package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

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
