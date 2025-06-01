package resolver

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
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

// Environment represents a scope in the scope chain
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
	Variables      map[*ast.Name]*Variable // Name → Variable mapping
	ScopeDepths    map[*ast.Name]int       // Name → scope distance
	ViewParameters map[string]*Variable    // View parameter tracking
	CellVars       map[string]bool         // Variables needing cells
	FreeVars       map[string]bool         // Free variables
	Errors         []error                 // Resolution errors

	// View composition support
	Views        map[string]*ast.ViewStmt           // View name → ViewStmt mapping (module level views)
	ViewElements map[*ast.HTMLElement]*ast.ViewStmt // HTMLElement → ViewStmt mapping (for composition)
}

// NewResolutionTable returns a new ResolutionTable with all internal maps and slices initialized for variable resolution and view composition.
func NewResolutionTable() *ResolutionTable {
	return &ResolutionTable{
		Variables:      make(map[*ast.Name]*Variable),
		ScopeDepths:    make(map[*ast.Name]int),
		ViewParameters: make(map[string]*Variable),
		CellVars:       make(map[string]bool),
		FreeVars:       make(map[string]bool),
		Errors:         []error{},
		Views:          make(map[string]*ast.ViewStmt),
		ViewElements:   make(map[*ast.HTMLElement]*ast.ViewStmt),
	}
}
