package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// viewStatement parses a view statement according to the grammar:
// view_def: 'view' NAME [type_params] '(' [params] ')' ['->' expression] ':' view_block
func (p *Parser) viewStatement() (ast.Stmt, error) {
	// Consume the 'view' keyword
	viewToken, err := p.consume(lexer.View, "expected 'view'")
	if err != nil {
		return nil, err
	}

	// Parse the view name
	nameToken, err := p.consume(lexer.Identifier, "expected view name")
	if err != nil {
		return nil, err
	}
	viewName := &ast.Name{
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
	_, err = p.consume(lexer.LeftParen, "expected '(' after view name")
	if err != nil {
		return nil, err
	}

	// Parse view parameters (if any)
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
	_, err = p.consume(lexer.RightParen, "expected ')' after view parameters")
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
	_, err = p.consume(lexer.Colon, "expected ':' after view definition")
	if err != nil {
		return nil, err
	}

	// Parse the view body
	body, err := p.viewBlock()
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

	return &ast.ViewStmt{
		Name:       viewName,
		TypeParams: typeParams,
		Params:     parameterList,
		ReturnType: returnType,
		Body:       body,
		IsAsync:    false, // Views are not async in this POC
		Kind:       ast.ViewKindServerView,

		Span: lexer.Span{Start: viewToken.Start(), End: endPos},
	}, nil
}

// viewBlock parses a view block which can contain Python statements and HTML elements
func (p *Parser) viewBlock() ([]ast.Stmt, error) {
	// Check if this is a simple statement block (single line)
	if !p.check(lexer.Newline) {
		stmt, err := p.viewStatement_inner()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	// Otherwise expect NEWLINE INDENT statements DEDENT
	_, err := p.consume(lexer.Newline, "expected newline")
	if err != nil {
		return nil, err
	}

	// Consume all remaining newlines
	for p.check(lexer.Newline) {
		p.advance()
	}

	_, err = p.consume(lexer.Indent, "expected indented block")
	if err != nil {
		return nil, err
	}

	statements := []ast.Stmt{}
	for !p.isAtEnd() && !p.check(lexer.Dedent) {
		stmt, err := p.viewStatement_inner()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)

		// Consume all the newlines we see
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	_, err = p.consume(lexer.Dedent, "expected dedent at end of view block")
	if err != nil {
		return nil, err
	}

	return statements, nil
}

// viewStatement_inner parses statements that can appear inside a view body
func (p *Parser) viewStatement_inner() (ast.Stmt, error) {
	// Check for HTML elements first
	if p.check(lexer.TagOpen) {
		return p.htmlElement()
	}

	// Check for compound statements
	switch p.peek().Type {
	case lexer.Def:
		return p.viewFunctionDef()
	case lexer.Async:
		// For async statements, we need to look ahead
		if p.checkNext(lexer.Def) {
			return p.viewFunctionDef()
		}
		if p.checkNext(lexer.For) {
			return p.viewForStatement()
		}
		if p.checkNext(lexer.With) {
			// TODO: Implement async with in views when needed
			return nil, p.error(p.peek(), "async with statements not implemented in views yet")
		}
		// Fall through to simple statements for other async cases
		return p.simpleStatement()
	case lexer.If:
		return p.viewIfStatement()
	case lexer.While:
		return p.viewWhileStatement()
	case lexer.For:
		return p.viewForStatement()
	case lexer.Try:
		return p.viewTryStatement()
	case lexer.Match:
		return p.viewMatchStatement()
	}

	// Check for simple statements
	switch p.peek().Type {
	case lexer.Return:
		return p.returnStatement()
	case lexer.Pass:
		return p.passStatement()
	// TODO: Add more simple statements as needed
	default:
		// Fall back to regular statement parsing for Python statements
		return p.simpleStatement()
	}
}

// htmlElement parses an HTML element
func (p *Parser) htmlElement() (ast.Stmt, error) {
	// Consume opening '<'
	openToken, err := p.consume(lexer.TagOpen, "expected '<'")
	if err != nil {
		return nil, err
	}

	// Parse tag name
	tagNameToken, err := p.consume(lexer.Identifier, "expected tag name")
	if err != nil {
		return nil, err
	}

	// Parse attributes
	var attributes []ast.HTMLAttribute
	for !p.check(lexer.TagSelfClose) && !p.check(lexer.TagClose) && p.check(lexer.Identifier) {
		attr, err := p.htmlAttribute()
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attr)
	}

	// Check for self-closing tag
	if p.match(lexer.TagSelfClose) {
		return &ast.HTMLElement{
			Type:       ast.HTMLSelfClosingTag,
			TagName:    tagNameToken,
			Attributes: attributes,
			Content:    nil,
			IsClosing:  false,
			Span:       lexer.Span{Start: openToken.Start(), End: p.previous().End()},
		}, nil
	}

	// Consume closing '>'
	_, err = p.consume(lexer.TagClose, "expected '>' after tag")
	if err != nil {
		return nil, err
	}

	// Parse element content
	content, elementType, err := p.htmlElementContent(tagNameToken)
	if err != nil {
		return nil, err
	}

	return &ast.HTMLElement{
		Type:       elementType,
		TagName:    tagNameToken,
		Attributes: attributes,
		Content:    content,
		IsClosing:  false,
		Span:       lexer.Span{Start: openToken.Start(), End: p.previous().End()},
	}, nil
}

// htmlElementContent parses the content of an HTML element and determines if it's single-line or multiline
func (p *Parser) htmlElementContent(tagNameToken lexer.Token) ([]ast.Stmt, ast.HTMLElementType, error) {
	var content []ast.Stmt

	// Check if this is immediately followed by a closing tag (empty element)
	if p.check(lexer.TagCloseStart) {
		err := p.consumeClosingTag(tagNameToken)
		if err != nil {
			return nil, ast.HTMLSingleLineElement, err
		}
		return content, ast.HTMLSingleLineElement, nil
	}

	// Check if content starts on a new line (multiline element)
	if p.check(lexer.Newline) {
		return p.parseMultilineContent(tagNameToken)
	}

	// Parse single-line content
	return p.parseSingleLineContent(tagNameToken)
}

// parseMultilineContent parses multiline HTML element content
func (p *Parser) parseMultilineContent(tagNameToken lexer.Token) ([]ast.Stmt, ast.HTMLElementType, error) {
	var content []ast.Stmt

	// Consume the newline
	_, err := p.consume(lexer.Newline, "expected newline")
	if err != nil {
		return nil, ast.HTMLMultilineElement, err
	}

	// Consume all remaining newlines
	for p.check(lexer.Newline) {
		p.advance()
	}

	// Expect indented content
	_, err = p.consume(lexer.Indent, "expected indented content")
	if err != nil {
		return nil, ast.HTMLMultilineElement, err
	}

	// Parse content until dedent
	for !p.isAtEnd() && !p.check(lexer.Dedent) {
		// Skip empty lines
		if p.check(lexer.Newline) {
			p.advance()
			continue
		}

		// Handle nested HTML elements or Python statements only
		// Multiline HTML does NOT support raw HTML text
		stmt, err := p.viewStatement_inner()
		if err != nil {
			return nil, ast.HTMLMultilineElement, err
		}

		if stmt != nil {
			content = append(content, stmt)
		}

		// Consume newlines after statements
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	// Consume dedent
	_, err = p.consume(lexer.Dedent, "expected dedent")
	if err != nil {
		return nil, ast.HTMLMultilineElement, err
	}

	// Parse closing tag
	err = p.consumeClosingTag(tagNameToken)
	if err != nil {
		return nil, ast.HTMLMultilineElement, err
	}

	return content, ast.HTMLMultilineElement, nil
}

// parseSingleLineContent parses single-line HTML element content
func (p *Parser) parseSingleLineContent(tagNameToken lexer.Token) ([]ast.Stmt, ast.HTMLElementType, error) {
	var content []ast.Stmt

	// Collect HTML content parts (text and interpolations)
	htmlParts, err := p.parseHTMLContentParts()
	if err != nil {
		return nil, ast.HTMLSingleLineElement, err
	}

	// If we have HTML content, create an HTMLContent node
	if len(htmlParts) > 0 {
		htmlContent := &ast.HTMLContent{
			Parts: htmlParts,
			Span:  lexer.Span{Start: htmlParts[0].GetSpan().Start, End: htmlParts[len(htmlParts)-1].GetSpan().End},
		}
		content = append(content, htmlContent)
	}

	// Check for nested HTML elements after text content
	for !p.check(lexer.TagCloseStart) && !p.check(lexer.Newline) && !p.isAtEnd() {
		if p.check(lexer.TagOpen) {
			nestedElement, err := p.htmlElement()
			if err != nil {
				return nil, ast.HTMLSingleLineElement, err
			}
			content = append(content, nestedElement)

			// Parse any additional HTML content after the nested element
			moreParts, err := p.parseHTMLContentParts()
			if err != nil {
				return nil, ast.HTMLSingleLineElement, err
			}
			if len(moreParts) > 0 {
				moreContent := &ast.HTMLContent{
					Parts: moreParts,
					Span:  lexer.Span{Start: moreParts[0].GetSpan().Start, End: moreParts[len(moreParts)-1].GetSpan().End},
				}
				content = append(content, moreContent)
			}
		} else {
			break
		}
	}

	// Parse closing tag
	err = p.consumeClosingTag(tagNameToken)
	if err != nil {
		return nil, ast.HTMLSingleLineElement, err
	}

	return content, ast.HTMLSingleLineElement, nil
}

// parseHTMLContentParts parses consecutive HTML text and interpolations
func (p *Parser) parseHTMLContentParts() ([]ast.HTMLContentPart, error) {
	var parts []ast.HTMLContentPart

	for !p.check(lexer.TagCloseStart) && !p.check(lexer.TagOpen) && !p.check(lexer.Newline) && !p.isAtEnd() {
		if p.check(lexer.HTMLTextInline) {
			// Handle HTML text
			textToken := p.advance()
			htmlText := &ast.HTMLText{
				Value: textToken.Literal.(string),
				Span:  lexer.Span{Start: textToken.Start(), End: textToken.End()},
			}
			parts = append(parts, htmlText)
		} else if p.check(lexer.HTMLInterpolationStart) {
			// Handle interpolation {expression}
			startToken := p.advance() // consume '{'
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}
			endToken, err := p.consume(lexer.HTMLInterpolationEnd, "expected '}' after interpolation")
			if err != nil {
				return nil, err
			}

			htmlInterpolation := &ast.HTMLInterpolation{
				Expression: expr,
				Span:       lexer.Span{Start: startToken.Start(), End: endToken.End()},
			}
			parts = append(parts, htmlInterpolation)
		} else {
			break
		}
	}

	return parts, nil
}

// consumeClosingTag parses and validates a closing tag
func (p *Parser) consumeClosingTag(expectedTagName lexer.Token) error {
	_, err := p.consume(lexer.TagCloseStart, "expected closing tag")
	if err != nil {
		return err
	}

	closingTagName, err := p.consume(lexer.Identifier, "expected closing tag name")
	if err != nil {
		return err
	}

	// Verify tag names match
	if expectedTagName.Lexeme != closingTagName.Lexeme {
		return p.error(closingTagName, "closing tag name doesn't match opening tag")
	}

	_, err = p.consume(lexer.TagClose, "expected '>' after closing tag")
	return err
}

// htmlAttribute parses an HTML attribute
func (p *Parser) htmlAttribute() (ast.HTMLAttribute, error) {
	// Parse attribute name
	nameToken, err := p.consume(lexer.Identifier, "expected attribute name")
	if err != nil {
		return ast.HTMLAttribute{}, err
	}

	// Check for attribute value
	if p.match(lexer.Equal) {
		// Parse attribute value
		value, err := p.htmlAttributeValue()
		if err != nil {
			return ast.HTMLAttribute{}, err
		}

		return ast.HTMLAttribute{
			Name:  nameToken,
			Value: value,
			Span:  lexer.Span{Start: nameToken.Start(), End: value.GetSpan().End},
		}, nil
	}

	// Boolean attribute (no value)
	return ast.HTMLAttribute{
		Name:  nameToken,
		Value: nil,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}, nil
}

// htmlAttributeValue parses the value part of an HTML attribute
func (p *Parser) htmlAttributeValue() (ast.Expr, error) {
	// Handle string literal values
	if p.check(lexer.String) {
		stringToken := p.advance()
		return &ast.Literal{
			Token: stringToken,
			Value: stringToken.Literal,
			Type:  ast.LiteralTypeString,
			Span:  lexer.Span{Start: stringToken.Start(), End: stringToken.End()},
		}, nil
	}

	// Handle expression values {expression}
	if p.match(lexer.HTMLInterpolationStart) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(lexer.HTMLInterpolationEnd, "expected '}' after expression")
		if err != nil {
			return nil, err
		}
		return expr, nil
	}

	// Handle number literals
	if p.check(lexer.Number) {
		numberToken := p.advance()
		return &ast.Literal{
			Token: numberToken,
			Value: numberToken.Literal,
			Type:  ast.LiteralTypeNumber,
			Span:  lexer.Span{Start: numberToken.Start(), End: numberToken.End()},
		}, nil
	}

	// Handle boolean literals
	if p.check(lexer.True) || p.check(lexer.False) {
		boolToken := p.advance()
		return &ast.Literal{
			Token: boolToken,
			Value: boolToken.Literal,
			Type:  ast.LiteralTypeBool,
			Span:  lexer.Span{Start: boolToken.Start(), End: boolToken.End()},
		}, nil
	}

	// TODO: Handle interpolated strings like class="btn {variant} active"
	// This would require lexer support for STRING_PART tokens or
	// a different parsing approach for mixed string/interpolation content

	return nil, p.error(p.peek(), "expected string, number, boolean, or expression for attribute value")
}
