package symbol

import (
	"topple/compiler/ast"
)

// Collector extracts exportable symbols from an AST
type Collector struct {
	filePath string             // Current file being processed
	symbols  map[string]*Symbol // Collected symbols
	errors   []error            // Collection errors
}

// NewCollector creates a new symbol collector
func NewCollector(filePath string) *Collector {
	return &Collector{
		filePath: filePath,
		symbols:  make(map[string]*Symbol),
		errors:   []error{},
	}
}

// CollectFromModule collects all top-level symbols from a module
func (c *Collector) CollectFromModule(module *ast.Module) *ModuleSymbols {
	// Reset state
	c.symbols = make(map[string]*Symbol)
	c.errors = []error{}

	// Visit all top-level statements
	for _, stmt := range module.Body {
		c.visitStatement(stmt)
	}

	// Create module symbols
	moduleSymbols := NewModuleSymbols(c.filePath)
	for _, symbol := range c.symbols {
		moduleSymbols.AddSymbol(symbol)
	}

	return moduleSymbols
}

// visitStatement processes a single statement
func (c *Collector) visitStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ViewStmt:
		c.addSymbol(s.Name.Token.Lexeme, SymbolView, s)
	case *ast.Function:
		c.addSymbol(s.Name.Token.Lexeme, SymbolFunction, s)
	case *ast.Class:
		c.addSymbol(s.Name.Token.Lexeme, SymbolClass, s)
	case *ast.AssignStmt:
		// Only collect simple module-level assignments
		c.collectAssignmentTargets(s)
	case *ast.AnnotationStmt:
		// Type annotations at module level (e.g., x: int)
		if name, ok := s.Target.(*ast.Name); ok {
			c.addSymbol(name.Token.Lexeme, SymbolVariable, s)
		}
		// Ignore other statements (imports, expressions, control flow, etc.)
	}
}

// collectAssignmentTargets extracts variable names from assignment targets
func (c *Collector) collectAssignmentTargets(assign *ast.AssignStmt) {
	for _, target := range assign.Targets {
		switch t := target.(type) {
		case *ast.Name:
			// Simple assignment: x = 1
			c.addSymbol(t.Token.Lexeme, SymbolVariable, assign)
		case *ast.TupleExpr:
			// Tuple unpacking: a, b = 1, 2
			c.collectTupleTargets(t, assign)
		case *ast.ListExpr:
			// List unpacking: [a, b] = [1, 2]
			c.collectListTargets(t, assign)
			// Ignore subscript, attribute assignments (not module-level symbols)
		}
	}
}

// collectTupleTargets extracts names from tuple unpacking
func (c *Collector) collectTupleTargets(tuple *ast.TupleExpr, assign *ast.AssignStmt) {
	for _, elem := range tuple.Elements {
		if name, ok := elem.(*ast.Name); ok {
			c.addSymbol(name.Token.Lexeme, SymbolVariable, assign)
		}
	}
}

// collectListTargets extracts names from list unpacking
func (c *Collector) collectListTargets(list *ast.ListExpr, assign *ast.AssignStmt) {
	for _, elem := range list.Elements {
		if name, ok := elem.(*ast.Name); ok {
			c.addSymbol(name.Token.Lexeme, SymbolVariable, assign)
		}
	}
}

// addSymbol adds a symbol to the collection
func (c *Collector) addSymbol(name string, symbolType SymbolType, node ast.Node) {
	// Check for duplicates
	if _, exists := c.symbols[name]; exists {
		// Note: We don't error on duplicates here, just overwrite
		// The resolver will handle duplicate definition errors
	}

	symbol := &Symbol{
		Name:       name,
		Type:       symbolType,
		Node:       node,
		Location:   extractLocation(node, c.filePath),
		Visibility: determineVisibility(name),
	}

	c.symbols[name] = symbol
}

// GetErrors returns any collection errors
func (c *Collector) GetErrors() []error {
	return c.errors
}
