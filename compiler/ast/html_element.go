package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

// HTMLElementType represents the type of HTML element
type HTMLElementType int

const (
	HTMLOpenTag           HTMLElementType = iota // <div>
	HTMLCloseTag                                 // </div>
	HTMLSelfClosingTag                           // <img />
	HTMLMultilineElement                         // <div>...content...</div> (multiline)
	HTMLSingleLineElement                        // <span>content</span> (single line)
)

// HTMLAttribute represents an HTML attribute
type HTMLAttribute struct {
	Name  lexer.Token // Attribute name
	Value Expr        // Attribute value (can be string literal or expression)
	Span  lexer.Span
}

// HTMLElement represents an HTML element statement
type HTMLElement struct {
	Type       HTMLElementType // Type of HTML element
	TagName    lexer.Token     // Tag name (e.g., "div", "span")
	Attributes []HTMLAttribute // HTML attributes
	Content    []Stmt          // Content inside the element (for container elements)
	IsClosing  bool            // Whether this is a closing tag

	Span lexer.Span
}

func (h *HTMLElement) isStmt() {}

func (h *HTMLElement) GetSpan() lexer.Span {
	return h.Span
}

func (h *HTMLElement) Accept(visitor Visitor) {
	visitor.VisitHTMLElement(h)
}

func (h *HTMLElement) String() string {
	switch h.Type {
	case HTMLOpenTag:
		return fmt.Sprintf("HTMLOpenTag(<%s>)", h.TagName.Lexeme)
	case HTMLCloseTag:
		return fmt.Sprintf("HTMLCloseTag(</%s>)", h.TagName.Lexeme)
	case HTMLSelfClosingTag:
		return fmt.Sprintf("HTMLSelfClosingTag(<%s />)", h.TagName.Lexeme)
	case HTMLMultilineElement:
		return fmt.Sprintf("HTMLMultilineElement(<%s>...content...</%s>)", h.TagName.Lexeme, h.TagName.Lexeme)
	case HTMLSingleLineElement:
		return fmt.Sprintf("HTMLSingleLineElement(<%s>content</%s>)", h.TagName.Lexeme, h.TagName.Lexeme)
	default:
		return fmt.Sprintf("HTMLElement(%s)", h.TagName.Lexeme)
	}
}
