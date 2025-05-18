package compiler

import "fmt"

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	VisitModule(m *Module) Visitor
	VisitExprStmt(e *ExprStmt) Visitor
	VisitTypeAlias(t *TypeAlias) Visitor
	VisitReturnStmt(r *ReturnStmt) Visitor
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
