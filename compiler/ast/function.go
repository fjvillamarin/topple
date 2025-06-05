package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

type Function struct {
	Name           *Name
	TypeParameters []*TypeParam
	ReturnType     Expr
	Parameters     *ParameterList
	Body           []Stmt
	IsAsync        bool

	Span lexer.Span
}

func (f *Function) isStmt() {}

func (f *Function) GetSpan() lexer.Span {
	return f.Span
}

func (f *Function) Accept(visitor Visitor) {
	visitor.VisitFunction(f)
}

func (f *Function) String() string {
	return fmt.Sprintf("def %s%s%s", f.Name, f.Parameters.String(), f.ReturnType.String())
}
