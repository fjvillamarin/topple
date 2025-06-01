package codegen

import (
	"biscuit/compiler/ast"
	"fmt"
	"strings"
)

type CodeGenerator struct {
	builder strings.Builder
	indent  int

	// Additional fields for proper code generation
	needsNewline bool
	atLineStart  bool

	ast.Visitor
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		atLineStart: true,
	}
}

// Generate generates Python code from the given AST node
func (cg *CodeGenerator) Generate(node ast.Node) string {
	cg.builder.Reset()
	cg.indent = 0
	cg.needsNewline = false
	cg.atLineStart = true

	node.Accept(cg)
	return cg.builder.String()
}

// Helper methods for formatting
func (cg *CodeGenerator) write(s string) {
	if cg.atLineStart && cg.indent > 0 && s != "\n" {
		cg.builder.WriteString(strings.Repeat("    ", cg.indent))
		cg.atLineStart = false
	}
	cg.builder.WriteString(s)
	if s == "\n" {
		cg.atLineStart = true
	}
}

func (cg *CodeGenerator) writef(format string, args ...interface{}) {
	cg.write(fmt.Sprintf(format, args...))
}

func (cg *CodeGenerator) newline() {
	cg.write("\n")
}

func (cg *CodeGenerator) increaseIndent() {
	cg.indent++
}

func (cg *CodeGenerator) decreaseIndent() {
	cg.indent--
}

func (cg *CodeGenerator) writeStmts(stmts []ast.Stmt) {
	for _, stmt := range stmts {
		stmt.Accept(cg)
	}
}

// Expression visitors
func (cg *CodeGenerator) VisitName(n *ast.Name) ast.Visitor {
	cg.write(n.Token.Lexeme)
	return cg
}

func (cg *CodeGenerator) VisitLiteral(l *ast.Literal) ast.Visitor {
	// Handle the case where literal type might be wrong - check actual value type
	switch v := l.Value.(type) {
	case string:
		if l.Type == ast.LiteralTypeString {
			// Correct string literal
			cg.write(fmt.Sprintf("\"%s\"", v))
		} else {
			// String value but wrong type - treat as string anyway
			cg.write(fmt.Sprintf("\"%s\"", v))
		}
	case int:
		cg.write(fmt.Sprintf("%d", v))
	case int64:
		cg.write(fmt.Sprintf("%d", v))
	case float64:
		cg.write(fmt.Sprintf("%g", v))
	case bool:
		if v {
			cg.write("True")
		} else {
			cg.write("False")
		}
	default:
		// Fallback for other types
		if l.Type == ast.LiteralTypeNone {
			cg.write("None")
		} else {
			cg.write(fmt.Sprintf("%v", v))
		}
	}
	return cg
}

func (cg *CodeGenerator) VisitAttribute(a *ast.Attribute) ast.Visitor {
	a.Object.Accept(cg)
	cg.write(".")
	cg.write(a.Name.Lexeme)
	return cg
}

func (cg *CodeGenerator) VisitCall(c *ast.Call) ast.Visitor {
	c.Callee.Accept(cg)
	cg.write("(")
	for i, arg := range c.Arguments {
		if i > 0 {
			cg.write(", ")
		}
		arg.Accept(cg)
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSubscript(s *ast.Subscript) ast.Visitor {
	s.Object.Accept(cg)
	cg.write("[")
	for i, index := range s.Indices {
		if i > 0 {
			cg.write(", ")
		}
		index.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitBinary(b *ast.Binary) ast.Visitor {
	b.Left.Accept(cg)
	cg.writef(" %s ", b.Operator.Lexeme)
	b.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitUnary(u *ast.Unary) ast.Visitor {
	cg.write(u.Operator.Lexeme)
	if u.Operator.Lexeme == "not" {
		cg.write(" ")
	}
	u.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitAssignExpr(a *ast.AssignExpr) ast.Visitor {
	a.Left.Accept(cg)
	cg.write(" := ")
	a.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitStarExpr(s *ast.StarExpr) ast.Visitor {
	cg.write("*")
	s.Expr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitTernaryExpr(t *ast.TernaryExpr) ast.Visitor {
	t.TrueExpr.Accept(cg)
	cg.write(" if ")
	t.Condition.Accept(cg)
	cg.write(" else ")
	t.FalseExpr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitListExpr(l *ast.ListExpr) ast.Visitor {
	cg.write("[")
	for i, elem := range l.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitTupleExpr(t *ast.TupleExpr) ast.Visitor {
	cg.write("(")
	for i, elem := range t.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	if len(t.Elements) == 1 {
		cg.write(",")
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSetExpr(s *ast.SetExpr) ast.Visitor {
	cg.write("{")
	for i, elem := range s.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitDictExpr(d *ast.DictExpr) ast.Visitor {
	cg.write("{")
	for i, pair := range d.Pairs {
		if i > 0 {
			cg.write(", ")
		}
		switch p := pair.(type) {
		case *ast.KeyValuePair:
			p.Key.Accept(cg)
			cg.write(": ")
			p.Value.Accept(cg)
		case *ast.DoubleStarredPair:
			cg.write("**")
			p.Expr.Accept(cg)
		}
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitListComp(lc *ast.ListComp) ast.Visitor {
	cg.write("[")
	lc.Element.Accept(cg)
	for _, clause := range lc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitSetComp(sc *ast.SetComp) ast.Visitor {
	cg.write("{")
	sc.Element.Accept(cg)
	for _, clause := range sc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitDictComp(dc *ast.DictComp) ast.Visitor {
	cg.write("{")
	dc.Key.Accept(cg)
	cg.write(": ")
	dc.Value.Accept(cg)
	for _, clause := range dc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitGenExpr(ge *ast.GenExpr) ast.Visitor {
	cg.write("(")
	ge.Element.Accept(cg)
	for _, clause := range ge.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write(")")
	return cg
}

// Helper method for writing ForIfClause
func (cg *CodeGenerator) writeForIfClause(clause ast.ForIfClause) {
	if clause.IsAsync {
		cg.write("async ")
	}
	cg.write("for ")
	clause.Target.Accept(cg)
	cg.write(" in ")
	clause.Iter.Accept(cg)
	for _, ifCond := range clause.Ifs {
		cg.write(" if ")
		ifCond.Accept(cg)
	}
}

func (cg *CodeGenerator) VisitYieldExpr(y *ast.YieldExpr) ast.Visitor {
	cg.write("yield")
	if y.Value != nil {
		cg.write(" ")
		y.Value.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitGroupExpr(g *ast.GroupExpr) ast.Visitor {
	cg.write("(")
	g.Expression.Accept(cg)
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitTypeParamExpr(t *ast.TypeParam) ast.Visitor {
	if t.IsStar {
		cg.write("*")
	} else if t.IsDoubleStar {
		cg.write("**")
	}
	cg.write(t.Name.Lexeme)
	if t.Bound != nil {
		cg.write(": ")
		t.Bound.Accept(cg)
	}
	if t.Default != nil {
		cg.write(" = ")
		t.Default.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitSlice(s *ast.Slice) ast.Visitor {
	if s.StartIndex != nil {
		s.StartIndex.Accept(cg)
	}
	cg.write(":")
	if s.EndIndex != nil {
		s.EndIndex.Accept(cg)
	}
	if s.Step != nil {
		cg.write(":")
		s.Step.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitAwaitExpr(a *ast.AwaitExpr) ast.Visitor {
	cg.write("await ")
	a.Expr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitArgument(a *ast.Argument) ast.Visitor {
	if a.Name != nil {
		a.Name.Accept(cg)
		cg.write("=")
	}
	if a.IsStar {
		cg.write("*")
	} else if a.IsDoubleStar {
		cg.write("**")
	}
	a.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitLambda(l *ast.Lambda) ast.Visitor {
	cg.write("lambda")
	if l.Parameters != nil && len(l.Parameters.Parameters) > 0 {
		cg.write(" ")
		l.Parameters.Accept(cg)
	}
	cg.write(": ")
	l.Body.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitFString(f *ast.FString) ast.Visitor {
	cg.write("f\"")
	for _, part := range f.Parts {
		part.Accept(cg)
	}
	cg.write("\"")
	return cg
}

func (cg *CodeGenerator) VisitFStringMiddle(f *ast.FStringMiddle) ast.Visitor {
	cg.write(f.Value)
	return cg
}

func (cg *CodeGenerator) VisitFStringReplacementField(f *ast.FStringReplacementField) ast.Visitor {
	cg.write("{")
	f.Expression.Accept(cg)
	if f.Conversion != nil {
		f.Conversion.Accept(cg)
	}
	if f.FormatSpec != nil {
		f.FormatSpec.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitFStringConversion(f *ast.FStringConversion) ast.Visitor {
	cg.write("!")
	cg.write(f.Type)
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatSpec(f *ast.FStringFormatSpec) ast.Visitor {
	cg.write(":")
	for _, part := range f.Spec {
		part.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatMiddle(f *ast.FStringFormatMiddle) ast.Visitor {
	cg.write(f.Value)
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatReplacementField(f *ast.FStringFormatReplacementField) ast.Visitor {
	cg.write("{")
	f.Expression.Accept(cg)
	cg.write("}")
	return cg
}

// Statement visitors
func (cg *CodeGenerator) Visit(node ast.Node) ast.Visitor {
	// Generic visit method - delegate to specific visitor
	node.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitModule(m *ast.Module) ast.Visitor {
	cg.writeStmts(m.Body)
	return cg
}

func (cg *CodeGenerator) VisitExprStmt(e *ast.ExprStmt) ast.Visitor {
	e.Expr.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitTypeAlias(t *ast.TypeAlias) ast.Visitor {
	cg.write("type ")
	cg.write(t.Name.Lexeme)
	if len(t.Params) > 0 {
		cg.write("[")
		for i, param := range t.Params {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}
	cg.write(" = ")
	t.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitReturnStmt(r *ast.ReturnStmt) ast.Visitor {
	cg.write("return")
	if r.Value != nil {
		cg.write(" ")
		r.Value.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitRaiseStmt(r *ast.RaiseStmt) ast.Visitor {
	cg.write("raise")
	if r.HasException {
		cg.write(" ")
		r.Exception.Accept(cg)
		if r.HasFrom {
			cg.write(" from ")
			r.FromExpr.Accept(cg)
		}
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitPassStmt(p *ast.PassStmt) ast.Visitor {
	cg.write("pass")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitBreakStmt(b *ast.BreakStmt) ast.Visitor {
	cg.write("break")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitContinueStmt(c *ast.ContinueStmt) ast.Visitor {
	cg.write("continue")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitYieldStmt(y *ast.YieldStmt) ast.Visitor {
	y.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAssertStmt(a *ast.AssertStmt) ast.Visitor {
	cg.write("assert ")
	a.Test.Accept(cg)
	if a.Message != nil {
		cg.write(", ")
		a.Message.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitGlobalStmt(g *ast.GlobalStmt) ast.Visitor {
	cg.write("global ")
	for i, name := range g.Names {
		if i > 0 {
			cg.write(", ")
		}
		name.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitNonlocalStmt(n *ast.NonlocalStmt) ast.Visitor {
	cg.write("nonlocal ")
	for i, name := range n.Names {
		if i > 0 {
			cg.write(", ")
		}
		name.Accept(cg)
	}
	cg.newline()
	return cg
}

// Helper method for writing import names
func (cg *CodeGenerator) writeImportName(name *ast.ImportName) {
	cg.writeDottedName(name.DottedName)
	if name.AsName != nil {
		cg.write(" as ")
		name.AsName.Accept(cg)
	}
}

// Helper method for writing dotted names
func (cg *CodeGenerator) writeDottedName(name *ast.DottedName) {
	for i, part := range name.Names {
		if i > 0 {
			cg.write(".")
		}
		part.Accept(cg)
	}
}

func (cg *CodeGenerator) VisitImportStmt(i *ast.ImportStmt) ast.Visitor {
	cg.write("import ")
	for idx, alias := range i.Names {
		if idx > 0 {
			cg.write(", ")
		}
		cg.writeImportName(alias)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor {
	cg.write("from ")
	if i.DotCount > 0 {
		cg.write(strings.Repeat(".", i.DotCount))
	}
	if i.DottedName != nil {
		cg.writeDottedName(i.DottedName)
	}
	cg.write(" import ")
	if i.IsWildcard {
		cg.write("*")
	} else {
		for idx, alias := range i.Names {
			if idx > 0 {
				cg.write(", ")
			}
			cg.writeImportName(alias)
		}
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAssignStmt(a *ast.AssignStmt) ast.Visitor {
	for i, target := range a.Targets {
		if i > 0 {
			cg.write(" = ")
		}
		target.Accept(cg)
	}
	cg.write(" = ")
	a.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAnnotationStmt(a *ast.AnnotationStmt) ast.Visitor {
	a.Target.Accept(cg)
	cg.write(": ")
	a.Type.Accept(cg)
	if a.HasValue {
		cg.write(" = ")
		a.Value.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
	for i, stmt := range m.Stmts {
		if i > 0 {
			cg.write("; ")
		}
		stmt.Accept(cg)
	}
	return cg
}

// Compound statement visitors
func (cg *CodeGenerator) VisitIf(i *ast.If) ast.Visitor {
	cg.write("if ")
	i.Condition.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(i.Body)
	cg.decreaseIndent()

	if len(i.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(i.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitWhile(w *ast.While) ast.Visitor {
	cg.write("while ")
	w.Test.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(w.Body)
	cg.decreaseIndent()

	if len(w.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(w.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitFor(f *ast.For) ast.Visitor {
	if f.IsAsync {
		cg.write("async ")
	}
	cg.write("for ")
	f.Target.Accept(cg)
	cg.write(" in ")
	f.Iterable.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(f.Body)
	cg.decreaseIndent()

	if len(f.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(f.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitWith(w *ast.With) ast.Visitor {
	if w.IsAsync {
		cg.write("async ")
	}
	cg.write("with ")
	for i, item := range w.Items {
		if i > 0 {
			cg.write(", ")
		}
		item.Expr.Accept(cg)
		if item.As != nil {
			cg.write(" as ")
			item.As.Accept(cg)
		}
	}
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(w.Body)
	cg.decreaseIndent()
	return cg
}

func (cg *CodeGenerator) VisitTry(t *ast.Try) ast.Visitor {
	cg.write("try:")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(t.Body)
	cg.decreaseIndent()

	for _, handler := range t.Excepts {
		cg.write("except")
		if handler.Type != nil {
			cg.write(" ")
			handler.Type.Accept(cg)
			if handler.Name != nil {
				cg.write(" as ")
				handler.Name.Accept(cg)
			}
		}
		cg.write(":")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(handler.Body)
		cg.decreaseIndent()
	}

	if len(t.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(t.Else)
		cg.decreaseIndent()
	}

	if len(t.Finally) > 0 {
		cg.write("finally:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(t.Finally)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitDecorator(d *ast.Decorator) ast.Visitor {
	cg.write("@")
	d.Expr.Accept(cg)
	cg.newline()
	d.Stmt.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitClass(c *ast.Class) ast.Visitor {
	cg.write("class ")
	c.Name.Accept(cg)

	if len(c.TypeParams) > 0 {
		cg.write("[")
		for i, param := range c.TypeParams {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}

	if len(c.Args) > 0 {
		cg.write("(")
		for i, base := range c.Args {
			if i > 0 {
				cg.write(", ")
			}
			base.Accept(cg)
		}
		cg.write(")")
	}

	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(c.Body)
	cg.decreaseIndent()
	return cg
}

func (cg *CodeGenerator) VisitFunction(f *ast.Function) ast.Visitor {
	if f.IsAsync {
		cg.write("async ")
	}
	cg.write("def ")
	f.Name.Accept(cg)

	if len(f.TypeParameters) > 0 {
		cg.write("[")
		for i, param := range f.TypeParameters {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}

	cg.write("(")
	if f.Parameters != nil {
		f.Parameters.Accept(cg)
	}
	cg.write(")")

	if f.ReturnType != nil {
		cg.write(" -> ")
		f.ReturnType.Accept(cg)
	}

	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(f.Body)
	cg.decreaseIndent()
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitParameterList(p *ast.ParameterList) ast.Visitor {
	for i, param := range p.Parameters {
		if i > 0 {
			cg.write(", ")
		}
		param.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitParameter(p *ast.Parameter) ast.Visitor {
	if p.IsStar {
		cg.write("*")
	} else if p.IsDoubleStar {
		cg.write("**")
	}

	if p.Name != nil {
		p.Name.Accept(cg)
	}

	if p.Annotation != nil {
		cg.write(": ")
		p.Annotation.Accept(cg)
	}

	if p.Default != nil {
		cg.write(" = ")
		p.Default.Accept(cg)
	}

	return cg
}

func (cg *CodeGenerator) VisitMatch(m *ast.MatchStmt) ast.Visitor {
	cg.write("match ")
	m.Subject.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()

	for _, caseStmt := range m.Cases {
		for _, pattern := range caseStmt.Patterns {
			cg.write("case ")
			pattern.Accept(cg)
			if caseStmt.Guard != nil {
				cg.write(" if ")
				caseStmt.Guard.Accept(cg)
			}
			cg.write(":")
			cg.newline()
			cg.increaseIndent()
			cg.writeStmts(caseStmt.Body)
			cg.decreaseIndent()
		}
	}

	cg.decreaseIndent()
	return cg
}

// Pattern visitors (simplified implementations)
func (cg *CodeGenerator) VisitLiteralPattern(lp *ast.LiteralPattern) ast.Visitor {
	lp.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitCapturePattern(cp *ast.CapturePattern) ast.Visitor {
	cp.Name.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitWildcardPattern(wp *ast.WildcardPattern) ast.Visitor {
	cg.write("_")
	return cg
}

func (cg *CodeGenerator) VisitValuePattern(vp *ast.ValuePattern) ast.Visitor {
	vp.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitGroupPattern(gp *ast.GroupPattern) ast.Visitor {
	cg.write("(")
	gp.Pattern.Accept(cg)
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSequencePattern(sp *ast.SequencePattern) ast.Visitor {
	cg.write("[")
	for i, pattern := range sp.Patterns {
		if i > 0 {
			cg.write(", ")
		}
		pattern.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitStarPattern(sp *ast.StarPattern) ast.Visitor {
	cg.write("*")
	if sp.Pattern != nil {
		sp.Pattern.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitMappingPattern(mp *ast.MappingPattern) ast.Visitor {
	cg.write("{")
	for i, pair := range mp.Pairs {
		if i > 0 {
			cg.write(", ")
		}
		pair.Key.Accept(cg)
		cg.write(": ")
		pair.Pattern.Accept(cg)
	}
	if mp.DoubleStar != nil {
		if len(mp.Pairs) > 0 {
			cg.write(", ")
		}
		cg.write("**")
		mp.DoubleStar.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitClassPattern(cp *ast.ClassPattern) ast.Visitor {
	cp.Class.Accept(cg)
	cg.write("(")
	for i, pattern := range cp.Patterns {
		if i > 0 {
			cg.write(", ")
		}
		pattern.Accept(cg)
	}
	for i, kwPattern := range cp.KwdPatterns {
		if i > 0 || len(cp.Patterns) > 0 {
			cg.write(", ")
		}
		// cg.write(kwPattern.Arg)
		kwPattern.Name.Accept(cg)
		cg.write("=")
		kwPattern.Pattern.Accept(cg)
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitAsPattern(ap *ast.AsPattern) ast.Visitor {
	if ap.Pattern != nil {
		ap.Pattern.Accept(cg)
		cg.write(" as ")
	}
	ap.Target.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitOrPattern(op *ast.OrPattern) ast.Visitor {
	for i, pattern := range op.Patterns {
		if i > 0 {
			cg.write(" | ")
		}
		pattern.Accept(cg)
	}
	return cg
}

// PSX-specific visitors that should not be called after mutation
func (cg *CodeGenerator) VisitViewStmt(v *ast.ViewStmt) ast.Visitor {
	// This should never be called since ViewStmt nodes are transformed to Class nodes by the mutator
	panic("ViewStmt nodes should be transformed to Class nodes before code generation")
}

func (cg *CodeGenerator) VisitHTMLElement(h *ast.HTMLElement) ast.Visitor {
	// This should never be called since HTML elements are transformed by the mutator
	panic("HTMLElement nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLContent(h *ast.HTMLContent) ast.Visitor {
	// This should never be called since HTML content is transformed by the mutator
	panic("HTMLContent nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLText(h *ast.HTMLText) ast.Visitor {
	// This should never be called since HTML text is transformed by the mutator
	panic("HTMLText nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLInterpolation(h *ast.HTMLInterpolation) ast.Visitor {
	// This should never be called since HTML interpolation is transformed by the mutator
	panic("HTMLInterpolation nodes should be transformed before code generation")
}
