package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// forIfClauses parses one or more for_if_clause according to the grammar:
// for_if_clauses: for_if_clause+
func (p *Parser) forIfClauses() ([]ast.ForIfClause, error) {
	var clauses []ast.ForIfClause

	// Parse the first clause
	clause, err := p.forIfClause()
	if err != nil {
		return nil, err
	}
	clauses = append(clauses, clause)

	// Parse additional clauses
	for p.check(lexer.For) || p.check(lexer.Async) {
		clause, err := p.forIfClause()
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	return clauses, nil
}

// forIfClause parses a single for_if_clause according to the grammar:
// for_if_clause:
//
//	| 'async' 'for' star_targets 'in' ~ disjunction ('if' disjunction )*
//	| 'for' star_targets 'in' ~ disjunction ('if' disjunction )*
func (p *Parser) forIfClause() (ast.ForIfClause, error) {
	var clause ast.ForIfClause

	// Check for async
	if p.match(lexer.Async) {
		clause.IsAsync = true
	}

	// Consume 'for'
	_, err := p.consume(lexer.For, "expected 'for'")
	if err != nil {
		return clause, err
	}

	// Parse star_targets (the loop variables)
	targets, err := p.starTargets()
	if err != nil {
		return clause, err
	}

	// Convert multiple targets to a tuple, or use single target directly
	if len(targets) == 1 {
		clause.Target = targets[0]
	} else {
		// Multiple targets - create a tuple
		clause.Target = &ast.TupleExpr{
			Elements: targets,
			Span:     lexer.Span{Start: targets[0].GetSpan().Start, End: targets[len(targets)-1].GetSpan().End},
		}
	}

	// Consume 'in'
	_, err = p.consume(lexer.In, "expected 'in' after for targets")
	if err != nil {
		return clause, err
	}

	// Parse the iterable expression (disjunction)
	iter, err := p.disjunction()
	if err != nil {
		return clause, err
	}
	clause.Iter = iter

	// Parse zero or more 'if' conditions
	for p.match(lexer.If) {
		condition, err := p.disjunction()
		if err != nil {
			return clause, err
		}
		clause.Ifs = append(clause.Ifs, condition)
	}

	return clause, nil
}

// listComp parses a list comprehension according to the grammar:
// listcomp: '[' named_expression for_if_clauses ']'
func (p *Parser) listComp(element ast.Expr, leftBracket lexer.Token) (ast.Expr, error) {
	// Parse for_if_clauses
	clauses, err := p.forIfClauses()
	if err != nil {
		return nil, err
	}

	// Consume the closing ']'
	rightBracket, err := p.consume(lexer.RightBracket, "expected ']' after list comprehension")
	if err != nil {
		return nil, err
	}

	return &ast.ListComp{
		Element: element,
		Clauses: clauses,
		Span:    lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
	}, nil
}

// setComp parses a set comprehension according to the grammar:
// setcomp: '{' named_expression for_if_clauses '}'
func (p *Parser) setComp(element ast.Expr, leftBrace lexer.Token) (ast.Expr, error) {
	// Parse for_if_clauses
	clauses, err := p.forIfClauses()
	if err != nil {
		return nil, err
	}

	// Consume the closing '}'
	rightBrace, err := p.consume(lexer.RightBrace, "expected '}' after set comprehension")
	if err != nil {
		return nil, err
	}

	return &ast.SetComp{
		Element: element,
		Clauses: clauses,
		Span:    lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
	}, nil
}

// dictComp parses a dictionary comprehension according to the grammar:
// dictcomp: '{' kvpair for_if_clauses '}'
func (p *Parser) dictComp(key, value ast.Expr, leftBrace lexer.Token) (ast.Expr, error) {
	// Parse for_if_clauses
	clauses, err := p.forIfClauses()
	if err != nil {
		return nil, err
	}

	// Consume the closing '}'
	rightBrace, err := p.consume(lexer.RightBrace, "expected '}' after dictionary comprehension")
	if err != nil {
		return nil, err
	}

	return &ast.DictComp{
		Key:     key,
		Value:   value,
		Clauses: clauses,
		Span:    lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
	}, nil
}

// listOrListComp determines whether to parse a list literal or list comprehension
func (p *Parser) listOrListComp() (ast.Expr, error) {
	// Consume the opening '['
	leftBracket, err := p.consume(lexer.LeftBracket, "expected '['")
	if err != nil {
		return nil, err
	}

	// Check for empty list
	if p.check(lexer.RightBracket) {
		rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
		if err != nil {
			return nil, err
		}
		return &ast.ListExpr{
			Elements: []ast.Expr{},
			Span:     lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
		}, nil
	}

	// Save position to potentially backtrack
	startPos := p.Current

	// Parse the first expression
	element, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Check if this is a comprehension (look for 'for' or 'async')
	if p.check(lexer.For) || p.check(lexer.Async) {
		// It's a list comprehension
		return p.listComp(element, leftBracket)
	}

	// It's a regular list - restore position and parse as list
	p.Current = startPos
	return p.listFromBracket(leftBracket)
}

// listFromBracket parses a regular list starting from an already consumed '['
func (p *Parser) listFromBracket(leftBracket lexer.Token) (ast.Expr, error) {
	var elements []ast.Expr

	// Parse the first element
	element, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}
	elements = append(elements, element)

	// Parse additional elements separated by commas
	for p.match(lexer.Comma) {
		// Allow trailing comma
		if p.check(lexer.RightBracket) {
			break
		}

		element, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}

	// Consume the closing ']'
	rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
	if err != nil {
		return nil, err
	}

	return &ast.ListExpr{
		Elements: elements,
		Span:     lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
	}, nil
}

// genExpr parses a generator expression according to the grammar:
// genexp: '(' ( assignment_expression | expression !':=') for_if_clauses ')'
func (p *Parser) genExpr(element ast.Expr, leftParen lexer.Token) (ast.Expr, error) {
	// Parse for_if_clauses
	clauses, err := p.forIfClauses()
	if err != nil {
		return nil, err
	}

	// Consume the closing ')'
	rightParen, err := p.consume(lexer.RightParen, "expected ')' after generator expression")
	if err != nil {
		return nil, err
	}

	return &ast.GenExpr{
		Element: element,
		Clauses: clauses,
		Span:    lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
	}, nil
}
