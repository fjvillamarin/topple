package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// RaiseStmt represents a 'raise' statement.
type RaiseStmt struct {
	Exception    Expr
	FromExpr     Expr
	HasException bool
	HasFrom      bool

	span lexer.Span
}

func NewRaiseStmt(exception Expr, fromExpr Expr, hasException bool, hasFrom bool, span lexer.Span) *RaiseStmt {
	return &RaiseStmt{
		Exception:    exception,
		FromExpr:     fromExpr,
		HasException: hasException,
		HasFrom:      hasFrom,

		span: span,
	}
}

func (r *RaiseStmt) isStmt() {}

func (r *RaiseStmt) Span() lexer.Span {
	return r.span
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
