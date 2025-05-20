package nodes

import (
	"biscuit/compiler/lexer"
)

// Node is the base interface for all AST nodes.
type Node interface {
	Accept(visitor Visitor)

	Span() lexer.Span
	String() string
}

// Expr is the interface for all expression nodes.
type Expr interface {
	Node
	isExpr()
} // only *expressions* satisfy this

// Stmt is the interface for all statement nodes.
type Stmt interface {
	Node
	isStmt()
} // only *statements* satisfy this

// Visitor is the interface for visitors that traverse the AST.
type Visitor interface {
	ExprVisitor
	StmtVisitor
}

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
	VisitAwaitExpr(a *AwaitExpr) Visitor
}

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	Visit(node Node) Visitor
	VisitModule(m *Module) Visitor
	VisitExprStmt(e *ExprStmt) Visitor
	VisitTypeAlias(t *TypeAlias) Visitor
	VisitReturnStmt(r *ReturnStmt) Visitor
	VisitRaiseStmt(r *RaiseStmt) Visitor
	VisitPassStmt(p *PassStmt) Visitor
	VisitBreakStmt(b *BreakStmt) Visitor
	VisitContinueStmt(c *ContinueStmt) Visitor
	VisitYieldStmt(y *YieldStmt) Visitor
	VisitAssertStmt(a *AssertStmt) Visitor
	VisitGlobalStmt(g *GlobalStmt) Visitor
	VisitNonlocalStmt(n *NonlocalStmt) Visitor
	VisitImportStmt(i *ImportStmt) Visitor
	VisitImportFromStmt(i *ImportFromStmt) Visitor
	VisitAssignStmt(a *AssignStmt) Visitor
	VisitAugAssignStmt(a *AugAssignStmt) Visitor
	VisitAnnotationStmt(a *AnnotationStmt) Visitor
	VisitMultiStmt(m *MultiStmt) Visitor
}
