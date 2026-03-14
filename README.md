# Topple

A Python transpiler that compiles PSX (Python Syntax eXtension) files to standard Python. Write HTML-like syntax within Python code for creating dynamic web UIs, similar to JSX/TSX in the JavaScript ecosystem.

## Prerequisites

- **Python 3.12+**
- **[mise](https://mise.jdx.dev/)** (recommended) or **Go 1.23+** for installing the compiler
- **[Poetry](https://python-poetry.org/)** (recommended) or **pip** for managing Python dependencies

Topple has two components you need to install:

| Component | What it does | How it's distributed |
|-----------|-------------|---------------------|
| **Compiler** (`topple` CLI) | Compiles `.psx` files to `.py` | Go binary via mise, `go install`, or GitHub Releases |
| **Runtime** (`topple` package) | Provides `BaseView`, `el()`, etc. at runtime | Python package via PyPI |

## Installation

### Recommended: mise + Poetry

**1. Add the compiler** to your project's `.mise.toml`:

```toml
[tools]
"go:github.com/fjvillamarin/topple/cmd/topple" = "latest"
```

**2. Add the runtime** to your `pyproject.toml`:

```toml
[tool.poetry.dependencies]
topple = "^0.1.0"
```

**3. Install both:**

```bash
mise install        # installs the topple CLI
poetry install      # installs the Python runtime
```

Verify the installation:

```bash
topple --version
python -c "import topple; print(topple.__version__)"
```

### Alternative: go install + pip

If you don't use mise/Poetry:

```bash
# Compiler (requires Go 1.23+)
go install github.com/fjvillamarin/topple/cmd/topple@latest

# Runtime
pip install topple
```

### Alternative: Pre-built binaries

Download the compiler binary for your platform from [GitHub Releases](https://github.com/fjvillamarin/topple/releases), then install the runtime:

```bash
pip install topple
```

## Quick Start

**1. Create a `.psx` file:**

```python
# hello.psx
view HelloWorld(name: str = "World"):
    <div>Hello, {name}!</div>
```

**2. Compile it:**

```bash
topple compile hello.psx -o hello.py
```

**3. Use it in your Python app:**

```python
from hello import HelloWorld

view = HelloWorld(name="Topple")
print(view.render())
# <div>Hello, Topple!</div>
```

## Usage

### Project Structure

A typical project using Topple looks like this:

```
my-app/
├── .mise.toml          # Compiler tool config
├── pyproject.toml      # Python dependencies (includes topple runtime)
├── views/
│   ├── components.psx  # PSX source files
│   ├── components.py   # Compiled output (generated, git-ignored)
│   ├── pages.psx
│   └── pages.py
└── main.py             # Your application
```

### Compiling

```bash
# Compile a single file
topple compile views/pages.psx -o views/pages.py

# Compile a directory recursively
topple compile views/ -r

# Watch for changes during development
topple watch views/ -r
```

When compiling a directory with `-r`, Topple writes each `.py` file next to its `.psx` source.

### Using with FastAPI

```python
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from views.pages import LandingPage

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
async def home():
    page = LandingPage()
    return page.render()
```

### CLI Reference

| Command | Description |
|---------|-------------|
| `topple compile <file\|dir>` | Compile PSX to Python |
| `topple compile <src> -o <out>` | Compile with custom output path |
| `topple compile <dir> -r` | Compile directory recursively |
| `topple watch <dir> -r` | Watch and recompile on changes |
| `topple scan <file>` | Debug: show token stream |
| `topple parse <file>` | Debug: show AST |
| `topple parse <file> -d` | Debug: show AST with resolution info |

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
