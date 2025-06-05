# 🐛 Parser Bug: Try Statement Clause Ordering Validation

## Summary
The parser does not properly validate the ordering of clauses in try statements. Python requires strict ordering: `try` → `except*` → `else?` → `finally?`, but the parser accepts invalid orderings like `finally` before `except`, only failing with generic "expected" errors rather than specific ordering violations.

## Priority: **MEDIUM** 🟡
While this allows invalid Python syntax to be partially parsed, it will eventually fail during parsing, just with less helpful error messages.

## Bug Details

### Location
- **File**: `compiler/parser/try_test.go:679`
- **Test**: `TestTryCombinations/finally_before_except`
- **Parser File**: `compiler/parser/try.go:40-74`

### Current Behavior
```python
# Input (invalid Python)
try:
    pass
finally:
    pass
except:
    pass

# Current: Parser consumes 'try' and 'finally', then fails at 'except'
# Error: Generic "expected" error
# Should be: "SyntaxError: 'except' clause cannot appear after 'finally'"
```

### Expected Behavior
The parser should enforce Python's clause ordering rules:
1. `try` block (required)
2. One or more `except` blocks (optional)
3. `else` block (optional, only valid if except blocks exist)
4. `finally` block (optional, must be last if present)

## Root Cause Analysis

The current parser structure in `try.go`:

```go
if p.check(lexer.Except) {
    // Parse except blocks
    for p.match(lexer.Except) { ... }
    
    // Parse optional else
    if p.match(lexer.Else) { ... }
    
    // Parse optional finally
    if p.match(lexer.Finally) { ... }
    
} else if p.match(lexer.Finally) {
    // Just finally without except
    ...
}
```

This structure assumes correct input ordering. When the input has incorrect ordering:
1. The parser doesn't see `except` first (because `finally` comes first)
2. It takes the `else if` branch and parses `finally`
3. The remaining `except` in the input causes a generic parsing error
4. No specific clause ordering error is generated

## Action Plan

### Option 1: Post-Parse Validation (Simple)
After parsing all parts, check if any tokens remain that violate ordering:

```go
// After parsing try/except/else/finally
if p.check(lexer.Except) {
    return nil, p.error(p.peek(), "'except' clause cannot appear after 'finally'")
}
if p.check(lexer.Else) {
    if len(finallyBody) > 0 {
        return nil, p.error(p.peek(), "'else' clause cannot appear after 'finally'")
    }
    if len(excepts) == 0 {
        return nil, p.error(p.peek(), "'else' clause requires at least one 'except'")
    }
}
```

### Option 2: State Machine Approach (Comprehensive)
Track what clauses have been seen and validate transitions:

```go
type tryParseState int

const (
    tryStateBody tryParseState = iota
    tryStateExcept
    tryStateElse
    tryStateFinally
    tryStateDone
)

func (p *Parser) tryStatement() (ast.Stmt, error) {
    state := tryStateBody
    // ... parse try block ...
    
    for !p.isAtEnd() {
        switch {
        case p.check(lexer.Except):
            if state > tryStateExcept {
                return nil, p.error(p.peek(), 
                    "'except' must come before 'else' and 'finally'")
            }
            state = tryStateExcept
            // ... parse except ...
            
        case p.check(lexer.Else):
            if state != tryStateExcept {
                return nil, p.error(p.peek(), 
                    "'else' requires preceding 'except' clause")
            }
            state = tryStateElse
            // ... parse else ...
            
        case p.check(lexer.Finally):
            if state == tryStateDone {
                return nil, p.error(p.peek(), 
                    "unexpected 'finally' clause")
            }
            state = tryStateFinally
            // ... parse finally ...
            break // finally must be last
            
        default:
            break
        }
    }
}
```

### Option 3: Lookahead Validation (Minimal Change)
Before taking the `else if p.match(lexer.Finally)` branch, peek ahead:

```go
} else if p.match(lexer.Finally) {
    // Parse finally
    finallyBody, err = p.parseFinallyBlock()
    if err != nil {
        return nil, err
    }
    
    // Check for invalid clauses after finally
    if p.check(lexer.Except) {
        return nil, p.error(p.peek(), "'except' cannot appear after 'finally'")
    }
    if p.check(lexer.Else) {
        return nil, p.error(p.peek(), "'else' cannot appear after 'finally'")
    }
}
```

## Test Cases to Verify

```python
# Valid orderings
try: pass
except: pass

try: pass
except: pass
else: pass

try: pass
except: pass
finally: pass

try: pass
except: pass
else: pass
finally: pass

try: pass
finally: pass

# Invalid orderings (should error)
try: pass
finally: pass
except: pass  # Error: except after finally

try: pass
finally: pass
else: pass    # Error: else after finally

try: pass
else: pass    # Error: else without except
except: pass

try: pass
except: pass
finally: pass
else: pass    # Error: else after finally

# Multiple except blocks
try: pass
except ValueError: pass
except KeyError: pass
finally: pass
except: pass  # Error: except after finally
```

## Expected Impact

### Fixes
- Clear error messages for clause ordering violations
- Python-compliant try statement validation
- Better developer experience with specific errors

### Implementation Complexity
- Option 1 (Post-parse): Easiest, minimal code change
- Option 2 (State machine): Most robust but requires refactoring
- Option 3 (Lookahead): Good balance of simplicity and effectiveness

## Recommendation

Implement **Option 3** (Lookahead Validation) as it:
- Requires minimal changes to existing code
- Provides specific error messages
- Catches errors at the right parsing stage
- Maintains the current parsing structure

## References
- [Python Grammar - Try Statement](https://docs.python.org/3/reference/compound_stmts.html#the-try-statement)
- Python Language Reference 8.4: The try statement
- Similar validation in CPython parser