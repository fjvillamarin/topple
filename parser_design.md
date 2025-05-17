## Parser Implementation Plan

This document outlines a staged roadmap for turning `compiler/parser.go` into a **full-fledged Python 3 parser**.  Each stage introduces a slice of grammar rules and AST nodes while keeping the code runnable and testable, mirroring the incremental style used in *Crafting Interpreters*.

---

### Guiding Principles

1. **Single-Pass Recursive-Descent**  
   We mimic CPython's PEG grammar but implement it with classic recursive-descent functions (`parseExpr()`, `parseStmt()`, …).  Expressions use Pratt parsing for precedence.
2. **Thin AST Layer**  
   Re-use / extend the `Stmt` and `Expr` hierarchies already present in `compiler`.
3. **Always-Green Main**  
   The parser must compile at every stage; new grammar pieces are gated behind parser methods but unused code is allowed to return errors like `ErrUnimplemented`.
4. **Tests Drive Growth**  
   Each stage ships with unit tests that lock in behaviour and help refactors.

---

### Stage 0 – Skeleton & Infrastructure *(already partially done)*

* Files: `scanner.go`, `token.go`, `parser.go`  
* AST: `Expr`, `Stmt`, `Module`, `ExprStmt`  
* Capability: Accept a module containing **one empty newline**.

Action items:
1. Fix compile error in `statement()` by returning an `ExprStmt` or `nil, error`.
2. Add helper `synchronize()` for panic-mode error recovery.

---

### Stage 1 – Expression Statements

Subset of grammar:
```
file         : [statements] ENDMARKER
statements   : statement+
statement    : simple_stmts
simple_stmts : star_expressions NEWLINE
star_expressions : star_expression (',' star_expression)* [',']
star_expression  : expression
```

Goals:
* Parse a **single-line expression** terminated by NEWLINE or EOF.
* Support literals, identifiers, parenthesised expressions.

Tasks:
1. Pratt parser with precedence levels (primary → power → factor → term → sum → …)
2. Implement `parseExpression(precedence int)` helper.
3. new AST nodes: `Literal`, `Identifier`, `Unary`, `Binary`, `Grouping`.
4. Unit tests: `1 + 2 * 3`, `(a)`.

---

### Stage 2 – Full Expression Grammar

Add every rule under "EXPRESSIONS" section of `grammar_python.md` **except** `yield_expr`, comprehensions, and assignment expressions.

Highlights:
* `comparison`, boolean `and`/`or`, `not` inversion.
* `if‐expr` (ternary).
* Attribute access, calls, subscripts.

New AST nodes: `Call`, `Attribute`, `Subscript`, `Conditional`.

---

### Stage 3 – Simple Statements

Enable the entire `simple_stmt` list **excluding** `import`, `global`, `nonlocal`, `assert`, `del`, `raise`, `yield_stmt`.

Rules tackled:
* `assignment`
* `return_stmt` (only inside function bodies once those exist)

Details:
* Handle multi-target and unpacking assignments `(star_targets '=' )+ RHS`.
* Introduce `Assign` and `AugAssign` nodes.
* Parse type annotations but store them as optional `Expr` on the node.

AST additions: `Assign`, `AugAssign`, `AnnAssign`.

---

### Stage 4 – Block & Indentation Management

We need to mirror Python's INDENT/DEDENT tokens.  Scanner must already emit these; parser now groups them into `block` nodes.

* Add `Block` node containing `[]Stmt`.
* Extend `statement()` to dispatch between `simple_stmts` and `compound_stmt`.

Compound statements tackled in later stages.

---

### Stage 5 – Control Flow Statements

Implement subset of `compound_stmt`:
* `if_stmt`
* `while_stmt`
* `for_stmt` (non-async)

AST: `If`, `While`, `For`, `Break`, `Continue`.

---

### Stage 6 – Functions & Classes

* `function_def`, decorators, and `return_stmt` fully functional.
* `class_def` with base list.
* Add scopes for `global` / `nonlocal` semantic checks later.

AST: `FunctionDef`, `ClassDef`, `Parameter`, etc.

---

### Stage 7 – Imports & Exceptions

* `import_stmt`, `try_stmt`, `raise_stmt`, `assert_stmt`.
* `with_stmt` (non-async).

New nodes: `Import`, `ImportFrom`, `Try`, `With`, `Raise`, `Assert`.

---

### Stage 8 – Advanced Expressions

* `lambda`, `yield_expr`, comprehensions (`listcomp`, `setcomp`, `dictcomp`, `genexp`).
* `await`, async `for` / `with`.
* Assignment expressions (`:=`).

---

### Stage 9 – Pattern Matching (Python 3.10+)

Implement `match_stmt` and all `pattern` sub-rules.

---

### Stage 10 – Final Polish

* Complete coverage for invalid_* rules for better error messages.
* Docstrings and f-strings.
* Static type-checking hooks (`TYPE_COMMENT`, type ignores).
* Linter passes.

---

## File & Function Layout Sketch

```
compiler/
  parser.go      -- entry; public API Parse(tokens)
  parser_expr.go -- expression Pratt parsing helpers
  parser_stmt.go -- statements & blocks
  ast.go         -- (existing) Stmt/Expr hierarchy
  errors.go      -- ParseError, ErrUnimplemented
```

Each grammar rule `x` maps to `parseX()` returning *node* and error.  The *first* alternative is attempted, backtracking is avoided by ordering calls in the same way as written in grammar (like CPython does).

---

## Testing Strategy

1. Stage-specific golden tests in `tests/parser_stageN_test.go`.
2. Each test feeds tokens from the existing scanner and checks the produced AST via `String()`.
3. Property tests (gofuzz) later to ensure round-trip between AST → pretty-print → scanner → parser.

---

## Next Steps

1. **Finish Stage 0** – fix the missing return and write minimal tests.
2. Create `parser_expr.go` with Pratt parser for Stage 1.
3. Iterate! 