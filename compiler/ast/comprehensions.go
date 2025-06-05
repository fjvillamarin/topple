package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
	"strings"
)

// ListComp represents a list comprehension: [expression for_if_clauses]
type ListComp struct {
	Element Expr          // The expression to evaluate for each iteration
	Clauses []ForIfClause // One or more for/if clauses

	Span lexer.Span
}

func (lc *ListComp) isExpr() {}

func (lc *ListComp) GetSpan() lexer.Span {
	return lc.Span
}

func (lc *ListComp) String() string {
	var clauses []string
	for _, clause := range lc.Clauses {
		clauses = append(clauses, clause.String())
	}
	return fmt.Sprintf("[%s %s]", lc.Element.String(), strings.Join(clauses, " "))
}

func (lc *ListComp) Accept(visitor Visitor) {
	visitor.VisitListComp(lc)
}

// SetComp represents a set comprehension: {expression for_if_clauses}
type SetComp struct {
	Element Expr          // The expression to evaluate for each iteration
	Clauses []ForIfClause // One or more for/if clauses

	Span lexer.Span
}

func (sc *SetComp) isExpr() {}

func (sc *SetComp) GetSpan() lexer.Span {
	return sc.Span
}

func (sc *SetComp) String() string {
	var clauses []string
	for _, clause := range sc.Clauses {
		clauses = append(clauses, clause.String())
	}
	return fmt.Sprintf("{%s %s}", sc.Element.String(), strings.Join(clauses, " "))
}

func (sc *SetComp) Accept(visitor Visitor) {
	visitor.VisitSetComp(sc)
}

// DictComp represents a dictionary comprehension: {key: value for_if_clauses}
type DictComp struct {
	Key     Expr          // The key expression
	Value   Expr          // The value expression
	Clauses []ForIfClause // One or more for/if clauses

	Span lexer.Span
}

func (dc *DictComp) isExpr() {}

func (dc *DictComp) GetSpan() lexer.Span {
	return dc.Span
}

func (dc *DictComp) String() string {
	var clauses []string
	for _, clause := range dc.Clauses {
		clauses = append(clauses, clause.String())
	}
	return fmt.Sprintf("{%s: %s %s}", dc.Key.String(), dc.Value.String(), strings.Join(clauses, " "))
}

func (dc *DictComp) Accept(visitor Visitor) {
	visitor.VisitDictComp(dc)
}

// GenExpr represents a generator expression: (expression for_if_clauses)
type GenExpr struct {
	Element Expr          // The expression to evaluate for each iteration
	Clauses []ForIfClause // One or more for/if clauses

	Span lexer.Span
}

func (ge *GenExpr) isExpr() {}

func (ge *GenExpr) GetSpan() lexer.Span {
	return ge.Span
}

func (ge *GenExpr) String() string {
	var clauses []string
	for _, clause := range ge.Clauses {
		clauses = append(clauses, clause.String())
	}
	return fmt.Sprintf("(%s %s)", ge.Element.String(), strings.Join(clauses, " "))
}

func (ge *GenExpr) Accept(visitor Visitor) {
	visitor.VisitGenExpr(ge)
}
