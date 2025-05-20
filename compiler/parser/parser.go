package parser

import (
	"biscuit/compiler/ast/nodes"
	"biscuit/compiler/lexer"
)

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

// statement parses a single statement.
func (p *Parser) statement() (nodes.Stmt, error) {
	return p.simpleStatement()
}

// simpleStatement parses an expression statement.
func (p *Parser) simpleStatement() (nodes.Stmt, error) {
	// Check for keywords first
	switch p.peek().Type {
	case lexer.Type:
		return p.typeAlias()
	case lexer.Return:
		return p.returnStatement()
	case lexer.Import, lexer.From:
		return p.importStatement()
	case lexer.Raise:
		return p.raiseStatement()
	case lexer.Pass:
		return p.passStatement()
	case lexer.Del:
		return p.delStatement()
	case lexer.Yield:
		return p.yieldStatement()
	case lexer.Assert:
		return p.assertStatement()
	case lexer.Break:
		return p.breakStatement()
	case lexer.Continue:
		return p.continueStatement()
	case lexer.Global:
		return p.globalStatement()
	case lexer.Nonlocal:
		return p.nonlocalStatement()
	}

	// Check for assignment before expression
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) || p.check(lexer.Star) {
		// Save current position
		currentPos := p.Current

		// Try to parse as assignment
		stmt, err := p.assignment()
		if err == nil {
			return stmt, nil
		}

		// If assignment parsing failed, restore position and try as expression
		p.Current = currentPos
	}

	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	return nodes.NewExprStmt(expr, lexer.Span{Start: expr.Span().Start, End: expr.Span().End}), nil
}

// typeAlias parses a type alias statement as per the grammar:
// type_alias: "type" NAME [type_params] '=' expression
func (p *Parser) typeAlias() (nodes.Stmt, error) {
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
	var params []nodes.Expr = nil
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

	return nodes.NewTypeAlias(name, params, expr, lexer.Span{Start: typeToken.Start(), End: expr.Span().End}), nil
}

// typeParams parses type parameters as per the grammar:
// type_params: '[' type_param_seq ']'
// type_param_seq: ','.type_param+ [',']
func (p *Parser) typeParams() ([]nodes.Expr, error) {
	params := []nodes.Expr{}

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
func (p *Parser) typeParam() (nodes.Expr, error) {
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
	var bound nodes.Expr = nil
	if !isStar && !isDoubleStar && p.match(lexer.Colon) {
		bound, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional default
	var defaultValue nodes.Expr = nil
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
		endPos = defaultValue.Span().End
	} else if bound != nil {
		endPos = bound.Span().End
	}

	return nodes.NewTypeParamExpr(name, bound, defaultValue, isStar, isDoubleStar, lexer.Span{Start: startPos, End: endPos}), nil
}

func (p *Parser) returnStatement() (nodes.Stmt, error) {
	// Consume the 'return' keyword
	returnToken, err := p.consume(lexer.Return, "expected 'return'")
	if err != nil {
		return nil, err
	}

	// Exit early if there's no return expression
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) {
		p.advance()
		return nodes.NewReturnStmt(nil, lexer.Span{Start: returnToken.Start(), End: returnToken.End()}), nil
	}

	// Parse the return expression
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	return nodes.NewReturnStmt(expr, lexer.Span{Start: returnToken.Start(), End: expr.Span().End}), nil
}

func (p *Parser) importStatement() (nodes.Stmt, error) {
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
func (p *Parser) importName() (nodes.Stmt, error) {
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
	endPos := names[len(names)-1].Span().End
	return nodes.NewImportStmt(names, lexer.Span{Start: importToken.Start(), End: endPos}), nil
}

// importFrom handles:
// import_from:
//
//	| 'from' ('.' | '...')* dotted_name 'import' import_from_targets
//	| 'from' ('.' | '...')+ 'import' import_from_targets
func (p *Parser) importFrom() (nodes.Stmt, error) {
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
	var dottedName *nodes.DottedName = nil
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
	var names []*nodes.ImportName

	// Check for '*'
	if p.match(lexer.Star) {
		isWildcard = true
		names = []*nodes.ImportName{}
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
		endPos = names[len(names)-1].Span().End
	} else {
		return nil, p.error(p.previous(), "expected import target after 'import'")
	}

	return nodes.NewImportFromStmt(dottedName, dotCount, names, isWildcard, lexer.Span{Start: fromToken.Start(), End: endPos}), nil
}

// parseDottedAsNames handles: dotted_as_names: ','.dotted_as_name+
func (p *Parser) parseDottedAsNames() ([]*nodes.ImportName, error) {
	var names []*nodes.ImportName

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
func (p *Parser) parseDottedAsName() (*nodes.ImportName, error) {
	// Start position is the start of the dotted name
	startPos := p.peek().Start()

	// Parse the dotted_name
	dottedName, err := p.parseDottedName()
	if err != nil {
		return nil, err
	}

	// Check for optional 'as NAME'
	var asName *nodes.Name = nil
	endPos := dottedName.Span().End

	if p.match(lexer.As) {
		// Parse the alias name
		nameToken, err := p.consume(lexer.Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
		endPos = asName.Span().End
	}

	return nodes.NewImportName(dottedName, asName, lexer.Span{Start: startPos, End: endPos}), nil
}

// parseDottedName handles:
// dotted_name:
//
//	| dotted_name '.' NAME
//	| NAME
func (p *Parser) parseDottedName() (*nodes.DottedName, error) {
	startPos := p.peek().Start()

	// First name
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
	names := []*nodes.Name{name}
	endPos := name.Span().End

	// Parse additional names with dots
	for p.match(lexer.Dot) {
		nameToken, err = p.consume(lexer.Identifier, "expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		name = nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
		names = append(names, name)
		endPos = name.Span().End
	}

	return nodes.NewDottedName(names, lexer.Span{Start: startPos, End: endPos}), nil
}

// parseImportFromAsNames handles: import_from_as_names: ','.import_from_as_name+
func (p *Parser) parseImportFromAsNames() ([]*nodes.ImportName, error) {
	var names []*nodes.ImportName

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
func (p *Parser) parseImportFromAsName() (*nodes.ImportName, error) {
	startPos := p.peek().Start()

	// Parse the first name
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
	dottedName := nodes.NewDottedName([]*nodes.Name{name}, lexer.Span{Start: name.Span().Start, End: name.Span().End})

	// Check for optional 'as NAME'
	var asName *nodes.Name = nil
	endPos := dottedName.Span().End

	if p.match(lexer.As) {
		// Parse the alias name
		aliasToken, err := p.consume(lexer.Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = nodes.NewName(aliasToken, lexer.Span{Start: aliasToken.Start(), End: aliasToken.End()})
		endPos = asName.Span().End
	}

	return nodes.NewImportName(dottedName, asName, lexer.Span{Start: startPos, End: endPos}), nil
}

func (p *Parser) raiseStatement() (nodes.Stmt, error) {
	// Consume the 'raise' keyword
	raiseToken, err := p.consume(lexer.Raise, "expected 'raise'")
	if err != nil {
		return nil, err
	}

	// Check if there's an expression after 'raise'
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) {
		// Just a 'raise' with no exception
		endPos := raiseToken.End()
		return nodes.NewRaiseStmt(nil, nil, false, false, lexer.Span{Start: raiseToken.Start(), End: endPos}), nil
	}

	// Parse the exception expression
	exception, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check if there's a 'from' clause
	hasFrom := false
	var fromExpr nodes.Expr = nil
	endPos := exception.Span().End

	if p.match(lexer.From) {
		hasFrom = true
		// Parse the 'from' expression
		fromExpr, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = fromExpr.Span().End
	}

	return nodes.NewRaiseStmt(exception, fromExpr, true, hasFrom, lexer.Span{Start: raiseToken.Start(), End: endPos}), nil
}

func (p *Parser) passStatement() (nodes.Stmt, error) {
	// Consume the 'pass' keyword
	passToken := p.advance()
	return nodes.NewPassStmt(lexer.Span{Start: passToken.Start(), End: passToken.End()}), nil
}

func (p *Parser) delStatement() (nodes.Stmt, error) {
	// Consume the 'del' keyword
	delToken, err := p.consume(lexer.Del, "expected 'del'")
	if err != nil {
		return nil, err
	}

	// Parse the del_targets
	targets, err := p.delTargets()
	if err != nil {
		return nil, err
	}

	// TODO: Create and return a DelStmt node
	// For now, return a placeholder
	return nodes.NewExprStmt(targets, lexer.Span{Start: delToken.Start(), End: targets.Span().End}), nil
}

func (p *Parser) yieldStatement() (nodes.Stmt, error) {
	expr, err := p.yieldExpression()
	if err != nil {
		return nil, err
	}
	return nodes.NewYieldStmt(expr, lexer.Span{Start: expr.Span().Start, End: expr.Span().End}), nil
}

func (p *Parser) assertStatement() (nodes.Stmt, error) {
	// Consume the 'assert' keyword
	assertToken, err := p.consume(lexer.Assert, "expected 'assert'")
	if err != nil {
		return nil, err
	}

	// Parse the test expression
	test, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check for optional message expression
	var message nodes.Expr = nil
	endPos := test.Span().End

	if p.match(lexer.Comma) {
		// Parse the message expression
		message, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = message.Span().End
	}

	return nodes.NewAssertStmt(test, message, lexer.Span{Start: assertToken.Start(), End: endPos}), nil
}

func (p *Parser) breakStatement() (nodes.Stmt, error) {
	// Consume the 'break' keyword
	breakToken := p.advance()
	return nodes.NewBreakStmt(lexer.Span{Start: breakToken.Start(), End: breakToken.End()}), nil
}

func (p *Parser) continueStatement() (nodes.Stmt, error) {
	// Consume the 'continue' keyword
	continueToken := p.advance()
	return nodes.NewContinueStmt(lexer.Span{Start: continueToken.Start(), End: continueToken.End()}), nil
}

func (p *Parser) globalStatement() (nodes.Stmt, error) {
	// Consume the 'global' keyword
	globalToken, err := p.consume(lexer.Global, "expected 'global'")
	if err != nil {
		return nil, err
	}

	// Parse at least one identifier
	names, err := p.parseNameList()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, p.error(p.previous(), "expected at least one identifier after 'global'")
	}

	// Get the end position from the last name
	endPos := names[len(names)-1].Span().End
	return nodes.NewGlobalStmt(names, lexer.Span{Start: globalToken.Start(), End: endPos}), nil
}

func (p *Parser) nonlocalStatement() (nodes.Stmt, error) {
	// Consume the 'nonlocal' keyword
	nonlocalToken, err := p.consume(lexer.Nonlocal, "expected 'nonlocal'")
	if err != nil {
		return nil, err
	}

	// Parse at least one identifier
	names, err := p.parseNameList()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, p.error(p.previous(), "expected at least one identifier after 'nonlocal'")
	}

	// Get the end position from the last name
	endPos := names[len(names)-1].Span().End
	return nodes.NewNonlocalStmt(names, lexer.Span{Start: nonlocalToken.Start(), End: endPos}), nil
}

// parseNameList is a helper function to parse a comma-separated list of identifiers
func (p *Parser) parseNameList() ([]*nodes.Name, error) {
	names := []*nodes.Name{}

	// Parse first identifier
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	// Create a Name node from the token
	name := nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
	names = append(names, name)

	// Parse additional identifiers separated by commas
	for p.match(lexer.Comma) {
		nameToken, err = p.consume(lexer.Identifier, "expected identifier after ','")
		if err != nil {
			return nil, err
		}

		// Create a Name node from the token
		name = nodes.NewName(nameToken, lexer.Span{Start: nameToken.Start(), End: nameToken.End()})
		names = append(names, name)
	}

	return names, nil
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

// expression parses an expression.
func (p *Parser) expression() (nodes.Expr, error) {
	expr, err := p.disjunction()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.If) {
		condition, err := p.disjunction()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(lexer.Else, "expected 'else' after condition")
		if err != nil {
			return nil, err
		}
		elseExpr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return nodes.NewTernaryExpr(condition, expr, elseExpr, lexer.Span{Start: expr.Span().Start, End: elseExpr.Span().End}), nil
	}

	return expr, nil
}

// disjunction parses a disjunction expression.
func (p *Parser) disjunction() (nodes.Expr, error) {
	expr, err := p.conjunction()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Or) {
		operator := p.previous()
		right, err := p.conjunction()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// conjunction parses a conjunction expression.
func (p *Parser) conjunction() (nodes.Expr, error) {
	expr, err := p.inversion()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.And) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// inversion parses an inversion expression.
func (p *Parser) inversion() (nodes.Expr, error) {
	if p.match(lexer.Not) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		return nodes.NewUnary(operator, right, lexer.Span{Start: operator.Start(), End: right.Span().End}), nil
	}

	return p.comparison()
}

// comparison parses a comparison expression.
func (p *Parser) comparison() (nodes.Expr, error) {
	left, err := p.bitwiseOr()
	if err != nil {
		return nil, err
	}

	// First check for any comparison operator to start a comparison chain
	if isComparisonOperator(p.peek().Type) {
		// We're going to build a chain of comparisons
		operands := []nodes.Expr{left}
		operators := []lexer.Token{}

		// Keep consuming comparison operators and their right operands
		for isComparisonOperator(p.peek().Type) {
			// Regular comparison operator - now includes IsNot and NotIn
			p.advance()
			operators = append(operators, p.previous())

			// Parse the right operand
			right, err := p.bitwiseOr()
			if err != nil {
				return nil, err
			}
			operands = append(operands, right)
		}

		// Handle a single comparison (most common case)
		if len(operands) == 2 {
			return nodes.NewBinary(operands[0], operators[0], operands[1], lexer.Span{Start: operands[0].Span().Start, End: operands[1].Span().End}), nil
		}

		// Handle chained comparisons (a < b < c becomes (a < b) and (b < c))
		var result nodes.Expr
		for i := 0; i < len(operators); i++ {
			comparison := nodes.NewBinary(operands[i], operators[i], operands[i+1],
				lexer.Span{Start: operands[i].Span().Start, End: operands[i+1].Span().End})

			if i == 0 {
				result = comparison
			} else {
				// Create an AND expression linking the comparisons
				andToken := lexer.Token{Type: lexer.And, Lexeme: "and"}
				result = nodes.NewBinary(result, andToken, comparison,
					lexer.Span{Start: result.Span().Start, End: comparison.Span().End})
			}
		}
		return result, nil
	}

	return left, nil
}

// Helper function to check if a token type is a comparison operator
func isComparisonOperator(tokenType lexer.TokenType) bool {
	return tokenType == lexer.EqualEqual || tokenType == lexer.BangEqual ||
		tokenType == lexer.Less || tokenType == lexer.LessEqual ||
		tokenType == lexer.Greater || tokenType == lexer.GreaterEqual ||
		tokenType == lexer.In || tokenType == lexer.Is || tokenType == lexer.IsNot || tokenType == lexer.NotIn
}

// bitwise_or parses a bitwise OR expression.
func (p *Parser) bitwiseOr() (nodes.Expr, error) {
	expr, err := p.bitwiseXor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Pipe) {
		operator := p.previous()
		right, err := p.bitwiseXor()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// bitwiseXor parses a bitwise XOR expression.
func (p *Parser) bitwiseXor() (nodes.Expr, error) {
	expr, err := p.bitwiseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Caret) {
		operator := p.previous()
		right, err := p.bitwiseAnd()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// bitwiseAnd parses a bitwise AND expression.
func (p *Parser) bitwiseAnd() (nodes.Expr, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Ampersand) {
		operator := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// shiftExpr parses a shift expression.
func (p *Parser) shiftExpr() (nodes.Expr, error) {
	expr, err := p.sum()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.LessLess, lexer.GreaterGreater) {
		operator := p.previous()
		right, err := p.sum()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// sum parses a sum expression.
func (p *Parser) sum() (nodes.Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Plus, lexer.Minus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// term parses a term expression.
func (p *Parser) term() (nodes.Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Star, lexer.Slash, lexer.SlashSlash, lexer.Percent, lexer.At) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// factor parses a factor expression.
func (p *Parser) factor() (nodes.Expr, error) {
	if p.match(lexer.Plus, lexer.Minus, lexer.Tilde) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return nodes.NewUnary(operator, right, lexer.Span{Start: operator.Start(), End: right.Span().End}), nil
	}

	return p.power()
}

// power parses a power expression.
func (p *Parser) power() (nodes.Expr, error) {
	expr, err := p.await()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.StarStar) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return nodes.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End}), nil
	}

	return expr, nil
}

func (p *Parser) await() (nodes.Expr, error) {
	// Check if the current token is an await keyword
	if p.match(lexer.Await) {
		awaitToken := p.previous()
		expr, err := p.primary()
		if err != nil {
			return nil, err
		}
		return nodes.NewAwaitExpr(expr, lexer.Span{Start: awaitToken.Start(), End: expr.Span().End}), nil
	}

	return p.primary()
}

// primary parses a primary expression.
func (p *Parser) primary() (nodes.Expr, error) {
	// Parse the initial atom
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Keep parsing postfix operations while they exist
	for {
		if p.match(lexer.Dot) {
			// Handle attribute access: expr.NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = nodes.NewAttribute(expr, name, lexer.Span{Start: expr.Span().Start, End: name.End()})
		} else if p.match(lexer.LeftParen) {
			// Handle function call: expr(args)
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(lexer.LeftBracket) {
			// Handle subscript access: expr[index] or expr[slice]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = nodes.NewSubscript(expr, indices, lexer.Span{Start: expr.Span().Start, End: right.End()})
		} else {
			// No more postfix operations
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee nodes.Expr) (nodes.Expr, error) {
	args := []nodes.Expr{}
	if !p.check(lexer.RightParen) {
		// Parse first argument
		arg, err := p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional arguments
		for p.match(lexer.Comma) {
			// If there's a right parenthesis after the comma,
			// it's a trailing comma, so we're done parsing arguments
			if p.check(lexer.RightParen) {
				break
			}

			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	right, err := p.consume(lexer.RightParen, "expected ')' after arguments")
	if err != nil {
		return nil, err
	}

	return nodes.NewCall(callee, args, lexer.Span{Start: callee.Span().Start, End: right.End()}), nil
}

// atom parses an atom.
func (p *Parser) atom() (nodes.Expr, error) {
	if p.match(lexer.False) {
		return nodes.NewLiteral(p.previous(), false, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.True) {
		return nodes.NewLiteral(p.previous(), true, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.None) {
		return nodes.NewLiteral(p.previous(), nil, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Number, lexer.String) {
		return nodes.NewLiteral(p.previous(), p.previous().Literal, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Ellipsis) {
		return nodes.NewLiteral(p.previous(), nil, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Identifier) {
		return nodes.NewName(p.previous(), lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.check(lexer.LeftParen) {
		// This could be either a tuple or a group (parenthesized expression)
		return p.tuple()
	}

	if p.check(lexer.LeftBracket) {
		// List literal
		return p.list()
	}

	if p.check(lexer.LeftBrace) {
		// Set literal
		return p.set()
	}

	if p.check(lexer.Yield) {
		// Yield expression
		return p.yieldExpression()
	}

	return nil, p.error(p.peek(), "unexpected token")
}

func (p *Parser) list() (nodes.Expr, error) {
	// Expect opening bracket
	leftBracket, err := p.consume(lexer.LeftBracket, "expected '['")
	if err != nil {
		return nil, err
	}

	elements := []nodes.Expr{}

	// Parse elements if the list is not empty
	if !p.check(lexer.RightBracket) {
		// Parse star named expressions
		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)

		// Parse additional elements separated by commas
		for p.match(lexer.Comma) {
			// Allow trailing comma
			if p.check(lexer.RightBracket) {
				break
			}

			expr, err := p.starNamedExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)
		}
	}

	// Expect closing bracket
	rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
	if err != nil {
		return nil, err
	}

	return nodes.NewListExpr(elements, lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()}), nil
}

func (p *Parser) tuple() (nodes.Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(lexer.LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Empty tuple
	if p.check(lexer.RightParen) {
		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return nodes.NewTupleExpr([]nodes.Expr{}, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
	}

	// Parse the first expression (could be a star expression or named expression)
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}

	// If the next token is 'yield', it's a group
	if p.checkNext(lexer.Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return nodes.NewGroupExpr(expr, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
	}

	// If there's a comma, it's a tuple
	if p.match(lexer.Comma) {
		elements := []nodes.Expr{expr}

		// Parse additional elements if present
		if !p.check(lexer.RightParen) {
			for {
				nextExpr, err := p.starNamedExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, nextExpr)

				if !p.match(lexer.Comma) {
					break
				}

				// Allow trailing comma
				if p.check(lexer.RightParen) {
					break
				}
			}
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return nodes.NewTupleExpr(elements, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
	} else {
		// No comma, so it's a group
		// Groups can only contain named expressions, not star expressions
		_, isStarExpr := expr.(*nodes.StarExpr)
		if isStarExpr {
			return nil, p.error(p.previous(), "starred expression cannot appear in a group")
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return nodes.NewGroupExpr(expr, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
	}
}

func (p *Parser) set() (nodes.Expr, error) {
	// Expect opening brace
	leftBrace, err := p.consume(lexer.LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	elements := []nodes.Expr{}

	// Parse elements
	// First element
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}
	elements = append(elements, expr)

	// Parse additional elements separated by commas
	for p.match(lexer.Comma) {
		// Allow trailing comma
		if p.check(lexer.RightBrace) {
			break
		}

		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	// Expect closing brace
	rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
	if err != nil {
		return nil, err
	}

	return nodes.NewSetExpr(elements, lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()}), nil
}

func (p *Parser) yieldExpression() (nodes.Expr, error) {
	// Expect 'yield' keyword
	yieldToken, err := p.consume(lexer.Yield, "expected 'yield'")
	if err != nil {
		return nil, err
	}

	// Check for 'yield from' form
	isFrom := false
	if p.match(lexer.From) {
		isFrom = true
		// Parse the expression after 'yield from'
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return nodes.NewYieldExpr(isFrom, expr, lexer.Span{Start: yieldToken.Start(), End: expr.Span().End}), nil
	}

	// Check if there's an expression after 'yield'
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) || p.check(lexer.RightParen) || p.check(lexer.Comma) {
		// No expression, yield on its own
		return nodes.NewYieldExpr(false, nil, lexer.Span{Start: yieldToken.Start(), End: yieldToken.End()}), nil
	}

	// Parse star expressions after 'yield'
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}
	return nodes.NewYieldExpr(false, expr, lexer.Span{Start: yieldToken.Start(), End: expr.Span().End}), nil
}

func (p *Parser) starExpressions() (nodes.Expr, error) {
	// Parse the first star expression
	expr, err := p.starExpression()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the expression as is
	if !p.match(lexer.Comma) {
		return expr, nil
	}

	// We have a comma, so this is a tuple of expressions
	elements := []nodes.Expr{expr}

	// Allow trailing comma with no following expression
	if !p.check(lexer.Newline) && !p.check(lexer.RightParen) && !p.check(lexer.RightBracket) && !p.check(lexer.RightBrace) && !p.check(lexer.Semicolon) && !p.isAtEnd() {
		// Parse subsequent expressions
		for {
			expr, err := p.starExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.Newline) || p.check(lexer.RightParen) || p.check(lexer.RightBracket) || p.check(lexer.RightBrace) || p.check(lexer.Semicolon) || p.isAtEnd() {
				break
			}
		}
	}

	// Create a tuple with the collected expressions
	return nodes.NewTupleExpr(elements, lexer.Span{Start: elements[0].Span().Start, End: elements[len(elements)-1].Span().End}), nil
}

func (p *Parser) group() (nodes.Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(lexer.LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Check for yield expression
	if p.check(lexer.Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		// Expect closing parenthesis
		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}

		return nodes.NewGroupExpr(expr, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
	}

	// Parse named expression
	expr, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect closing parenthesis
	rightParen, err := p.consume(lexer.RightParen, "expected ')'")
	if err != nil {
		return nil, err
	}

	return nodes.NewGroupExpr(expr, lexer.Span{Start: leftParen.Start(), End: rightParen.End()}), nil
}

func (p *Parser) starExpression() (nodes.Expr, error) {
	if p.match(lexer.Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return nodes.NewStarExpr(expr, lexer.Span{Start: star.Start(), End: expr.Span().End}), nil
	}

	// Not a star expression, parse as a regular expression
	return p.expression()
}

func (p *Parser) namedExpression() (nodes.Expr, error) {
	// Handle assignment expressions first (the walrus operator)
	if p.check(lexer.Identifier) && p.checkNext(lexer.Walrus) {
		name, err := p.consume(lexer.Identifier, "expected identifier") // Consume the identifier
		if err != nil {
			return nil, err
		}
		_, err = p.consume(lexer.Walrus, "expected ':=' after identifier") // Consume the ':=' operator
		if err != nil {
			return nil, err
		}

		value, err := p.expression()
		if err != nil {
			return nil, err
		}

		return nodes.NewAssignExpr(
			nodes.NewName(name, lexer.Span{Start: name.Start(), End: name.End()}),
			value,
			lexer.Span{Start: name.Start(), End: value.Span().End},
		), nil
	}

	// If not an assignment expression, parse a regular expression
	return p.expression()
}

func (p *Parser) starNamedExpression() (nodes.Expr, error) {
	if p.match(lexer.Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return nodes.NewStarExpr(expr, lexer.Span{Start: star.Start(), End: expr.Span().End}), nil
	}

	// Not a star expression, parse as a regular expression
	return p.namedExpression()
}

// slice parses a single slice element as per the grammar:
// slice:
//
//	| [expression] ':' [expression] [':' [expression] ]
//	| named_expression
func (p *Parser) slice() (nodes.Expr, error) {
	// Check if this is a slice notation or just an expression
	// We need to look ahead to see if there's a colon after the first expression (if any)
	startPos := p.peek().Start()

	// Empty slice is allowed (:)
	if p.check(lexer.Colon) {
		// No start expression, consume the colon
		p.advance()

		var end nodes.Expr
		var err error

		// Check for end expression after colon
		if !p.check(lexer.Colon) && !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
			end, err = p.expression()
			if err != nil {
				return nil, err
			}
		}

		// Check for step (second colon)
		var step nodes.Expr
		if p.match(lexer.Colon) {
			// Parse optional step
			if !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
				step, err = p.expression()
				if err != nil {
					return nil, err
				}
			}
		}

		endPos := p.previous().End()
		if step != nil {
			endPos = step.Span().End
		} else if end != nil {
			endPos = end.Span().End
		}

		return nodes.NewSlice(nil, end, step, lexer.Span{Start: startPos, End: endPos}), nil
	}

	// There's an expression before any potential colon
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	// If there's no colon after the expression, this is just a regular index
	if !p.match(lexer.Colon) {
		return expr, nil
	}

	// We have a slice with a start expression
	var end nodes.Expr

	// Check for end expression after colon
	if !p.check(lexer.Colon) && !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
		end, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Check for step (second colon)
	var step nodes.Expr
	if p.match(lexer.Colon) {
		// Parse optional step
		if !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
			step, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
	}

	endPos := p.previous().End()
	if step != nil {
		endPos = step.Span().End
	} else if end != nil {
		endPos = end.Span().End
	} else {
		endPos = expr.Span().End
	}

	return nodes.NewSlice(expr, end, step, lexer.Span{Start: startPos, End: endPos}), nil
}

// slices parses one or more slice elements as per the grammar:
// slices:
//
//	| slice !','
//	| ','.(slice | starred_expression)+ [',']
func (p *Parser) slices() ([]nodes.Expr, error) {
	// Parse the first slice
	first, err := p.slice()
	if err != nil {
		return nil, err
	}

	// Initialize the slice with the first element
	indices := []nodes.Expr{first}

	// If there's no comma, it's just a single index/slice
	if !p.match(lexer.Comma) {
		return indices, nil
	}

	// Parse additional slice or starred_expression elements
	for !p.check(lexer.RightBracket) {
		// Handle starred expressions
		if p.match(lexer.Star) {
			// This is a starred expression like *args
			star := p.previous()
			expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
			if err != nil {
				return nil, err
			}
			indices = append(indices, nodes.NewStarExpr(expr, lexer.Span{Start: star.Start(), End: expr.Span().End}))
		} else {
			// Regular slice expression
			expr, err := p.slice()
			if err != nil {
				return nil, err
			}
			indices = append(indices, expr)
		}

		// Break if no more commas
		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if p.check(lexer.RightBracket) {
			break
		}
	}

	return indices, nil
}

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

func (p *Parser) consume(t lexer.TokenType, message string) (lexer.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return lexer.Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(token lexer.Token, message string) error {
	return &ParseError{Token: token, Message: message}
}

// match checks if the current token is one of the given types.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) checkNext(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekN(1).Type == t
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.Current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.EOF
}

func (p *Parser) peek() lexer.Token {
	return p.Tokens[p.Current]
}

func (p *Parser) peekN(n int) lexer.Token {
	return p.Tokens[p.Current+n]
}

func (p *Parser) previous() lexer.Token {
	return p.Tokens[p.Current-1]
}

// tPrimary parses a primary expression that must be followed by an accessor.
// This is used for parsing chained targets in assignment contexts.
// Implementation of the grammar rule:
//
// t_primary:
//
//	| t_primary '.' NAME &t_lookahead
//	| t_primary '[' slices ']' &t_lookahead
//	| t_primary genexp &t_lookahead
//	| t_primary '(' [arguments] ')' &t_lookahead
//	| atom &t_lookahead
func (p *Parser) tPrimary() (nodes.Expr, error) {
	// Start with the base case: atom &t_lookahead
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Check if there's a lookahead token - required for all t_primary rules
	// This implements the &t_lookahead part of the grammar
	if !p.tLookahead() {
		return nil, p.error(p.peek(), "expected accessor token ('.', '[', or '(')")
	}

	// Iteratively handle the recursive cases for t_primary
	// This transforms the left-recursive grammar into an iterative implementation
	for {
		// Save the current position in case we need to restore it
		originalPosition := p.Current

		if p.match(lexer.Dot) {
			// Rule: t_primary '.' NAME &t_lookahead
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = nodes.NewAttribute(expr, name, lexer.Span{Start: expr.Span().Start, End: name.End()})
		} else if p.match(lexer.LeftParen) {
			// Rule: t_primary '(' [arguments] ')' &t_lookahead
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}

			// TODO: check for genexp if call fails
		} else if p.match(lexer.LeftBracket) {
			// Rule: t_primary '[' slices ']' &t_lookahead
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = nodes.NewSubscript(expr, indices, lexer.Span{Start: expr.Span().Start, End: right.End()})
		} else {
			// If we didn't consume any accessor, we're done
			// Either we have atom &t_lookahead or we've finished a chain
			break
		}

		// Each accessor must be followed by another lookahead token
		// This implements the &t_lookahead at the end of each recursive rule
		if !p.tLookahead() {
			// If not, we've been too greedy and consumed too many tokens
			// Restore the position and return what we had before
			p.Current = originalPosition
			break
		}
	}

	return expr, nil
}

// tLookahead checks if the current token is one of the t_lookahead tokens: '(', '[', '.'
func (p *Parser) tLookahead() bool {
	if p.isAtEnd() {
		return false
	}
	tokenType := p.peek().Type
	return tokenType == lexer.LeftParen || tokenType == lexer.LeftBracket || tokenType == lexer.Dot
}

// singleSubscriptAttributeTarget parses a single target with attribute or subscription
// as per the grammar:
//
// single_subscript_attribute_target:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
func (p *Parser) singleSubscriptAttributeTarget() (nodes.Expr, error) {
	// Parse the t_primary expression
	expr, err := p.tPrimary()
	if err != nil {
		return nil, err
	}

	// Check which form it is
	if p.match(lexer.Dot) {
		// Handle attribute access: t_primary.NAME
		name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		result := nodes.NewAttribute(expr, name, lexer.Span{Start: expr.Span().Start, End: name.End()})

		// Check negative lookahead - must NOT be followed by another accessor
		if p.tLookahead() {
			return nil, p.error(p.peek(), "unexpected accessor after attribute target")
		}

		return result, nil
	} else if p.match(lexer.LeftBracket) {
		// Handle subscript access: t_primary[slices]
		indices, err := p.slices()
		if err != nil {
			return nil, err
		}

		right, err := p.consume(lexer.RightBracket, "expected ']' after index")
		if err != nil {
			return nil, err
		}
		result := nodes.NewSubscript(expr, indices, lexer.Span{Start: expr.Span().Start, End: right.End()})

		// Check negative lookahead - must NOT be followed by another accessor
		if p.tLookahead() {
			return nil, p.error(p.peek(), "unexpected accessor after subscript target")
		}

		return result, nil
	}

	return nil, p.error(p.peek(), "expected '.' or '[' after primary expression")
}

// singleTarget parses a single target as per the grammar:
//
// single_target:
//
//	| single_subscript_attribute_target
//	| NAME
//	| '(' single_target ')'
func (p *Parser) singleTarget() (nodes.Expr, error) {
	if p.check(lexer.Identifier) {
		// Handle the NAME case first
		// But first check if it might be a single_subscript_attribute_target
		// by seeing if there's a lookahead accessor after the identifier
		if p.checkNext(lexer.Dot) || p.checkNext(lexer.LeftBracket) || p.checkNext(lexer.LeftParen) {
			return p.singleSubscriptAttributeTarget()
		}

		// Just a NAME
		name := p.advance()
		return nodes.NewName(name, lexer.Span{Start: name.Start(), End: name.End()}), nil
	} else if p.match(lexer.LeftParen) {
		// Handle parenthesized form: '(' single_target ')'
		target, err := p.singleTarget()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(lexer.RightParen, "expected ')' after target")
		if err != nil {
			return nil, err
		}

		return nodes.NewGroupExpr(target, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	// Try to parse as single_subscript_attribute_target
	return p.singleSubscriptAttributeTarget()
}

// starTarget parses a star target as per the grammar:
//
// star_target:
//
//	| '*' (!'*' star_target)
//	| target_with_star_atom
func (p *Parser) starTarget() (nodes.Expr, error) {
	if p.match(lexer.Star) {
		// Handle starred expression
		star := p.previous()

		// Check not followed by another star
		if p.check(lexer.Star) {
			return nil, p.error(p.peek(), "cannot use ** in target expressions")
		}

		// Parse the inner star_target
		expr, err := p.starTarget()
		if err != nil {
			return nil, err
		}

		return nodes.NewStarExpr(expr, lexer.Span{Start: star.Start(), End: expr.Span().End}), nil
	}

	// Not a starred expression, parse as target_with_star_atom
	return p.targetWithStarAtom()
}

// starTargets parses star targets as per the grammar:
//
// star_targets:
//
//	| star_target !','
//	| star_target (',' star_target)* [',']
func (p *Parser) starTargets() ([]nodes.Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	targets := []nodes.Expr{target}

	// If there's no comma, return the single target
	if !p.match(lexer.Comma) {
		return targets, nil
	}

	// We've consumed a comma
	// Check if we're at the end (trailing comma case)
	if p.check(lexer.RightParen) || p.check(lexer.RightBracket) ||
		p.check(lexer.Colon) || p.check(lexer.Equal) || p.check(lexer.Newline) ||
		p.check(lexer.Semicolon) || p.isAtEnd() {
		// Just a trailing comma, we're done
		return targets, nil
	}

	// Keep parsing targets and commas
	for {
		target, err = p.starTarget()
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)

		// If no more commas, we're done
		if !p.match(lexer.Comma) {
			break
		}

		// If we've found a trailing comma, we're done
		if p.check(lexer.RightParen) || p.check(lexer.RightBracket) ||
			p.check(lexer.Colon) || p.check(lexer.Equal) || p.check(lexer.Newline) ||
			p.check(lexer.Semicolon) || p.isAtEnd() {
			break
		}
	}

	return targets, nil
}

// parseStarTargetSequence parses a sequence of star targets for either list or tuple contexts
// isTuple indicates whether tuple rules should be enforced (requiring comma for single element)
// closingToken specifies the token that would terminate the sequence (RightParen or RightBracket)
func (p *Parser) parseStarTargetSequence(isTuple bool, closingToken lexer.TokenType) ([]nodes.Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	elements := []nodes.Expr{target}

	// For tuples, a comma is required
	if isTuple && !p.match(lexer.Comma) {
		return nil, p.error(p.peek(), "expected ',' after target in tuple")
	} else if !isTuple {
		// For lists, comma is optional
		if !p.match(lexer.Comma) {
			return elements, nil
		}
	}

	// Check if it's just a single-element tuple (star_target ',')
	if isTuple && p.check(closingToken) {
		return elements, nil
	}

	// If we don't have another target after comma for a list, we're done
	if !isTuple && p.check(closingToken) {
		return elements, nil
	}

	// Parse additional targets if there are any
	if !p.check(closingToken) {
		// For tuples, we need at least a second element if not single-element
		if isTuple {
			target, err = p.starTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			// Parse more elements with commas
			for p.match(lexer.Comma) {
				// Check for trailing comma
				if p.check(closingToken) {
					break
				}

				target, err = p.starTarget()
				if err != nil {
					return nil, err
				}
				elements = append(elements, target)
			}
		} else {
			// For lists, parse additional elements while there are commas
			for {
				target, err = p.starTarget()
				if err != nil {
					return nil, err
				}
				elements = append(elements, target)

				if !p.match(lexer.Comma) {
					break
				}

				// Check for trailing comma
				if p.check(closingToken) {
					break
				}
			}
		}
	}

	return elements, nil
}

// starAtom parses a star atom as per the grammar:
//
// star_atom:
//
//	| NAME
//	| '(' target_with_star_atom ')'
//	| '(' [star_targets_tuple_seq] ')'
//	| '[' [star_targets_list_seq] ']'
func (p *Parser) starAtom() (nodes.Expr, error) {
	startPos := p.peek().Start()

	if p.match(lexer.Identifier) {
		// Handle simple NAME case
		name := p.previous()
		return nodes.NewName(name, lexer.Span{Start: startPos, End: name.End()}), nil
	} else if p.match(lexer.LeftParen) {
		// Handle parenthesized forms
		if p.match(lexer.RightParen) {
			// Empty tuple
			return nodes.NewTupleExpr([]nodes.Expr{}, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

		// Try to parse as target_with_star_atom first
		nextPos := p.Current
		target, err := p.targetWithStarAtom()

		if err == nil {
			// Successfully parsed as target_with_star_atom
			// Consume the closing parenthesis
			_, err = p.consume(lexer.RightParen, "expected ')' after target")
			if err != nil {
				goto tryStartTargetSequence
			}
			return nodes.NewGroupExpr(target, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

	tryStartTargetSequence:
		// Restore position and try as star_targets_tuple_seq
		p.Current = nextPos

		elements, err := p.parseStarTargetSequence(true, lexer.RightParen)
		if err != nil {
			return nil, err
		}

		// Consume the closing parenthesis
		_, err = p.consume(lexer.RightParen, "expected ')' after tuple targets")
		if err != nil {
			return nil, err
		}

		return nodes.NewTupleExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
	} else if p.match(lexer.LeftBracket) {
		// Handle list form
		if p.match(lexer.RightBracket) {
			// Empty list
			return nodes.NewListExpr([]nodes.Expr{}, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

		elements, err := p.parseStarTargetSequence(false, lexer.RightBracket)
		if err != nil {
			return nil, err
		}

		// Consume the closing bracket
		_, err = p.consume(lexer.RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return nodes.NewListExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
	}

	return nil, p.error(p.peek(), "expected NAME, '(' or '[' in star atom")
}

// targetWithStarAtom parses a target with star atom as per the grammar:
//
// target_with_star_atom:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
//	| star_atom
func (p *Parser) targetWithStarAtom() (nodes.Expr, error) {
	// Try to parse as t_primary if the next token could start a t_primary
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) ||
		p.check(lexer.False) || p.check(lexer.True) || p.check(lexer.None) ||
		p.check(lexer.Number) || p.check(lexer.String) || p.check(lexer.Ellipsis) {

		// First, save the current position
		startPos := p.Current

		// Try to parse a t_primary followed by '.'
		primary, err := p.tPrimary()
		if err != nil {
			// We failed to parse as t_primary, so try as star_atom
			// Restore the position and try again
			p.Current = startPos
			goto tryStarAtom
		}

		if p.match(lexer.Dot) {
			// Handle attribute access: t_primary.NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after attribute target")
			}

			return nodes.NewAttribute(primary, name, lexer.Span{Start: primary.Span().Start, End: name.End()}), nil
		}

		// Restore position and try t_primary followed by '['
		if p.match(lexer.LeftBracket) {
			// Handle subscript access: t_primary[slices]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after subscript target")
			}

			return nodes.NewSubscript(primary, indices, lexer.Span{Start: primary.Span().Start, End: right.End()}), nil
		}

		// Reset position if we couldn't match t_primary with an accessor
		p.Current = startPos
	}

tryStarAtom:
	// If we couldn't parse as t_primary with an accessor, try as star_atom
	return p.starAtom()
}

// delTarget parses a target for the del statement as per the grammar:
// del_target:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
//	| del_t_atom
func (p *Parser) delTarget() (nodes.Expr, error) {
	// Try to parse as t_primary if the next token could start a t_primary
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) ||
		p.check(lexer.False) || p.check(lexer.True) || p.check(lexer.None) ||
		p.check(lexer.Number) || p.check(lexer.String) || p.check(lexer.Ellipsis) {

		// First, save the current position
		startPos := p.Current

		// Try to parse a t_primary followed by '.'
		primary, err := p.tPrimary()
		if err == nil && p.match(lexer.Dot) {
			// Handle attribute access: t_primary.NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			result := nodes.NewAttribute(primary, name, lexer.Span{Start: primary.Span().Start, End: name.End()})

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after attribute in del target")
			}

			return result, nil
		}

		// Restore position and try t_primary followed by '['
		p.Current = startPos
		primary, err = p.tPrimary()
		if err == nil && p.match(lexer.LeftBracket) {
			// Handle subscript access: t_primary[slices]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			result := nodes.NewSubscript(primary, indices, lexer.Span{Start: primary.Span().Start, End: right.End()})

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after subscript in del target")
			}

			return result, nil
		}

		// Reset position if we couldn't match t_primary with an accessor
		p.Current = startPos
	}

	// If we couldn't parse as t_primary with an accessor, try as del_t_atom
	return p.delTAtom()
}

// delTAtom parses a del_t_atom as per the grammar:
// del_t_atom:
//
//	| NAME
//	| '(' del_target ')'
//	| '(' [del_targets] ')'
//	| '[' [del_targets] ']'
func (p *Parser) delTAtom() (nodes.Expr, error) {
	startPos := p.peek().Start()

	if p.check(lexer.Identifier) {
		// Handle simple NAME case
		name := p.advance()
		return nodes.NewName(name, lexer.Span{Start: startPos, End: name.End()}), nil
	} else if p.match(lexer.LeftParen) {
		// Handle parenthesized forms
		if p.match(lexer.RightParen) {
			// Empty tuple
			return nodes.NewTupleExpr([]nodes.Expr{}, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

		// Try to parse as single del_target first
		nextPos := p.Current
		target, err := p.delTarget()

		if err == nil {
			// Check if there's a comma after, which means it's a tuple
			if p.match(lexer.Comma) {
				// Start a tuple with the first target
				elements := []nodes.Expr{target}

				// Check for empty rest of tuple
				if p.match(lexer.RightParen) {
					return nodes.NewTupleExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
				}

				// Parse rest of del_targets
				for !p.check(lexer.RightParen) {
					target, err = p.delTarget()
					if err != nil {
						return nil, err
					}
					elements = append(elements, target)

					if !p.match(lexer.Comma) {
						break
					}
				}

				// Consume closing parenthesis
				_, err = p.consume(lexer.RightParen, "expected ')' after del targets")
				if err != nil {
					return nil, err
				}

				return nodes.NewTupleExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
			}

			// No comma, so it's a grouped expression
			_, err = p.consume(lexer.RightParen, "expected ')' after target")
			if err != nil {
				return nil, err
			}
			return nodes.NewGroupExpr(target, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

		// Restore position and try as del_targets (tuple)
		p.Current = nextPos

		// Parse del_targets as a sequence
		var elements []nodes.Expr
		for !p.check(lexer.RightParen) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.RightParen) {
				break
			}
		}

		// Consume the closing parenthesis
		_, err = p.consume(lexer.RightParen, "expected ')' after tuple targets")
		if err != nil {
			return nil, err
		}

		return nodes.NewTupleExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
	} else if p.match(lexer.LeftBracket) {
		// Handle list form
		if p.match(lexer.RightBracket) {
			// Empty list
			return nodes.NewListExpr([]nodes.Expr{}, lexer.Span{Start: startPos, End: p.previous().End()}), nil
		}

		// Parse del_targets as a sequence
		var elements []nodes.Expr
		for !p.check(lexer.RightBracket) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.RightBracket) {
				break
			}
		}

		// Consume the closing bracket
		_, err := p.consume(lexer.RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return nodes.NewListExpr(elements, lexer.Span{Start: startPos, End: p.previous().End()}), nil
	}

	return nil, p.error(p.peek(), "expected NAME, '(' or '[' in del target atom")
}

// delTargets parses del targets as per the grammar:
// del_targets: ','.del_target+ [',']
func (p *Parser) delTargets() (nodes.Expr, error) {
	// Parse the first del_target
	target, err := p.delTarget()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the single target
	if !p.match(lexer.Comma) {
		return target, nil
	}

	// We have a comma, so this is a tuple of targets
	elements := []nodes.Expr{target}

	// Parse additional targets if any
	for !p.check(lexer.Newline) && !p.check(lexer.Semicolon) && !p.isAtEnd() {
		// Allow trailing comma
		if p.check(lexer.Newline) || p.check(lexer.Semicolon) || p.isAtEnd() {
			break
		}

		target, err = p.delTarget()
		if err != nil {
			return nil, err
		}
		elements = append(elements, target)

		// Expect a comma after each target except possibly the last
		if !p.match(lexer.Comma) {
			break
		}
	}

	// Create a tuple expression with the targets
	return nodes.NewTupleExpr(elements, lexer.Span{Start: elements[0].Span().Start, End: elements[len(elements)-1].Span().End}), nil
}

// annotatedRhs parses the right-hand side of an annotated assignment:
// annotated_rhs: yield_expr | star_expressions
func (p *Parser) annotatedRhs() (nodes.Expr, error) {
	if p.check(lexer.Yield) {
		return p.yieldExpression()
	}
	return p.starExpressions()
}

// augassign parses an augmented assignment operator:
// augassign:
//
//	| '+=' | '-=' | '*=' | '@=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>=' | '**=' | '//='
func (p *Parser) augassign() (lexer.Token, error) {
	if p.match(lexer.PlusEqual, lexer.MinusEqual, lexer.StarEqual, lexer.AtEqual, lexer.SlashEqual, lexer.PercentEqual,
		lexer.AmpEqual, lexer.PipeEqual, lexer.CaretEqual, lexer.LessLessEqual, lexer.GreaterGreaterEqual,
		lexer.StarStarEqual, lexer.SlashSlashEqual) {
		return p.previous(), nil
	}
	return lexer.Token{}, p.error(p.peek(), "expected augmented assignment operator")
}

// assignment parses an assignment statement.
func (p *Parser) assignment() (nodes.Stmt, error) {
	startPos := p.peek().Start()
	originalPos := p.Current

	// Try form 1: NAME ':' expression ['=' annotated_rhs]
	if p.check(lexer.Identifier) && p.checkNext(lexer.Colon) {
		name := p.advance()                                         // Consume the NAME
		_, err := p.consume(lexer.Colon, "expected ':' after name") // Consume the ':'
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr nodes.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create a variable annotation statement
		nameExpr := nodes.NewName(name, lexer.Span{Start: name.Start(), End: name.End()})
		return nodes.NewAnnotationStmt(nameExpr, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

	// ('(' single_target ')' | single_subscript_attribute_target) ':' expression ['=' annotated_rhs]

	// First try '(' single_target ')'
	if p.match(lexer.LeftParen) {
		target, err := p.singleTarget()
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Successfully parsed single_target, now expect closing paren
		_, err = p.consume(lexer.RightParen, "expected ')' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Now expect colon
		_, err = p.consume(lexer.Colon, "expected ':' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr nodes.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create annotation statement
		return nodes.NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

trySingleSubscriptAttributeTarget:
	// If any part fails, restore position and try next alternative
	p.Current = originalPos

	// Try single_subscript_attribute_target ':' ...
	target, err := p.singleSubscriptAttributeTarget()
	if err == nil {
		_, err = p.consume(lexer.Colon, "expected ':' after target")
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err == nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr nodes.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create annotation statement
		return nodes.NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

	// Restore position and try form 3: (star_targets '=' )+ (yield_expr | star_expressions) !'=' [TYPE_COMMENT]
	p.Current = originalPos

	// Parse the first star_targets
	targets, err := p.starTargets()
	if err == nil {
		// We have valid targets, now check for '='
		if p.check(lexer.Equal) {
			// Start building the chain of targets
			var targetChain [][]nodes.Expr
			targetChain = append(targetChain, targets)

			lastPos := p.Current
			// Parse additional star_targets '=' pairs
			for p.match(lexer.Equal) {
				moreTargets, err := p.starTargets()
				if err != nil {
					// We've probably consumed the right-hand side expression
					// so we need to restore the position, which should be the last equal sign
					p.Current = lastPos
					break
				}
				targetChain = append(targetChain, moreTargets)
				lastPos = p.Current
			}

			_, err = p.consume(lexer.Equal, "expected '=' after targets")
			if err != nil {
				return nil, err
			}

			// Parse the right-hand side expression
			var rhs nodes.Expr
			if p.check(lexer.Yield) {
				rhs, err = p.yieldExpression()
			} else {
				rhs, err = p.starExpressions()
			}
			if err != nil {
				return nil, err
			}

			// Make sure '=' doesn't follow (used in the grammar to disambiguate)
			if p.check(lexer.Equal) {
				return nil, p.error(p.peek(), "unexpected '=' in assignment")
			}

			// For chain assignments (a = b = c = 1), we create multiple AssignStmt nodes
			// The last one gets the right-hand side expression, and then assign left to right
			// We iterate over the targetChain, and create an AssignStmt for each target
			// TODO: we should assign the RHS expression to a temp variable, and then assign the temp variable to the targets
			var stmts []nodes.Stmt
			for i := 0; i < len(targetChain); i++ {
				stmts = append(stmts, nodes.NewAssignStmt(targetChain[i], rhs, lexer.Span{Start: startPos, End: rhs.Span().End}))
			}
			return nodes.NewMultiStmt(stmts, lexer.Span{Start: startPos, End: rhs.Span().End}), nil
		}
	}

	// Restore position and try form 4: single_target augassign ~ (yield_expr | star_expressions)
	p.Current = originalPos

	// Parse single_target
	singleTarget, err := p.singleTarget()
	if err != nil {
		return nil, err
	}

	// Try to parse augassign
	if p.match(lexer.PlusEqual, lexer.MinusEqual, lexer.StarEqual, lexer.AtEqual, lexer.SlashEqual, lexer.PercentEqual,
		lexer.AmpEqual, lexer.PipeEqual, lexer.CaretEqual, lexer.LessLessEqual, lexer.GreaterGreaterEqual,
		lexer.StarStarEqual, lexer.SlashSlashEqual) {

		op := p.previous()

		// Parse the right-hand side expression
		var value nodes.Expr
		if p.check(lexer.Yield) {
			value, err = p.yieldExpression()
		} else {
			value, err = p.starExpressions()
		}
		if err != nil {
			return nil, err
		}

		// Create and return an AugAssignStmt node
		return nodes.NewAugAssignStmt(singleTarget, op, value, lexer.Span{Start: startPos, End: value.Span().End}), nil
	}

	// If we get here, none of the assignment forms matched
	return nil, p.error(p.peek(), "invalid assignment")
}
