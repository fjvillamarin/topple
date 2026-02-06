# Topple Bug Status Report
**Generated**: 2026-02-06
**Verified Against**: Current repository test suite and CLI parser behavior

## Executive Summary

- **Current test status**: `go test ./...` passes
- **Open parser bug**: 1
- **Open codegen correctness bugs**: 0
- **Open quality consistency item**: 1 (newline-at-EOF policy mismatch)

## Confirmed Open Issues

### 1. Async view declarations are not parsed
- **Area**: Parser
- **Status**: Open
- **Repro**:
  - Create a `.psx` file with `async view AsyncComponent():`
  - Run `topple parse <file>`
  - Result: `at 'async': unexpected token`
- **Code pointers**:
  - `compiler/parser/statements.go` handles `async for`, `async with`, and `async def`, but not `async view`
  - `compiler/parser/view.go` currently sets `IsAsync: false`
- **Impact**: Async view syntax is unsupported

### 2. Newline-at-EOF style consistency is not enforced in generated code
- **Area**: Codegen quality policy
- **Status**: Open (style/consistency)
- **Observed**:
  - Many generated and expected golden files do not end with `\n`
  - Test comparison normalizes away trailing newline differences
- **Impact**: Inconsistent file endings relative to stated style guidance

## Closed / Outdated Items

The previously tracked parser and codegen bug lists included many items that are now fixed or no longer applicable to the current architecture. Notable examples verified as passing in current tests include:

- Operator precedence (`TestBitwiseOr/mixed_with_comparison`)
- Star-target validation (`TestStarTargets/multiple_stars_in_same_level`)
- F-string equality parsing (`TestFStringDebugExpressions/equals_in_expression`)
- Function parameter pattern parsing (`TestFunctionParameterPatterns/complete_parameter_pattern`)
- Relative imports and import edge cases (`TestRelativeImports/*`, `TestImportEdgeCases/*`)
- Target parsing edge cases (`TestTargetParsing/*`, `TestTargetEdgeCases/*`, `TestTargetComplexNesting/*`)
- Try clause ordering (`TestTryCombinations/finally_before_except`)
- Codegen empty set output (`TestCodeGeneration/collections/empty_set`)
- Codegen string quoting (`TestCodeGeneration/literals/string`)
- Codegen default parameter formatting (`TestCodeGeneration/functions/function_with_defaults`)

## Recommended Next Actions

1. Implement parser support for `async view` and add dedicated parser tests.
2. Decide whether newline-at-EOF is required; if yes, enforce it in codegen and golden tests, otherwise remove the policy text from docs.
3. Keep this report as a living status file and avoid duplicating stale bug inventories across multiple docs.
