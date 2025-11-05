# 🐛 Parser Bug: Missing Validation for Multiple Starred Expressions

## Summary
The parser incorrectly accepts multiple starred expressions at the same nesting level in assignment targets (e.g., `for *a, *b in items:`), which is invalid Python syntax. Python only allows one starred expression per assignment level.

## Priority: **MEDIUM** 🟡
This allows invalid Python syntax to pass through the parser, potentially causing runtime errors.

## Bug Details

### Location
- **File**: `compiler/parser/for_test.go:333`
- **Test**: `TestStarTargets/multiple_stars_in_same_level`
- **Parser Files**: `compiler/parser/targets.go`, `compiler/parser/for.go`

### Current Behavior
```python
# Input (should be invalid)
for *a, *b in items:
    pass

# Currently: Parser accepts this without error
# Expected: SyntaxError: multiple starred expressions in assignment
```

### Python Reference
Python's grammar restricts starred expressions in assignment targets:
```python
# Valid: Only one starred expression
a, *b, c = [1, 2, 3, 4, 5]  # a=1, b=[2,3,4], c=5

# Invalid: Multiple starred expressions
*a, *b = [1, 2, 3]  # SyntaxError: multiple starred expressions in assignment
```

## Root Cause Analysis

After examining the code, I found that:

1. **No validation in `starTargets()`**: The function in `compiler/parser/targets.go` (lines 265-307) simply collects star targets without checking if multiple starred expressions exist at the same level.

2. **No validation in `parseStarTargetSequence()`**: Similarly, this function (lines 312+) parses sequences without validating the starred expression count.

3. **Missing validation logic**: The parser needs to track whether a starred expression has already been encountered at the current nesting level and reject additional ones.

### Code Analysis

In `starTargets()`:
```go
// The function just appends targets without validation
for {
    target, err = p.starTarget()
    if err != nil {
        return nil, err
    }
    targets = append(targets, target)  // No check for multiple stars!
    
    if !p.match(lexer.Comma) {
        break
    }
    // ...
}
```

## Action Plan

### 1. **Add Starred Expression Validation**
Create a helper function to validate that only one starred expression exists at each nesting level:

```go
func validateStarredExpressions(targets []ast.Expr) error {
    starCount := 0
    for _, target := range targets {
        if _, isStar := target.(*ast.StarExpr); isStar {
            starCount++
            if starCount > 1 {
                return NewParseError(target.GetSpan(), 
                    "multiple starred expressions in assignment")
            }
        }
    }
    return nil
}
```

### 2. **Integrate Validation in Parser**
Add validation calls in:
- `starTargets()` - after collecting all targets
- `parseStarTargetSequence()` - after collecting sequence elements
- Any other places where star targets are collected

### 3. **Handle Nested Structures**
The validation needs to be aware of nesting levels:
- `(a, *b)` and `(*c, d)` in the same assignment is valid
- `*a, *b` at the same level is invalid

## Implementation Steps

1. **Create validation function**: Add `validateStarredExpressions()` to targets.go
2. **Add validation calls**: Insert validation after target collection
3. **Update tests**: Ensure the test expects an error
4. **Test edge cases**: Verify nested structures work correctly

## The Fix

```go
// In compiler/parser/targets.go, after line 305 in starTargets()
func (p *Parser) starTargets() ([]ast.Expr, error) {
    // ... existing code to collect targets ...
    
    // Add validation before returning
    if err := p.validateStarredTargets(targets); err != nil {
        return nil, err
    }
    
    return targets, nil
}

// New validation function
func (p *Parser) validateStarredTargets(targets []ast.Expr) error {
    starCount := 0
    var starToken lexer.Token
    
    for _, target := range targets {
        if star, isStar := target.(*ast.StarExpr); isStar {
            starCount++
            if starCount > 1 {
                return p.error(starToken, "multiple starred expressions in assignment")
            }
            // Track the token for error reporting
            starToken = p.Tokens[p.Current-1] // Approximate token position
        }
    }
    return nil
}
```

## Test Cases to Verify

```python
# Should fail
for *a, *b in items: pass
*x, *y = [1, 2, 3]
a, *b, c, *d = range(10)

# Should pass
for *a, b in items: pass
for a, *b, c in items: pass
for (a, *b), (*c, d) in items: pass  # Different nesting levels
```

## Expected Impact

### Fixes
- Proper validation of starred expression syntax
- Alignment with Python's assignment rules
- Test `TestStarTargets/multiple_stars_in_same_level` will pass

### Additional Considerations
- Need to ensure nested structures still work correctly
- May need to handle comprehension targets differently
- Error messages should be clear and match Python's

## References
- [PEP 3132](https://www.python.org/dev/peps/pep-3132/) - Extended Iterable Unpacking
- Python Grammar: `star_targets` rule restricts multiple stars
- CPython source: `Python/ast.c` validates starred expressions