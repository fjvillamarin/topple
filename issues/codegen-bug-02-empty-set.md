# 🐛 Codegen Bug: Empty Set Generates Dict Literal

## Summary
The code generator incorrectly generates `{}` for an empty `SetExpr`, which creates an empty dictionary in Python, not an empty set. This is a critical semantic error that will cause runtime type errors.

## Priority: **MEDIUM** 🟡
This generates semantically incorrect code that will cause runtime errors when set operations are attempted on what is actually a dict.

## Bug Details

### Location
- **File**: `compiler/codegen/expressions.go:148-158`
- **Method**: `VisitSetExpr`
- **Test**: Would be in `codegen_emptyset_test.go` (mentioned in bug report)

### Current Behavior
```go
// Input AST
&ast.SetExpr{Elements: []ast.Expr{}}

// Generated output:
{}

// This creates a DICT, not a SET!
```

### Expected Behavior
```python
set()
```

### Python Semantics
In Python:
- `{}` creates an empty **dictionary**
- `set()` creates an empty **set**
- `{1, 2, 3}` creates a set with elements

This is a historical quirk in Python - sets were added after dicts, so `{}` was already taken.

## Root Cause Analysis

The current implementation doesn't special-case empty sets:

```go
func (cg *CodeGenerator) VisitSetExpr(s *ast.SetExpr) ast.Visitor {
    cg.write("{")
    for i, elem := range s.Elements {
        if i > 0 {
            cg.write(", ")
        }
        elem.Accept(cg)
    }
    cg.write("}")  // BUG: Empty sets become {} which is a dict!
    return cg
}
```

## Action Plan

### Solution: Check for Empty Set
Add a special case for empty sets:

```go
func (cg *CodeGenerator) VisitSetExpr(s *ast.SetExpr) ast.Visitor {
    // Special case: empty set must use set() constructor
    if len(s.Elements) == 0 {
        cg.write("set()")
        return cg
    }
    
    // Non-empty sets use {} syntax
    cg.write("{")
    for i, elem := range s.Elements {
        if i > 0 {
            cg.write(", ")
        }
        elem.Accept(cg)
    }
    cg.write("}")
    return cg
}
```

## Test Cases

```python
# Empty set (the bug case)
s1 = set()              # NOT: s1 = {}

# Non-empty sets (these work correctly)
s2 = {1, 2, 3}
s3 = {"a", "b", "c"}
s4 = {x for x in range(5)}  # set comprehension

# Empty dict (for comparison)
d1 = {}
d2 = dict()

# Type checking
assert type(set()) == set
assert type({}) == dict
assert type({1}) == set
```

## Impact Analysis

### Runtime Errors
This bug will cause runtime errors when set operations are used:
```python
# Generated code: s = {}  (actually a dict)
s = {}

# These will fail:
s.add(1)        # AttributeError: 'dict' object has no attribute 'add'
s.remove(2)     # AttributeError: 'dict' object has no attribute 'remove'
s.union({3})    # AttributeError: 'dict' object has no attribute 'union'

# These will "work" but with wrong semantics:
len(s)          # Works for both dict and set
if 1 in s:      # Works but checks keys for dict
for x in s:     # Iterates over keys for dict
```

### Detection
The bug is subtle because some operations work on both types:
- `len()`, `in`, iteration all work
- Only set-specific methods will fail
- Type checks will reveal the issue

## Expected Impact

### Fixes
- Empty sets will be correctly generated as `set()`
- Set operations will work as expected
- Type consistency will be maintained

### No Breaking Changes
- Non-empty sets already work correctly
- Only affects the edge case of empty sets

## Additional Considerations

### Set Literals vs Constructor
Python has two ways to create sets:
```python
# Literal syntax (only for non-empty)
s = {1, 2, 3}

# Constructor syntax (works for any case)
s = set()          # empty
s = set([1, 2, 3]) # from iterable
```

The fix uses the minimal change - only using `set()` for the empty case.

### Similar Issues
Check if there are similar issues with:
- Empty tuples: `()` is correct
- Empty lists: `[]` is correct
- Empty dicts: `{}` is correct
- Only sets have this special case!

## References
- [Python Set Documentation](https://docs.python.org/3/library/stdtypes.html#set)
- [PEP 218](https://www.python.org/dev/peps/pep-0218/) - Adding a Built-In Set Object Type
- Historical note: Sets were added in Python 2.4, after `{}` was established for dicts