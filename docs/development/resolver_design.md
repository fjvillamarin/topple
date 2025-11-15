# Variable Resolver Design for Topple Compiler

## Overview

The Variable Resolver is a compiler pass that runs after parsing and before code generation. It implements Python's LEGB (Local, Enclosing, Global, Built-in) scoping rules and binds variable names to their definitions, enabling proper variable resolution and scope analysis.

**Status**: ✅ Fully implemented and integrated into the compilation pipeline

**Location**: `compiler/resolver/`
- `types.go` - Core data structures (Scope, Binding, Variable, ResolutionTable)
- `resolver.go` - Main resolver logic (scope management, LEGB resolution)
- `visitor.go` - AST visitor implementations for all node types
- `debug.go` - Debug printing utilities for resolution results

## Goals

1. **Variable Binding**: Bind every `Name` node in the AST to its corresponding variable definition ✅
2. **Scope Analysis**: Track variable scopes and detect scope-related errors ✅
3. **Python Compatibility**: Handle Python's complex scoping rules (global/nonlocal declarations, class scopes, comprehensions) ✅
4. **Topple Extensions**: Support Topple-specific constructs like view parameters ✅
5. **Error Detection**: Identify undefined variables, invalid global/nonlocal declarations ✅
6. **View Composition**: Track view definitions and HTML element references for view composition ✅
7. **Closure Analysis**: Identify cell variables and free variables for proper closure handling ✅

## Current Implementation Architecture

The resolver uses a **scope chain system** with rich binding tracking. This differs from a simple environment stack by explicitly modeling:
- Individual scopes as first-class objects with unique IDs
- Bindings as the connection between a name and a variable in a specific scope
- Shadowing relationships between bindings in nested scopes
- Both legacy pointer-based and modern scope chain-based lookups

### Key Components

1. **Scope Chain**: Linked list of `Scope` objects from current scope to module scope
2. **Binding System**: Maps variable names to `Variable` metadata within each scope
3. **Resolution Table**: Output containing all resolution results, mappings, and metadata
4. **Visitor Pattern**: Traverses AST and performs resolution using visitor methods

## Core Data Structures (Current Implementation)

### Variable Metadata

```go
type Variable struct {
    Name           string        // Variable name
    DefinitionDepth int          // Absolute scope depth where defined
    State          VariableState // Current state of the variable
    
    // Scope behavior flags
    IsParameter    bool          // Function/view parameter
    IsGlobal       bool          // Declared with 'global'
    IsNonlocal     bool          // Declared with 'nonlocal'
    IsImported     bool          // Bound by import statement
    IsViewParameter bool         // Topple view parameter
    IsExceptionVar bool          // Exception handler variable
    
    // Usage tracking
    FirstDefSpan   lexer.Span    // Where first defined
    FirstUseSpan   lexer.Span    // Where first used
    UsedBeforeDef  bool          // Python late binding detection
    
    // Closure analysis
    IsCaptured     bool          // Used in nested functions
    IsCell         bool          // Needs cell variable for closures
}

type VariableState int
const (
    VariableUndefined VariableState = iota
    VariableDeclared     // Declared but not assigned
    VariableDefined      // Has been assigned
    VariableUsed         // Has been referenced
)
```

### Environment/Scope Chain

```go
type Environment struct {
    Enclosing   *Environment              // Parent scope
    Values      map[string]*Variable      // Variable bindings in this scope
    ScopeType   ScopeType                // Type of scope
    
    // Python-specific scope declarations
    Globals     map[string]*Variable      // 'global' declarations
    Nonlocals   map[string]*Variable      // 'nonlocal' declarations
    IsClassScope bool                     // Class scopes have special rules
}

type ScopeType int
const (
    ModuleScopeType ScopeType = iota
    FunctionScopeType
    ClassScopeType
    ViewScopeType              // Topple view scope
    ComprehensionScopeType
    ExceptScopeType            // Exception handler scope
    WithScopeType             // With statement scope
)
```

### Resolver State

```go
type Resolver struct {
    // Scope management
    Scopes        []*Environment          // Stack of active scopes
    Current       *Environment            // Current scope
    ModuleGlobals map[string]*Variable    // Module-level variables
    
    // Resolution results
    Variables     map[*ast.Name]*Variable // Name node → Variable mapping
    ScopeDepths   map[*ast.Name]int      // Name node → scope distance
    
    // Closure analysis
    CellVars      map[string]bool         // Variables needing cells
    FreeVars      map[string]bool         // Free variables
    
    // Error tracking
    Errors        []error
    
    // Context tracking
    InClassScope    bool
    InFunctionScope bool
    InViewScope     bool                  // Topple views
    CurrentFunction *ast.Function
    CurrentView     *ast.ViewStmt
}
```

## Python Scoping Rules Implementation

### LEGB Resolution Order

```go
func (r *Resolver) ResolveName(name *ast.Name) error {
    varName := name.Token.Lexeme
    currentDepth := len(r.Scopes) - 1
    
    // Check for global/nonlocal declarations in current scope
    if globalVar := r.Current.Globals[varName]; globalVar != nil {
        r.Variables[name] = globalVar
        globalVar.IsUsed = true
        return nil
    }
    
    if nonlocalVar := r.Current.Nonlocals[varName]; nonlocalVar != nil {
        r.Variables[name] = nonlocalVar
        nonlocalVar.IsUsed = true
        nonlocalVar.IsCaptured = true
        return nil
    }
    
    // LEGB Resolution: Local → Enclosing → Global → Builtin
    
    // 1. Local scope
    if variable, exists := r.Current.Values[varName]; exists {
        r.Variables[name] = variable
        variable.IsUsed = true
        r.ScopeDepths[name] = 0
        return nil
    }
    
    // 2. Enclosing scopes (skip class scopes for function lookups)
    distance := 1
    for i := len(r.Scopes) - 2; i >= 1; i-- {
        scope := r.Scopes[i]
        
        // Class scopes don't participate in LEGB for nested functions
        if scope.IsClassScope && r.InFunctionScope {
            distance++
            continue
        }
        
        if variable, exists := scope.Values[varName]; exists {
            r.Variables[name] = variable
            variable.IsUsed = true
            variable.IsCaptured = true
            r.ScopeDepths[name] = distance
            r.CellVars[varName] = true
            return nil
        }
        distance++
    }
    
    // 3. Global scope (module level)
    if variable, exists := r.ModuleGlobals[varName]; exists {
        r.Variables[name] = variable
        variable.IsUsed = true
        return nil
    }
    
    // 4. Built-ins (handled at runtime)
    return fmt.Errorf("name '%s' is not defined", varName)
}
```

### Global/Nonlocal Declaration Handling

```go
func (r *Resolver) VisitGlobalStmt(g *ast.GlobalStmt) Visitor {
    if r.Current.ScopeType == ModuleScopeType {
        r.ReportError(fmt.Errorf("'global' declaration at module level"))
        return r
    }
    
    for _, name := range g.Names {
        varName := name.Token.Lexeme
        
        // Create or reference module-level variable
        if variable, exists := r.ModuleGlobals[varName]; exists {
            r.Current.Globals[varName] = variable
        } else {
            variable := &Variable{
                Name: varName,
                IsGlobal: true,
                DefinitionDepth: 0,
                FirstDefSpan: name.Span,
            }
            r.ModuleGlobals[varName] = variable
            r.Current.Globals[varName] = variable
        }
    }
    return r
}

func (r *Resolver) VisitNonlocalStmt(n *ast.NonlocalStmt) Visitor {
    if r.Current.ScopeType != FunctionScopeType && r.Current.ScopeType != ViewScopeType {
        r.ReportError(fmt.Errorf("'nonlocal' declaration not in function scope"))
        return r
    }
    
    for _, name := range n.Names {
        varName := name.Token.Lexeme
        
        // Find in enclosing scopes (not global, not current)
        found := false
        for i := len(r.Scopes) - 2; i >= 1; i-- {
            if variable, exists := r.Scopes[i].Values[varName]; exists {
                r.Current.Nonlocals[varName] = variable
                found = true
                break
            }
        }
        
        if !found {
            r.ReportError(fmt.Errorf("no binding for nonlocal '%s' found", varName))
        }
    }
    return r
}
```

## Visitor Implementation

### Scope Management

```go
func (r *Resolver) BeginScope(scopeType ScopeType) {
    env := &Environment{
        Enclosing: r.Current,
        Values:    make(map[string]*Variable),
        ScopeType: scopeType,
        Globals:   make(map[string]*Variable),
        Nonlocals: make(map[string]*Variable),
        IsClassScope: scopeType == ClassScopeType,
    }
    r.Scopes = append(r.Scopes, env)
    r.Current = env
}

func (r *Resolver) EndScope() {
    if len(r.Scopes) > 0 {
        r.Scopes = r.Scopes[:len(r.Scopes)-1]
        if len(r.Scopes) > 0 {
            r.Current = r.Scopes[len(r.Scopes)-1]
        } else {
            r.Current = nil
        }
    }
}

func (r *Resolver) DefineVariable(name string, span lexer.Span) *Variable {
    variable := &Variable{
        Name: name,
        DefinitionDepth: len(r.Scopes) - 1,
        State: VariableDeclared,
        FirstDefSpan: span,
    }
    r.Current.Values[name] = variable
    return variable
}
```

### Statement Visitors

```go
func (r *Resolver) VisitFunction(f *ast.Function) Visitor {
    // Function name is bound in enclosing scope
    if f.Name != nil {
        r.DefineVariable(f.Name.Token.Lexeme, f.Name.Span)
    }
    
    // Function body has its own scope
    r.BeginScope(FunctionScopeType)
    oldFunction := r.CurrentFunction
    oldInFunction := r.InFunctionScope
    r.CurrentFunction = f
    r.InFunctionScope = true
    
    defer func() {
        r.EndScope()
        r.CurrentFunction = oldFunction
        r.InFunctionScope = oldInFunction
    }()
    
    // Parameters create local bindings
    if f.Parameters != nil {
        for _, param := range f.Parameters.Parameters {
            if param.Name != nil {
                variable := r.DefineVariable(param.Name.Token.Lexeme, param.Name.Span)
                variable.IsParameter = true
                variable.State = VariableDefined
            }
            
            // Visit default values in enclosing scope
            if param.Default != nil {
                param.Default.Accept(r)
            }
        }
    }
    
    // Visit function body
    for _, stmt := range f.Body {
        stmt.Accept(r)
    }
    
    return r
}

func (r *Resolver) VisitViewStmt(v *ast.ViewStmt) Visitor {
    // View name is bound in enclosing scope
    if v.Name != nil {
        r.DefineVariable(v.Name.Token.Lexeme, v.Name.Span)
    }
    
    // View body has its own scope
    r.BeginScope(ViewScopeType)
    oldView := r.CurrentView
    oldInView := r.InViewScope
    r.CurrentView = v
    r.InViewScope = true
    
    defer func() {
        r.EndScope()
        r.CurrentView = oldView
        r.InViewScope = oldInView
    }()
    
    // View parameters create local bindings
    if v.Params != nil {
        for _, param := range v.Params.Parameters {
            if param.Name != nil {
                variable := r.DefineVariable(param.Name.Token.Lexeme, param.Name.Span)
                variable.IsParameter = true
                variable.IsViewParameter = true
                variable.State = VariableDefined
            }
        }
    }
    
    // Visit view body
    for _, stmt := range v.Body {
        stmt.Accept(r)
    }
    
    return r
}

func (r *Resolver) VisitAssignStmt(a *ast.AssignStmt) Visitor {
    // Visit the value first (right-hand side)
    a.Value.Accept(r)
    
    // Then handle targets (which create/update bindings)
    for _, target := range a.Targets {
        r.AnalyzeAssignmentTarget(target)
    }
    
    return r
}

func (r *Resolver) AnalyzeAssignmentTarget(target ast.Expr) {
    switch t := target.(type) {
    case *ast.Name:
        // Simple assignment - define or update variable
        if variable, exists := r.Current.Values[t.Token.Lexeme]; exists {
            variable.State = VariableDefined
        } else {
            variable := r.DefineVariable(t.Token.Lexeme, t.Span)
            variable.State = VariableDefined
        }
        
    case *ast.TupleExpr:
        // Tuple unpacking
        for _, elem := range t.Elements {
            r.AnalyzeAssignmentTarget(elem)
        }
        
    case *ast.ListExpr:
        // List unpacking
        for _, elem := range t.Elements {
            r.AnalyzeAssignmentTarget(elem)
        }
        
    case *ast.StarExpr:
        // Starred expression in unpacking
        r.AnalyzeAssignmentTarget(t.Value)
        
    case *ast.Attribute:
        // Attribute assignment - only resolve the object
        t.Value.Accept(r)
        
    case *ast.Subscript:
        // Subscript assignment - resolve both parts
        t.Value.Accept(r)
        t.Slice.Accept(r)
        
    default:
        r.ReportError(fmt.Errorf("invalid assignment target"))
    }
}
```

### Expression Visitors

```go
func (r *Resolver) VisitName(name *ast.Name) Visitor {
    if err := r.ResolveName(name); err != nil {
        r.ReportError(err)
    }
    return r
}

func (r *Resolver) VisitListComp(lc *ast.ListComp) Visitor {
    // List comprehensions have their own scope in Python 3+
    r.BeginScope(ComprehensionScopeType)
    defer r.EndScope()
    
    // Process generators
    for _, comp := range lc.Generators {
        // Visit iterable in enclosing scope
        comp.Iter.Accept(r)
        
        // Target creates binding in comprehension scope
        r.AnalyzeAssignmentTarget(comp.Target)
        
        // Conditions in comprehension scope
        for _, condition := range comp.Ifs {
            condition.Accept(r)
        }
    }
    
    // Element expression in comprehension scope
    lc.Element.Accept(r)
    
    return r
}
```

## Integration with Compiler Pipeline

### Resolution Result

```go
type ResolutionTable struct {
    Variables       map[*ast.Name]*Variable // Name → Variable mapping
    ScopeDepths     map[*ast.Name]int      // Name → scope distance
    ViewParameters  map[string]*Variable    // View parameter tracking
    CellVars        map[string]bool         // Variables needing cells
    FreeVars        map[string]bool         // Free variables
    Errors          []error                 // Resolution errors
}
```

### Compiler Integration

```go
func (c *StandardCompiler) Compile(ctx context.Context, file File) ([]byte, []error) {
    ast, errors := Parse(file.Content)
    if len(errors) > 0 {
        return nil, errors
    }

    // Variable resolution phase
    resolver := NewResolver()
    resolutionTable, err := resolver.Resolve(ast)
    if err != nil {
        return nil, []error{err}
    }
    if len(resolutionTable.Errors) > 0 {
        return nil, resolutionTable.Errors
    }

    // Transformation phase (with resolution info)
    transformerVisitor := transformers.NewTransformerVisitor(
        transformers.NewViewTransformer(resolutionTable))
    ast, err = transformerVisitor.TransformModule(ast)
    if err != nil {
        return nil, []error{err}
    }

    // Code generation phase
    generator := codegen.NewCodeGeneratorWithResolution(resolutionTable)
    result := generator.Generate(ast)

    return []byte(result), nil
}
```

## Special Considerations

### Class Scope Isolation

Python class scopes don't participate in the LEGB chain for nested functions:

```python
x = "global"

class C:
    x = "class"
    
    def method(self):
        def inner():
            return x  # Refers to "global", not "class"
        return inner()
```

The resolver handles this by skipping class scopes when resolving names in function contexts.

### View Parameter Transformation

Topple view parameters need special handling for the transformation phase:

```go
func (r *Resolver) VisitViewStmt(v *ast.ViewStmt) Visitor {
    // ... scope setup ...
    
    // Mark view parameters for transformation
    if v.Params != nil {
        for _, param := range v.Params.Parameters {
            if param.Name != nil {
                variable := r.DefineVariable(param.Name.Token.Lexeme, param.Name.Span)
                variable.IsViewParameter = true  // Special flag for transformer
            }
        }
    }
    
    // ... rest of method ...
}
```

### Comprehension Scoping

Python 3+ comprehensions have their own scope:

```python
x = "outer"
result = [x for x in range(3)]  # x is local to comprehension
print(x)  # Still "outer"
```

### Exception Variable Scoping

Exception variables in `except` clauses have special scoping rules and are automatically deleted at the end of the handler.

## Error Detection

The resolver detects various scoping errors:

1. **Undefined variables**: Names that don't resolve to any binding
2. **Invalid global declarations**: `global` at module level
3. **Invalid nonlocal declarations**: `nonlocal` without enclosing binding
4. **Global/nonlocal conflicts**: Same name declared as both global and nonlocal

## Performance Considerations

- Variable lookups use hash maps for O(1) average case
- Scope depth calculation is O(depth) but depth is typically small
- Memory usage scales with number of variables and scopes
- Resolution table provides fast lookup during code generation

## Current Status and Features

### Implemented Features

1. **Full LEGB Resolution** ✅
   - Local, Enclosing, Global, Builtin name resolution
   - Proper handling of global/nonlocal declarations
   - Class scope isolation (classes don't participate in LEGB for nested functions)

2. **Scope Management** ✅
   - Module, function, class, view, comprehension, except, and with scopes
   - Scope chain with unique IDs for each scope
   - Depth counters for nested scopes

3. **Binding Tracking** ✅
   - Name → Variable → Binding relationships
   - Shadowing detection and tracking
   - Scope-specific binding resolution

4. **Variable Metadata** ✅
   - State tracking (undefined, declared, defined, used)
   - Classification (parameter, global, nonlocal, imported, view parameter, exception var)
   - Usage tracking (first definition span, first use span)
   - Closure analysis (captured, cell variable)

5. **View Composition Support** ✅
   - Tracks module-level view definitions
   - Maps HTML elements to view references
   - Enables view nesting and composition

6. **Error Detection** ✅
   - Global declarations at module level
   - Nonlocal declarations without enclosing binding
   - Invalid assignment targets

7. **Debug Output** ✅
   - `DebugPrintResolutionTable()` - Comprehensive resolution results
   - `DebugPrintResolver()` - Current resolver state during resolution
   - `DebugPrintCurrentScope()` - Specific scope information
   - Formatted tables with variable states, types, depths, and flags

### Integration Points

```go
// In compiler/compiler.go
func (c *StandardCompiler) Compile(ctx context.Context, file File) ([]byte, []error) {
    ast, errors := Parse(file.Content)

    // Variable resolution phase
    r := resolver.NewResolver()
    resolutionTable, err := r.Resolve(ast)

    // Transformation phase uses resolution table
    transformerVisitor := transformers.NewTransformerVisitor()
    ast, err = transformerVisitor.TransformModule(ast, resolutionTable)

    // Code generation
    generator := codegen.NewCodeGenerator()
    result := generator.Generate(ast)

    return []byte(result), nil
}
```

### Debug Capabilities

The resolver has comprehensive debug output capabilities (in `debug.go`):

1. **Resolution Table Debug Output**
   - Errors with detailed messages
   - Variables table (name, state, type, depth, flags)
   - View parameters listing
   - View composition information
   - Closure analysis (cell vars, free vars)
   - Summary statistics

2. **Resolver State Debug Output**
   - Current scope chain visualization
   - Context information (function/class/view depth)
   - Module globals listing
   - Scope-specific variable listings

3. **Formatted Output Features**
   - Color-coded sections (using Unicode box-drawing characters)
   - Sorted variable lists for consistency
   - Duplicate detection (groups by Variable object)
   - Reference counting for multiply-referenced variables

## Implemented Features (Available Now)

1. **JSON Output Format** ✅
   - Serializes resolution table to `.res.json` for machine consumption
   - Includes scopes, variables, views, closure analysis, and diagnostics
   - Usage: `topple parse file.psx --format=json`
   - Location: `compiler/resolver/json.go`

2. **Human-Readable Output** ✅
   - Generates `.res` text files with formatted resolution info
   - Well-structured sections with visual separators
   - Usage: `topple parse file.psx --format=text`
   - Location: `compiler/resolver/text_output.go`

3. **Parse Command Integration** ✅
   - Resolution output fully integrated into `topple parse` command
   - Supports single files and recursive directory processing
   - Usage: `topple parse file.psx -w --format=all` (generates .ast, .res, .res.json)
   - Location: `cmd/parse.go`

## Future Extensions

1. **Type annotations**: Could be extended to track type information
2. **Dataflow analysis**: Track variable assignments and usage patterns
3. **Optimization hints**: Identify variables that can be optimized
4. **Import analysis**: Better handling of imported names
5. **Topple-specific constructs**: Slot parameters, special view behaviors
6. **Enhanced LSP integration**: Real-time resolution feedback in editors
7. **Web-based visualization**: Interactive exploration of resolution data