package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// importStatement parses an import statement.
func (p *Parser) importStatement() (ast.Stmt, error) {
	startToken := p.peek()

	// Check which type of import statement this is
	if p.check(lexer.Import) {
		return p.importName()
	} else if p.check(lexer.From) {
		return p.importFrom()
	}

	return nil, p.error(startToken, "expected 'import' or 'from'")
}

// importName handles: import_name: 'import' dotted_as_names
func (p *Parser) importName() (ast.Stmt, error) {
	// Consume the 'import' keyword
	importToken, err := p.consume(lexer.Import, "expected 'import'")
	if err != nil {
		return nil, err
	}

	// Parse dotted_as_names
	names, err := p.parseDottedAsNames()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, p.error(p.previous(), "expected module name after 'import'")
	}

	// Use the end position of the last name
	endPos := names[len(names)-1].GetSpan().End
	return &ast.ImportStmt{
		Names: names,
		Span:  lexer.Span{Start: importToken.Start(), End: endPos},
	}, nil
}

// importFrom handles:
// import_from:
//
//	| 'from' ('.' | '...')* dotted_name 'import' import_from_targets
//	| 'from' ('.' | '...')+ 'import' import_from_targets
func (p *Parser) importFrom() (ast.Stmt, error) {
	// Consume the 'from' keyword
	fromToken, err := p.consume(lexer.From, "expected 'from'")
	if err != nil {
		return nil, err
	}

	// Parse leading dots for relative imports
	dotCount := 0
	for p.match(lexer.Dot) || p.match(lexer.Ellipsis) {
		// Count '.' as 1, '...' as 3
		if p.previous().Type == lexer.Dot {
			dotCount += 1
		} else {
			dotCount += 3
		}
	}

	// Parse dotted_name if present (not present for relative-only imports)
	var dottedName *ast.DottedName = nil
	if dotCount == 0 || !p.check(lexer.Import) {
		dottedName, err = p.parseDottedName()
		if err != nil {
			return nil, err
		}
	}

	// Consume the 'import' keyword
	_, err = p.consume(lexer.Import, "expected 'import' after module name")
	if err != nil {
		return nil, err
	}

	// Parse import_from_targets
	isWildcard := false
	var names []*ast.ImportName

	// Check for '*'
	if p.match(lexer.Star) {
		isWildcard = true
		names = []*ast.ImportName{}
	} else if p.match(lexer.LeftParen) {
		// '(' import_from_as_names [','] ')'
		names, err = p.parseImportFromAsNames()
		if err != nil {
			return nil, err
		}

		// Allow optional trailing comma
		p.match(lexer.Comma)

		// Consume closing parenthesis
		_, err = p.consume(lexer.RightParen, "expected ')' after import names")
		if err != nil {
			return nil, err
		}
	} else {
		// import_from_as_names !','
		names, err = p.parseImportFromAsNames()
		if err != nil {
			return nil, err
		}
	}

	// Determine the end position
	var endPos lexer.Position
	if isWildcard {
		endPos = p.previous().End()
	} else if len(names) > 0 {
		endPos = names[len(names)-1].GetSpan().End
	} else {
		return nil, p.error(p.previous(), "expected import target after 'import'")
	}

	return &ast.ImportFromStmt{
		DottedName: dottedName,
		DotCount:   dotCount,
		Names:      names,
		IsWildcard: isWildcard,

		Span: lexer.Span{Start: fromToken.Start(), End: endPos},
	}, nil
}

// parseDottedAsNames handles: dotted_as_names: ','.dotted_as_name+
func (p *Parser) parseDottedAsNames() ([]*ast.ImportName, error) {
	var names []*ast.ImportName

	// Parse first dotted_as_name
	name, err := p.parseDottedAsName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	// Parse additional dotted_as_names separated by commas
	for p.match(lexer.Comma) {
		name, err = p.parseDottedAsName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

// parseDottedAsName handles: dotted_as_name: dotted_name ['as' NAME]
func (p *Parser) parseDottedAsName() (*ast.ImportName, error) {
	// Start position is the start of the dotted name
	startPos := p.peek().Start()

	// Parse the dotted_name
	dottedName, err := p.parseDottedName()
	if err != nil {
		return nil, err
	}

	// Check for optional 'as NAME'
	var asName *ast.Name = nil
	endPos := dottedName.GetSpan().End

	if p.match(lexer.As) {
		// Parse the alias name
		nameToken, err := p.consume(lexer.Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = &ast.Name{
			Token: nameToken,

			Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
		}
		endPos = asName.GetSpan().End
	}

	return &ast.ImportName{
		DottedName: dottedName,
		AsName:     asName,

		Span: lexer.Span{Start: startPos, End: endPos},
	}, nil
}

// parseDottedName handles:
// dotted_name:
//
//	| dotted_name '.' NAME
//	| NAME
func (p *Parser) parseDottedName() (*ast.DottedName, error) {
	startPos := p.peek().Start()

	// First name
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := &ast.Name{
		Token: nameToken,

		Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}
	names := []*ast.Name{name}
	endPos := name.GetSpan().End

	// Parse additional names with dots
	for p.match(lexer.Dot) {
		nameToken, err = p.consume(lexer.Identifier, "expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		name = &ast.Name{
			Token: nameToken,

			Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
		}
		names = append(names, name)
		endPos = name.GetSpan().End
	}

	return &ast.DottedName{
		Names: names,

		Span: lexer.Span{Start: startPos, End: endPos},
	}, nil
}

// parseImportFromAsNames handles: import_from_as_names: ','.import_from_as_name+
func (p *Parser) parseImportFromAsNames() ([]*ast.ImportName, error) {
	var names []*ast.ImportName

	// Parse first import_from_as_name
	name, err := p.parseImportFromAsName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	// Parse additional import_from_as_names separated by commas
	for p.match(lexer.Comma) {
		// If we see a closing parenthesis next, it's a trailing comma
		if p.check(lexer.RightParen) {
			break
		}

		name, err = p.parseImportFromAsName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

// parseImportFromAsName handles: import_from_as_name: NAME ['as' NAME]
func (p *Parser) parseImportFromAsName() (*ast.ImportName, error) {
	startPos := p.peek().Start()

	// Parse the first name
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := &ast.Name{
		Token: nameToken,

		Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}
	dottedName := &ast.DottedName{
		Names: []*ast.Name{name},
		Span:  lexer.Span{Start: name.GetSpan().Start, End: name.GetSpan().End},
	}

	// Check for optional 'as NAME'
	var asName *ast.Name = nil
	endPos := dottedName.GetSpan().End

	if p.match(lexer.As) {
		// Parse the alias name
		aliasToken, err := p.consume(lexer.Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = &ast.Name{
			Token: aliasToken,

			Span: lexer.Span{Start: aliasToken.Start(), End: aliasToken.End()},
		}
		endPos = asName.GetSpan().End
	}

	return &ast.ImportName{
		DottedName: dottedName,
		AsName:     asName,

		Span: lexer.Span{Start: startPos, End: endPos},
	}, nil
}
