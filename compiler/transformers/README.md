# View Transformer Tests

This directory contains tests for the Sylfie view transformer that follows the same pattern as the codegen tests.

## Test Structure

```
testdata/
├── input/          # Go files that return ViewStmt AST nodes
│   ├── basic/
│   ├── html/
│   └── slots/
├── expected/       # Expected Python output
│   └── ... (same structure as input)
└── generated/      # Generated Python output (gitignored)
    └── ... (same structure as input)
```

## Running Tests

```bash
# Run all tests
go test -v

# Update golden files
UPDATE_GOLDEN=1 go test -v

# Run specific test
go test -v -run TestBasicViewTransformation
```

## Adding New Tests

1. Create an input file in `testdata/input/category/test_name.go`:
```go
package category

import "sylfie/compiler/ast"

func GetAST() ast.Node {
    return &ast.ViewStmt{
        // Build your ViewStmt AST here
    }
}
```

2. Add the test case to the appropriate test function:
```go
{
    name:     "test_name",
    category: "category",
    testFile: "test_name",
}
```

3. Add the AST builder to `view_ast_builders.go`:
```go
"category/test_name": func() ast.Node {
    // Return the same AST as in the input file
},
```

4. Run with `UPDATE_GOLDEN=1` to create the expected output:
```bash
UPDATE_GOLDEN=1 go test -v -run test_name
```

5. Review the generated file in `testdata/expected/category/test_name.py`

## Test Categories

- **basic** - Simple views, parameters, type annotations
- **html** - HTML elements, attributes, nesting
- **slots** - Default slots, named slots, fallback content
- **control_flow** - If/for/while statements in views
- **composition** - Views using other views
- **edge_cases** - Empty views, multiple roots, etc.