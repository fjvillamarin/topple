package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

type LiteralType int

const (
	LiteralTypeString LiteralType = iota
	LiteralTypeNumber
	LiteralTypeBool
	LiteralTypeNone
)

// Literal represents a literal value (number, string, etc.).
type Literal struct {
	Token lexer.Token
	Value any
	Type  LiteralType

	Span lexer.Span
}

func (l *Literal) isExpr() {}

func (l *Literal) GetSpan() lexer.Span {
	return l.Span
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Accept calls the VisitLiteral method on the visitor
func (l *Literal) Accept(visitor Visitor) {
	visitor.VisitLiteral(l)
}
