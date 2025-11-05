package parser

import (
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Test sum operations (addition and subtraction)
func TestSum(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedOpType   lexer.TokenType
		shouldValidateOp bool
	}{
		{"simple addition", "x + y", lexer.Plus, true},
		{"simple subtraction", "x - y", lexer.Minus, true},
		{"chained addition", "x + y + z", lexer.Plus, false}, // outermost varies
		{"mixed operations", "x + y - z", lexer.Minus, true}, // validates outermost
		{"with numbers", "10 + 20", lexer.Plus, true},
		{"with parentheses", "(x + y) - z", lexer.Minus, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.sum()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Validate operator if specified
			if test.shouldValidateOp {
				if binary, ok := expr.(*ast.Binary); ok {
					if binary.Operator.Type != test.expectedOpType {
						t.Errorf("Expected operator %v, got %v", test.expectedOpType, binary.Operator.Type)
					}
				} else {
					t.Errorf("Expected binary expression for operator validation")
				}
			}
		})
	}
}

// Test term operations (multiplication, division, modulo)
func TestTerm(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedOpType   lexer.TokenType
		shouldValidateOp bool
	}{
		{"multiplication", "x * y", lexer.Star, true},
		{"division", "x / y", lexer.Slash, true},
		{"floor division", "x // y", lexer.SlashSlash, true},
		{"modulo", "x % y", lexer.Percent, true},
		{"chained multiply", "x * y * z", lexer.Star, false}, // outermost varies
		{"mixed operations", "x * y / z", lexer.Slash, true}, // validates outermost
		{"with parentheses", "(x * y) % z", lexer.Percent, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.term()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Validate operator if specified
			if test.shouldValidateOp {
				if binary, ok := expr.(*ast.Binary); ok {
					if binary.Operator.Type != test.expectedOpType {
						t.Errorf("Expected operator %v, got %v", test.expectedOpType, binary.Operator.Type)
					}
				} else {
					t.Errorf("Expected binary expression for operator validation")
				}
			}
		})
	}
}

// Test factor operations (unary operators)
func TestFactor(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedOpType   lexer.TokenType
		shouldValidateOp bool
	}{
		{"unary plus", "+x", lexer.Plus, true},
		{"unary minus", "-x", lexer.Minus, true},
		{"bitwise not", "~x", lexer.Tilde, true},
		{"double negative", "--x", lexer.Minus, true}, // outermost operator
		{"mixed unary", "-+x", lexer.Minus, true},     // outermost operator
		{"unary on number", "-42", lexer.Minus, true},
		{"unary on expression", "-(x + y)", lexer.Minus, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.factor()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Check for unary expression
			unary, ok := expr.(*ast.Unary)
			if !ok {
				t.Fatalf("Expected Unary, got %T", expr)
			}

			// Validate operator if specified
			if test.shouldValidateOp {
				if unary.Operator.Type != test.expectedOpType {
					t.Errorf("Expected operator %v, got %v", test.expectedOpType, unary.Operator.Type)
				}
			}
		})
	}
}

// Test power operations
func TestPower(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectBinary   bool
		expectedOpType lexer.TokenType
	}{
		{"simple power", "x ** y", true, lexer.StarStar},
		{"chained power", "x ** y ** z", true, lexer.StarStar},
		{"negative base", "-x ** 2", false, lexer.Minus}, // ** has higher precedence than unary -, so this is -(x ** 2)
		{"parenthesized base", "(-x) ** 2", true, lexer.StarStar},
		{"power of power", "(x ** 2) ** 3", true, lexer.StarStar},
		{"with numbers", "2 ** 8", true, lexer.StarStar},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.factor()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Validate expression type and operator
			if test.expectBinary {
				binary, ok := expr.(*ast.Binary)
				if !ok {
					t.Fatalf("Expected Binary, got %T", expr)
				}
				if binary.Operator.Type != test.expectedOpType {
					t.Errorf("Expected operator %v, got %v", test.expectedOpType, binary.Operator.Type)
				}
			} else {
				// For non-binary expressions, check if it's unary
				unary, ok := expr.(*ast.Unary)
				if !ok {
					t.Fatalf("Expected Unary, got %T", expr)
				}
				if unary.Operator.Type != test.expectedOpType {
					t.Errorf("Expected operator %v, got %v", test.expectedOpType, unary.Operator.Type)
				}
			}
		})
	}
}

// Test operator precedence
func TestArithmeticPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTopOp lexer.TokenType // The outermost/lowest precedence operator
		description   string
	}{
		{"multiply before add", "x + y * z", lexer.Plus, "multiplication binds tighter"},
		{"power before multiply", "x * y ** z", lexer.Star, "power binds tighter"},
		{"unary before binary", "-x * y", lexer.Star, "unary minus before multiply"},
		{"parentheses override", "(x + y) * z", lexer.Star, "parentheses change order"},
		{"complex expression", "x + y * z ** 2 - w / v", lexer.Minus, "subtraction is lowest precedence"},
		{"floor div precedence", "x + y // z - w", lexer.Minus, "floor division same as division"},
		{"modulo precedence", "x + y % z * w", lexer.Plus, "addition is lowest precedence"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			// Parse through sum() which handles the full expression
			expr, err := parser.sum()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression")
				return
			}

			// Validate that the top-level operator matches expected precedence
			if binary, ok := expr.(*ast.Binary); ok {
				if binary.Operator.Type != test.expectedTopOp {
					t.Errorf("Expected top-level operator %v, got %v - %s",
						test.expectedTopOp, binary.Operator.Type, test.description)
				}
			} else {
				t.Errorf("Expected binary expression at top level for precedence test")
			}
		})
	}
}

// Test error cases
func TestArithmeticErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"incomplete addition", "x +"},
		{"incomplete subtraction", "x -"},
		{"incomplete multiplication", "x *"},
		{"incomplete division", "x /"},
		{"invalid double star", "x * * y"}, // Two stars without space is not ** operator
		{"missing operand", "* y"},
		{"invalid power", "x ** ** y"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			_, err := parser.sum()
			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.input)
			}
		})
	}
}
