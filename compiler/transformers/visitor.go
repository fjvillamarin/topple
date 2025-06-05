package transformers

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/resolver"
	"fmt"
)

// TransformerVisitor is a visitor that transforms PSX-specific AST nodes into Python AST nodes
type TransformerVisitor struct {
	// Track transformations
	hasTransformed bool
	errors         []error

	// AST visitor implementation
	ast.Visitor
}

// NewTransformerVisitor creates a new TransformerVisitor
func NewTransformerVisitor() *TransformerVisitor {
	return &TransformerVisitor{
		hasTransformed: false,
		errors:         []error{},
	}
}

// TransformModule transforms a module by replacing ViewStmt nodes with Class nodes
func (mv *TransformerVisitor) TransformModule(module *ast.Module, resolutionTable *resolver.ResolutionTable) (*ast.Module, error) {
	// Create view transformer with resolution table
	viewTransformer := NewViewTransformer(resolutionTable)

	// Transform the module body
	transformedBody, err := mv.transformStatements(module.Body, viewTransformer)
	if err != nil {
		return nil, err
	}

	// Add required imports if any views were transformed
	if mv.hasTransformed {
		imports := viewTransformer.GetRequiredImports()
		// Prepend imports to the module body
		allStmts := make([]ast.Stmt, 0, len(imports)+len(transformedBody))
		for _, imp := range imports {
			allStmts = append(allStmts, imp)
		}
		allStmts = append(allStmts, transformedBody...)
		transformedBody = allStmts
	}

	return &ast.Module{
		Body: transformedBody,
		Span: module.Span,
	}, nil
}

// transformStatements transforms a slice of statements, replacing ViewStmt with Class
func (mv *TransformerVisitor) transformStatements(stmts []ast.Stmt, viewTransformer *ViewTransformer) ([]ast.Stmt, error) {
	var transformed []ast.Stmt

	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ViewStmt:
			// Transform ViewStmt to Class using the configured view transformer
			class, err := viewTransformer.TransformViewToClass(s)
			if err != nil {
				return nil, fmt.Errorf("failed to transform view %s: %w", s.Name.Token.Lexeme, err)
			}
			transformed = append(transformed, class)
			mv.hasTransformed = true

		default:
			// Keep other statements as-is
			transformed = append(transformed, stmt)
		}
	}

	return transformed, nil
}

// Visitor interface methods - most just delegate to default behavior
// We only override the ones we need to transform

func (mv *TransformerVisitor) Visit(node ast.Node) ast.Visitor {
	// Default implementation - just traverse
	return mv
}

func (mv *TransformerVisitor) VisitModule(m *ast.Module) ast.Visitor {
	// We handle modules specially in MutateModule
	return mv
}

func (mv *TransformerVisitor) VisitViewStmt(v *ast.ViewStmt) ast.Visitor {
	// Views are transformed by MutateModule, not during traversal
	return mv
}

// Expression visitors - default implementations
func (mv *TransformerVisitor) VisitName(n *ast.Name) ast.Visitor                   { return mv }
func (mv *TransformerVisitor) VisitLiteral(l *ast.Literal) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitAttribute(a *ast.Attribute) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitCall(c *ast.Call) ast.Visitor                   { return mv }
func (mv *TransformerVisitor) VisitSubscript(s *ast.Subscript) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitBinary(b *ast.Binary) ast.Visitor               { return mv }
func (mv *TransformerVisitor) VisitUnary(u *ast.Unary) ast.Visitor                 { return mv }
func (mv *TransformerVisitor) VisitAssignExpr(a *ast.AssignExpr) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitStarExpr(s *ast.StarExpr) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitTernaryExpr(t *ast.TernaryExpr) ast.Visitor     { return mv }
func (mv *TransformerVisitor) VisitListExpr(l *ast.ListExpr) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitTupleExpr(t *ast.TupleExpr) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitSetExpr(s *ast.SetExpr) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitDictExpr(d *ast.DictExpr) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitListComp(lc *ast.ListComp) ast.Visitor          { return mv }
func (mv *TransformerVisitor) VisitSetComp(sc *ast.SetComp) ast.Visitor            { return mv }
func (mv *TransformerVisitor) VisitDictComp(dc *ast.DictComp) ast.Visitor          { return mv }
func (mv *TransformerVisitor) VisitGenExpr(ge *ast.GenExpr) ast.Visitor            { return mv }
func (mv *TransformerVisitor) VisitYieldExpr(y *ast.YieldExpr) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitGroupExpr(g *ast.GroupExpr) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitParameterList(p *ast.ParameterList) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitParameter(p *ast.Parameter) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitTypeParamExpr(t *ast.TypeParam) ast.Visitor     { return mv }
func (mv *TransformerVisitor) VisitSlice(s *ast.Slice) ast.Visitor                 { return mv }
func (mv *TransformerVisitor) VisitAwaitExpr(a *ast.AwaitExpr) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitArgument(a *ast.Argument) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitLambda(l *ast.Lambda) ast.Visitor               { return mv }
func (mv *TransformerVisitor) VisitFString(f *ast.FString) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitFStringMiddle(f *ast.FStringMiddle) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitFStringReplacementField(f *ast.FStringReplacementField) ast.Visitor {
	return mv
}
func (mv *TransformerVisitor) VisitFStringConversion(f *ast.FStringConversion) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitFStringFormatSpec(f *ast.FStringFormatSpec) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitFStringFormatMiddle(f *ast.FStringFormatMiddle) ast.Visitor {
	return mv
}
func (mv *TransformerVisitor) VisitFStringFormatReplacementField(f *ast.FStringFormatReplacementField) ast.Visitor {
	return mv
}

// Statement visitors - default implementations
func (mv *TransformerVisitor) VisitExprStmt(e *ast.ExprStmt) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitAssignStmt(a *ast.AssignStmt) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitAnnotationStmt(a *ast.AnnotationStmt) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitReturnStmt(r *ast.ReturnStmt) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitRaiseStmt(r *ast.RaiseStmt) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitPassStmt(p *ast.PassStmt) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitYieldStmt(y *ast.YieldStmt) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitAssertStmt(a *ast.AssertStmt) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitBreakStmt(b *ast.BreakStmt) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitContinueStmt(c *ast.ContinueStmt) ast.Visitor     { return mv }
func (mv *TransformerVisitor) VisitGlobalStmt(g *ast.GlobalStmt) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitNonlocalStmt(n *ast.NonlocalStmt) ast.Visitor     { return mv }
func (mv *TransformerVisitor) VisitImportStmt(i *ast.ImportStmt) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitTypeAlias(t *ast.TypeAlias) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitDecorator(d *ast.Decorator) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor           { return mv }

// Compound statement visitors
func (mv *TransformerVisitor) VisitIf(i *ast.If) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitWhile(w *ast.While) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitFor(f *ast.For) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitWith(w *ast.With) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitTry(t *ast.Try) ast.Visitor           { return mv }
func (mv *TransformerVisitor) VisitClass(c *ast.Class) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitFunction(f *ast.Function) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitMatch(m *ast.MatchStmt) ast.Visitor   { return mv }

// Pattern visitors
func (mv *TransformerVisitor) VisitLiteralPattern(lp *ast.LiteralPattern) ast.Visitor   { return mv }
func (mv *TransformerVisitor) VisitCapturePattern(cp *ast.CapturePattern) ast.Visitor   { return mv }
func (mv *TransformerVisitor) VisitWildcardPattern(wp *ast.WildcardPattern) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitValuePattern(vp *ast.ValuePattern) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitGroupPattern(gp *ast.GroupPattern) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitSequencePattern(sp *ast.SequencePattern) ast.Visitor { return mv }
func (mv *TransformerVisitor) VisitStarPattern(sp *ast.StarPattern) ast.Visitor         { return mv }
func (mv *TransformerVisitor) VisitMappingPattern(mp *ast.MappingPattern) ast.Visitor   { return mv }
func (mv *TransformerVisitor) VisitClassPattern(cp *ast.ClassPattern) ast.Visitor       { return mv }
func (mv *TransformerVisitor) VisitAsPattern(ap *ast.AsPattern) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitOrPattern(op *ast.OrPattern) ast.Visitor             { return mv }

// HTML-specific visitors - these remain unchanged for now since ViewMutator handles the transformation
func (mv *TransformerVisitor) VisitHTMLElement(h *ast.HTMLElement) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitHTMLContent(h *ast.HTMLContent) ast.Visitor             { return mv }
func (mv *TransformerVisitor) VisitHTMLText(h *ast.HTMLText) ast.Visitor                   { return mv }
func (mv *TransformerVisitor) VisitHTMLInterpolation(h *ast.HTMLInterpolation) ast.Visitor { return mv }
