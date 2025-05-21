package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// typeAlias parses a type alias statement as per the grammar:
// type_alias: "type" NAME [type_params] '=' expression
func (p *Parser) typeAlias() (ast.Stmt, error) {
	// Consume the 'type' keyword
	typeToken, err := p.consume(lexer.Type, "expected 'type'")
	if err != nil {
		return nil, err
	}

	// Parse the type name
	name, err := p.consume(lexer.Identifier, "expected type name")
	if err != nil {
		return nil, err
	}

	// Check for '[' to parse type parameters
	var params []ast.Expr = nil
	if p.match(lexer.LeftBracket) {
		// Parse the type parameters
		params, err = p.typeParams()
		if err != nil {
			return nil, err
		}
	}

	// Consume the '='
	_, err = p.consume(lexer.Equal, "expected '='")
	if err != nil {
		return nil, err
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	return &ast.TypeAlias{
		Name:   name,
		Params: params,
		Value:  expr,

		Span: lexer.Span{Start: typeToken.Start(), End: expr.GetSpan().End},
	}, nil
}

// typeParams parses type parameters as per the grammar:
// type_params: '[' type_param_seq ']'
// type_param_seq: ','.type_param+ [',']
func (p *Parser) typeParams() ([]ast.Expr, error) {
	params := []ast.Expr{}

	// Parse type parameters until we hit a closing bracket
	for !p.check(lexer.RightBracket) && !p.isAtEnd() {
		// Parse a single type parameter
		param, err := p.typeParam()
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		// If no comma, we're done with the parameter list
		if !p.match(lexer.Comma) {
			break
		}
	}

	// Consume the closing bracket
	_, err := p.consume(lexer.RightBracket, "expected ']' after type parameters")
	if err != nil {
		return nil, err
	}

	return params, nil
}

// typeParam parses a single type parameter as per the grammar:
// type_param:
//
//	| NAME [type_param_bound] [type_param_default]
//	| '*' NAME [type_param_starred_default]
//	| '**' NAME [type_param_default]
func (p *Parser) typeParam() (ast.Expr, error) {
	startPos := p.peek().Start()
	isStar := false
	isDoubleStar := false

	// Check for star parameters
	if p.match(lexer.Star) {
		isStar = true
	} else if p.match(lexer.StarStar) {
		isDoubleStar = true
	}

	// Parse the parameter name
	name, err := p.consume(lexer.Identifier, "expected parameter name")
	if err != nil {
		return nil, err
	}

	// Parse optional bound (: expression)
	var bound ast.Expr = nil
	if !isStar && !isDoubleStar && p.match(lexer.Colon) {
		bound, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional default
	var defaultValue ast.Expr = nil
	if p.match(lexer.Equal) {
		if isStar {
			// For star parameters, the default is a star_expression
			defaultValue, err = p.starExpression()
		} else {
			// For regular and double-star parameters, the default is a regular expression
			defaultValue, err = p.expression()
		}

		if err != nil {
			return nil, err
		}
	}

	endPos := p.previous().End()
	if defaultValue != nil {
		endPos = defaultValue.GetSpan().End
	} else if bound != nil {
		endPos = bound.GetSpan().End
	}

	return &ast.TypeParamExpr{
		Name:         name,
		Bound:        bound,
		Default:      defaultValue,
		IsStar:       isStar,
		IsDoubleStar: isDoubleStar,

		Span: lexer.Span{Start: startPos, End: endPos},
	}, nil
}
