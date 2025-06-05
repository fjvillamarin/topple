package resolver

import (
	"fmt"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// Resolver implements variable resolution for Python-like scoping
type Resolver struct {
	// Scope management
	Scopes        []*Environment       // Stack of active scopes
	Current       *Environment         // Current scope
	ModuleGlobals map[string]*Variable // Module-level variables

	// Resolution results
	Variables   map[*ast.Name]*Variable // Name node → Variable mapping
	ScopeDepths map[*ast.Name]int       // Name node → scope distance

	// Closure analysis
	CellVars map[string]bool // Variables needing cells
	FreeVars map[string]bool // Free variables

	// View composition support
	Views        map[string]*ast.ViewStmt           // View name → ViewStmt mapping
	ViewElements map[*ast.HTMLElement]*ast.ViewStmt // HTMLElement → ViewStmt mapping

	// Error tracking
	Errors []error

	// Context tracking - use counters for nested scopes
	FunctionScopeDepth int // How many function scopes deep we are
	ClassScopeDepth    int // How many class scopes deep we are
	ViewScopeDepth     int // How many view scopes deep we are
	CurrentFunction    *ast.Function
	CurrentView        *ast.ViewStmt

	// Embed visitor interface
	ast.Visitor
}

// NewResolver constructs and initializes a new Resolver for variable and view resolution within a module.
func NewResolver() *Resolver {
	resolver := &Resolver{
		Scopes:        []*Environment{},
		ModuleGlobals: make(map[string]*Variable),
		Variables:     make(map[*ast.Name]*Variable),
		ScopeDepths:   make(map[*ast.Name]int),
		CellVars:      make(map[string]bool),
		FreeVars:      make(map[string]bool),
		Errors:        []error{},
		Views:         make(map[string]*ast.ViewStmt),
		ViewElements:  make(map[*ast.HTMLElement]*ast.ViewStmt),
	}

	// Begin with module scope
	resolver.BeginScope(ModuleScopeType)

	return resolver
}

// Resolve performs variable resolution on the given module
func (r *Resolver) Resolve(module *ast.Module) (*ResolutionTable, error) {
	// Visit the module to perform resolution
	module.Accept(r)

	// Create and return resolution table
	table := &ResolutionTable{
		Variables:      r.Variables,
		ScopeDepths:    r.ScopeDepths,
		ViewParameters: make(map[string]*Variable),
		CellVars:       r.CellVars,
		FreeVars:       r.FreeVars,
		Errors:         r.Errors,
		Views:          r.Views,
		ViewElements:   r.ViewElements,
	}

	// Extract view parameters
	for name, variable := range r.ModuleGlobals {
		if variable.IsViewParameter {
			table.ViewParameters[name] = variable
		}
	}
	// Extract view parameters from all resolved variables
	for _, variable := range r.Variables {
		if variable.IsViewParameter {
			table.ViewParameters[variable.Name] = variable
		}
	}

	if len(r.Errors) > 0 {
		return table, fmt.Errorf("resolution failed with %d errors", len(r.Errors))
	}

	return table, nil
}

// BeginScope creates a new scope
func (r *Resolver) BeginScope(scopeType ScopeType) {
	env := &Environment{
		Enclosing:    r.Current,
		Values:       make(map[string]*Variable),
		ScopeType:    scopeType,
		Globals:      make(map[string]*Variable),
		Nonlocals:    make(map[string]*Variable),
		IsClassScope: scopeType == ClassScopeType,
	}
	r.Scopes = append(r.Scopes, env)
	r.Current = env

	// Update context counters
	switch scopeType {
	case ClassScopeType:
		r.ClassScopeDepth++
	case FunctionScopeType:
		r.FunctionScopeDepth++
	case ViewScopeType:
		r.ViewScopeDepth++
	}
}

// EndScope removes the current scope
func (r *Resolver) EndScope() {
	if len(r.Scopes) > 0 {
		// Update context counters based on scope being removed
		switch r.Current.ScopeType {
		case ClassScopeType:
			r.ClassScopeDepth--
		case FunctionScopeType:
			r.FunctionScopeDepth--
		case ViewScopeType:
			r.ViewScopeDepth--
		}

		r.Scopes = r.Scopes[:len(r.Scopes)-1]
		if len(r.Scopes) > 0 {
			r.Current = r.Scopes[len(r.Scopes)-1]
		} else {
			r.Current = nil
		}
	}
}

// InFunctionScope returns true if we're currently in any function scope
func (r *Resolver) InFunctionScope() bool {
	return r.FunctionScopeDepth > 0 || r.ViewScopeDepth > 0
}

// DefineVariable creates a new variable in the current scope
func (r *Resolver) DefineVariable(name string, span lexer.Span) *Variable {
	variable := &Variable{
		Name:            name,
		DefinitionDepth: len(r.Scopes) - 1,
		State:           VariableDeclared,
		FirstDefSpan:    span,
	}

	if r.Current != nil {
		r.Current.Values[name] = variable
		// Also store module-level variables in ModuleGlobals for LEGB resolution
		if r.Current.ScopeType == ModuleScopeType {
			r.ModuleGlobals[name] = variable
		}
	} else {
		// No current scope, store in module globals
		r.ModuleGlobals[name] = variable
	}

	return variable
}

// ResolveName implements LEGB resolution for a name
func (r *Resolver) ResolveName(name *ast.Name) error {
	varName := name.Token.Lexeme

	// Check for global/nonlocal declarations in current scope
	if r.Current != nil {
		if globalVar := r.Current.Globals[varName]; globalVar != nil {
			r.Variables[name] = globalVar
			globalVar.IsUsed = true
			r.ScopeDepths[name] = 0 // Global is always depth 0
			return nil
		}

		if nonlocalVar := r.Current.Nonlocals[varName]; nonlocalVar != nil {
			r.Variables[name] = nonlocalVar
			nonlocalVar.IsUsed = true
			nonlocalVar.IsCaptured = true
			nonlocalVar.IsCell = true

			// Find the actual depth of the nonlocal variable
			distance := 0
			for i := len(r.Scopes) - 2; i >= 0; i-- {
				distance++
				scope := r.Scopes[i]
				if variable, exists := scope.Values[varName]; exists && variable == nonlocalVar {
					r.ScopeDepths[name] = distance
					break
				}
			}

			// This is a free variable in the current scope
			r.FreeVars[varName] = true
			r.CellVars[varName] = true
			return nil
		}
	}

	// LEGB Resolution: Local → Enclosing → Global → Builtin

	// 1. Local scope
	if r.Current != nil {
		if variable, exists := r.Current.Values[varName]; exists {
			r.Variables[name] = variable
			variable.IsUsed = true
			// Local variables have depth based on current scope distance from module
			r.ScopeDepths[name] = len(r.Scopes) - 1
			return nil
		}
	}

	// 2. Enclosing scopes (skip class scopes for function lookups)
	distance := 1
	for i := len(r.Scopes) - 2; i >= 1; i-- {
		scope := r.Scopes[i]

		// Class scopes don't participate in LEGB for nested functions
		if scope.IsClassScope && r.InFunctionScope() {
			distance++
			continue
		}

		if variable, exists := scope.Values[varName]; exists {
			r.Variables[name] = variable
			variable.IsUsed = true

			// Mark as captured if we're accessing from a nested scope
			if r.InFunctionScope() || r.ViewScopeDepth > 0 {
				variable.IsCaptured = true
				variable.IsCell = true
				r.CellVars[varName] = true
				// This is a free variable in the current scope
				r.FreeVars[varName] = true
			}

			// Set the scope depth (the scope index where variable was found)
			r.ScopeDepths[name] = i
			return nil
		}
		distance++
	}

	// 3. Global scope (module level)
	if variable, exists := r.ModuleGlobals[varName]; exists {
		r.Variables[name] = variable
		variable.IsUsed = true
		r.ScopeDepths[name] = 0 // Module scope is depth 0
		return nil
	}

	// 4. Built-ins (handled at runtime, don't error here)
	// For now, create a placeholder global variable for unknown names
	variable := &Variable{
		Name:            varName,
		DefinitionDepth: 0,
		State:           VariableUndefined,
		FirstUseSpan:    name.Span,
	}
	r.ModuleGlobals[varName] = variable
	r.Variables[name] = variable
	variable.IsUsed = true
	r.ScopeDepths[name] = 0 // Unknown variables are treated as global

	return nil
}

// ReportError adds an error to the error list
func (r *Resolver) ReportError(err error) {
	r.Errors = append(r.Errors, err)
}
