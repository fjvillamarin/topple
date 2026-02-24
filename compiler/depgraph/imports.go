package depgraph

import (
	"context"
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/module"
	"strings"
)

// ExtractImports extracts all imports from a module AST.
// It resolves import paths using the provided module resolver.
// Unresolved imports are skipped (they will be reported as errors during resolution phase).
func ExtractImports(
	astModule *ast.Module,
	sourceFile string,
	resolver module.Resolver,
) ([]*Import, error) {
	extractor := &importExtractor{
		sourceFile: sourceFile,
		resolver:   resolver,
		imports:    []*Import{},
	}

	// Visit all top-level statements
	for _, stmt := range astModule.Body {
		extractor.visitStatement(stmt)
	}

	return extractor.imports, nil
}

// importExtractor is a stateful visitor for extracting imports
type importExtractor struct {
	sourceFile string
	resolver   module.Resolver
	imports    []*Import
}

// visitStatement checks if a statement is an import and extracts it
func (e *importExtractor) visitStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ImportStmt:
		e.handleImportStmt(s)
	case *ast.ImportFromStmt:
		e.handleImportFromStmt(s)
	}
}

// handleImportStmt processes "import x" and "import x as y" statements
func (e *importExtractor) handleImportStmt(stmt *ast.ImportStmt) {
	for _, name := range stmt.Names {
		modulePath := convertDottedNameToPath(name.DottedName)
		filePath, err := e.resolver.ResolveAbsolute(context.Background(), modulePath)
		if err != nil {
			// Skip unresolved imports - they will be caught during resolution
			continue
		}

		e.imports = append(e.imports, &Import{
			Statement:  stmt,
			ModulePath: filePath,
			Names:      []string{}, // import x doesn't import specific names
			IsWildcard: false,
			Location:   extractLocation(stmt),
		})
	}
}

// handleImportFromStmt processes "from x import y" and "from . import y" statements
func (e *importExtractor) handleImportFromStmt(stmt *ast.ImportFromStmt) {
	var filePath string
	var err error

	if stmt.DotCount > 0 {
		// Relative import: from . import x, from ..pkg import y
		modulePath := ""
		if stmt.DottedName != nil {
			modulePath = convertDottedNameToPath(stmt.DottedName)
		}
		filePath, err = e.resolver.ResolveRelative(
			context.Background(),
			stmt.DotCount,
			modulePath,
			e.sourceFile,
		)
	} else {
		// Absolute import: from x import y
		modulePath := convertDottedNameToPath(stmt.DottedName)
		filePath, err = e.resolver.ResolveAbsolute(context.Background(), modulePath)
	}

	if err != nil {
		// Skip unresolved imports - they will be caught during resolution
		return
	}

	var names []string
	if !stmt.IsWildcard {
		names = make([]string, len(stmt.Names))
		for i, name := range stmt.Names {
			names[i] = name.DottedName.Names[0].Token.Lexeme
		}
	}

	e.imports = append(e.imports, &Import{
		Statement:  stmt,
		ModulePath: filePath,
		Names:      names,
		IsWildcard: stmt.IsWildcard,
		Location:   extractLocation(stmt),
	})
}

// convertDottedNameToPath converts a dotted name AST node to a module path string
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

// extractLocation extracts source location from an AST node
func extractLocation(node ast.Node) Location {
	span := node.GetSpan()
	return Location{
		Line:   span.Start.Line,
		Column: span.Start.Column,
	}
}
