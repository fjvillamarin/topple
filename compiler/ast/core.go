package ast

import (
	"sylfie/compiler/lexer"
)

// Node is the base interface for all AST nodes.
type Node interface {
	Accept(visitor Visitor)

	GetSpan() lexer.Span
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

	// Generic visit method for fallback
	Visit(node Node) Visitor
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
	VisitDictExpr(d *DictExpr) Visitor
	VisitListComp(lc *ListComp) Visitor
	VisitSetComp(sc *SetComp) Visitor
	VisitDictComp(dc *DictComp) Visitor
	VisitGenExpr(ge *GenExpr) Visitor
	VisitYieldExpr(y *YieldExpr) Visitor
	VisitGroupExpr(g *GroupExpr) Visitor
	VisitParameterList(p *ParameterList) Visitor
	VisitParameter(p *Parameter) Visitor
	VisitTypeParamExpr(t *TypeParam) Visitor
	VisitSlice(s *Slice) Visitor
	VisitAwaitExpr(a *AwaitExpr) Visitor
	VisitArgument(a *Argument) Visitor
	VisitLambda(l *Lambda) Visitor
	VisitFString(f *FString) Visitor
	VisitFStringMiddle(f *FStringMiddle) Visitor
	VisitFStringReplacementField(f *FStringReplacementField) Visitor
	VisitFStringConversion(f *FStringConversion) Visitor
	VisitFStringFormatSpec(f *FStringFormatSpec) Visitor
	VisitFStringFormatMiddle(f *FStringFormatMiddle) Visitor
	VisitFStringFormatReplacementField(f *FStringFormatReplacementField) Visitor
}

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	VisitExprStmt(e *ExprStmt) Visitor
	VisitAssignStmt(a *AssignStmt) Visitor
	VisitAnnotationStmt(a *AnnotationStmt) Visitor
	VisitReturnStmt(r *ReturnStmt) Visitor
	VisitRaiseStmt(r *RaiseStmt) Visitor
	VisitPassStmt(p *PassStmt) Visitor
	VisitYieldStmt(y *YieldStmt) Visitor
	VisitAssertStmt(a *AssertStmt) Visitor
	VisitBreakStmt(b *BreakStmt) Visitor
	VisitContinueStmt(c *ContinueStmt) Visitor
	VisitGlobalStmt(g *GlobalStmt) Visitor
	VisitNonlocalStmt(n *NonlocalStmt) Visitor
	VisitImportStmt(i *ImportStmt) Visitor
	VisitImportFromStmt(i *ImportFromStmt) Visitor
	VisitTypeAlias(t *TypeAlias) Visitor
	VisitDecorator(d *Decorator) Visitor
	VisitModule(m *Module) Visitor
	VisitMultiStmt(m *MultiStmt) Visitor

	// Compound statements (using correct type names)
	VisitIf(i *If) Visitor
	VisitWhile(w *While) Visitor
	VisitFor(f *For) Visitor
	VisitWith(w *With) Visitor
	VisitTry(t *Try) Visitor
	VisitClass(c *Class) Visitor
	VisitFunction(f *Function) Visitor
	VisitMatch(m *MatchStmt) Visitor

	// Pattern visitors (for match statements)
	VisitLiteralPattern(lp *LiteralPattern) Visitor
	VisitCapturePattern(cp *CapturePattern) Visitor
	VisitWildcardPattern(wp *WildcardPattern) Visitor
	VisitValuePattern(vp *ValuePattern) Visitor
	VisitGroupPattern(gp *GroupPattern) Visitor
	VisitSequencePattern(sp *SequencePattern) Visitor
	VisitStarPattern(sp *StarPattern) Visitor
	VisitMappingPattern(mp *MappingPattern) Visitor
	VisitClassPattern(cp *ClassPattern) Visitor
	VisitAsPattern(ap *AsPattern) Visitor
	VisitOrPattern(op *OrPattern) Visitor

	// New Biscuit-specific nodes
	VisitViewStmt(v *ViewStmt) Visitor
	VisitHTMLElement(h *HTMLElement) Visitor
	VisitHTMLContent(h *HTMLContent) Visitor
	VisitHTMLText(h *HTMLText) Visitor
	VisitHTMLInterpolation(h *HTMLInterpolation) Visitor
}
