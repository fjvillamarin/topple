# 🐛 Parser Bug: Incomplete "not in" Expression Accepted as Valid

## Summary
The parser incorrectly accepts incomplete "not in" expressions like `x not` as valid syntax, treating the trailing `not` as a unary NOT operator. This should be a syntax error as it represents an incomplete expression.

## Priority: **LOW** 🟢
This is an edge case error handling issue that affects parser robustness but is unlikely to occur in real code.

## Bug Details

### Location
- **File**: `compiler/parser/comparisons_test.go:284`
- **Test**: `TestComparisonErrors/incomplete_not_in`
- **Parser Files**: `compiler/parser/expressions.go` (inversion function), `compiler/lexer/scanner.go`

### Current Behavior
```python
# Input
x not

# Currently: Parser accepts this, treating it as:
# - Expression: x
# - Followed by: not (waiting for operand)

# Expected: SyntaxError - incomplete expression
```

### Python Reference
In Python, a trailing `not` without an operand is invalid:
```python
>>> x not
  File "<stdin>", line 1
    x not
        ^
SyntaxError: invalid syntax
```

## Root Cause Analysis

The issue occurs because of how the parser handles expressions:

1. **Scanner behavior**: The scanner correctly combines `not in` into a `NotIn` token when they appear together (lines 167-182 in scanner.go)
2. **Parser behavior**: When parsing `x not`:
   - `x` is parsed as a complete expression
   - The parser returns `x` successfully
   - The trailing `not` token is left unconsumed
   - In statement parsing, this might be caught as an unexpected token
   - But in expression parsing alone, it's not detected

The problem is in the expression parsing chain. When `comparison()` is called:
- It calls `bitwiseOr()` which eventually parses `x`
- It returns `x` as a valid expression
- The `not` token remains unconsumed

## Action Plan

### Option 1: Context-Aware Parsing (Recommended)
This is likely a test expectation issue. In a full parsing context, the trailing `not` would be caught as an unexpected token when trying to parse the next statement or expression.

### Option 2: Lookahead Validation
Add validation in the comparison parser to detect incomplete "not in" patterns:

```go
// In comparison() or a helper function
func (p *Parser) checkIncompleteNotIn() error {
    if p.check(lexer.Not) && !p.checkNext(lexer.In) {
        return p.error(p.peek(), "incomplete 'not in' expression")
    }
    return nil
}
```

### Option 3: Scanner-Level Detection
Modify the scanner to detect and report incomplete "not in" patterns, though this violates separation of concerns.

## Implementation Considerations

### Why This Might Not Be a Bug

Looking at the test more carefully:
- The test calls `parser.comparison()` directly
- In a real parsing scenario, `statement()` would be called
- The unconsumed `not` token would cause an error at the statement level

This suggests the test might be testing at the wrong level. The parser is correct to parse `x` as a valid expression and leave `not` for subsequent parsing.

### The Real Issue

The test expects `comparison()` to fail, but `comparison()` correctly parses `x` and returns. The `not` token would be handled (and error) at a higher parsing level.

## Recommended Fix

### Option 1: Adjust Test Expectations (Preferred)
The test should parse at the statement level, not the expression level:

```go
// Instead of:
_, err := parser.comparison()

// Should be:
_, err := parser.statement()  // or parser.Parse()
```

### Option 2: Add Expression Validation
If we want `comparison()` to catch this, we need to add validation:

```go
func (p *Parser) comparison() (ast.Expr, error) {
    left, err := p.bitwiseOr()
    if err != nil {
        return nil, err
    }
    
    // Check for incomplete "not in"
    if p.isAtEnd() && p.previous().Type == lexer.Not {
        return nil, p.error(p.previous(), "incomplete 'not in' expression")
    }
    
    // ... rest of comparison parsing
}
```

## Test Cases to Consider

```python
# Should all be syntax errors:
x not
x is not
x not and y
if x not: pass

# Valid cases:
x not in y
x is not None
not x
if not x: pass
```

## Expected Impact

### If Test Adjustment:
- No parser changes needed
- Test will properly validate statement-level parsing
- Maintains clean separation of concerns

### If Parser Enhancement:
- Better error messages for incomplete expressions
- Might catch other edge cases
- Could add complexity to expression parsing

## Recommendation

This appears to be a **test issue** rather than a parser bug. The parser correctly handles the expression `x` and leaves `not` for subsequent parsing, which is the expected behavior for recursive descent parsers. The test should be adjusted to parse at the appropriate level (statement or module) where the trailing `not` would be properly detected as a syntax error.

## References
- Python's parser would catch this at the statement level
- Recursive descent parsers typically handle such errors at higher parsing levels
- Similar to how `x +` would parse `x` successfully at the expression level