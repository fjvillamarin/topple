# Import System Implementation Plan for Topple Compiler

## Overview

This document outlines the high-level plan for implementing cross-file import support in the Topple compiler. Import support will enable PSX files to import views, functions, classes, and variables from other PSX files, allowing modular application development.

**Status**: 📋 Planning Phase

**Target Completion**: 3-4 weeks

**Related Issues**: TBD

## Goals

1. **Module Resolution**: Translate import statements to actual `.psx` file paths ⏳
2. **Cross-File Symbol Resolution**: Resolve imported symbols across file boundaries ⏳
3. **Dependency Tracking**: Build and validate dependency graphs between files ⏳
4. **Compilation Ordering**: Compile files in correct topological order ⏳
5. **Error Detection**: Validate imports and detect circular dependencies ⏳
6. **Python Compatibility**: Support Python's full import semantics (absolute, relative, wildcards) ⏳

## Current State Analysis

### ✅ What Already Works

1. **Import Parsing (100% Complete)**
   - Location: `compiler/parser/imports.go` (320 lines)
   - Supports: `import x`, `from x import y`, relative imports, wildcards, aliases
   - Full test coverage: `compiler/parser/imports_test.go` (665 lines)
   - AST nodes: `ImportStmt`, `ImportFromStmt`, `ImportName`, `DottedName`

2. **Infrastructure Ready**
   - `Variable.IsImported` field exists (`compiler/resolver/types.go:28`)
   - Filesystem abstraction complete (`internal/filesystem/filesystem.go`)
   - Scope chain system supports extensibility
   - Error handling infrastructure in place

3. **Code Generation**
   - Import statements pass through to generated Python
   - Location: `compiler/codegen/statements.go:118-151`

### ❌ What's Missing

1. **Module Resolution Engine**
   - No translation from import paths to file paths
   - No package structure understanding
   - No relative import resolution

2. **Cross-File Symbol Registry**
   - No way to track exported symbols from files
   - No symbol lookup across files
   - No symbol origin tracking

3. **Resolver Integration**
   - Import visitor methods are empty stubs (`compiler/resolver/visitor.go:582-583`)
   - No cross-file name resolution in `ResolveName()`
   - Import statements don't create bindings

4. **Multi-File Compilation**
   - Compiler interface is single-file only (`compiler/compiler.go`)
   - No dependency graph construction
   - No compilation ordering
   - No circular dependency detection

## High-Level Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Multi-File Compiler                      │
│                                                             │
│  ┌──────────────┐      ┌──────────────┐                   │
│  │  File Set    │─────▶│  Module      │                   │
│  │  Collector   │      │  Resolver    │                   │
│  └──────────────┘      └──────┬───────┘                   │
│                               │                            │
│                               ▼                            │
│                        ┌──────────────┐                   │
│                        │  Dependency  │                   │
│                        │  Graph       │                   │
│                        └──────┬───────┘                   │
│                               │                            │
│                               ▼                            │
│  ┌──────────────┐      ┌──────────────┐                   │
│  │  Symbol      │◀─────│  Ordered     │                   │
│  │  Registry    │      │  Compiler    │                   │
│  └──────────────┘      └──────────────┘                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### New Components

#### 1. Module Resolver
**Purpose**: Translate import paths to filesystem paths

**Responsibilities**:
- Map import paths (`my_module`) to `.psx` files (`./my_module.psx`)
- Handle relative imports (`.`, `..`) relative to current file
- Support package structures (directories with `__init__.psx`)
- Validate import paths exist
- Provide clear error messages for missing modules

**Location**: `compiler/moduleresolver/` (new package)

#### 2. Symbol Registry
**Purpose**: Track exported symbols from all compiled files

**Responsibilities**:
- Collect top-level symbols from each file (views, functions, classes, variables)
- Provide symbol lookup by qualified name
- Track symbol origins (source file and location)
- Support wildcard import expansion
- Handle symbol visibility rules

**Location**: `compiler/symbolregistry/` (new package)

#### 3. Dependency Graph
**Purpose**: Track and order file dependencies

**Responsibilities**:
- Build directed graph from import statements
- Perform topological sort for compilation order
- Detect circular dependencies
- Track which files need recompilation

**Location**: `compiler/depgraph/` (new package)

#### 4. Multi-File Compiler
**Purpose**: Orchestrate compilation of multiple related files

**Responsibilities**:
- Accept multiple input files or directories
- Coordinate module resolution, symbol collection, and compilation
- Manage shared compilation context
- Produce all output files

**Location**: `compiler/compiler.go` (extend existing)

### Data Flow

```
Input Files (.psx)
       ↓
1. Parse All Files → AST Forest
       ↓
2. Extract Imports → Dependency Graph
       ↓
3. Topological Sort → Compilation Order
       ↓
4. For Each File (in order):
   a. Resolve Symbols (with registry)
   b. Transform Views
   c. Generate Code
   d. Export Symbols to Registry
       ↓
Output Files (.py)
```

## Implementation Phases

### Phase 1: Module Resolution Foundation (Week 1)

**Goal**: Implement the module resolution engine

**Tasks**:
1. Create `ModuleResolver` interface and implementation
2. Implement absolute import path resolution
3. Implement relative import path resolution (`.`, `..`, `...`)
4. Add package discovery (`__init__.psx` handling)
5. Create comprehensive test suite
6. Handle `.psx` extension automatically

**Deliverables**:
- `compiler/moduleresolver/resolver.go`
- `compiler/moduleresolver/resolver_test.go`
- Support for resolving `from module import x` to file paths

**Key Design Decisions**:
- Search path configuration (cwd, explicit paths)
- Package structure rules
- `__init__.psx` behavior
- Error messages for missing modules

**Success Criteria**:
- Can resolve absolute imports to files
- Can resolve relative imports to files
- Proper error messages for invalid imports
- All tests passing

---

### Phase 2: Symbol Registry (Week 1-2)

**Goal**: Build cross-file symbol tracking system

**Tasks**:
1. Create `SymbolRegistry` data structure
2. Implement symbol collection from AST
3. Add symbol lookup methods (by qualified name)
4. Implement wildcard import expansion
5. Track symbol metadata (location, type, visibility)
6. Add conflict detection (duplicate exports)

**Deliverables**:
- `compiler/symbolregistry/registry.go`
- `compiler/symbolregistry/registry_test.go`
- Symbol export mechanism

**Key Data Structures**:
```go
type SymbolRegistry struct {
    // Module name → exported symbols
    modules map[string]*ModuleSymbols
}

type ModuleSymbols struct {
    FilePath string
    Symbols  map[string]*Symbol
}

type Symbol struct {
    Name       string
    Type       SymbolType  // View, Function, Class, Variable
    Location   ast.Node
    Visibility Visibility  // Public, Private
}
```

**Success Criteria**:
- Can collect symbols from parsed files
- Can lookup symbols by module and name
- Wildcard imports return all public symbols
- Proper handling of symbol visibility

---

### Phase 3: Resolver Integration (Week 2)

**Goal**: Enable cross-file name resolution

**Tasks**:
1. Implement `VisitImportStmt()` in resolver
2. Implement `VisitImportFromStmt()` in resolver
3. Create import bindings in module scope
4. Extend `ResolveName()` to check imported symbols
5. Set `Variable.IsImported = true` for imports
6. Add import-related error detection
7. Update resolution table to include imports

**Deliverables**:
- Updated `compiler/resolver/visitor.go`
- Updated `compiler/resolver/resolver.go`
- Import resolution tests

**Key Changes**:
```go
func (v *ResolverVisitor) VisitImportFromStmt(i *ast.ImportFromStmt) {
    // 1. Resolve module path to file
    modulePath := v.moduleResolver.Resolve(i.DottedName)

    // 2. Get symbols from registry
    symbols := v.symbolRegistry.GetModuleSymbols(modulePath)

    // 3. Create bindings for imported names
    for _, name := range i.Names {
        symbol := symbols.Lookup(name.Name)
        v.defineImportedVariable(name, symbol)
    }
}
```

**Success Criteria**:
- Import statements create proper bindings
- Imported symbols resolve correctly
- `IsImported` flag set appropriately
- Undefined import errors reported
- Resolution table includes import info

---

### Phase 4: Dependency Graph (Week 2-3)

**Goal**: Track and order file dependencies

**Tasks**:
1. Create `DependencyGraph` data structure
2. Extract imports from AST to build graph
3. Implement topological sort algorithm
4. Add circular dependency detection
5. Generate clear error messages for cycles
6. Support incremental graph updates

**Deliverables**:
- `compiler/depgraph/graph.go`
- `compiler/depgraph/graph_test.go`
- Compilation ordering logic

**Key Data Structures**:
```go
type DependencyGraph struct {
    nodes map[string]*FileNode
    edges map[string][]string  // file → dependencies
}

type FileNode struct {
    FilePath string
    AST      *ast.Module
    Imports  []*Import
}

type Import struct {
    Statement  ast.Node
    TargetFile string
    IsRelative bool
}
```

**Algorithms**:
- Kahn's algorithm for topological sort
- DFS for cycle detection

**Success Criteria**:
- Correct compilation order computed
- Circular dependencies detected
- Clear error messages with cycle path
- Handles complex dependency patterns

---

### Phase 5: Multi-File Compilation Pipeline (Week 3)

**Goal**: Orchestrate multi-file compilation

**Tasks**:
1. Create `MultiFileCompiler` interface
2. Implement file collection (explicit + recursive)
3. Integrate module resolver, registry, and dep graph
4. Add two-phase compilation:
   - Phase 1: Parse all, collect symbols
   - Phase 2: Resolve and generate
5. Share compilation context across files
6. Update CLI to support multi-file mode

**Deliverables**:
- `compiler/multifile.go` (new)
- Updated `cmd/compile.go`
- Multi-file compilation API

**New Compiler Interface**:
```go
type MultiFileCompiler interface {
    CompileProject(ctx context.Context, opts CompileOptions) (*ProjectOutput, []error)
}

type CompileOptions struct {
    Files       []string  // Explicit file list
    Directories []string  // Directories to scan
    Recursive   bool
    SearchPaths []string  // Import search paths
}

type ProjectOutput struct {
    Files map[string][]byte  // filepath → generated code
    Registry *SymbolRegistry
    Graph *DependencyGraph
}
```

**Compilation Flow**:
```go
1. Collect all .psx files
2. Parse all files → AST forest
3. Build dependency graph
4. Topological sort → compilation order
5. For each file (ordered):
   a. Resolve (with registry context)
   b. Transform
   c. Generate
   d. Add symbols to registry
6. Return all outputs
```

**Success Criteria**:
- Can compile multiple interdependent files
- Symbols resolve across files
- Correct compilation order
- All files produce output
- No regressions in single-file mode

---

### Phase 6: Testing & Polish (Week 3-4)

**Goal**: Comprehensive testing and edge case handling

**Tasks**:
1. Add golden file tests for imports
2. Test wildcard imports (`from x import *`)
3. Test relative imports at various depths
4. Test circular dependency errors
5. Test missing import errors
6. Test symbol conflicts
7. Integration tests with view composition
8. Documentation updates
9. Error message improvements
10. Performance testing with large projects

**Test Categories**:
- `testdata/input/imports/absolute/` - Absolute imports
- `testdata/input/imports/relative/` - Relative imports
- `testdata/input/imports/wildcard/` - Wildcard imports
- `testdata/input/imports/circular/` - Circular dependency errors
- `testdata/input/imports/views/` - Importing views
- `testdata/input/imports/packages/` - Package structures

**Example Test Case**:
```
testdata/input/imports/views/
  ├── components.psx    # Exports MyButton view
  └── app.psx          # Imports MyButton

Expected:
  ├── components.py    # Generated with MyButton class
  └── app.py          # Generated with import from components
```

**Success Criteria**:
- 100+ import test cases passing
- All edge cases covered
- Clear, helpful error messages
- Documentation complete
- Performance acceptable (< 100ms per file)

---

## Key Design Decisions

### 1. Symbol Visibility Rules

**Decision Required**: Which symbols are exported?

**Options**:
- A. **All top-level names** (simplest)
- B. **Exclude underscore-prefixed** (`_private`)
- C. **Explicit `__all__`** (Python-standard)

**Recommendation**: Start with A, add B in future

**Rationale**: Simplicity for v1, matches Python default behavior

---

### 2. Package Structure

**Decision Required**: How are packages defined?

**Options**:
- A. **Directories are packages** (automatic)
- B. **Require `__init__.psx`** (explicit)
- C. **No package support** (only flat modules)

**Recommendation**: A with optional `__init__.psx`

**Rationale**: Matches Python 3 namespace packages, minimal ceremony

---

### 3. Circular Dependencies

**Decision Required**: How to handle circular imports?

**Options**:
- A. **Forbid entirely** (compile error)
- B. **Allow with restrictions** (runtime error if accessed during init)
- C. **Full Python compatibility** (complex)

**Recommendation**: A for v1

**Rationale**: Simpler implementation, encourages better architecture. Can relax later.

---

### 4. Compilation Modes

**Decision Required**: When to compile dependencies?

**Options**:
- A. **Always compile all dependencies** (eager)
- B. **Compile only changed files** (incremental)
- C. **Compile on-demand** (lazy)

**Recommendation**: A for v1, add B later

**Rationale**: Correctness first, optimization later

---

### 5. Import Path Resolution

**Decision Required**: Where to search for modules?

**Options**:
- A. **Only current directory**
- B. **Configurable search paths**
- C. **Python sys.path compatibility**

**Recommendation**: A for v1, add B in v2

**Rationale**: Keep it simple initially

---

### 6. Output Structure

**Decision Required**: How to organize generated files?

**Options**:
- A. **Mirror input structure** (`src/x.psx` → `out/x.py`)
- B. **Flatten to single directory**
- C. **User-configurable**

**Recommendation**: A

**Rationale**: Maintains project structure, supports packages

---

## Testing Strategy

### Unit Tests

**Per Component**:
- Module Resolver: 50+ test cases
- Symbol Registry: 40+ test cases
- Dependency Graph: 30+ test cases
- Resolver Integration: 60+ test cases

**Coverage Target**: 85%+

### Integration Tests

**Golden File Tests**:
- Organize by import type in `testdata/input/imports/`
- Each test: multiple `.psx` files → multiple `.py` files
- Use `test-golden-category CATEGORY=imports`

**Test Matrix**:
```
Import Type × Feature
├── Absolute × (View, Function, Class, Variable)
├── Relative × (., .., ...)
├── Wildcard × (*, specific names)
├── Alias × (as keyword)
└── Mixed × (multiple in one file)
```

### Error Tests

**Negative Cases**:
- Missing module
- Missing symbol
- Circular dependency
- Invalid relative import
- Symbol conflicts

### End-to-End Tests

**Real-World Scenarios**:
```
examples/multi_file_app/
  ├── components/
  │   ├── button.psx
  │   └── card.psx
  ├── layouts/
  │   └── main.psx
  └── app.psx
```

Compile and verify:
- All files compile
- Imports resolve
- Generated code runs
- Views compose correctly

---

## Risk Assessment & Mitigation

### Risk 1: Circular Dependencies
**Impact**: High - Could crash compiler or create invalid output
**Likelihood**: Medium - Common in real projects
**Mitigation**:
- Implement robust cycle detection (Kahn's algorithm)
- Clear error messages showing cycle path
- Early detection in dependency graph phase

### Risk 2: Symbol Conflicts
**Impact**: Medium - Ambiguous symbol resolution
**Likelihood**: Low - Good naming conventions help
**Mitigation**:
- Detect conflicts during symbol registration
- Report conflicts with file locations
- Consider qualified name support

### Risk 3: Performance at Scale
**Impact**: Medium - Slow compilation on large projects
**Likelihood**: Medium - As projects grow
**Mitigation**:
- Profile early with large test projects
- Implement caching in Phase 2
- Incremental compilation in future

### Risk 4: Python Compatibility
**Impact**: Low - Generated code might not match Python behavior
**Likelihood**: Low - Import semantics well-defined
**Mitigation**:
- Test against Python reference implementation
- Document any deviations
- Focus on common cases first

### Risk 5: Breaking Changes
**Impact**: High - Existing code might break
**Likelihood**: Low - Additive feature mostly
**Mitigation**:
- Maintain backward compatibility for single-file mode
- Comprehensive regression tests
- Feature flag for multi-file mode initially

---

## Success Metrics

### Functional Metrics
- ✅ Can import views from other files
- ✅ Can import functions, classes, variables
- ✅ Relative imports work correctly
- ✅ Circular dependencies detected
- ✅ Missing imports reported clearly
- ✅ All tests passing (100%)

### Quality Metrics
- ✅ Test coverage > 85%
- ✅ Zero regressions in existing tests
- ✅ Documentation complete
- ✅ Error messages actionable

### Performance Metrics
- ✅ < 100ms per file compilation
- ✅ < 500ms for 50-file project
- ✅ Memory usage reasonable (< 100MB for typical project)

---

## Future Enhancements (Post-v1)

### Phase 7: Incremental Compilation
- Cache compiled outputs
- Track file modifications
- Recompile only changed files + dependents
- Significant speedup for large projects

### Phase 8: Package Management
- Support for third-party PSX packages
- Package discovery mechanism
- Version management
- Distribution format

### Phase 9: Advanced Features
- Star import expansion (`from x import *`)
- `__all__` export control
- Lazy imports (reduce startup time)
- Import hooks for custom resolution

### Phase 10: IDE Support
- LSP integration for go-to-definition across files
- Auto-import suggestions
- Refactoring support (rename across files)

---

## Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| 1. Module Resolution | 3-5 days | Module resolver with tests |
| 2. Symbol Registry | 3-4 days | Symbol tracking system |
| 3. Resolver Integration | 5-7 days | Cross-file name resolution |
| 4. Dependency Graph | 3-5 days | Dep graph with cycle detection |
| 5. Multi-File Pipeline | 5-7 days | Complete multi-file compilation |
| 6. Testing & Polish | 4-5 days | Comprehensive tests, docs |
| **Total** | **3-4 weeks** | **Full import support** |

---

## References

### Related Documents
- `docs/development/resolver_design.md` - Current resolver architecture
- `docs/architecture.md` - Overall compiler architecture
- `docs/development/ast_reference.md` - AST node reference

### Code Locations
- `compiler/parser/imports.go` - Import parsing (complete)
- `compiler/resolver/types.go` - Variable metadata (has `IsImported`)
- `compiler/resolver/visitor.go:582-583` - Empty import visitors
- `compiler/codegen/statements.go:118-151` - Import code generation
- `internal/filesystem/filesystem.go` - Filesystem utilities

### Python References
- [PEP 328 - Imports: Multi-Line and Absolute/Relative](https://peps.python.org/pep-0328/)
- [Python Import System Documentation](https://docs.python.org/3/reference/import.html)
- [Python Module Search Path](https://docs.python.org/3/tutorial/modules.html#the-module-search-path)

---

## Approval & Sign-off

**Created**: 2025-11-15
**Author**: Claude Code
**Status**: Awaiting Review

**Next Steps**:
1. Review this plan with project stakeholders
2. Create GitHub issue for tracking
3. Begin Phase 1 implementation
4. Iterate based on learnings

---

*This plan is a living document and will be updated as implementation progresses.*
