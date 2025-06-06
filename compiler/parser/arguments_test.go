package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
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
		name                string
		input               string
		expectedArgs        int
		expectedStars       int
		expectedKwargs      int
		expectedDoubleStars int
		hasError            bool
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
			name:           "single keyword argument",
			input:          "func(x=42)",
			expectedArgs:   1,
			expectedKwargs: 1,
		},
		{
			name:           "multiple keyword arguments",
			input:          "func(x=1, y=2)",
			expectedArgs:   2,
			expectedKwargs: 2,
		},
		{
			name:           "mixed positional and keyword",
			input:          "func(1, 2, x=3, y=4)",
			expectedArgs:   4,
			expectedKwargs: 2,
		},
		{
			name:          "starred argument",
			input:         "func(*args)",
			expectedArgs:  1,
			expectedStars: 1,
		},
		{
			name:                "double starred argument",
			input:               "func(**kwargs)",
			expectedArgs:        1,
			expectedDoubleStars: 1,
		},
		{
			name:                "complex argument mix",
			input:               "func(1, *args, x=2, **kwargs)",
			expectedArgs:        4,
			expectedStars:       1,
			expectedKwargs:      1,
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
			name:         "missing argument after comma",
			input:        "func(1,)",
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
		name                string
		input               string
		argIndex            int
		expectStar          bool
		expectDoubleStar    bool
		expectKeyword       bool
		expectedKeywordName string
		hasError            bool
	}{
		{
			name:     "positional argument",
			input:    "func(42)",
			argIndex: 0,
		},
		{
			name:       "starred argument",
			input:      "func(*items)",
			argIndex:   0,
			expectStar: true,
		},
		{
			name:             "double starred argument",
			input:            "func(**mapping)",
			argIndex:         0,
			expectDoubleStar: true,
		},
		{
			name:                "keyword argument",
			input:               "func(x=42)",
			argIndex:            0,
			expectKeyword:       true,
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
		name              string
		input             string
		argIndex          int
		expectedValueType string // "literal", "name", "binary", "call", etc.
		hasError          bool
	}{
		{
			name:              "number literal",
			input:             "func(42)",
			argIndex:          0,
			expectedValueType: "literal",
		},
		{
			name:              "string literal",
			input:             `func("hello")`,
			argIndex:          0,
			expectedValueType: "literal",
		},
		{
			name:              "variable name",
			input:             "func(x)",
			argIndex:          0,
			expectedValueType: "name",
		},
		{
			name:              "binary expression",
			input:             "func(a + b)",
			argIndex:          0,
			expectedValueType: "binary",
		},
		{
			name:              "function call",
			input:             "func(other())",
			argIndex:          0,
			expectedValueType: "call",
		},
		{
			name:              "attribute access",
			input:             "func(obj.method)",
			argIndex:          0,
			expectedValueType: "attribute",
		},
		{
			name:              "subscript access",
			input:             "func(arr[0])",
			argIndex:          0,
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

// TestGeneratorExpressionArguments tests generator expressions as function arguments
func TestGeneratorExpressionArguments(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		hasError        bool
		description     string
		expectGenExprAt int  // -1 if not checking, otherwise the index of expected GenExpr
		expectWrapped   bool // whether the GenExpr is wrapped in GroupExpr
	}{
		{
			name:            "simple generator expression",
			input:           "any(x > 0 for x in values)",
			hasError:        false,
			description:     "bare generator expression as sole argument",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "generator with conditional",
			input:           "sum(x**2 for x in range(10) if x % 2 == 0)",
			hasError:        false,
			description:     "generator expression with if clause",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "generator with multiple clauses",
			input:           "all(x < y for x in list1 for y in list2)",
			hasError:        false,
			description:     "generator expression with multiple for clauses",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "generator with keyword argument",
			input:           "sorted(x for x in items, reverse=True)",
			hasError:        false,
			description:     "generator expression followed by keyword argument",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "parenthesized generator expression",
			input:           "func((x for x in items))",
			hasError:        false,
			description:     "explicitly parenthesized generator expression",
			expectGenExprAt: 0,
			expectWrapped:   false, // Parser handles parentheses as part of generator syntax
		},
		{
			name:            "generator in max function",
			input:           "max(len(s) for s in strings)",
			hasError:        false,
			description:     "generator expression with function call in element",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "generator in join",
			input:           "''.join(c for c in text if c.isalpha())",
			hasError:        false,
			description:     "generator expression in method call",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "nested generator expressions",
			input:           "sum(sum(x for x in row) for row in matrix)",
			hasError:        false,
			description:     "generator expression within generator expression",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		{
			name:            "async generator expression",
			input:           "func(x async for x in async_iter)",
			hasError:        false,
			description:     "async generator expression",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		// These should work but might have edge cases
		{
			name:            "generator with starred arg after",
			input:           "func(x for x in items, *args)",
			hasError:        false,
			description:     "generator expression followed by starred argument",
			expectGenExprAt: 0,
			expectWrapped:   false,
		},
		// Note: Python actually allows these cases - the restriction is syntactic not semantic
		// A bare generator can appear anywhere a regular expression can appear
		{
			name:            "generator not first positional",
			input:           "func(1, (x for x in items))",
			hasError:        false,
			description:     "parenthesized generator expression as second argument",
			expectGenExprAt: 1,
			expectWrapped:   false, // Parser handles parentheses as part of generator syntax
		},
		{
			name:            "multiple parenthesized generators",
			input:           "func((x for x in items), (y for y in other))",
			hasError:        false,
			description:     "multiple parenthesized generator expressions",
			expectGenExprAt: 0,
			expectWrapped:   false, // Parser handles parentheses as part of generator syntax
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
					t.Errorf("Expected error for %s, but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.description, err)
				return
			}

			// Verify it's a call expression
			call, ok := expr.(*ast.Call)
			if !ok {
				t.Errorf("Expected *ast.Call but got %T", expr)
				return
			}

			// Verify generator expression if expected
			if tt.expectGenExprAt >= 0 {
				if tt.expectGenExprAt >= len(call.Arguments) {
					t.Errorf("Expected GenExpr at argument index %d, but only have %d arguments",
						tt.expectGenExprAt, len(call.Arguments))
					return
				}

				arg := call.Arguments[tt.expectGenExprAt]

				if tt.expectWrapped {
					// Expect GenExpr wrapped in GroupExpr
					group, ok := arg.Value.(*ast.GroupExpr)
					if !ok {
						t.Errorf("Expected argument at index %d to be GroupExpr, but got %T",
							tt.expectGenExprAt, arg.Value)
						return
					}
					if _, ok := group.Expression.(*ast.GenExpr); !ok {
						t.Errorf("Expected GroupExpr to contain GenExpr, but got %T",
							group.Expression)
					}
				} else {
					// Expect bare GenExpr
					if _, ok := arg.Value.(*ast.GenExpr); !ok {
						t.Errorf("Expected argument at index %d to be GenExpr, but got %T",
							tt.expectGenExprAt, arg.Value)
					}
				}
			}
		})
	}
}
