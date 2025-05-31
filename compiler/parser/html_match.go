package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// htmlMatchStatement parses a match statement within HTML content context.
// Uses the same AST node as regular match statements but uses HTML content block parsing.
func (p *Parser) htmlMatchStatement() (ast.Stmt, error) {
	matchToken, err := p.consume(lexer.Match, "expected 'match'")
	if err != nil {
		return nil, err
	}

	// Parse subject expression
	subject, err := p.subjectExpr()
	if err != nil {
		return nil, err
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after match subject")
	if err != nil {
		return nil, err
	}

	// Consume NEWLINE
	_, err = p.consume(lexer.Newline, "expected newline after ':'")
	if err != nil {
		return nil, err
	}

	// Consume all remaining newlines (handles comments)
	for p.check(lexer.Newline) {
		p.advance()
	}

	// Consume INDENT
	_, err = p.consume(lexer.Indent, "expected indented block after match")
	if err != nil {
		return nil, err
	}

	// Parse one or more case blocks using HTML context parsing
	var cases []ast.CaseBlock
	for p.check(lexer.Case) {
		caseBlock, err := p.htmlCaseBlock()
		if err != nil {
			return nil, err
		}
		cases = append(cases, caseBlock)

		// Skip newlines between case blocks
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	if len(cases) == 0 {
		return nil, p.error(p.peek(), "expected at least one case block")
	}

	// Consume DEDENT
	dedentToken, err := p.consume(lexer.Dedent, "expected dedent after match cases")
	if err != nil {
		return nil, err
	}

	return &ast.MatchStmt{
		Subject: subject,
		Cases:   cases,
		Span:    lexer.Span{Start: matchToken.Start(), End: dedentToken.End()},
	}, nil
}

// htmlCaseBlock parses a case block within HTML content context.
func (p *Parser) htmlCaseBlock() (ast.CaseBlock, error) {
	caseToken, err := p.consume(lexer.Case, "expected 'case'")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse patterns
	patterns, err := p.patterns()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse optional guard
	var guard ast.Expr
	if p.match(lexer.If) {
		guard, err = p.namedExpression()
		if err != nil {
			return ast.CaseBlock{}, err
		}
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after case pattern")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse block using HTML content parsing (no function definitions allowed)
	body, err := p.htmlContentBlock()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	return ast.CaseBlock{
		Patterns: patterns,
		Guard:    guard,
		Body:     body,
		Span:     lexer.Span{Start: caseToken.Start(), End: body[len(body)-1].GetSpan().End},
	}, nil
}
