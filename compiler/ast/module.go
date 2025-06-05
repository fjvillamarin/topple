package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// Module is the root node of a program, containing a list of statements.
type Module struct {
	Body []Stmt

	Span lexer.Span
}

func (m *Module) isStmt() {}

func (m *Module) GetSpan() lexer.Span {
	return m.Span
}

// Accept calls the VisitModule method on the visitor
func (m *Module) Accept(visitor Visitor) {
	visitor.VisitModule(m)
}

func (m *Module) String() string {
	return fmt.Sprintf("Module(%d stmts)", len(m.Body))
}
