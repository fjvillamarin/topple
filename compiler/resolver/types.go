package resolver

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// VariableState represents the current state of a variable
type VariableState int

const (
	VariableUndefined VariableState = iota
	VariableDeclared                // Declared but not assigned
	VariableDefined                 // Has been assigned
	VariableUsed                    // Has been referenced
)

// Variable represents metadata about a variable in the code
type Variable struct {
	Name            string        // Variable name
	DefinitionDepth int           // Absolute scope depth where defined
	State           VariableState // Current state of the variable

	// Scope behavior flags
	IsParameter     bool // Function/view parameter
	IsGlobal        bool // Declared with 'global'
	IsNonlocal      bool // Declared with 'nonlocal'
	IsImported      bool // Bound by import statement
	IsViewParameter bool // Biscuit view parameter
	IsExceptionVar  bool // Exception handler variable
	IsUsed          bool // Has been referenced

	// Usage tracking
	FirstDefSpan  lexer.Span // Where first defined
	FirstUseSpan  lexer.Span // Where first used
	UsedBeforeDef bool       // Python late binding detection

	// Closure analysis
	IsCaptured bool // Used in nested functions
	IsCell     bool // Needs cell variable for closures
}

// ScopeType represents the type of scope
type ScopeType int

const (
	ModuleScopeType ScopeType = iota
	FunctionScopeType
	ClassScopeType
	ViewScopeType // Biscuit view scope
	ComprehensionScopeType
	ExceptScopeType // Exception handler scope
	WithScopeType   // With statement scope
)

// Scope represents a lexical scope in the source code
type Scope struct {
	// Identification
	ID        int       // Unique scope identifier
	ScopeType ScopeType // Type of scope (module, function, view, etc.)
	Parent    *Scope    // Parent scope (for chain traversal)

	// Variable bindings in this scope
	Bindings map[string]*Binding // Variables defined at THIS level only

	// Python scope modifiers
	Globals   map[string]*Binding // 'global' declarations in this scope
	Nonlocals map[string]*Binding // 'nonlocal' declarations in this scope

	// Metadata
	IsClassScope bool     // Special LEGB handling for classes
	ASTNode      ast.Node // Associated AST node (ViewStmt, Function, etc.)
}

// Binding represents a single binding of a variable name in a specific scope
type Binding struct {
	Name       string    // Variable name
	Variable   *Variable // Variable metadata (shared across scopes)
	Scope      *Scope    // Scope where this binding was created
	ShadowedBy *Binding  // Points to inner scope binding that shadows this
}

// NewScope creates a new scope with the given ID, type, and parent
func NewScope(id int, scopeType ScopeType, parent *Scope) *Scope {
	scope := &Scope{
		ID:           id,
		ScopeType:    scopeType,
		Parent:       parent,
		Bindings:     make(map[string]*Binding),
		Globals:      make(map[string]*Binding),
		Nonlocals:    make(map[string]*Binding),
		IsClassScope: scopeType == ClassScopeType,
	}
	return scope
}

// GetBinding looks up a binding in this scope only (does not traverse parent chain)
func (s *Scope) GetBinding(name string) (*Binding, bool) {
	if s == nil {
		return nil, false
	}
	binding, exists := s.Bindings[name]
	return binding, exists
}

// ResolveBinding looks up a binding in the scope chain (this scope and all parents)
func (s *Scope) ResolveBinding(name string) *Binding {
	for scope := s; scope != nil; scope = scope.Parent {
		if binding, exists := scope.Bindings[name]; exists {
			return binding
		}
	}
	return nil
}

// Environment represents a scope in the scope chain (DEPRECATED - will be removed)
type Environment struct {
	Enclosing *Environment         // Parent scope
	Values    map[string]*Variable // Variable bindings in this scope
	ScopeType ScopeType            // Type of scope

	// Python-specific scope declarations
	Globals      map[string]*Variable // 'global' declarations
	Nonlocals    map[string]*Variable // 'nonlocal' declarations
	IsClassScope bool                 // Class scopes have special rules
}

// ResolutionTable holds the results of variable resolution
type ResolutionTable struct {
	// Legacy pointer-based lookups (kept for backward compatibility)
	Variables   map[*ast.Name]*Variable // Name → Variable mapping
	ScopeDepths map[*ast.Name]int       // Name → scope distance

	// New scope chain-based lookups
	NameToBinding map[*ast.Name]*Binding // Name → specific binding
	Scopes        map[int]*Scope         // Scope ID → Scope
	NodeScopes    map[ast.Node]*Scope    // AST node → declaring scope

	// View tracking
	ViewParameters map[string]*Variable // View parameter tracking

	// Closure analysis
	CellVars map[string]bool // Variables needing cells
	FreeVars map[string]bool // Free variables

	// Errors
	Errors []error // Resolution errors

	// View composition support
	Views        map[string]*ast.ViewStmt           // View name → ViewStmt mapping (module level views)
	ViewElements map[*ast.HTMLElement]*ast.ViewStmt // HTMLElement → ViewStmt mapping (for composition)
}

// NewResolutionTable returns a new ResolutionTable with all internal maps and slices initialized for variable resolution and view composition.
func NewResolutionTable() *ResolutionTable {
	return &ResolutionTable{
		Variables:      make(map[*ast.Name]*Variable),
		ScopeDepths:    make(map[*ast.Name]int),
		NameToBinding:  make(map[*ast.Name]*Binding),
		Scopes:         make(map[int]*Scope),
		NodeScopes:     make(map[ast.Node]*Scope),
		ViewParameters: make(map[string]*Variable),
		CellVars:       make(map[string]bool),
		FreeVars:       make(map[string]bool),
		Errors:         []error{},
		Views:          make(map[string]*ast.ViewStmt),
		ViewElements:   make(map[*ast.HTMLElement]*ast.ViewStmt),
	}
}
