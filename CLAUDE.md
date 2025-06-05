# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sylfie is a Python transpiler written in Go that compiles `.psx` (Python Sylfie eXtension) files to standard Python. It allows developers to write HTML-like syntax within Python code for creating dynamic web UIs, similar to JSX/TSX in the JavaScript ecosystem.

**Key Concepts:**
- Views are components that compile to Python classes inheriting from `BaseView`
- HTML elements become `el()` function calls with proper escaping
- Full Python syntax is supported within views
- The runtime provides caching to prevent unnecessary re-renders

## Key Architecture Components

### Compilation Pipeline
1. **Lexer** (`compiler/lexer/`) - Tokenizes PSX source code into tokens
2. **Parser** (`compiler/parser/`) - Builds Abstract Syntax Tree (AST) from tokens
3. **Resolver** (`compiler/resolver/`) - Resolves names, scopes, and types
4. **Transformers** (`compiler/transformers/`) - Transforms view syntax to Python classes
5. **Codegen** (`compiler/codegen/`) - Generates Python code from AST

### AST System
- All AST nodes are in `compiler/ast/`
- Each node type has an `Accept(visitor)` method for the visitor pattern
- Helper functions for node creation:
  - `ast.N("name")` - Name/identifier node
  - `ast.S("string")` - String literal
  - `ast.I(42)` - Integer literal
  - `ast.F(3.14)` - Float literal
  - `ast.B(true)` - Boolean literal
  - `ast.HElement("div", children...)` - HTML element
  - `ast.HView("ViewName", params, body...)` - View definition

## Essential Commands

### Building and Running
```bash
# Build the compiler
make build

# Compile PSX files recursively
make run

# Watch mode for development
make watch

# Start FastAPI dev server
make web
```

### Testing
```bash
# Run all unit tests
make test

# Run codegen tests specifically
make test-codegen

# Run view transformer tests
make test-view-transformer

# Update view transformer golden files
make test-view-transformer-update

# Run all golden file tests (e2e compiler tests)
make test-golden

# Update all golden files
make test-golden-update

# Run specific test category
make test-golden-category CATEGORY=literals

# Run single test
make test-golden-single TEST=literals/string

# Update single golden file
UPDATE_GOLDEN=1 go test ./compiler/codegen -run TestLiterals/string

# View available test categories
make test-golden-list

# Compare generated vs expected files
make test-golden-diff CATEGORY=literals
make test-golden-diff-single TEST=literals/string
```

### CLI Commands
```bash
# Compile a PSX file
sylfie compile file.psx

# Compile with custom output
sylfie compile file.psx -o output.py

# Compile directory recursively
sylfie compile src/ -r

# Watch for changes
sylfie watch src/

# Debug: show tokens
sylfie scan file.psx

# Debug: show AST
sylfie parse file.psx

# Debug: show AST with resolution info
sylfie parse file.psx -d
```

## Code Generation Standards

### Critical Rules
1. **PEP 8 Compliance**: Generated Python must follow PEP 8
   - Default parameters: `def foo(x=1)` NOT `def foo(x = 1)`
   - Single newline at EOF
   
2. **String Literals**: Must include quotes
   ```python
   "hello world"  # Correct
   hello world    # Wrong
   ```

3. **Empty Sets**: Use `set()` NOT `{}`
   ```python
   empty = set()  # Correct
   empty = {}     # Wrong - creates dict
   ```

4. **Testing Pattern**: Golden file structure
   - Input: `testdata/input/category/test.py`
   - Expected: `testdata/expected/category/test.py`
   - Generated: `testdata/generated/category/test.py`
   - Always write to `generated/` for comparison
   - Use `UPDATE_GOLDEN=1` to update expected files

5. **Import Organization**: Standard library → third-party → local imports

6. **Consistent Newlines**: All files must end with exactly one newline

### Known Bugs (from codegen_bugs.md)
1. **MultiStmt**: Generates `a\n; b\n; c` instead of `a; b; c` (High priority)
2. **Empty Set**: Generates `{}` instead of `set()` (Medium priority)  
3. **Default Parameters**: Adds spaces around `=` (Low priority - style issue)

## View Transformation Pattern

Views in PSX become Python classes:
```python
# Input (.psx)
view HelloWorld(name: str = "World"):
    <div>Hello, {name}!</div>

# Output (.py)
from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child

class HelloWorld(BaseView):
    def __init__(self, name: str = "World"):
        super().__init__()
        self.name = name
    
    def _render(self) -> Element:
        return el("div", f"Hello, {self.name}!")
```

**Key Transformation Rules:**
- View parameters become `__init__` parameters
- View body becomes `_render()` method returning Element
- HTML elements use `el(tag, attrs, children)` function
- Text content is automatically escaped
- Interpolations `{expr}` become f-string expressions

## Development Workflow

1. **Always run tests** before submitting changes
2. **Update golden files** when changing expected output: `UPDATE_GOLDEN=1`
3. **Check generated files** in `testdata/generated/` to verify output
4. **Follow existing patterns** - the codebase has consistent patterns for:
   - Visitor methods: `VisitNodeType(node *ast.NodeType) string`
   - Error handling: Return empty string on errors
   - Test structure: Table-driven tests with clear test names
   - Helper functions for common operations

### Common Development Tasks

```bash
# When adding a new AST node type:
1. Define the node in compiler/ast/
2. Add Accept() method for visitor pattern
3. Add parsing logic in compiler/parser/
4. Add VisitNodeType() in codegen and transformers
5. Add tests with golden files

# When fixing a codegen bug:
1. Add failing test case first
2. Fix the Visit method in codegen.go
3. Run UPDATE_GOLDEN=1 to update expected output
4. Verify the generated Python is correct
```

## File Types
- `.psx` - Source files with view syntax
- `.py` - Generated Python output
- `.ast` - AST debug output (from `parse` command)
- `.tok` - Token debug output (from `scan` command)

## Project Structure

```
compiler/
├── ast/          # AST node definitions
├── lexer/        # Tokenization (scanner.go, token.go)
├── parser/       # Parsing logic for all Python/PSX constructs
├── resolver/     # Name resolution and scope analysis
├── transformers/ # View → Python class transformation
├── codegen/      # Python code generation
│   └── testdata/ # Golden file tests
└── e2e_test.go   # End-to-end compiler tests
```

## Test Categories

The golden file tests cover these categories:
- `literals` - String, int, float, bool literals
- `expressions` - Binary ops, calls, comprehensions
- `statements` - Assignments, control flow
- `functions` - Function definitions with various features
- `classes` - Class definitions
- `control_flow` - if/for/while/match statements
- `collections` - Lists, dicts, sets, tuples
- `patterns` - Match statement patterns
- `comprehensions` - List/dict/set comprehensions
- `fstring` - F-string expressions

## Current Status
- ✅ Core Python features fully supported
- ✅ Basic view transformation working
- ✅ HTML element generation
- ✅ Expression interpolation
- ⚠️ Codegen has several known bugs (see above)
- 🚧 Advanced view features in progress (components, slots)

## Important Notes

1. **Runtime Dependency**: Generated code requires `runtime.py` to be available
2. **Go Version**: Requires Go 1.23+ for building
3. **Python Version**: Generated code targets Python 3.12+
4. **Branch**: Active development on `feat/codegen-refactor`