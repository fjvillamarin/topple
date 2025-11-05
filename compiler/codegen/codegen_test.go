package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// TestCodeGeneration tests code generation using golden files
func TestCodeGeneration(t *testing.T) {
	updateGolden := os.Getenv("UPDATE_GOLDEN") == "1"

	testCases := []struct {
		category   string
		name       string
		buildAST   func() ast.Node
		skipReason string // If non-empty, test will be skipped with this reason
	}{
		// Literals
		{
			category: "literals",
			name:     "integer",
			buildAST: func() ast.Node {
				return &ast.Literal{Value: int64(42), Type: ast.LiteralTypeNumber}
			},
		},
		{
			category: "literals",
			name:     "float",
			buildAST: func() ast.Node {
				return &ast.Literal{Value: float64(3.14), Type: ast.LiteralTypeNumber}
			},
		},
		{
			category: "literals",
			name:     "string",
			buildAST: func() ast.Node {
				return &ast.Literal{Value: "hello world", Type: ast.LiteralTypeString}
			},
		},
		{
			category: "literals",
			name:     "boolean_true",
			buildAST: func() ast.Node {
				return &ast.Literal{Value: true, Type: ast.LiteralTypeBool}
			},
		},
		{
			category: "literals",
			name:     "boolean_false",
			buildAST: func() ast.Node {
				return &ast.Literal{Value: false, Type: ast.LiteralTypeBool}
			},
		},
		{
			category: "literals",
			name:     "none",
			buildAST: func() ast.Node {
				return &ast.Literal{Type: ast.LiteralTypeNone}
			},
		},

		// Collections
		{
			category: "collections",
			name:     "empty_list",
			buildAST: func() ast.Node {
				return &ast.ListExpr{Elements: []ast.Expr{}}
			},
		},
		{
			category: "collections",
			name:     "simple_list",
			buildAST: func() ast.Node {
				return &ast.ListExpr{
					Elements: []ast.Expr{
						&ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
						&ast.Literal{Value: int64(2), Type: ast.LiteralTypeNumber},
						&ast.Literal{Value: int64(3), Type: ast.LiteralTypeNumber},
					},
				}
			},
		},
		{
			category: "collections",
			name:     "empty_tuple",
			buildAST: func() ast.Node {
				return &ast.TupleExpr{Elements: []ast.Expr{}}
			},
		},
		{
			category: "collections",
			name:     "single_element_tuple",
			buildAST: func() ast.Node {
				return &ast.TupleExpr{
					Elements: []ast.Expr{
						&ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
					},
				}
			},
		},
		{
			category: "collections",
			name:     "empty_dict",
			buildAST: func() ast.Node {
				return &ast.DictExpr{Pairs: []ast.DictPair{}}
			},
		},
		{
			category: "collections",
			name:     "simple_dict",
			buildAST: func() ast.Node {
				return &ast.DictExpr{
					Pairs: []ast.DictPair{
						&ast.KeyValuePair{
							Key:   &ast.Literal{Value: "key", Type: ast.LiteralTypeString},
							Value: &ast.Literal{Value: "value", Type: ast.LiteralTypeString},
						},
					},
				}
			},
		},
		{
			category: "collections",
			name:     "empty_set",
			buildAST: func() ast.Node {
				return &ast.SetExpr{Elements: []ast.Expr{}}
			},
		},
		{
			category: "collections",
			name:     "simple_set",
			buildAST: func() ast.Node {
				return &ast.SetExpr{
					Elements: []ast.Expr{
						&ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
						&ast.Literal{Value: int64(2), Type: ast.LiteralTypeNumber},
					},
				}
			},
		},

		// Expressions
		{
			category: "expressions",
			name:     "binary_add",
			buildAST: func() ast.Node {
				return &ast.Binary{
					Left:     &ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
					Operator: lexer.Token{Type: lexer.Plus, Lexeme: "+"},
					Right:    &ast.Literal{Value: int64(2), Type: ast.LiteralTypeNumber},
				}
			},
		},
		{
			category: "expressions",
			name:     "attribute_access",
			buildAST: func() ast.Node {
				return &ast.Attribute{
					Object: &ast.Name{Token: lexer.Token{Lexeme: "obj"}},
					Name:   lexer.Token{Lexeme: "attr"},
				}
			},
		},
		{
			category: "expressions",
			name:     "function_call",
			buildAST: func() ast.Node {
				return &ast.Call{
					Callee: &ast.Name{Token: lexer.Token{Lexeme: "print"}},
					Arguments: []*ast.Argument{
						{Value: &ast.Literal{Value: "hello", Type: ast.LiteralTypeString}},
					},
				}
			},
		},
		{
			category: "expressions",
			name:     "subscript",
			buildAST: func() ast.Node {
				return &ast.Subscript{
					Object: &ast.Name{Token: lexer.Token{Lexeme: "lst"}},
					Indices: []ast.Expr{
						&ast.Literal{Value: int64(0), Type: ast.LiteralTypeNumber},
					},
				}
			},
		},
		{
			category: "expressions",
			name:     "slice_expression",
			buildAST: func() ast.Node {
				return &ast.Subscript{
					Object: &ast.Name{Token: lexer.Token{Lexeme: "lst"}},
					Indices: []ast.Expr{
						&ast.Slice{
							StartIndex: &ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
							EndIndex:   &ast.Literal{Value: int64(5), Type: ast.LiteralTypeNumber},
						},
					},
				}
			},
		},
		{
			category: "expressions",
			name:     "lambda",
			buildAST: func() ast.Node {
				return &ast.Lambda{
					Parameters: &ast.ParameterList{
						Parameters: []*ast.Parameter{
							{Name: &ast.Name{Token: lexer.Token{Lexeme: "x"}}},
						},
					},
					Body: &ast.Binary{
						Left:     &ast.Name{Token: lexer.Token{Lexeme: "x"}},
						Operator: lexer.Token{Type: lexer.Plus, Lexeme: "+"},
						Right:    &ast.Literal{Value: int64(1), Type: ast.LiteralTypeNumber},
					},
				}
			},
		},
		{
			category: "expressions",
			name:     "ternary",
			buildAST: func() ast.Node {
				return &ast.TernaryExpr{
					Condition: &ast.Name{Token: lexer.Token{Lexeme: "condition"}},
					TrueExpr:  &ast.Literal{Value: "yes", Type: ast.LiteralTypeString},
					FalseExpr: &ast.Literal{Value: "no", Type: ast.LiteralTypeString},
				}
			},
		},

		// Statements
		{
			category: "statements",
			name:     "simple_assignment",
			buildAST: func() ast.Node {
				return &ast.AssignStmt{
					Targets: []ast.Expr{&ast.Name{Token: lexer.Token{Lexeme: "x"}}},
					Value:   &ast.Literal{Value: int64(42), Type: ast.LiteralTypeNumber},
				}
			},
		},
		{
			category: "statements",
			name:     "if_statement",
			buildAST: func() ast.Node {
				return &ast.If{
					Condition: &ast.Name{Token: lexer.Token{Lexeme: "condition"}},
					Body: []ast.Stmt{
						&ast.ExprStmt{
							Expr: &ast.Call{
								Callee:    &ast.Name{Token: lexer.Token{Lexeme: "print"}},
								Arguments: []*ast.Argument{{Value: &ast.Literal{Value: "true", Type: ast.LiteralTypeString}}},
							},
						},
					},
				}
			},
		},
		{
			category: "statements",
			name:     "for_loop",
			buildAST: func() ast.Node {
				return &ast.For{
					Target:   &ast.Name{Token: lexer.Token{Lexeme: "i"}},
					Iterable: &ast.Name{Token: lexer.Token{Lexeme: "items"}},
					Body: []ast.Stmt{
						&ast.ExprStmt{
							Expr: &ast.Call{
								Callee:    &ast.Name{Token: lexer.Token{Lexeme: "print"}},
								Arguments: []*ast.Argument{{Value: &ast.Name{Token: lexer.Token{Lexeme: "i"}}}},
							},
						},
					},
				}
			},
		},
		{
			category: "statements",
			name:     "while_loop",
			buildAST: func() ast.Node {
				return &ast.While{
					Test: &ast.Name{Token: lexer.Token{Lexeme: "condition"}},
					Body: []ast.Stmt{
						&ast.PassStmt{},
					},
				}
			},
		},
		{
			category: "statements",
			name:     "return_statement",
			buildAST: func() ast.Node {
				return &ast.ReturnStmt{
					Value: &ast.Literal{Value: int64(42), Type: ast.LiteralTypeNumber},
				}
			},
		},

		// Functions
		{
			category: "functions",
			name:     "simple_function",
			buildAST: func() ast.Node {
				return &ast.Function{
					Name: &ast.Name{Token: lexer.Token{Lexeme: "greet"}},
					Parameters: &ast.ParameterList{
						Parameters: []*ast.Parameter{
							{Name: &ast.Name{Token: lexer.Token{Lexeme: "name"}}},
						},
					},
					Body: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.Binary{
								Left:     &ast.Literal{Value: "Hello, ", Type: ast.LiteralTypeString},
								Operator: lexer.Token{Type: lexer.Plus, Lexeme: "+"},
								Right:    &ast.Name{Token: lexer.Token{Lexeme: "name"}},
							},
						},
					},
				}
			},
		},
		{
			category: "functions",
			name:     "function_with_defaults",
			buildAST: func() ast.Node {
				return &ast.Function{
					Name: &ast.Name{Token: lexer.Token{Lexeme: "greet"}},
					Parameters: &ast.ParameterList{
						Parameters: []*ast.Parameter{
							{
								Name:    &ast.Name{Token: lexer.Token{Lexeme: "name"}},
								Default: &ast.Literal{Value: "World", Type: ast.LiteralTypeString},
							},
						},
					},
					Body: []ast.Stmt{
						&ast.PassStmt{},
					},
				}
			},
		},

		// Classes
		{
			category: "classes",
			name:     "simple_class",
			buildAST: func() ast.Node {
				return &ast.Class{
					Name: &ast.Name{Token: lexer.Token{Lexeme: "MyClass"}},
					Body: []ast.Stmt{
						&ast.PassStmt{},
					},
				}
			},
		},
		{
			category: "classes",
			name:     "class_with_method",
			buildAST: func() ast.Node {
				return &ast.Class{
					Name: &ast.Name{Token: lexer.Token{Lexeme: "Person"}},
					Body: []ast.Stmt{
						&ast.Function{
							Name: &ast.Name{Token: lexer.Token{Lexeme: "__init__"}},
							Parameters: &ast.ParameterList{
								Parameters: []*ast.Parameter{
									{Name: &ast.Name{Token: lexer.Token{Lexeme: "self"}}},
									{Name: &ast.Name{Token: lexer.Token{Lexeme: "name"}}},
								},
							},
							Body: []ast.Stmt{
								&ast.AssignStmt{
									Targets: []ast.Expr{
										&ast.Attribute{
											Object: &ast.Name{Token: lexer.Token{Lexeme: "self"}},
											Name:   lexer.Token{Lexeme: "name"},
										},
									},
									Value: &ast.Name{Token: lexer.Token{Lexeme: "name"}},
								},
							},
						},
					},
				}
			},
		},

		// Comprehensions
		{
			category: "comprehensions",
			name:     "list_comprehension",
			buildAST: func() ast.Node {
				return &ast.ListComp{
					Element: &ast.Name{Token: lexer.Token{Lexeme: "x"}},
					Clauses: []ast.ForIfClause{
						{
							Target: &ast.Name{Token: lexer.Token{Lexeme: "x"}},
							Iter:   &ast.Name{Token: lexer.Token{Lexeme: "items"}},
						},
					},
				}
			},
		},
		{
			category: "comprehensions",
			name:     "dict_comprehension",
			buildAST: func() ast.Node {
				return &ast.DictComp{
					Key:   &ast.Name{Token: lexer.Token{Lexeme: "k"}},
					Value: &ast.Name{Token: lexer.Token{Lexeme: "v"}},
					Clauses: []ast.ForIfClause{
						{
							Target: &ast.TupleExpr{
								Elements: []ast.Expr{
									&ast.Name{Token: lexer.Token{Lexeme: "k"}},
									&ast.Name{Token: lexer.Token{Lexeme: "v"}},
								},
							},
							Iter: &ast.Call{
								Callee: &ast.Attribute{
									Object: &ast.Name{Token: lexer.Token{Lexeme: "d"}},
									Name:   lexer.Token{Lexeme: "items"},
								},
								Arguments: []*ast.Argument{},
							},
						},
					},
				}
			},
		},
		{
			category: "comprehensions",
			name:     "set_comprehension",
			buildAST: func() ast.Node {
				return &ast.SetComp{
					Element: &ast.Name{Token: lexer.Token{Lexeme: "x"}},
					Clauses: []ast.ForIfClause{
						{
							Target: &ast.Name{Token: lexer.Token{Lexeme: "x"}},
							Iter:   &ast.Name{Token: lexer.Token{Lexeme: "items"}},
						},
					},
				}
			},
		},

		// F-strings
		{
			category: "fstring",
			name:     "simple_fstring",
			buildAST: func() ast.Node {
				return &ast.FString{
					Parts: []ast.FStringPart{
						&ast.FStringMiddle{Value: "Hello, "},
						&ast.FStringReplacementField{
							Expression: &ast.Name{Token: lexer.Token{Lexeme: "name"}},
						},
						&ast.FStringMiddle{Value: "!"},
					},
				}
			},
		},

		// Match statements
		{
			category: "patterns",
			name:     "match_statement",
			buildAST: func() ast.Node {
				return &ast.MatchStmt{
					Subject: &ast.Name{Token: lexer.Token{Lexeme: "status"}},
					Cases: []ast.CaseBlock{
						{
							Patterns: []ast.Pattern{
								&ast.LiteralPattern{
									Value: &ast.Literal{Value: int64(200), Type: ast.LiteralTypeNumber},
								},
							},
							Body: []ast.Stmt{
								&ast.ReturnStmt{
									Value: &ast.Literal{Value: "OK", Type: ast.LiteralTypeString},
								},
							},
						},
						{
							Patterns: []ast.Pattern{
								&ast.WildcardPattern{},
							},
							Body: []ast.Stmt{
								&ast.ReturnStmt{
									Value: &ast.Literal{Value: "Unknown", Type: ast.LiteralTypeString},
								},
							},
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			if tc.skipReason != "" {
				t.Skip(tc.skipReason)
			}

			// Generate code
			gen := NewCodeGenerator()
			generated := gen.Generate(tc.buildAST())

			// Define paths
			expectedPath := filepath.Join("testdata", "expected", tc.category, tc.name+".py")
			generatedDir := filepath.Join("testdata", "generated", tc.category)
			generatedPath := filepath.Join(generatedDir, tc.name+".py")

			// Always write generated output to testdata/generated for comparison
			if err := os.MkdirAll(generatedDir, 0755); err != nil {
				t.Fatalf("Failed to create generated directory: %v", err)
			}
			err := os.WriteFile(generatedPath, []byte(generated), 0644)
			if err != nil {
				t.Fatalf("Failed to write generated file: %v", err)
			}

			if updateGolden {
				// Ensure expected directory exists
				if err := os.MkdirAll(filepath.Dir(expectedPath), 0755); err != nil {
					t.Fatalf("Failed to create expected directory: %v", err)
				}

				// Update golden file
				err := os.WriteFile(expectedPath, []byte(generated), 0644)
				if err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", expectedPath)
				t.Logf("Generated file: %s", generatedPath)
				return
			}

			// Compare with expected output
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				if os.IsNotExist(err) {
					// Create golden file if it doesn't exist
					if err := os.MkdirAll(filepath.Dir(expectedPath), 0755); err != nil {
						t.Fatalf("Failed to create expected directory: %v", err)
					}
					err := os.WriteFile(expectedPath, []byte(generated), 0644)
					if err != nil {
						t.Fatalf("Failed to create golden file: %v", err)
					}
					t.Logf("Created golden file: %s", expectedPath)
					t.Logf("Generated file: %s", generatedPath)
					return
				}
				t.Fatalf("Failed to read expected file: %v", err)
			}

			// Compare
			if normalizeOutput(generated) != normalizeOutput(string(expected)) {
				t.Errorf("Output differs from expected:\nExpected file: %s\nGenerated file: %s\n\nDifferences:\nGot:\n%s\nExpected:\n%s", expectedPath, generatedPath, generated, string(expected))
				showDiff(t, string(expected), generated)
			} else {
				t.Logf("✅ Test passed - Generated file: %s", generatedPath)
			}
		})
	}
}

// normalizeOutput normalizes the output by trimming trailing whitespace
func normalizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// showDiff shows the difference between expected and generated strings
func showDiff(t *testing.T, expected, generated string) {
	expectedLines := strings.Split(expected, "\n")
	generatedLines := strings.Split(generated, "\n")

	maxLines := len(expectedLines)
	if len(generatedLines) > maxLines {
		maxLines = len(generatedLines)
	}

	t.Logf("\n=== Diff ===")
	for i := 0; i < maxLines; i++ {
		var expectedLine, generatedLine string
		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(generatedLines) {
			generatedLine = generatedLines[i]
		}

		if expectedLine != generatedLine {
			t.Logf("Line %d:", i+1)
			t.Logf("  Expected:  %q", expectedLine)
			t.Logf("  Generated: %q", generatedLine)
		}
	}
}
