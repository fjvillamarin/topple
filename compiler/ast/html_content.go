package ast

import (
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
)

// HTMLContent represents HTML content with mixed text and interpolations: Hello {name}!
type HTMLContent struct {
	Parts []HTMLContentPart // The parts of the HTML content (text and interpolations)

	Span lexer.Span
}

func (h *HTMLContent) isStmt() {}

func (h *HTMLContent) GetSpan() lexer.Span {
	return h.Span
}

func (h *HTMLContent) String() string {
	var parts []string
	for _, part := range h.Parts {
		parts = append(parts, part.String())
	}
	return strings.Join(parts, "")
}

func (h *HTMLContent) Accept(visitor Visitor) {
	visitor.VisitHTMLContent(h)
}

// HTMLContentPart is the interface for parts of HTML content
type HTMLContentPart interface {
	Node
	isHTMLContentPart()
}

// HTMLText represents literal text in HTML content
type HTMLText struct {
	Value string // The literal text

	Span lexer.Span
}

func (h *HTMLText) isHTMLContentPart() {}

func (h *HTMLText) GetSpan() lexer.Span {
	return h.Span
}

func (h *HTMLText) String() string {
	return h.Value
}

func (h *HTMLText) Accept(visitor Visitor) {
	visitor.VisitHTMLText(h)
}

// HTMLInterpolation represents an interpolated expression in HTML content: {expression}
type HTMLInterpolation struct {
	Expression Expr // The expression to evaluate

	Span lexer.Span
}

func (h *HTMLInterpolation) isHTMLContentPart() {}

func (h *HTMLInterpolation) GetSpan() lexer.Span {
	return h.Span
}

func (h *HTMLInterpolation) String() string {
	return "{" + h.Expression.String() + "}"
}

func (h *HTMLInterpolation) Accept(visitor Visitor) {
	visitor.VisitHTMLInterpolation(h)
}
