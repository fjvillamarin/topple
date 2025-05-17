# Biscuit Compiler CLI Design

## Overview

The Biscuit compiler CLI transforms Biscuit template files (`.bsct`) into standard Python code. It provides a simple yet powerful interface for developers to compile their Biscuit views while offering various options for customization and optimization.

## Command Structure

```
biscuit [command] [options] [arguments]
```

### Main Commands

- `compile`: Compile Biscuit files to Python
- `watch`: Watch for changes and compile on the fly
- `check`: Validate Biscuit files without compilation
- `version`: Show version information
- `help`: Display help information

## Core Command: `compile`

```
biscuit compile [options] <input> [output]
```

### Arguments

- `input`: Path to a Biscuit file or directory (required)
- `output`: Output directory for compiled Python files (default: same as input)

### Options

- `-r, --recursive`: Process directories recursively
- `-o, --output <dir>`: Specify output directory (alternative to positional argument)
- `-p, --python-version <version>`: Target Python version (default: 3.9)
- `-i, --include <pattern>`: Only include files matching pattern (e.g., "views/*.bsct")
- `-e, --exclude <pattern>`: Exclude files matching pattern
- `-v, --verbose`: Enable verbose output
- `-q, --quiet`: Suppress all output except errors
- `--no-type-hints`: Omit type hints in the generated code
- `--minify`: Generate minimal code (remove comments, whitespace)
- `--source-map`: Generate source maps for debugging
- `--runtime <path>`: Path to custom runtime module (default: biscuit.py)
- `--imports <file>`: Path to a file defining custom imports
- `--ast`: Output the Biscuit AST instead of Python code
- `--dry-run`: Parse and validate but don't write files

## Watch Command

```
biscuit watch [options] <directory>
```

### Arguments

- `directory`: Directory to watch for changes (required)

### Options

- Same as `compile` command plus:
- `--delay <ms>`: Debounce delay in milliseconds (default: 300)
- `--clear`: Clear terminal on each compilation
- `--server`: Run a development server for previewing

## Check Command

```
biscuit check [options] <input>
```

### Arguments

- `input`: Path to a Biscuit file or directory (required)

### Options

- `-r, --recursive`: Process directories recursively
- `-i, --include <pattern>`: Only include files matching pattern
- `-e, --exclude <pattern>`: Exclude files matching pattern
- `-v, --verbose`: Enable verbose output

## Configuration Files

The CLI supports multiple configuration file formats to accommodate various project setups. Configuration files are searched in the following order (with later entries overriding earlier ones):

### 1. TOML (pyproject.toml)

```toml
[tool.biscuit]
input_dir = "./views"
output_dir = "./dist"
python_version = "3.10"
include = ["**/*.bsct"]
exclude = ["**/_*.bsct"]
no_type_hints = false
minify = false
source_maps = true
runtime = "./custom_runtime.py"
imports = "./custom_imports.py"

[tool.biscuit.fastapi]
mode = true
html_response = true
router = "app.router"

[tool.biscuit.optimization]
inline_single_use = true
merge_writers = true
static_analysis = true
```

### 2. YAML (.biscuit.yaml or .biscuit.yml)

```yaml
input_dir: "./views"
output_dir: "./dist"
python_version: "3.10"
include: 
  - "**/*.bsct"
exclude:
  - "**/_*.bsct"
no_type_hints: false
minify: false
source_maps: true
runtime: "./custom_runtime.py"
imports: "./custom_imports.py"

fastapi:
  mode: true
  html_response: true
  router: "app.router"

optimization:
  inline_single_use: true
  merge_writers: true
  static_analysis: true
```

### 3. JSON (.biscuit.json)

```json
{
  "input_dir": "./views",
  "output_dir": "./dist",
  "python_version": "3.10",
  "include": ["**/*.bsct"],
  "exclude": ["**/_*.bsct"],
  "no_type_hints": false,
  "minify": false,
  "source_maps": true,
  "runtime": "./custom_runtime.py",
  "imports": "./custom_imports.py",
  "fastapi": {
    "mode": true,
    "html_response": true,
    "router": "app.router"
  },
  "optimization": {
    "inline_single_use": true,
    "merge_writers": true,
    "static_analysis": true
  }
}
```

### 4. INI (setup.cfg)

```ini
[biscuit]
input_dir = ./views
output_dir = ./dist
python_version = 3.10
include = **/*.bsct
exclude = **/_*.bsct
no_type_hints = false
minify = false
source_maps = true
runtime = ./custom_runtime.py
imports = ./custom_imports.py

[biscuit.fastapi]
mode = true
html_response = true
router = app.router

[biscuit.optimization]
inline_single_use = true
merge_writers = true
static_analysis = true
```

### Configuration Loading Order

1. CLI arguments (highest priority)
2. Project-specific config file in current directory
3. User-specific config in home directory
4. Default built-in configuration

## Integration with FastAPI

Special flags for FastAPI integration:

```
--fastapi-mode: Generate code optimized for FastAPI
--html-response: Wrap all views with HTMLResponse
--router <module>: Specify router module for route declarations
```

## Optimization Options

```
--optimize: Enable all optimizations
--inline-single-use: Inline views that are used only once
--merge-writers: Merge adjacent writer calls
--static-analysis: Perform static analysis for optimizations
```

## Example Usage

```bash
# Compile a single file
biscuit compile view.bsct

# Compile a directory recursively
biscuit compile --recursive ./views ./dist

# Watch a directory for changes
biscuit watch --recursive ./views

# Validate files without compilation
biscuit check --recursive ./views

# Compile with optimizations for FastAPI
biscuit compile --recursive --fastapi-mode --html-response ./views

# Specify a different config file
biscuit compile --config ./configs/biscuit.yaml ./views
```

## Error Handling

The CLI provides clear error messages for common issues:

- Syntax errors in Biscuit files
- Missing required parameters
- Type errors in view declarations
- Undefined views in composition
- File access issues

Each error includes the file, line number, and column information when available.

## Python Integration

Since Biscuit is designed for Python projects, the CLI offers seamless integration:

- Respects Python package structure
- Auto-detection of virtual environments
- Compatibility with common Python development tools
- Support for poetry, pip, and other package managers
- Automatic import resolution based on Python module paths