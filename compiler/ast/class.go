package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

type Class struct {
	Name       *Name
	TypeParams []TypeParam
	Args       []*Argument
	Body       []Stmt

	Span lexer.Span
}

func (c *Class) isStmt() {}

func (c *Class) GetSpan() lexer.Span {
	return c.Span
}

func (c *Class) Accept(visitor Visitor) {
	visitor.VisitClass(c)
}

func (c *Class) String() string {
	return fmt.Sprintf("Class(%s)", c.Name)
}
