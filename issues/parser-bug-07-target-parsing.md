# 🐛 Parser Bug: Complex Target Parsing Failures

## Summary
The parser fails to correctly parse various target expressions used in assignments, particularly simple subscript targets like `arr[0]` and complex chained expressions like `obj.method()[0].attr`. The parser's lookahead requirements and backtracking logic cause valid Python assignment targets to be rejected.

## Priority: **HIGH** 🔴
Assignment target parsing is fundamental to Python syntax. This bug prevents basic operations like `arr[0] = value` from parsing correctly.

## Bug Details

### Location
- **Files**: `compiler/parser/targets_test.go` (multiple test cases)
- **Parser File**: `compiler/parser/targets.go`
- **Key Issues**:
  - Line 189: Incorrect routing in `singleTarget()`
  - Line 28-30: Overly strict lookahead requirement in `tPrimary()`
  - Line 505-511: Backtracking issues in `targetWithStarAtom()`

### Current Behavior

#### Issue 1: Simple Subscript Targets
```python
# Input
arr[0] = value

# Error
"expected '.' or '[' after primary expression"

# Root cause: Parser expects chained accessors even for simple subscripts
```

#### Issue 2: Complex Chained Expressions
```python
# Input
obj.method()[0].attr = value

# Parsing issues:
# - AST node type incorrectly identified
# - Final operation type not properly recognized
# - Method calls in targets not handled correctly
```

#### Issue 3: Nested Subscripts
```python
# Input
matrix[row][col] = value

# Error: Parser fails on deeply nested subscript chains
```

## Root Cause Analysis

The parser has several interconnected issues:

### 1. **Incorrect Routing Logic**
In `singleTarget()` (line 189), when it sees `arr` followed by `[`, it routes to `singleSubscriptAttributeTarget()`, which then calls `tPrimary()`. But `tPrimary()` is designed for expressions that MUST have chained accessors, not simple ones.

### 2. **Overly Strict Lookahead**
`tPrimary()` requires a lookahead token after the initial atom (line 28-30). This makes sense for grammar rule `t_primary: atom &t_lookahead`, but fails for simple cases like `arr[0]` where there's no further chaining.

### 3. **Grammar Interpretation Issues**
The Python grammar's target rules are complex:
```
single_target:
    | single_subscript_attribute_target
    | NAME
    | '(' single_target ')'

single_subscript_attribute_target:
    | t_primary '.' NAME !t_lookahead
    | t_primary '[' slices ']' !t_lookahead

t_primary:
    | atom &t_lookahead
    | t_primary '.' NAME &t_lookahead
    | t_primary '[' slices ']' &t_lookahead
    | t_primary '(' [arguments] ')' &t_lookahead
```

The parser is trying to follow these rules literally but missing the intent: `t_primary` is for chained expressions, while `single_subscript_attribute_target` can handle both chained and simple cases.

## Action Plan

### Fix 1: Improve `singleTarget()` Routing
Don't immediately route to `singleSubscriptAttributeTarget()` for all subscripts:

```go
func (p *Parser) singleTarget() (ast.Expr, error) {
    if p.check(lexer.Identifier) {
        // Parse the identifier first
        name := p.advance()
        
        // Now check what follows
        if p.match(lexer.LeftBracket) {
            // Handle simple subscript: name[index]
            indices, err := p.slices()
            if err != nil {
                return nil, err
            }
            _, err = p.consume(lexer.RightBracket, "expected ']'")
            if err != nil {
                return nil, err
            }
            
            // Check if there's more chaining
            if p.tLookahead() {
                // Continue as t_primary...
                // (backtrack and use existing logic)
            } else {
                // Simple subscript, return it
                return &ast.Subscript{
                    Object: &ast.Name{Token: name},
                    Indices: indices,
                }, nil
            }
        } else if p.check(lexer.Dot) || p.check(lexer.LeftParen) {
            // Has chaining, use existing logic
            p.Current-- // backtrack the name
            return p.singleSubscriptAttributeTarget()
        } else {
            // Just a name
            return &ast.Name{Token: name}, nil
        }
    }
    // ... rest of function
}
```

### Fix 2: Make `tPrimary()` More Flexible
Allow `tPrimary()` to handle cases without further lookahead:

```go
func (p *Parser) tPrimary() (ast.Expr, error) {
    expr, err := p.atom()
    if err != nil {
        return nil, err
    }
    
    // Make lookahead optional for the initial atom
    if !p.tLookahead() {
        // No chaining, just return the atom
        // This allows simple cases like 'arr' in 'arr[0]'
        return expr, nil
    }
    
    // ... rest of existing logic for chained accessors
}
```

### Fix 3: Separate Simple and Complex Target Parsing
Create dedicated functions for different complexity levels:

```go
// For simple targets like 'x', 'arr[0]', 'obj.attr'
func (p *Parser) simpleTarget() (ast.Expr, error) {
    // Handle non-chained targets
}

// For complex chained targets
func (p *Parser) complexTarget() (ast.Expr, error) {
    // Handle chained accessors with proper backtracking
}
```

## Test Cases to Verify

```python
# Simple cases that should work
x = 1                          # Name target
arr[0] = 2                     # Simple subscript
obj.attr = 3                   # Simple attribute
(x) = 4                        # Parenthesized target

# Complex cases
obj.method()[0] = value        # Method call + subscript
matrix[i][j] = value           # Nested subscripts
data[key].value = x            # Subscript + attribute
obj.a.b.c = value              # Chained attributes

# Star targets
*rest = [1, 2, 3]              # Star expression
a, *b, c = items               # Star in sequence

# Edge cases
obj[a:b][c] = value            # Slice + subscript
func()[0][1].attr = value      # Complex chaining
((x)) = value                  # Nested parentheses
```

## Expected Impact

### Fixes
- Basic subscript assignments will work
- Complex chained targets will parse correctly
- Python-compatible assignment target parsing

### Risks
- Need careful testing of all target forms
- Backtracking logic must be precise
- Performance impact should be minimal

## Implementation Priority

1. **Fix simple subscripts first** - Most common use case
2. **Then handle complex chaining** - Less common but important
3. **Finally optimize and refactor** - Clean up the code

## References
- Python Grammar: [Full Grammar Specification](https://docs.python.org/3/reference/grammar.html)
- PEP 572 (Assignment Expressions) for target rules
- CPython's parser implementation for reference