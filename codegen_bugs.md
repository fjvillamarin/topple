# Codegen Bugs

This document tracks bugs found in the code generator during testing.

## Bug #1: MultiStmt generates incorrect output

**Status**: Open
**Found**: During POC testing with literals
**Test**: `TestMultiStmtBug` in `codegen_multistmt_test.go`

### Description
The `VisitMultiStmt` method generates incorrect output for multiple statements on one line. 

### Current Behavior
```go
input := &ast.MultiStmt{
    Stmts: []ast.Stmt{
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "a"}}},
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "b"}}},
        &ast.ExprStmt{Expr: &ast.Name{Token: lexer.Token{Lexeme: "c"}}},
    },
}
```

Generates:
```
a
; b
; c
```

### Expected Behavior
Should generate:
```
a; b; c
```

### Root Cause
The implementation in `codegen.go` at `VisitMultiStmt` is calling `Accept` on each statement, which adds newlines after each ExprStmt. The semicolons are being added between already-newlined statements.

### Fix Required
The `VisitMultiStmt` method needs to be rewritten to handle the semicolon-separated statements on a single line without triggering the automatic newline behavior of ExprStmt.

## Bug #2: Empty set literal ambiguity

**Status**: Confirmed
**Found**: During POC testing
**Test**: `TestEmptySetBug` in `codegen_emptyset_test.go`

### Description
An empty `SetExpr` generates `{}` which is actually an empty dict in Python, not an empty set.

### Current Behavior
```go
&ast.SetExpr{Elements: []ast.Expr{}}
```
Generates: `{}`

### Expected Behavior
Should generate: `set()` for an empty set, as `{}` creates an empty dict in Python.

### Impact
Medium - This will cause runtime errors where code expects a set but gets a dict. Both empty dict and empty set generate identical output, making them indistinguishable.

### Fix Required
The `VisitSetExpr` method in `codegen.go` needs to check if the set is empty and generate `set()` instead of `{}`.

## Bug #3: String literals missing quotes

**Status**: Open
**Found**: During golden file review
**Test**: String literal tests in consolidated `codegen_test.go`

### Description
String literals are generated without quotes, making them invalid Python code.

### Current Behavior
```go
&ast.Literal{Token: lexer.Token{Type: lexer.STRING, Lexeme: "hello world"}}
```
Generates: `hello world`

### Expected Behavior
Should generate: `"hello world"` or `'hello world'`

### Impact
High - Generated code is syntactically invalid Python

### Files Affected
- `testdata/expected/literals/string.py`
- `testdata/expected/expressions/literal_string.py`

### Fix Required
The `VisitLiteral` method needs to add quotes around string literals when generating code.

## Bug #4: Default parameter formatting

**Status**: Open
**Found**: During golden file review
**Test**: Function parameter tests in consolidated `codegen_test.go`

### Description
Default parameters in functions are generated with spaces around the equals sign, which violates PEP 8.

### Current Behavior
```
def greet(name = "World"):
```

### Expected Behavior
Should generate: `def greet(name="World"):`

### Impact
Low - Code works but violates Python style guidelines

### Files Affected
- `testdata/expected/functions/function_with_defaults.py`

### Fix Required
The parameter generation code needs to remove spaces around `=` for default values.

## Bug #5: Inconsistent trailing newlines

**Status**: Open
**Found**: During golden file review
**Test**: Various tests in consolidated `codegen_test.go`

### Description
Generated files have inconsistent trailing newlines - some have them, some don't.

### Current Behavior
Some files end with a newline, others don't

### Expected Behavior
All generated Python files should consistently end with a single newline character

### Impact
Low - Cosmetic issue but affects file consistency

### Fix Required
Ensure all code generation ends with exactly one newline character.

## Summary

### Test Results
- Golden file tests created and consolidated into single `codegen_test.go`
- Multiple bugs found during golden file generation and review

### Bug Severity
1. **String literals missing quotes** - High: Generates syntactically incorrect Python code
2. **MultiStmt** - High: Generates syntactically incorrect Python code  
3. **Empty Set** - Medium: Generates wrong type (dict instead of set)
4. **Default parameter formatting** - Low: Style violation
5. **Inconsistent trailing newlines** - Low: Cosmetic issue

### Test Files Status
- All individual test files consolidated into `codegen_test.go`
- Golden file test system implemented
- Generated golden files reveal multiple formatting and syntax issues