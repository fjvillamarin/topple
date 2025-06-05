# 🐛 Codegen Bug: MultiStmt Generates Incorrect Output

## Summary
The `VisitMultiStmt` method generates syntactically incorrect Python code by placing semicolons at the beginning of new lines instead of keeping all statements on a single line separated by semicolons.

## Priority: **HIGH** 🔴
This generates syntactically invalid Python code, making multi-statement lines unusable.

## Bug Details

### Location
- **File**: `compiler/codegen/statements.go:178-186`
- **Method**: `VisitMultiStmt`
- **Test**: Would be in `codegen_multistmt_test.go` (mentioned in bug report)

### Current Behavior
```go
// Input AST
&ast.MultiStmt{
    Stmts: []ast.Stmt{
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "a"}}},
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "b"}}},
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "c"}}},
    },
}

// Generated output:
a
; b
; c

// This is INVALID Python syntax!
```

### Expected Behavior
```python
a; b; c
```

In Python, semicolons are used to separate multiple statements on the same line.

## Root Cause Analysis

The issue is in the `VisitMultiStmt` implementation:

```go
func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
    for i, stmt := range m.Stmts {
        if i > 0 {
            cg.write("; ")
        }
        stmt.Accept(cg)  // BUG: This calls VisitExprStmt which adds newline!
    }
    return cg
}
```

The problem:
1. `stmt.Accept(cg)` calls the visitor for each statement
2. `VisitExprStmt` (line 10-14) automatically adds a newline after the expression
3. The semicolon is written AFTER the previous statement's newline
4. Result: semicolons appear at the start of lines

## Action Plan

### Solution 1: Special Handling in MultiStmt (Recommended)
Don't use the statement's Accept method directly. Instead, generate the expression without the automatic newline:

```go
func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
    for i, stmt := range m.Stmts {
        if i > 0 {
            cg.write("; ")
        }
        
        // Special handling for each statement type
        switch s := stmt.(type) {
        case *ast.ExprStmt:
            // Generate expression without newline
            s.Expr.Accept(cg)
        case *ast.AssignStmt:
            // Generate assignment without newline
            for i, target := range s.Targets {
                if i > 0 {
                    cg.write(" = ")
                }
                target.Accept(cg)
            }
            cg.write(" = ")
            s.Value.Accept(cg)
        // Add other statement types as needed
        default:
            // Fallback - but this will add unwanted newlines
            s.Accept(cg)
        }
    }
    cg.newline()  // Single newline at the end
    return cg
}
```

### Solution 2: Add Context Flag
Add a flag to suppress newlines when generating statements in a multi-statement context:

```go
type CodeGenerator struct {
    // ... existing fields ...
    suppressNewline bool
}

func (cg *CodeGenerator) VisitExprStmt(e *ast.ExprStmt) ast.Visitor {
    e.Expr.Accept(cg)
    if !cg.suppressNewline {
        cg.newline()
    }
    return cg
}

func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
    cg.suppressNewline = true
    defer func() { cg.suppressNewline = false }()
    
    for i, stmt := range m.Stmts {
        if i > 0 {
            cg.write("; ")
        }
        stmt.Accept(cg)
    }
    cg.newline()
    return cg
}
```

### Solution 3: Extract Statement Body Generation
Create helper methods that generate statement bodies without newlines:

```go
func (cg *CodeGenerator) generateExprStmtBody(e *ast.ExprStmt) {
    e.Expr.Accept(cg)
}

func (cg *CodeGenerator) VisitExprStmt(e *ast.ExprStmt) ast.Visitor {
    cg.generateExprStmtBody(e)
    cg.newline()
    return cg
}

func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
    for i, stmt := range m.Stmts {
        if i > 0 {
            cg.write("; ")
        }
        
        switch s := stmt.(type) {
        case *ast.ExprStmt:
            cg.generateExprStmtBody(s)
        // ... other cases
        }
    }
    cg.newline()
    return cg
}
```

## Test Cases

```python
# Basic multi-statement
a; b; c

# With different statement types
x = 1; y = 2; print(x + y)

# Complex expressions
func1(); obj.method(); x[0] = 5

# With assignments
a = 1; b = a + 1; c = b * 2

# Mixed statements
import sys; x = 1; print(x)
```

## Expected Impact

### Fixes
- Multi-statement lines will generate valid Python syntax
- Enables compact code generation where appropriate
- Aligns with Python's semicolon usage

### Implementation Notes
- Solution 1 is most straightforward but requires handling each statement type
- Solution 2 is cleaner but adds state to the generator
- Solution 3 provides best separation of concerns

## References
- Python allows multiple statements on one line separated by semicolons
- [Python Grammar](https://docs.python.org/3/reference/grammar.html): `simple_stmt: small_stmt (';' small_stmt)* [';'] NEWLINE`
- Common use case: one-liners, compact initialization, etc.