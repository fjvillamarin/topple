package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a set expression
func parseSetExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.set()
}

// Helper function to parse dict or set expression
func parseDictOrSetExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.dictOrSet()
}

// Helper function to validate set expression structure
func validateSetExpression(t *testing.T, expr ast.Expr, expectedElements int) *ast.SetExpr {
	set, ok := expr.(*ast.SetExpr)
	if !ok {
		t.Fatalf("Expected *ast.SetExpr, got %T", expr)
	}

	if expectedElements >= 0 && len(set.Elements) != expectedElements {
		t.Errorf("Expected %d elements, got %d", expectedElements, len(set.Elements))
	}

	return set
}

// Helper function to validate set comprehension structure
func validateSetComprehension(t *testing.T, expr ast.Expr, expectedGenerators int, hasCondition bool) *ast.SetComp {
	comp, ok := expr.(*ast.SetComp)
	if !ok {
		t.Fatalf("Expected *ast.SetComp, got %T", expr)
	}

	if comp.Element == nil {
		t.Fatal("Set comprehension missing element expression")
	}

	if len(comp.Clauses) == 0 {
		t.Fatal("Set comprehension missing for clauses")
	}

	genCount := len(comp.Clauses)
	if expectedGenerators >= 0 && genCount != expectedGenerators {
		t.Errorf("Expected %d generators, got %d", expectedGenerators, genCount)
	}

	return comp
}

// Helper function to count starred expressions in sets
func countStarredExpressionsInSet(set *ast.SetExpr) int {
	count := 0
	for _, elem := range set.Elements {
		if _, ok := elem.(*ast.StarExpr); ok {
			count++
		}
	}
	return count
}

// Helper function to get set expression type string
func getSetExpressionType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.DictExpr:
		return "DictExpr"
	case *ast.SetExpr:
		return "SetExpr"
	case *ast.DictComp:
		return "DictComp"
	case *ast.SetComp:
		return "SetComp"
	default:
		return "unknown"
	}
}

// Test comprehensive set expression functionality
func TestSetExpressions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		elemCount    int
		hasError     bool
		description  string
	}{
		// Basic set expressions
		{
			name:        "single element",
			input:       "{1}",
			elemCount:   1,
			description: "set with single numeric element",
		},
		{
			name:        "multiple elements",
			input:       "{1, 2, 3}",
			elemCount:   3,
			description: "set with multiple numeric elements",
		},
		{
			name:        "with duplicates",
			input:       "{1, 2, 2, 3}",
			elemCount:   4,
			description: "set with duplicate elements (parser doesn't dedupe)",
		},
		{
			name:        "mixed types",
			input:       "{1, 'hello', True}",
			elemCount:   3,
			description: "set with mixed literal types",
		},
		{
			name:        "trailing comma",
			input:       "{1, 2, 3,}",
			elemCount:   3,
			description: "set with trailing comma",
		},

		// Expression elements
		{
			name:        "expressions",
			input:       "{x + 1, y * 2}",
			elemCount:   2,
			description: "set with expression elements",
		},
		{
			name:        "function calls",
			input:       "{func(), method(), len(x)}",
			elemCount:   3,
			description: "set with function call elements",
		},
		{
			name:        "nested structures",
			input:       "{(1, 2), (3, 4)}",
			elemCount:   2,
			description: "set containing tuples",
		},
		{
			name:        "complex expressions",
			input:       "{obj.attr, data[key], item.method()}",
			elemCount:   3,
			description: "set with complex member access expressions",
		},

		// Error cases
		{
			name:        "unclosed set",
			input:       "{1, 2, 3",
			hasError:    true,
			description: "set missing closing brace",
		},
		{
			name:        "missing comma",
			input:       "{1 2 3}",
			hasError:    true,
			description: "set elements without comma separator",
		},
		{
			name:        "empty element",
			input:       "{1, , 3}",
			hasError:    true,
			description: "set with empty element between commas",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseSetExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateSetExpression(t, expr, test.elemCount)
		})
	}
}

// Test set comprehensions and their variations
func TestSetComprehensions(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedGenerators int
		hasCondition      bool
		hasError          bool
		description       string
	}{
		// Basic comprehensions
		{
			name:              "simple comprehension",
			input:             "{x for x in range(10)}",
			expectedGenerators: 1,
			hasCondition:      false,
			description:       "basic set comprehension with single generator",
		},
		{
			name:              "with condition",
			input:             "{x for x in items if x > 0}",
			expectedGenerators: 1,
			hasCondition:      true,
			description:       "set comprehension with filter condition",
		},
		{
			name:              "expression transform",
			input:             "{x * 2 for x in numbers}",
			expectedGenerators: 1,
			hasCondition:      false,
			description:       "set comprehension with expression transformation",
		},
		{
			name:              "nested generators",
			input:             "{x + y for x in range(3) for y in range(3)}",
			expectedGenerators: 2,
			hasCondition:      false,
			description:       "set comprehension with multiple generators",
		},
		{
			name:              "complex comprehension",
			input:             "{process(x) for x in data if valid(x)}",
			expectedGenerators: 1,
			hasCondition:      true,
			description:       "set comprehension with function call and condition",
		},
		{
			name:              "unique values",
			input:             "{x % 3 for x in range(10)}",
			expectedGenerators: 1,
			hasCondition:      false,
			description:       "set comprehension creating unique values",
		},

		// Advanced comprehensions
		{
			name:              "tuple unpacking",
			input:             "{x + y for x, y in pairs}",
			expectedGenerators: 1,
			hasCondition:      false,
			description:       "set comprehension with tuple unpacking",
		},
		{
			name:              "nested iteration",
			input:             "{item for row in matrix for item in row}",
			expectedGenerators: 2,
			hasCondition:      false,
			description:       "set comprehension flattening nested structure",
		},
		{
			name:              "complex condition",
			input:             "{x for x in items if x > 0 and x < 10}",
			expectedGenerators: 1,
			hasCondition:      true,
			description:       "set comprehension with complex boolean condition",
		},

		// Error cases
		{
			name:        "invalid comprehension",
			input:       "{x for}",
			hasError:    true,
			description: "comprehension missing variable and iterable",
		},
		{
			name:        "missing in keyword",
			input:       "{x for x range(10)}",
			hasError:    true,
			description: "comprehension missing 'in' keyword",
		},
		{
			name:        "incomplete condition",
			input:       "{x for x in items if}",
			hasError:    true,
			description: "comprehension with incomplete condition",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseDictOrSetExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateSetComprehension(t, expr, test.expectedGenerators, test.hasCondition)
		})
	}
}

// Test starred expressions in sets
func TestSetStarredExpressions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStars int
		hasError      bool
		description   string
	}{
		{
			name:          "starred element",
			input:         "{1, *other}",
			expectedStars: 1,
			description:   "set with starred element",
		},
		{
			name:          "multiple starred",
			input:         "{*a, *b, 3}",
			expectedStars: 2,
			description:   "set with multiple starred elements",
		},
		{
			name:          "starred expression",
			input:         "{*(x for x in range(3))}",
			expectedStars: 1,
			description:   "set with starred generator expression",
		},
		{
			name:          "starred call",
			input:         "{*get_items()}",
			expectedStars: 1,
			description:   "set with starred function call",
		},
		{
			name:          "mixed starred and regular",
			input:         "{1, *items, 'hello', *more}",
			expectedStars: 2,
			description:   "set with mixed regular and starred elements",
		},

		// Error cases
		{
			name:        "invalid starred",
			input:       "{*}",
			hasError:    true,
			description: "starred expression without operand",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseSetExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			set := validateSetExpression(t, expr, -1)

			starCount := countStarredExpressionsInSet(set)
			if test.expectedStars >= 0 && starCount != test.expectedStars {
				t.Errorf("Expected %d starred expressions, got %d for %s",
					test.expectedStars, starCount, test.description)
			}

			if test.expectedStars > 0 && starCount == 0 {
				t.Errorf("Expected at least one starred expression for %s", test.description)
			}
		})
	}
}

// Test disambiguation between dict and set
func TestDictSetDisambiguation(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		description  string
	}{
		{
			name:         "empty dict",
			input:        "{}",
			expectedType: "DictExpr",
			description:  "empty braces should parse as dict, not set",
		},
		{
			name:         "single element set",
			input:        "{1}",
			expectedType: "SetExpr",
			description:  "single element in braces should be set",
		},
		{
			name:         "dict with colon",
			input:        "{1: 2}",
			expectedType: "DictExpr",
			description:  "key-value pair should be dict",
		},
		{
			name:         "set comprehension",
			input:        "{x for x in range(3)}",
			expectedType: "SetComp",
			description:  "comprehension without colon should be set comprehension",
		},
		{
			name:         "dict comprehension",
			input:        "{x: x**2 for x in range(3)}",
			expectedType: "DictComp",
			description:  "comprehension with colon should be dict comprehension",
		},
		{
			name:         "dict unpacking",
			input:        "{**other}",
			expectedType: "DictExpr",
			description:  "double-star unpacking should be dict",
		},
		{
			name:         "set with starred",
			input:        "{*other}",
			expectedType: "SetExpr",
			description:  "single-star unpacking should be set",
		},
		{
			name:         "multiple dict pairs",
			input:        "{1: 'a', 2: 'b'}",
			expectedType: "DictExpr",
			description:  "multiple key-value pairs should be dict",
		},
		{
			name:         "multiple set elements",
			input:        "{1, 2, 3}",
			expectedType: "SetExpr",
			description:  "multiple elements without colons should be set",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseDictOrSetExpression(t, test.input)
			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			actualType := getSetExpressionType(expr)
			if actualType != test.expectedType {
				t.Errorf("Expected %s, got %s for %s",
					test.expectedType, actualType, test.description)
			}
		})
	}
}

// Test set edge cases and complex scenarios
func TestSetEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		errorContains string
		description  string
	}{
		// Complex valid cases
		{
			name:        "deeply nested structures",
			input:       "{((1, 2), (3, 4)), [(5, 6), (7, 8)]}",
			description: "set with deeply nested tuples and lists",
		},
		{
			name:        "mixed comprehensions and literals",
			input:       "{x for x in range(3)} | {4, 5, 6}",
			description: "set operations with comprehensions",
		},
		{
			name:        "complex expressions in set",
			input:       "{obj.method().attr, func(a, b, c), x if cond else y}",
			description: "set with complex member expressions",
		},
		{
			name:        "set with lambda expressions",
			input:       "{lambda x: x + 1, lambda y: y * 2}",
			description: "set containing lambda expressions",
		},

		// Error cases
		{
			name:          "invalid element syntax",
			input:         "{1, 2, def}",
			hasError:      true,
			errorContains: "unexpected",
			description:   "set with invalid element syntax",
		},
		{
			name:          "invalid comprehension syntax",
			input:         "{x for x in for y in items}",
			hasError:      true,
			description:   "malformed comprehension syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseDictOrSetExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorContains, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			// For successful cases, just verify we got a valid expression
			if expr == nil {
				t.Errorf("Expected valid expression but got nil for %s", test.description)
			}
		})
	}
}