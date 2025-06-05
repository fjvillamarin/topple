package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
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
	switch l.Type {
	case LiteralTypeString:
		return fmt.Sprintf("\"%s\"", l.Value)
	case LiteralTypeNumber:
		// Handle numeric literals properly
		switch v := l.Value.(type) {
		case int:
			return fmt.Sprintf("%d", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%g", v)
		case string:
			// If stored as string, use it directly
			return v
		default:
			// Fallback - should not reach here if lexer works correctly
			return fmt.Sprintf("%v", v)
		}
	case LiteralTypeBool:
		switch v := l.Value.(type) {
		case bool:
			if v {
				return "True"
			} else {
				return "False"
			}
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	case LiteralTypeNone:
		return "None"
	default:
		return fmt.Sprintf("%v", l.Value)
	}
}

// Accept calls the VisitLiteral method on the visitor
func (l *Literal) Accept(visitor Visitor) {
	visitor.VisitLiteral(l)
}
