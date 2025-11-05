# Topple Known Issues

**Last Updated**: 2025-11-05

## Active Issues (2)

### 🔴 High Priority

1. **[Issue #67: Parser: Add async view statement support](https://github.com/fjvillamarin/topple/issues/67)**
   - **File**: [parser-bug-05-async-view-support.md](parser-bug-05-async-view-support.md)
   - **Status**: Active, needs fix
   - **Impact**: Blocks async view functionality
   - **Tests Failing**: 2
   - **Recommendation**: **Fix this first**

### 🟡 Medium Priority

2. **[Issue #68: Parser: Fix complex target parsing edge cases](https://github.com/fjvillamarin/topple/issues/68)**
   - **File**: [parser-bug-07-target-parsing.md](parser-bug-07-target-parsing.md)
   - **Status**: Mostly fixed, 3 edge cases remaining
   - **Impact**: Complex target parsing edge cases only
   - **Tests Failing**: 10 (all edge cases)
   - **Note**: Basic target parsing works fine

---

## Archived Issues

### Fixed (10 bugs) ✅
See [archive/fixed/](archive/fixed/) for:
- All codegen bugs (3) - empty set, default params, multistmt
- Operator precedence
- E2E compilation
- Multiple starred expressions
- Relative imports
- Try clause ordering
- Function parameters
- Generator expressions

### Deferred (1 bug) 🟢
See [archive/deferred/](archive/deferred/) for:
- Incomplete "not in" validation (low priority edge case)

---

## Summary Statistics

- **Total documented bugs**: 23
- **Fixed**: 17 (74%)
- **Active**: 2 (9%)
- **Deferred**: 1 (4%)
- **Test pass rate**: 94%

---

## See Also

- [BUG_STATUS_REPORT.md](../BUG_STATUS_REPORT.md) - Full audit of all bugs
- [docs/development/codegen_bugs.md](../docs/development/codegen_bugs.md) - Codegen bug documentation
- [docs/development/parser_bugs.md](../docs/development/parser_bugs.md) - Parser bug documentation
