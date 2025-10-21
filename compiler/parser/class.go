package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// classStatement parses a class statement according to the grammar:
//
// class_def:
//
//	| decorators class_def_raw
//	| class_def_raw
//
// class_def_raw:
//
//	| 'class' NAME [type_params] ['(' [arguments] ')' ] ':' block
func (p *Parser) classStatement() (ast.Stmt, error) {
	// Consume the 'class' keyword
	classToken, err := p.consume(lexer.Class, "expected 'class'")
	if err != nil {
		return nil, err
	}

	// Parse the class name
	nameToken, err := p.consume(lexer.Identifier, "expected class name")
	if err != nil {
		return nil, err
	}
	className := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Parse optional type parameters - converting from Expr to TypeParam
	var typeParams []ast.TypeParam
	if p.check(lexer.LeftBracket) {
		// For simplicity, we'll assume typeParams returns compatible type params
		// In a real implementation, this would need proper conversion or a dedicated type param parser
		exprs, err := p.typeParams()
		if err != nil {
			return nil, err
		}

		// Convert from Expr to TypeParam (this is a simplification)
		typeParams = make([]ast.TypeParam, len(exprs))
		for i, expr := range exprs {
			if tpExpr, ok := expr.(*ast.TypeParam); ok {
				typeParams[i] = *tpExpr
			}
		}
	}

	// Parse optional parent classes/arguments
	var args []*ast.Argument
	if p.match(lexer.LeftParen) {
		// Check for empty argument list
		if !p.check(lexer.RightParen) {
			args, err = p.arguments()
			if err != nil {
				return nil, err
			}
		}

		// Consume the closing parenthesis
		_, err = p.consume(lexer.RightParen, "expected ')' after class inheritance list")
		if err != nil {
			return nil, err
		}
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after class definition")
	if err != nil {
		return nil, err
	}

	// Parse the class body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	return &ast.Class{
		Name:       className,
		TypeParams: typeParams,
		Args:       args,
		Body:       body,
		Span:       lexer.Span{Start: classToken.Start(), End: endPos},
	}, nil
}
