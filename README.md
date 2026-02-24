# Topple

A Python transpiler that compiles PSX (Python Syntax eXtension) files to standard Python. Write HTML-like syntax within Python code for creating dynamic web UIs, similar to JSX/TSX in the JavaScript ecosystem.

## Project Setup

Add Topple to any Python project with [mise](https://mise.jdx.dev/) and [Poetry](https://python-poetry.org/):

**`.mise.toml`** — add the compiler:

```toml
[tools]
"go:github.com/fjvillamarin/topple/cmd/topple" = "latest"
```

**`pyproject.toml`** — add the runtime:

```toml
[tool.poetry.dependencies]
topple = "^0.1.0"
```

Then install both:

```bash
mise install        # installs the topple CLI
poetry install      # installs the Python runtime
```

### Alternative Installation

If you don't use mise/Poetry, you can install each component directly:

```bash
# Compiler (requires Go 1.23+)
go install github.com/fjvillamarin/topple/cmd/topple@latest

# Runtime
pip install topple
```

Pre-built binaries are also available from [GitHub Releases](https://github.com/fjvillamarin/topple/releases).

## Quick Start

Create a `.psx` file:

```python
# hello.psx
view HelloWorld(name: str = "World"):
    <div>Hello, {name}!</div>
```

Compile it:

```bash
topple compile hello.psx -o hello.py
```

Use it in your Python app:

```python
from hello import HelloWorld

view = HelloWorld(name="Topple")
print(view.render())
# <div>Hello, Topple!</div>
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
