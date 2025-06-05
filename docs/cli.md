# Sylfie CLI Reference

## Overview

The Sylfie compiler provides a command-line interface for transforming `.psx` (Python Syntax eXtended) files into standard Python code. It offers various commands for compilation, development, and debugging.

## Installation

```bash
# Build from source
mise run build

# The binary will be in bin/sylfie
# Add to PATH or copy to system location:
cp bin/sylfie /usr/local/bin/
```

## Command Structure

```bash
sylfie <command> [options] [arguments]
```

## Commands

### compile

Compile PSX files to Python.

```bash
sylfie compile [options] <input>
```

**Arguments:**
- `input`: Path to a .psx file or directory

**Options:**
- `-o, --output <path>`: Output file or directory (default: same location as input)
- `-r, --recursive`: Process directories recursively
- `--debug`: Enable debug output

**Examples:**
```bash
# Compile a single file
sylfie compile hello.psx

# Compile to specific output
sylfie compile hello.psx -o hello.py

# Compile directory recursively
sylfie compile src/ -r

# Compile with custom output directory
sylfie compile src/ -o dist/ -r
```

### watch

Watch files for changes and recompile automatically.

```bash
sylfie watch [options] <paths...>
```

**Arguments:**
- `paths`: One or more paths to watch (files or directories)

**Options:**
- `-o, --output <dir>`: Output directory for compiled files
- `--debug`: Enable debug output

**Examples:**
```bash
# Watch a single file
sylfie watch hello.psx

# Watch multiple directories
sylfie watch src/ templates/ components/

# Watch with custom output
sylfie watch src/ -o dist/
```

### scan

Tokenize a file and display the token stream (for debugging).

```bash
sylfie scan [options] <input>
```

**Arguments:**
- `input`: Path to a .psx file

**Options:**
- `--debug`: Enable debug output

**Example:**
```bash
# Display tokens for a file
sylfie scan hello.psx
```

### parse

Parse a file and display the AST (for debugging).

```bash
sylfie parse [options] <input>
```

**Arguments:**
- `input`: Path to a .psx file

**Options:**
- `--debug`: Enable debug output

**Example:**
```bash
# Display AST for a file
sylfie parse hello.psx
```

## File Extensions

- `.psx`: Sylfie source files (Python Syntax eXtended)
- `.py`: Generated Python output files

## Output Format

The compiler generates standard Python code with:
- Required runtime imports
- View classes extending `BaseView`
- Type annotations preserved from source
- Automatic HTML escaping for security

## Error Handling

The compiler provides detailed error messages including:
- File path and location (line:column)
- Error type and description
- Source code context
- Suggestions for common issues

Example error output:
```
Error in hello.psx:5:12
SyntaxError: Unclosed HTML tag 'div'
    3 | view HelloWorld():
    4 |     <div>
    5 |         <h1>Hello
                    ^
    6 |     </div>
```

## Development Workflow

### Basic Development

1. Create `.psx` files with your views
2. Run `sylfie compile` to generate Python
3. Import and use the generated classes

### Watch Mode Development

1. Start watch mode: `sylfie watch src/`
2. Edit `.psx` files in your editor
3. Files are automatically recompiled on save
4. See compilation errors in terminal

### Integration with Build Tools

#### Mise Tasks
```toml
# In .mise.toml
[tasks.compile-views]
description = "Compile all PSX views"
run = "sylfie compile src/ -r"

[tasks.dev]
description = "Start development mode"
run = "sylfie watch src/"

[tasks.clean-views]
description = "Clean generated Python files"
run = 'find src -name "*.py" -type f | grep -E "view.*\.py$" | xargs rm -f'
```

#### Poetry/PyProject
```toml
[tool.poetry.scripts]
build-views = "sylfie compile src/ -r"
dev = "sylfie watch src/"
```

## Performance Considerations

- The compiler is written in Go for fast compilation
- Watch mode uses efficient file system monitoring
- Only changed files are recompiled
- Compilation is typically < 100ms for most files

## Debugging

### Token Analysis
Use `sylfie scan` to debug lexical issues:
```bash
sylfie scan problematic.psx > tokens.txt
```

### AST Analysis
Use `sylfie parse` to debug parsing issues:
```bash
sylfie parse problematic.psx > ast.txt
```

### Verbose Output
Add `--debug` to any command for detailed logging:
```bash
sylfie compile hello.psx --debug
```

## Environment Variables

- `SYLFIE_DEBUG`: Set to "1" to enable debug output globally
- `SYLFIE_RUNTIME`: Path to custom runtime module

## Exit Codes

- `0`: Success
- `1`: General error (file not found, permissions, etc.)
- `2`: Syntax error in source file
- `3`: Compilation error
- `4`: I/O error

## See Also

- [Language Grammar](grammar_psx.md) - PSX syntax reference
- [Architecture](architecture.md) - Compiler internals
- [Examples](../examples/) - Example applications