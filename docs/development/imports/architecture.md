# Import System Architecture

## Overview

This document provides detailed technical architecture for the import resolution system in Topple. It complements the high-level implementation plan with specific data structures, algorithms, and integration points.

## System Components

### 1. Module Resolver

**Package**: `compiler/module/`

**Purpose**: Translate import paths to filesystem paths

#### Interface Design

```go
package moduleresolver

import (
    "context"
    "github.com/topple-psx/topple/internal/filesystem"
)

// ModuleResolver translates import paths to file paths
type ModuleResolver interface {
    // Resolve an absolute import path to a file path
    // Example: "my_module" -> "/path/to/my_module.psx"
    ResolveAbsolute(ctx context.Context, modulePath string) (string, error)

    // Resolve a relative import path from a source file
    // Example: ".sibling", "/path/to/current.psx" -> "/path/to/sibling.psx"
    ResolveRelative(ctx context.Context, dotCount int, modulePath string, sourceFile string) (string, error)

    // Check if a module exists
    Exists(ctx context.Context, modulePath string) bool

    // Get search paths for module resolution
    SearchPaths() []string
}

// Config for module resolution
type Config struct {
    // Root directory for resolution (usually cwd or project root)
    RootDir string

    // Additional search paths (future: for package management)
    SearchPaths []string

    // Filesystem abstraction
    FileSystem filesystem.FileSystem
}

// Standard implementation
type StandardResolver struct {
    config Config
    cache  map[string]string // Import path -> resolved file path
}

func NewResolver(config Config) *StandardResolver {
    return &StandardResolver{
        config: config,
        cache:  make(map[string]string),
    }
}
```

#### Resolution Algorithm

**Absolute Import Resolution**:
```
Input: "my_module"
Process:
  1. For each search path in [RootDir, SearchPaths...]:
     a. Try: searchPath/my_module.psx
     b. Try: searchPath/my_module/__init__.psx
     c. If found, return absolute path
  2. Return error: module not found

Examples:
  "components" -> ./components.psx (file)
  "components" -> ./components/__init__.psx (package)
  "lib.utils" -> ./lib/utils.psx (nested module)
```

**Relative Import Resolution**:
```
Input: dotCount=1, modulePath="sibling", sourceFile="/proj/app.psx"
Process:
  1. Get source directory: /proj/
  2. Navigate up by dotCount-1: /proj/ (1-1=0, stay)
  3. Append modulePath: /proj/sibling
  4. Try: /proj/sibling.psx
  5. Try: /proj/sibling/__init__.psx
  6. If found, return absolute path
  7. Return error: module not found

Examples:
  Source: /proj/views/app.psx
  "." -> /proj/views/__init__.psx (current package)
  ".components" -> /proj/views/components.psx
  "..utils" -> /proj/utils.psx (parent level)
  "...config" -> /config.psx (two levels up)
```

#### Error Handling

```go
type ResolutionError struct {
    ImportPath   string
    SourceFile   string
    SearchedPaths []string
    ErrorType    ErrorType
}

type ErrorType int

const (
    ModuleNotFound ErrorType = iota
    AmbiguousModule  // Multiple candidates found
    InvalidRelativeImport  // Too many dots
    InvalidPath  // Malformed import path
)

func (e *ResolutionError) Error() string {
    switch e.ErrorType {
    case ModuleNotFound:
        return fmt.Sprintf(
            "cannot resolve import '%s'\n"+
            "  in file: %s\n"+
            "  searched: %v",
            e.ImportPath, e.SourceFile, e.SearchedPaths,
        )
    // ... other cases
    }
}
```

---

### 2. Symbol Registry

**Package**: `compiler/symbol/`

**Purpose**: Track and lookup symbols exported from modules

#### Data Structures

```go
package symbolregistry

import (
    "github.com/topple-psx/topple/compiler/ast"
)

// SymbolRegistry tracks all exported symbols across modules
type SymbolRegistry struct {
    modules map[string]*ModuleSymbols  // Absolute file path -> symbols
    mu      sync.RWMutex                // Thread-safe access
}

// ModuleSymbols represents all symbols exported from one module
type ModuleSymbols struct {
    FilePath string
    Symbols  map[string]*Symbol  // Symbol name -> Symbol
}

// Symbol represents an exported name (view, function, class, variable)
type Symbol struct {
    Name       string
    Type       SymbolType
    Node       ast.Node      // Original AST node
    Location   Location      // File and line number
    Visibility Visibility    // Public or Private
    Docstring  string        // For future documentation support
}

type SymbolType int

const (
    SymbolView SymbolType = iota
    SymbolFunction
    SymbolClass
    SymbolVariable
)

type Visibility int

const (
    Public Visibility = iota   // Normal names
    Private                     // Underscore-prefixed
)

type Location struct {
    File   string
    Line   int
    Column int
}
```

#### Interface

```go
type Registry interface {
    // Register symbols from a module
    RegisterModule(filePath string, symbols *ModuleSymbols) error

    // Get all symbols from a module
    GetModuleSymbols(filePath string) (*ModuleSymbols, error)

    // Lookup a specific symbol in a module
    LookupSymbol(filePath string, symbolName string) (*Symbol, error)

    // Get all public symbols (for wildcard imports)
    GetPublicSymbols(filePath string) ([]*Symbol, error)

    // Check if module is registered
    HasModule(filePath string) bool

    // Clear all symbols (for testing)
    Clear()
}
```

#### Symbol Collection

```go
// SymbolCollector extracts exportable symbols from AST
type SymbolCollector struct {
    currentFile string
    symbols     map[string]*Symbol
}

func (sc *SymbolCollector) CollectFromAST(module *ast.Module, filePath string) *ModuleSymbols {
    sc.currentFile = filePath
    sc.symbols = make(map[string]*Symbol)

    // Visit all top-level statements
    for _, stmt := range module.Body {
        sc.visitStatement(stmt)
    }

    return &ModuleSymbols{
        FilePath: filePath,
        Symbols:  sc.symbols,
    }
}

func (sc *SymbolCollector) visitStatement(stmt ast.Stmt) {
    switch s := stmt.(type) {
    case *ast.ViewDef:
        sc.addSymbol(s.Name, SymbolView, s)
    case *ast.FunctionDef:
        sc.addSymbol(s.Name, SymbolFunction, s)
    case *ast.ClassDef:
        sc.addSymbol(s.Name.Value, SymbolClass, s)
    case *ast.AssignStmt:
        // Track module-level variables
        for _, target := range s.Targets {
            if name, ok := target.(*ast.Name); ok {
                sc.addSymbol(name.Value, SymbolVariable, s)
            }
        }
    }
}

func (sc *SymbolCollector) addSymbol(name string, typ SymbolType, node ast.Node) {
    visibility := Public
    if len(name) > 0 && name[0] == '_' {
        visibility = Private
    }

    sc.symbols[name] = &Symbol{
        Name:       name,
        Type:       typ,
        Node:       node,
        Location:   extractLocation(node),
        Visibility: visibility,
    }
}
```

---

### 3. Dependency Graph

**Package**: `compiler/depgraph/`

**Purpose**: Build and analyze file dependencies

#### Data Structures

```go
package depgraph

import (
    "github.com/topple-psx/topple/compiler/ast"
)

// DependencyGraph represents dependencies between files
type DependencyGraph struct {
    nodes map[string]*FileNode     // File path -> node
    edges map[string][]string      // File path -> dependencies
}

// FileNode represents a single file in the graph
type FileNode struct {
    FilePath string
    AST      *ast.Module
    Imports  []*Import
}

// Import represents a single import statement
type Import struct {
    Statement  ast.Node     // ImportStmt or ImportFromStmt
    ModulePath string       // Resolved file path
    Names      []string     // Imported names (empty for "import x")
    IsWildcard bool
    Location   Location
}

type Location struct {
    Line   int
    Column int
}
```

#### Graph Operations

```go
type Graph interface {
    // Add a file to the graph
    AddFile(filePath string, ast *ast.Module) error

    // Add a dependency edge
    AddDependency(from, to string) error

    // Get compilation order (topological sort)
    GetCompilationOrder() ([]string, error)

    // Detect circular dependencies
    DetectCycles() ([][]string, error)

    // Get direct dependencies of a file
    GetDependencies(filePath string) []string

    // Get reverse dependencies (files that import this file)
    GetDependents(filePath string) []string
}
```

#### Topological Sort (Kahn's Algorithm)

```go
func (g *DependencyGraph) GetCompilationOrder() ([]string, error) {
    // 1. Calculate in-degrees (number of dependencies)
    inDegree := make(map[string]int)
    for file := range g.nodes {
        inDegree[file] = 0
    }
    for _, deps := range g.edges {
        for _, dep := range deps {
            inDegree[dep]++
        }
    }

    // 2. Queue files with no dependencies
    queue := []string{}
    for file, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, file)
        }
    }

    // 3. Process files in order
    result := []string{}
    for len(queue) > 0 {
        // Dequeue
        file := queue[0]
        queue = queue[1:]
        result = append(result, file)

        // Reduce in-degree of dependents
        for _, dep := range g.edges[file] {
            inDegree[dep]--
            if inDegree[dep] == 0 {
                queue = append(queue, dep)
            }
        }
    }

    // 4. Check for cycles
    if len(result) != len(g.nodes) {
        cycles := g.detectCyclesInternal()
        return nil, &CycleError{Cycles: cycles}
    }

    return result, nil
}
```

#### Cycle Detection (DFS)

```go
func (g *DependencyGraph) DetectCycles() ([][]string, error) {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    cycles := [][]string{}

    for file := range g.nodes {
        if !visited[file] {
            path := []string{}
            g.dfsCycle(file, visited, recStack, path, &cycles)
        }
    }

    if len(cycles) > 0 {
        return cycles, &CycleError{Cycles: cycles}
    }
    return nil, nil
}

func (g *DependencyGraph) dfsCycle(
    file string,
    visited map[string]bool,
    recStack map[string]bool,
    path []string,
    cycles *[][]string,
) {
    visited[file] = true
    recStack[file] = true
    path = append(path, file)

    for _, dep := range g.edges[file] {
        if !visited[dep] {
            g.dfsCycle(dep, visited, recStack, path, cycles)
        } else if recStack[dep] {
            // Found cycle
            cycle := extractCycle(path, dep)
            *cycles = append(*cycles, cycle)
        }
    }

    recStack[file] = false
}
```

#### Error Reporting

```go
type CycleError struct {
    Cycles [][]string
}

func (e *CycleError) Error() string {
    var sb strings.Builder
    sb.WriteString("circular dependencies detected:\n")

    for i, cycle := range e.Cycles {
        sb.WriteString(fmt.Sprintf("  Cycle %d:\n", i+1))
        for j, file := range cycle {
            sb.WriteString(fmt.Sprintf("    %s\n", file))
            if j < len(cycle)-1 {
                sb.WriteString("     ↓ imports\n")
            } else {
                sb.WriteString("     ↓ imports (circular)\n")
            }
        }
    }

    return sb.String()
}
```

---

### 4. Resolver Integration

#### Updated Visitor Methods

```go
// In compiler/resolver/visitor.go

func (v *ResolverVisitor) VisitImportStmt(i *ast.ImportStmt) ast.Visitor {
    for _, alias := range i.Names {
        // Resolve module path
        modulePath, err := v.moduleResolver.ResolveAbsolute(
            v.ctx,
            alias.DottedName.String(),
        )
        if err != nil {
            v.addError(i, err.Error())
            continue
        }

        // Get module symbols
        moduleSymbols, err := v.symbolRegistry.GetModuleSymbols(modulePath)
        if err != nil {
            v.addError(i, fmt.Sprintf("module '%s' not found", modulePath))
            continue
        }

        // Create binding for imported module name
        name := alias.AsName
        if name == "" {
            // Use first component of dotted name
            name = alias.DottedName.Names[0]
        }

        v.defineImportedModule(name, modulePath, i)
    }

    return v
}

func (v *ResolverVisitor) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor {
    // Resolve module path
    var modulePath string
    var err error

    if i.DotCount > 0 {
        // Relative import
        modulePath, err = v.moduleResolver.ResolveRelative(
            v.ctx,
            i.DotCount,
            i.DottedName.String(),
            v.currentFile,
        )
    } else {
        // Absolute import
        modulePath, err = v.moduleResolver.ResolveAbsolute(
            v.ctx,
            i.DottedName.String(),
        )
    }

    if err != nil {
        v.addError(i, err.Error())
        return v
    }

    // Get module symbols
    moduleSymbols, err := v.symbolRegistry.GetModuleSymbols(modulePath)
    if err != nil {
        v.addError(i, fmt.Sprintf("module '%s' not found", modulePath))
        return v
    }

    // Handle wildcard vs specific imports
    if i.IsWildcard {
        // Import all public symbols
        symbols := moduleSymbols.GetPublicSymbols()
        for _, symbol := range symbols {
            v.defineImportedSymbol(symbol.Name, modulePath, symbol, i)
        }
    } else {
        // Import specific names
        for _, name := range i.Names {
            symbol, err := moduleSymbols.LookupSymbol(name.Name)
            if err != nil {
                v.addError(i, fmt.Sprintf(
                    "cannot import '%s' from '%s': symbol not found",
                    name.Name,
                    modulePath,
                ))
                continue
            }

            importName := name.AsName
            if importName == "" {
                importName = name.Name
            }

            v.defineImportedSymbol(importName, modulePath, symbol, i)
        }
    }

    return v
}
```

#### New Resolver Methods

```go
// In compiler/resolver/resolver.go

func (r *Resolver) defineImportedModule(
    name string,
    modulePath string,
    node ast.Node,
) {
    variable := &Variable{
        Name:           name,
        DefinitionDepth: r.currentScope.GetDepth(),
        State:          Defined,
        IsImported:     true,
        ImportedFrom:   modulePath,
    }

    binding := &Binding{
        Name:     name,
        Variable: variable,
        Scope:    r.currentScope,
        Node:     node,
    }

    r.currentScope.Bindings[name] = binding
}

func (r *Resolver) defineImportedSymbol(
    name string,
    modulePath string,
    symbol *Symbol,
    node ast.Node,
) {
    variable := &Variable{
        Name:           name,
        DefinitionDepth: r.currentScope.GetDepth(),
        State:          Defined,
        IsImported:     true,
        ImportedFrom:   modulePath,
        ImportedSymbol: symbol,
    }

    binding := &Binding{
        Name:     name,
        Variable: variable,
        Scope:    r.currentScope,
        Node:     node,
    }

    r.currentScope.Bindings[name] = binding
}
```

#### Updated Variable Struct

```go
// In compiler/resolver/types.go

type Variable struct {
    Name           string
    DefinitionDepth int
    State          VariableState

    // Import tracking
    IsImported     bool
    ImportedFrom   string   // Module file path
    ImportedSymbol *Symbol  // Actual symbol from other module

    // Existing fields...
    IsParameter    bool
    IsGlobal       bool
    IsNonlocal     bool
    // ...
}
```

---

### 5. Multi-File Compiler

#### New Compiler Interface

```go
// In compiler/compiler.go

// MultiFileCompiler compiles multiple interdependent files
type MultiFileCompiler struct {
    config           CompilerConfig
    fileSystem       filesystem.FileSystem
    moduleResolver   *moduleresolver.StandardResolver
    symbolRegistry   *symbolregistry.SymbolRegistry
    dependencyGraph  *depgraph.DependencyGraph
}

type CompilerConfig struct {
    RootDir     string
    SearchPaths []string
    Recursive   bool
}

func NewMultiFileCompiler(config CompilerConfig) *MultiFileCompiler {
    fs := filesystem.NewOSFileSystem()

    return &MultiFileCompiler{
        config:     config,
        fileSystem: fs,
        moduleResolver: moduleresolver.NewResolver(moduleresolver.Config{
            RootDir:    config.RootDir,
            SearchPaths: config.SearchPaths,
            FileSystem: fs,
        }),
        symbolRegistry:  symbolregistry.NewRegistry(),
        dependencyGraph: depgraph.NewGraph(),
    }
}
```

#### Compilation Pipeline

```go
func (c *MultiFileCompiler) CompileProject(
    ctx context.Context,
    files []string,
) (*ProjectOutput, []error) {
    var allErrors []error

    // Phase 1: Parse all files
    astMap := make(map[string]*ast.Module)
    for _, file := range files {
        content, err := c.fileSystem.ReadFile(file)
        if err != nil {
            allErrors = append(allErrors, err)
            continue
        }

        // Lex + Parse
        tokens := lexer.Scan(content)
        module, errors := parser.Parse(tokens)
        if len(errors) > 0 {
            allErrors = append(allErrors, errors...)
            continue
        }

        astMap[file] = module
        c.dependencyGraph.AddFile(file, module)
    }

    if len(allErrors) > 0 {
        return nil, allErrors
    }

    // Phase 2: Extract imports and build dependency graph
    for file, module := range astMap {
        imports := extractImports(module, file, c.moduleResolver)
        for _, imp := range imports {
            c.dependencyGraph.AddDependency(file, imp.ModulePath)
        }
    }

    // Phase 3: Get compilation order
    order, err := c.dependencyGraph.GetCompilationOrder()
    if err != nil {
        return nil, []error{err}
    }

    // Phase 4: Collect symbols (first pass - no resolution yet)
    for _, file := range order {
        symbols := collectSymbols(astMap[file], file)
        c.symbolRegistry.RegisterModule(file, symbols)
    }

    // Phase 5: Compile each file with full context
    outputs := make(map[string][]byte)
    for _, file := range order {
        module := astMap[file]

        // Resolve with full symbol registry
        resTable, errors := c.resolve(ctx, module, file)
        if len(errors) > 0 {
            allErrors = append(allErrors, errors...)
            continue
        }

        // Transform
        transformed := c.transform(module, resTable)

        // Generate
        code := c.generate(transformed)

        outputs[file] = code
    }

    if len(allErrors) > 0 {
        return nil, allErrors
    }

    return &ProjectOutput{
        Files:    outputs,
        Registry: c.symbolRegistry,
        Graph:    c.dependencyGraph,
    }, nil
}

type ProjectOutput struct {
    Files    map[string][]byte
    Registry *symbolregistry.SymbolRegistry
    Graph    *depgraph.DependencyGraph
}
```

---

## Integration Points

### CLI Integration

```go
// In cmd/compile.go

func compileCommand() *cobra.Command {
    var recursive bool
    var multiFile bool

    cmd := &cobra.Command{
        Use: "compile [files...]",
        RunE: func(cmd *cobra.Command, args []string) error {
            if multiFile || len(args) > 1 {
                return compileMultiFile(args, recursive)
            }
            return compileSingleFile(args[0])
        },
    }

    cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "compile recursively")
    cmd.Flags().BoolVar(&multiFile, "multi-file", false, "enable multi-file mode")

    return cmd
}

func compileMultiFile(paths []string, recursive bool) error {
    compiler := compiler.NewMultiFileCompiler(compiler.CompilerConfig{
        RootDir:   ".",
        Recursive: recursive,
    })

    files, err := collectFiles(paths, recursive)
    if err != nil {
        return err
    }

    output, errors := compiler.CompileProject(context.Background(), files)
    if len(errors) > 0 {
        for _, err := range errors {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        }
        return fmt.Errorf("compilation failed with %d errors", len(errors))
    }

    // Write outputs
    for file, code := range output.Files {
        outputPath := getOutputPath(file)
        if err := os.WriteFile(outputPath, code, 0644); err != nil {
            return err
        }
        fmt.Printf("Compiled: %s -> %s\n", file, outputPath)
    }

    return nil
}
```

---

## Testing Infrastructure

### Module Resolver Tests

```go
// compiler/module/resolver_test.go

func TestResolveAbsolute(t *testing.T) {
    tests := []struct {
        name       string
        modulePath string
        files      map[string]string  // Virtual filesystem
        want       string
        wantErr    bool
    }{
        {
            name:       "simple module",
            modulePath: "my_module",
            files: map[string]string{
                "my_module.psx": "view X(): <div/>",
            },
            want:    "/my_module.psx",
            wantErr: false,
        },
        {
            name:       "package with __init__",
            modulePath: "components",
            files: map[string]string{
                "components/__init__.psx": "view Y(): <div/>",
            },
            want:    "/components/__init__.psx",
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fs := filesystem.NewMemoryFileSystem(tt.files)
            resolver := NewResolver(Config{
                RootDir:    "/",
                FileSystem: fs,
            })

            got, err := resolver.ResolveAbsolute(context.Background(), tt.modulePath)
            if (err != nil) != tt.wantErr {
                t.Errorf("ResolveAbsolute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ResolveAbsolute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Test Structure

```
compiler/testdata/input/imports/
├── absolute/
│   ├── module.psx          # Simple module
│   ├── importer.psx        # Imports from module
│   └── expected/
│       ├── module.py
│       └── importer.py
├── relative/
│   ├── dir1/
│   │   └── file1.psx
│   ├── dir2/
│   │   └── file2.psx       # Imports from ../dir1/file1
│   └── expected/
│       ├── dir1/file1.py
│       └── dir2/file2.py
└── circular/
    ├── a.psx               # Imports b
    ├── b.psx               # Imports a
    └── error.txt           # Expected error message
```

---

## Performance Considerations

### Caching Strategy

```go
// ModuleResolver with caching
type CachedResolver struct {
    inner *StandardResolver
    cache sync.Map  // import path -> resolved path
}

func (r *CachedResolver) ResolveAbsolute(ctx context.Context, path string) (string, error) {
    if cached, ok := r.cache.Load(path); ok {
        return cached.(string), nil
    }

    resolved, err := r.inner.ResolveAbsolute(ctx, path)
    if err == nil {
        r.cache.Store(path, resolved)
    }

    return resolved, err
}
```

### Parallel Compilation

```go
// Compile independent files in parallel
func (c *MultiFileCompiler) CompileParallel(
    files []string,
    order []string,
) (map[string][]byte, []error) {
    // Group by dependency level
    levels := c.dependencyGraph.GetLevels(order)

    outputs := make(map[string][]byte)
    var mu sync.Mutex
    var allErrors []error

    // Compile each level in parallel
    for _, level := range levels {
        var wg sync.WaitGroup
        for _, file := range level {
            wg.Add(1)
            go func(f string) {
                defer wg.Done()

                code, err := c.compileFile(f)
                mu.Lock()
                if err != nil {
                    allErrors = append(allErrors, err)
                } else {
                    outputs[f] = code
                }
                mu.Unlock()
            }(file)
        }
        wg.Wait()

        // Stop if any errors in this level
        if len(allErrors) > 0 {
            break
        }
    }

    return outputs, allErrors
}
```

---

## Migration Path

### Backward Compatibility

The import system must maintain backward compatibility with single-file compilation:

```go
// SingleFileCompiler remains unchanged
type SingleFileCompiler struct {
    // ... existing implementation
}

func (c *SingleFileCompiler) Compile(
    ctx context.Context,
    file File,
) ([]byte, []error) {
    // Existing single-file logic unchanged
}

// CLI automatically selects mode
func compile(files []string) error {
    if len(files) == 1 && !hasImports(files[0]) {
        // Use fast single-file path
        return compileSingle(files[0])
    }
    // Use multi-file path
    return compileMulti(files)
}
```

---

*This architecture supports incremental implementation, allowing each component to be developed and tested independently before integration.*
