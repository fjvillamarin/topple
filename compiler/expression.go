package compiler

import (
	"fmt"
)

// ExprVisitor is the interface for visitors that traverse expressions.
type ExprVisitor interface {
	VisitName(n *Name) Visitor
	VisitLiteral(l *Literal) Visitor
	VisitAttribute(a *Attribute) Visitor
	VisitCall(c *Call) Visitor
	VisitSubscript(s *Subscript) Visitor
	VisitBinary(b *Binary) Visitor
	VisitUnary(u *Unary) Visitor
	VisitAssignExpr(a *AssignExpr) Visitor
	VisitStarExpr(s *StarExpr) Visitor
	VisitTernaryExpr(t *TernaryExpr) Visitor
	VisitListExpr(l *ListExpr) Visitor
	VisitTupleExpr(t *TupleExpr) Visitor
	VisitSetExpr(s *SetExpr) Visitor
	VisitYieldExpr(y *YieldExpr) Visitor
	VisitGroupExpr(g *GroupExpr) Visitor
	VisitTypeParamExpr(t *TypeParamExpr) Visitor
	VisitSlice(s *Slice) Visitor
}

// Name represents an identifier expression.
type Name struct {
	BaseNode
	Token Token
}

func NewName(token Token, startPos Position, endPos Position) *Name {
	return &Name{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Token: token,
	}
}

func (n *Name) isExpr() {}

func (n *Name) String() string {
	return n.Token.Lexeme
}

// Accept calls the VisitName method on the visitor
func (n *Name) Accept(visitor Visitor) {
	visitor.VisitName(n)
}

// Literal represents a literal value (number, string, etc.).
type Literal struct {
	BaseNode
	Token Token
	Value any
}

func NewLiteral(token Token, value any, startPos Position, endPos Position) *Literal {
	return &Literal{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Token: token,
		Value: value,
	}
}

func (l *Literal) isExpr() {}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Accept calls the VisitLiteral method on the visitor
func (l *Literal) Accept(visitor Visitor) {
	visitor.VisitLiteral(l)
}

// Attribute represents an attribute access expression (obj.attr)
type Attribute struct {
	BaseNode
	Object Expr
	Name   Token
}

func NewAttribute(object Expr, name Token, startPos Position, endPos Position) *Attribute {
	return &Attribute{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Object: object,
		Name:   name,
	}
}

func (a *Attribute) isExpr() {}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v.%s", a.Object, a.Name.Lexeme)
}

// Accept calls the VisitAttribute method on the visitor
func (a *Attribute) Accept(visitor Visitor) {
	visitor.VisitAttribute(a)
}

// Call represents a function call expression (func(args))
type Call struct {
	BaseNode
	Callee    Expr
	Arguments []Expr
}

func NewCall(callee Expr, arguments []Expr, startPos Position, endPos Position) *Call {
	return &Call{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Callee:    callee,
		Arguments: arguments,
	}
}

func (c *Call) isExpr() {}

func (c *Call) String() string {
	return fmt.Sprintf("%v()", c.Callee)
}

// Accept calls the VisitCall method on the visitor
func (c *Call) Accept(visitor Visitor) {
	visitor.VisitCall(c)
}

// Subscript represents a subscript access expression (obj[index] or obj[start:end:step])
type Subscript struct {
	BaseNode
	Object  Expr
	Indices []Expr // Multiple indices or slices
}

func NewSubscript(object Expr, indices []Expr, startPos Position, endPos Position) *Subscript {
	return &Subscript{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Object:  object,
		Indices: indices,
	}
}

func (s *Subscript) isExpr() {}

func (s *Subscript) String() string {
	return fmt.Sprintf("%v[...]", s.Object)
}

// Accept calls the VisitSubscript method on the visitor
func (s *Subscript) Accept(visitor Visitor) {
	visitor.VisitSubscript(s)
}

// Binary represents a binary operation expression (left op right)
type Binary struct {
	BaseNode
	Left     Expr
	Operator Token
	Right    Expr
}

func NewBinary(left Expr, operator Token, right Expr, startPos Position, endPos Position) *Binary {
	return &Binary{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (b *Binary) isExpr() {}

func (b *Binary) String() string {
	return fmt.Sprintf("%v %s %v", b.Left, b.Operator.Lexeme, b.Right)
}

// Accept calls the VisitBinary method on the visitor
func (b *Binary) Accept(visitor Visitor) {
	visitor.VisitBinary(b)
}

// Unary represents a unary operation expression (-expr)
type Unary struct {
	BaseNode
	Operator Token
	Right    Expr
}

func NewUnary(operator Token, right Expr, startPos Position, endPos Position) *Unary {
	return &Unary{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Operator: operator,
		Right:    right,
	}
}

func (u *Unary) isExpr() {}

func (u *Unary) String() string {
	return fmt.Sprintf("%s %v", u.Operator.Lexeme, u.Right)
}

// Accept calls the VisitUnary method on the visitor
func (u *Unary) Accept(visitor Visitor) {
	visitor.VisitUnary(u)
}

// AssignExpr represents an assignment expression (left = right)
type AssignExpr struct {
	BaseNode
	Left  Expr
	Right Expr
}

func NewAssignExpr(left Expr, right Expr, startPos Position, endPos Position) *AssignExpr {
	return &AssignExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Left:  left,
		Right: right,
	}
}

func (a *AssignExpr) isExpr() {}

func (a *AssignExpr) String() string {
	return fmt.Sprintf("%v = %v", a.Left, a.Right)
}

// Accept calls the VisitAssignExpr method on the visitor
func (a *AssignExpr) Accept(visitor Visitor) {
	visitor.VisitAssignExpr(a)
}

// StarExpr represents a star expression (*expr)
type StarExpr struct {
	BaseNode
	Expr Expr
}

func NewStarExpr(expr Expr, startPos Position, endPos Position) *StarExpr {
	return &StarExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Expr: expr,
	}
}

func (s *StarExpr) isExpr() {}

func (s *StarExpr) String() string {
	return fmt.Sprintf("*%v", s.Expr)
}

// Accept calls the VisitStarExpr method on the visitor
func (s *StarExpr) Accept(visitor Visitor) {
	visitor.VisitStarExpr(s)
}

// TernaryExpr represents a ternary expression (if condition then true else false)
type TernaryExpr struct {
	BaseNode
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func NewTernaryExpr(condition Expr, trueExpr Expr, falseExpr Expr, startPos Position, endPos Position) *TernaryExpr {
	return &TernaryExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Condition: condition,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

func (t *TernaryExpr) isExpr() {}

func (t *TernaryExpr) String() string {
	return fmt.Sprintf("%v ? %v : %v", t.Condition, t.TrueExpr, t.FalseExpr)
}

// Accept calls the VisitTernaryExpr method on the visitor
func (t *TernaryExpr) Accept(visitor Visitor) {
	visitor.VisitTernaryExpr(t)
}

// ListExpr represents a list expression [items]
type ListExpr struct {
	BaseNode
	Elements []Expr
}

func NewListExpr(elements []Expr, startPos Position, endPos Position) *ListExpr {
	return &ListExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Elements: elements,
	}
}

func (l *ListExpr) isExpr() {}

func (l *ListExpr) String() string {
	return fmt.Sprintf("[...]")
}

// Accept calls the VisitListExpr method on the visitor
func (l *ListExpr) Accept(visitor Visitor) {
	visitor.VisitListExpr(l)
}

// TupleExpr represents a tuple expression (items)
type TupleExpr struct {
	BaseNode
	Elements []Expr
}

func NewTupleExpr(elements []Expr, startPos Position, endPos Position) *TupleExpr {
	return &TupleExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Elements: elements,
	}
}

func (t *TupleExpr) isExpr() {}

func (t *TupleExpr) String() string {
	return fmt.Sprintf("(...)")
}

// Accept calls the VisitTupleExpr method on the visitor
func (t *TupleExpr) Accept(visitor Visitor) {
	visitor.VisitTupleExpr(t)
}

// SetExpr represents a set expression {items}
type SetExpr struct {
	BaseNode
	Elements []Expr
}

func NewSetExpr(elements []Expr, startPos Position, endPos Position) *SetExpr {
	return &SetExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Elements: elements,
	}
}

func (s *SetExpr) isExpr() {}

func (s *SetExpr) String() string {
	return fmt.Sprintf("{...}")
}

// Accept calls the VisitSetExpr method on the visitor
func (s *SetExpr) Accept(visitor Visitor) {
	visitor.VisitSetExpr(s)
}

// YieldExpr represents a yield expression (yield value)
type YieldExpr struct {
	BaseNode
	IsFrom bool
	Value  Expr
}

func NewYieldExpr(isFrom bool, value Expr, startPos Position, endPos Position) *YieldExpr {
	return &YieldExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		IsFrom: isFrom,
		Value:  value,
	}
}

func (y *YieldExpr) isExpr() {}

func (y *YieldExpr) String() string {
	if y.IsFrom {
		return fmt.Sprintf("yield from %v", y.Value)
	}
	return fmt.Sprintf("yield %v", y.Value)
}

// Accept calls the VisitYieldExpr method on the visitor
func (y *YieldExpr) Accept(visitor Visitor) {
	visitor.VisitYieldExpr(y)
}

// GroupExpr represents a parenthesized expression (expr)
type GroupExpr struct {
	BaseNode
	Expression Expr
}

func NewGroupExpr(expression Expr, startPos Position, endPos Position) *GroupExpr {
	return &GroupExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Expression: expression,
	}
}

func (g *GroupExpr) isExpr() {}

func (g *GroupExpr) String() string {
	return fmt.Sprintf("(%v)", g.Expression)
}

// Accept calls the VisitGroupExpr method on the visitor
func (g *GroupExpr) Accept(visitor Visitor) {
	visitor.VisitGroupExpr(g)
}

// TypeParamExpr represents a type parameter expression
type TypeParamExpr struct {
	BaseNode
	Name         Token
	Bound        Expr // Optional bound (: expression)
	Default      Expr // Optional default (= expression)
	IsStar       bool // Whether this is a *NAME parameter
	IsDoubleStar bool // Whether this is a **NAME parameter
}

func NewTypeParamExpr(name Token, bound Expr, defaultValue Expr, isStar bool, isDoubleStar bool,
	startPos Position, endPos Position) *TypeParamExpr {
	return &TypeParamExpr{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Name:         name,
		Bound:        bound,
		Default:      defaultValue,
		IsStar:       isStar,
		IsDoubleStar: isDoubleStar,
	}
}

func (t *TypeParamExpr) isExpr() {}

// Accept calls the VisitTypeParamExpr method on the visitor
func (t *TypeParamExpr) Accept(visitor Visitor) {
	visitor.VisitTypeParamExpr(t)
}

func (t *TypeParamExpr) String() string {
	prefix := ""
	if t.IsStar {
		prefix = "*"
	} else if t.IsDoubleStar {
		prefix = "**"
	}
	return fmt.Sprintf("%s%s", prefix, t.Name.Lexeme)
}

// Slice represents a slice expression (start:end:step)
type Slice struct {
	BaseNode
	StartIndex Expr // Optional start index
	EndIndex   Expr // Optional end index
	Step       Expr // Optional step
}

func NewSlice(start Expr, end Expr, step Expr, startPos Position, endPos Position) *Slice {
	return &Slice{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		StartIndex: start,
		EndIndex:   end,
		Step:       step,
	}
}

func (s *Slice) isExpr() {}

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

// Start returns the start position of the node
func (s *Slice) Start() Position {
	return s.StartPos
}

// End returns the end position of the node
func (s *Slice) End() Position {
	return s.EndPos
}

// Accept calls the VisitSlice method on the visitor
func (s *Slice) Accept(visitor Visitor) {
	visitor.VisitSlice(s)
}
