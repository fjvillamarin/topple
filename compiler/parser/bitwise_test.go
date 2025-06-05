package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

func TestBitwiseOperations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasError       bool
		expectBinary   bool
		expectedOpType lexer.TokenType
	}{
		// Bitwise OR operations
		{"simple bitwise or", "a | b", false, true, lexer.Pipe},
		{"chained bitwise or", "a | b | c", false, true, lexer.Pipe},
		{"bitwise or with numbers", "5 | 3", false, true, lexer.Pipe},
		{"bitwise or with expressions", "(a + b) | (c - d)", false, true, lexer.Pipe},
		{"mixed precedence or/and", "a & b | c", false, true, lexer.Pipe},
		{"bitwise or with variables", "flags | new_flag", false, true, lexer.Pipe},

		// Bitwise XOR operations  
		{"simple bitwise xor", "a ^ b", false, true, lexer.Caret},
		{"chained bitwise xor", "a ^ b ^ c", false, true, lexer.Caret},
		{"bitwise xor with numbers", "5 ^ 3", false, true, lexer.Caret},
		{"bitwise xor with expressions", "(a | b) ^ (c & d)", false, true, lexer.Caret},
		{"bitwise xor with variables", "hash1 ^ hash2", false, true, lexer.Caret},

		// Bitwise AND operations
		{"simple bitwise and", "a & b", false, true, lexer.Ampersand},
		{"chained bitwise and", "a & b & c", false, true, lexer.Ampersand},
		{"bitwise and with numbers", "7 & 3", false, true, lexer.Ampersand},
		{"bitwise and with expressions", "(a << 1) & mask", false, true, lexer.Ampersand},
		{"bitwise and with variables", "permissions & required", false, true, lexer.Ampersand},

		// Shift operations
		{"left shift", "a << 2", false, true, lexer.LessLess},
		{"right shift", "a >> 2", false, true, lexer.GreaterGreater},
		{"chained left shift", "a << 2 << 1", false, true, lexer.LessLess},
		{"chained right shift", "a >> 1 >> 2", false, true, lexer.GreaterGreater},
		{"shift with expressions", "(a + b) << count", false, true, lexer.LessLess},
		{"shift with variables", "value >> shift_amount", false, true, lexer.GreaterGreater},

		// Error cases  
		{"incomplete bitwise or", "a |", true, false, lexer.EOF},
		{"missing operand or", "| b", true, false, lexer.EOF},
		{"incomplete bitwise xor", "a ^", true, false, lexer.EOF},
		{"missing operand xor", "^ b", true, false, lexer.EOF},
		{"incomplete bitwise and", "a &", true, false, lexer.EOF},
		{"missing operand and", "& b", true, false, lexer.EOF},
		{"incomplete left shift", "a <<", true, false, lexer.EOF},
		{"missing operand left shift", "<< b", true, false, lexer.EOF},
		{"incomplete right shift", "a >>", true, false, lexer.EOF},
		{"missing operand right shift", ">> b", true, false, lexer.EOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if expr == nil {
				t.Errorf("Expression should not be nil")
				return
			}

			// Validate binary expression if expected
			if tt.expectBinary {
				binary, ok := expr.(*ast.Binary)
				if !ok {
					t.Errorf("Expected *ast.Binary but got %T", expr)
					return
				}
				if binary.Operator.Type != tt.expectedOpType {
					t.Errorf("Expected operator %v but got %v", tt.expectedOpType, binary.Operator.Type)
				}
			}
		})
	}
}

func TestBitwisePrecedence(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTopOp  lexer.TokenType
		description    string
	}{
		{"or lower than xor", "a ^ b | c", lexer.Pipe, "| has lower precedence than ^"},
		{"xor lower than and", "a & b ^ c", lexer.Caret, "^ has lower precedence than &"},
		{"and lower than shift", "a << 1 & b", lexer.Ampersand, "& has lower precedence than <<"},
		{"shift lower than addition", "a + b << 2", lexer.LessLess, "<< has lower precedence than +"},
		{"complex precedence", "a + b << 1 & c ^ d | e", lexer.Pipe, "| is lowest precedence"},
		{"parentheses override", "(a | b) & c", lexer.Ampersand, "parentheses change precedence"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.input, err)
			}

			if expr == nil {
				t.Error("Expected non-nil expression")
				return
			}

			// Validate that the top-level operator matches expected precedence
			if binary, ok := expr.(*ast.Binary); ok {
				if binary.Operator.Type != tt.expectedTopOp {
					t.Errorf("Expected top-level operator %v, got %v - %s", 
						tt.expectedTopOp, binary.Operator.Type, tt.description)
				}
			} else {
				t.Errorf("Expected binary expression at top level for precedence test")
			}
		})
	}
}

func TestBitwiseWithComplexExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"bitwise with function calls", "hash_func(a) ^ hash_func(b)"},
		{"bitwise with array access", "data[i] | flags[j]"},
		{"bitwise with attribute access", "obj.flags & obj.mask"},
		{"bitwise with nested expressions", "(a + b) & (c << d)"},
		{"mixed bitwise operations", "a | b ^ c & d"},
		{"bitwise with unary", "~a & b"},
		{"bitwise with comparison", "(a > b) & (c < d)"}, // boolean context
		{"complex bit manipulation", "(value & mask) | (bits << shift)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if err != nil {
				t.Errorf("Failed to parse %s: %v", tt.input, err)
				return
			}

			if expr == nil {
				t.Errorf("Expression should not be nil")
			}

			// Just verify we got a valid expression - complex validation would
			// require detailed AST walking
		})
	}
}

func TestUnaryBitwiseOperations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasError       bool
		expectedOpType lexer.TokenType
	}{
		{"bitwise not simple", "~a", false, lexer.Tilde},
		{"bitwise not number", "~42", false, lexer.Tilde},
		{"bitwise not expression", "~(a + b)", false, lexer.Tilde},
		{"double bitwise not", "~~a", false, lexer.Tilde}, // outermost operator
		{"bitwise not with other unary", "-~a", false, lexer.Minus}, // outermost operator
		{"bitwise not missing operand", "~", true, lexer.EOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if expr == nil {
				t.Errorf("Expression should not be nil")
				return
			}

			// Check for unary expression
			unary, ok := expr.(*ast.Unary)
			if !ok {
				t.Errorf("Expected *ast.Unary but got %T", expr)
				return
			}

			if unary.Operator.Type != tt.expectedOpType {
				t.Errorf("Expected operator %v but got %v", tt.expectedOpType, unary.Operator.Type)
			}
		})
	}
}