package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

func (p *Parser) elseBlock() ([]ast.Stmt, error) {
	// Check if there's an else block
	if !p.match(lexer.Else) {
		return nil, nil
	}

	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'else'")
	if err != nil {
		return nil, err
	}

	// Parse else body
	elseBody, err := p.block()
	if err != nil {
		return nil, err
	}

	return elseBody, nil
}
