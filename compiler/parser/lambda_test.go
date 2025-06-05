package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a lambda expression
func parseLambdaExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to parse a statement containing a lambda
func parseLambdaStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.statement()
}

// Helper function to validate lambda structure
func validateLambda(t *testing.T, expr ast.Expr, expectedParams int, bodyType string) *ast.Lambda {
	lambda, ok := expr.(*ast.Lambda)
	if !ok {
		t.Fatalf("Expected *ast.Lambda, got %T", expr)
	}

	if lambda.Body == nil {
		t.Fatal("Lambda body should not be nil")
	}

	// Check parameter count
	var actualParams int
	if lambda.Parameters != nil {
		actualParams = len(lambda.Parameters.Parameters)
	}

	if expectedParams >= 0 && actualParams != expectedParams {
		t.Errorf("Expected %d parameters, got %d", expectedParams, actualParams)
	}

	// Check body type if specified
	if bodyType != "" {
		actualBodyType := getLambdaExpressionTypeName(lambda.Body)
		if actualBodyType != bodyType {
			t.Errorf("Expected body type %s, got %s", bodyType, actualBodyType)
		}
	}

	return lambda
}

// Helper function to get expression type name for lambda tests
func getLambdaExpressionTypeName(expr ast.Expr) string {
	if expr == nil {
		return "nil"
	}
	
	switch expr.(type) {
	case *ast.Literal:
		return "*ast.Literal"
	case *ast.Name:
		return "*ast.Name"
	case *ast.Binary:
		return "*ast.Binary"
	case *ast.Call:
		return "*ast.Call"
	case *ast.TernaryExpr:
		return "*ast.TernaryExpr"
	case *ast.Lambda:
		return "*ast.Lambda"
	case *ast.ListExpr:
		return "*ast.ListExpr"
	case *ast.DictExpr:
		return "*ast.DictExpr"
	case *ast.Unary:
		return "*ast.Unary"
	case *ast.TupleExpr:
		return "*ast.TupleExpr"
	default:
		return "unknown"
	}
}

// Test comprehensive lambda expression functionality
func TestLambdaExpressions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		expectedParams int
		bodyType     string
		description  string
	}{
		// Basic lambda expressions
		{
			name:         "simple lambda",
			input:        "lambda x: x + 1",
			expectedParams: 1,
			bodyType:     "*ast.Binary",
			description:  "basic lambda with single parameter and binary operation body",
		},
		{
			name:         "lambda with no parameters",
			input:        "lambda: 42",
			expectedParams: 0,
			bodyType:     "*ast.Literal",
			description:  "lambda with no parameters and literal body",
		},
		{
			name:         "lambda with multiple parameters",
			input:        "lambda x, y: x + y",
			expectedParams: 2,
			bodyType:     "*ast.Binary",
			description:  "lambda with multiple positional parameters",
		},
		{
			name:         "lambda with name reference",
			input:        "lambda x: x",
			expectedParams: 1,
			bodyType:     "*ast.Name",
			description:  "lambda returning parameter as-is",
		},

		// Error cases
		{
			name:        "lambda without colon",
			input:       "lambda x x + 1",
			hasError:    true,
			description: "lambda missing colon separator",
		},
		{
			name:        "lambda without body",
			input:       "lambda x:",
			hasError:    true,
			description: "lambda missing body expression",
		},
		{
			name:        "empty lambda",
			input:       "lambda",
			hasError:    true,
			description: "incomplete lambda expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseLambdaExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateLambda(t, expr, test.expectedParams, test.bodyType)
		})
	}
}

// Test lambda expressions in function calls and assignments
func TestLambdaInContext(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		description string
	}{
		{
			name:        "lambda assigned to variable",
			input:       "f = lambda x: x * 2",
			description: "lambda in assignment statement",
		},
		{
			name:        "lambda as function argument",
			input:       "result = map(lambda x: x ** 2, numbers)",
			description: "lambda as function call argument",
		},
		{
			name:        "lambda in function call with keyword",
			input:       "sorted(items, key=lambda x: x.name)",
			description: "lambda as keyword argument in function call",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Try parsing as statement first
			stmt, err := parseLambdaStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if stmt == nil {
				t.Errorf("Expected valid statement but got nil for %s", test.description)
			}
		})
	}
}

// Test lambda body expressions
func TestLambdaBody(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		bodyType    string
		hasError    bool
		description string
	}{
		{
			name:        "literal body",
			input:       "lambda: 42",
			bodyType:    "*ast.Literal",
			description: "lambda with literal number body",
		},
		{
			name:        "name body",
			input:       "lambda x: x",
			bodyType:    "*ast.Name",
			description: "lambda returning parameter",
		},
		{
			name:        "binary operation body",
			input:       "lambda x, y: x + y",
			bodyType:    "*ast.Binary",
			description: "lambda with arithmetic operation body",
		},
		{
			name:        "call expression body",
			input:       "lambda x: len(x)",
			bodyType:    "*ast.Call",
			description: "lambda with function call body",
		},
		{
			name:        "conditional expression body",
			input:       "lambda x: x if x > 0 else -x",
			bodyType:    "*ast.TernaryExpr",
			description: "lambda with ternary expression body",
		},
		{
			name:        "nested lambda body",
			input:       "lambda x: lambda y: x + y",
			bodyType:    "*ast.Lambda",
			description: "lambda returning another lambda (closure)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseLambdaExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateLambda(t, expr, -1, test.bodyType)
		})
	}
}

// Test lambda parameter patterns
func TestLambdaParameters(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedParams int
		hasError       bool
		description    string
	}{
		{
			name:           "no parameters",
			input:          "lambda: 42",
			expectedParams: 0,
			description:    "lambda with no parameters",
		},
		{
			name:           "one parameter",
			input:          "lambda x: x",
			expectedParams: 1,
			description:    "lambda with single positional parameter",
		},
		{
			name:           "two parameters",
			input:          "lambda x, y: x + y",
			expectedParams: 2,
			description:    "lambda with multiple positional parameters",
		},
		{
			name:           "parameter with default",
			input:          "lambda x, y=10: x + y",
			expectedParams: 2,
			description:    "lambda with default parameter value",
		},
		{
			name:           "varargs parameter",
			input:          "lambda *args: sum(args)",
			expectedParams: 1,
			description:    "lambda with variable arguments",
		},
		{
			name:           "kwargs parameter",
			input:          "lambda **kwargs: len(kwargs)",
			expectedParams: 1,
			description:    "lambda with keyword arguments",
		},
		{
			name:        "invalid parameter name - number",
			input:       "lambda 123: x",
			hasError:    true,
			description: "lambda with numeric parameter name",
		},
		{
			name:        "invalid parameter syntax",
			input:       "lambda def: def + 1",
			hasError:    true,
			description: "lambda using reserved keyword as parameter",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseLambdaExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateLambda(t, expr, test.expectedParams, "")
		})
	}
}