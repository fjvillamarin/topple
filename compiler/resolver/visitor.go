package resolver

import (
	"biscuit/compiler/ast"
	"fmt"
)

// Visitor implementation - this is where the main resolution logic goes

// Visit is the fallback visitor method
func (r *Resolver) Visit(node ast.Node) ast.Visitor {
	// Default implementation - just traverse
	return r
}

// ===== Module and Top-level =====

func (r *Resolver) VisitModule(m *ast.Module) ast.Visitor {
	// Visit all statements in the module
	for _, stmt := range m.Body {
		if stmt != nil {
			stmt.Accept(r)
		}
	}
	return r
}

// ===== Expression Visitors =====

func (r *Resolver) VisitName(name *ast.Name) ast.Visitor {
	if err := r.ResolveName(name); err != nil {
		r.ReportError(err)
	}
	return r
}

func (r *Resolver) VisitLiteral(l *ast.Literal) ast.Visitor {
	// Literals don't introduce or reference variables
	return r
}

func (r *Resolver) VisitAttribute(a *ast.Attribute) ast.Visitor {
	// Only resolve the object part (x in x.attr)
	if a.Object != nil {
		a.Object.Accept(r)
	}
	return r
}

func (r *Resolver) VisitCall(c *ast.Call) ast.Visitor {
	// Resolve callee and arguments
	if c.Callee != nil {
		c.Callee.Accept(r)
	}
	for _, arg := range c.Arguments {
		if arg != nil && arg.Value != nil {
			arg.Value.Accept(r)
		}
	}
	return r
}

func (r *Resolver) VisitSubscript(s *ast.Subscript) ast.Visitor {
	// Resolve both object and indices
	if s.Object != nil {
		s.Object.Accept(r)
	}
	for _, index := range s.Indices {
		if index != nil {
			index.Accept(r)
		}
	}
	return r
}

func (r *Resolver) VisitBinary(b *ast.Binary) ast.Visitor {
	// Resolve left and right operands
	if b.Left != nil {
		b.Left.Accept(r)
	}
	if b.Right != nil {
		b.Right.Accept(r)
	}
	return r
}

func (r *Resolver) VisitUnary(u *ast.Unary) ast.Visitor {
	// Resolve operand
	if u.Right != nil {
		u.Right.Accept(r)
	}
	return r
}

// ===== Statement Visitors =====

func (r *Resolver) VisitExprStmt(e *ast.ExprStmt) ast.Visitor {
	// Resolve the expression
	if e.Expr != nil {
		e.Expr.Accept(r)
	}
	return r
}

func (r *Resolver) VisitAssignStmt(a *ast.AssignStmt) ast.Visitor {
	// Visit the value first (right-hand side)
	if a.Value != nil {
		a.Value.Accept(r)
	}

	// Then handle targets (which create/update bindings)
	for _, target := range a.Targets {
		r.AnalyzeAssignmentTarget(target)
	}

	return r
}

func (r *Resolver) VisitGlobalStmt(g *ast.GlobalStmt) ast.Visitor {
	if r.Current.ScopeType == ModuleScopeType {
		r.ReportError(fmt.Errorf("'global' declaration at module level"))
		return r
	}

	for _, name := range g.Names {
		varName := name.Token.Lexeme

		// Create or reference module-level variable
		if variable, exists := r.ModuleGlobals[varName]; exists {
			// Mark existing variable as global for summary counting
			variable.IsGlobal = true
			r.Current.Globals[varName] = variable
		} else {
			variable := &Variable{
				Name:            varName,
				IsGlobal:        true,
				DefinitionDepth: 0,
				FirstDefSpan:    name.Span,
			}
			r.ModuleGlobals[varName] = variable
			r.Current.Globals[varName] = variable
		}
	}
	return r
}

func (r *Resolver) VisitNonlocalStmt(n *ast.NonlocalStmt) ast.Visitor {
	if r.Current == nil {
		r.ReportError(fmt.Errorf("'nonlocal' declaration outside any scope"))
		return r
	}

	if r.Current.ScopeType != FunctionScopeType && r.Current.ScopeType != ViewScopeType {
		r.ReportError(fmt.Errorf("'nonlocal' declaration not in function scope"))
		return r
	}

	for _, name := range n.Names {
		varName := name.Token.Lexeme

		// Find in enclosing scopes (not global, not current)
		found := false
		for i := len(r.Scopes) - 2; i >= 1; i-- {
			scope := r.Scopes[i]
			if variable, exists := scope.Values[varName]; exists {
				r.Current.Nonlocals[varName] = variable
				// Mark the variable as captured since it's being accessed from nested scope
				variable.IsCaptured = true
				// Mark the variable as nonlocal for proper summary counting
				variable.IsNonlocal = true
				found = true
				break
			}
		}

		if !found {
			r.ReportError(fmt.Errorf("no binding for nonlocal '%s' found", varName))
		}
	}
	return r
}

// ===== Function and View Visitors =====

func (r *Resolver) VisitFunction(f *ast.Function) ast.Visitor {
	// Function name is bound in enclosing scope
	if f.Name != nil {
		variable := r.DefineVariable(f.Name.Token.Lexeme, f.Name.Span)
		variable.State = VariableDefined // Mark function as defined, not just declared
		r.Variables[f.Name] = variable
		r.ScopeDepths[f.Name] = len(r.Scopes) - 1 // Current scope depth
	}

	// Function body has its own scope
	r.BeginScope(FunctionScopeType)
	oldFunction := r.CurrentFunction
	r.CurrentFunction = f

	defer func() {
		r.EndScope()
		r.CurrentFunction = oldFunction
	}()

	// Parameters create local bindings
	if f.Parameters != nil {
		for _, param := range f.Parameters.Parameters {
			if param.Name != nil {
				variable := r.DefineVariable(param.Name.Token.Lexeme, param.Name.Span)
				variable.IsParameter = true
				variable.State = VariableDefined
				// IMPORTANT: Add parameter name to Variables map
				r.Variables[param.Name] = variable
				r.ScopeDepths[param.Name] = len(r.Scopes) - 1 // Current scope depth
			}

			// Visit default values in enclosing scope
			if param.Default != nil {
				param.Default.Accept(r)
			}
		}
	}

	// Visit function body
	for _, stmt := range f.Body {
		stmt.Accept(r)
	}

	return r
}

func (r *Resolver) VisitViewStmt(v *ast.ViewStmt) ast.Visitor {
	// View name is bound in enclosing scope
	if v.Name != nil {
		variable := r.DefineVariable(v.Name.Token.Lexeme, v.Name.Span)
		variable.State = VariableDefined // Mark view as defined, not just declared
		r.Variables[v.Name] = variable
		r.ScopeDepths[v.Name] = len(r.Scopes) - 1 // Current scope depth
	}

	// View body has its own scope
	r.BeginScope(ViewScopeType)
	oldView := r.CurrentView
	r.CurrentView = v

	defer func() {
		r.EndScope()
		r.CurrentView = oldView
	}()

	// View parameters create local bindings
	if v.Params != nil {
		for _, param := range v.Params.Parameters {
			if param.Name != nil {
				variable := r.DefineVariable(param.Name.Token.Lexeme, param.Name.Span)
				variable.IsParameter = true
				variable.IsViewParameter = true
				variable.State = VariableDefined
				// IMPORTANT: Add parameter name to Variables map
				r.Variables[param.Name] = variable
				r.ScopeDepths[param.Name] = len(r.Scopes) - 1 // Current scope depth
			}
		}
	}

	// Visit view body
	for _, stmt := range v.Body {
		stmt.Accept(r)
	}

	return r
}

func (r *Resolver) VisitClass(c *ast.Class) ast.Visitor {
	// Class name is bound in enclosing scope
	if c.Name != nil {
		r.DefineVariable(c.Name.Token.Lexeme, c.Name.Span)
	}

	// Class body has its own scope
	r.BeginScope(ClassScopeType)
	defer r.EndScope()

	// Visit class body
	for _, stmt := range c.Body {
		stmt.Accept(r)
	}

	return r
}

// ===== Assignment Target Analysis =====

func (r *Resolver) AnalyzeAssignmentTarget(target ast.Expr) {
	switch t := target.(type) {
	case *ast.Name:
		varName := t.Token.Lexeme

		// Check if this variable has a global declaration in current scope
		if r.Current != nil && r.Current.Globals[varName] != nil {
			// This is a global variable - assign to the global
			globalVar := r.Current.Globals[varName]
			globalVar.State = VariableDefined
			r.Variables[t] = globalVar
			r.ScopeDepths[t] = 0 // Global scope
			return
		}

		// Check if this variable has a nonlocal declaration in current scope
		if r.Current != nil && r.Current.Nonlocals[varName] != nil {
			// This is a nonlocal variable - assign to the nonlocal
			nonlocalVar := r.Current.Nonlocals[varName]
			nonlocalVar.State = VariableDefined
			r.Variables[t] = nonlocalVar
			// Find the scope depth where this nonlocal variable was originally defined
			for i := len(r.Scopes) - 1; i >= 0; i-- {
				if r.Scopes[i].Values[varName] == nonlocalVar {
					r.ScopeDepths[t] = i
					break
				}
			}
			return
		}

		// Simple assignment - define or update variable in current scope
		if variable, exists := r.Current.Values[varName]; exists {
			// Variable already exists in current scope, update its state
			variable.State = VariableDefined
			r.Variables[t] = variable
			r.ScopeDepths[t] = len(r.Scopes) - 1 // Current scope depth
		} else {
			// Variable doesn't exist in current scope, create new one
			variable := r.DefineVariable(varName, t.Span)
			variable.State = VariableDefined
			r.Variables[t] = variable
			r.ScopeDepths[t] = len(r.Scopes) - 1 // Current scope depth
		}

	case *ast.TupleExpr:
		// Tuple unpacking
		for _, elem := range t.Elements {
			r.AnalyzeAssignmentTarget(elem)
		}

	case *ast.ListExpr:
		// List unpacking
		for _, elem := range t.Elements {
			r.AnalyzeAssignmentTarget(elem)
		}

	case *ast.StarExpr:
		// Starred expression in unpacking
		r.AnalyzeAssignmentTarget(t.Expr)

	case *ast.Attribute:
		// Attribute assignment - only resolve the object
		t.Object.Accept(r)

	case *ast.Subscript:
		// Subscript assignment - resolve both parts
		t.Object.Accept(r)
		for _, index := range t.Indices {
			if index != nil {
				index.Accept(r)
			}
		}

	default:
		r.ReportError(fmt.Errorf("invalid assignment target"))
	}
}

// ===== Placeholder visitors for other nodes =====
// TODO: Implement these as needed

func (r *Resolver) VisitAssignExpr(a *ast.AssignExpr) ast.Visitor { return r }
func (r *Resolver) VisitStarExpr(s *ast.StarExpr) ast.Visitor     { return r }
func (r *Resolver) VisitTernaryExpr(t *ast.TernaryExpr) ast.Visitor {
	// Visit condition, true expression, and false expression
	if t.Condition != nil {
		t.Condition.Accept(r)
	}
	if t.TrueExpr != nil {
		t.TrueExpr.Accept(r)
	}
	if t.FalseExpr != nil {
		t.FalseExpr.Accept(r)
	}
	return r
}
func (r *Resolver) VisitListExpr(l *ast.ListExpr) ast.Visitor {
	// Visit all elements in the list
	for _, element := range l.Elements {
		if element != nil {
			element.Accept(r)
		}
	}
	return r
}
func (r *Resolver) VisitTupleExpr(t *ast.TupleExpr) ast.Visitor {
	// Visit all elements in the tuple
	for _, element := range t.Elements {
		if element != nil {
			element.Accept(r)
		}
	}
	return r
}
func (r *Resolver) VisitSetExpr(s *ast.SetExpr) ast.Visitor { return r }
func (r *Resolver) VisitDictExpr(d *ast.DictExpr) ast.Visitor {
	// Visit all key-value pairs in the dictionary
	for _, pair := range d.Pairs {
		if pair != nil {
			switch p := pair.(type) {
			case *ast.KeyValuePair:
				if p.Key != nil {
					p.Key.Accept(r)
				}
				if p.Value != nil {
					p.Value.Accept(r)
				}
			case *ast.DoubleStarredPair:
				if p.Expr != nil {
					p.Expr.Accept(r)
				}
			}
		}
	}
	return r
}
func (r *Resolver) VisitListComp(lc *ast.ListComp) ast.Visitor  { return r }
func (r *Resolver) VisitSetComp(sc *ast.SetComp) ast.Visitor    { return r }
func (r *Resolver) VisitDictComp(dc *ast.DictComp) ast.Visitor  { return r }
func (r *Resolver) VisitGenExpr(ge *ast.GenExpr) ast.Visitor    { return r }
func (r *Resolver) VisitYieldExpr(y *ast.YieldExpr) ast.Visitor { return r }
func (r *Resolver) VisitGroupExpr(g *ast.GroupExpr) ast.Visitor {
	// Visit the grouped expression
	if g.Expression != nil {
		g.Expression.Accept(r)
	}
	return r
}
func (r *Resolver) VisitParameterList(p *ast.ParameterList) ast.Visitor { return r }
func (r *Resolver) VisitParameter(p *ast.Parameter) ast.Visitor         { return r }
func (r *Resolver) VisitTypeParamExpr(t *ast.TypeParam) ast.Visitor     { return r }
func (r *Resolver) VisitSlice(s *ast.Slice) ast.Visitor                 { return r }
func (r *Resolver) VisitAwaitExpr(a *ast.AwaitExpr) ast.Visitor         { return r }
func (r *Resolver) VisitArgument(a *ast.Argument) ast.Visitor           { return r }
func (r *Resolver) VisitLambda(l *ast.Lambda) ast.Visitor               { return r }
func (r *Resolver) VisitFString(f *ast.FString) ast.Visitor {
	// Visit all parts of the f-string
	for _, part := range f.Parts {
		if part != nil {
			part.Accept(r)
		}
	}
	return r
}
func (r *Resolver) VisitFStringMiddle(f *ast.FStringMiddle) ast.Visitor { return r }
func (r *Resolver) VisitFStringReplacementField(f *ast.FStringReplacementField) ast.Visitor {
	// Visit the expression inside the replacement field
	if f.Expression != nil {
		f.Expression.Accept(r)
	}
	return r
}
func (r *Resolver) VisitFStringConversion(f *ast.FStringConversion) ast.Visitor { return r }
func (r *Resolver) VisitFStringFormatSpec(f *ast.FStringFormatSpec) ast.Visitor { return r }
func (r *Resolver) VisitFStringFormatMiddle(f *ast.FStringFormatMiddle) ast.Visitor {
	return r
}
func (r *Resolver) VisitFStringFormatReplacementField(f *ast.FStringFormatReplacementField) ast.Visitor {
	return r
}

func (r *Resolver) VisitAnnotationStmt(a *ast.AnnotationStmt) ast.Visitor { return r }
func (r *Resolver) VisitReturnStmt(ret *ast.ReturnStmt) ast.Visitor {
	if ret.Value != nil {
		ret.Value.Accept(r)
	}
	return r
}
func (r *Resolver) VisitRaiseStmt(rs *ast.RaiseStmt) ast.Visitor          { return r }
func (r *Resolver) VisitPassStmt(p *ast.PassStmt) ast.Visitor             { return r }
func (r *Resolver) VisitYieldStmt(y *ast.YieldStmt) ast.Visitor           { return r }
func (r *Resolver) VisitAssertStmt(a *ast.AssertStmt) ast.Visitor         { return r }
func (r *Resolver) VisitBreakStmt(b *ast.BreakStmt) ast.Visitor           { return r }
func (r *Resolver) VisitContinueStmt(c *ast.ContinueStmt) ast.Visitor     { return r }
func (r *Resolver) VisitImportStmt(i *ast.ImportStmt) ast.Visitor         { return r }
func (r *Resolver) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor { return r }
func (r *Resolver) VisitTypeAlias(t *ast.TypeAlias) ast.Visitor           { return r }
func (r *Resolver) VisitDecorator(d *ast.Decorator) ast.Visitor           { return r }
func (r *Resolver) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
	// Visit all sub-statements
	for _, stmt := range m.Stmts {
		stmt.Accept(r)
	}

	return r
}

func (r *Resolver) VisitIf(i *ast.If) ast.Visitor {
	// Visit the condition
	if i.Condition != nil {
		i.Condition.Accept(r)
	}

	// Visit the body
	for _, stmt := range i.Body {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	// Visit the else clause if present
	for _, stmt := range i.Else {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	return r
}

func (r *Resolver) VisitWhile(w *ast.While) ast.Visitor {
	// Visit the test condition
	if w.Test != nil {
		w.Test.Accept(r)
	}

	// Visit the body
	for _, stmt := range w.Body {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	// Visit the else clause if present
	for _, stmt := range w.Else {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	return r
}

func (r *Resolver) VisitFor(f *ast.For) ast.Visitor {
	// Visit the iterable first
	if f.Iterable != nil {
		f.Iterable.Accept(r)
	}

	// Visit the target (loop variable)
	if f.Target != nil {
		f.Target.Accept(r)
	}

	// Visit the body
	for _, stmt := range f.Body {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	// Visit the else clause if present
	for _, stmt := range f.Else {
		if stmt != nil {
			stmt.Accept(r)
		}
	}

	return r
}

func (r *Resolver) VisitWith(w *ast.With) ast.Visitor       { return r }
func (r *Resolver) VisitTry(t *ast.Try) ast.Visitor         { return r }
func (r *Resolver) VisitMatch(m *ast.MatchStmt) ast.Visitor { return r }

func (r *Resolver) VisitLiteralPattern(lp *ast.LiteralPattern) ast.Visitor   { return r }
func (r *Resolver) VisitCapturePattern(cp *ast.CapturePattern) ast.Visitor   { return r }
func (r *Resolver) VisitWildcardPattern(wp *ast.WildcardPattern) ast.Visitor { return r }
func (r *Resolver) VisitValuePattern(vp *ast.ValuePattern) ast.Visitor       { return r }
func (r *Resolver) VisitGroupPattern(gp *ast.GroupPattern) ast.Visitor       { return r }
func (r *Resolver) VisitSequencePattern(sp *ast.SequencePattern) ast.Visitor { return r }
func (r *Resolver) VisitStarPattern(sp *ast.StarPattern) ast.Visitor         { return r }
func (r *Resolver) VisitMappingPattern(mp *ast.MappingPattern) ast.Visitor   { return r }
func (r *Resolver) VisitClassPattern(cp *ast.ClassPattern) ast.Visitor       { return r }
func (r *Resolver) VisitAsPattern(ap *ast.AsPattern) ast.Visitor             { return r }
func (r *Resolver) VisitOrPattern(op *ast.OrPattern) ast.Visitor             { return r }

func (r *Resolver) VisitHTMLElement(h *ast.HTMLElement) ast.Visitor {
	// Visit all attributes first - they contain expressions that need resolution
	for _, attr := range h.Attributes {
		if attr.Value != nil {
			attr.Value.Accept(r)
		}
	}

	// Visit all content inside the HTML element
	for _, content := range h.Content {
		if content != nil {
			content.Accept(r)
		}
	}
	return r
}

func (r *Resolver) VisitHTMLContent(h *ast.HTMLContent) ast.Visitor {
	// Visit all parts of the HTML content
	for _, part := range h.Parts {
		if part != nil {
			part.Accept(r)
		}
	}
	return r
}

func (r *Resolver) VisitHTMLText(h *ast.HTMLText) ast.Visitor {
	// HTML text doesn't contain variables to resolve
	return r
}

func (r *Resolver) VisitHTMLInterpolation(h *ast.HTMLInterpolation) ast.Visitor {
	// Visit the expression inside the interpolation
	if h.Expression != nil {
		h.Expression.Accept(r)
	}
	return r
}
