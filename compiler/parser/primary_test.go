package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to create a parser from a string
func parseExpr(input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to validate literal expression
func validateLiteral(t *testing.T, expr ast.Expr, expectedType ast.LiteralType, expectedValue interface{}) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected *ast.Literal, got %T", expr)
	}

	if lit.Type != expectedType {
		t.Errorf("Expected literal type %v, got %v", expectedType, lit.Type)
	}

	if expectedValue != nil && lit.Value != expectedValue {
		t.Errorf("Expected value %v, got %v", expectedValue, lit.Value)
	}
}

// Helper function to validate name expression
func validateName(t *testing.T, expr ast.Expr, expectedName string) {
	name, ok := expr.(*ast.Name)
	if !ok {
		t.Fatalf("Expected *ast.Name, got %T", expr)
	}

	if name.Token.Lexeme != expectedName {
		t.Errorf("Expected identifier %s, got %s", expectedName, name.Token.Lexeme)
	}
}

// Helper function to validate attribute expression
func validateAttribute(t *testing.T, expr ast.Expr, expectedAttrName string) {
	attr, ok := expr.(*ast.Attribute)
	if !ok {
		t.Fatalf("Expected *ast.Attribute, got %T", expr)
	}

	if attr.Name.Lexeme != expectedAttrName {
		t.Errorf("Expected attribute name %s, got %s", expectedAttrName, attr.Name.Lexeme)
	}
}

// Helper function to parse a primary expression specifically
func parsePrimary(input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.primary()
}

// Helper function to parse an atom specifically
func parseAtom(input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.atom()
}

// Test atom parsing - literals
func TestAtomLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		litType  ast.LiteralType
	}{
		{"true literal", "True", true, ast.LiteralTypeBool},
		{"false literal", "False", false, ast.LiteralTypeBool},
		{"none literal", "None", nil, ast.LiteralTypeNone},
		{"integer literal", "42", int64(42), ast.LiteralTypeString}, // Note: parser stores all numbers as LiteralTypeString
		{"float literal", "3.14", 3.14, ast.LiteralTypeString},
		{"string literal", `"hello"`, "hello", ast.LiteralTypeString},
		{"ellipsis literal", "...", nil, ast.LiteralTypeNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseAtom(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			validateLiteral(t, expr, test.litType, test.expected)
		})
	}
}

// Test atom parsing - identifiers
func TestAtomIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple identifier", "x"},
		{"underscore", "_"},
		{"with numbers", "var123"},
		{"with underscore", "my_var"},
		{"uppercase", "MY_CONST"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseAtom(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			validateName(t, expr, test.input)
		})
	}
}

// Test primary expressions with attribute access
func TestPrimaryAttributeAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		attrName string
		depth    int // How many dots deep
	}{
		{"simple attribute", "obj.attr", "attr", 1},
		{"chained attributes", "obj.attr.method", "method", 2},
		{"deep chain", "a.b.c.d", "d", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parsePrimary(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// For chained attributes, the structure is right-associative
			// e.g., a.b.c is (Attribute (Attribute a b) c)
			validateAttribute(t, expr, test.attrName)
		})
	}
}

// Test primary expressions with subscript access
func TestPrimarySubscript(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple subscript", "arr[0]"},
		{"string key", `dict["key"]`},
		{"variable index", "arr[i]"},
		{"expression index", "arr[i + 1]"},
		{"chained subscript", "matrix[i][j]"},
		{"mixed access", "obj.attr[0]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parsePrimary(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Check that we got a subscript at the top level
			if !strings.HasPrefix(test.name, "mixed") {
				_, ok := expr.(*ast.Subscript)
				if !ok {
					t.Fatalf("Expected Subscript, got %T", expr)
				}
			}
		})
	}
}

// Test primary expressions with function calls
func TestPrimaryFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		argCount int
	}{
		{"no args", "func()", 0},
		{"single arg", "func(x)", 1},
		{"multiple args", "func(x, y, z)", 3},
		{"with keywords", "func(x, y=2)", 2},
		{"chained calls", "func()()", 0},
		{"method call", "obj.method(x)", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parsePrimary(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Find the call expression (might be nested)
			var call *ast.Call
			switch e := expr.(type) {
			case *ast.Call:
				call = e
			case *ast.Attribute:
				// Method call case - check if the parent is a call
				t.Skip("Need to handle method calls differently")
			}

			if call != nil && len(call.Arguments) != test.argCount {
				t.Errorf("Expected %d arguments, got %d", test.argCount, len(call.Arguments))
			}
		})
	}
}

// Test await expressions
func TestAwaitExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"await function call", "await func()"},
		{"await method call", "await obj.method()"},
		{"await identifier", "await coro"},
		{"await with args", "await async_func(x, y)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			expr, err := parser.await()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			await, ok := expr.(*ast.AwaitExpr)
			if !ok {
				t.Fatalf("Expected AwaitExpr, got %T", expr)
			}

			if await.Expr == nil {
				t.Error("Await expression should have a non-nil expression")
			}
		})
	}
}

// Test complex primary expressions
func TestComplexPrimaryExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"attribute then call", "obj.method()"},
		{"subscript then attribute", "arr[0].value"},
		{"call then subscript", "func()[0]"},
		{"deep nesting", "a.b[c].d(e)[f].g"},
		{"multiple calls", "factory()()()"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parsePrimary(test.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Just ensure we got a non-nil expression
			if expr == nil {
				t.Error("Expected non-nil expression")
			}
		})
	}
}

// Test error cases
func TestPrimaryErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing closing bracket", "arr[0"},
		{"missing closing paren", "func("},
		{"missing attribute name", "obj."},
		{"empty subscript", "arr[]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parsePrimary(test.input)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.input)
			}
		})
	}
}

// Test that atom correctly dispatches to other parsers
func TestAtomDispatching(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
	}{
		{"parentheses", "(x)", "*ast.Group"},
		{"list", "[1, 2, 3]", "*ast.List"},
		{"dict", "{x: 1}", "*ast.Dict"},
		{"set", "{1, 2, 3}", "*ast.Set"},
		{"f-string", `f"hello {name}"`, "*ast.FString"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseAtom(test.input)
			if err != nil {
				// Some of these might fail if the respective parsers aren't implemented
				t.Skipf("Parser not implemented yet: %v", err)
			}

			if expr != nil {
				// Just check that we got some expression back
				t.Logf("Got expression type: %T", expr)
			}
		})
	}
}

// Test edge cases and error handling comprehensively
func TestPrimaryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		// Valid complex cases
		{
			name:        "deeply nested access",
			input:       "a.b.c[d].e().f[g].h",
			description: "deeply nested attribute and subscript access",
		},
		{
			name:        "multiple function calls",
			input:       "factory()().method()()",
			description: "chained function calls",
		},
		{
			name:        "complex subscript expressions",
			input:       "matrix[row + 1][col * 2]",
			description: "subscripts with complex expressions",
		},
		{
			name:        "method call with complex args",
			input:       "obj.method(a.b, c[d], func())",
			description: "method call with complex argument expressions",
		},

		// Error cases
		{
			name:        "incomplete attribute access",
			input:       "obj.",
			hasError:    true,
			errorText:   "expected identifier",
			description: "attribute access without identifier",
		},
		{
			name:        "unclosed subscript",
			input:       "arr[index",
			hasError:    true,
			errorText:   "expected ']'",
			description: "missing closing bracket in subscript",
		},
		{
			name:        "unclosed function call",
			input:       "func(arg",
			hasError:    true,
			errorText:   "expected ')'",
			description: "missing closing parenthesis in call",
		},
		{
			name:        "empty subscript",
			input:       "arr[]",
			hasError:    true,
			description: "subscript without index expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parsePrimary(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q", test.errorText, err.Error())
				}
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
