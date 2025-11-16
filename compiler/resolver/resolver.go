package resolver

import (
	"fmt"
	"topple/compiler/ast"
	"topple/compiler/lexer"
	"topple/compiler/module"
	"topple/compiler/symbol"
)

// Resolver implements variable resolution for Python-like scoping
type Resolver struct {
	// Scope management (new scope chain system)
	ScopeChain  *Scope         // Current scope in the chain
	AllScopes   map[int]*Scope // All scopes by ID
	NextScopeID int            // Scope ID generator

	// Module-level variables
	ModuleGlobals map[string]*Variable // Module-level variables

	// Resolution results (legacy pointer-based)
	Variables   map[*ast.Name]*Variable // Name node → Variable mapping (for backward compat)
	ScopeDepths map[*ast.Name]int       // Name node → scope distance (for backward compat)

	// Resolution results (new scope chain-based)
	NameToBinding map[*ast.Name]*Binding // Name → specific binding
	NodeScopes    map[ast.Node]*Scope    // AST node → declaring scope

	// Closure analysis
	CellVars map[string]bool // Variables needing cells
	FreeVars map[string]bool // Free variables

	// View composition support
	Views        map[string]*ast.ViewStmt           // View name → ViewStmt mapping
	ViewElements map[*ast.HTMLElement]*ast.ViewStmt // HTMLElement → ViewStmt mapping

	// Import resolution support
	ModuleResolver *module.StandardResolver // Resolves import paths to file paths
	SymbolRegistry *symbol.Registry         // Cross-file symbol registry
	SourceFilePath string                   // Current source file being resolved

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
	return NewResolverWithDeps(nil, nil, "")
}

// NewResolverWithDeps creates a resolver with import resolution support.
// Parameters:
//   - moduleResolver: resolves import paths to file paths (optional, can be nil)
//   - symbolRegistry: cross-file symbol registry (optional, can be nil)
//   - sourceFilePath: path to the current source file being resolved (optional, can be empty)
func NewResolverWithDeps(moduleResolver *module.StandardResolver, symbolRegistry *symbol.Registry, sourceFilePath string) *Resolver {
	resolver := &Resolver{
		AllScopes:      make(map[int]*Scope),
		NextScopeID:    0,
		ModuleGlobals:  make(map[string]*Variable),
		Variables:      make(map[*ast.Name]*Variable),
		ScopeDepths:    make(map[*ast.Name]int),
		NameToBinding:  make(map[*ast.Name]*Binding),
		NodeScopes:     make(map[ast.Node]*Scope),
		CellVars:       make(map[string]bool),
		FreeVars:       make(map[string]bool),
		Errors:         []error{},
		Views:          make(map[string]*ast.ViewStmt),
		ViewElements:   make(map[*ast.HTMLElement]*ast.ViewStmt),
		ModuleResolver: moduleResolver,
		SymbolRegistry: symbolRegistry,
		SourceFilePath: sourceFilePath,
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
		NameToBinding:  r.NameToBinding,
		Scopes:         r.AllScopes,
		NodeScopes:     r.NodeScopes,
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
	newScope := NewScope(r.NextScopeID, scopeType, r.ScopeChain)
	r.NextScopeID++
	r.AllScopes[newScope.ID] = newScope
	r.ScopeChain = newScope

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
	if r.ScopeChain != nil {
		// Update context counters based on scope being removed
		switch r.ScopeChain.ScopeType {
		case ClassScopeType:
			r.ClassScopeDepth--
		case FunctionScopeType:
			r.FunctionScopeDepth--
		case ViewScopeType:
			r.ViewScopeDepth--
		}

		// Pop scope chain
		r.ScopeChain = r.ScopeChain.Parent
	}
}

// InFunctionScope returns true if we're currently in any function scope
func (r *Resolver) InFunctionScope() bool {
	return r.FunctionScopeDepth > 0 || r.ViewScopeDepth > 0
}

// DefineVariable creates a new variable in the current scope
func (r *Resolver) DefineVariable(name string, span lexer.Span) *Variable {
	// Calculate depth by walking up the scope chain
	depth := 0
	for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
		depth++
	}
	depth-- // Adjust for current scope

	variable := &Variable{
		Name:            name,
		DefinitionDepth: depth,
		State:           VariableDeclared,
		FirstDefSpan:    span,
	}

	if r.ScopeChain != nil {
		// Create binding in current scope
		binding := &Binding{
			Name:     name,
			Variable: variable,
			Scope:    r.ScopeChain,
		}

		// Check for shadowing - look in parent scopes for same name
		if r.ScopeChain.Parent != nil {
			if parentBinding := r.ScopeChain.Parent.ResolveBinding(name); parentBinding != nil {
				parentBinding.ShadowedBy = binding
			}
		}

		// Add binding to current scope
		r.ScopeChain.Bindings[name] = binding

		// Also store module-level variables in ModuleGlobals for LEGB resolution
		if r.ScopeChain.ScopeType == ModuleScopeType {
			r.ModuleGlobals[name] = variable
		}
	} else {
		// No current scope, store in module globals
		r.ModuleGlobals[name] = variable
	}

	return variable
}

// DefineImportedVariable creates a binding for an imported symbol.
// It marks the variable as imported and defined, since imports are always resolved at module scope.
func (r *Resolver) DefineImportedVariable(name string, span lexer.Span) *Variable {
	variable := r.DefineVariable(name, span)
	variable.IsImported = true
	variable.State = VariableDefined
	return variable
}

// ResolveName implements LEGB resolution for a name
func (r *Resolver) ResolveName(name *ast.Name) error {
	varName := name.Token.Lexeme

	// Helper function to calculate absolute scope depth (from module scope)
	calculateDepth := func(targetScope *Scope) int {
		depth := 0
		for scope := targetScope; scope != nil && scope.Parent != nil; scope = scope.Parent {
			depth++
		}
		return depth
	}

	// Check for global/nonlocal declarations in current scope
	if r.ScopeChain != nil {
		if globalBinding := r.ScopeChain.Globals[varName]; globalBinding != nil {
			r.Variables[name] = globalBinding.Variable
			r.NameToBinding[name] = globalBinding
			r.NodeScopes[name] = globalBinding.Scope
			globalBinding.Variable.IsUsed = true
			r.ScopeDepths[name] = 0 // Global is always depth 0
			return nil
		}

		if nonlocalBinding := r.ScopeChain.Nonlocals[varName]; nonlocalBinding != nil {
			r.Variables[name] = nonlocalBinding.Variable
			r.NameToBinding[name] = nonlocalBinding
			r.NodeScopes[name] = nonlocalBinding.Scope
			nonlocalBinding.Variable.IsUsed = true
			nonlocalBinding.Variable.IsCaptured = true
			nonlocalBinding.Variable.IsCell = true

			// Calculate depth
			r.ScopeDepths[name] = calculateDepth(nonlocalBinding.Scope)

			// This is a free variable in the current scope
			r.FreeVars[varName] = true
			r.CellVars[varName] = true
			return nil
		}
	}

	// LEGB Resolution: Local → Enclosing → Global → Builtin
	// Walk the scope chain looking for the name
	for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
		// Class scopes don't participate in LEGB for nested functions
		if scope.IsClassScope && r.InFunctionScope() {
			continue
		}

		if binding, exists := scope.Bindings[varName]; exists {
			r.Variables[name] = binding.Variable
			r.NameToBinding[name] = binding
			r.NodeScopes[name] = scope
			binding.Variable.IsUsed = true

			// Calculate depth
			depth := calculateDepth(scope)
			r.ScopeDepths[name] = depth

			// Mark as captured if we're accessing from a nested scope
			if scope != r.ScopeChain && (r.InFunctionScope() || r.ViewScopeDepth > 0) {
				binding.Variable.IsCaptured = true
				binding.Variable.IsCell = true
				r.CellVars[varName] = true
				r.FreeVars[varName] = true
			}

			return nil
		}
	}

	// Check global scope (module level)
	if variable, exists := r.ModuleGlobals[varName]; exists {
		r.Variables[name] = variable
		variable.IsUsed = true
		r.ScopeDepths[name] = 0 // Module scope is depth 0
		// No binding to track for module globals defined before scope chain
		return nil
	}

	// Built-ins (handled at runtime, don't error here)
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
