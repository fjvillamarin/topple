package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a list expression
func parseListExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.list()
}

// Helper function to validate list expression structure
func validateListExpression(t *testing.T, expr ast.Expr, expectedCount int) *ast.ListExpr {
	list, ok := expr.(*ast.ListExpr)
	if !ok {
		t.Fatalf("Expected *ast.ListExpr, got %T", expr)
	}

	if expectedCount >= 0 && len(list.Elements) != expectedCount {
		t.Errorf("Expected %d elements, got %d", expectedCount, len(list.Elements))
	}

	return list
}

// Helper function to validate list comprehension structure
func validateListComprehension(t *testing.T, expr ast.Expr, expectedGenerators int, expectedConditions bool) *ast.ListComp {
	comp, ok := expr.(*ast.ListComp)
	if !ok {
		t.Fatalf("Expected *ast.ListComp, got %T", expr)
	}

	if comp.Element == nil {
		t.Fatal("List comprehension missing element expression")
	}

	if len(comp.Clauses) == 0 {
		t.Fatal("List comprehension missing for clauses")
	}

	// Count generators (for clauses)
	genCount := len(comp.Clauses)
	if expectedGenerators >= 0 && genCount != expectedGenerators {
		t.Errorf("Expected %d generators, got %d", expectedGenerators, genCount)
	}

	return comp
}

// Helper function to count starred expressions in list
func countStarredExpressions(list *ast.ListExpr) int {
	count := 0
	for _, elem := range list.Elements {
		if _, ok := elem.(*ast.StarExpr); ok {
			count++
		}
	}
	return count
}

// Helper function to get element types in list
func getElementTypes(list *ast.ListExpr) []string {
	types := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		switch elem.(type) {
		case *ast.Literal:
			types[i] = "literal"
		case *ast.Name:
			types[i] = "name"
		case *ast.Binary:
			types[i] = "binary"
		case *ast.Call:
			types[i] = "call"
		case *ast.ListExpr:
			types[i] = "list"
		case *ast.Lambda:
			types[i] = "lambda"
		case *ast.StarExpr:
			types[i] = "starred"
		default:
			types[i] = "other"
		}
	}
	return types
}

// Test comprehensive list expression functionality
func TestListExpressions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCount int
		elementTypes  []string
		description   string
	}{
		// Basic list expressions
		{
			name:          "empty list",
			input:         "[]",
			expectedCount: 0,
			elementTypes:  []string{},
			description:   "empty list literal",
		},
		{
			name:          "single element",
			input:         "[1]",
			expectedCount: 1,
			elementTypes:  []string{"literal"},
			description:   "list with single literal element",
		},
		{
			name:          "multiple elements",
			input:         "[1, 2, 3]",
			expectedCount: 3,
			elementTypes:  []string{"literal", "literal", "literal"},
			description:   "list with multiple literal elements",
		},
		{
			name:          "trailing comma",
			input:         "[1, 2, 3,]",
			expectedCount: 3,
			elementTypes:  []string{"literal", "literal", "literal"},
			description:   "list with trailing comma",
		},

		// Different element types
		{
			name:          "mixed types",
			input:         "[1, 'hello', True, None]",
			expectedCount: 4,
			elementTypes:  []string{"literal", "literal", "literal", "literal"},
			description:   "list with mixed literal types",
		},
		{
			name:          "variables",
			input:         "[x, y, z]",
			expectedCount: 3,
			elementTypes:  []string{"name", "name", "name"},
			description:   "list with variable names",
		},
		{
			name:          "expressions",
			input:         "[x + 1, y * 2, z - 3]",
			expectedCount: 3,
			elementTypes:  []string{"binary", "binary", "binary"},
			description:   "list with arithmetic expressions",
		},
		{
			name:          "function calls",
			input:         "[func(), method(), len(x)]",
			expectedCount: 3,
			elementTypes:  []string{"call", "call", "call"},
			description:   "list with function calls",
		},

		// Nested structures
		{
			name:          "nested lists",
			input:         "[[1, 2], [3, 4]]",
			expectedCount: 2,
			elementTypes:  []string{"list", "list"},
			description:   "list containing nested lists",
		},
		{
			name:          "deeply nested",
			input:         "[[[1]], [[2, 3]], [4]]",
			expectedCount: 3,
			elementTypes:  []string{"list", "list", "list"},
			description:   "deeply nested list structure",
		},
		{
			name:          "mixed nesting",
			input:         "[1, [2, 3], 4]",
			expectedCount: 3,
			elementTypes:  []string{"literal", "list", "literal"},
			description:   "list with mixed literal and nested elements",
		},

		// Complex expressions
		{
			name:          "attribute access",
			input:         "[obj.attr, data.value, item.name]",
			expectedCount: 3,
			description:   "list with attribute access expressions",
		},
		{
			name:          "subscript access",
			input:         "[data[0], items[key], matrix[i][j]]",
			expectedCount: 3,
			description:   "list with subscript access expressions",
		},
		{
			name:          "conditional expressions",
			input:         "[x if condition else y, a if flag else b]",
			expectedCount: 2,
			description:   "list with ternary expressions",
		},

		// Error cases
		{
			name:        "unclosed list",
			input:       "[1, 2, 3",
			hasError:    true,
			description: "list missing closing bracket",
		},
		{
			name:        "missing comma",
			input:       "[1 2 3]",
			hasError:    true,
			description: "list elements without comma separator",
		},
		{
			name:        "empty element",
			input:       "[1, , 3]",
			hasError:    true,
			description: "list with empty element between commas",
		},
		{
			name:        "invalid element",
			input:       "[1, 2, def]",
			hasError:    true,
			description: "list with invalid element syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseListExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			list := validateListExpression(t, expr, test.expectedCount)

			// Validate element types if specified
			if len(test.elementTypes) > 0 {
				actualTypes := getElementTypes(list)
				if len(actualTypes) != len(test.elementTypes) {
					t.Errorf("Expected %d element types, got %d", len(test.elementTypes), len(actualTypes))
				} else {
					for i, expectedType := range test.elementTypes {
						if i < len(actualTypes) && actualTypes[i] != expectedType {
							t.Errorf("Element %d: expected type %s, got %s", i, expectedType, actualTypes[i])
						}
					}
				}
			}
		})
	}
}

// Test list comprehensions and their variations
func TestListComprehensions(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		hasError           bool
		expectedGenerators int
		hasCondition       bool
		description        string
	}{
		// Basic comprehensions
		{
			name:               "simple comprehension",
			input:              "[x for x in range(10)]",
			expectedGenerators: 1,
			hasCondition:       false,
			description:        "basic list comprehension with single generator",
		},
		{
			name:               "with condition",
			input:              "[x for x in range(10) if x % 2 == 0]",
			expectedGenerators: 1,
			hasCondition:       true,
			description:        "list comprehension with filter condition",
		},
		{
			name:               "expression transform",
			input:              "[x * 2 for x in numbers]",
			expectedGenerators: 1,
			hasCondition:       false,
			description:        "list comprehension with expression transformation",
		},
		{
			name:               "nested generators",
			input:              "[x + y for x in range(3) for y in range(3)]",
			expectedGenerators: 2,
			hasCondition:       false,
			description:        "list comprehension with multiple generators",
		},
		{
			name:               "multiple conditions",
			input:              "[x for x in items if x > 0 if x < 10]",
			expectedGenerators: 1,
			hasCondition:       true,
			description:        "list comprehension with multiple filter conditions",
		},

		// Complex comprehensions
		{
			name:               "function call in element",
			input:              "[func(x) for x in data]",
			expectedGenerators: 1,
			hasCondition:       false,
			description:        "list comprehension with function call in element expression",
		},
		{
			name:               "attribute access in condition",
			input:              "[item for item in objects if item.is_valid]",
			expectedGenerators: 1,
			hasCondition:       true,
			description:        "list comprehension with attribute access in condition",
		},
		{
			name:               "complex expression and condition",
			input:              "[item.name.upper() for item in data if item.active and item.visible]",
			expectedGenerators: 1,
			hasCondition:       true,
			description:        "list comprehension with complex element expression and condition",
		},

		// Nested data access
		{
			name:               "tuple unpacking",
			input:              "[x + y for x, y in pairs]",
			expectedGenerators: 1,
			hasCondition:       false,
			description:        "list comprehension with tuple unpacking in generator",
		},
		{
			name:               "nested iteration",
			input:              "[cell for row in matrix for cell in row]",
			expectedGenerators: 2,
			hasCondition:       false,
			description:        "list comprehension flattening nested structure",
		},

		// Error cases
		{
			name:        "invalid comprehension syntax",
			input:       "[x for]",
			hasError:    true,
			description: "comprehension missing variable and iterable",
		},
		{
			name:        "missing in keyword",
			input:       "[x for x range(10)]",
			hasError:    true,
			description: "comprehension missing 'in' keyword",
		},
		{
			name:        "incomplete condition",
			input:       "[x for x in items if]",
			hasError:    true,
			description: "comprehension with incomplete condition",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseListExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateListComprehension(t, expr, test.expectedGenerators, test.hasCondition)
		})
	}
}

// Test starred expressions in lists
func TestListStarredExpressions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedStars int
		description   string
	}{
		{
			name:          "starred at beginning",
			input:         "[*first, 2, 3]",
			expectedStars: 1,
			description:   "starred expression at beginning of list",
		},
		{
			name:          "starred at end",
			input:         "[1, 2, *rest]",
			expectedStars: 1,
			description:   "starred expression at end of list",
		},
		{
			name:          "starred in middle",
			input:         "[1, *middle, 3]",
			expectedStars: 1,
			description:   "starred expression in middle of list",
		},
		{
			name:          "multiple starred",
			input:         "[*a, *b, *c]",
			expectedStars: 3,
			description:   "multiple starred expressions in list",
		},
		{
			name:          "starred with expressions",
			input:         "[x, *items, y + 1]",
			expectedStars: 1,
			description:   "starred expression mixed with regular expressions",
		},
		{
			name:          "starred function call",
			input:         "[*get_items(), 42]",
			expectedStars: 1,
			description:   "starred function call in list",
		},

		// Error cases
		{
			name:        "invalid starred",
			input:       "[*]",
			hasError:    true,
			description: "starred expression without operand",
		},
		{
			name:        "starred with invalid syntax",
			input:       "[*1 2]",
			hasError:    true,
			description: "starred expression with invalid syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseListExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			list := validateListExpression(t, expr, -1)

			// Count starred expressions
			starCount := countStarredExpressions(list)
			if test.expectedStars >= 0 && starCount != test.expectedStars {
				t.Errorf("Expected %d starred expressions for %s, got %d", test.expectedStars, test.description, starCount)
			}

			// Verify at least one starred expression exists if expected
			if test.expectedStars > 0 && starCount == 0 {
				t.Errorf("Expected at least one starred expression for %s", test.description)
			}
		})
	}
}

// Test list edge cases and complex scenarios
func TestListEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		// Complex valid cases
		{
			name:        "deeply nested structure",
			input:       "[[[1, 2]], [[3, [4, 5]], 6], [7]]",
			description: "complex deeply nested list structure",
		},
		{
			name:        "mixed comprehensions and literals",
			input:       "[[x for x in range(3)], [1, 2, 3], [y * 2 for y in data]]",
			description: "list containing both comprehensions and literals",
		},
		{
			name:        "complex expressions in list",
			input:       "[obj.method().attr, func(a, b, c), x if cond else y]",
			description: "list with complex member expressions",
		},
		{
			name:        "list with lambda expressions",
			input:       "[lambda x: x + 1, lambda y: y * 2]",
			description: "list containing lambda expressions",
		},

		// Error cases with specific error validation
		{
			name:        "invalid element syntax",
			input:       "[1, 2, def]",
			hasError:    true,
			description: "list with invalid element syntax",
		},
		{
			name:        "invalid comprehension",
			input:       "[x for x in for y in items]",
			hasError:    true,
			description: "malformed comprehension syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseListExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			// For successful cases, just verify we got a valid list or comprehension
			if expr == nil {
				t.Errorf("Expected valid expression but got nil for %s", test.description)
			}
		})
	}
}
