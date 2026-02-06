# Parser Bug Status

**Last Verified**: 2026-02-06

This file tracks parser bugs that are still relevant to the current codebase.

## Open Bugs

### 1. Async view statement support is missing
- **Status**: Open
- **Severity**: High
- **Input**:
  ```python
  async view AsyncComponent():
      <div>Hello</div>
  ```
- **Current behavior**: parser reports `at 'async': unexpected token`
- **Expected behavior**: parser accepts `async view` and marks the resulting view node as async
- **Impact**: async view declarations cannot be used
- **Likely implementation points**:
  - `compiler/parser/statements.go`: add async+view branch in statement dispatch
  - `compiler/parser/view.go`: support async flag on parsed `ViewStmt`

## Closed / Outdated Entries

The previous version of this file listed many bugs that no longer reproduce in the current suite. These now pass in current tests, including:

- Operator precedence and chained comparisons
- Multiple-star validation in for-targets
- F-string equality expressions
- Function parameter pattern parsing
- Relative imports and import edge cases
- Generator expressions in call/if contexts
- Target parsing edge cases and complex nesting
- Try clause ordering validation

## Notes on Intentional Behavior

- `except 123:` remains parse-time valid and is treated as a runtime concern, matching current test expectations.
- Grouped starred expression handling follows current tests where `(*args)` is invalid unless tuple syntax is used (for example `(*args,)`).

## Current Health Snapshot

- `go test ./compiler/parser` passes.
- No currently failing parser tests in this repository snapshot.
