package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// forStatement parses a for statement.
func (p *Parser) forStatement() (ast.Stmt, error) {
	// Check for async
	isAsync := false
	var startToken lexer.Token

	if p.match(lexer.Async) {
		isAsync = true
		startToken = p.previous()
		_, err := p.consume(lexer.For, "expected 'for' after 'async'")
		if err != nil {
			return nil, err
		}
	} else {
		// Consume the 'for' token
		var err error
		startToken, err = p.consume(lexer.For, "expected 'for'")
		if err != nil {
			return nil, err
		}
	}

	// Parse the target (star_targets)
	target, err := p.starTargets()
	if err != nil {
		return nil, err
	}

	// Expect 'in' keyword
	_, err = p.consume(lexer.In, "expected 'in' after target in for statement")
	if err != nil {
		return nil, err
	}

	// Parse the iterable (star_expressions)
	iterable, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after iterable in for statement")
	if err != nil {
		return nil, err
	}

	// Parse the body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Check for else
	elseBody, err := p.elseBlock()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(elseBody) > 0 {
		endPos = elseBody[len(elseBody)-1].GetSpan().End
	} else if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	// Handle the case where target is a list of expressions
	var targetExpr ast.Expr
	if len(target) == 1 {
		targetExpr = target[0]
	} else {
		// If multiple targets, create a tuple
		targetExpr = &ast.TupleExpr{
			Elements: target,
			Span:     lexer.Span{Start: target[0].GetSpan().Start, End: target[len(target)-1].GetSpan().End},
		}
	}

	return &ast.For{
		Target:   targetExpr,
		Iterable: iterable,
		Body:     body,
		Else:     elseBody,
		IsAsync:  isAsync,
		Span:     lexer.Span{Start: startToken.Start(), End: endPos},
	}, nil
}
