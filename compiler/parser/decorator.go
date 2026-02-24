package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// decorator parses a decorator followed by a statement
// @decorator
// statement
func (p *Parser) decorator() (ast.Stmt, error) {
	// Consume the '@' token
	atToken := p.advance()

	// Parse the decorator expression (should be a named_expression)
	expr, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect a newline after the decorator
	_, err = p.consume(lexer.Newline, "expected newline after decorator")
	if err != nil {
		return nil, err
	}

	// Parse the statement that follows (which could be another decorator or a class/function)
	decorated, err := p.statement()
	if err != nil {
		return nil, err
	}

	// Only classes and functions can be decorated
	if !isDecoratable(decorated) {
		return nil, p.error(p.previous(), "only class and function definitions can be decorated")
	}

	// Create a decorator node that wraps the decorated statement
	return &ast.Decorator{
		Expr: expr,
		Stmt: decorated,
		Span: lexer.Span{Start: atToken.Start(), End: decorated.GetSpan().End},
	}, nil
}

// isDecoratable checks if a statement can be decorated (classes and functions)
func isDecoratable(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.Class, *ast.Function, *ast.Decorator:
		// Classes, functions, and other decorators can be decorated
		return true
	default:
		return false
	}
}
