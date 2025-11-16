package resolver

import (
	"context"
	"fmt"
	"strings"
	"topple/compiler/ast"
	"topple/compiler/symbol"
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
	if r.ScopeChain.ScopeType == ModuleScopeType {
		r.ReportError(fmt.Errorf("'global' declaration at module level"))
		return r
	}

	for _, name := range g.Names {
		varName := name.Token.Lexeme

		// Create or reference module-level variable
		if variable, exists := r.ModuleGlobals[varName]; exists {
			// Find or create binding in module scope
			moduleScope := r.ScopeChain
			for moduleScope.Parent != nil {
				moduleScope = moduleScope.Parent
			}
			binding := &Binding{
				Name:     varName,
				Variable: variable,
				Scope:    moduleScope,
			}

			// Mark existing variable as global for summary counting
			variable.IsGlobal = true
			r.ScopeChain.Globals[varName] = binding
		} else {
			// Find module scope
			moduleScope := r.ScopeChain
			for moduleScope.Parent != nil {
				moduleScope = moduleScope.Parent
			}

			variable := &Variable{
				Name:            varName,
				IsGlobal:        true,
				DefinitionDepth: 0,
				FirstDefSpan:    name.Span,
			}
			binding := &Binding{
				Name:     varName,
				Variable: variable,
				Scope:    moduleScope,
			}
			r.ModuleGlobals[varName] = variable
			r.ScopeChain.Globals[varName] = binding
		}
	}
	return r
}

func (r *Resolver) VisitNonlocalStmt(n *ast.NonlocalStmt) ast.Visitor {
	if r.ScopeChain == nil {
		r.ReportError(fmt.Errorf("'nonlocal' declaration outside any scope"))
		return r
	}

	if r.ScopeChain.ScopeType != FunctionScopeType && r.ScopeChain.ScopeType != ViewScopeType {
		r.ReportError(fmt.Errorf("'nonlocal' declaration not in function scope"))
		return r
	}

	for _, name := range n.Names {
		varName := name.Token.Lexeme

		// Find in enclosing scopes (not global, not current)
		found := false
		for scope := r.ScopeChain.Parent; scope != nil && scope.ScopeType != ModuleScopeType; scope = scope.Parent {
			if binding, exists := scope.Bindings[varName]; exists {
				r.ScopeChain.Nonlocals[varName] = binding
				// Mark the variable as captured since it's being accessed from nested scope
				binding.Variable.IsCaptured = true
				// Mark the variable as nonlocal for proper summary counting
				binding.Variable.IsNonlocal = true
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

		// Track binding for the function name
		if binding, exists := r.ScopeChain.Bindings[f.Name.Token.Lexeme]; exists {
			r.NameToBinding[f.Name] = binding
			r.NodeScopes[f.Name] = r.ScopeChain
		}

		// Calculate scope depth
		depth := 0
		for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
			depth++
		}
		r.ScopeDepths[f.Name] = depth - 1
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

				// Track binding for the parameter
				if binding, exists := r.ScopeChain.Bindings[param.Name.Token.Lexeme]; exists {
					r.NameToBinding[param.Name] = binding
					r.NodeScopes[param.Name] = r.ScopeChain
				}

				// Calculate scope depth
				depth := 0
				for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
					depth++
				}
				r.ScopeDepths[param.Name] = depth - 1
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

		// Track binding for the view name
		if binding, exists := r.ScopeChain.Bindings[v.Name.Token.Lexeme]; exists {
			r.NameToBinding[v.Name] = binding
			r.NodeScopes[v.Name] = r.ScopeChain
		}

		// Calculate scope depth
		depth := 0
		for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
			depth++
		}
		r.ScopeDepths[v.Name] = depth - 1

		// Track view definition for composition (only at module level)
		if r.ScopeChain != nil && r.ScopeChain.ScopeType == ModuleScopeType {
			r.Views[v.Name.Token.Lexeme] = v
		}
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
				variable.IsViewParameter = true // CRITICAL: Mark as view parameter
				variable.State = VariableDefined
				// IMPORTANT: Add parameter name to Variables map
				r.Variables[param.Name] = variable

				// Track binding for the view parameter - THIS IS THE KEY FIX
				if binding, exists := r.ScopeChain.Bindings[param.Name.Token.Lexeme]; exists {
					r.NameToBinding[param.Name] = binding
					r.NodeScopes[param.Name] = r.ScopeChain
				}

				// Calculate scope depth
				depth := 0
				for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
					depth++
				}
				r.ScopeDepths[param.Name] = depth - 1
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
		if r.ScopeChain != nil && r.ScopeChain.Globals[varName] != nil {
			// This is a global variable - assign to the global
			globalBinding := r.ScopeChain.Globals[varName]
			globalBinding.Variable.State = VariableDefined
			r.Variables[t] = globalBinding.Variable
			r.NameToBinding[t] = globalBinding
			r.NodeScopes[t] = globalBinding.Scope
			r.ScopeDepths[t] = 0 // Global scope
			return
		}

		// Check if this variable has a nonlocal declaration in current scope
		if r.ScopeChain != nil && r.ScopeChain.Nonlocals[varName] != nil {
			// This is a nonlocal variable - assign to the nonlocal
			nonlocalBinding := r.ScopeChain.Nonlocals[varName]
			nonlocalBinding.Variable.State = VariableDefined
			r.Variables[t] = nonlocalBinding.Variable
			r.NameToBinding[t] = nonlocalBinding
			r.NodeScopes[t] = nonlocalBinding.Scope
			// Calculate scope depth
			depth := 0
			for scope := r.ScopeChain; scope != nil && scope != nonlocalBinding.Scope; scope = scope.Parent {
				depth++
			}
			r.ScopeDepths[t] = depth
			return
		}

		// Simple assignment - define or update variable in current scope
		if binding, exists := r.ScopeChain.Bindings[varName]; exists {
			// Variable already exists in current scope, update its state
			binding.Variable.State = VariableDefined
			r.Variables[t] = binding.Variable
			r.NameToBinding[t] = binding
			r.NodeScopes[t] = r.ScopeChain
			// Calculate scope depth
			depth := 0
			for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
				depth++
			}
			r.ScopeDepths[t] = depth - 1
		} else {
			// Variable doesn't exist in current scope, create new one
			variable := r.DefineVariable(varName, t.Span)
			variable.State = VariableDefined
			r.Variables[t] = variable
			// Track the new binding
			if binding, exists := r.ScopeChain.Bindings[varName]; exists {
				r.NameToBinding[t] = binding
				r.NodeScopes[t] = r.ScopeChain
			}
			// Calculate scope depth
			depth := 0
			for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
				depth++
			}
			r.ScopeDepths[t] = depth - 1
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
func (r *Resolver) VisitRaiseStmt(rs *ast.RaiseStmt) ast.Visitor      { return r }
func (r *Resolver) VisitPassStmt(p *ast.PassStmt) ast.Visitor         { return r }
func (r *Resolver) VisitYieldStmt(y *ast.YieldStmt) ast.Visitor       { return r }
func (r *Resolver) VisitAssertStmt(a *ast.AssertStmt) ast.Visitor     { return r }
func (r *Resolver) VisitBreakStmt(b *ast.BreakStmt) ast.Visitor       { return r }
func (r *Resolver) VisitContinueStmt(c *ast.ContinueStmt) ast.Visitor { return r }

// convertDottedNameToPath converts a dotted name AST node to a module path string.
// For example, "os.path" becomes "os.path".
func convertDottedNameToPath(dottedName *ast.DottedName) string {
	if dottedName == nil {
		return ""
	}

	parts := make([]string, len(dottedName.Names))
	for i, name := range dottedName.Names {
		parts[i] = name.Token.Lexeme
	}
	return strings.Join(parts, ".")
}

// VisitImportStmt resolves import statements: import x, import x.y, import x as y
func (r *Resolver) VisitImportStmt(i *ast.ImportStmt) ast.Visitor {
	// If no module resolver available, skip import resolution
	if r.ModuleResolver == nil {
		return r
	}

	// For each imported module: import os.path as p
	for _, importName := range i.Names {
		// 1. Convert DottedName to module path string
		modulePath := convertDottedNameToPath(importName.DottedName)

		// 2. Resolve to file path using ModuleResolver
		filePath, err := r.ModuleResolver.ResolveAbsolute(context.Background(), modulePath)
		if err != nil {
			r.ReportError(fmt.Errorf("cannot import '%s': %w", modulePath, err))
			continue
		}

		// 3. Determine binding name (alias or first component)
		var bindingName string
		var nameNode *ast.Name
		if importName.AsName != nil {
			// import x.y as z -> bind as "z"
			bindingName = importName.AsName.Token.Lexeme
			nameNode = importName.AsName
		} else {
			// import x.y -> bind as "x"
			bindingName = importName.DottedName.Names[0].Token.Lexeme
			nameNode = importName.DottedName.Names[0]
		}

		// 4. Create imported variable binding
		variable := r.DefineImportedVariable(bindingName, importName.GetSpan())
		variable.ImportSource = filePath

		// 5. Track in Variables map for backward compatibility
		if nameNode != nil {
			r.Variables[nameNode] = variable
			r.ScopeDepths[nameNode] = 0 // Imports are always at module level

			// Track in new binding system
			if binding, exists := r.ScopeChain.Bindings[bindingName]; exists {
				r.NameToBinding[nameNode] = binding
			}
		}

		// Store the resolved file path for future dependency graph analysis
		// (Could be tracked in ResolutionTable.ImportedModules if we add that field)
		_ = filePath
	}

	return r
}

// VisitImportFromStmt resolves from...import statements: from x import y, from . import y, from x import *
func (r *Resolver) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor {
	// If no module resolver or symbol registry available, skip import resolution
	if r.ModuleResolver == nil || r.SymbolRegistry == nil {
		return r
	}

	// 1. Resolve module path
	var modulePath string
	var filePath string
	var err error

	if i.DotCount > 0 {
		// Relative import: from . import x, from .. import x, from .pkg import x
		modulePath = ""
		if i.DottedName != nil {
			modulePath = convertDottedNameToPath(i.DottedName)
		}
		filePath, err = r.ModuleResolver.ResolveRelative(
			context.Background(),
			i.DotCount,
			modulePath,
			r.SourceFilePath,
		)
		if err != nil {
			r.ReportError(fmt.Errorf("cannot resolve relative import: %w", err))
			return r
		}
	} else {
		// Absolute import: from x import y
		modulePath = convertDottedNameToPath(i.DottedName)
		filePath, err = r.ModuleResolver.ResolveAbsolute(
			context.Background(),
			modulePath,
		)
		if err != nil {
			r.ReportError(fmt.Errorf("cannot import from '%s': %w", modulePath, err))
			return r
		}
	}

	// 2. Handle wildcard vs specific imports
	if i.IsWildcard {
		// from module import *
		symbols, err := r.SymbolRegistry.GetPublicSymbols(filePath)
		if err != nil {
			r.ReportError(fmt.Errorf("cannot get symbols from '%s': %w", modulePath, err))
			return r
		}

		for _, sym := range symbols {
			variable := r.DefineImportedVariable(sym.Name, i.Span)
			variable.ImportSource = filePath
		}
	} else {
		// from module import x, y as z
		for _, importName := range i.Names {
			symbolName := importName.DottedName.Names[0].Token.Lexeme

			// Lookup symbol in registry
			sym, err := r.SymbolRegistry.LookupSymbol(filePath, symbolName)
			if err != nil {
				r.ReportError(fmt.Errorf("cannot import '%s' from '%s': %w",
					symbolName, modulePath, err))
				continue
			}

			// Determine local binding name
			bindingName := symbolName
			var nameNode *ast.Name
			if importName.AsName != nil {
				bindingName = importName.AsName.Token.Lexeme
				nameNode = importName.AsName
			} else {
				nameNode = importName.DottedName.Names[0]
			}

			// Create binding
			variable := r.DefineImportedVariable(bindingName, importName.GetSpan())
			variable.ImportSource = filePath

			// Track in Variables map for backward compatibility
			if nameNode != nil {
				r.Variables[nameNode] = variable
				r.ScopeDepths[nameNode] = 0 // Imports are always at module level

				// Track in new binding system
				if binding, exists := r.ScopeChain.Bindings[bindingName]; exists {
					r.NameToBinding[nameNode] = binding
				}
			}

			// Store reference to original symbol (for documentation/debugging)
			_ = sym
		}
	}

	return r
}
func (r *Resolver) VisitTypeAlias(t *ast.TypeAlias) ast.Visitor { return r }
func (r *Resolver) VisitDecorator(d *ast.Decorator) ast.Visitor { return r }
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
	// Check if this HTML element references a view (for composition)
	tagName := h.TagName.Lexeme

	// First check: same-file view
	if viewStmt, exists := r.Views[tagName]; exists {
		// This HTML element is actually a view reference - bind it
		r.ViewElements[h] = viewStmt
	} else if r.SymbolRegistry != nil {
		// Second check: imported view
		// Look up the name in module globals to see if it's imported
		if variable, exists := r.ModuleGlobals[tagName]; exists && variable.IsImported {
			var foundView *ast.ViewStmt

			// Try ImportSource first (O(1) lookup) if available
			if variable.ImportSource != "" {
				if sym, err := r.SymbolRegistry.LookupSymbol(variable.ImportSource, tagName); err == nil {
					if sym.Type == symbol.SymbolView {
						if viewStmt, ok := sym.Node.(*ast.ViewStmt); ok {
							foundView = viewStmt
						}
					}
				}
			}

			// If not found via ImportSource (e.g., re-exported from __init__.psx),
			// search all registered modules (O(n) fallback)
			if foundView == nil {
				for _, filePath := range r.SymbolRegistry.GetAllModules() {
					if sym, err := r.SymbolRegistry.LookupSymbol(filePath, tagName); err == nil {
						if sym.Type == symbol.SymbolView {
							if viewStmt, ok := sym.Node.(*ast.ViewStmt); ok {
								foundView = viewStmt
								break
							}
						}
					}
				}
			}

			if foundView != nil {
				r.ViewElements[h] = foundView
			}
		}
	}

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
