package transformers

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// GetRequiredImports returns the import statements required for the transformed view
func (vm *ViewTransformer) GetRequiredImports() []*ast.ImportFromStmt {
	var imports []*ast.ImportFromStmt

	if vm.needsRuntimeImports {
		// Create single combined import: from topple.psx import BaseView, Element, el, escape, fragment, raw
		runtimeImport := &ast.ImportFromStmt{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{
							Lexeme: "topple",
							Type:   lexer.Identifier,
						},
						Span: lexer.Span{},
					},
					{
						Token: lexer.Token{
							Lexeme: "psx",
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
				{
					DottedName: &ast.DottedName{
						Names: []*ast.Name{
							{
								Token: lexer.Token{
									Lexeme: "raw",
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
		imports = append(imports, runtimeImport)
	}

	return imports
}
