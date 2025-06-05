package transformers

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// GetRequiredImports returns the import statements required for the transformed view
func (vm *ViewTransformer) GetRequiredImports() []*ast.ImportFromStmt {
	var imports []*ast.ImportFromStmt

	if vm.needsRuntimeImports {
		// Import BaseView
		baseViewImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "psx_runtime",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			Names: []*ast.ImportName{
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "BaseView",
									Type:   lexer.Identifier,
								},
								Span: lexer.Span{},
							},
						},
						Span: lexer.Span{},
					},
					AsName: nil,
					Span:   lexer.Span{},
				},
			},
			Span: lexer.Span{},
		}
		imports = append(imports, baseViewImport)

		// Import Element type
		elementImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "psx_runtime",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			Names: []*ast.ImportName{
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "Element",
									Type:   lexer.Identifier,
								},
								Span: lexer.Span{},
							},
						},
						Span: lexer.Span{},
					},
					AsName: nil,
					Span:   lexer.Span{},
				},
			},
			Span: lexer.Span{},
		}
		imports = append(imports, elementImport)

		// Import el function
		elImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "psx_runtime",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			Names: []*ast.ImportName{
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "el",
									Type:   lexer.Identifier,
								},
								Span: lexer.Span{},
							},
						},
						Span: lexer.Span{},
					},
					AsName: nil,
					Span:   lexer.Span{},
				},
			},
			Span: lexer.Span{},
		}
		imports = append(imports, elImport)

		// Import escape function
		escapeImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "psx_runtime",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			Names: []*ast.ImportName{
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "escape",
									Type:   lexer.Identifier,
								},
								Span: lexer.Span{},
							},
						},
						Span: lexer.Span{},
					},
					AsName: nil,
					Span:   lexer.Span{},
				},
			},
			Span: lexer.Span{},
		}
		imports = append(imports, escapeImport)

		// Import fragment function
		fragmentImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "psx_runtime",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			Names: []*ast.ImportName{
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "fragment",
									Type:   lexer.Identifier,
								},
								Span: lexer.Span{},
							},
						},
						Span: lexer.Span{},
					},
					AsName: nil,
					Span:   lexer.Span{},
				},
			},
			Span: lexer.Span{},
		}
		imports = append(imports, fragmentImport)
	}

	return imports
}