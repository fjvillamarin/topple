package compiler

// The scaffold parses only *one-line expression statements* so that you can
// start writing tests immediately and grow the grammar feature-by-feature.

type Parser struct {
	Tokens  []Token
	Current int
	Errors  []error
}

// NewParser returns a new parser instance.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens:  tokens,
		Current: 0,
		Errors:  []error{},
	}
}

// Parse parses the tokens and returns a list of statements.
// It will attempt to recover from errors and return all encountered errors.
func (p *Parser) Parse() (*Module, []error) {
	stmts := []Stmt{}

	for !p.isAtEnd() {
		// Skip over any blank lines between statements. A blank line is just a
		// NEWLINE token that is not part of any real statement. By consuming
		// these eagerly we avoid producing ExprStmt nodes whose Value is nil
		// when the source line is empty.
		for p.check(Newline) {
			p.advance()
		}

		// If we have reached EOF after skipping newlines, break out of the
		// loop early so we don't attempt to parse a statement starting at EOF.
		if p.isAtEnd() {
			break
		}

		stmt, err := p.statement()
		if err != nil {
			p.Errors = append(p.Errors, err)
			return nil, p.Errors
		}
		stmts = append(stmts, stmt)
	}

	return &Module{Body: stmts}, p.Errors
}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

// statement parses a single statement.
func (p *Parser) statement() (Stmt, error) {
	return p.simpleStatement()
}

// simpleStatement parses an expression statement.
func (p *Parser) simpleStatement() (Stmt, error) {
	// Check for keywords first
	switch p.peek().Type {
	case Type:
		return p.typeAlias()
	case Return:
		return p.returnStatement()
	case Import, From:
		return p.importStatement()
	case Raise:
		return p.raiseStatement()
	case Pass:
		return p.passStatement()
	case Del:
		return p.delStatement()
	case Yield:
		return p.yieldStatement()
	case Assert:
		return p.assertStatement()
	case Break:
		return p.breakStatement()
	case Continue:
		return p.continueStatement()
	case Global:
		return p.globalStatement()
	case Nonlocal:
		return p.nonlocalStatement()
	}

	// Check for assignment before expression
	if p.check(Identifier) || p.check(LeftParen) || p.check(LeftBracket) || p.check(Star) {
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

	return NewExprStmt(expr, expr.Start(), expr.End()), nil
}

// typeAlias parses a type alias statement as per the grammar:
// type_alias: "type" NAME [type_params] '=' expression
func (p *Parser) typeAlias() (Stmt, error) {
	// Consume the 'type' keyword
	typeToken, err := p.consume(Type, "expected 'type'")
	if err != nil {
		return nil, err
	}

	// Parse the type name
	name, err := p.consume(Identifier, "expected type name")
	if err != nil {
		return nil, err
	}

	// Check for '[' to parse type parameters
	var params []Expr = nil
	if p.match(LeftBracket) {
		// Parse the type parameters
		params, err = p.typeParams()
		if err != nil {
			return nil, err
		}
	}

	// Consume the '='
	_, err = p.consume(Equal, "expected '='")
	if err != nil {
		return nil, err
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	return NewTypeAlias(name, params, expr, typeToken.Start(), expr.End()), nil
}

// typeParams parses type parameters as per the grammar:
// type_params: '[' type_param_seq ']'
// type_param_seq: ','.type_param+ [',']
func (p *Parser) typeParams() ([]Expr, error) {
	params := []Expr{}

	// Parse type parameters until we hit a closing bracket
	for !p.check(RightBracket) && !p.isAtEnd() {
		// Parse a single type parameter
		param, err := p.typeParam()
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		// If no comma, we're done with the parameter list
		if !p.match(Comma) {
			break
		}
	}

	// Consume the closing bracket
	_, err := p.consume(RightBracket, "expected ']' after type parameters")
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
func (p *Parser) typeParam() (Expr, error) {
	startPos := p.peek().Start()
	isStar := false
	isDoubleStar := false

	// Check for star parameters
	if p.match(Star) {
		isStar = true
	} else if p.match(StarStar) {
		isDoubleStar = true
	}

	// Parse the parameter name
	name, err := p.consume(Identifier, "expected parameter name")
	if err != nil {
		return nil, err
	}

	// Parse optional bound (: expression)
	var bound Expr = nil
	if !isStar && !isDoubleStar && p.match(Colon) {
		bound, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional default
	var defaultValue Expr = nil
	if p.match(Equal) {
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
		endPos = defaultValue.End()
	} else if bound != nil {
		endPos = bound.End()
	}

	return NewTypeParamExpr(name, bound, defaultValue, isStar, isDoubleStar, startPos, endPos), nil
}

func (p *Parser) returnStatement() (Stmt, error) {
	// Consume the 'return' keyword
	returnToken, err := p.consume(Return, "expected 'return'")
	if err != nil {
		return nil, err
	}

	// Exit early if there's no return expression
	if p.isAtEnd() || p.check(Newline) || p.check(Semicolon) {
		p.advance()
		return NewReturnStmt(nil, returnToken.Start(), returnToken.End()), nil
	}

	// Parse the return expression
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	return NewReturnStmt(expr, returnToken.Start(), expr.End()), nil
}

func (p *Parser) importStatement() (Stmt, error) {
	startToken := p.peek()

	// Check which type of import statement this is
	if p.check(Import) {
		return p.importName()
	} else if p.check(From) {
		return p.importFrom()
	}

	return nil, p.error(startToken, "expected 'import' or 'from'")
}

// importName handles: import_name: 'import' dotted_as_names
func (p *Parser) importName() (Stmt, error) {
	// Consume the 'import' keyword
	importToken, err := p.consume(Import, "expected 'import'")
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
	endPos := names[len(names)-1].End()
	return NewImportStmt(names, importToken.Start(), endPos), nil
}

// importFrom handles:
// import_from:
//
//	| 'from' ('.' | '...')* dotted_name 'import' import_from_targets
//	| 'from' ('.' | '...')+ 'import' import_from_targets
func (p *Parser) importFrom() (Stmt, error) {
	// Consume the 'from' keyword
	fromToken, err := p.consume(From, "expected 'from'")
	if err != nil {
		return nil, err
	}

	// Parse leading dots for relative imports
	dotCount := 0
	for p.match(Dot) || p.match(Ellipsis) {
		// Count '.' as 1, '...' as 3
		if p.previous().Type == Dot {
			dotCount += 1
		} else {
			dotCount += 3
		}
	}

	// Parse dotted_name if present (not present for relative-only imports)
	var dottedName *DottedName = nil
	if dotCount == 0 || !p.check(Import) {
		dottedName, err = p.parseDottedName()
		if err != nil {
			return nil, err
		}
	}

	// Consume the 'import' keyword
	_, err = p.consume(Import, "expected 'import' after module name")
	if err != nil {
		return nil, err
	}

	// Parse import_from_targets
	isWildcard := false
	var names []*ImportName

	// Check for '*'
	if p.match(Star) {
		isWildcard = true
		names = []*ImportName{}
	} else if p.match(LeftParen) {
		// '(' import_from_as_names [','] ')'
		names, err = p.parseImportFromAsNames()
		if err != nil {
			return nil, err
		}

		// Allow optional trailing comma
		p.match(Comma)

		// Consume closing parenthesis
		_, err = p.consume(RightParen, "expected ')' after import names")
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
	var endPos Position
	if isWildcard {
		endPos = p.previous().End()
	} else if len(names) > 0 {
		endPos = names[len(names)-1].End()
	} else {
		return nil, p.error(p.previous(), "expected import target after 'import'")
	}

	return NewImportFromStmt(dottedName, dotCount, names, isWildcard, fromToken.Start(), endPos), nil
}

// parseDottedAsNames handles: dotted_as_names: ','.dotted_as_name+
func (p *Parser) parseDottedAsNames() ([]*ImportName, error) {
	var names []*ImportName

	// Parse first dotted_as_name
	name, err := p.parseDottedAsName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	// Parse additional dotted_as_names separated by commas
	for p.match(Comma) {
		name, err = p.parseDottedAsName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

// parseDottedAsName handles: dotted_as_name: dotted_name ['as' NAME]
func (p *Parser) parseDottedAsName() (*ImportName, error) {
	// Start position is the start of the dotted name
	startPos := p.peek().Start()

	// Parse the dotted_name
	dottedName, err := p.parseDottedName()
	if err != nil {
		return nil, err
	}

	// Check for optional 'as NAME'
	var asName *Name = nil
	endPos := dottedName.End()

	if p.match(As) {
		// Parse the alias name
		nameToken, err := p.consume(Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = NewName(nameToken, nameToken.Start(), nameToken.End())
		endPos = asName.End()
	}

	return NewImportName(dottedName, asName, startPos, endPos), nil
}

// parseDottedName handles:
// dotted_name:
//
//	| dotted_name '.' NAME
//	| NAME
func (p *Parser) parseDottedName() (*DottedName, error) {
	startPos := p.peek().Start()

	// First name
	nameToken, err := p.consume(Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := NewName(nameToken, nameToken.Start(), nameToken.End())
	names := []*Name{name}
	endPos := name.End()

	// Parse additional names with dots
	for p.match(Dot) {
		nameToken, err = p.consume(Identifier, "expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		name = NewName(nameToken, nameToken.Start(), nameToken.End())
		names = append(names, name)
		endPos = name.End()
	}

	return NewDottedName(names, startPos, endPos), nil
}

// parseImportFromAsNames handles: import_from_as_names: ','.import_from_as_name+
func (p *Parser) parseImportFromAsNames() ([]*ImportName, error) {
	var names []*ImportName

	// Parse first import_from_as_name
	name, err := p.parseImportFromAsName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	// Parse additional import_from_as_names separated by commas
	for p.match(Comma) {
		// If we see a closing parenthesis next, it's a trailing comma
		if p.check(RightParen) {
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
func (p *Parser) parseImportFromAsName() (*ImportName, error) {
	startPos := p.peek().Start()

	// Parse the first name
	nameToken, err := p.consume(Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := NewName(nameToken, nameToken.Start(), nameToken.End())
	dottedName := NewDottedName([]*Name{name}, name.Start(), name.End())

	// Check for optional 'as NAME'
	var asName *Name = nil
	endPos := dottedName.End()

	if p.match(As) {
		// Parse the alias name
		aliasToken, err := p.consume(Identifier, "expected identifier after 'as'")
		if err != nil {
			return nil, err
		}
		asName = NewName(aliasToken, aliasToken.Start(), aliasToken.End())
		endPos = asName.End()
	}

	return NewImportName(dottedName, asName, startPos, endPos), nil
}

func (p *Parser) raiseStatement() (Stmt, error) {
	// Consume the 'raise' keyword
	raiseToken, err := p.consume(Raise, "expected 'raise'")
	if err != nil {
		return nil, err
	}

	// Check if there's an expression after 'raise'
	if p.isAtEnd() || p.check(Newline) || p.check(Semicolon) {
		// Just a 'raise' with no exception
		endPos := raiseToken.End()
		return NewRaiseStmt(nil, nil, false, false, raiseToken.Start(), endPos), nil
	}

	// Parse the exception expression
	exception, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check if there's a 'from' clause
	hasFrom := false
	var fromExpr Expr = nil
	endPos := exception.End()

	if p.match(From) {
		hasFrom = true
		// Parse the 'from' expression
		fromExpr, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = fromExpr.End()
	}

	return NewRaiseStmt(exception, fromExpr, true, hasFrom, raiseToken.Start(), endPos), nil
}

func (p *Parser) passStatement() (Stmt, error) {
	// Consume the 'pass' keyword
	passToken := p.advance()
	return NewPassStmt(passToken.Start(), passToken.End()), nil
}

func (p *Parser) delStatement() (Stmt, error) {
	// Consume the 'del' keyword
	delToken, err := p.consume(Del, "expected 'del'")
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
	return NewExprStmt(targets, delToken.Start(), targets.End()), nil
}

func (p *Parser) yieldStatement() (Stmt, error) {
	expr, err := p.yieldExpression()
	if err != nil {
		return nil, err
	}
	return NewYieldStmt(expr, expr.Start(), expr.End()), nil
}

func (p *Parser) assertStatement() (Stmt, error) {
	// Consume the 'assert' keyword
	assertToken, err := p.consume(Assert, "expected 'assert'")
	if err != nil {
		return nil, err
	}

	// Parse the test expression
	test, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check for optional message expression
	var message Expr = nil
	endPos := test.End()

	if p.match(Comma) {
		// Parse the message expression
		message, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = message.End()
	}

	return NewAssertStmt(test, message, assertToken.Start(), endPos), nil
}

func (p *Parser) breakStatement() (Stmt, error) {
	// Consume the 'break' keyword
	breakToken := p.advance()
	return NewBreakStmt(breakToken.Start(), breakToken.End()), nil
}

func (p *Parser) continueStatement() (Stmt, error) {
	// Consume the 'continue' keyword
	continueToken := p.advance()
	return NewContinueStmt(continueToken.Start(), continueToken.End()), nil
}

func (p *Parser) globalStatement() (Stmt, error) {
	// Consume the 'global' keyword
	globalToken, err := p.consume(Global, "expected 'global'")
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
	endPos := names[len(names)-1].End()
	return NewGlobalStmt(names, globalToken.Start(), endPos), nil
}

func (p *Parser) nonlocalStatement() (Stmt, error) {
	// Consume the 'nonlocal' keyword
	nonlocalToken, err := p.consume(Nonlocal, "expected 'nonlocal'")
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
	endPos := names[len(names)-1].End()
	return NewNonlocalStmt(names, nonlocalToken.Start(), endPos), nil
}

// parseNameList is a helper function to parse a comma-separated list of identifiers
func (p *Parser) parseNameList() ([]*Name, error) {
	names := []*Name{}

	// Parse first identifier
	nameToken, err := p.consume(Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	// Create a Name node from the token
	name := NewName(nameToken, nameToken.Start(), nameToken.End())
	names = append(names, name)

	// Parse additional identifiers separated by commas
	for p.match(Comma) {
		nameToken, err = p.consume(Identifier, "expected identifier after ','")
		if err != nil {
			return nil, err
		}

		// Create a Name node from the token
		name = NewName(nameToken, nameToken.Start(), nameToken.End())
		names = append(names, name)
	}

	return names, nil
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

// expression parses an expression.
func (p *Parser) expression() (Expr, error) {
	expr, err := p.disjunction()
	if err != nil {
		return nil, err
	}

	if p.match(If) {
		condition, err := p.disjunction()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(Else, "expected 'else' after condition")
		if err != nil {
			return nil, err
		}
		elseExpr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return NewTernaryExpr(condition, expr, elseExpr, expr.Start(), elseExpr.End()), nil
	}

	return expr, nil
}

// disjunction parses a disjunction expression.
func (p *Parser) disjunction() (Expr, error) {
	expr, err := p.conjunction()
	if err != nil {
		return nil, err
	}

	for p.match(Or) {
		operator := p.previous()
		right, err := p.conjunction()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// conjunction parses a conjunction expression.
func (p *Parser) conjunction() (Expr, error) {
	expr, err := p.inversion()
	if err != nil {
		return nil, err
	}

	for p.match(And) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// inversion parses an inversion expression.
func (p *Parser) inversion() (Expr, error) {
	if p.match(Not) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right, operator.Start(), right.End()), nil
	}

	return p.comparison()
}

// comparison parses a comparison expression.
func (p *Parser) comparison() (Expr, error) {
	left, err := p.bitwiseOr()
	if err != nil {
		return nil, err
	}

	// First check for any comparison operator to start a comparison chain
	if isComparisonOperator(p.peek().Type) {
		// We're going to build a chain of comparisons
		operands := []Expr{left}
		operators := []Token{}

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
			return NewBinary(operands[0], operators[0], operands[1], operands[0].Start(), operands[1].End()), nil
		}

		// Handle chained comparisons (a < b < c becomes (a < b) and (b < c))
		var result Expr
		for i := 0; i < len(operators); i++ {
			comparison := NewBinary(operands[i], operators[i], operands[i+1],
				operands[i].Start(), operands[i+1].End())

			if i == 0 {
				result = comparison
			} else {
				// Create an AND expression linking the comparisons
				andToken := Token{Type: And, Lexeme: "and"}
				result = NewBinary(result, andToken, comparison,
					result.Start(), comparison.End())
			}
		}
		return result, nil
	}

	return left, nil
}

// Helper function to check if a token type is a comparison operator
func isComparisonOperator(tokenType TokenType) bool {
	return tokenType == EqualEqual || tokenType == BangEqual ||
		tokenType == Less || tokenType == LessEqual ||
		tokenType == Greater || tokenType == GreaterEqual ||
		tokenType == In || tokenType == Is || tokenType == IsNot || tokenType == NotIn
}

// bitwise_or parses a bitwise OR expression.
func (p *Parser) bitwiseOr() (Expr, error) {
	expr, err := p.bitwiseXor()
	if err != nil {
		return nil, err
	}

	for p.match(Pipe) {
		operator := p.previous()
		right, err := p.bitwiseXor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// bitwiseXor parses a bitwise XOR expression.
func (p *Parser) bitwiseXor() (Expr, error) {
	expr, err := p.bitwiseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(Caret) {
		operator := p.previous()
		right, err := p.bitwiseAnd()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// bitwiseAnd parses a bitwise AND expression.
func (p *Parser) bitwiseAnd() (Expr, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	for p.match(Ampersand) {
		operator := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// shiftExpr parses a shift expression.
func (p *Parser) shiftExpr() (Expr, error) {
	expr, err := p.sum()
	if err != nil {
		return nil, err
	}

	for p.match(LessLess, GreaterGreater) {
		operator := p.previous()
		right, err := p.sum()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// sum parses a sum expression.
func (p *Parser) sum() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(Plus, Minus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// term parses a term expression.
func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(Star, Slash, SlashSlash, Percent, At) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// factor parses a factor expression.
func (p *Parser) factor() (Expr, error) {
	if p.match(Plus, Minus, Tilde) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right, operator.Start(), right.End()), nil
	}

	return p.power()
}

// power parses a power expression.
func (p *Parser) power() (Expr, error) {
	expr, err := p.await()
	if err != nil {
		return nil, err
	}

	if p.match(StarStar) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return NewBinary(expr, operator, right, expr.Start(), right.End()), nil
	}

	return expr, nil
}

func (p *Parser) await() (Expr, error) {
	// Check if the current token is an await keyword
	if p.match(Await) {
		awaitToken := p.previous()
		expr, err := p.primary()
		if err != nil {
			return nil, err
		}
		return NewAwaitExpr(expr, awaitToken.Start(), expr.End()), nil
	}

	return p.primary()
}

// primary parses a primary expression.
func (p *Parser) primary() (Expr, error) {
	// Parse the initial atom
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Keep parsing postfix operations while they exist
	for {
		if p.match(Dot) {
			// Handle attribute access: expr.NAME
			name, err := p.consume(Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = NewAttribute(expr, name, expr.Start(), name.End())
		} else if p.match(LeftParen) {
			// Handle function call: expr(args)
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(LeftBracket) {
			// Handle subscript access: expr[index] or expr[slice]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = NewSubscript(expr, indices, expr.Start(), right.End())
		} else {
			// No more postfix operations
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := []Expr{}
	if !p.check(RightParen) {
		// Parse first argument
		arg, err := p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional arguments
		for p.match(Comma) {
			// If there's a right parenthesis after the comma,
			// it's a trailing comma, so we're done parsing arguments
			if p.check(RightParen) {
				break
			}

			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	right, err := p.consume(RightParen, "expected ')' after arguments")
	if err != nil {
		return nil, err
	}

	return NewCall(callee, args, callee.Start(), right.End()), nil
}

// atom parses an atom.
func (p *Parser) atom() (Expr, error) {
	if p.match(False) {
		return NewLiteral(p.previous(), false, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(True) {
		return NewLiteral(p.previous(), true, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(None) {
		return NewLiteral(p.previous(), nil, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Number, String) {
		return NewLiteral(p.previous(), p.previous().Literal, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Ellipsis) {
		return NewLiteral(p.previous(), nil, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Identifier) {
		return NewName(p.previous(), p.previous().Start(), p.previous().End()), nil
	}

	if p.check(LeftParen) {
		// This could be either a tuple or a group (parenthesized expression)
		return p.tuple()
	}

	if p.check(LeftBracket) {
		// List literal
		return p.list()
	}

	if p.check(LeftBrace) {
		// Set literal
		return p.set()
	}

	if p.check(Yield) {
		// Yield expression
		return p.yieldExpression()
	}

	return nil, p.error(p.peek(), "unexpected token")
}

func (p *Parser) list() (Expr, error) {
	// Expect opening bracket
	leftBracket, err := p.consume(LeftBracket, "expected '['")
	if err != nil {
		return nil, err
	}

	elements := []Expr{}

	// Parse elements if the list is not empty
	if !p.check(RightBracket) {
		// Parse star named expressions
		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)

		// Parse additional elements separated by commas
		for p.match(Comma) {
			// Allow trailing comma
			if p.check(RightBracket) {
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
	rightBracket, err := p.consume(RightBracket, "expected ']'")
	if err != nil {
		return nil, err
	}

	return NewListExpr(elements, leftBracket.Start(), rightBracket.End()), nil
}

func (p *Parser) tuple() (Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Empty tuple
	if p.check(RightParen) {
		rightParen, err := p.consume(RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return NewTupleExpr([]Expr{}, leftParen.Start(), rightParen.End()), nil
	}

	// Parse the first expression (could be a star expression or named expression)
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}

	// If the next token is 'yield', it's a group
	if p.checkNext(Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		rightParen, err := p.consume(RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return NewGroupExpr(expr, leftParen.Start(), rightParen.End()), nil
	}

	// If there's a comma, it's a tuple
	if p.match(Comma) {
		elements := []Expr{expr}

		// Parse additional elements if present
		if !p.check(RightParen) {
			for {
				nextExpr, err := p.starNamedExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, nextExpr)

				if !p.match(Comma) {
					break
				}

				// Allow trailing comma
				if p.check(RightParen) {
					break
				}
			}
		}

		rightParen, err := p.consume(RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return NewTupleExpr(elements, leftParen.Start(), rightParen.End()), nil
	} else {
		// No comma, so it's a group
		// Groups can only contain named expressions, not star expressions
		_, isStarExpr := expr.(*StarExpr)
		if isStarExpr {
			return nil, p.error(p.previous(), "starred expression cannot appear in a group")
		}

		rightParen, err := p.consume(RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return NewGroupExpr(expr, leftParen.Start(), rightParen.End()), nil
	}
}

func (p *Parser) set() (Expr, error) {
	// Expect opening brace
	leftBrace, err := p.consume(LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	elements := []Expr{}

	// Parse elements
	// First element
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}
	elements = append(elements, expr)

	// Parse additional elements separated by commas
	for p.match(Comma) {
		// Allow trailing comma
		if p.check(RightBrace) {
			break
		}

		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	// Expect closing brace
	rightBrace, err := p.consume(RightBrace, "expected '}'")
	if err != nil {
		return nil, err
	}

	return NewSetExpr(elements, leftBrace.Start(), rightBrace.End()), nil
}

func (p *Parser) yieldExpression() (Expr, error) {
	// Expect 'yield' keyword
	yieldToken, err := p.consume(Yield, "expected 'yield'")
	if err != nil {
		return nil, err
	}

	// Check for 'yield from' form
	isFrom := false
	if p.match(From) {
		isFrom = true
		// Parse the expression after 'yield from'
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return NewYieldExpr(isFrom, expr, yieldToken.Start(), expr.End()), nil
	}

	// Check if there's an expression after 'yield'
	if p.isAtEnd() || p.check(Newline) || p.check(Semicolon) || p.check(RightParen) || p.check(Comma) {
		// No expression, yield on its own
		return NewYieldExpr(false, nil, yieldToken.Start(), yieldToken.End()), nil
	}

	// Parse star expressions after 'yield'
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}
	return NewYieldExpr(false, expr, yieldToken.Start(), expr.End()), nil
}

func (p *Parser) starExpressions() (Expr, error) {
	// Parse the first star expression
	expr, err := p.starExpression()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the expression as is
	if !p.match(Comma) {
		return expr, nil
	}

	// We have a comma, so this is a tuple of expressions
	elements := []Expr{expr}

	// Allow trailing comma with no following expression
	if !p.check(Newline) && !p.check(RightParen) && !p.check(RightBracket) && !p.check(RightBrace) && !p.check(Semicolon) && !p.isAtEnd() {
		// Parse subsequent expressions
		for {
			expr, err := p.starExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)

			if !p.match(Comma) {
				break
			}

			// Allow trailing comma
			if p.check(Newline) || p.check(RightParen) || p.check(RightBracket) || p.check(RightBrace) || p.check(Semicolon) || p.isAtEnd() {
				break
			}
		}
	}

	// Create a tuple with the collected expressions
	return NewTupleExpr(elements, elements[0].Start(), elements[len(elements)-1].End()), nil
}

func (p *Parser) group() (Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Check for yield expression
	if p.check(Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		// Expect closing parenthesis
		rightParen, err := p.consume(RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}

		return NewGroupExpr(expr, leftParen.Start(), rightParen.End()), nil
	}

	// Parse named expression
	expr, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect closing parenthesis
	rightParen, err := p.consume(RightParen, "expected ')'")
	if err != nil {
		return nil, err
	}

	return NewGroupExpr(expr, leftParen.Start(), rightParen.End()), nil
}

func (p *Parser) starExpression() (Expr, error) {
	if p.match(Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return NewStarExpr(expr, star.Start(), expr.End()), nil
	}

	// Not a star expression, parse as a regular expression
	return p.expression()
}

func (p *Parser) namedExpression() (Expr, error) {
	// Handle assignment expressions first (the walrus operator)
	if p.check(Identifier) && p.checkNext(Walrus) {
		name, err := p.consume(Identifier, "expected identifier") // Consume the identifier
		if err != nil {
			return nil, err
		}
		_, err = p.consume(Walrus, "expected ':=' after identifier") // Consume the ':=' operator
		if err != nil {
			return nil, err
		}

		value, err := p.expression()
		if err != nil {
			return nil, err
		}

		return NewAssignExpr(
			NewName(name, name.Start(), name.End()),
			value,
			name.Start(),
			value.End(),
		), nil
	}

	// If not an assignment expression, parse a regular expression
	return p.expression()
}

func (p *Parser) starNamedExpression() (Expr, error) {
	if p.match(Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return NewStarExpr(expr, star.Start(), expr.End()), nil
	}

	// Not a star expression, parse as a regular expression
	return p.namedExpression()
}

// slice parses a single slice element as per the grammar:
// slice:
//
//	| [expression] ':' [expression] [':' [expression] ]
//	| named_expression
func (p *Parser) slice() (Expr, error) {
	// Check if this is a slice notation or just an expression
	// We need to look ahead to see if there's a colon after the first expression (if any)
	startPos := p.peek().Start()

	// Empty slice is allowed (:)
	if p.check(Colon) {
		// No start expression, consume the colon
		p.advance()

		var end Expr
		var err error

		// Check for end expression after colon
		if !p.check(Colon) && !p.check(RightBracket) && !p.check(Comma) {
			end, err = p.expression()
			if err != nil {
				return nil, err
			}
		}

		// Check for step (second colon)
		var step Expr
		if p.match(Colon) {
			// Parse optional step
			if !p.check(RightBracket) && !p.check(Comma) {
				step, err = p.expression()
				if err != nil {
					return nil, err
				}
			}
		}

		endPos := p.previous().End()
		if step != nil {
			endPos = step.End()
		} else if end != nil {
			endPos = end.End()
		}

		return NewSlice(nil, end, step, startPos, endPos), nil
	}

	// There's an expression before any potential colon
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	// If there's no colon after the expression, this is just a regular index
	if !p.match(Colon) {
		return expr, nil
	}

	// We have a slice with a start expression
	var end Expr

	// Check for end expression after colon
	if !p.check(Colon) && !p.check(RightBracket) && !p.check(Comma) {
		end, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Check for step (second colon)
	var step Expr
	if p.match(Colon) {
		// Parse optional step
		if !p.check(RightBracket) && !p.check(Comma) {
			step, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
	}

	endPos := p.previous().End()
	if step != nil {
		endPos = step.End()
	} else if end != nil {
		endPos = end.End()
	} else {
		endPos = expr.End()
	}

	return NewSlice(expr, end, step, startPos, endPos), nil
}

// slices parses one or more slice elements as per the grammar:
// slices:
//
//	| slice !','
//	| ','.(slice | starred_expression)+ [',']
func (p *Parser) slices() ([]Expr, error) {
	// Parse the first slice
	first, err := p.slice()
	if err != nil {
		return nil, err
	}

	// Initialize the slice with the first element
	indices := []Expr{first}

	// If there's no comma, it's just a single index/slice
	if !p.match(Comma) {
		return indices, nil
	}

	// Parse additional slice or starred_expression elements
	for !p.check(RightBracket) {
		// Handle starred expressions
		if p.match(Star) {
			// This is a starred expression like *args
			star := p.previous()
			expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
			if err != nil {
				return nil, err
			}
			indices = append(indices, NewStarExpr(expr, star.Start(), expr.End()))
		} else {
			// Regular slice expression
			expr, err := p.slice()
			if err != nil {
				return nil, err
			}
			indices = append(indices, expr)
		}

		// Break if no more commas
		if !p.match(Comma) {
			break
		}

		// Allow trailing comma
		if p.check(RightBracket) {
			break
		}
	}

	return indices, nil
}

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

func (p *Parser) consume(t TokenType, message string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(token Token, message string) error {
	return &ParseError{Token: token, Message: message}
}

// match checks if the current token is one of the given types.
func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) checkNext(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekN(1).Type == t
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.Current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) peek() Token {
	return p.Tokens[p.Current]
}

func (p *Parser) peekN(n int) Token {
	return p.Tokens[p.Current+n]
}

func (p *Parser) previous() Token {
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
func (p *Parser) tPrimary() (Expr, error) {
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

		if p.match(Dot) {
			// Rule: t_primary '.' NAME &t_lookahead
			name, err := p.consume(Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = NewAttribute(expr, name, expr.Start(), name.End())
		} else if p.match(LeftParen) {
			// Rule: t_primary '(' [arguments] ')' &t_lookahead
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}

			// TODO: check for genexp if call fails
		} else if p.match(LeftBracket) {
			// Rule: t_primary '[' slices ']' &t_lookahead
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = NewSubscript(expr, indices, expr.Start(), right.End())
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
	return tokenType == LeftParen || tokenType == LeftBracket || tokenType == Dot
}

// singleSubscriptAttributeTarget parses a single target with attribute or subscription
// as per the grammar:
//
// single_subscript_attribute_target:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
func (p *Parser) singleSubscriptAttributeTarget() (Expr, error) {
	// Parse the t_primary expression
	expr, err := p.tPrimary()
	if err != nil {
		return nil, err
	}

	// Check which form it is
	if p.match(Dot) {
		// Handle attribute access: t_primary.NAME
		name, err := p.consume(Identifier, "expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		result := NewAttribute(expr, name, expr.Start(), name.End())

		// Check negative lookahead - must NOT be followed by another accessor
		if p.tLookahead() {
			return nil, p.error(p.peek(), "unexpected accessor after attribute target")
		}

		return result, nil
	} else if p.match(LeftBracket) {
		// Handle subscript access: t_primary[slices]
		indices, err := p.slices()
		if err != nil {
			return nil, err
		}

		right, err := p.consume(RightBracket, "expected ']' after index")
		if err != nil {
			return nil, err
		}
		result := NewSubscript(expr, indices, expr.Start(), right.End())

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
func (p *Parser) singleTarget() (Expr, error) {
	if p.check(Identifier) {
		// Handle the NAME case first
		// But first check if it might be a single_subscript_attribute_target
		// by seeing if there's a lookahead accessor after the identifier
		if p.checkNext(Dot) || p.checkNext(LeftBracket) || p.checkNext(LeftParen) {
			return p.singleSubscriptAttributeTarget()
		}

		// Just a NAME
		name := p.advance()
		return NewName(name, name.Start(), name.End()), nil
	} else if p.match(LeftParen) {
		// Handle parenthesized form: '(' single_target ')'
		target, err := p.singleTarget()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(RightParen, "expected ')' after target")
		if err != nil {
			return nil, err
		}

		return NewGroupExpr(target, p.previous().Start(), p.previous().End()), nil
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
func (p *Parser) starTarget() (Expr, error) {
	if p.match(Star) {
		// Handle starred expression
		star := p.previous()

		// Check not followed by another star
		if p.check(Star) {
			return nil, p.error(p.peek(), "cannot use ** in target expressions")
		}

		// Parse the inner star_target
		expr, err := p.starTarget()
		if err != nil {
			return nil, err
		}

		return NewStarExpr(expr, star.Start(), expr.End()), nil
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
func (p *Parser) starTargets() ([]Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	targets := []Expr{target}

	// If there's no comma, return the single target
	if !p.match(Comma) {
		return targets, nil
	}

	// We've consumed a comma
	// Check if we're at the end (trailing comma case)
	if p.check(RightParen) || p.check(RightBracket) ||
		p.check(Colon) || p.check(Equal) || p.check(Newline) ||
		p.check(Semicolon) || p.isAtEnd() {
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
		if !p.match(Comma) {
			break
		}

		// If we've found a trailing comma, we're done
		if p.check(RightParen) || p.check(RightBracket) ||
			p.check(Colon) || p.check(Equal) || p.check(Newline) ||
			p.check(Semicolon) || p.isAtEnd() {
			break
		}
	}

	return targets, nil
}

// parseStarTargetSequence parses a sequence of star targets for either list or tuple contexts
// isTuple indicates whether tuple rules should be enforced (requiring comma for single element)
// closingToken specifies the token that would terminate the sequence (RightParen or RightBracket)
func (p *Parser) parseStarTargetSequence(isTuple bool, closingToken TokenType) ([]Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	elements := []Expr{target}

	// For tuples, a comma is required
	if isTuple && !p.match(Comma) {
		return nil, p.error(p.peek(), "expected ',' after target in tuple")
	} else if !isTuple {
		// For lists, comma is optional
		if !p.match(Comma) {
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
			for p.match(Comma) {
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

				if !p.match(Comma) {
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
func (p *Parser) starAtom() (Expr, error) {
	startPos := p.peek().Start()

	if p.match(Identifier) {
		// Handle simple NAME case
		name := p.previous()
		return NewName(name, startPos, name.End()), nil
	} else if p.match(LeftParen) {
		// Handle parenthesized forms
		if p.match(RightParen) {
			// Empty tuple
			return NewTupleExpr([]Expr{}, startPos, p.previous().End()), nil
		}

		// Try to parse as target_with_star_atom first
		nextPos := p.Current
		target, err := p.targetWithStarAtom()

		if err == nil {
			// Successfully parsed as target_with_star_atom
			// Consume the closing parenthesis
			_, err = p.consume(RightParen, "expected ')' after target")
			if err != nil {
				goto tryStartTargetSequence
			}
			return NewGroupExpr(target, startPos, p.previous().End()), nil
		}

	tryStartTargetSequence:
		// Restore position and try as star_targets_tuple_seq
		p.Current = nextPos

		elements, err := p.parseStarTargetSequence(true, RightParen)
		if err != nil {
			return nil, err
		}

		// Consume the closing parenthesis
		_, err = p.consume(RightParen, "expected ')' after tuple targets")
		if err != nil {
			return nil, err
		}

		return NewTupleExpr(elements, startPos, p.previous().End()), nil
	} else if p.match(LeftBracket) {
		// Handle list form
		if p.match(RightBracket) {
			// Empty list
			return NewListExpr([]Expr{}, startPos, p.previous().End()), nil
		}

		elements, err := p.parseStarTargetSequence(false, RightBracket)
		if err != nil {
			return nil, err
		}

		// Consume the closing bracket
		_, err = p.consume(RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return NewListExpr(elements, startPos, p.previous().End()), nil
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
func (p *Parser) targetWithStarAtom() (Expr, error) {
	// Try to parse as t_primary if the next token could start a t_primary
	if p.check(Identifier) || p.check(LeftParen) || p.check(LeftBracket) ||
		p.check(False) || p.check(True) || p.check(None) ||
		p.check(Number) || p.check(String) || p.check(Ellipsis) {

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

		if p.match(Dot) {
			// Handle attribute access: t_primary.NAME
			name, err := p.consume(Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after attribute target")
			}

			return NewAttribute(primary, name, primary.Start(), name.End()), nil
		}

		// Restore position and try t_primary followed by '['
		if p.match(LeftBracket) {
			// Handle subscript access: t_primary[slices]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after subscript target")
			}

			return NewSubscript(primary, indices, primary.Start(), right.End()), nil
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
func (p *Parser) delTarget() (Expr, error) {
	// Try to parse as t_primary if the next token could start a t_primary
	if p.check(Identifier) || p.check(LeftParen) || p.check(LeftBracket) ||
		p.check(False) || p.check(True) || p.check(None) ||
		p.check(Number) || p.check(String) || p.check(Ellipsis) {

		// First, save the current position
		startPos := p.Current

		// Try to parse a t_primary followed by '.'
		primary, err := p.tPrimary()
		if err == nil && p.match(Dot) {
			// Handle attribute access: t_primary.NAME
			name, err := p.consume(Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			result := NewAttribute(primary, name, primary.Start(), name.End())

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after attribute in del target")
			}

			return result, nil
		}

		// Restore position and try t_primary followed by '['
		p.Current = startPos
		primary, err = p.tPrimary()
		if err == nil && p.match(LeftBracket) {
			// Handle subscript access: t_primary[slices]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			result := NewSubscript(primary, indices, primary.Start(), right.End())

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
func (p *Parser) delTAtom() (Expr, error) {
	startPos := p.peek().Start()

	if p.check(Identifier) {
		// Handle simple NAME case
		name := p.advance()
		return NewName(name, startPos, name.End()), nil
	} else if p.match(LeftParen) {
		// Handle parenthesized forms
		if p.match(RightParen) {
			// Empty tuple
			return NewTupleExpr([]Expr{}, startPos, p.previous().End()), nil
		}

		// Try to parse as single del_target first
		nextPos := p.Current
		target, err := p.delTarget()

		if err == nil {
			// Check if there's a comma after, which means it's a tuple
			if p.match(Comma) {
				// Start a tuple with the first target
				elements := []Expr{target}

				// Check for empty rest of tuple
				if p.match(RightParen) {
					return NewTupleExpr(elements, startPos, p.previous().End()), nil
				}

				// Parse rest of del_targets
				for !p.check(RightParen) {
					target, err = p.delTarget()
					if err != nil {
						return nil, err
					}
					elements = append(elements, target)

					if !p.match(Comma) {
						break
					}
				}

				// Consume closing parenthesis
				_, err = p.consume(RightParen, "expected ')' after del targets")
				if err != nil {
					return nil, err
				}

				return NewTupleExpr(elements, startPos, p.previous().End()), nil
			}

			// No comma, so it's a grouped expression
			_, err = p.consume(RightParen, "expected ')' after target")
			if err != nil {
				return nil, err
			}
			return NewGroupExpr(target, startPos, p.previous().End()), nil
		}

		// Restore position and try as del_targets (tuple)
		p.Current = nextPos

		// Parse del_targets as a sequence
		var elements []Expr
		for !p.check(RightParen) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(Comma) {
				break
			}

			// Allow trailing comma
			if p.check(RightParen) {
				break
			}
		}

		// Consume the closing parenthesis
		_, err = p.consume(RightParen, "expected ')' after tuple targets")
		if err != nil {
			return nil, err
		}

		return NewTupleExpr(elements, startPos, p.previous().End()), nil
	} else if p.match(LeftBracket) {
		// Handle list form
		if p.match(RightBracket) {
			// Empty list
			return NewListExpr([]Expr{}, startPos, p.previous().End()), nil
		}

		// Parse del_targets as a sequence
		var elements []Expr
		for !p.check(RightBracket) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(Comma) {
				break
			}

			// Allow trailing comma
			if p.check(RightBracket) {
				break
			}
		}

		// Consume the closing bracket
		_, err := p.consume(RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return NewListExpr(elements, startPos, p.previous().End()), nil
	}

	return nil, p.error(p.peek(), "expected NAME, '(' or '[' in del target atom")
}

// delTargets parses del targets as per the grammar:
// del_targets: ','.del_target+ [',']
func (p *Parser) delTargets() (Expr, error) {
	// Parse the first del_target
	target, err := p.delTarget()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the single target
	if !p.match(Comma) {
		return target, nil
	}

	// We have a comma, so this is a tuple of targets
	elements := []Expr{target}

	// Parse additional targets if any
	for !p.check(Newline) && !p.check(Semicolon) && !p.isAtEnd() {
		// Allow trailing comma
		if p.check(Newline) || p.check(Semicolon) || p.isAtEnd() {
			break
		}

		target, err = p.delTarget()
		if err != nil {
			return nil, err
		}
		elements = append(elements, target)

		// Expect a comma after each target except possibly the last
		if !p.match(Comma) {
			break
		}
	}

	// Create a tuple expression with the targets
	return NewTupleExpr(elements, elements[0].Start(), elements[len(elements)-1].End()), nil
}

// annotatedRhs parses the right-hand side of an annotated assignment:
// annotated_rhs: yield_expr | star_expressions
func (p *Parser) annotatedRhs() (Expr, error) {
	if p.check(Yield) {
		return p.yieldExpression()
	}
	return p.starExpressions()
}

// augassign parses an augmented assignment operator:
// augassign:
//
//	| '+=' | '-=' | '*=' | '@=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>=' | '**=' | '//='
func (p *Parser) augassign() (Token, error) {
	if p.match(PlusEqual, MinusEqual, StarEqual, AtEqual, SlashEqual, PercentEqual,
		AmpEqual, PipeEqual, CaretEqual, LessLessEqual, GreaterGreaterEqual,
		StarStarEqual, SlashSlashEqual) {
		return p.previous(), nil
	}
	return Token{}, p.error(p.peek(), "expected augmented assignment operator")
}

// assignment parses an assignment statement.
func (p *Parser) assignment() (Stmt, error) {
	startPos := p.peek().Start()
	originalPos := p.Current

	// Try form 1: NAME ':' expression ['=' annotated_rhs]
	if p.check(Identifier) && p.checkNext(Colon) {
		name := p.advance()                                   // Consume the NAME
		_, err := p.consume(Colon, "expected ':' after name") // Consume the ':'
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr Expr = nil
		hasValue := false
		if p.match(Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.End()
		if valueExpr != nil {
			endPos = valueExpr.End()
		}

		// Create a variable annotation statement
		nameExpr := NewName(name, name.Start(), name.End())
		return NewAnnotationStmt(nameExpr, typeExpr, valueExpr, hasValue, startPos, endPos), nil
	}

	// ('(' single_target ')' | single_subscript_attribute_target) ':' expression ['=' annotated_rhs]

	// First try '(' single_target ')'
	if p.match(LeftParen) {
		target, err := p.singleTarget()
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Successfully parsed single_target, now expect closing paren
		_, err = p.consume(RightParen, "expected ')' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Now expect colon
		_, err = p.consume(Colon, "expected ':' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr Expr = nil
		hasValue := false
		if p.match(Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.End()
		if valueExpr != nil {
			endPos = valueExpr.End()
		}

		// Create annotation statement
		return NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, startPos, endPos), nil
	}

trySingleSubscriptAttributeTarget:
	// If any part fails, restore position and try next alternative
	p.Current = originalPos

	// Try single_subscript_attribute_target ':' ...
	target, err := p.singleSubscriptAttributeTarget()
	if err == nil {
		_, err = p.consume(Colon, "expected ':' after target")
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err == nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr Expr = nil
		hasValue := false
		if p.match(Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.End()
		if valueExpr != nil {
			endPos = valueExpr.End()
		}

		// Create annotation statement
		return NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, startPos, endPos), nil
	}

	// Restore position and try form 3: (star_targets '=' )+ (yield_expr | star_expressions) !'=' [TYPE_COMMENT]
	p.Current = originalPos

	// Parse the first star_targets
	targets, err := p.starTargets()
	if err == nil {
		// We have valid targets, now check for '='
		if p.check(Equal) {
			// Start building the chain of targets
			var targetChain [][]Expr
			targetChain = append(targetChain, targets)

			lastPos := p.Current
			// Parse additional star_targets '=' pairs
			for p.match(Equal) {
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

			_, err = p.consume(Equal, "expected '=' after targets")
			if err != nil {
				return nil, err
			}

			// Parse the right-hand side expression
			var rhs Expr
			if p.check(Yield) {
				rhs, err = p.yieldExpression()
			} else {
				rhs, err = p.starExpressions()
			}
			if err != nil {
				return nil, err
			}

			// Make sure '=' doesn't follow (used in the grammar to disambiguate)
			if p.check(Equal) {
				return nil, p.error(p.peek(), "unexpected '=' in assignment")
			}

			// For chain assignments (a = b = c = 1), we create multiple AssignStmt nodes
			// The last one gets the right-hand side expression, and then assign left to right
			// We iterate over the targetChain, and create an AssignStmt for each target
			// TODO: we should assign the RHS expression to a temp variable, and then assign the temp variable to the targets
			var stmts []Stmt
			for i := 0; i < len(targetChain); i++ {
				stmts = append(stmts, NewAssignStmt(targetChain[i], rhs, startPos, rhs.End()))
			}
			return NewMultiStmt(stmts, startPos, rhs.End()), nil
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
	if p.match(PlusEqual, MinusEqual, StarEqual, AtEqual, SlashEqual, PercentEqual,
		AmpEqual, PipeEqual, CaretEqual, LessLessEqual, GreaterGreaterEqual,
		StarStarEqual, SlashSlashEqual) {

		op := p.previous()

		// Parse the right-hand side expression
		var value Expr
		if p.check(Yield) {
			value, err = p.yieldExpression()
		} else {
			value, err = p.starExpressions()
		}
		if err != nil {
			return nil, err
		}

		// Create and return an AugAssignStmt node
		return NewAugAssignStmt(singleTarget, op, value, startPos, value.End()), nil
	}

	// If we get here, none of the assignment forms matched
	return nil, p.error(p.peek(), "invalid assignment")
}
