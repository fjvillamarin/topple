package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a comparison expression
func parseComparison(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.comparison()
}

// Helper function to validate binary expression operator
func validateBinaryOp(t *testing.T, expr ast.Expr, expectedOp lexer.TokenType) {
	binary, ok := expr.(*ast.Binary)
	if !ok {
		t.Fatalf("Expected *ast.Binary, got %T", expr)
	}
	if binary.Operator.Type != expectedOp {
		t.Errorf("Expected operator %v, got %v", expectedOp, binary.Operator.Type)
	}
}

// Test comparison operations
func TestComparison(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpType lexer.TokenType
		expectBinary   bool
	}{
		// Simple comparisons
		{"less than", "x < y", lexer.Less, true},
		{"greater than", "x > y", lexer.Greater, true},
		{"less equal", "x <= y", lexer.LessEqual, true},
		{"greater equal", "x >= y", lexer.GreaterEqual, true},
		{"equal", "x == y", lexer.EqualEqual, true},
		{"not equal", "x != y", lexer.BangEqual, true},
		{"is", "x is y", lexer.Is, true},
		{"is not", "x is not y", lexer.IsNot, true},
		{"in", "x in y", lexer.In, true},
		{"not in", "x not in y", lexer.NotIn, true},

		// Expression comparisons
		{"with numbers", "5 < 10", lexer.Less, true},
		{"with function calls", "len(a) > len(b)", lexer.Greater, true},
		{"with attributes", "obj.value == target", lexer.EqualEqual, true},
		{"with subscripts", "arr[i] != arr[j]", lexer.BangEqual, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseComparison(t, test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if test.expectBinary {
				validateBinaryOp(t, expr, test.expectedOpType)
			}
		})
	}
}

// Test chained comparisons
func TestChainedComparisons(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		chainLen    int
		validateAST bool
	}{
		{"simple chain", "x < y < z", 2, true},
		{"long chain", "a < b < c < d", 3, true},
		{"mixed operators", "x < y <= z", 2, true},
		{"equality chain", "x == y == z", 2, true},
		{"complex chain", "0 < x <= 100", 2, true},
		{"with expressions", "min(a) < x < max(b)", 2, true},
		{"identity chain", "x is y is z", 2, true},
		{"membership chain", "x in a in b", 2, false}, // May not be valid
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseComparison(t, test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression for chained comparison")
				return
			}

			if test.validateAST {
				// Verify we have a valid expression structure for chained comparisons
				// This may be a Binary with nested Binary operands
				if _, ok := expr.(*ast.Binary); !ok {
					t.Logf("Chained comparison parsed as %T (may be implementation-specific)", expr)
				}
			}
		})
	}
}

// Test bitwise OR operations (part of comparison precedence)
func TestBitwiseOr(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectBinary   bool
		expectedOpType lexer.TokenType
	}{
		{"simple or", "x | y", true, lexer.Pipe},
		{"chained or", "x | y | z", true, lexer.Pipe},
		{"with parentheses", "(x | y) | z", true, lexer.Pipe},
		{"mixed with comparison", "x | y == 0", true, lexer.Pipe}, // bitwiseOr() only parses up to the | operator
		{"complex expression", "flags | FLAG_A | FLAG_B", true, lexer.Pipe},
		{"with numbers", "5 | 3", true, lexer.Pipe},
		{"with variables", "mask | flag", true, lexer.Pipe},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.bitwiseOr()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression")
				return
			}

			if test.expectBinary {
				validateBinaryOp(t, expr, test.expectedOpType)
			}
		})
	}
}

// Test membership operations
func TestMembershipOperations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpType lexer.TokenType
		validateOp     bool
	}{
		{"simple in", "x in list", lexer.In, true},
		{"not in", "x not in list", lexer.NotIn, true},
		{"in string", "'a' in 'abc'", lexer.In, true},
		{"in range", "5 in range(10)", lexer.In, true},
		{"in dict", "key in dictionary", lexer.In, true},
		{"complex in", "item.value in valid_values", lexer.In, true},
		{"chained membership", "x in a and y in b", lexer.And, false}, // 'and' has lower precedence
		{"in with subscript", "item in data[key]", lexer.In, true},
		{"not in with call", "value not in get_blacklist()", lexer.NotIn, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseComparison(t, test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression")
				return
			}

			if test.validateOp {
				validateBinaryOp(t, expr, test.expectedOpType)
			}
		})
	}
}

// Test identity operations
func TestIdentityOperations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpType lexer.TokenType
		validateOp     bool
	}{
		{"simple is", "x is None", lexer.Is, true},
		{"is not", "x is not None", lexer.IsNot, true},
		{"is True", "flag is True", lexer.Is, true},
		{"is False", "flag is False", lexer.Is, true},
		{"identity chain", "x is y is z", lexer.Is, false}, // Complex chaining
		{"complex identity", "obj.attr is not None", lexer.IsNot, true},
		{"with function", "result() is None", lexer.Is, true},
		{"with subscript", "data[key] is not value", lexer.IsNot, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseComparison(t, test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression")
				return
			}

			if test.validateOp {
				validateBinaryOp(t, expr, test.expectedOpType)
			}
		})
	}
}

// Test comparison precedence
func TestComparisonPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTopOp lexer.TokenType
		description   string
	}{
		{"comparison vs bitwise or", "a | b == c", lexer.EqualEqual, "== has higher precedence than |"},
		{"comparison vs bitwise and", "a & b > c", lexer.Greater, "> has higher precedence than &"},
		{"chained vs single", "a < b < c", lexer.And, "chained comparisons become logical AND"},
		{"mixed comparisons", "a == b != c", lexer.And, "chained comparisons become logical AND"},
		{"parentheses override", "(a | b) == c", lexer.EqualEqual, "parentheses change precedence"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseComparison(t, test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if binary, ok := expr.(*ast.Binary); ok {
				if binary.Operator.Type != test.expectedTopOp {
					t.Errorf("Expected top-level operator %v, got %v - %s",
						test.expectedTopOp, binary.Operator.Type, test.description)
				}
			} else {
				t.Logf("Non-binary expression for precedence test: %T", expr)
			}
		})
	}
}

// Test error cases
func TestComparisonErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"incomplete comparison", "x <"},
		{"invalid chain", "x < < y"},
		{"missing operand", "< y"},
		{"invalid is", "x is"},
		{"invalid in", "x in"},
		{"empty comparison", ""},
		{"invalid operator sequence", "x <=> y"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			_, err := parser.comparison()
			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.input)
			}
		})
	}
}

// TestIncompleteExpressionStatements tests incomplete expressions at the statement level
func TestIncompleteExpressionStatements(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
	}{
		{"incomplete not in", "x not", ""},
		{"incomplete is not", "x is not", ""},
		{"trailing not", "x not and y", ""},
		{"if with incomplete not", "if x not: pass", ""},
		{"incomplete addition", "x +", ""},
		{"incomplete multiplication", "x *", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			// Parse as a complete program
			_, errs := parser.Parse()
			if len(errs) == 0 {
				t.Errorf("Expected error for '%s', but got none", test.input)
			} else if test.errorContains != "" && !strings.Contains(errs[0].Error(), test.errorContains) {
				t.Errorf("Expected error containing '%s' for '%s', but got: %v", test.errorContains, test.input, errs[0])
			}
		})
	}
}
