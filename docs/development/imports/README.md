# Import System Documentation

This directory contains the complete planning and design documentation for implementing cross-file import support in the Topple compiler.

## Documents

### 1. [Implementation Plan](./implementation_plan.md)
**High-level roadmap for the full implementation**

- Overview and goals
- Current state analysis (what works, what's missing)
- Implementation phases (6 phases over 3-4 weeks)
- Key design decisions required
- Testing strategy
- Risk assessment
- Success metrics

**Start here** for understanding the project scope and timeline.

---

### 2. [Architecture](./architecture.md)
**Detailed technical architecture and data structures**

- Component designs (Module Resolver, Symbol Registry, Dependency Graph)
- Algorithms (topological sort, cycle detection)
- Data structures and interfaces
- Resolver integration details
- Multi-file compiler pipeline
- Performance considerations
- Testing infrastructure

**Reference this** when implementing each component.

---

### 3. [Examples](./examples.md)
**Concrete examples and use cases**

- Basic import scenarios (views, functions, classes)
- Relative imports and packages
- Wildcard imports and aliases
- Multi-level dependencies
- Error cases with expected messages
- Integration tests structure
- Edge cases and migration examples

**Use this** for writing tests and validating behavior.

---

## Quick Reference

### Implementation Phases

| Phase | Focus | Duration |
|-------|-------|----------|
| 1 | Module Resolution | 3-5 days |
| 2 | Symbol Registry | 3-4 days |
| 3 | Resolver Integration | 5-7 days |
| 4 | Dependency Graph | 3-5 days |
| 5 | Multi-File Pipeline | 5-7 days |
| 6 | Testing & Polish | 4-5 days |

### Key Components

```
┌─────────────────────────────────────────┐
│        Multi-File Compiler              │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │   Module     │  │   Symbol     │   │
│  │   Resolver   │  │   Registry   │   │
│  └──────────────┘  └──────────────┘   │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │  Dependency  │  │   Resolver   │   │
│  │    Graph     │  │ Integration  │   │
│  └──────────────┘  └──────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

### New Packages

- `compiler/moduleresolver/` - Import path → file path resolution
- `compiler/symbolregistry/` - Cross-file symbol tracking
- `compiler/depgraph/` - Dependency graph and ordering

### Modified Components

- `compiler/resolver/visitor.go` - Import statement visitors (currently empty)
- `compiler/resolver/types.go` - Enhanced with import metadata
- `compiler/compiler.go` - Multi-file compilation interface

### Testing

```bash
# Run import-specific tests
mise run test-golden-category CATEGORY=imports

# Update golden files
UPDATE_GOLDEN=1 mise run test-golden-category CATEGORY=imports

# Run all tests
mise run test
```

---

## Current Status

### ✅ Complete
- Import parsing (all Python import syntax supported)
- AST nodes for imports
- Filesystem infrastructure
- Scope chain system

### ⏳ In Progress
- _No active work yet - planning phase_

### 📋 To Do
- Module resolver implementation
- Symbol registry implementation
- Resolver integration
- Dependency graph
- Multi-file compiler
- Comprehensive testing

---

## Design Decisions Summary

### Decided

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Compilation Flow | Two-phase (parse all → compile ordered) | Enables full symbol visibility |
| Cycle Handling | Forbid circular imports | Simpler implementation, better architecture |
| Package Structure | Directories as packages | Matches Python 3 namespace packages |

### To Be Decided

| Decision | Options | Impact |
|----------|---------|--------|
| Symbol Visibility | All top-level vs explicit `__all__` | Affects wildcard imports |
| Search Paths | CWD only vs configurable | Package management support |
| Incremental Compilation | v1 or later | Performance vs complexity |

---

## Code Locations

### Existing (Already Works)

| Component | File | Status |
|-----------|------|--------|
| Import Parsing | `compiler/parser/imports.go` | ✅ Complete (320 lines) |
| Import Tests | `compiler/parser/imports_test.go` | ✅ Complete (665 lines) |
| AST Nodes | `compiler/ast/import_*.go` | ✅ Complete |
| Filesystem Utils | `internal/filesystem/filesystem.go` | ✅ Complete (524 lines) |
| Code Generation | `compiler/codegen/statements.go:118-151` | ✅ Pass-through only |

### To Be Created

| Component | File | Status |
|-----------|------|--------|
| Module Resolver | `compiler/moduleresolver/resolver.go` | ⏳ Phase 1 |
| Symbol Registry | `compiler/symbolregistry/registry.go` | ⏳ Phase 2 |
| Dependency Graph | `compiler/depgraph/graph.go` | ⏳ Phase 4 |
| Multi-File Compiler | `compiler/multifile.go` | ⏳ Phase 5 |

### To Be Modified

| Component | File | Current State |
|-----------|------|---------------|
| Resolver Visitors | `compiler/resolver/visitor.go:582-583` | Empty stubs |
| Variable Metadata | `compiler/resolver/types.go` | Has `IsImported` field |
| Main Resolver | `compiler/resolver/resolver.go` | Single-file only |
| Compiler Interface | `compiler/compiler.go` | Single-file API |
| CLI | `cmd/compile.go` | No multi-file support |

---

## Getting Started

### For Implementation

1. **Read the Implementation Plan** to understand scope and phases
2. **Review Architecture** for component designs
3. **Check Examples** for expected behavior
4. **Start with Phase 1** (Module Resolver)
5. **Write tests first** before implementing

### For Review

1. **Implementation Plan** - Does the roadmap make sense?
2. **Architecture** - Are the components well-designed?
3. **Examples** - Do these cover the important use cases?
4. **Design Decisions** - What choices should we make?

### For Testing

1. **Create test files** in `compiler/testdata/input/imports/`
2. **Follow golden file pattern** from existing tests
3. **Run with UPDATE_GOLDEN=1** to generate expected output
4. **Verify generated Python** is correct

---

## Related Documentation

- [Resolver Design](../resolver_design.md) - Current resolver architecture
- [Architecture](../../architecture.md) - Overall compiler architecture
- [Golden Tests](../../golden_tests.md) - Testing methodology
- [AST Reference](../ast_reference.md) - AST node definitions

---

## Next Steps

1. ✅ Complete planning documentation (this document)
2. Create GitHub issue to track implementation
3. Set up initial test structure
4. Begin Phase 1: Module Resolver
5. Iterate through phases with continuous testing

---

## Questions or Feedback

If you have questions or suggestions about this design:

1. Review the relevant document (plan, architecture, or examples)
2. Check if it's a design decision that needs to be made
3. Consider the impact on other components
4. Open an issue or discussion

---

**Status**: Planning Complete ✅
**Next**: Ready for implementation
**Target**: Import support in 3-4 weeks

---

*Last Updated: 2025-11-15*
