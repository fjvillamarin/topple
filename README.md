# Topple

A Python transpiler that compiles PSX (Python Syntax eXtension) files to standard Python. Write HTML-like syntax within Python code for creating dynamic web UIs, similar to JSX/TSX in the JavaScript ecosystem.

## Installation

### Compiler (Go binary)

Install with `go install`:

```bash
go install github.com/fjvillamarin/topple/cmd/topple@latest
```

Or with [mise](https://mise.jdx.dev/):

```bash
mise use go:github.com/fjvillamarin/topple/cmd/topple
```

Or download a pre-built binary from [GitHub Releases](https://github.com/fjvillamarin/topple/releases).

### Python Runtime

The generated Python code requires the Topple runtime:

```bash
pip install topple
```

## Quick Start

```python
# hello.psx
view HelloWorld(name: str = "World"):
    <div>Hello, {name}!</div>
```

```bash
topple compile hello.psx -o hello.py
```

## Known Issues

**Last Updated**: 2026-02-06

## Active Issues (1)

### High Priority

1. **[Issue #67: Parser: Add async view statement support](https://github.com/fjvillamarin/topple/issues/67)**
   - **Status**: Active, needs parser support for `async view`
   - **Repro**: `topple parse` on a file that starts with `async view ...` returns `at 'async': unexpected token`
   - **Impact**: Async view declarations are currently unsupported

## Recently Resolved

- **[Issue #68: Parser: Fix complex target parsing edge cases](https://github.com/fjvillamarin/topple/issues/68)** now passes current parser tests.
- Historical codegen/parser bugs previously tracked in this repo have been fixed or are no longer applicable to the current architecture.

## Verification Snapshot

- `go test ./...` passes as of 2026-02-06.

## See Also

- [BUG_STATUS_REPORT.md](BUG_STATUS_REPORT.md) - Current bug status summary
- [docs/development/codegen_bugs.md](docs/development/codegen_bugs.md) - Codegen status and remaining quality notes
- [docs/development/parser_bugs.md](docs/development/parser_bugs.md) - Parser status and remaining open parser issue
