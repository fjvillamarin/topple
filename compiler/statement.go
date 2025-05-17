package compiler

import "fmt"

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	VisitModule(m *Module) Visitor
	VisitExprStmt(e *ExprStmt) Visitor
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
