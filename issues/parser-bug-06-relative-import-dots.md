# 🐛 Parser Bug: Incorrect Relative Import Dot Counting

## Summary
The parser incorrectly counts dots in relative imports. When parsing `from .. import module`, it counts only 1 dot instead of 2. This is because the scanner creates a single `Dot` token with lexeme ".." for consecutive dots, but the parser only increments the counter by 1 per token, not by the actual number of dots in the lexeme.

## Priority: **HIGH** 🔴
Relative imports are a fundamental Python feature, and incorrect dot counting will cause imports to resolve to the wrong package level.

## Bug Details

### Location
- **File**: `compiler/parser/imports_test.go:81`
- **Test**: Multiple relative import tests
- **Parser File**: `compiler/parser/imports.go:61-69`
- **Scanner File**: `compiler/lexer/scanner.go:459-467`

### Current Behavior
```python
# Input
from .. import module

# Parser sees:
# Token: Dot with lexeme ".."
# Counts as: 1 dot (incorrect)
# Should count as: 2 dots
```

### Scanner Behavior (Correct)
The scanner correctly tokenizes dots:
- `.` → `Dot` token with lexeme "."
- `..` → `Dot` token with lexeme ".." (single token, not two)
- `...` → `Ellipsis` token with lexeme "..."
- `....` → `Ellipsis` token + `Dot` token

### Parser Behavior (Incorrect)
```go
// Current code in imports.go lines 62-68
for p.match(lexer.Dot) || p.match(lexer.Ellipsis) {
    if p.previous().Type == lexer.Dot {
        dotCount += 1  // BUG: Always adds 1, regardless of lexeme
    } else {
        dotCount += 3  // Correct for Ellipsis
    }
}
```

## Root Cause Analysis

The issue is a mismatch between scanner and parser expectations:

1. **Scanner Design**: The scanner creates a single token for consecutive dots (up to 2), with the full lexeme preserved. This is actually good design as it prevents ambiguity.

2. **Parser Assumption**: The parser assumes each `Dot` token represents exactly one dot, but doesn't check the lexeme length.

3. **The Bug**: When the parser sees a `Dot` token, it always adds 1 to `dotCount`, even if the token's lexeme is ".." (which should add 2).

## Action Plan

### Fix the Parser's Dot Counting
Update the parser to count the actual dots in the lexeme:

```go
// Fixed version
for p.match(lexer.Dot) || p.match(lexer.Ellipsis) {
    prev := p.previous()
    if prev.Type == lexer.Dot {
        // Count the actual dots in the lexeme
        dotCount += len(prev.Lexeme)
    } else { // Ellipsis
        dotCount += 3
    }
}
```

### Alternative Fix (More Explicit)
Be more explicit about what we're counting:

```go
for p.match(lexer.Dot) || p.match(lexer.Ellipsis) {
    prev := p.previous()
    switch prev.Type {
    case lexer.Dot:
        // Handle both "." and ".."
        switch prev.Lexeme {
        case ".":
            dotCount += 1
        case "..":
            dotCount += 2
        default:
            return nil, p.error(prev, "unexpected dot sequence")
        }
    case lexer.Ellipsis:
        dotCount += 3
    }
}
```

## Test Cases to Verify

```python
# Single dot - current package
from . import module           # dotCount = 1 ✓

# Double dot - parent package  
from .. import module          # dotCount = 2 (currently 1 ❌)

# Triple dot - grandparent
from ... import module         # dotCount = 3 ✓

# Four dots - great-grandparent
from .... import module        # dotCount = 4 (needs verification)

# Mixed with module paths
from ..utils import helper     # dotCount = 2, module = "utils"
from ...lib.core import func   # dotCount = 3, module = "lib.core"
```

## Implementation Steps

1. **Update Parser**: Modify the dot counting logic in `imports.go`
2. **Add Tests**: Ensure all dot counting scenarios are tested
3. **Verify Edge Cases**: Test sequences like "...." (4 dots) which tokenize as Ellipsis + Dot
4. **Update Documentation**: Document the expected behavior for relative imports

## Expected Impact

### Fixes
- Relative imports with `..` will correctly resolve to parent package
- Test cases for relative imports will pass
- Python-compatible relative import behavior

### Risks
- Need to ensure the fix handles all possible dot sequences
- Edge cases like "...." need careful handling
- May affect existing code that worked around this bug

## Additional Notes

The scanner's behavior of creating single tokens for ".." is actually good design because:
1. It prevents ambiguity in parsing
2. It maintains the exact input sequence
3. It's consistent with how other multi-character operators are handled

The fix is straightforward - just count the characters in the lexeme rather than assuming each token equals one dot.

## References
- Python relative imports: [PEP 328](https://www.python.org/dev/peps/pep-0328/)
- Scanner dot handling: `compiler/lexer/scanner.go:459-467`
- Import parser: `compiler/parser/imports.go:61-69`