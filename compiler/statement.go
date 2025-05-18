package compiler

import "fmt"

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	VisitModule(m *Module) Visitor
	VisitExprStmt(e *ExprStmt) Visitor
	VisitTypeAlias(t *TypeAlias) Visitor
	VisitReturnStmt(r *ReturnStmt) Visitor
	VisitRaiseStmt(r *RaiseStmt) Visitor
	VisitPassStmt(p *PassStmt) Visitor
	VisitBreakStmt(b *BreakStmt) Visitor
	VisitContinueStmt(c *ContinueStmt) Visitor
}

// Module is the root node of a program, containing a list of statements.
type Module struct {
	BaseNode
	Body []Stmt
}

func NewModule(body []Stmt, startPos Position, endPos Position) *Module {
	return &Module{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Body: body,
	}
}

func (m *Module) isStmt() {}

// Accept calls the VisitModule method on the visitor
func (m *Module) Accept(visitor Visitor) {
	visitor.VisitModule(m)
}

func (m *Module) String() string {
	return fmt.Sprintf("Module(%d stmts)", len(m.Body))
}

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	BaseNode
	Value Expr
}

func NewExprStmt(value Expr, startPos Position, endPos Position) *ExprStmt {
	return &ExprStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Value: value,
	}
}

func (e *ExprStmt) isStmt() {}

// Accept calls the VisitExprStmt method on the visitor
func (e *ExprStmt) Accept(visitor Visitor) {
	visitor.VisitExprStmt(e)
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%s)", e.Value)
}

// TypeAlias represents a 'type' statement.
type TypeAlias struct {
	BaseNode
	Name   Token
	Params []Expr // Type parameters as expressions
	Value  Expr   // The type value expression
}

func NewTypeAlias(name Token, params []Expr, value Expr, startPos Position, endPos Position) *TypeAlias {
	return &TypeAlias{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Name:   name,
		Params: params,
		Value:  value,
	}
}

func (t *TypeAlias) isStmt() {}

// Accept calls the VisitTypeAlias method on the visitor
func (t *TypeAlias) Accept(visitor Visitor) {
	visitor.VisitTypeAlias(t)
}

func (t *TypeAlias) String() string {
	return fmt.Sprintf("TypeAlias(%s)", t.Name.Lexeme)
}

// ReturnStmt represents a 'return' statement.
type ReturnStmt struct {
	BaseNode
	Value Expr
}

func NewReturnStmt(value Expr, startPos Position, endPos Position) *ReturnStmt {
	return &ReturnStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Value: value,
	}
}

func (r *ReturnStmt) isStmt() {}

func (r *ReturnStmt) Accept(visitor Visitor) {
	visitor.VisitReturnStmt(r)
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt(%s)", r.Value)
}

// RaiseStmt represents a 'raise' statement.
type RaiseStmt struct {
	BaseNode
	Exception    Expr
	FromExpr     Expr
	HasException bool
	HasFrom      bool
}

func NewRaiseStmt(exception Expr, fromExpr Expr, hasException bool, hasFrom bool, startPos Position, endPos Position) *RaiseStmt {
	return &RaiseStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Exception:    exception,
		FromExpr:     fromExpr,
		HasException: hasException,
		HasFrom:      hasFrom,
	}
}

func (r *RaiseStmt) isStmt() {}

func (r *RaiseStmt) Accept(visitor Visitor) {
	visitor.VisitRaiseStmt(r)
}

func (r *RaiseStmt) String() string {
	if r.HasException {
		if r.HasFrom {
			return fmt.Sprintf("RaiseStmt(%s from %s)", r.Exception, r.FromExpr)
		}
		return fmt.Sprintf("RaiseStmt(%s)", r.Exception)
	}
	return "RaiseStmt()"
}

// PassStmt represents a 'pass' statement, which is a no-op.
type PassStmt struct {
	BaseNode
}

func NewPassStmt(startPos Position, endPos Position) *PassStmt {
	return &PassStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (p *PassStmt) isStmt() {}

func (p *PassStmt) Accept(visitor Visitor) {
	visitor.VisitPassStmt(p)
}

func (p *PassStmt) String() string {
	return "PassStmt()"
}

// BreakStmt represents a 'break' statement.
type BreakStmt struct {
	BaseNode
}

func NewBreakStmt(startPos Position, endPos Position) *BreakStmt {
	return &BreakStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (b *BreakStmt) isStmt() {}

func (b *BreakStmt) Accept(visitor Visitor) {
	visitor.VisitBreakStmt(b)
}

func (b *BreakStmt) String() string {
	return "BreakStmt()"
}

// ContinueStmt represents a 'continue' statement.
type ContinueStmt struct {
	BaseNode
}

func NewContinueStmt(startPos Position, endPos Position) *ContinueStmt {
	return &ContinueStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (c *ContinueStmt) isStmt() {}

func (c *ContinueStmt) Accept(visitor Visitor) {
	visitor.VisitContinueStmt(c)
}

func (c *ContinueStmt) String() string {
	return "ContinueStmt()"
}
