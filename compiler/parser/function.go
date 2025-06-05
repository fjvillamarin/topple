package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// functionDef parses a function definition according to the grammar:
//
// function_def:
//
//	| decorators function_def_raw
//	| function_def_raw
func (p *Parser) functionDef() (ast.Stmt, error) {
	// Check for decorators - handled at the statement level now
	// Decorators wrap around the function instead of being part of it

	// Parse the function definition
	return p.functionDefRaw()
}

// functionDefRaw parses a raw function definition according to the grammar:
//
// function_def_raw:
//
//	| 'def' NAME [type_params] '(' [params] ')' ['->' expression ] ':' [func_type_comment] block
//	| 'async' 'def' NAME [type_params] '(' [params] ')' ['->' expression ] ':' [func_type_comment] block
func (p *Parser) functionDefRaw() (ast.Stmt, error) {
	// Track if this is an async function
	isAsync := false
	var startToken lexer.Token

	// Check for 'async' keyword
	if p.match(lexer.Async) {
		isAsync = true
		startToken = p.previous()
		_, err := p.consume(lexer.Def, "expected 'def' after 'async'")
		if err != nil {
			return nil, err
		}
	} else {
		// Consume the 'def' keyword
		var err error
		startToken, err = p.consume(lexer.Def, "expected 'def'")
		if err != nil {
			return nil, err
		}
	}

	// Parse the function name
	nameToken, err := p.consume(lexer.Identifier, "expected function name")
	if err != nil {
		return nil, err
	}
	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Parse optional type parameters
	var typeParams []*ast.TypeParam
	if p.check(lexer.LeftBracket) {
		// Get the type parameters using the existing parser function
		paramExprs, err := p.typeParams()
		if err != nil {
			return nil, err
		}

		// Convert from Expr to *TypeParam
		typeParams = make([]*ast.TypeParam, 0, len(paramExprs))
		for _, expr := range paramExprs {
			if tp, ok := expr.(*ast.TypeParam); ok {
				typeParams = append(typeParams, tp)
			}
		}
	}

	// Expect opening parenthesis for parameters
	_, err = p.consume(lexer.LeftParen, "expected '(' after function name")
	if err != nil {
		return nil, err
	}

	// Parse function parameters (if any)
	var parameterList *ast.ParameterList
	if !p.check(lexer.RightParen) {
		parameterList, err = p.params()
		if err != nil {
			return nil, err
		}
	} else {
		// Empty parameter list
		parameterList = &ast.ParameterList{
			Parameters:  []*ast.Parameter{},
			SlashIndex:  -1,
			VarArgIndex: -1,
			KwArgIndex:  -1,
			Span:        lexer.Span{Start: p.previous().End(), End: p.peek().Start()},
		}
	}

	// Expect closing parenthesis
	_, err = p.consume(lexer.RightParen, "expected ')' after function parameters")
	if err != nil {
		return nil, err
	}

	// Parse optional return type annotation
	var returnType ast.Expr
	if p.match(lexer.Arrow) {
		returnType, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after function header")
	if err != nil {
		return nil, err
	}

	// Ignore func_type_comment as instructed

	// Parse function body
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

	// Create and return the Function node
	return &ast.Function{
		Name:           name,
		TypeParameters: typeParams,
		ReturnType:     returnType,
		Parameters:     parameterList,
		Body:           body,
		IsAsync:        isAsync,
		Span:           lexer.Span{Start: startToken.Start(), End: endPos},
	}, nil
}
