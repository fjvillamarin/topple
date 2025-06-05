package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// params parses function parameters according to the grammar:
// params:
//
//	| parameters
func (p *Parser) params() (*ast.ParameterList, error) {
	return p.parameters()
}

// parameters parses a parameter list according to the grammar:
// parameters:
//
//	| slash_no_default param_no_default* param_with_default* [star_etc]
//	| slash_with_default param_with_default* [star_etc]
//	| param_no_default+ param_with_default* [star_etc]
//	| param_with_default+ [star_etc]
//	| star_etc
func (p *Parser) parameters() (*ast.ParameterList, error) {
	startPos := p.peek().Start()

	// Create an empty parameter list
	paramList := &ast.ParameterList{
		Parameters:  []*ast.Parameter{},
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
	}

	// Try to parse the parameter list
	// We'll use a more flexible approach rather than trying to strictly match the grammar rules
	// This allows us to handle a variety of parameter patterns

	// Handle empty parameter list
	if p.check(lexer.RightParen) {
		paramList.Span = lexer.Span{Start: startPos, End: p.peek().Start()}
		return paramList, nil
	}

	// First, try to parse any regular parameters (with or without defaults)
	hasSlash := false
	hasStarArg := false

	// Track where we are in the parameter parsing process
	// 0 = before slash, 1 = after slash but before star, 2 = after star
	paramPhase := 0

	for !p.check(lexer.RightParen) && !p.isAtEnd() {
		// Check for star parameter (*args)
		if p.match(lexer.Star) {
			if hasStarArg {
				return nil, p.error(p.previous(), "only one *args parameter allowed")
			}

			starToken := p.previous()

			// Check for empty star parameter (*,) - this is the keyword-only marker
			if p.match(lexer.Comma) {
				// This is a bare star, marking keyword-only parameters
				// Don't add it as a parameter, just mark the phase
				hasStarArg = true
				paramPhase = 2
				continue
			}

			// Parse the parameter after *
			if !p.check(lexer.Identifier) {
				return nil, p.error(p.peek(), "expected parameter name after '*'")
			}

			// Parse the parameter
			param, err := p.parseParam()
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
			paramPhase = 2

			// Consume comma if present
			p.match(lexer.Comma)
		} else if p.match(lexer.StarStar) {
			// Check for double star parameter (**kwargs)
			doubleStarToken := p.previous()

			// Parse the parameter after **
			if !p.check(lexer.Identifier) {
				return nil, p.error(p.peek(), "expected parameter name after '**'")
			}

			// Parse the parameter
			param, err := p.parseParam()
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
			p.match(lexer.Comma)
			if !p.check(lexer.RightParen) {
				return nil, p.error(p.peek(), "**kwargs must be the last parameter")
			}
		} else if p.match(lexer.Slash) {
			// Positional-only parameters separator
			if hasSlash {
				return nil, p.error(p.previous(), "multiple '/' in parameter list")
			}
			if hasStarArg {
				return nil, p.error(p.previous(), "'/' after *args")
			}
			if len(paramList.Parameters) == 0 {
				return nil, p.error(p.previous(), "'/' without preceding parameters")
			}

			// Mark the position in the parameter list where the slash occurs
			slashIndex := len(paramList.Parameters) - 1
			paramList.SlashIndex = slashIndex
			paramList.HasSlash = true
			hasSlash = true
			paramPhase = 1

			// Consume comma if present - trailing comma after slash is allowed
			p.match(lexer.Comma)
		} else if p.check(lexer.Identifier) {
			// Regular parameter
			param, err := p.parseParam()
			if err != nil {
				return nil, err
			}

			// Check that we're not adding a parameter without default after one with default
			// unless we're after a star arg
			if param.Default == nil && paramPhase < 2 {
				// Check if we've already seen a parameter with default
				for _, prevParam := range paramList.Parameters {
					if prevParam.Default != nil && !prevParam.IsStar && !prevParam.IsDoubleStar {
						if paramPhase == 0 && !hasSlash {
							return nil, p.error(p.previous(), "non-default parameter follows default parameter")
						}
					}
				}
			}

			// Add the parameter
			paramList.Parameters = append(paramList.Parameters, param)

			// Consume comma if present
			p.match(lexer.Comma)
		} else {
			return nil, p.error(p.peek(), "unexpected token in parameter list")
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

// parseParam parses a single parameter including annotation and default value if present
func (p *Parser) parseParam() (*ast.Parameter, error) {
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

	// Check for annotation
	if p.match(lexer.Colon) {
		// Parse the annotation
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Set the annotation
		param.Annotation = expr

		// Update the span to include the annotation
		param.Span = lexer.Span{Start: param.GetSpan().Start, End: expr.GetSpan().End}
	}

	// Check for default value
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
