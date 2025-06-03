package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a grouped expression
func parseGroupExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	
	// Check for lexer errors
	if len(scanner.Errors) > 0 {
		t.Fatalf("Lexer errors encountered: %v", scanner.Errors)
	}
	
	parser := NewParser(tokens)
	expr, err := parser.expression()
	
	// Check for parser errors
	if len(parser.Errors) > 0 {
		t.Fatalf("Parser errors encountered: %v", parser.Errors)
	}
	
	return expr, err
}

// Helper function to get expression type as string for group tests
func getGroupExpressionType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.Literal:
		return "literal"
	case *ast.Name:
		return "name"
	case *ast.Binary:
		return "binary"
	case *ast.Unary:
		return "unary"
	case *ast.Call:
		return "call"
	case *ast.Attribute:
		return "attribute"
	case *ast.Subscript:
		return "subscript"
	case *ast.ListExpr:
		return "list"
	case *ast.ListComp:
		return "list_comp"
	case *ast.DictExpr:
		return "dict"
	case *ast.DictComp:
		return "dict_comp"
	case *ast.SetExpr:
		return "set"
	case *ast.SetComp:
		return "set_comp"
	case *ast.TupleExpr:
		return "tuple"
	case *ast.TernaryExpr:
		return "ternary"
	case *ast.Lambda:
		return "lambda"
	case *ast.YieldExpr:
		return "yield"
	case *ast.GroupExpr:
		return "group"
	case *ast.StarExpr:
		return "star"
	case *ast.FString:
		return "fstring"
	default:
		return "unknown"
	}
}

// Helper function to validate group expression structure
func validateGroupExpr(t *testing.T, expr ast.Expr, expectedInnerType string) *ast.GroupExpr {
	group, ok := expr.(*ast.GroupExpr)
	if !ok {
		t.Fatalf("Expected *ast.GroupExpr, got %T", expr)
	}

	if group.Expression == nil {
		t.Fatal("Group expression should have an inner expression")
	}

	if expectedInnerType != "" {
		actualType := getGroupExpressionType(group.Expression)
		if actualType != expectedInnerType {
			t.Errorf("Expected inner expression type %s, got %s (actual: %T)", 
				expectedInnerType, actualType, group.Expression)
		}
	}

	return group
}

// Helper function to count nested group expression levels
func countNestedGroups(expr ast.Expr) int {
	count := 0
	current := expr
	for {
		if group, ok := current.(*ast.GroupExpr); ok {
			count++
			current = group.Expression
		} else {
			break
		}
	}
	return count
}

// Helper function to validate any expression type
func validateExpressionType(t *testing.T, expr ast.Expr, expectedType string) {
	actualType := getGroupExpressionType(expr)
	if actualType != expectedType {
		t.Errorf("Expected expression type %s, got %s (actual: %T)", 
			expectedType, actualType, expr)
	}
}

// Test comprehensive grouped expression functionality
func TestGroupedExpressions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedInner string
		isGroup       bool
		description   string
	}{
		// Basic grouped expressions
		{
			name:          "simple literal",
			input:         "(42)",
			expectedInner: "literal",
			isGroup:       true,
			description:   "grouped integer literal",
		},
		{
			name:          "grouped variable",
			input:         "(x)",
			expectedInner: "name",
			isGroup:       true,
			description:   "grouped variable name",
		},
		{
			name:          "grouped string",
			input:         `("hello")`,
			expectedInner: "literal",
			isGroup:       true,
			description:   "grouped string literal",
		},
		{
			name:          "grouped boolean",
			input:         "(True)",
			expectedInner: "literal",
			isGroup:       true,
			description:   "grouped boolean literal",
		},
		{
			name:          "grouped None",
			input:         "(None)",
			expectedInner: "literal",
			isGroup:       true,
			description:   "grouped None literal",
		},

		// Grouped operations
		{
			name:          "grouped arithmetic",
			input:         "(a + b)",
			expectedInner: "binary",
			isGroup:       true,
			description:   "grouped arithmetic operation",
		},
		{
			name:          "grouped comparison",
			input:         "(x < 10)",
			expectedInner: "binary",
			isGroup:       true,
			description:   "grouped comparison operation",
		},
		{
			name:          "grouped boolean operation",
			input:         "(x and y)",
			expectedInner: "binary",
			isGroup:       true,
			description:   "grouped boolean operation",
		},
		{
			name:          "grouped unary operation",
			input:         "(-x)",
			expectedInner: "unary",
			isGroup:       true,
			description:   "grouped unary operation",
		},

		// Grouped complex expressions
		{
			name:          "grouped function call",
			input:         "(func())",
			expectedInner: "call",
			isGroup:       true,
			description:   "grouped function call",
		},
		{
			name:          "grouped attribute access",
			input:         "(obj.attr)",
			expectedInner: "attribute",
			isGroup:       true,
			description:   "grouped attribute access",
		},
		{
			name:          "grouped subscript",
			input:         "(arr[0])",
			expectedInner: "subscript",
			isGroup:       true,
			description:   "grouped subscript access",
		},
		{
			name:          "grouped list",
			input:         "([1, 2, 3])",
			expectedInner: "list",
			isGroup:       true,
			description:   "grouped list literal",
		},
		{
			name:          "grouped dict",
			input:         `({"key": "value"})`,
			expectedInner: "dict",
			isGroup:       true,
			description:   "grouped dictionary literal",
		},
		{
			name:          "grouped set",
			input:         "({1, 2, 3})",
			expectedInner: "set",
			isGroup:       true,
			description:   "grouped set literal",
		},
		{
			name:          "grouped tuple",
			input:         "((1, 2, 3))",
			expectedInner: "tuple",
			isGroup:       true,
			description:   "grouped tuple literal",
		},

		// Advanced grouped expressions
		{
			name:          "grouped conditional",
			input:         "(x if condition else y)",
			expectedInner: "ternary",
			isGroup:       true,
			description:   "grouped ternary expression",
		},
		{
			name:          "grouped list comprehension",
			input:         "([x for x in items])",
			expectedInner: "list_comp",  // Parser correctly identifies this as list comprehension
			isGroup:       true,
			description:   "grouped list comprehension",
		},
		{
			name:          "grouped lambda",
			input:         "(lambda x: x * 2)",
			expectedInner: "lambda",
			isGroup:       true,
			description:   "grouped lambda expression",
		},
		{
			name:          "grouped yield",
			input:         "(yield value)",
			expectedInner: "yield",
			isGroup:       true,
			description:   "grouped yield expression",
		},
		{
			name:          "grouped yield from",
			input:         "(yield from items)",
			expectedInner: "yield",
			isGroup:       true,
			description:   "grouped yield from expression",
		},

		// Special cases
		{
			name:        "empty parentheses",
			input:       "()",
			isGroup:     false,
			description: "empty parentheses create tuple",
		},
		{
			name:          "single element tuple",
			input:         "(x,)",
			expectedInner: "",  // Don't check inner type as this might be parsed differently
			isGroup:       false,
			description:   "single element tuple with trailing comma",
		},

		// Error cases
		{
			name:        "unclosed parenthesis",
			input:       "(42",
			hasError:    true,
			description: "missing closing parenthesis",
		},
		{
			name:        "empty expression in parens",
			input:       "(,)",
			hasError:    true,
			description: "invalid comma without expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseGroupExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if test.isGroup {
				validateGroupExpr(t, expr, test.expectedInner)
			} else {
				// Verify it's not a group expression (e.g., empty parens create tuple)
				if _, ok := expr.(*ast.GroupExpr); ok && test.expectedInner == "" {
					t.Errorf("Expected non-group expression for %s, but got GroupExpr", test.description)
				}
			}
		})
	}
}

// Test nested group expressions
func TestNestedGroupExpressions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		expectedNest int
		description  string
	}{
		{
			name:         "double nested",
			input:        "((42))",
			expectedNest: 2,
			description:  "two levels of grouping",
		},
		{
			name:         "triple nested",
			input:        "(((x)))",
			expectedNest: 3,
			description:  "three levels of grouping",
		},
		{
			name:         "quadruple nested",
			input:        "((((value))))",
			expectedNest: 4,
			description:  "four levels of grouping",
		},
		{
			name:         "nested with operation",
			input:        "((a + b))",
			expectedNest: 2,
			description:  "nested grouping around operation",
		},
		{
			name:         "asymmetric nesting",
			input:        "(((a + b)) * ((c - d)))",
			expectedNest: 1,
			description:  "top-level group with nested groups inside",
		},
		{
			name:         "complex nested",
			input:        "((a + (b * c)) / ((d - e) + f))",
			expectedNest: 1,
			description:  "complex expression with multiple nested groups",
		},

		// Error cases
		{
			name:        "mismatched nested parens",
			input:       "(((42))",
			hasError:    true,
			description: "unbalanced nested parentheses",
		},
		{
			name:        "unclosed nested",
			input:       "((42",
			hasError:    true,
			description: "unclosed nested parentheses",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseGroupExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			nestCount := countNestedGroups(expr)
			if nestCount != test.expectedNest {
				t.Errorf("Expected %d nesting levels for %s, got %d", test.expectedNest, test.description, nestCount)
			}
		})
	}
}

// Test grouping effects on operator precedence
func TestGroupingPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		expectGroup bool
		description string
	}{
		{
			name:        "grouping changes multiplication precedence",
			input:       "(a + b) * c",
			expectGroup: false, // The outer expression is multiplication, not a group
			description: "grouped addition multiplied by variable",
		},
		{
			name:        "grouping changes division precedence",
			input:       "x / (y + z)",
			expectGroup: false, // The outer expression is division, not a group
			description: "division by grouped addition",
		},
		{
			name:        "comparison with grouped expressions",
			input:       "(x + y) > (a - b)",
			expectGroup: false, // The outer expression is comparison, not a group
			description: "comparison between two grouped expressions",
		},
		{
			name:        "boolean logic with grouping",
			input:       "(x and y) or (a and b)",
			expectGroup: false, // The outer expression is boolean or, not a group
			description: "or between grouped and expressions",
		},
		{
			name:        "nested precedence override",
			input:       "((a + b) * c) / d",
			expectGroup: false, // The outer expression is division, not a group
			description: "division with grouped multiplication",
		},
		{
			name:        "power with grouping",
			input:       "(a + b) ** 2",
			expectGroup: false, // The outer expression is power, not a group
			description: "grouped addition raised to power",
		},
		{
			name:        "unary with grouping",
			input:       "-(a + b)",
			expectGroup: false, // The outer expression is unary, not a group
			description: "negation of grouped addition",
		},

		// These should maintain group structure
		{
			name:        "grouped complex expression",
			input:       "((x + y) * (a - b))",
			expectGroup: true,
			description: "grouped multiplication of grouped expressions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseGroupExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			isGroup := getGroupExpressionType(expr) == "group"
			if isGroup != test.expectGroup {
				t.Errorf("Expected group=%v for %s, got group=%v (actual type: %s)", 
					test.expectGroup, test.description, isGroup, getGroupExpressionType(expr))
			}
		})
	}
}

// Test error cases and edge scenarios
func TestGroupExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		errorContains string
		description  string
	}{
		{
			name:          "deeply nested expression",
			input:         "(((((((a + b)))))))",
			hasError:      false,
			description:   "very deeply nested grouping",
		},
		{
			name:          "grouped complex chained operations",
			input:         "(a.b().c[0].d)",
			hasError:      false,
			description:   "grouped method chaining and attribute access",
		},
		{
			name:          "grouped f-string",
			input:         `(f"hello {name}")`,
			hasError:      false,
			description:   "grouped f-string expression",
		},
		{
			name:          "grouped starred expression",
			input:         "(*args)",
			hasError:      false,
			description:   "grouped starred expression",
		},
		{
			name:          "grouped complex comprehension",
			input:         "([x for x in range(10) if x % 2 == 0])",
			hasError:      false,
			description:   "grouped list comprehension with condition",
		},

		// Error cases
		{
			name:          "unclosed parenthesis",
			input:         "(42",
			hasError:      true,
			errorContains: "expected",
			description:   "missing closing parenthesis",
		},
		// Note: "42)" doesn't produce an error because parser.group() only parses the "42" part
		// The extra ")" would be caught by the caller context, not the group parser itself
		{
			name:          "nested unclosed",
			input:         "((42",
			hasError:      true,
			errorContains: "expected",
			description:   "nested missing closing parenthesis",
		},
		{
			name:          "mismatched nesting",
			input:         "(((42))",
			hasError:      true,
			description:   "unbalanced nested parentheses",
		},
		{
			name:          "invalid expression in group",
			input:         "(def)",
			hasError:      true,
			description:   "invalid keyword in group",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseGroupExpression(t, test.input)

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

			// For successful cases, just verify we got a valid expression
			if expr == nil {
				t.Error("Expression should not be nil")
			}
		})
	}
}

// Test special parentheses cases (empty, tuples)
func TestSpecialParenthesesCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		expectedType string
		description  string
	}{
		{
			name:         "empty parentheses",
			input:        "()",
			expectedType: "tuple",
			description:  "empty parentheses create empty tuple",
		},
		{
			name:         "single element with comma",
			input:        "(x,)",
			expectedType: "tuple",
			description:  "single element tuple with trailing comma",
		},
		{
			name:         "multiple elements",
			input:        "(a, b, c)",
			expectedType: "tuple",
			description:  "multi-element tuple",
		},
		{
			name:         "single element without comma",
			input:        "(x)",
			expectedType: "group",
			description:  "single element without comma is grouping",
		},
		{
			name:         "nested tuple in group",
			input:        "((a, b))",
			expectedType: "group",
			description:  "grouped tuple expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseGroupExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateExpressionType(t, expr, test.expectedType)
		})
	}
}