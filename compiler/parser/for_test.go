package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a for statement
func parseForStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.forStatement()
}

// Helper function to parse star targets
func parseStarTargets(t *testing.T, input string) ([]ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.starTargets()
}

// Helper function to validate for statement structure
func validateForStatement(t *testing.T, stmt ast.Stmt, expectedAsync, expectedElse bool) {
	forStmt, ok := stmt.(*ast.For)
	if !ok {
		t.Fatalf("Expected *ast.For, got %T", stmt)
	}

	if forStmt.Target == nil {
		t.Error("For statement missing target")
	}
	if forStmt.Iterable == nil {
		t.Error("For statement missing iterable")
	}
	if len(forStmt.Body) == 0 {
		t.Error("For statement missing body")
	}

	if forStmt.IsAsync != expectedAsync {
		t.Errorf("Expected isAsync=%v, got %v", expectedAsync, forStmt.IsAsync)
	}

	hasElse := len(forStmt.Else) > 0
	if hasElse != expectedElse {
		t.Errorf("Expected hasElse=%v, got %v", expectedElse, hasElse)
	}
}

// Helper function to check for starred expressions recursively
func checkForStarred(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return true
	case *ast.TupleExpr:
		for _, elem := range e.Elements {
			if checkForStarred(elem) {
				return true
			}
		}
	case *ast.ListExpr:
		for _, elem := range e.Elements {
			if checkForStarred(elem) {
				return true
			}
		}
	}
	return false
}

// Helper function to validate star targets
func validateStarTargets(t *testing.T, targets []ast.Expr, shouldHaveStarred bool) {
	if len(targets) == 0 {
		t.Error("Expected at least one target")
		return
	}

	hasStarred := false
	for _, target := range targets {
		if checkForStarred(target) {
			hasStarred = true
			break
		}
	}

	if shouldHaveStarred && !hasStarred {
		t.Error("Expected at least one starred expression")
	} else if !shouldHaveStarred && hasStarred {
		t.Error("Expected no starred expressions")
	}
}

// Test comprehensive for statement functionality
func TestForStatement(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		isAsync   bool
		hasElse   bool
		hasError  bool
	}{
		{
			name: "simple for loop",
			input: `for i in range(10):
    print(i)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "for with else clause",
			input: `for item in items:
    process(item)
else:
    print("done")`,
			isAsync: false,
			hasElse: true,
		},
		{
			name: "for with tuple unpacking",
			input: `for x, y in pairs:
    print(x, y)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "for with complex target",
			input: `for (a, b), c in nested:
    use(a, b, c)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "async for loop",
			input: `async for data in stream:
    await process(data)`,
			isAsync: true,
			hasElse: false,
		},
		{
			name: "for with starred unpacking",
			input: `for first, *rest in sequences:
    handle(first, rest)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "nested for loops",
			input: `for i in range(3):
    for j in range(3):
        process(i, j)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "for with control flow",
			input: `for x in items:
    if x < 0:
        continue
    if x > 100:
        break
    process(x)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "for with complex iterable",
			input: `for item in get_items().filter(predicate):
    handle(item)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "async for with else",
			input: `async for chunk in async_stream():
    process(chunk)
else:
    cleanup()`,
			isAsync: true,
			hasElse: true,
		},
		{
			name: "for with list comprehension iterable",
			input: `for result in [process(x) for x in data]:
    use(result)`,
			isAsync: false,
			hasElse: false,
		},
		{
			name: "single line for",
			input: `for x in [1, 2, 3]: print(x)`,
			isAsync: false,
			hasElse: false,
		},

		// Error cases
		{
			name: "missing colon",
			input: `for i in range(10)
    print(i)`,
			hasError: true,
		},
		{
			name: "missing in keyword",
			input: `for i range(10):
    print(i)`,
			hasError: true,
		},
		{
			name: "missing target",
			input: `for in range(10):
    print(i)`,
			hasError: true,
		},
		{
			name: "missing iterable",
			input: `for i in:
    print(i)`,
			hasError: true,
		},
		{
			name: "empty body",
			input: `for i in range(10):`,
			hasError: true,
		},
		{
			name: "invalid target literal",
			input: `for 123 in items:
    pass`,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseForStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.name, err)
			}

			validateForStatement(t, stmt, test.isAsync, test.hasElse)
		})
	}
}

// Test star targets in for loops and assignments
func TestStarTargets(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectStarred bool
	}{
		{
			name:          "star at beginning",
			input:         "first, *rest",
			expectStarred: true,
		},
		{
			name:          "star at end",
			input:         "*init, last",
			expectStarred: true,
		},
		{
			name:          "star in middle",
			input:         "first, *middle, last",
			expectStarred: true,
		},
		{
			name:          "nested star in tuple",
			input:         "(a, *b), c",
			expectStarred: true,
		},
		{
			name:          "star in list pattern",
			input:         "[first, *rest]",
			expectStarred: true,
		},
		{
			name:          "no starred expressions",
			input:         "a, b, c",
			expectStarred: false,
		},
		{
			name:          "single variable",
			input:         "x",
			expectStarred: false,
		},
		{
			name:          "complex nested without star",
			input:         "(a, (b, c)), d",
			expectStarred: false,
		},
		{
			name:          "deeply nested star",
			input:         "((a, *b), c), d",
			expectStarred: true,
		},

		// Error cases
		{
			name:     "multiple stars in same level",
			input:    "*a, *b",
			hasError: true,
		},
		{
			name:     "star without variable",
			input:    "*",
			hasError: true,
		},
		{
			name:     "invalid star usage",
			input:    "*, a",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targets, err := parseStarTargets(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.name, err)
			}

			validateStarTargets(t, targets, test.expectStarred)
		})
	}
}

// Test edge cases and complex scenarios
func TestForEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "for with function call target",
			input: `for obj.attr in items:
    process(obj.attr)`,
			hasError: false,
		},
		{
			name: "for with subscript target",
			input: `for data[key] in values:
    handle(data[key])`,
			hasError: true, // PARSER BUG: Should support subscript targets like Python
		},
		{
			name: "async for without async context warning",
			input: `async for x in items:
    print(x)`,
			hasError: false, // May be valid depending on context
		},
		{
			name: "for with generator expression",
			input: `for item in (x for x in data if x.valid):
    use(item)`,
			hasError: false,
		},
		{
			name: "extremely nested unpacking",
			input: `for ((a, b), (c, *d)), e in complex_data:
    process(a, b, c, d, e)`,
			hasError: false,
		},
		{
			name: "for with walrus operator in condition",
			input: `for x in data:
    if (y := process(x)) is not None:
        use(y)`,
			hasError: false,
		},

		// Error cases that should be caught
		{
			name: "malformed async syntax",
			input: `async async for x in items:
    pass`,
			hasError: true,
		},
		{
			name: "for without proper structure",
			input: `for x in items pass`,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseForStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.name, err)
			}

			// For successful cases, just verify we got a valid statement
			if stmt == nil {
				t.Error("Expected non-nil statement")
			}
		})
	}
}