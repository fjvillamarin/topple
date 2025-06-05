package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

// Slice represents a slice expression (start:end:step)
type Slice struct {
	StartIndex Expr // Optional start index
	EndIndex   Expr // Optional end index
	Step       Expr // Optional step

	Span lexer.Span
}

func (s *Slice) isExpr() {}

func (s *Slice) GetSpan() lexer.Span {
	return s.Span
}

func (s *Slice) String() string {
	var startStr, endStr, stepStr string
	if s.StartIndex != nil {
		startStr = s.StartIndex.String()
	}
	if s.EndIndex != nil {
		endStr = s.EndIndex.String()
	}
	if s.Step != nil {
		stepStr = ":" + s.Step.String()
	}
	return fmt.Sprintf("%s:%s%s", startStr, endStr, stepStr)
}

// Accept calls the VisitSlice method on the visitor
func (s *Slice) Accept(visitor Visitor) {
	visitor.VisitSlice(s)
}
