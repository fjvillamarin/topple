package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"testing"
)

// Helper function to classify argument type
func getArgumentType(arg *ast.Argument) string {
	switch {
	case arg.IsDoubleStar:
		return "doublestar"
	case arg.IsStar:
		return "star"
	case arg.Name != nil:
		return "kw"
	default:
		return "pos"
	}
}

func TestFunctionCallArguments(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedArgs   int
		expectedStars  int
		expectedKwargs int
		expectedDoubleStars int
		hasError       bool
	}{
		{
			name:         "no arguments",
			input:        "func()",
			expectedArgs: 0,
		},
		{
			name:         "single positional argument",
			input:        "func(42)",
			expectedArgs: 1,
		},
		{
			name:         "multiple positional arguments",
			input:        "func(1, 2, 3)",
			expectedArgs: 3,
		},
		{
			name:         "single keyword argument",
			input:        "func(x=42)",
			expectedArgs: 1,
			expectedKwargs: 1,
		},
		{
			name:         "multiple keyword arguments",
			input:        "func(x=1, y=2)",
			expectedArgs: 2,
			expectedKwargs: 2,
		},
		{
			name:         "mixed positional and keyword",
			input:        "func(1, 2, x=3, y=4)",
			expectedArgs: 4,
			expectedKwargs: 2,
		},
		{
			name:         "starred argument",
			input:        "func(*args)",
			expectedArgs: 1,
			expectedStars: 1,
		},
		{
			name:         "double starred argument",
			input:        "func(**kwargs)",
			expectedArgs: 1,
			expectedDoubleStars: 1,
		},
		{
			name:         "complex argument mix",
			input:        "func(1, *args, x=2, **kwargs)",
			expectedArgs: 4,
			expectedStars: 1,
			expectedKwargs: 1,
			expectedDoubleStars: 1,
		},
		{
			name:         "trailing comma allowed",
			input:        "func(1, 2, 3,)",
			expectedArgs: 3,
		},
		{
			name:     "keyword before positional",
			input:    "func(x=1, 2)",
			hasError: true,
		},
		{
			name:     "missing argument after comma",
			input:    "func(1,)",
			expectedArgs: 1,
		},
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

			call, ok := expr.(*ast.Call)
			if !ok {
				t.Errorf("Expected *ast.Call but got %T", expr)
				return
			}

			if len(call.Arguments) != tt.expectedArgs {
				t.Errorf("Expected %d arguments but got %d", tt.expectedArgs, len(call.Arguments))
			}

			// Count argument types
			actualStars := 0
			actualKwargs := 0
			actualDoubleStars := 0

			for _, arg := range call.Arguments {
				if arg.IsStar {
					actualStars++
				}
				if arg.IsDoubleStar {
					actualDoubleStars++
				}
				if arg.Name != nil {
					actualKwargs++
				}
			}

			if actualStars != tt.expectedStars {
				t.Errorf("Expected %d starred arguments but got %d", tt.expectedStars, actualStars)
			}

			if actualKwargs != tt.expectedKwargs {
				t.Errorf("Expected %d keyword arguments but got %d", tt.expectedKwargs, actualKwargs)
			}

			if actualDoubleStars != tt.expectedDoubleStars {
				t.Errorf("Expected %d double-starred arguments but got %d", tt.expectedDoubleStars, actualDoubleStars)
			}
		})
	}
}

func TestArgumentTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		argIndex int
		expectStar bool
		expectDoubleStar bool
		expectKeyword bool
		expectedKeywordName string
		hasError bool
	}{
		{
			name:     "positional argument",
			input:    "func(42)",
			argIndex: 0,
		},
		{
			name:     "starred argument",
			input:    "func(*items)",
			argIndex: 0,
			expectStar: true,
		},
		{
			name:     "double starred argument",
			input:    "func(**mapping)",
			argIndex: 0,
			expectDoubleStar: true,
		},
		{
			name:     "keyword argument",
			input:    "func(x=42)",
			argIndex: 0,
			expectKeyword: true,
			expectedKeywordName: "x",
		},
		{
			name:     "complex expression as argument",
			input:    "func(a + b * 2)",
			argIndex: 0,
		},
		{
			name:     "nested function call as argument",
			input:    "func(other(1, 2))",
			argIndex: 0,
		},
		{
			name:     "list comprehension as argument",
			input:    "func([x for x in items])",
			argIndex: 0,
		},
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

			call, ok := expr.(*ast.Call)
			if !ok {
				t.Errorf("Expected *ast.Call but got %T", expr)
				return
			}

			if tt.argIndex >= len(call.Arguments) {
				t.Errorf("Argument index %d out of range (have %d arguments)", tt.argIndex, len(call.Arguments))
				return
			}

			arg := call.Arguments[tt.argIndex]

			if arg.IsStar != tt.expectStar {
				t.Errorf("Expected IsStar=%v but got %v", tt.expectStar, arg.IsStar)
			}

			if arg.IsDoubleStar != tt.expectDoubleStar {
				t.Errorf("Expected IsDoubleStar=%v but got %v", tt.expectDoubleStar, arg.IsDoubleStar)
			}

			if tt.expectKeyword {
				if arg.Name == nil {
					t.Errorf("Expected keyword argument but got none")
				} else if arg.Name.Token.Lexeme != tt.expectedKeywordName {
					t.Errorf("Expected keyword name %q but got %q", tt.expectedKeywordName, arg.Name.Token.Lexeme)
				}
			} else {
				if arg.Name != nil {
					t.Errorf("Expected positional argument but got keyword")
				}
			}

			if arg.Value == nil {
				t.Errorf("Argument should have a value")
			}
		})
	}
}

func TestArgumentOrdering(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		positions []string // Expected argument types in order: "pos", "star", "kw", "doublestar"
		hasError  bool
	}{
		{
			name:      "correct ordering: positional then keyword",
			input:     "func(1, 2, x=3, y=4)",
			positions: []string{"pos", "pos", "kw", "kw"},
		},
		{
			name:      "correct ordering: positional, star, keyword, doublestar",
			input:     "func(1, *args, x=2, **kwargs)",
			positions: []string{"pos", "star", "kw", "doublestar"},
		},
		{
			name:      "multiple starred arguments",
			input:     "func(*args1, *args2)",
			positions: []string{"star", "star"},
		},
		{
			name:      "multiple double-starred arguments",
			input:     "func(**kwargs1, **kwargs2)",
			positions: []string{"doublestar", "doublestar"},
		},
		{
			name:      "all argument types",
			input:     "func(1, 2, *args, x=3, y=4, **kwargs)",
			positions: []string{"pos", "pos", "star", "kw", "kw", "doublestar"},
		},
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

			call, ok := expr.(*ast.Call)
			if !ok {
				t.Errorf("Expected *ast.Call but got %T", expr)
				return
			}

			if len(call.Arguments) != len(tt.positions) {
				t.Errorf("Expected %d arguments but got %d", len(tt.positions), len(call.Arguments))
				return
			}

			for i, expectedType := range tt.positions {
				arg := call.Arguments[i]
				var actualType string

				switch {
				case arg.IsDoubleStar:
					actualType = "doublestar"
				case arg.IsStar:
					actualType = "star"
				case arg.Name != nil:
					actualType = "kw"
				default:
					actualType = "pos"
				}

				if actualType != expectedType {
					t.Errorf("Argument %d: expected %s but got %s", i, expectedType, actualType)
				}
			}
		})
	}
}

func TestArgumentExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		argIndex int
		expectedValueType string // "literal", "name", "binary", "call", etc.
		hasError bool
	}{
		{
			name:     "number literal",
			input:    "func(42)",
			argIndex: 0,
			expectedValueType: "literal",
		},
		{
			name:     "string literal",
			input:    `func("hello")`,
			argIndex: 0,
			expectedValueType: "literal",
		},
		{
			name:     "variable name",
			input:    "func(x)",
			argIndex: 0,
			expectedValueType: "name",
		},
		{
			name:     "binary expression",
			input:    "func(a + b)",
			argIndex: 0,
			expectedValueType: "binary",
		},
		{
			name:     "function call",
			input:    "func(other())",
			argIndex: 0,
			expectedValueType: "call",
		},
		{
			name:     "attribute access",
			input:    "func(obj.method)",
			argIndex: 0,
			expectedValueType: "attribute",
		},
		{
			name:     "subscript access",
			input:    "func(arr[0])",
			argIndex: 0,
			expectedValueType: "subscript",
		},
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

			call, ok := expr.(*ast.Call)
			if !ok {
				t.Errorf("Expected *ast.Call but got %T", expr)
				return
			}

			if tt.argIndex >= len(call.Arguments) {
				t.Errorf("Argument index %d out of range", tt.argIndex)
				return
			}

			arg := call.Arguments[tt.argIndex]
			if arg.Value == nil {
				t.Errorf("Argument should have a value")
				return
			}

			var actualType string
			switch arg.Value.(type) {
			case *ast.Literal:
				actualType = "literal"
			case *ast.Name:
				actualType = "name"
			case *ast.Binary:
				actualType = "binary"
			case *ast.Call:
				actualType = "call"
			case *ast.Attribute:
				actualType = "attribute"
			case *ast.Subscript:
				actualType = "subscript"
			default:
				actualType = "unknown"
			}

			if actualType != tt.expectedValueType {
				t.Errorf("Expected value type %s but got %s (actual type: %T)", 
					tt.expectedValueType, actualType, arg.Value)
			}
		})
	}
}

func TestArgumentErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed function call",
			input: "func(1, 2",
		},
		{
			name:  "invalid starred syntax",
			input: "func(*)",
		},
		{
			name:  "invalid double starred syntax",
			input: "func(**)",
		},
		{
			name:  "keyword without value",
			input: "func(x=)",
		},
		{
			name:  "invalid keyword name",
			input: "func(123=value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			_, err := parser.expression()

			if err == nil {
				t.Errorf("Expected error but got none for input: %s", tt.input)
			}
		})
	}
}