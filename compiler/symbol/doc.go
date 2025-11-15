// Package symbol provides cross-file symbol tracking for the Topple compiler.
//
// The symbol registry collects and manages exportable symbols (views, functions,
// classes, variables) from parsed PSX modules. It supports:
//   - Symbol collection from AST nodes
//   - Symbol lookup by module path and name
//   - Wildcard import expansion (all public symbols)
//   - Symbol visibility rules (public vs private)
//
// # Usage
//
//	registry := symbol.NewRegistry()
//
//	// Collect symbols from a parsed module
//	collector := symbol.NewCollector(filePath)
//	moduleSymbols := collector.CollectFromModule(astModule)
//	registry.RegisterModule(filePath, moduleSymbols)
//
//	// Lookup symbols
//	symbol, err := registry.LookupSymbol(filePath, "MyView")
//	publicSymbols, err := registry.GetPublicSymbols(filePath)
//
// # Integration
//
// Phase 1: Module Resolver provides file paths for resolution
//
// Phase 3: Resolver uses registry for import statement processing
package symbol
