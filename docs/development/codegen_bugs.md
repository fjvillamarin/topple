# Codegen Bug Status

**Last Verified**: 2026-02-06

This file tracks current codegen bugs and quality gaps.

## Open Items

### 1. Newline-at-EOF consistency is not enforced
- **Status**: Open (style/consistency)
- **Severity**: Low
- **Details**:
  - Generated golden outputs are accepted with or without trailing newline.
  - Current test comparator trims trailing newlines, so this is not a test failure.
- **Decision needed**:
  - Either enforce final newline in all generated files and golden outputs, or
  - Remove the strict newline requirement from style guidance.

## Closed / No Longer Applicable

### MultiStmt output formatting
- **Status**: Closed by architecture
- **Details**: `MultiStmt` is expected to be unwrapped before codegen. `VisitMultiStmt` panics if reached.

### Empty set emitted as `{}`
- **Status**: Fixed
- **Verification**: `TestCodeGeneration/collections/empty_set` passes and output is `set()`.

### String literals missing quotes
- **Status**: Fixed
- **Verification**: `TestCodeGeneration/literals/string` passes and output is quoted.

### Default parameter spacing (`name = "x"`)
- **Status**: Fixed
- **Verification**: `TestCodeGeneration/functions/function_with_defaults` passes with `name="World"`.

## Current Health Snapshot

- `go test ./compiler/codegen` passes.
- No active codegen correctness bug is currently failing tests.
