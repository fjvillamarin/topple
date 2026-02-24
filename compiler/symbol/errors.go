package symbol

import (
	"fmt"
	"strings"
)

// ErrorType represents different types of symbol errors
type ErrorType int

const (
	ModuleNotRegistered ErrorType = iota
	SymbolNotFound
	DuplicateSymbol
	InvalidSymbol
)

// RegistryError represents a symbol registry error
type RegistryError struct {
	Type       ErrorType
	ModulePath string
	SymbolName string
	Location   *Location
	Message    string
}

// Error implements the error interface
func (e *RegistryError) Error() string {
	switch e.Type {
	case ModuleNotRegistered:
		return fmt.Sprintf("module not registered: %s", e.ModulePath)
	case SymbolNotFound:
		return fmt.Sprintf("symbol '%s' not found in module '%s'", e.SymbolName, e.ModulePath)
	case DuplicateSymbol:
		loc := ""
		if e.Location != nil {
			loc = fmt.Sprintf(" at %s:%d:%d", e.Location.File, e.Location.Line, e.Location.Column)
		}
		return fmt.Sprintf("duplicate symbol '%s' in module '%s'%s", e.SymbolName, e.ModulePath, loc)
	case InvalidSymbol:
		return fmt.Sprintf("invalid symbol: %s", e.Message)
	default:
		return fmt.Sprintf("symbol error: %s", e.Message)
	}
}

// newModuleNotRegisteredError creates a ModuleNotRegistered error
func newModuleNotRegisteredError(modulePath string) error {
	return &RegistryError{
		Type:       ModuleNotRegistered,
		ModulePath: modulePath,
	}
}

// newSymbolNotFoundError creates a SymbolNotFound error
func newSymbolNotFoundError(modulePath, symbolName string) error {
	return &RegistryError{
		Type:       SymbolNotFound,
		ModulePath: modulePath,
		SymbolName: symbolName,
	}
}

// newDuplicateSymbolError creates a DuplicateSymbol error
func newDuplicateSymbolError(modulePath, symbolName string, location *Location) error {
	return &RegistryError{
		Type:       DuplicateSymbol,
		ModulePath: modulePath,
		SymbolName: symbolName,
		Location:   location,
	}
}

// CollectionError represents errors during symbol collection
type CollectionError struct {
	FilePath string
	Errors   []error
}

// Error implements the error interface
func (e *CollectionError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("collection error in %s: unknown error", e.FilePath)
	}

	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("collection errors in %s:\n  - %s",
		e.FilePath,
		strings.Join(messages, "\n  - "))
}
