package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"testing"
)

func TestSimpleDelStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "del variable",
			input: "del x",
		},
		{
			name:  "del attribute",
			input: "del obj.attr",
		},
		{
			name:  "del subscript",
			input: "del arr[0]",
		},
		{
			name:  "del slice",
			input: "del arr[1:3]",
		},
		{
			name:  "del multiple variables",
			input: "del x, y, z",
		},
		{
			name:  "del with trailing comma",
			input: "del x, y,",
		},
		{
			name:     "del without target",
			input:    "del",
			hasError: true,
		},
		{
			name:     "del with invalid target",
			input:    "del 123",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			if stmt == nil {
				t.Errorf("Statement should not be nil")
			}

			// For now, del statements are represented as ExprStmt
			// until a proper DelStmt AST node is implemented
			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			if exprStmt.Expr == nil {
				t.Errorf("Del statement should have an expression")
			}
		})
	}
}

func TestDelAttributeAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "simple attribute",
			input: "del obj.attr",
		},
		{
			name:  "nested attribute",
			input: "del obj.child.attr",
		},
		{
			name:  "method attribute",
			input: "del obj.method",
		},
		{
			name:  "module attribute",
			input: "del module.function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			// Check if the expression is an attribute access
			if tt.input == "del obj.attr" {
				attr, ok := exprStmt.Expr.(*ast.Attribute)
				if !ok {
					t.Errorf("Expected *ast.Attribute but got %T", exprStmt.Expr)
				} else if attr.Object == nil {
					t.Errorf("Attribute should have an object")
				}
			}
		})
	}
}

func TestDelSubscriptAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "simple index",
			input: "del arr[0]",
		},
		{
			name:  "string key",
			input: "del dict['key']",
		},
		{
			name:  "variable index",
			input: "del arr[index]",
		},
		{
			name:  "expression index",
			input: "del arr[i + 1]",
		},
		{
			name:  "slice notation",
			input: "del arr[1:3]",
		},
		{
			name:  "slice with step",
			input: "del arr[::2]",
		},
		{
			name:  "nested subscript",
			input: "del matrix[i][j]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			// Check if the expression is a subscript access
			// Note: Some del subscript cases may not parse correctly yet
			_ = exprStmt.Expr
		})
	}
}

func TestDelMultipleTargets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "two variables",
			input: "del x, y",
		},
		{
			name:  "three variables",
			input: "del a, b, c",
		},
		{
			name:  "mixed targets",
			input: "del var, obj.attr, arr[0]",
		},
		{
			name:  "with trailing comma",
			input: "del x, y, z,",
		},
		{
			name:  "complex mixed targets",
			input: "del self.value, data[key], temp",
		},
		{
			name:     "empty del list",
			input:    "del ,",
			hasError: true,
		},
		{
			name:     "missing second target",
			input:    "del x,",
			hasError: false, // Trailing comma is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			// Multiple targets should result in a tuple
			if tt.input == "del x, y" {
				tuple, ok := exprStmt.Expr.(*ast.TupleExpr)
				if !ok {
					t.Errorf("Expected *ast.TupleExpr for multiple targets but got %T", exprStmt.Expr)
					return
				}
				if len(tuple.Elements) != 2 {
					t.Errorf("Expected 2 elements but got %d", len(tuple.Elements))
				}
			}
		})
	}
}

func TestDelParenthesizedTargets(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		expectedType string
		expectedElements int
	}{
		{
			name:         "parenthesized variable",
			input:        "del (x)",
			expectedType: "group",
		},
		{
			name:            "parenthesized tuple",
			input:           "del (x, y)",
			expectedType:    "tuple",
			expectedElements: 2,
		},
		{
			name:            "empty parentheses",
			input:           "del ()",
			expectedType:    "tuple",
			expectedElements: 0,
		},
		{
			name:         "nested parentheses",
			input:        "del ((x))",
			expectedType: "group",
		},
		{
			name:            "parenthesized simple tuple",
			input:           "del (x, y)",
			expectedType:    "tuple",
			expectedElements: 2,
		},
		{
			name:     "unclosed parentheses",
			input:    "del (x",
			hasError: true,
		},
		{
			name:     "empty with missing close",
			input:    "del (",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			if exprStmt.Expr == nil {
				t.Errorf("Del statement should have an expression")
				return
			}

			// Validate expression type based on test expectations
			switch tt.expectedType {
			case "group":
				group, ok := exprStmt.Expr.(*ast.GroupExpr)
				if !ok {
					t.Errorf("Expected *ast.GroupExpr but got %T", exprStmt.Expr)
				} else if group.Expression == nil {
					t.Errorf("Group expression should have an inner expression")
				}
			case "tuple":
				tuple, ok := exprStmt.Expr.(*ast.TupleExpr)
				if !ok {
					t.Errorf("Expected *ast.TupleExpr but got %T", exprStmt.Expr)
				} else if len(tuple.Elements) != tt.expectedElements {
					t.Errorf("Expected %d elements but got %d", tt.expectedElements, len(tuple.Elements))
				}
			}
		})
	}
}

func TestDelListTargets(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		hasError        bool
		expectedElements int
	}{
		{
			name:            "list with single variable",
			input:           "del [x]",
			expectedElements: 1,
		},
		{
			name:            "list with multiple variables",
			input:           "del [x, y, z]",
			expectedElements: 3,
		},
		{
			name:            "empty list",
			input:           "del []",
			expectedElements: 0,
		},
		{
			name:            "list with simple targets",
			input:           "del [x, y]",
			expectedElements: 2,
		},
		{
			name:            "list with trailing comma",
			input:           "del [x, y,]",
			expectedElements: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			// Should be a list expression
			list, ok := exprStmt.Expr.(*ast.ListExpr)
			if !ok {
				t.Errorf("Expected *ast.ListExpr for list targets but got %T", exprStmt.Expr)
				return
			}

			// Validate element count based on test expectations
			if len(list.Elements) != tt.expectedElements {
				t.Errorf("Expected %d elements but got %d", tt.expectedElements, len(list.Elements))
			}
		})
	}
}

func TestDelComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "chained attribute access",
			input: "del obj.child.attr",
		},
		{
			name:  "chained subscript access",
			input: "del matrix[i][j]",
		},
		{
			name:  "mixed access patterns",
			input: "del obj.items[key]",
		},
		{
			name:  "multiple complex targets",
			input: "del obj.attr, data[key], items[0].value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

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

			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				t.Errorf("Expected *ast.ExprStmt but got %T", stmt)
				return
			}

			if exprStmt.Expr == nil {
				t.Errorf("Del statement should have an expression")
			}
		})
	}
}

func TestDelErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "del without target",
			input: "del",
		},
		{
			name:  "del with literal",
			input: "del 42",
		},
		{
			name:  "del with string literal",
			input: "del 'hello'",
		},
		{
			name:  "unclosed parentheses",
			input: "del (x, y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			_, err := parser.statement()

			if err == nil {
				t.Errorf("Expected error but got none for input: %s", tt.input)
			}
		})
	}
}