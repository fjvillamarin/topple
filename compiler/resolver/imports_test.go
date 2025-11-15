package resolver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
	"topple/compiler/module"
	"topple/compiler/symbol"
	"topple/internal/filesystem"
)

// mockFileSystem implements filesystem.FileSystem for testing
type mockFileSystem struct {
	files map[string]bool // path -> exists
}

func newMockFS(files map[string]bool) *mockFileSystem {
	return &mockFileSystem{files: files}
}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) Exists(path string) (bool, error) {
	exists, ok := m.files[path]
	if !ok {
		return false, nil
	}
	return exists, nil
}

func (m *mockFileSystem) IsDir(path string) (bool, error) {
	// Check if any file starts with this path + "/"
	prefix := path + string(filepath.Separator)
	for p := range m.files {
		if strings.HasPrefix(p, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockFileSystem) ListFiles(dir string, recursive bool) ([]string, error) {
	return nil, nil
}

func (m *mockFileSystem) ListPSXFiles(dir string, recursive bool) ([]string, error) {
	return nil, nil
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) ResolvePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (m *mockFileSystem) RelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

func (m *mockFileSystem) AbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (m *mockFileSystem) JoinPaths(elem ...string) string {
	return filepath.Join(elem...)
}

func (m *mockFileSystem) GetOutputPath(inputPath, outputDir string) (string, error) {
	return "", nil
}

func (m *mockFileSystem) WatchFiles(ctx context.Context, dirs []string, recursive bool) (<-chan filesystem.FileEvent, error) {
	return nil, nil
}

func (m *mockFileSystem) StopWatching() error {
	return nil
}

// Helper function to create a test Name node
func createName(name string) *ast.Name {
	return &ast.Name{
		Token: lexer.Token{Lexeme: name, Type: lexer.Identifier},
		Span:  lexer.Span{},
	}
}

// Helper function to create a DottedName from parts
func createDottedName(parts ...string) *ast.DottedName {
	names := make([]*ast.Name, len(parts))
	for i, part := range parts {
		names[i] = createName(part)
	}
	return &ast.DottedName{
		Names: names,
		Span:  lexer.Span{},
	}
}

// setupTestEnvironment creates a mock module resolver and symbol registry for testing
func setupTestEnvironment() (*module.StandardResolver, *symbol.Registry) {
	// Create mock filesystem
	mockFS := newMockFS(map[string]bool{
		"/project/utils.psx":               true,
		"/project/helpers.psx":             true,
		"/project/lib/core.psx":            true,
		"/project/components/__init__.psx": true,
		"/project/components/button.psx":   true,
	})

	// Create module resolver
	moduleResolver := module.NewResolver(module.Config{
		RootDir:    "/project",
		FileSystem: mockFS,
	})

	// Create symbol registry
	symbolRegistry := symbol.NewRegistry()

	// Register utils module with symbols
	utilsSymbols := symbol.NewModuleSymbols("/project/utils.psx")
	utilsSymbols.AddSymbol(&symbol.Symbol{
		Name:       "MyView",
		Type:       symbol.SymbolView,
		Visibility: symbol.Public,
	})
	utilsSymbols.AddSymbol(&symbol.Symbol{
		Name:       "helper_func",
		Type:       symbol.SymbolFunction,
		Visibility: symbol.Public,
	})
	utilsSymbols.AddSymbol(&symbol.Symbol{
		Name:       "_private_func",
		Type:       symbol.SymbolFunction,
		Visibility: symbol.Private,
	})
	symbolRegistry.RegisterModule("/project/utils.psx", utilsSymbols)

	// Register helpers module with symbols
	helpersSymbols := symbol.NewModuleSymbols("/project/helpers.psx")
	helpersSymbols.AddSymbol(&symbol.Symbol{
		Name:       "format_string",
		Type:       symbol.SymbolFunction,
		Visibility: symbol.Public,
	})
	symbolRegistry.RegisterModule("/project/helpers.psx", helpersSymbols)

	// Register lib.core module with symbols
	coreSymbols := symbol.NewModuleSymbols("/project/lib/core.psx")
	coreSymbols.AddSymbol(&symbol.Symbol{
		Name:       "CoreView",
		Type:       symbol.SymbolView,
		Visibility: symbol.Public,
	})
	symbolRegistry.RegisterModule("/project/lib/core.psx", coreSymbols)

	return moduleResolver, symbolRegistry
}

func TestImportStmt_SimpleImport(t *testing.T) {
	moduleResolver, _ := setupTestEnvironment()

	dottedName := createDottedName("utils")
	importStmt := &ast.ImportStmt{
		Names: []*ast.ImportName{
			{
				DottedName: dottedName,
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, nil, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that "utils" is defined as imported variable
	variable, exists := resolver.ModuleGlobals["utils"]
	if !exists {
		t.Fatal("Expected 'utils' to be defined in module globals")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}

	if variable.State != VariableDefined {
		t.Errorf("Expected variable state to be Defined, got %v", variable.State)
	}

	// Verify it's tracked in Variables map (using the first name from dotted name)
	if _, exists := resolver.Variables[dottedName.Names[0]]; !exists {
		t.Error("Expected import name to be tracked in Variables map")
	}
}

func TestImportStmt_WithAlias(t *testing.T) {
	moduleResolver, _ := setupTestEnvironment()

	aliasName := createName("u")
	importStmt := &ast.ImportStmt{
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("utils"),
				AsName:     aliasName,
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, nil, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that "u" (alias) is defined, not "utils"
	if _, exists := resolver.ModuleGlobals["utils"]; exists {
		t.Error("Expected 'utils' not to be defined (using alias)")
	}

	variable, exists := resolver.ModuleGlobals["u"]
	if !exists {
		t.Fatal("Expected 'u' (alias) to be defined in module globals")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}
}

func TestImportStmt_DottedPath(t *testing.T) {
	moduleResolver, _ := setupTestEnvironment()

	importStmt := &ast.ImportStmt{
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("lib", "core"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, nil, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// For "import lib.core", only "lib" should be bound
	variable, exists := resolver.ModuleGlobals["lib"]
	if !exists {
		t.Fatal("Expected 'lib' to be defined in module globals")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}
}

func TestImportStmt_ModuleNotFound(t *testing.T) {
	moduleResolver, _ := setupTestEnvironment()

	importStmt := &ast.ImportStmt{
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("nonexistent"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, nil, "/project/main.psx")
	table, err := resolver.Resolve(module)

	// Resolve() returns an error when there are errors in the table
	if err == nil {
		t.Fatal("Expected Resolve to return error for nonexistent module")
	}

	// Should have error about module not found
	if table == nil || len(table.Errors) == 0 {
		t.Fatal("Expected error for nonexistent module in table")
	}
}

func TestImportFromStmt_SpecificImport(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("MyView"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that "MyView" is defined
	variable, exists := resolver.ModuleGlobals["MyView"]
	if !exists {
		t.Fatal("Expected 'MyView' to be defined in module globals")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}
}

func TestImportFromStmt_WithAlias(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	aliasName := createName("MV")
	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("MyView"),
				AsName:     aliasName,
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that "MV" (alias) is defined, not "MyView"
	if _, exists := resolver.ModuleGlobals["MyView"]; exists {
		t.Error("Expected 'MyView' not to be defined (using alias)")
	}

	variable, exists := resolver.ModuleGlobals["MV"]
	if !exists {
		t.Fatal("Expected 'MV' (alias) to be defined in module globals")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}
}

func TestImportFromStmt_MultipleImports(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("MyView"),
				Span:       lexer.Span{},
			},
			{
				DottedName: createDottedName("helper_func"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check both symbols are defined
	expectedSymbols := []string{"MyView", "helper_func"}
	for _, name := range expectedSymbols {
		variable, exists := resolver.ModuleGlobals[name]
		if !exists {
			t.Errorf("Expected '%s' to be defined in module globals", name)
			continue
		}

		if !variable.IsImported {
			t.Errorf("Expected '%s' to be marked as imported", name)
		}
	}
}

func TestImportFromStmt_Wildcard(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		IsWildcard: true,
		Span:       lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that public symbols are imported (not private ones)
	expectedSymbols := []string{"MyView", "helper_func"}
	for _, name := range expectedSymbols {
		variable, exists := resolver.ModuleGlobals[name]
		if !exists {
			t.Errorf("Expected '%s' to be defined from wildcard import", name)
			continue
		}

		if !variable.IsImported {
			t.Errorf("Expected '%s' to be marked as imported", name)
		}
	}

	// Private symbols should NOT be imported
	if _, exists := resolver.ModuleGlobals["_private_func"]; exists {
		t.Error("Expected private symbol '_private_func' not to be imported with wildcard")
	}
}

func TestImportFromStmt_RelativeImport(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	// from .helpers import format_string (relative import from same directory)
	// Note: "from . import helpers" would try to import helpers module, but we're importing
	// from .helpers (the helpers.psx file) to get specific symbols
	importStmt := &ast.ImportFromStmt{
		DotCount:   1,
		DottedName: createDottedName("helpers"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("format_string"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// "format_string" should be defined
	variable, exists := resolver.ModuleGlobals["format_string"]
	if !exists {
		t.Fatal("Expected 'format_string' to be defined from relative import")
	}

	if !variable.IsImported {
		t.Error("Expected variable to be marked as imported")
	}
}

func TestImportFromStmt_SymbolNotFound(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("NonexistentSymbol"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	// Resolve() returns an error when there are errors in the table
	if err == nil {
		t.Fatal("Expected Resolve to return error for nonexistent symbol")
	}

	// Should have error about symbol not found
	if table == nil || len(table.Errors) == 0 {
		t.Fatal("Expected error for nonexistent symbol in table")
	}
}

func TestImportedVariableUsage(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	// from utils import MyView
	// x = MyView  # Using imported name
	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("MyView"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	usageName := createName("MyView")
	assignStmt := &ast.AssignStmt{
		Targets: []ast.Expr{createName("x")},
		Value:   usageName,
		Span:    lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt, assignStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	// Check that MyView is imported
	importedVar, exists := resolver.ModuleGlobals["MyView"]
	if !exists {
		t.Fatal("Expected 'MyView' to be defined")
	}

	if !importedVar.IsImported {
		t.Error("Expected variable to be marked as imported")
	}

	// Check that usage of MyView resolves correctly
	usageVar, exists := resolver.Variables[usageName]
	if !exists {
		t.Fatal("Expected usage of 'MyView' to be tracked")
	}

	if usageVar.Name != "MyView" {
		t.Errorf("Expected variable name 'MyView', got '%s'", usageVar.Name)
	}
}

func TestImportShadowing(t *testing.T) {
	moduleResolver, symbolRegistry := setupTestEnvironment()

	// from utils import MyView
	// MyView = "shadowed"  # Local variable shadows import
	importStmt := &ast.ImportFromStmt{
		DottedName: createDottedName("utils"),
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("MyView"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	assignStmt := &ast.AssignStmt{
		Targets: []ast.Expr{createName("MyView")},
		Value: &ast.Literal{
			Token: lexer.Token{Lexeme: "shadowed", Type: lexer.String},
			Span:  lexer.Span{},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt, assignStmt},
		Span: lexer.Span{},
	}

	resolver := NewResolverWithDeps(moduleResolver, symbolRegistry, "/project/main.psx")
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors (shadowing is allowed in Python), got: %v", table.Errors)
	}

	// MyView should exist in ModuleGlobals
	// The latest assignment should shadow the import
	variable, exists := resolver.ModuleGlobals["MyView"]
	if !exists {
		t.Fatal("Expected 'MyView' to be defined")
	}

	// The variable should be imported first, but then reassigned
	// In Python, imports are just assignments, so this is valid
	if !variable.IsImported {
		t.Error("Expected original variable to be marked as imported")
	}
}

func TestNoResolverProvided(t *testing.T) {
	// Test that imports are gracefully skipped when no resolver is provided
	importStmt := &ast.ImportStmt{
		Names: []*ast.ImportName{
			{
				DottedName: createDottedName("utils"),
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{importStmt},
		Span: lexer.Span{},
	}

	// Use NewResolver() which doesn't have module resolver
	resolver := NewResolver()
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Should have no errors and no imported symbols
	if len(table.Errors) > 0 {
		t.Fatalf("Expected no errors, got: %v", table.Errors)
	}

	if len(resolver.ModuleGlobals) > 0 {
		t.Error("Expected no symbols to be imported without module resolver")
	}
}

func TestConvertDottedNameToPath(t *testing.T) {
	tests := []struct {
		name     string
		input    *ast.DottedName
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "single name",
			input:    createDottedName("utils"),
			expected: "utils",
		},
		{
			name:     "dotted name",
			input:    createDottedName("os", "path"),
			expected: "os.path",
		},
		{
			name:     "deeply nested",
			input:    createDottedName("lib", "core", "utils"),
			expected: "lib.core.utils",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDottedNameToPath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
