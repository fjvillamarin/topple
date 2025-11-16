package module

import (
	"fmt"
	"strings"
)

// ErrorType categorizes module resolution errors
type ErrorType int

const (
	ModuleNotFound ErrorType = iota
	InvalidRelativeImport
	InvalidPath
	TooManyDots
)

// ResolutionError represents a module resolution failure
type ResolutionError struct {
	ImportPath    string
	SourceFile    string
	SearchedPaths []string
	ErrorType     ErrorType
	Details       string
}

func (e *ResolutionError) Error() string {
	var sb strings.Builder

	switch e.ErrorType {
	case ModuleNotFound:
		sb.WriteString(fmt.Sprintf("cannot resolve import '%s'", e.ImportPath))
		if e.SourceFile != "" {
			sb.WriteString(fmt.Sprintf("\n  in file: %s", e.SourceFile))
		}
		if len(e.SearchedPaths) > 0 {
			sb.WriteString("\n  searched:")
			for _, path := range e.SearchedPaths {
				sb.WriteString(fmt.Sprintf("\n    - %s", path))
			}
		}

	case InvalidRelativeImport:
		sb.WriteString(fmt.Sprintf("invalid relative import '%s'", e.ImportPath))
		if e.SourceFile != "" {
			sb.WriteString(fmt.Sprintf("\n  in file: %s", e.SourceFile))
		}
		if e.Details != "" {
			sb.WriteString(fmt.Sprintf("\n  %s", e.Details))
		}

	case TooManyDots:
		sb.WriteString(fmt.Sprintf("relative import has too many dots: %s", e.ImportPath))
		if e.SourceFile != "" {
			sb.WriteString(fmt.Sprintf("\n  in file: %s", e.SourceFile))
		}
		sb.WriteString("\n  cannot navigate above root directory")

	case InvalidPath:
		sb.WriteString(fmt.Sprintf("invalid import path: %s", e.ImportPath))
		if e.Details != "" {
			sb.WriteString(fmt.Sprintf("\n  %s", e.Details))
		}
	}

	return sb.String()
}

func newModuleNotFoundError(importPath, sourceFile string, searchedPaths []string) error {
	return &ResolutionError{
		ImportPath:    importPath,
		SourceFile:    sourceFile,
		SearchedPaths: searchedPaths,
		ErrorType:     ModuleNotFound,
	}
}

func newInvalidRelativeImportError(importPath, sourceFile, details string) error {
	return &ResolutionError{
		ImportPath: importPath,
		SourceFile: sourceFile,
		ErrorType:  InvalidRelativeImport,
		Details:    details,
	}
}

func newTooManyDotsError(importPath, sourceFile string) error {
	return &ResolutionError{
		ImportPath: importPath,
		SourceFile: sourceFile,
		ErrorType:  TooManyDots,
	}
}
