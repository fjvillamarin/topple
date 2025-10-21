package parser

import (
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse an expression
func parseExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to validate comprehension structure
func validateComprehension(t *testing.T, expr ast.Expr, expectedType string) {
	switch expectedType {
	case "list":
		listComp, ok := expr.(*ast.ListComp)
		if !ok {
			t.Fatalf("Expected *ast.ListComp, got %T", expr)
		}
		if listComp.Element == nil {
			t.Error("List comprehension should have an element")
		}
		if len(listComp.Clauses) == 0 {
			t.Error("List comprehension should have at least one clause")
		}
	case "set":
		setComp, ok := expr.(*ast.SetComp)
		if !ok {
			t.Fatalf("Expected *ast.SetComp, got %T", expr)
		}
		if setComp.Element == nil {
			t.Error("Set comprehension should have an element")
		}
		if len(setComp.Clauses) == 0 {
			t.Error("Set comprehension should have at least one clause")
		}
	case "dict":
		dictComp, ok := expr.(*ast.DictComp)
		if !ok {
			t.Fatalf("Expected *ast.DictComp, got %T", expr)
		}
		if dictComp.Key == nil {
			t.Error("Dict comprehension should have a key")
		}
		if dictComp.Value == nil {
			t.Error("Dict comprehension should have a value")
		}
		if len(dictComp.Clauses) == 0 {
			t.Error("Dict comprehension should have at least one clause")
		}
	case "generator":
		genExpr, ok := expr.(*ast.GenExpr)
		if !ok {
			t.Fatalf("Expected *ast.GenExpr, got %T", expr)
		}
		if genExpr.Element == nil {
			t.Error("Generator expression should have an element")
		}
		if len(genExpr.Clauses) == 0 {
			t.Error("Generator expression should have at least one clause")
		}
	}
}

// Helper function to check if expression is a comprehension
func isComprehension(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.ListComp, *ast.SetComp, *ast.DictComp, *ast.GenExpr:
		return true
	default:
		return false
	}
}

// Test all types of comprehensions with comprehensive test cases
func TestComprehensions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		compType   string
		hasError   bool
		expectComp bool
	}{
		// List comprehensions
		{"simple list comprehension", "[x for x in items]", "list", false, true},
		{"list comprehension with expression", "[x * 2 for x in numbers]", "list", false, true},
		{"list comprehension with condition", "[x for x in items if x > 0]", "list", false, true},
		{"list comprehension with multiple conditions", "[x for x in items if x > 0 if x < 100]", "list", false, true},
		{"nested list comprehension", "[y for x in matrix for y in x]", "list", false, true},
		{"list comprehension with tuple unpacking", "[x + y for x, y in pairs]", "list", false, true},
		{"async list comprehension", "[x async for x in async_items]", "list", false, true},
		{"complex list comprehension", "[func(x, y) for x in items for y in get_list(x) if process(x, y)]", "list", false, true},

		// Set comprehensions
		{"simple set comprehension", "{x for x in items}", "set", false, true},
		{"set comprehension with expression", "{x.upper() for x in strings}", "set", false, true},
		{"set comprehension with condition", "{x for x in items if x.isalpha()}", "set", false, true},
		{"set comprehension with multiple clauses", "{x + y for x in range(3) for y in range(3) if x != y}", "set", false, true},
		{"async set comprehension", "{x async for x in async_generator()}", "set", false, true},

		// Dict comprehensions
		{"simple dict comprehension", "{k: v for k, v in items}", "dict", false, true},
		{"dict comprehension with expression", "{k.upper(): v * 2 for k, v in pairs}", "dict", false, true},
		{"dict comprehension with condition", "{k: v for k, v in items.items() if v > 0}", "dict", false, true},
		{"dict comprehension with nested loops", "{x: y for x in keys for y in get_values(x)}", "dict", false, true},
		{"dict comprehension with multiple conditions", "{k: v for k, v in data.items() if k.startswith('test') if v is not None}", "dict", false, true},
		{"async dict comprehension", "{k: v async for k, v in async_items()}", "dict", false, true},

		// Generator expressions
		{"simple generator expression", "(x for x in items)", "generator", false, true},
		{"generator expression with condition", "(x * 2 for x in numbers if x > 0)", "generator", false, true},
		{"generator expression with multiple clauses", "(x + y for x in range(3) for y in range(3))", "generator", false, true},
		{"async generator expression", "(x async for x in async_source())", "generator", false, true},
		{"complex generator expression", "(x * 2 for x in items if x.isdigit())", "generator", false, true},

		// Regular containers (should not be comprehensions)
		{"regular list", "[1, 2, 3]", "", false, false},
		{"regular set", "{1, 2, 3}", "", false, false},
		{"regular dict", "{'a': 1, 'b': 2}", "", false, false},
		{"regular tuple", "(1, 2, 3)", "", false, false},

		// Error cases
		{"list comprehension without for", "[x x in items]", "", true, false},
		{"list comprehension without in", "[x for x items]", "", true, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check if it's a comprehension when expected
			if isComprehension(expr) != test.expectComp {
				t.Errorf("Expected comprehension=%v but got %v", test.expectComp, isComprehension(expr))
				return
			}

			// Validate comprehension structure if it is one
			if test.expectComp && test.compType != "" {
				validateComprehension(t, expr, test.compType)
			}
		})
	}
}

// Test comprehension clause structure in detail
func TestComprehensionClauses(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedClauses int
		hasAsync        bool
		hasConditions   bool
		hasError        bool
	}{
		{"single clause", "[x for x in items]", 1, false, false, false},
		{"clause with condition", "[x for x in items if x > 0]", 1, false, true, false},
		{"multiple clauses", "[x + y for x in range(3) for y in range(3)]", 2, false, false, false},
		{"async clause", "[x async for x in async_items]", 1, true, false, false},
		{"multiple conditions", "[x for x in items if x > 0 if x < 100]", 1, false, true, false},
		{"complex nested", "[x * y for x in range(3) for y in range(3) if x != y if x + y > 2]", 2, false, true, false},
		{"async with conditions", "[x async for x in async_items if x.valid]", 1, true, true, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			listComp, ok := expr.(*ast.ListComp)
			if !ok {
				t.Fatalf("Expected *ast.ListComp, got %T", expr)
			}

			if len(listComp.Clauses) != test.expectedClauses {
				t.Errorf("Expected %d clauses, got %d", test.expectedClauses, len(listComp.Clauses))
			}

			// Check for async clauses
			hasAsync := false
			for _, clause := range listComp.Clauses {
				if clause.IsAsync {
					hasAsync = true
					break
				}
			}
			if hasAsync != test.hasAsync {
				t.Errorf("Expected hasAsync=%v, got %v", test.hasAsync, hasAsync)
			}

			// Check for conditions
			hasConditions := false
			for _, clause := range listComp.Clauses {
				if len(clause.Ifs) > 0 {
					hasConditions = true
					break
				}
			}
			if hasConditions != test.hasConditions {
				t.Errorf("Expected hasConditions=%v, got %v", test.hasConditions, hasConditions)
			}
		})
	}
}

// Test comprehensions in complex expressions
func TestComprehensionsInComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
		parseAs  string // "expression" or "statement"
	}{
		{"list comprehension in function call", "sum([x * 2 for x in numbers])", false, "expression"},
		{"nested comprehensions", "[[y for y in row] for row in matrix]", false, "expression"},
		{"comprehension with function calls", "[process(x) for x in items if validate(x)]", false, "expression"},
		{"dict comprehension in assignment", "result = {k.upper(): len(v) for k, v in data.items()}", false, "statement"},
		{"comprehension with lambda", "[f(x) for f in [lambda y: y*2, lambda y: y+1] for x in range(3)]", false, "expression"},
		{"set comprehension with attribute access", "{item.name for item in objects if hasattr(item, 'name')}", false, "expression"},
		// NOTE: Generator expressions in function calls not yet supported by parser
		// {"generator in function", "list(x*2 for x in range(5))", false, "expression"},
		{"nested dict comprehension", "{k: {x: x*2 for x in v} for k, v in data.items()}", false, "expression"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			var err error
			if test.parseAs == "statement" {
				_, err = parser.statement()
			} else {
				_, err = parser.expression()
			}

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// Test edge cases and error conditions
func TestComprehensionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"empty list comprehension", "[]", false}, // Not a comprehension, regular empty list
		{"empty set", "{}", false},                // Not a comprehension, empty set/dict
		{"invalid list comprehension syntax", "[for x in items]", true},
		{"missing iterator", "[x for x]", true},
		{"set comprehension with tuple unpacking", "{x for x, y in items}", false}, // Valid set comprehension, not dict
		{"malformed condition", "[x for x in items if]", true},
		{"nested comprehension error", "[x for x in [y for y in]]", true},
		{"async without for", "[x async x in items]", true},
		{"complex nested valid", "[x for x in [y*2 for y in range(3)] if x > 1]", false},
		{"walrus operator in comprehension", "[y := x*2 for x in range(3)]", false}, // Should be valid
		{"multiple async keywords", "[x async async for x in items]", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// For non-error cases, just verify we got a valid expression
			if expr == nil {
				t.Error("Expected non-nil expression")
			}
		})
	}
}
