package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

type ViewKind string

const (
	ViewKindServerView ViewKind = "server_view"
	ViewKindClientView ViewKind = "client_view"
)

// ViewStmt represents a view definition statement
type ViewStmt struct {
	Name       *Name          // View name
	TypeParams []*TypeParam   // Type parameters
	Params     *ParameterList // Function parameters
	ReturnType Expr           // Return type annotation (optional)
	Body       []Stmt         // View body containing Python statements and HTML elements
	IsAsync    bool           // Whether it's an async view
	Kind       ViewKind       // View kind

	Span lexer.Span
}

func (v *ViewStmt) isStmt() {}

func (v *ViewStmt) GetSpan() lexer.Span {
	return v.Span
}

func (v *ViewStmt) Accept(visitor Visitor) {
	visitor.VisitViewStmt(v)
}

func (v *ViewStmt) String() string {
	return fmt.Sprintf("ViewStmt(%s)", v.Name.Token.Lexeme)
}
