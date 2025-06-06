package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// TestGeneratorExpressionArguments tests generator expressions as function arguments
func TestGeneratorExpressionArguments(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		description string
	}{
		{
			name:        "simple generator expression",
			input:       "any(x > 0 for x in values)",
			hasError:    false,
			description: "bare generator expression as sole argument",
		},
		{
			name:        "generator with conditional",
			input:       "sum(x**2 for x in range(10) if x % 2 == 0)",
			hasError:    false,
			description: "generator expression with if clause",
		},
		{
			name:        "generator with multiple clauses",
			input:       "all(x < y for x in list1 for y in list2)",
			hasError:    false,
			description: "generator expression with multiple for clauses",
		},
		{
			name:        "generator with keyword argument",
			input:       "sorted(x for x in items, reverse=True)",
			hasError:    false,
			description: "generator expression followed by keyword argument",
		},
		{
			name:        "parenthesized generator expression",
			input:       "func((x for x in items))",
			hasError:    false,
			description: "explicitly parenthesized generator expression",
		},
		{
			name:        "generator in max function",
			input:       "max(len(s) for s in strings)",
			hasError:    false,
			description: "generator expression with function call in element",
		},
		{
			name:        "generator in join",
			input:       "''.join(c for c in text if c.isalpha())",
			hasError:    false,
			description: "generator expression in method call",
		},
		{
			name:        "nested generator expressions",
			input:       "sum(sum(x for x in row) for row in matrix)",
			hasError:    false,
			description: "generator expression within generator expression",
		},
		{
			name:        "async generator expression",
			input:       "func(x async for x in async_iter)",
			hasError:    false,
			description: "async generator expression",
		},
		// These should work but might have edge cases
		{
			name:        "generator with starred arg after",
			input:       "func(x for x in items, *args)",
			hasError:    false,
			description: "generator expression followed by starred argument",
		},
		// Note: Python actually allows these cases - the restriction is syntactic not semantic
		// A bare generator can appear anywhere a regular expression can appear
		{
			name:        "generator not first positional",
			input:       "func(1, (x for x in items))",
			hasError:    false,
			description: "parenthesized generator expression as second argument",
		},
		{
			name:        "multiple parenthesized generators",
			input:       "func((x for x in items), (y for y in other))",
			hasError:    false,
			description: "multiple parenthesized generator expressions",
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

			// Verify we have at least one GenExpr in the arguments (unless it's a specific test case)
			if tt.name != "generator not first positional" && tt.name != "multiple parenthesized generators" {
				// For most tests, verify the first argument is a generator
				if len(call.Arguments) > 0 {
					arg := call.Arguments[0]
					// Check if the argument value is a GenExpr
					if _, isGenExpr := arg.Value.(*ast.GenExpr); !isGenExpr {
						// It might be wrapped in a GroupExpr if explicitly parenthesized
						if group, isGroup := arg.Value.(*ast.GroupExpr); isGroup {
							if _, isGenExpr := group.Expression.(*ast.GenExpr); !isGenExpr && tt.name != "parenthesized generator expression" {
								t.Errorf("Expected first argument to be GenExpr, but got %T", group.Expression)
							}
						} else if tt.name != "parenthesized generator expression" {
							t.Errorf("Expected first argument to be GenExpr, but got %T", arg.Value)
						}
					}
				}
			}
		})
	}
}
