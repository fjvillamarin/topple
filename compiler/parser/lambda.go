package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// lambdef parses a lambda expression according to the grammar:
// lambdef: 'lambda' [lambda_params] ':' expression
func (p *Parser) lambdef() (ast.Expr, error) {
	// Consume the 'lambda' keyword
	lambdaToken, err := p.consume(lexer.Lambda, "expected 'lambda'")
	if err != nil {
		return nil, err
	}

	// Parse optional lambda parameters
	var parameters *ast.ParameterList = nil
	if !p.check(lexer.Colon) {
		// Parse lambda parameters (simplified - we'll use regular parameter parsing)
		parameters, err = p.lambdaParams()
		if err != nil {
			return nil, err
		}
	}

	// Consume the ':'
	_, err = p.consume(lexer.Colon, "expected ':' after lambda parameters")
	if err != nil {
		return nil, err
	}

	// Parse the body expression
	body, err := p.expression()
	if err != nil {
		return nil, err
	}

	return &ast.Lambda{
		Parameters: parameters,
		Body:       body,

		Span: lexer.Span{Start: lambdaToken.Start(), End: body.GetSpan().End},
	}, nil
}

// lambdaParams parses lambda parameters according to the grammar:
// lambda_params: lambda_parameters
// lambda_parameters:
//
//	| lambda_param_no_default+ lambda_param_with_default* [lambda_star_etc]
//	| lambda_param_with_default+ [lambda_star_etc]
//	| lambda_star_etc
func (p *Parser) lambdaParams() (*ast.ParameterList, error) {
	startPos := p.peek().Start()

	// Create an empty parameter list
	paramList := &ast.ParameterList{
		Parameters:  []*ast.Parameter{},
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
	}

	// Handle empty parameter list (lambda:)
	if p.check(lexer.Colon) {
		paramList.Span = lexer.Span{Start: startPos, End: p.peek().Start()}
		return paramList, nil
	}

	hasStarArg := false

	for !p.check(lexer.Colon) && !p.isAtEnd() {
		// Check for star parameter (*args)
		if p.match(lexer.Star) {
			if hasStarArg {
				return nil, p.error(p.previous(), "only one *args parameter allowed")
			}

			starToken := p.previous()

			// Check for empty star parameter (*,)
			if p.match(lexer.Comma) {
				// Create empty vararg parameter
				varargParam := &ast.Parameter{
					IsStar: true,
					Span:   lexer.Span{Start: starToken.Start(), End: starToken.End()},
				}

				paramList.Parameters = append(paramList.Parameters, varargParam)
				paramList.HasVarArg = true
				paramList.VarArgIndex = len(paramList.Parameters) - 1

				hasStarArg = true
				continue
			}

			// Parse the parameter after *
			if !p.check(lexer.Identifier) {
				return nil, p.error(p.peek(), "expected parameter name after '*'")
			}

			// Parse lambda parameter (no type annotation)
			param, err := p.lambdaParam()
			if err != nil {
				return nil, err
			}

			// Set star flag
			param.IsStar = true
			param.Span = lexer.Span{Start: starToken.Start(), End: param.GetSpan().End}

			// Add to parameter list
			paramList.Parameters = append(paramList.Parameters, param)
			paramList.HasVarArg = true
			paramList.VarArgIndex = len(paramList.Parameters) - 1

			hasStarArg = true

			// Consume comma if present
			if p.match(lexer.Comma) && p.check(lexer.Colon) {
				break
			}
		} else if p.match(lexer.StarStar) {
			// Check for double star parameter (**kwargs)
			doubleStarToken := p.previous()

			// Parse the parameter after **
			if !p.check(lexer.Identifier) {
				return nil, p.error(p.peek(), "expected parameter name after '**'")
			}

			// Parse lambda parameter (no type annotation)
			param, err := p.lambdaParam()
			if err != nil {
				return nil, err
			}

			// Set double star flag
			param.IsDoubleStar = true
			param.Span = lexer.Span{Start: doubleStarToken.Start(), End: param.GetSpan().End}

			// Add to parameter list
			paramList.Parameters = append(paramList.Parameters, param)
			paramList.HasKwArg = true
			paramList.KwArgIndex = len(paramList.Parameters) - 1

			// **kwargs must be the last parameter, but can have a trailing comma
			if p.match(lexer.Comma) && !p.check(lexer.Colon) {
				return nil, p.error(p.peek(), "**kwargs must be the last parameter")
			}
			break
		} else if p.check(lexer.Identifier) {
			// Regular lambda parameter
			param, err := p.lambdaParam()
			if err != nil {
				return nil, err
			}

			// Add the parameter
			paramList.Parameters = append(paramList.Parameters, param)

			// Consume comma if present
			if p.match(lexer.Comma) && p.check(lexer.Colon) {
				break
			}
		} else {
			return nil, p.error(p.peek(), "unexpected token in lambda parameter list")
		}
	}

	// Update the span based on the last parameter
	if len(paramList.Parameters) > 0 {
		endPos := paramList.Parameters[len(paramList.Parameters)-1].GetSpan().End
		paramList.Span = lexer.Span{Start: startPos, End: endPos}
	} else {
		endPos := p.peek().Start()
		paramList.Span = lexer.Span{Start: startPos, End: endPos}
	}

	return paramList, nil
}

// lambdaParam parses a single lambda parameter (no type annotations, just name and optional default)
// lambda_param: NAME
func (p *Parser) lambdaParam() (*ast.Parameter, error) {
	// Parse the parameter name
	nameToken, err := p.consume(lexer.Identifier, "expected parameter name")
	if err != nil {
		return nil, err
	}

	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Create parameter with name
	param := &ast.Parameter{
		Name: name,
		Span: name.Span,
	}

	// Check for default value (lambda parameters don't have type annotations)
	if p.match(lexer.Equal) {
		// Parse the default value expression
		defaultExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Set the default value
		param.Default = defaultExpr

		// Update the span to include the default
		param.Span = lexer.Span{Start: param.GetSpan().Start, End: defaultExpr.GetSpan().End}
	}

	return param, nil
}
