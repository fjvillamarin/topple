package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a tuple expression
func parseTupleExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.tuple()
}

// Helper function to parse an expression (may result in tuple)
func parseExpressionForTuple(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to validate tuple expression structure
func validateTupleExpression(t *testing.T, expr ast.Expr, expectedElements int, description string) *ast.TupleExpr {
	tuple, ok := expr.(*ast.TupleExpr)
	if !ok {
		t.Fatalf("Expected *ast.TupleExpr but got %T for %s", expr, description)
	}

	if expectedElements >= 0 && len(tuple.Elements) != expectedElements {
		t.Errorf("Expected %d elements but got %d for %s",
			expectedElements, len(tuple.Elements), description)
	}

	return tuple
}

// Helper function to validate generator expression structure
func validateGeneratorExpression(t *testing.T, expr ast.Expr, hasCondition bool, description string) *ast.GenExpr {
	gen, ok := expr.(*ast.GenExpr)
	if !ok {
		t.Fatalf("Expected *ast.GenExpr but got %T for %s", expr, description)
	}

	if gen.Element == nil {
		t.Errorf("Generator expression missing element for %s", description)
	}

	if len(gen.Clauses) == 0 {
		t.Errorf("Generator expression missing for clauses for %s", description)
	}

	// Check for condition
	actualHasCondition := false
	for _, clause := range gen.Clauses {
		if len(clause.Ifs) > 0 {
			actualHasCondition = true
			break
		}
	}

	if actualHasCondition != hasCondition {
		t.Errorf("Expected hasCondition=%v but got %v for %s",
			hasCondition, actualHasCondition, description)
	}

	return gen
}

// Helper function to count starred expressions in tuple
func countStarredExpressionsInTuple(tuple *ast.TupleExpr) int {
	count := 0
	for _, elem := range tuple.Elements {
		if _, ok := elem.(*ast.StarExpr); ok {
			count++
		}
	}
	return count
}

// Helper function to validate tuple parsing success
func validateTupleParseSuccess(t *testing.T, expr ast.Expr, err error, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if expr == nil {
		t.Fatalf("Expected expression but got nil for %s", description)
	}
}

// Helper function to validate tuple parsing error
func validateTupleParseError(t *testing.T, expr ast.Expr, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if expr != nil {
		t.Errorf("Expected nil expression on error, but got %T for %s", expr, description)
	}
}

// Helper function to check if expression is a group (parenthesized single expression)
func isGroupExpression(expr ast.Expr) bool {
	_, ok := expr.(*ast.GroupExpr)
	return ok
}

// Helper function to get expression type string
func getTupleExpressionType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.TupleExpr:
		return "TupleExpr"
	case *ast.GroupExpr:
		return "GroupExpr"
	case *ast.GenExpr:
		return "GenExpr"
	case *ast.ListComp:
		return "ListComp"
	default:
		return "unknown"
	}
}

// Helper function to analyze tuple elements for validation
func analyzeTupleElements(tuple *ast.TupleExpr) (starCount, nonStarCount int) {
	for _, elem := range tuple.Elements {
		if _, ok := elem.(*ast.StarExpr); ok {
			starCount++
		} else {
			nonStarCount++
		}
	}
	return
}

// Test comprehensive tuple expression parsing functionality
func TestTupleExpressions(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    string
		expectedElements int
		hasError        bool
		errorText       string
		description     string
	}{
		// Basic tuple expressions
		{
			name:            "empty tuple",
			input:           "()",
			expectedType:    "TupleExpr",
			expectedElements: 0,
			description:     "empty tuple expression",
		},
		{
			name:            "single element tuple with comma",
			input:           "(1,)",
			expectedType:    "TupleExpr",
			expectedElements: 1,
			description:     "single element tuple with trailing comma",
		},
		{
			name:            "two element tuple",
			input:           "(1, 2)",
			expectedType:    "TupleExpr",
			expectedElements: 2,
			description:     "simple two element tuple",
		},
		{
			name:            "multiple element tuple",
			input:           "(1, 2, 3, 4, 5)",
			expectedType:    "TupleExpr",
			expectedElements: 5,
			description:     "tuple with multiple elements",
		},
		{
			name:            "tuple with trailing comma",
			input:           "(1, 2, 3,)",
			expectedType:    "TupleExpr",
			expectedElements: 3,
			description:     "tuple with trailing comma",
		},

		// Expression elements
		{
			name:            "tuple with expressions",
			input:           "(x + 1, y * 2, z / 3)",
			expectedType:    "TupleExpr",
			expectedElements: 3,
			description:     "tuple with arithmetic expressions",
		},
		{
			name:            "tuple with function calls",
			input:           "(func(), method(), len(x))",
			expectedType:    "TupleExpr",
			expectedElements: 3,
			description:     "tuple with function call elements",
		},
		{
			name:            "tuple with mixed types",
			input:           "(1, 'hello', True, None)",
			expectedType:    "TupleExpr",
			expectedElements: 4,
			description:     "tuple with mixed literal types",
		},
		{
			name:            "nested tuples",
			input:           "((1, 2), (3, 4), (5, 6))",
			expectedType:    "TupleExpr",
			expectedElements: 3,
			description:     "tuple containing nested tuples",
		},

		// Complex expressions
		{
			name:            "tuple with comprehensions",
			input:           "([x for x in range(3)], {y for y in range(2)})",
			expectedType:    "TupleExpr",
			expectedElements: 2,
			description:     "tuple containing list and set comprehensions",
		},
		{
			name:            "tuple with lambda expressions",
			input:           "(lambda x: x + 1, lambda y: y * 2)",
			expectedType:    "TupleExpr",
			expectedElements: 2,
			description:     "tuple containing lambda expressions",
		},
		{
			name:            "tuple with conditional expressions",
			input:           "(x if condition else y, a if flag else b)",
			expectedType:    "TupleExpr",
			expectedElements: 2,
			description:     "tuple with conditional expressions",
		},

		// Parenthesized expressions (not tuples)
		{
			name:            "parenthesized single expression",
			input:           "(42)",
			expectedType:    "GroupExpr",
			expectedElements: -1, // Not applicable for GroupExpr
			description:     "single parenthesized expression without comma",
		},
		{
			name:            "parenthesized complex expression",
			input:           "(x + y * z)",
			expectedType:    "GroupExpr",
			expectedElements: -1, // Not applicable for GroupExpr
			description:     "parenthesized arithmetic expression",
		},

		// Generator expressions
		{
			name:            "simple generator expression",
			input:           "(x for x in range(10))",
			expectedType:    "GenExpr",
			expectedElements: -1, // Not applicable for GenExpr
			description:     "basic generator expression",
		},
		{
			name:            "generator with condition",
			input:           "(x for x in items if x > 0)",
			expectedType:    "GenExpr",
			expectedElements: -1, // Not applicable for GenExpr
			description:     "generator expression with condition",
		},

		// Error cases
		{
			name:        "unclosed tuple",
			input:       "(1, 2, 3",
			hasError:    true,
			errorText:   "expected",
			description: "tuple missing closing parenthesis",
		},
		{
			name:        "empty element in tuple",
			input:       "(1, , 3)",
			hasError:    true,
			errorText:   "expected",
			description: "tuple with empty element between commas",
		},
		{
			name:        "invalid expression in tuple",
			input:       "(1, def, 3)",
			hasError:    true,
			errorText:   "unexpected",
			description: "tuple with invalid expression element",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTupleExpression(t, test.input)

			if test.hasError {
				validateTupleParseError(t, expr, err, test.errorText, test.description)
				return
			}

			validateTupleParseSuccess(t, expr, err, test.description)

			// Validate expression type
			actualType := getTupleExpressionType(expr)
			if actualType != test.expectedType {
				t.Errorf("Expected %s but got %s for %s",
					test.expectedType, actualType, test.description)
			}

			// Validate element count for tuples
			if test.expectedType == "TupleExpr" && test.expectedElements >= 0 {
				validateTupleExpression(t, expr, test.expectedElements, test.description)
			}
		})
	}
}

// Test tuple expressions without parentheses (comma-separated expressions)
func TestTupleWithoutParentheses(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedElements int
		expectTuple     bool
		hasError        bool
		errorText       string
		description     string
	}{
		// Single expressions (parser.expression() behavior)
		{
			name:            "single expression",
			input:           "42",
			expectedElements: 1,
			expectTuple:     false,
			description:     "single expression without comma",
		},
		{
			name:            "single complex expression",
			input:           "obj.method().result[0]",
			expectedElements: 1,
			expectTuple:     false,
			description:     "single complex expression",
		},
		{
			name:            "first element of comma expression",
			input:           "1, 2",
			expectedElements: 1,
			expectTuple:     false,
			description:     "parser.expression() typically returns first element",
		},

		// Note: parser.expression() doesn't typically create tuples
		// This is context-dependent and tuples are usually created
		// in specific parsing contexts like assignments or function parameters,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseExpressionForTuple(t, test.input)

			if test.hasError {
				validateTupleParseError(t, expr, err, test.errorText, test.description)
				return
			}

			validateTupleParseSuccess(t, expr, err, test.description)

			if test.expectTuple {
				tuple := validateTupleExpression(t, expr, test.expectedElements, test.description)
				if tuple == nil {
					t.Errorf("Expected tuple but validation failed for %s", test.description)
				}
			} else {
				// Should not be a tuple
				if _, ok := expr.(*ast.TupleExpr); ok {
					t.Errorf("Expected non-tuple expression but got TupleExpr for %s", test.description)
				}
			}
		})
	}
}

// Test generator expressions parsing functionality
func TestGeneratorExpressions(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		hasCondition    bool
		expectedClauses int
		hasError        bool
		errorText       string
		description     string
	}{
		// Basic generator expressions
		{
			name:            "simple generator",
			input:           "(x for x in range(10))",
			hasCondition:    false,
			expectedClauses: 1,
			description:     "basic generator expression",
		},
		{
			name:            "generator with condition",
			input:           "(x for x in items if x > 0)",
			hasCondition:    true,
			expectedClauses: 1,
			description:     "generator expression with filter condition",
		},
		{
			name:            "expression transformation",
			input:           "(x * 2 for x in numbers)",
			hasCondition:    false,
			expectedClauses: 1,
			description:     "generator with expression transformation",
		},

		// Multiple generators
		{
			name:            "nested generators",
			input:           "(x + y for x in range(3) for y in range(3))",
			hasCondition:    false,
			expectedClauses: 2,
			description:     "generator with multiple for clauses",
		},
		{
			name:            "complex nested generators",
			input:           "(f(x, y, z) for x in data for y in items for z in values)",
			hasCondition:    false,
			expectedClauses: 3,
			description:     "generator with three nested for clauses",
		},

		// Complex generators
		{
			name:            "generator with complex condition",
			input:           "(process(x) for x in data if valid(x) and x.active)",
			hasCondition:    true,
			expectedClauses: 1,
			description:     "generator with complex boolean condition",
		},
		{
			name:            "generator with tuple unpacking",
			input:           "(x + y for x, y in pairs)",
			hasCondition:    false,
			expectedClauses: 1,
			description:     "generator with tuple unpacking in for clause",
		},
		{
			name:            "generator with attribute access",
			input:           "(item.value for item in objects if item.is_valid())",
			hasCondition:    true,
			expectedClauses: 1,
			description:     "generator with attribute access and method calls",
		},

		// Advanced generators
		{
			name:            "generator with subscript access",
			input:           "(data[key] for key in keys if key in data)",
			hasCondition:    true,
			expectedClauses: 1,
			description:     "generator with subscript access and membership test",
		},
		{
			name:            "generator with method chains",
			input:           "(obj.process().result for obj in items)",
			hasCondition:    false,
			expectedClauses: 1,
			description:     "generator with method call chains",
		},
		{
			name:            "generator with lambda",
			input:           "((lambda x: x * 2)(item) for item in values)",
			hasCondition:    false,
			expectedClauses: 1,
			description:     "generator with lambda expression",
		},

		// Error cases
		{
			name:        "incomplete generator",
			input:       "(x for)",
			hasError:    true,
			errorText:   "expected",
			description: "generator missing variable and iterable",
		},
		{
			name:        "missing in keyword",
			input:       "(x for x range(10))",
			hasError:    true,
			errorText:   "expected",
			description: "generator missing 'in' keyword",
		},
		{
			name:        "incomplete condition",
			input:       "(x for x in items if)",
			hasError:    true,
			errorText:   "expected",
			description: "generator with incomplete condition",
		},
		{
			name:        "invalid generator syntax",
			input:       "(for x in items)",
			hasError:    true,
			errorText:   "expected",
			description: "generator with missing element expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTupleExpression(t, test.input)

			if test.hasError {
				validateTupleParseError(t, expr, err, test.errorText, test.description)
				return
			}

			validateTupleParseSuccess(t, expr, err, test.description)
			gen := validateGeneratorExpression(t, expr, test.hasCondition, test.description)

			// Validate clause count
			if len(gen.Clauses) != test.expectedClauses {
				t.Errorf("Expected %d clauses but got %d for %s",
					test.expectedClauses, len(gen.Clauses), test.description)
			}
		})
	}
}

// Test starred expressions in tuples
func TestTupleWithStarredExpressions(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedStars   int
		expectedElements int
		hasError        bool
		errorText       string
		description     string
	}{
		// Basic starred expressions
		{
			name:            "starred at beginning",
			input:           "(*first, 2, 3)",
			expectedStars:   1,
			expectedElements: 3,
			description:     "tuple with starred element at beginning",
		},
		{
			name:            "starred at end",
			input:           "(1, 2, *rest)",
			expectedStars:   1,
			expectedElements: 3,
			description:     "tuple with starred element at end",
		},
		{
			name:            "starred in middle",
			input:           "(1, *middle, 3)",
			expectedStars:   1,
			expectedElements: 3,
			description:     "tuple with starred element in middle",
		},

		// Multiple starred expressions
		{
			name:            "multiple starred elements",
			input:           "(*first, *second, 3)",
			expectedStars:   2,
			expectedElements: 3,
			description:     "tuple with multiple starred elements",
		},
		{
			name:            "all starred elements",
			input:           "(*a, *b, *c)",
			expectedStars:   3,
			expectedElements: 3,
			description:     "tuple with all starred elements",
		},

		// Complex starred expressions
		{
			name:            "starred generator expression",
			input:           "(*(x for x in range(3)), 4)",
			expectedStars:   1,
			expectedElements: 2,
			description:     "tuple with starred generator expression",
		},
		{
			name:            "starred function call",
			input:           "(*get_items(), 'extra')",
			expectedStars:   1,
			expectedElements: 2,
			description:     "tuple with starred function call",
		},
		{
			name:            "starred attribute access",
			input:           "(*obj.items, *obj.extras)",
			expectedStars:   2,
			expectedElements: 2,
			description:     "tuple with starred attribute accesses",
		},
		{
			name:            "starred complex expressions",
			input:           "(*process(data), *filter(items), result)",
			expectedStars:   2,
			expectedElements: 3,
			description:     "tuple with complex starred expressions",
		},

		// Mixed starred and regular elements
		{
			name:            "mixed starred and regular",
			input:           "(first, *middle, 'literal', func(), *rest)",
			expectedStars:   2,
			expectedElements: 5,
			description:     "tuple with mix of starred and regular elements",
		},
		{
			name:            "single starred element",
			input:           "(*items,)",
			expectedStars:   1,
			expectedElements: 1,
			description:     "single starred element tuple",
		},

		// Error cases
		{
			name:        "invalid starred syntax",
			input:       "(*)",
			hasError:    true,
			errorText:   "expected",
			description: "starred expression without operand",
		},
		{
			name:        "incomplete starred expression",
			input:       "(*, 2, 3)",
			hasError:    true,
			errorText:   "expected",
			description: "incomplete starred expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTupleExpression(t, test.input)

			if test.hasError {
				validateTupleParseError(t, expr, err, test.errorText, test.description)
				return
			}

			validateTupleParseSuccess(t, expr, err, test.description)
			tuple := validateTupleExpression(t, expr, test.expectedElements, test.description)

			// Count starred expressions
			starCount := countStarredExpressionsInTuple(tuple)
			if starCount != test.expectedStars {
				t.Errorf("Expected %d starred expressions but got %d for %s",
					test.expectedStars, starCount, test.description)
			}

			// Validate that we have the expected number of starred elements
			if test.expectedStars > 0 && starCount == 0 {
				t.Errorf("Expected at least one starred expression for %s", test.description)
			}
		})
	}
}

// Test tuple error cases and edge conditions
func TestTupleErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorText   string
		description string
	}{
		{
			name:        "unclosed tuple",
			input:       "(1, 2, 3",
			errorText:   "expected",
			description: "tuple missing closing parenthesis",
		},
		{
			name:        "missing comma between elements",
			input:       "(1 2 3)",
			errorText:   "expected",
			description: "tuple elements without comma separator",
		},
		{
			name:        "empty element",
			input:       "(1, , 3)",
			errorText:   "expected",
			description: "tuple with empty element between commas",
		},
		{
			name:        "invalid generator syntax",
			input:       "(x for)",
			errorText:   "expected",
			description: "generator missing variable and iterable",
		},
		{
			name:        "missing in keyword",
			input:       "(x for x range(10))",
			errorText:   "expected",
			description: "generator missing 'in' keyword",
		},
		{
			name:        "incomplete condition",
			input:       "(x for x in items if)",
			errorText:   "expected",
			description: "generator with incomplete condition",
		},
		{
			name:        "invalid starred syntax",
			input:       "(*)",
			errorText:   "expected",
			description: "starred expression without operand",
		},
		{
			name:        "invalid element expression",
			input:       "(1, def, 3)",
			errorText:   "unexpected",
			description: "tuple with invalid expression element",
		},
		{
			name:        "unclosed generator",
			input:       "(x for x in items",
			errorText:   "expected",
			description: "generator expression missing closing parenthesis",
		},
		{
			name:        "malformed nested structure",
			input:       "((1, 2,), (3, 4)",
			errorText:   "expected",
			description: "nested tuple with malformed syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTupleExpression(t, test.input)
			validateTupleParseError(t, expr, err, test.errorText, test.description)
		})
	}
}

// Test edge cases and complex tuple scenarios
func TestTupleEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		// Complex valid cases
		{
			name: "deeply nested tuples",
			input: "(((1, 2), (3, 4)), ((5, 6), (7, 8)))",
			description: "deeply nested tuple structures",
		},
		{
			name: "tuple with comprehensions",
			input: "([x for x in range(3)], {y: y**2 for y in range(2)})",
			description: "tuple containing list and dict comprehensions",
		},
		{
			name: "tuple with lambda expressions",
			input: "(lambda x: x + 1, lambda y: y * 2, lambda z: z / 3)",
			description: "tuple containing multiple lambda expressions",
		},
		{
			name: "tuple with conditional expressions",
			input: "(x if condition else y, a if flag else b, c if test else d)",
			description: "tuple with multiple conditional expressions",
		},
		{
			name: "tuple with async expressions",
			input: "(await func1(), await func2(), regular_func())",
			description: "tuple with async await expressions",
		},
		{
			name: "tuple with star expressions",
			input: "(func(*args), method(**kwargs), normal_call())",
			description: "tuple with star expression function calls",
		},
		{
			name: "complex generator with multiple conditions",
			input: "(process(x, y) for x in data for y in items if x.valid and y.active)",
			description: "generator with multiple for clauses and conditions",
		},
		{
			name: "tuple with yield expressions",
			input: "(yield value1, yield value2, yield value3)",
			description: "tuple containing yield expressions",
		},

		// Edge error cases
		{
			name: "nested syntax error",
			input: "((1, 2), (3, def))",
			hasError: true,
			errorText: "unexpected",
			description: "syntax error in nested tuple",
		},
		{
			name: "generator with invalid expression",
			input: "(def for x in items)",
			hasError: true,
			errorText: "unexpected",
			description: "generator with invalid element expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTupleExpression(t, test.input)

			if test.hasError {
				validateTupleParseError(t, expr, err, test.errorText, test.description)
			} else {
				validateTupleParseSuccess(t, expr, err, test.description)

				// For successful cases, verify we got a reasonable expression
				exprType := getTupleExpressionType(expr)
				if exprType == "unknown" {
					t.Errorf("Got unknown expression type for %s (actual: %T)", test.description, expr)
				}
			}
		})
	}
}