# 🐛 Parser Bug: Incorrect Operator Precedence - Bitwise OR vs Equality

## Summary
The parser incorrectly gives bitwise OR (`|`) higher precedence than equality comparison (`==`), which violates Python's operator precedence rules. This causes expressions like `x | y == 0` to be parsed as `(x | y) == 0` instead of the correct `x | (y == 0)`.

## Priority: **HIGH** 🔴
This affects core expression evaluation and can lead to incorrect program behavior.

## Bug Details

### Location
- **File**: `compiler/parser/comparisons_test.go:121`
- **Test**: `TestBitwiseOr/mixed_with_comparison`
- **Parser Files**: `compiler/parser/bitwise.go`, `compiler/parser/comparisons.go`

### Current Behavior
```python
# Input
x | y == 0

# Currently parsed as:
(x | y) == 0

# Should be parsed as:
x | (y == 0)
```

### Python Reference
In Python's operator precedence table, comparison operators (`==`, `!=`, `<`, `>`, etc.) have higher precedence than bitwise operators (`|`, `&`, `^`).

```python
# Python correctly evaluates this as:
>>> x = 5
>>> y = 3
>>> x | y == 0  # Evaluates as: x | (y == 0) = 5 | False = 5
5
```

## Root Cause Analysis

After examining the code, I found the exact issue:

**In `compiler/parser/comparisons.go` line 10**, the `comparison()` function starts by calling `p.bitwiseOr()`:

```go
func (p *Parser) comparison() (ast.Expr, error) {
    left, err := p.bitwiseOr()  // BUG: This gives bitwise OR higher precedence!
    if err != nil {
        return nil, err
    }
    // ... rest of comparison parsing
}
```

This creates the precedence chain: 
`comparison` → `bitwiseOr` → `bitwiseXor` → `bitwiseAnd` → ...

But in Python, comparisons should have HIGHER precedence than bitwise operations!

### Current Precedence Chain (INCORRECT)
```
expression()
  → disjunction() (or)
    → conjunction() (and) 
      → inversion() (not)
        → comparison() (==, !=, <, >, etc.)
          → bitwiseOr() (|)     ❌ WRONG ORDER!
            → bitwiseXor() (^)
              → bitwiseAnd() (&)
                → shiftExpr() (<<, >>)
                  → sum() (+, -)
                    → ...
```

### Correct Python Precedence Chain
According to Python's grammar, the correct precedence should be:
```
expression()
  → disjunction() (or)
    → conjunction() (and)
      → inversion() (not)
        → comparison() (==, !=, <, >, etc.)
          → bitwiseOr() (|)     ✅ CORRECT: comparisons bind tighter
            → bitwiseXor() (^)
              → bitwiseAnd() (&)
                → ...
```

## Action Plan

### 1. **Fix the precedence chain**
The comparison operators need to parse their operands from a LOWER precedence level (higher binding strength). We need to find what comes after bitwise operations in the precedence chain.

### 2. **Specific Changes Required**
- In `comparison()`, change the operand parsing from `p.bitwiseOr()` to the next level down (likely `p.shiftExpr()` or `p.sum()`)
- Ensure the precedence chain is properly ordered according to Python's grammar

### 3. **Update Tests**
- The test expectation is correct - it expects `==` to have higher precedence
- After fixing, the test should pass

### 4. **Comprehensive Testing**
- Test all combinations of bitwise and comparison operators
- Verify against Python's actual behavior

## Implementation Steps

1. **Identify correct precedence level**: Check what `bitwiseAnd()` calls (it calls `shiftExpr()`)
2. **Update comparison parsing**: Change `comparison()` to call `p.shiftExpr()` instead of `p.bitwiseOr()`
3. **Run tests**: Verify the specific test now passes
4. **Check for regressions**: Ensure no other tests break

## The Fix

Based on my analysis, the fix should be:

```go
// In compiler/parser/comparisons.go
func (p *Parser) comparison() (ast.Expr, error) {
    left, err := p.shiftExpr()  // FIX: Skip bitwise operations!
    if err != nil {
        return nil, err
    }
    
    // ... rest of comparison parsing stays the same
    
    // When parsing right operands in the loop:
    right, err := p.shiftExpr()  // Also fix here!
    // ...
}
```

This ensures that comparison operators bind more tightly than bitwise operators, matching Python's behavior.

## Expected Impact

### Fixes
- Correct parsing of mixed bitwise and comparison expressions
- Alignment with Python's operator precedence rules
- Test `TestBitwiseOr/mixed_with_comparison` will pass

### Risks
- May affect existing code that accidentally relied on incorrect precedence
- Need to verify all expression parsing tests still pass

## References
- [Python Operator Precedence](https://docs.python.org/3/reference/expressions.html#operator-precedence)
- Python Grammar shows: `comparison: bitwise_or (comp_op bitwise_or)*`
- But the full precedence chain shows comparisons come BEFORE bitwise in evaluation order

## Verification
After fix, the expression `x | y == 0` should parse as:
```
BinaryOp(
    op=BitOr,
    left=Name(x),
    right=Compare(
        left=Name(y),
        ops=[Eq],
        comparators=[Constant(0)]
    )
)
```