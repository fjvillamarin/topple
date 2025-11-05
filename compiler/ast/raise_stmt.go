package ast

import (
	"fmt"
	"topple/compiler/lexer"
)

// RaiseStmt represents a 'raise' statement.
type RaiseStmt struct {
	Exception    Expr
	FromExpr     Expr
	HasException bool
	HasFrom      bool

	Span lexer.Span
}

func (r *RaiseStmt) isStmt() {}

func (r *RaiseStmt) GetSpan() lexer.Span {
	return r.Span
}

func (r *RaiseStmt) Accept(visitor Visitor) {
	visitor.VisitRaiseStmt(r)
}

func (r *RaiseStmt) String() string {
	if r.HasException {
		if r.HasFrom {
			return fmt.Sprintf("RaiseStmt(%s from %s)", r.Exception, r.FromExpr)
		}
		return fmt.Sprintf("RaiseStmt(%s)", r.Exception)
	}
	return "RaiseStmt()"
}
