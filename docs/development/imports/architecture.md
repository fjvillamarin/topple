# Import System Architecture

## Overview

This document provides detailed technical architecture for the import resolution system in Topple. It complements the high-level implementation plan with specific data structures, algorithms, and integration points.

## System Components

### 1. Module Resolver

**Package**: `compiler/module/`

**Purpose**: Translate import paths to filesystem paths

#### Interface Design

```go
package module

import (
    "context"
    "github.com/topple-psx/topple/internal/filesystem"
)

// Resolver translates import paths to file paths
type Resolver interface {
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
package symbol

import (
    "github.com/topple-psx/topple/compiler/ast"
)

// Registry tracks all exported symbols across modules
type Registry struct {
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
    RegisterModule(filePath string, symbols *ModuleSymbols)

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
// Collector extracts exportable symbols from AST
type Collector struct {
    currentFile string
    symbols     map[string]*Symbol
}

func (sc *Collector) CollectFromAST(module *ast.Module, filePath string) *ModuleSymbols {
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

func (sc *Collector) visitStatement(stmt ast.Stmt) {
    switch s := stmt.(type) {
    case *ast.ViewStmt:
        sc.addSymbol(s.Name, SymbolView, s)
    case *ast.Function:
        sc.addSymbol(s.Name, SymbolFunction, s)
    case *ast.Class:
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

func (sc *Collector) addSymbol(name string, typ SymbolType, node ast.Node) {
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

The `DependencyGraph` struct exports the following methods:

```go
// Add a file to the graph
func (g *DependencyGraph) AddFile(filePath string, module *ast.Module) error

// Add a dependency edge (from depends on to)
func (g *DependencyGraph) AddDependency(from, to string) error

// Get compilation order (topological sort)
func (g *DependencyGraph) GetCompilationOrder() ([]string, error)

// Detect circular dependencies
func (g *DependencyGraph) DetectCycles() ([][]string, error)

// Get direct dependencies of a file
func (g *DependencyGraph) GetDependencies(filePath string) []string

// Get reverse dependencies (files that import this file)
func (g *DependencyGraph) GetDependents(filePath string) []string

// Check if a file is in the graph
func (g *DependencyGraph) HasFile(filePath string) bool

// Get the FileNode for a given path
func (g *DependencyGraph) GetFileNode(filePath string) (*FileNode, bool)

// Get all file paths in the graph
func (g *DependencyGraph) GetAllFiles() []string

// Get count of files in the graph
func (g *DependencyGraph) FileCount() int

// Clear all files and dependencies
func (g *DependencyGraph) Clear()
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
        // Cycle detected - use cycle detection to get details
        cycles, _ := g.DetectCycles()
        return nil, NewCycleError(cycles)
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
        return cycles, NewCycleError(cycles)
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
            }
        }
        sb.WriteString("\n")
    }

    return sb.String()
}
```

---

### 4. Resolver Integration

> **Note**: The code examples in this section are **simplified illustrations** of the import resolution concepts. The actual implementation in `compiler/resolver/` uses different method names and internal structures. Refer to the source code for the precise API.

#### Import Statement Processing (Illustrative)

```go
// Simplified conceptual example - actual implementation differs
// See compiler/resolver/ for the real API

func (r *Resolver) VisitImportStmt(i *ast.ImportStmt) ast.Visitor {
    for _, alias := range i.Names {
        // Resolve module path using ModuleResolver
        modulePath, err := r.ModuleResolver.ResolveAbsolute(
            r.ctx,
            alias.DottedName.String(),
        )
        if err != nil {
            r.addError(i, err.Error())
            continue
        }

        // Get module symbols from SymbolRegistry
        moduleSymbols, err := r.SymbolRegistry.GetModuleSymbols(modulePath)
        if err != nil {
            r.addError(i, fmt.Sprintf("module '%s' not found", modulePath))
            continue
        }

        // Create binding for imported module name
        name := alias.AsName
        if name == "" {
            // Use first component of dotted name
            name = alias.DottedName.Names[0]
        }

        // Define the imported name in current scope (simplified)
        r.defineImported(name, modulePath, i)
    }

    return r
}

func (r *Resolver) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor {
    // Resolve module path
    var modulePath string
    var err error

    if i.DotCount > 0 {
        // Relative import
        modulePath, err = r.ModuleResolver.ResolveRelative(
            r.ctx,
            i.DotCount,
            i.DottedName.String(),
            r.currentFile,
        )
    } else {
        // Absolute import
        modulePath, err = r.ModuleResolver.ResolveAbsolute(
            r.ctx,
            i.DottedName.String(),
        )
    }

    if err != nil {
        r.addError(i, err.Error())
        return r
    }

    // Get module symbols from registry
    moduleSymbols, err := r.SymbolRegistry.GetModuleSymbols(modulePath)
    if err != nil {
        r.addError(i, fmt.Sprintf("module '%s' not found", modulePath))
        return r
    }

    // Handle wildcard vs specific imports
    if i.IsWildcard {
        // Import all public symbols
        symbols := moduleSymbols.GetPublicSymbols()
        for _, symbol := range symbols {
            r.defineImported(symbol.Name, modulePath, i)
        }
    } else {
        // Import specific names
        for _, name := range i.Names {
            symbol, exists := moduleSymbols.LookupSymbol(name.Name)
            if !exists {
                r.addError(i, fmt.Sprintf(
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

            r.defineImported(importName, modulePath, i)
        }
    }

    return r
}
```

#### Import Variable Tracking (Conceptual)

```go
// Simplified illustration of how imports are tracked
// Actual implementation in compiler/resolver/ differs

func (r *Resolver) defineImported(
    name string,
    modulePath string,
    node ast.Node,
) {
    // Create a variable to track the imported name
    variable := &Variable{
        Name:           name,
        DefinitionDepth: r.ScopeChain.GetDepth(),
        State:          Defined,
        IsImported:     true,
        ImportSource:   modulePath,  // Actual field name in codebase
    }

    // Store in the current scope
    // (Actual implementation uses ScopeChain and different internal structures)
    r.ScopeChain.Define(name, variable)
}
```

#### Variable Struct with Import Tracking

```go
// From compiler/resolver/types.go (actual fields)

type Variable struct {
    Name            string        // Variable name
    DefinitionDepth int           // Scope depth where defined
    State           VariableState // Current state

    // Scope behavior flags
    IsParameter     bool   // Function/view parameter
    IsGlobal        bool   // Declared with 'global'
    IsNonlocal      bool   // Declared with 'nonlocal'
    IsImported      bool   // Bound by import statement
    ImportSource    string // File path of imported module (if IsImported)
    IsViewParameter bool   // View parameter
    IsExceptionVar  bool   // Exception handler variable
    IsUsed          bool   // Has been referenced

    // Usage tracking
    FirstDefSpan  lexer.Span // Where first defined
    FirstUseSpan  lexer.Span // Where first used
    UsedBeforeDef bool       // Late binding detection
}
```

---

### 5. Multi-File Compiler

#### Multi-File Compiler Structure

```go
// From compiler/multifile.go (actual implementation)

// MultiFileCompiler compiles multiple interdependent PSX files
type MultiFileCompiler struct {
    logger         *slog.Logger
    fs             filesystem.FileSystem
    moduleResolver *module.StandardResolver   // Package: compiler/module
    symbolRegistry *symbol.Registry           // Package: compiler/symbol
    depGraph       *depgraph.DependencyGraph
}

// MultiFileOptions configures multi-file compilation
type MultiFileOptions struct {
    RootDir     string   // Project root for module resolution
    Files       []string // Explicit file list (absolute paths)
    SearchPaths []string // Additional search paths for imports
}

// MultiFileOutput contains the results of multi-file compilation
type MultiFileOutput struct {
    CompiledFiles map[string][]byte         // filepath -> generated Python code
    Registry      *symbol.Registry          // Symbol registry with all exports
    Graph         *depgraph.DependencyGraph // Dependency graph
    Errors        []*CompilationError       // All compilation errors
}

// CompilationError represents an error during multi-file compilation
type CompilationError struct {
    File    string // File where error occurred
    Stage   string // Compilation stage: "parse", "resolve", "transform", "codegen"
    Message string // Error message
    Details error  // Underlying error
}

func NewMultiFileCompiler(logger *slog.Logger) *MultiFileCompiler {
    return &MultiFileCompiler{
        logger:         logger,
        fs:             nil, // Initialized in CompileProject
        moduleResolver: nil, // Initialized in CompileProject
        symbolRegistry: symbol.NewRegistry(),
        depGraph:       depgraph.NewGraph(),
    }
}
```

#### Compilation Pipeline (Simplified Illustration)

> **Note**: This is a **simplified conceptual overview**. The actual implementation in `compiler/multifile.go` has more stages and error handling. See source code for details.

```go
// Actual signature from compiler/multifile.go
func (c *MultiFileCompiler) CompileProject(
    ctx context.Context,
    opts MultiFileOptions,
) (*MultiFileOutput, error) {
    output := &MultiFileOutput{
        CompiledFiles: make(map[string][]byte),
        Registry:      c.symbolRegistry,
        Graph:         c.depGraph,
        Errors:        []*CompilationError{},
    }

    // Stage 1: Initialize filesystem and module resolver
    c.fs = filesystem.NewFileSystem(c.logger)
    c.moduleResolver = module.NewResolver(module.Config{
        RootDir:     opts.RootDir,
        SearchPaths: opts.SearchPaths,
        FileSystem:  c.fs,
    })

    // Stage 2: Parse all files to AST
    astMap, parseErrs := c.parseAllFiles(ctx, opts.Files)
    if len(parseErrs) > 0 {
        output.Errors = append(output.Errors, parseErrs...)
        return output, fmt.Errorf("parsing failed with %d errors", len(parseErrs))
    }

    // Stage 3: Build dependency graph
    graphErrs := c.buildDependencyGraph(ctx, astMap)
    if len(graphErrs) > 0 {
        output.Errors = append(output.Errors, graphErrs...)
        return output, fmt.Errorf("dependency graph failed with %d errors", len(graphErrs))
    }

    // Stage 4: Get compilation order (topological sort)
    compilationOrder, err := c.depGraph.GetCompilationOrder()
    if err != nil {
        return output, fmt.Errorf("circular dependency detected: %w", err)
    }

    // Stage 5: Collect symbols from all files
    c.collectSymbols(ctx, astMap, compilationOrder)

    // Stage 6: Resolve and generate code for each file
    compileErrs := c.resolveAndGenerate(ctx, astMap, compilationOrder, output.CompiledFiles)
    if len(compileErrs) > 0 {
        output.Errors = append(output.Errors, compileErrs...)
    }

    if len(output.Errors) > 0 {
        return output, fmt.Errorf("compilation completed with %d errors", len(output.Errors))
    }

    return output, nil
}
```

---

## Integration Points

### CLI Integration

The command-line interface in `cmd/compile.go` handles:

1. **Argument parsing**: Input file/directory and optional output directory
2. **File discovery**: Recursively finding `.psx` files if needed
3. **Compilation mode selection**:
   - Single-file mode for individual `.psx` files
   - Multi-file mode for directories (uses `MultiFileCompiler`)
4. **Output writing**: Writing compiled Python files to appropriate locations
5. **Error reporting**: Displaying compilation errors with context

Key functions:
```go
// From cmd/compile.go (simplified)

// CompileCmd defines the compile command structure
type CompileCmd struct {
    Input  string `arg:"" required:""`
    Output string `arg:"" optional:""`
}

// compileMultiFile handles directory compilation with imports
func compileMultiFile(files []string, rootDir, outputDir string, log *slog.Logger, ctx context.Context) error {
    multiCompiler := compiler.NewMultiFileCompiler(log)

    opts := compiler.MultiFileOptions{
        RootDir:     rootDir,
        Files:       files,
        SearchPaths: []string{},
    }

    output, err := multiCompiler.CompileProject(ctx, opts)
    // ... handle errors and write output files
}
```

See `cmd/compile.go` for the complete implementation.

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
// CachedResolver wraps StandardResolver with caching
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
