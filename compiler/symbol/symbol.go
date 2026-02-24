package symbol

import (
	"github.com/fjvillamarin/topple/compiler/ast"
)

// SymbolType represents the type of symbol
type SymbolType int

const (
	SymbolView SymbolType = iota
	SymbolFunction
	SymbolClass
	SymbolVariable
)

// String returns the string representation of SymbolType
func (st SymbolType) String() string {
	switch st {
	case SymbolView:
		return "view"
	case SymbolFunction:
		return "function"
	case SymbolClass:
		return "class"
	case SymbolVariable:
		return "variable"
	default:
		return "unknown"
	}
}

// Visibility represents symbol visibility
type Visibility int

const (
	Public  Visibility = iota // Normal names
	Private                   // Underscore-prefixed names
)

// String returns the string representation of Visibility
func (v Visibility) String() string {
	switch v {
	case Public:
		return "public"
	case Private:
		return "private"
	default:
		return "unknown"
	}
}

// Symbol represents an exported name from a module
type Symbol struct {
	Name       string     // Symbol name (e.g., "MyView", "my_function")
	Type       SymbolType // Type of symbol
	Node       ast.Node   // Original AST node
	Location   Location   // Source location
	Visibility Visibility // Public or private
	Docstring  string     // Documentation (for future use)
}

// Location represents a position in source code
type Location struct {
	File   string // Absolute file path
	Line   int    // Line number (1-indexed)
	Column int    // Column number (1-indexed)
}

// ModuleSymbols represents all symbols exported from one module
type ModuleSymbols struct {
	FilePath string             // Absolute file path
	Symbols  map[string]*Symbol // Symbol name -> Symbol
}

// NewModuleSymbols creates a new ModuleSymbols
func NewModuleSymbols(filePath string) *ModuleSymbols {
	return &ModuleSymbols{
		FilePath: filePath,
		Symbols:  make(map[string]*Symbol),
	}
}

// AddSymbol adds a symbol to the module
func (ms *ModuleSymbols) AddSymbol(symbol *Symbol) {
	ms.Symbols[symbol.Name] = symbol
}

// LookupSymbol finds a symbol by name
func (ms *ModuleSymbols) LookupSymbol(name string) (*Symbol, bool) {
	symbol, exists := ms.Symbols[name]
	return symbol, exists
}

// GetPublicSymbols returns all public symbols
func (ms *ModuleSymbols) GetPublicSymbols() []*Symbol {
	var public []*Symbol
	for _, symbol := range ms.Symbols {
		if symbol.Visibility == Public {
			public = append(public, symbol)
		}
	}
	return public
}

// GetAllSymbols returns all symbols (public and private)
func (ms *ModuleSymbols) GetAllSymbols() []*Symbol {
	symbols := make([]*Symbol, 0, len(ms.Symbols))
	for _, symbol := range ms.Symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// extractLocation creates a Location from an AST node
func extractLocation(node ast.Node, filePath string) Location {
	span := node.GetSpan()
	return Location{
		File:   filePath,
		Line:   span.Start.Line,
		Column: span.Start.Column,
	}
}

// determineVisibility checks if a name is private (underscore-prefixed)
func determineVisibility(name string) Visibility {
	if len(name) > 0 && name[0] == '_' {
		return Private
	}
	return Public
}
