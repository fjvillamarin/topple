package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Module is the root node of a program, containing a list of statements.
type Module struct {
	Body []Stmt

	span lexer.Span
}

func NewModule(body []Stmt, span lexer.Span) *Module {
	return &Module{
		Body: body,

		span: span,
	}
}

func (m *Module) isStmt() {}

func (m *Module) Span() lexer.Span {
	return m.span
}

// Accept calls the VisitModule method on the visitor
func (m *Module) Accept(visitor Visitor) {
	visitor.VisitModule(m)
}

func (m *Module) String() string {
	return fmt.Sprintf("Module(%d stmts)", len(m.Body))
}
