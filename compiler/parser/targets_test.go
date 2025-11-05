package parser

import (
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse target using specific method
func parseTargetWithMethod(t *testing.T, input string, method string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)

	switch method {
	case "tPrimary":
		return parser.tPrimary()
	case "singleSubscriptAttributeTarget":
		return parser.singleSubscriptAttributeTarget()
	case "singleTarget":
		return parser.singleTarget()
	case "starTarget":
		return parser.starTarget()
	case "starAtom":
		return parser.starAtom()
	case "targetWithStarAtom":
		return parser.targetWithStarAtom()
	default:
		t.Fatalf("Unknown parser method: %s", method)
		return nil, nil
	}
}

// Helper function to parse star targets for target validation
func parseStarTargetsForTargets(t *testing.T, input string) ([]ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.starTargets()
}

// Helper function to check target lookahead
func checkTargetLookahead(t *testing.T, input string) bool {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.tLookahead()
}

// Helper function to get target node type
func getTargetNodeType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.Name:
		return "name"
	case *ast.Attribute:
		return "attribute"
	case *ast.Subscript:
		return "subscript"
	case *ast.Call:
		return "call"
	case *ast.GroupExpr:
		return "group"
	case *ast.StarExpr:
		return "star"
	case *ast.TupleExpr:
		return "tuple"
	case *ast.ListExpr:
		return "list"
	case *ast.Literal:
		return "literal"
	default:
		return "unknown"
	}
}

// Helper function to get element count from containers
func getTargetElementCount(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.TupleExpr:
		return len(e.Elements)
	case *ast.ListExpr:
		return len(e.Elements)
	default:
		return 0
	}
}

// Helper function to validate target parsing success
func validateTargetParseSuccess(t *testing.T, expr ast.Expr, err error, expectedType string, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if expr == nil {
		t.Fatalf("Expected target expression but got nil for %s", description)
	}

	actualType := getTargetNodeType(expr)
	if actualType != expectedType {
		t.Errorf("Expected %s target type but got %s (actual: %T) for %s",
			expectedType, actualType, expr, description)
	}
}

// Helper function to validate target parsing error
func validateTargetParseError(t *testing.T, expr ast.Expr, err error, expectedErrorText string, description string) {
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

// Helper function to validate star target properties
func validateStarTarget(t *testing.T, expr ast.Expr, expectStar bool, description string) {
	if expectStar {
		if _, ok := expr.(*ast.StarExpr); !ok {
			t.Errorf("Expected star expression but got %T for %s", expr, description)
		}
	} else {
		if _, ok := expr.(*ast.StarExpr); ok {
			t.Errorf("Expected non-star expression but got star for %s", description)
		}
	}
}

// Helper function to validate container element count
func validateElementCount(t *testing.T, expr ast.Expr, expectedCount int, description string) {
	actualCount := getTargetElementCount(expr)
	if actualCount != expectedCount {
		t.Errorf("Expected %d elements but got %d for %s",
			expectedCount, actualCount, description)
	}
}

// Test target parsing with various methods
func TestTargetParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		method       string
		expectedType string
		hasError     bool
		errorText    string
		isStar       bool
		elemCount    int
		description  string
	}{
		// tPrimary tests
		{
			name:         "chained attribute access",
			input:        "obj.attr.nested.value",
			method:       "tPrimary",
			expectedType: "attribute",
			description:  "tPrimary with deeply nested attribute access",
		},
		{
			name:         "chained subscript access",
			input:        "arr[0][1][2]",
			method:       "tPrimary",
			expectedType: "subscript",
			description:  "tPrimary with nested subscript access",
		},
		{
			name:         "mixed access pattern",
			input:        "obj.method()[0].attr",
			method:       "tPrimary",
			expectedType: "attribute",
			description:  "tPrimary with mixed method call, subscript, and attribute",
		},
		{
			name:        "atom without lookahead",
			input:       "x",
			method:      "tPrimary",
			hasError:    true,
			errorText:   "expected",
			description: "tPrimary should require accessor tokens",
		},

		// singleSubscriptAttributeTarget tests
		{
			name:         "method call with attribute",
			input:        "obj.method().result",
			method:       "singleSubscriptAttributeTarget",
			expectedType: "attribute",
			description:  "singleSubscriptAttributeTarget with method call and attribute",
		},
		{
			name:         "chained subscript access",
			input:        "matrix[row][col]",
			method:       "singleSubscriptAttributeTarget",
			expectedType: "subscript",
			description:  "singleSubscriptAttributeTarget with nested subscripts",
		},
		{
			name:         "simple attribute access",
			input:        "obj.attr",
			method:       "singleSubscriptAttributeTarget",
			expectedType: "attribute",
			description:  "singleSubscriptAttributeTarget with simple attribute",
		},

		// singleTarget tests
		{
			name:         "simple name target",
			input:        "x",
			method:       "singleTarget",
			expectedType: "name",
			description:  "singleTarget with simple identifier",
		},
		{
			name:         "parenthesized target",
			input:        "(x)",
			method:       "singleTarget",
			expectedType: "group",
			description:  "singleTarget with parenthesized identifier",
		},
		{
			name:         "attribute target",
			input:        "obj.attr",
			method:       "singleTarget",
			expectedType: "attribute",
			description:  "singleTarget with attribute access",
		},
		{
			name:         "subscript target",
			input:        "arr[0]",
			method:       "singleTarget",
			expectedType: "subscript",
			description:  "singleTarget with subscript access",
		},
		{
			name:         "complex chained target",
			input:        "obj.method()[key].value",
			method:       "singleTarget",
			expectedType: "attribute",
			description:  "singleTarget with complex access chain",
		},

		// starTarget tests
		{
			name:         "regular name target",
			input:        "x",
			method:       "starTarget",
			expectedType: "name",
			isStar:       false,
			description:  "starTarget with regular identifier",
		},
		{
			name:         "starred name target",
			input:        "*x",
			method:       "starTarget",
			expectedType: "star",
			isStar:       true,
			description:  "starTarget with starred identifier",
		},
		{
			name:         "starred attribute target",
			input:        "*obj.attr",
			method:       "starTarget",
			expectedType: "star",
			isStar:       true,
			description:  "starTarget with starred attribute access",
		},
		{
			name:         "starred subscript target",
			input:        "*items[key]",
			method:       "starTarget",
			expectedType: "star",
			isStar:       true,
			description:  "starTarget with starred subscript access",
		},
		{
			name:        "double star error",
			input:       "**x",
			method:      "starTarget",
			hasError:    true,
			errorText:   "expected",
			description: "starTarget should reject double star syntax",
		},

		// starAtom tests
		{
			name:         "simple name atom",
			input:        "x",
			method:       "starAtom",
			expectedType: "name",
			description:  "starAtom with simple identifier",
		},
		{
			name:         "parenthesized target atom",
			input:        "(x)",
			method:       "starAtom",
			expectedType: "group",
			description:  "starAtom with parenthesized target",
		},
		{
			name:         "empty tuple atom",
			input:        "()",
			method:       "starAtom",
			expectedType: "tuple",
			elemCount:    0,
			description:  "starAtom with empty tuple",
		},
		{
			name:         "tuple with elements",
			input:        "(x, y, z)",
			method:       "starAtom",
			expectedType: "tuple",
			elemCount:    3,
			description:  "starAtom with tuple containing elements",
		},
		{
			name:         "empty list atom",
			input:        "[]",
			method:       "starAtom",
			expectedType: "list",
			elemCount:    0,
			description:  "starAtom with empty list",
		},
		{
			name:         "list with elements",
			input:        "[x, y, z]",
			method:       "starAtom",
			expectedType: "list",
			elemCount:    3,
			description:  "starAtom with list containing elements",
		},
		{
			name:         "tuple with starred element",
			input:        "(x, *y, z)",
			method:       "starAtom",
			expectedType: "tuple",
			elemCount:    3,
			description:  "starAtom with tuple containing starred element",
		},
		{
			name:         "list with starred element",
			input:        "[a, *b, c]",
			method:       "starAtom",
			expectedType: "list",
			elemCount:    3,
			description:  "starAtom with list containing starred element",
		},

		// targetWithStarAtom tests
		{
			name:         "simple name with star atom",
			input:        "x",
			method:       "targetWithStarAtom",
			expectedType: "name",
			description:  "targetWithStarAtom with simple identifier",
		},
		{
			name:         "chained attribute with star atom",
			input:        "obj.method().result",
			method:       "targetWithStarAtom",
			expectedType: "attribute",
			description:  "targetWithStarAtom with complex attribute chain",
		},
		{
			name:         "subscript with star atom",
			input:        "data[key][index]",
			method:       "targetWithStarAtom",
			expectedType: "subscript",
			description:  "targetWithStarAtom with nested subscript access",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTargetWithMethod(t, test.input, test.method)

			if test.hasError {
				validateTargetParseError(t, expr, err, test.errorText, test.description)
				return
			}

			validateTargetParseSuccess(t, expr, err, test.expectedType, test.description)

			// Additional validations based on test properties
			if test.isStar || test.expectedType == "star" {
				validateStarTarget(t, expr, true, test.description)
			} else {
				validateStarTarget(t, expr, false, test.description)
			}

			// Validate element count for containers
			if test.elemCount > 0 || (test.elemCount == 0 && (test.expectedType == "tuple" || test.expectedType == "list")) {
				validateElementCount(t, expr, test.elemCount, test.description)
			}
		})
	}
}

// Test star target sequences and comma-separated target lists
func TestStarTargetSequences(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		hasError      bool
		errorText     string
		description   string
	}{
		{
			name:          "single target",
			input:         "x",
			expectedCount: 1,
			description:   "single identifier target",
		},
		{
			name:          "single target with trailing comma",
			input:         "x,",
			expectedCount: 1,
			description:   "single target with trailing comma",
		},
		{
			name:          "two simple targets",
			input:         "x, y",
			expectedCount: 2,
			description:   "two identifier targets",
		},
		{
			name:          "three targets with trailing comma",
			input:         "x, y, z,",
			expectedCount: 3,
			description:   "three targets with trailing comma",
		},
		{
			name:          "starred targets",
			input:         "*x, y, *z",
			expectedCount: 3,
			description:   "mixed starred and regular targets",
			hasError:      true, // Multiple starred expressions not allowed in Python
			errorText:     "multiple starred",
		},
		{
			name:          "complex mixed targets",
			input:         "a, *b, c.attr, d[0]",
			expectedCount: 4,
			description:   "mix of simple, starred, attribute, and subscript targets",
		},
		{
			name:          "deeply nested targets",
			input:         "obj.method()[0].attr, *items, data[key]",
			expectedCount: 3,
			description:   "complex nested target expressions",
		},
		{
			name:          "starred first element",
			input:         "*first, second, third",
			expectedCount: 3,
			description:   "starred element at beginning of sequence",
		},
		{
			name:          "starred last element",
			input:         "first, second, *rest",
			expectedCount: 3,
			description:   "starred element at end of sequence",
		},
		{
			name:          "parenthesized targets",
			input:         "(a, b), (c, d), *others",
			expectedCount: 3,
			description:   "parenthesized tuple targets with starred element",
		},
		{
			name:          "list targets",
			input:         "[a, b], [c, d], *remaining",
			expectedCount: 3,
			description:   "list targets with starred element",
		},
		{
			name:          "attribute and method targets",
			input:         "obj.attr, instance.method(), *extras",
			expectedCount: 3,
			description:   "attribute and method call targets",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targets, err := parseStarTargetsForTargets(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if len(targets) != test.expectedCount {
				t.Errorf("Expected %d targets but got %d for %s",
					test.expectedCount, len(targets), test.description)
			}

			// Validate that each target is a valid expression
			for i, target := range targets {
				if target == nil {
					t.Errorf("Target %d is nil for %s", i, test.description)
				}
			}
		})
	}
}

// Test target lookahead functionality
func TestTargetLookahead(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    bool
		description string
	}{
		{
			name:        "dot lookahead",
			input:       ".attr",
			expected:    true,
			description: "lookahead should detect dot for attribute access",
		},
		{
			name:        "bracket lookahead",
			input:       "[0]",
			expected:    true,
			description: "lookahead should detect bracket for subscript access",
		},
		{
			name:        "paren lookahead",
			input:       "()",
			expected:    true,
			description: "lookahead should detect paren for function call",
		},
		{
			name:        "no lookahead for identifier",
			input:       "x",
			expected:    false,
			description: "lookahead should not trigger for plain identifier",
		},
		{
			name:        "no lookahead for number",
			input:       "42",
			expected:    false,
			description: "lookahead should not trigger for numeric literal",
		},
		{
			name:        "no lookahead for string",
			input:       "'hello'",
			expected:    false,
			description: "lookahead should not trigger for string literal",
		},
		{
			name:        "empty input",
			input:       "",
			expected:    false,
			description: "lookahead should not trigger for empty input",
		},
		{
			name:        "operator lookahead",
			input:       "+",
			expected:    false,
			description: "lookahead should not trigger for operators",
		},
		{
			name:        "keyword lookahead",
			input:       "def",
			expected:    false,
			description: "lookahead should not trigger for keywords",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checkTargetLookahead(t, test.input)

			if result != test.expected {
				t.Errorf("tLookahead() = %v, expected %v for %s",
					result, test.expected, test.description)
			}
		})
	}
}

// Test target edge cases and error conditions
func TestTargetEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		method      string
		hasError    bool
		errorText   string
		description string
	}{
		// Complex valid cases
		{
			name:        "deeply nested attribute access",
			input:       "a.b.c.d.e.f",
			method:      "singleTarget",
			description: "very deep attribute chain",
		},
		{
			name:        "deeply nested subscript access",
			input:       "matrix[i][j][k][l]",
			method:      "singleTarget",
			description: "very deep subscript chain",
		},
		{
			name:        "complex mixed access",
			input:       "obj.method()[key].attr().value[0]",
			method:      "singleTarget",
			description: "complex chain of method calls, subscripts, and attributes",
		},
		{
			name:        "starred complex target",
			input:       "*obj.complex[key].method().result",
			method:      "starTarget",
			description: "starred expression with complex target",
		},
		{
			name:        "tuple with complex elements",
			input:       "(obj.attr, data[key], func().result)",
			method:      "starAtom",
			description: "tuple containing complex target expressions",
		},
		{
			name:        "list with starred elements",
			input:       "[first, *middle, last]",
			method:      "starAtom",
			description: "list with starred unpacking in middle",
		},

		// Error cases
		{
			name:        "invalid double star",
			input:       "**invalid",
			method:      "starTarget",
			hasError:    true,
			errorText:   "expected",
			description: "double star should be invalid in target context",
		},
		{
			name:        "incomplete attribute access",
			input:       "obj.",
			method:      "singleTarget",
			hasError:    true,
			errorText:   "expected",
			description: "incomplete attribute access missing identifier",
		},
		{
			name:        "incomplete subscript access",
			input:       "arr[",
			method:      "singleTarget",
			hasError:    true,
			errorText:   "expected",
			description: "incomplete subscript access missing closing bracket",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTargetWithMethod(t, test.input, test.method)

			if test.hasError {
				validateTargetParseError(t, expr, err, test.errorText, test.description)
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for %s: %v", test.description, err)
				}
				if expr == nil {
					t.Errorf("Expected valid expression for %s", test.description)
				}
			}
		})
	}
}

// Test comprehensive star target validation
func TestStarTargetValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStars int
		hasError      bool
		errorText     string
		description   string
	}{
		{
			name:          "no starred elements",
			input:         "a, b, c",
			expectedStars: 0,
			description:   "sequence with no starred elements",
		},
		{
			name:          "single starred element",
			input:         "a, *b, c",
			expectedStars: 1,
			description:   "sequence with one starred element",
		},
		{
			name:          "starred at beginning",
			input:         "*first, second, third",
			expectedStars: 1,
			description:   "starred element at beginning of sequence",
		},
		{
			name:          "starred at end",
			input:         "first, second, *rest",
			expectedStars: 1,
			description:   "starred element at end of sequence",
		},
		{
			name:          "multiple starred elements",
			input:         "*a, b, *c, d",
			expectedStars: 2,
			description:   "sequence with multiple starred elements",
			hasError:      true, // Multiple starred expressions not allowed in Python
			errorText:     "multiple starred",
		},
		{
			name:          "starred complex expressions",
			input:         "*obj.attr, regular, *data[key]",
			expectedStars: 2,
			description:   "starred elements with complex expressions",
			hasError:      true, // Multiple starred expressions not allowed in Python
			errorText:     "multiple starred",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targets, err := parseStarTargetsForTargets(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			// Count starred expressions
			starCount := 0
			for _, target := range targets {
				if _, ok := target.(*ast.StarExpr); ok {
					starCount++
				}
			}

			if starCount != test.expectedStars {
				t.Errorf("Expected %d starred elements but got %d for %s",
					test.expectedStars, starCount, test.description)
			}

			// Additional validation that non-starred elements are also valid
			for i, target := range targets {
				if target == nil {
					t.Errorf("Target %d is nil in %s", i, test.description)
				}
			}
		})
	}
}

// Test target parsing with complex nested structures
func TestTargetComplexNesting(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		method      string
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "deeply nested attribute chain",
			input:       "obj.module.klass.instance.attribute.value",
			method:      "singleTarget",
			description: "very deep attribute access chain in target",
		},
		{
			name:        "mixed access with method calls",
			input:       "obj.get_instance().data[key].process().result",
			method:      "singleTarget",
			description: "complex mixed access pattern in target",
		},
		{
			name:        "nested subscript and slice",
			input:       "matrix[row][col:end][index]",
			method:      "singleTarget",
			description: "nested subscript with slice notation",
		},
		{
			name:        "starred complex nested expression",
			input:       "*obj.method()[key].attr",
			method:      "starTarget",
			description: "starred expression with complex nested access",
		},
		{
			name:        "parenthesized complex target",
			input:       "(obj.method().result[index])",
			method:      "singleTarget",
			description: "parenthesized complex target expression",
		},
		{
			name:        "tuple with complex elements",
			input:       "(obj.attr, data[key], func().result)",
			method:      "starAtom",
			description: "tuple containing complex target expressions",
		},
		{
			name:        "list with mixed targets",
			input:       "[simple, obj.attr, *items, data[key]]",
			method:      "starAtom",
			description: "list with mix of simple and complex targets",
		},
		{
			name:        "nested tuple unpacking",
			input:       "((a, b), (c, d))",
			method:      "starAtom",
			description: "nested tuple structure for unpacking",
		},
		{
			name:        "list with starred and nested",
			input:       "[*first, (middle1, middle2), *last]",
			method:      "starAtom",
			description: "list with starred elements and nested tuples",
		},
		{
			name:        "complex attribute with slicing",
			input:       "obj.items[start:end:step].value",
			method:      "singleTarget",
			description: "attribute access with complex slicing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTargetWithMethod(t, test.input, test.method)

			if test.hasError {
				validateTargetParseError(t, expr, err, test.errorText, test.description)
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if expr == nil {
				t.Errorf("Expected valid expression for %s", test.description)
			}

			// Verify the expression is a reasonable target type
			targetType := getTargetNodeType(expr)
			if targetType == "unknown" {
				t.Errorf("Got unknown target type for %s (actual: %T)", test.description, expr)
			}
		})
	}
}
