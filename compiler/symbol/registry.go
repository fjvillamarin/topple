package symbol

import (
	"sync"
)

// Registry tracks all exported symbols across modules
type Registry struct {
	modules map[string]*ModuleSymbols // Absolute file path -> symbols
	mu      sync.RWMutex              // Thread-safe access
}

// NewRegistry creates a new symbol registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]*ModuleSymbols),
	}
}

// RegisterModule registers symbols from a module
func (r *Registry) RegisterModule(filePath string, symbols *ModuleSymbols) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store the module symbols
	r.modules[filePath] = symbols
	return nil
}

// GetModuleSymbols retrieves all symbols from a module
func (r *Registry) GetModuleSymbols(filePath string) (*ModuleSymbols, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	symbols, exists := r.modules[filePath]
	if !exists {
		return nil, newModuleNotRegisteredError(filePath)
	}
	return symbols, nil
}

// LookupSymbol finds a specific symbol in a module
func (r *Registry) LookupSymbol(filePath string, symbolName string) (*Symbol, error) {
	moduleSymbols, err := r.GetModuleSymbols(filePath)
	if err != nil {
		return nil, err
	}

	symbol, exists := moduleSymbols.LookupSymbol(symbolName)
	if !exists {
		return nil, newSymbolNotFoundError(filePath, symbolName)
	}
	return symbol, nil
}

// GetPublicSymbols returns all public symbols from a module (for wildcard imports)
func (r *Registry) GetPublicSymbols(filePath string) ([]*Symbol, error) {
	moduleSymbols, err := r.GetModuleSymbols(filePath)
	if err != nil {
		return nil, err
	}
	return moduleSymbols.GetPublicSymbols(), nil
}

// HasModule checks if a module is registered
func (r *Registry) HasModule(filePath string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.modules[filePath]
	return exists
}

// Clear removes all registered modules (useful for testing)
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules = make(map[string]*ModuleSymbols)
}

// GetAllModules returns all registered module paths
func (r *Registry) GetAllModules() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paths := make([]string, 0, len(r.modules))
	for path := range r.modules {
		paths = append(paths, path)
	}
	return paths
}

// Stats returns statistics about the registry
type Stats struct {
	ModuleCount   int
	SymbolCount   int
	ViewCount     int
	FunctionCount int
	ClassCount    int
	VariableCount int
}

// GetStats returns registry statistics
func (r *Registry) GetStats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := Stats{
		ModuleCount: len(r.modules),
	}

	for _, module := range r.modules {
		for _, symbol := range module.Symbols {
			stats.SymbolCount++
			switch symbol.Type {
			case SymbolView:
				stats.ViewCount++
			case SymbolFunction:
				stats.FunctionCount++
			case SymbolClass:
				stats.ClassCount++
			case SymbolVariable:
				stats.VariableCount++
			}
		}
	}

	return stats
}
