package ast

import (
	"fmt"
	"strings"
	"topple/compiler/lexer"
)

// DictExpr represents a dictionary literal: {key1: value1, key2: value2, **dict_expr}
type DictExpr struct {
	Pairs []DictPair // Key-value pairs and starred expressions

	Span lexer.Span
}

// DictPair represents either a key-value pair or a double-starred expression
type DictPair interface {
	isDictPair()
	GetSpan() lexer.Span
	String() string
}

// KeyValuePair represents a key: value pair in a dictionary
type KeyValuePair struct {
	Key   Expr // The key expression
	Value Expr // The value expression

	Span lexer.Span
}

// DoubleStarredPair represents **expression in a dictionary (dictionary unpacking)
type DoubleStarredPair struct {
	Expr Expr // The expression to unpack

	Span lexer.Span
}

// DictExpr methods
func (d *DictExpr) isExpr() {}

func (d *DictExpr) GetSpan() lexer.Span {
	return d.Span
}

func (d *DictExpr) String() string {
	if len(d.Pairs) == 0 {
		return "{}"
	}

	var pairs []string
	for _, pair := range d.Pairs {
		pairs = append(pairs, pair.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (d *DictExpr) Accept(visitor Visitor) {
	visitor.VisitDictExpr(d)
}

// KeyValuePair methods
func (kvp *KeyValuePair) isDictPair() {}

func (kvp *KeyValuePair) GetSpan() lexer.Span {
	return kvp.Span
}

func (kvp *KeyValuePair) String() string {
	return fmt.Sprintf("%s: %s", kvp.Key.String(), kvp.Value.String())
}

// DoubleStarredPair methods
func (dsp *DoubleStarredPair) isDictPair() {}

func (dsp *DoubleStarredPair) GetSpan() lexer.Span {
	return dsp.Span
}

func (dsp *DoubleStarredPair) String() string {
	return fmt.Sprintf("**%s", dsp.Expr.String())
}
