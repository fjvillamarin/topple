# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Topple is a Python transpiler written in Go that compiles `.psx` (Python Syntax eXtension) files to standard Python. It allows developers to write HTML-like syntax within Python code for creating dynamic web UIs, similar to JSX/TSX in the JavaScript ecosystem.

**PSX** is the syntax specification for embedding HTML-like markup in Python code. **Topple** is the reference implementation (compiler + runtime + tooling) for PSX.

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

The project uses `mise` for task management.

### Building and Running
```bash
# Build the compiler
mise run build

# Compile PSX files recursively
mise run run

# Watch mode for development
mise run watch

# Start FastAPI dev server
mise run web

# Format Go code
mise run fmt

# View all available tasks
mise run help
```

### Testing
```bash
# Run all unit tests
mise run test

# Run codegen tests specifically
mise run test-codegen

# Run view transformer tests
mise run test-view-transformer

# Update view transformer golden files
mise run test-view-transformer-update

# Run all golden file tests (e2e compiler tests)
mise run test-golden

# Update all golden files
mise run test-golden-update

# Run specific test category
mise run test-golden-category CATEGORY=literals

# Run single test
mise run test-golden-single TEST=literals/string

# Update single golden file
UPDATE_GOLDEN=1 mise run test-golden-single TEST=literals/string

# View available test categories
mise run test-golden-list

# Compare generated vs expected files
mise run test-golden-diff CATEGORY=literals
mise run test-golden-diff-single TEST=literals/string
```

### CLI Commands
```bash
# Compile a PSX file
topple compile file.psx

# Compile with custom output
topple compile file.psx -o output.py

# Compile directory recursively
topple compile src/ -r

# Watch for changes
topple watch src/

# Debug: show tokens
topple scan file.psx

# Debug: show AST
topple parse file.psx

# Debug: show AST with resolution info
topple parse file.psx -d
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
from topple.psx import BaseView, el, escape, Element, FragmentElement, fragment, render_child

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

### Working on GitHub Issues

When working on a GitHub issue (e.g., #22), follow these steps:

1. **Fetch the issue details** using GitHub CLI:
   ```bash
   gh issue view 22
   ```

2. **Create a branch** following the naming convention:
   - For bugs: `bug/issue-22`
   - For features: `feat/issue-22`
   - For documentation: `docs/issue-22`
   ```bash
   # Example for a bug fix
   git checkout -b bug/issue-22
   
   # Example for a feature
   git checkout -b feat/issue-22
   ```

3. **Create a todo list** to track your progress:
   - Use the TodoWrite tool to create a comprehensive list of tasks
   - Break down the issue into specific, actionable items
   - Include tasks for:
     - Understanding the issue and examining relevant code
     - Implementing the fix or feature
     - Writing/updating tests
     - Running tests and checking for regressions
     - Code formatting
     - Committing changes
     - Creating the pull request
   - Mark tasks as `in_progress` when you start them
   - Mark tasks as `completed` immediately after finishing them
   - Add new tasks as you discover additional work needed

4. **Work on the issue**:
   - Make incremental commits as you progress
   - Ensure all existing tests continue to pass
   - Add new tests if implementing a feature or fixing a bug
   - Follow the existing code style and patterns

5. **Testing requirements**:
   ```bash
   # Run all tests before committing
   mise run test
   
   # Run specific test categories related to your changes
   mise run test-golden-category CATEGORY=<relevant-category>
   
   # Update golden files if output changes are intentional
   mise run test-golden-update
   
   # Debug by running compiler on individual files
   topple compile path/to/test.psx        # See compiled output
   topple scan path/to/test.psx          # See token stream
   topple parse path/to/test.psx         # See AST structure
   topple parse path/to/test.psx -d      # See AST with resolution info
   ```

6. **Code formatting**:
   - **IMPORTANT**: Always run `mise run fmt` before committing Go code
   - This ensures consistent formatting across the codebase
   - The formatter will fix indentation, spacing, and add missing newlines

7. **Commit guidelines**:
   - Use conventional commit format: `type(scope): description` or `type: description`
     - `feat(parser): add support for async views`
     - `fix(codegen): handle empty set literals correctly`
     - `docs: update README with mise instructions`
     - `test(transformers): add tests for slot handling`
   - Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`
   - Common scopes: `parser`, `codegen`, `lexer`, `transformer`, `resolver`
   - Reference the issue number when relevant: `fix(parser): handle nested f-strings (#22)`
   - Ensure each commit represents a working state

8. **Commit your changes**:
   - Stage and commit your changes following the guidelines above
   - Make sure all tests pass before committing
   - Example:
     ```bash
     git add -A
     git commit -m "fix(parser): handle nested f-strings correctly (#22)"
     ```
   - Update your todo list to mark the commit task as completed

9. **Before pushing**:
   - Run the full test suite: `mise run test-all`
   - Verify no existing tests are broken
   - Check that new tests cover your changes
   - Run code formatting: `mise run fmt`
   - Review your changes: `git diff master`
   - Double-check all modified files make sense for the issue

10. **Creating the Pull Request**:
   ```bash
   # Push your branch
   git push -u origin bug/issue-22
   
   # Create PR using GitHub CLI
   gh pr create --title "fix(parser): handle nested f-strings correctly" \
                --body "Fixes #22\n\n## Summary\n- Fixed parser to handle nested f-strings\n- Added comprehensive tests\n\n## Test plan\n- Added golden file tests\n- All existing tests pass"
   ```
   - Use a meaningful PR title following the commit format
   - Reference the issue in the PR body (e.g., "Fixes #22")
   - Include a summary of changes
   - Describe the test plan
   - Let GitHub Actions run the full test suite

### General Development Guidelines

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

# When adding tests:
1. Follow existing test patterns in the relevant _test.go files
2. Use table-driven tests for multiple test cases
3. For golden file tests, add input files in testdata/input/
4. Run the test with UPDATE_GOLDEN=1 to generate expected output
5. Verify the expected output is correct before committing
```

### Testing Best Practices

1. **Unit Tests**: Follow the existing patterns in `*_test.go` files
   - Use descriptive test names that explain what is being tested
   - Group related tests using subtests (`t.Run()`)
   - Use table-driven tests for testing multiple scenarios

2. **Golden File Tests**: For end-to-end compiler tests
   - Place input `.psx` files in `compiler/testdata/input/<category>/`
   - Run `mise run test-golden-category-update CATEGORY=<category>` to generate expected output
   - Always verify the generated output is correct before committing

3. **Test Coverage**: Ensure your changes are covered by tests
   - Add tests for both success and error cases
   - Test edge cases and boundary conditions
   - If fixing a bug, add a test that would have caught the bug

## File Types
- `.psx` - Source files with view syntax
- `.py` - Generated Python output
- `.ast` - AST debug output (from `parse` command)
- `.tok` - Token debug output (from `scan` command)

## Project Structure

```
topple/
├── cmd/          # CLI commands (compile, watch, scan, parse)
├── compiler/
│   ├── ast/          # AST node definitions
│   ├── lexer/        # Tokenization (scanner.go, token.go)
│   ├── parser/       # Parsing logic for all Python/PSX constructs
│   ├── resolver/     # Name resolution and scope analysis
│   ├── transformers/ # View → Python class transformation
│   ├── codegen/      # Python code generation
│   └── testdata/     # Golden file tests
├── topple/       # Python runtime package
│   └── psx/      # PSX runtime module (BaseView, el, etc.)
└── .mise.toml    # Task runner configuration
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

1. **Runtime Dependency**: Generated code requires the Topple runtime package (`pip install topple`)
   - Imports from `topple.psx` (e.g., `from topple.psx import BaseView, el`)
2. **Go Version**: Requires Go 1.23+ for building
3. **Python Version**: Generated code targets Python 3.12+