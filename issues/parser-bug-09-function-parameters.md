# 🐛 Parser Bug: Incorrect Function Parameter Categorization

## Summary
The parser incorrectly categorizes function parameters when counting regular vs. default parameters. Specifically, it doesn't properly distinguish keyword-only parameters (those after `*args`) from regular parameters, leading to incorrect counts in complex parameter patterns like `def func(a, b=2, *args, c, d=4, **kwargs):`.

## Priority: **MEDIUM** 🟡
This affects the accuracy of parameter analysis but doesn't prevent parsing. It may impact tools that rely on accurate parameter categorization.

## Bug Details

### Location
- **File**: `compiler/parser/function_test.go:467`
- **Test**: `TestFunctionParameterPatterns/complete_parameter_pattern`
- **Helper Function**: `countParameterTypes()` at line 71

### Current Behavior
```python
# Input
def func(a, b=2, *args, c, d=4, **kwargs):
    pass

# Expected counts:
# - Regular parameters: 1 (only 'a')
# - Parameters with defaults: 3 (b=2, d=4)
# - Keyword-only regular: 1 (c - has no default but comes after *args)

# Actual counts (based on bug report):
# - Regular parameters: 2 (counting both 'a' and 'c')
# - Parameters with defaults: 2 (only b=2 and d=4)
```

### Python's Parameter Categories
1. **Positional-only** (before `/` if present)
2. **Regular** (can be passed positionally or by keyword)
3. **Keyword-only** (after `*args` or bare `*`)
4. **Var-positional** (`*args`)
5. **Var-keyword** (`**kwargs`)

Parameters can also have defaults in any category except var-positional/keyword.

## Root Cause Analysis

### Issue 1: Incomplete Parameter Model
The `Parameter` struct doesn't track whether a parameter is keyword-only:
```go
type Parameter struct {
    Name         *Name
    Annotation   Expr
    Default      Expr
    IsStar       bool  // *args
    IsDoubleStar bool  // **kwargs
    IsSlash      bool  // positional-only marker
    // MISSING: IsKeywordOnly bool
}
```

### Issue 2: Oversimplified Counting
The `countParameterTypes()` function uses a simple if-else chain:
```go
if param.IsDoubleStar {
    doubleStarred++
} else if param.IsStar {
    starred++
} else if param.Default != nil {
    withDefaults++
} else {
    regular++  // This catches both regular AND keyword-only params
}
```

This doesn't account for:
- Keyword-only parameters without defaults (like `c` in the example)
- The position of parameters relative to `*args`

### Issue 3: Test Expectation Confusion
The test expects `c` to be counted as "regular" because it has no default, but in Python terminology, `c` is "keyword-only" because it comes after `*args`.

## Action Plan

### Option 1: Add Keyword-Only Tracking (Recommended)
1. Add `IsKeywordOnly` field to `Parameter` struct
2. Update parser to set this field for parameters after `*args` or `*`
3. Update counting logic to handle keyword-only parameters

```go
// In Parameter struct
type Parameter struct {
    // ... existing fields ...
    IsKeywordOnly bool  // Whether parameter is keyword-only
}

// In parser, when parsing parameters after *args
if p.VarArgIndex >= 0 && i > p.VarArgIndex {
    param.IsKeywordOnly = true
}

// Updated counting
func countParameterTypes(params *ast.ParameterList) (
    regular, regularDefaults, kwOnly, kwOnlyDefaults, starred, doubleStarred int) {
    
    for _, param := range params.Parameters {
        switch {
        case param.IsDoubleStar:
            doubleStarred++
        case param.IsStar:
            starred++
        case param.IsKeywordOnly && param.Default != nil:
            kwOnlyDefaults++
        case param.IsKeywordOnly:
            kwOnly++
        case param.Default != nil:
            regularDefaults++
        default:
            regular++
        }
    }
    return
}
```

### Option 2: Fix Test Expectations (Simple)
If the goal is just to count parameters with/without defaults regardless of category:
```go
// Keep current counting but fix test expectations
// For def func(a, b=2, *args, c, d=4, **kwargs):
// - Without defaults: 2 (a, c)
// - With defaults: 2 (b, d)
// - Starred: 1 (*args)
// - Double-starred: 1 (**kwargs)
```

### Option 3: Use ParameterList Indices
Use the existing `VarArgIndex` to determine keyword-only status:
```go
func isKeywordOnly(params *ast.ParameterList, index int) bool {
    return params.VarArgIndex >= 0 && index > params.VarArgIndex &&
           !params.Parameters[index].IsStar && 
           !params.Parameters[index].IsDoubleStar
}
```

## Test Cases to Verify

```python
# Basic parameter types
def f1(a): pass                    # regular: 1
def f2(a=1): pass                  # defaults: 1
def f3(a, b=2): pass               # regular: 1, defaults: 1

# Keyword-only parameters
def f4(*, a): pass                 # kw-only: 1
def f5(*, a=1): pass               # kw-only-defaults: 1
def f6(*args, a): pass             # starred: 1, kw-only: 1
def f7(*args, a=1): pass           # starred: 1, kw-only-defaults: 1

# Complex patterns
def f8(a, b=1, *args, c, d=2, **kwargs): pass
# regular: 1 (a)
# defaults: 1 (b)
# starred: 1 (*args)
# kw-only: 1 (c)
# kw-only-defaults: 1 (d)
# double-starred: 1 (**kwargs)

# Positional-only
def f9(a, b=1, /, c, d=2): pass
# pos-only: 1 (a)
# pos-only-defaults: 1 (b)
# regular: 1 (c)
# defaults: 1 (d)
```

## Expected Impact

### Fixes
- Accurate parameter categorization
- Better alignment with Python's parameter model
- More precise parameter analysis for tools

### Risks
- Requires AST changes (adding field)
- May break existing code that relies on current categorization
- Test expectations need updating

## Recommendation

Implement **Option 1** (Add Keyword-Only Tracking) because:
1. It properly models Python's parameter system
2. Enables accurate analysis for tools and documentation
3. Aligns with Python's official terminology

The test expectations should also be clarified to use Python's standard terminology rather than mixing "regular" for both positional and keyword-only parameters.

## References
- [PEP 3102](https://www.python.org/dev/peps/pep-3102/) - Keyword-Only Arguments
- [Python Function Definitions](https://docs.python.org/3/reference/compound_stmts.html#function-definitions)
- Python's `inspect.Parameter` kinds for categorization model