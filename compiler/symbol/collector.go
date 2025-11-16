package symbol

import (
	"strings"

	"topple/compiler/ast"
	"topple/compiler/module"
)

// Collector extracts exportable symbols from an AST
type Collector struct {
	filePath       string                   // Current file being processed
	symbols        map[string]*Symbol       // Collected symbols
	registry       *Registry                // Symbol registry (for re-exports)
	moduleResolver *module.StandardResolver // Module resolver (for import paths)
}

// NewCollector creates a new symbol collector
func NewCollector(filePath string) *Collector {
	return &Collector{
		filePath:       filePath,
		symbols:        make(map[string]*Symbol),
		registry:       nil,
		moduleResolver: nil,
	}
}

// NewCollectorWithDeps creates a collector with registry and resolver for handling re-exports
func NewCollectorWithDeps(filePath string, registry *Registry, resolver *module.StandardResolver) *Collector {
	return &Collector{
		filePath:       filePath,
		symbols:        make(map[string]*Symbol),
		registry:       registry,
		moduleResolver: resolver,
	}
}

// CollectFromModule collects all top-level symbols from a module
func (c *Collector) CollectFromModule(module *ast.Module) *ModuleSymbols {
	// Reset state
	c.symbols = make(map[string]*Symbol)

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
	case *ast.ImportFromStmt:
		// Handle re-exports (e.g., from .layout import PageHeader)
		c.collectReexports(s)
	case *ast.ImportStmt:
		// Handle module imports (e.g., import foo as bar)
		c.collectModuleImports(s)
		// Ignore other statements (expressions, control flow, etc.)
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

// collectReexports handles ImportFrom statements for re-exporting symbols
func (c *Collector) collectReexports(stmt *ast.ImportFromStmt) {
	// If we don't have a registry or resolver, we can't resolve re-exports
	if c.registry == nil || c.moduleResolver == nil {
		return
	}

	// Resolve the module path
	var filePath string
	var err error

	if stmt.DotCount > 0 {
		// Relative import
		modulePath := ""
		if stmt.DottedName != nil {
			modulePath = convertDottedNameToPath(stmt.DottedName)
		}
		filePath, err = c.moduleResolver.ResolveRelative(
			nil, // context not used by resolver
			stmt.DotCount,
			modulePath,
			c.filePath,
		)
	} else {
		// Absolute import
		modulePath := convertDottedNameToPath(stmt.DottedName)
		filePath, err = c.moduleResolver.ResolveAbsolute(nil, modulePath)
	}

	if err != nil {
		// Can't resolve - skip re-export
		return
	}

	// Handle wildcard imports (from foo import *)
	if stmt.IsWildcard {
		symbols, err := c.registry.GetPublicSymbols(filePath)
		if err != nil {
			// Module not in registry yet - skip
			return
		}
		for _, symbol := range symbols {
			// Re-export the symbol (keep original type and node)
			c.symbols[symbol.Name] = symbol
		}
		return
	}

	// Handle specific imports (from foo import a, b as c)
	for _, importName := range stmt.Names {
		// Get the imported symbol name (simple name from DottedName)
		importedName := convertDottedNameToPath(importName.DottedName)

		// Get the binding name (after "as" if present)
		bindingName := importedName
		if importName.AsName != nil {
			bindingName = importName.AsName.Token.Lexeme
		}

		// Look up the symbol in the source module
		symbol, err := c.registry.LookupSymbol(filePath, importedName)
		if err != nil {
			// Symbol not found - skip
			continue
		}

		// Re-export the symbol with the binding name
		// Keep the original type and node, but update the name if aliased
		reexportedSymbol := &Symbol{
			Name:       bindingName,
			Type:       symbol.Type,
			Node:       symbol.Node,
			Location:   extractLocation(stmt, c.filePath),
			Visibility: determineVisibility(bindingName),
		}
		c.symbols[bindingName] = reexportedSymbol
	}
}

// collectModuleImports handles ImportStmt for module-level imports
func (c *Collector) collectModuleImports(stmt *ast.ImportStmt) {
	// Module imports (import foo, import bar as baz) are not typically re-exported
	// They create module references in the namespace but don't export the module itself
	// For now, we don't collect these as exportable symbols
}

// convertDottedNameToPath converts a DottedName AST node to a dot-separated path string
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
