# 🐛 Codegen Bug: Default Parameter Spacing Violates PEP 8

## Summary
The code generator adds spaces around the equals sign in function default parameters (`name = "World"`), which violates PEP 8 style guidelines. While the code is syntactically valid, it doesn't follow Python's standard formatting conventions.

## Priority: **LOW** 🟢
This is a style issue that doesn't affect functionality but impacts code quality and consistency with Python standards.

## Bug Details

### Location
- **File**: `compiler/codegen/compound.go:242-245`
- **Method**: `VisitParameter`
- **Evidence**: `testdata/expected/functions/function_with_defaults.py`

### Current Behavior
```python
# Generated output:
def greet(name = "World"):
    pass

# Spaces around = in default parameter
```

### Expected Behavior
```python
# PEP 8 compliant:
def greet(name="World"):
    pass

# No spaces around = in default parameter
```

### PEP 8 Guidelines
According to [PEP 8](https://www.python.org/dev/peps/pep-0008/#other-recommendations):
> Don't use spaces around the = sign when used to indicate a keyword argument or a default parameter value.

```python
# Correct:
def complex(real, imag=0.0):
    return magic(r=real, i=imag)

# Wrong:
def complex(real, imag = 0.0):
    return magic(r = real, i = imag)
```

## Root Cause Analysis

In `VisitParameter` method:

```go
func (cg *CodeGenerator) VisitParameter(p *ast.Parameter) ast.Visitor {
    // ... other code ...
    
    if p.Default != nil {
        cg.write(" = ")  // BUG: Spaces around equals sign
        p.Default.Accept(cg)
    }
    
    return cg
}
```

The issue is the string `" = "` which includes spaces before and after the equals sign.

## Action Plan

### Simple Fix
Change the spacing in default parameter generation:

```go
func (cg *CodeGenerator) VisitParameter(p *ast.Parameter) ast.Visitor {
    // ... other code ...
    
    if p.Default != nil {
        cg.write("=")  // FIX: No spaces around equals
        p.Default.Accept(cg)
    }
    
    return cg
}
```

### Additional Consideration
Also check keyword argument generation in `VisitArgument`:

```go
func (cg *CodeGenerator) VisitArgument(a *ast.Argument) ast.Visitor {
    if a.Name != nil {
        a.Name.Accept(cg)
        cg.write("=")  // Should also have no spaces
    }
    // ...
}
```

Looking at the code (expressions.go:284), this is already correct - it uses `cg.write("=")` without spaces.

## Test Cases

```python
# Simple default
def func1(x=1): pass

# Multiple defaults
def func2(a, b=2, c="hello"): pass

# Complex defaults
def func3(items=[], mapping={}): pass

# With type annotations
def func4(name: str = "default"): pass

# Mixed parameters
def func5(a, b=1, *args, c=2, **kwargs): pass

# Lambda defaults (if supported)
lambda x=0: x + 1

# Nested defaults
def func6(callback=lambda x: x**2): pass
```

## Impact Analysis

### Positive Impact
- PEP 8 compliance
- Better integration with Python linters (flake8, pylint)
- Consistent with Python community standards
- Professional code appearance

### No Functional Impact
- The generated code works identically
- Only affects whitespace
- No semantic changes

### Tool Integration
Many Python tools expect PEP 8 compliance:
- **Black formatter**: Would automatically fix this
- **flake8**: Would report `E251 unexpected spaces around keyword / parameter equals`
- **IDEs**: May show style warnings

## Related Issues

Check for similar spacing issues in:
- ✅ Keyword arguments in calls: Already correct (`cg.write("=")`)
- Type annotations: Uses `cg.write(": ")` which is correct
- Assignment statements: Uses `cg.write(" = ")` which is correct for regular assignments

The spacing rules are context-dependent:
- Regular assignment: `x = 1` (spaces)
- Default parameters: `def f(x=1)` (no spaces)
- Keyword arguments: `f(x=1)` (no spaces)
- Type annotations: `x: int = 1` (spaces around =, space after :)

## Expected Impact

### Fixes
- Generated functions will be PEP 8 compliant
- No linter warnings for parameter spacing
- Better consistency with Python ecosystem

### Minimal Change
- Only affects one line of code
- No risk of breaking functionality
- Easy to test and verify

## References
- [PEP 8 - Other Recommendations](https://www.python.org/dev/peps/pep-0008/#other-recommendations)
- [Python Style Guide](https://google.github.io/styleguide/pyguide.html#3.5-blank-lines)
- Common linter rules: E251 (flake8), C0326 (pylint)