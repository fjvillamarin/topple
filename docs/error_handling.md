# Sylfie Compiler Error Handling

This document describes how the Sylfie compiler handles and reports errors throughout the compilation pipeline.

## Error Philosophy

The Sylfie compiler follows these principles for error handling:

1. **Clear and Actionable**: Error messages should clearly explain what went wrong and how to fix it
2. **Precise Location**: Errors include exact file, line, and column information
3. **Context-Aware**: Show relevant source code context around errors
4. **Recovery**: Continue parsing/compilation to report multiple errors when possible
5. **User-Friendly**: Avoid compiler jargon; use terms familiar to Python developers

## Error Types

### Lexical Errors

Errors detected during tokenization:

```
Error: Unterminated string literal
  --> hello.psx:3:14
   |
 3 |     message = "Hello, world
   |               ^~~~~~~~~~~~~~ expected closing quote
```

Common lexical errors:
- Unterminated strings
- Invalid numeric literals
- Illegal characters
- Mixed indentation (tabs vs spaces)
- Invalid escape sequences

### Syntax Errors

Errors detected during parsing:

```
Error: Expected identifier after 'view' keyword
  --> components.psx:5:5
   |
 5 | view ():
   |      ^ expected view name here
```

Common syntax errors:
- Missing colons after statements
- Unmatched brackets/parentheses
- Invalid view syntax
- Malformed HTML elements
- Missing required keywords

### Semantic Errors

Errors detected during semantic analysis:

```
Error: Undefined variable 'user_name'
  --> profile.psx:8:22
   |
 8 |         <h1>Welcome, {user_name}!</h1>
   |                       ^^^^^^^^^ 'user_name' is not defined
   |
help: did you mean 'username'?
```

Common semantic errors:
- Undefined variables
- Invalid scope usage (global/nonlocal)
- Type mismatches
- Duplicate definitions
- Invalid view composition

### Resolution Errors

Errors during name resolution:

```
Error: Cannot use 'global' declaration in module scope
  --> app.psx:2:1
   |
 2 | global x
   | ^^^^^^^^ 'global' can only be used inside functions
```

## Error Recovery

The compiler attempts to recover from errors to provide multiple diagnostics:

### Synchronization Points

The parser synchronizes at:
- Statement boundaries (newlines)
- Block delimiters (dedents)
- HTML element boundaries
- View definitions

### Example Multi-Error Report

```
Found 3 errors in app.psx:

Error: Missing closing tag for 'div'
  --> app.psx:10:5
   |
10 |     <div class="container">
   |     ^^^^^^^^^^^^^^^^^^^^^^^ opened here
...
15 |     <span>Content</span>
   |                        ^ expected </div> before </span>

Error: Undefined variable 'itms'
  --> app.psx:12:20
   |
12 |         for item in itms:
   |                     ^^^^ 'itms' is not defined
   |
help: did you mean 'items'?

Error: Invalid attribute syntax
  --> app.psx:13:21
   |
13 |             <li class={>Item</li>
   |                       ^ expected expression after '{'
```

## Error Message Format

### Standard Format

```
Error: <description>
  --> <file>:<line>:<column>
   |
<line> | <source code>
   |     <marker and annotation>
   |
help: <suggestion if available>
```

### With Context

For complex errors, additional context is shown:

```
Error: View 'UserCard' is not defined
  --> dashboard.psx:25:9
   |
23 |     <div class="users">
24 |         for user in users:
25 |             <UserCard user={user} />
   |             ^^^^^^^^ unknown view
   |
note: available views in scope:
  - ProfileCard (imported from components)
  - ItemCard (defined at line 5)
```

## Error Categories

### HTML-Specific Errors

```
Error: Self-closing tag 'div' cannot have content
  --> layout.psx:7:5
   |
 7 |     <div class="wrapper" />
   |     ^^^^^^^^^^^^^^^^^^^^^^
 8 |         <p>This content is invalid</p>
   |         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ remove content or use opening/closing tags
```

### Attribute Errors

```
Error: Duplicate attribute 'class'
  --> button.psx:4:35
   |
 4 |     <button class="btn" disabled class="primary">
   |             -----------          ^^^^^^^^^^^^^^^ duplicate attribute
   |             first defined here
```

### View-Specific Errors

```
Error: View parameter 'age' has no type annotation
  --> profile.psx:1:29
   |
 1 | view UserProfile(name: str, age):
   |                             ^^^ add type annotation
   |
help: specify a type like 'age: int'
```

### Slot Errors

```
Error: Named slot 'header' does not exist in view 'Layout'
  --> page.psx:15:13
   |
15 |         <h1 slot="header">Page Title</h1>
   |             ^^^^^^^^^^^^^ unknown slot
   |
note: Layout defines these slots:
  - title
  - content (default)
  - footer
```

## Warning System

Non-critical issues are reported as warnings:

```
Warning: Unused variable 'count'
  --> stats.psx:5:5
   |
 5 |     count = len(items)
   |     ^^^^^ variable is assigned but never used
```

Common warnings:
- Unused variables
- Shadowed names
- Deprecated syntax
- Performance hints

## IDE Integration

Error messages are designed for IDE integration:

### Machine-Readable Format

```json
{
  "type": "error",
  "code": "E001",
  "message": "Undefined variable 'user_name'",
  "file": "app.psx",
  "line": 8,
  "column": 22,
  "endLine": 8,
  "endColumn": 31,
  "severity": "error",
  "suggestions": ["username", "user_id"]
}
```

### Language Server Protocol

The compiler can output LSP-compatible diagnostics for real-time IDE feedback.

## Custom Error Types

### Runtime Errors

```go
type RuntimeError struct {
    Token   lexer.Token
    Message string
}
```

### Resolver Errors

```go
type ResolverError struct {
    Message string
    Line    int
    Column  int
}
```

## Best Practices for Error Messages

1. **Be Specific**: "Expected closing tag for 'div'" not "Invalid syntax"
2. **Show Context**: Include relevant source lines
3. **Suggest Fixes**: "did you mean 'username'?"
4. **Use Examples**: Show correct syntax when helpful
5. **Group Related**: Combine related errors (e.g., all undefined variables)

## Debugging Errors

### Verbose Mode

```bash
sylfie compile app.psx --debug
```

Shows additional information:
- Token stream for lexical errors
- AST dump for parsing errors
- Resolution table for semantic errors

### Error Codes

Each error has a unique code for documentation:

- `E001-E099`: Lexical errors
- `E100-E199`: Syntax errors  
- `E200-E299`: Semantic errors
- `E300-E399`: View-specific errors
- `E400-E499`: HTML/attribute errors
- `W001-W099`: Warnings

## Common Error Patterns

### Missing Imports

```
Error: Name 'Request' is not defined [E201]
  --> api.psx:5:23
   |
 5 | view Handler(request: Request):
   |                       ^^^^^^^ undefined type
   |
help: add 'from fastapi import Request' at the top of the file
```

### Type Mismatches

```
Error: Cannot iterate over type 'int' [E205]
  --> loop.psx:7:17
   |
 7 |     for item in count:
   |                 ^^^^^ 'count' is of type 'int', not iterable
   |
help: did you mean 'for i in range(count):'?
```

### HTML Context Errors

```
Error: Python statement not allowed inside HTML attribute [E401]
  --> form.psx:10:27
   |
10 |     <input value={if cond: "yes" else: "no"}>
   |                   ^^^^^^^^^^^^^^^^^^^^^^^^^^ use a variable or expression
   |
help: compute the value before the element:
      value = "yes" if cond else "no"
      <input value={value}>
```

## See Also

- [Architecture Guide](architecture.md) - Compiler pipeline details
- [CLI Reference](cli.md) - Debug options and flags
- [Testing Strategy](tests.md) - Error testing approach