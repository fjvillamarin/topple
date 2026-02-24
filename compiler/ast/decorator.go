package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

type Decorator struct {
	Expr Expr
	Stmt Stmt

	Span lexer.Span
}

func (d *Decorator) isStmt() {}

func (d *Decorator) GetSpan() lexer.Span {
	return d.Span
}

func (d *Decorator) Accept(visitor Visitor) {
	visitor.VisitDecorator(d)
}

func (d *Decorator) String() string {
	return fmt.Sprintf("Decorator(%s)", d.Expr)
}
