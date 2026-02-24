# Topple Compiler Error Handling

This document describes how the Topple compiler handles and reports errors throughout the compilation pipeline.

## Error Types

### Lexical Errors

Errors detected during tokenization:
- Unterminated strings
- Invalid numeric literals
- Illegal characters
- Mixed indentation (tabs vs spaces)
- Invalid escape sequences

### Syntax Errors

Errors detected during parsing:
- Missing colons after statements
- Unmatched brackets/parentheses
- Invalid view syntax
- Malformed HTML elements
- Missing required keywords
- Unclosed HTML tags

### Resolution Errors

Errors during name resolution:
- Undefined view components referenced in HTML elements
- Invalid global/nonlocal usage

### Transformation Errors

Errors during view transformation:
- View elements with unsupported nested content (e.g., template slots on view elements)

## Error Recovery

The compiler attempts to recover from errors to provide multiple diagnostics. The parser synchronizes at statement boundaries and block delimiters to continue parsing after encountering errors.

## Error Message Format

Errors include the file path and a description of the issue:

```
file.psx: at 'token': error description
```

For compilation errors in views, the error is included in the generated output:

```python
# COMPILATION_ERRORS: [error description]
```

## Debugging

### Token Analysis
Use `topple scan` to debug lexical issues:
```bash
topple scan problematic.psx
```

### AST Analysis
Use `topple parse` to debug parsing issues:
```bash
topple parse problematic.psx
```

### Debug Output
Add `-d` to any command for detailed logging:
```bash
topple compile hello.psx -d
```

## See Also

- [Architecture Guide](architecture.md) - Compiler pipeline details
- [CLI Reference](cli.md) - Debug options and flags
- [Testing Strategy](tests.md) - Error testing approach
