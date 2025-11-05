**GitHub Issue**: [#67](https://github.com/fjvillamarin/topple/issues/67)


# 🐛 Parser Bug: Async View Statement Support Missing

## Summary
The parser does not support `async view` statements. When parsing statements, it checks for `async def`, `async for`, and `async with`, but not `async view`. This causes a syntax error when trying to define asynchronous view components.

## Priority: **HIGH** 🔴
Async views are an important feature for modern web frameworks, especially for server-side rendering with asynchronous data fetching.

## Bug Details

### Location
- **File**: `compiler/parser/view_test.go:56`
- **Test**: `TestAsyncViewStatements/async_view`
- **Parser Files**: `compiler/parser/statements.go` (line 29-39), `compiler/parser/view.go`

### Current Behavior
```python
# Input
async view fetch_data():
    data = await api.fetch()
    <div>{data}</div>

# Error
at 'async': unexpected token

# Parser expects 'view' keyword directly, not preceded by 'async'
```

### Expected Behavior
Should parse async views similar to async functions:
```python
# Should compile to an async view class
class fetch_data(BaseView):
    async def _render(self):
        data = await api.fetch()
        return el("div", data)
```

## Root Cause Analysis

The issue is in the statement parsing logic:

1. **In `statements.go` (lines 29-39)**:
```go
case lexer.Async:
    // For async statements, we need to look ahead one more token
    if p.checkNext(lexer.For) {
        return p.forStatement()
    }
    if p.checkNext(lexer.With) {
        return p.withStatement()
    }
    if p.checkNext(lexer.Def) {
        return p.functionDef()
    }
    // MISSING: check for lexer.View!
```

2. **In `statements.go` (line 11)**:
```go
case lexer.View:
    return p.viewStatement()  // Only handles non-async views
```

3. **In `view.go`**:
The `viewStatement()` function doesn't handle the async modifier - it expects the `view` token immediately.

## Action Plan

### 1. **Add Async View Check in Statement Parser**
In `statements.go`, add a check for `async view`:
```go
case lexer.Async:
    if p.checkNext(lexer.For) {
        return p.forStatement()
    }
    if p.checkNext(lexer.With) {
        return p.withStatement()
    }
    if p.checkNext(lexer.Def) {
        return p.functionDef()
    }
    if p.checkNext(lexer.View) {  // ADD THIS
        return p.viewStatement()
    }
```

### 2. **Update viewStatement to Handle Async**
Modify `viewStatement()` to check for and consume the optional `async` keyword:
```go
func (p *Parser) viewStatement() (ast.Stmt, error) {
    // Check for async modifier
    isAsync := false
    var startToken lexer.Token
    
    if p.match(lexer.Async) {
        isAsync = true
        startToken, err := p.consume(lexer.View, "expected 'view' after 'async'")
        if err != nil {
            return nil, err
        }
    } else {
        // Consume the 'view' keyword
        var err error
        startToken, err = p.consume(lexer.View, "expected 'view'")
        if err != nil {
            return nil, err
        }
    }
    
    // ... rest of parsing logic ...
    
    return &ast.ViewStmt{
        Name:       viewName,
        TypeParams: typeParams,
        Params:     parameterList,
        ReturnType: returnType,
        Body:       body,
        IsAsync:    isAsync,  // Set based on async modifier
        Kind:       ast.ViewKindServerView,
        Span:       lexer.Span{Start: startToken.Start(), End: endPos},
    }, nil
}
```

### 3. **Update AST ViewStmt**
Ensure the `IsAsync` field in `ast.ViewStmt` is properly used (it's already there but hardcoded to false).

## Implementation Steps

1. **Update statement parser**: Add the `async view` check
2. **Modify viewStatement**: Handle optional async modifier
3. **Update tests**: Ensure async view tests pass
4. **Update transformers**: Ensure async views generate async render methods

## Test Cases to Verify

```python
# Basic async view
async view LoadingData():
    data = await fetch_data()
    <div>{data}</div>

# Async view with parameters
async view UserProfile(user_id: int):
    user = await get_user(user_id)
    <h1>{user.name}</h1>

# Async view with error handling
async view SafeFetch():
    try:
        result = await risky_operation()
        <div class="success">{result}</div>
    except Exception as e:
        <div class="error">{str(e)}</div>

# Mixed async/sync in same file
view Header():
    <header>Logo</header>

async view Content():
    data = await load_content()
    <main>{data}</main>
```

## Expected Impact

### Fixes
- Async view statements will be properly parsed
- Test `TestAsyncViewStatements` will pass
- Enables asynchronous data fetching in views

### Additional Work Required
- Transformer needs to generate async `_render()` methods
- Runtime might need updates to handle async rendering
- Documentation for async view usage

## References
- Python async/await syntax
- Similar frameworks: React Server Components, Fresh (Deno)
- View statement parsing starts at line 10 in view.go
- Async statement handling at lines 29-39 in statements.go