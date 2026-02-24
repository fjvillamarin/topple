package depgraph

import (
	"context"
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"github.com/fjvillamarin/topple/compiler/module"
	"github.com/fjvillamarin/topple/internal/filesystem"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// mockResolver implements module.Resolver for testing
type mockResolver struct {
	files map[string]string // Module path -> file path
}

func newMockResolver(files map[string]string) module.Resolver {
	mockFS := &mockFileSystem{files: make(map[string]bool)}
	for _, filePath := range files {
		mockFS.files[filePath] = true
	}

	return module.NewResolver(module.Config{
		RootDir:    "/project",
		FileSystem: mockFS,
	})
}

// mockFileSystem implements filesystem.FileSystem for testing
type mockFileSystem struct {
	files map[string]bool
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

// Helper functions

func createEmptyModule() *ast.Module {
	return &ast.Module{
		Body: []ast.Stmt{},
		Span: lexer.Span{},
	}
}

// === Graph Construction Tests ===

func TestNewGraph(t *testing.T) {
	graph := NewGraph()

	if graph == nil {
		t.Fatal("NewGraph() returned nil")
	}

	if graph.FileCount() != 0 {
		t.Errorf("new graph should have 0 files, got %d", graph.FileCount())
	}
}

func TestAddFile(t *testing.T) {
	graph := NewGraph()
	module := createEmptyModule()

	err := graph.AddFile("/project/a.psx", module)
	if err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	if !graph.HasFile("/project/a.psx") {
		t.Error("file not found in graph after AddFile()")
	}

	if graph.FileCount() != 1 {
		t.Errorf("expected 1 file, got %d", graph.FileCount())
	}
}

func TestAddDuplicateFile(t *testing.T) {
	graph := NewGraph()
	module := createEmptyModule()

	graph.AddFile("/project/a.psx", module)
	err := graph.AddFile("/project/a.psx", module)

	if err == nil {
		t.Error("AddFile() should return error for duplicate file")
	}
}

func TestAddDependency(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)

	err := graph.AddDependency("/project/b.psx", "/project/a.psx")
	if err != nil {
		t.Fatalf("AddDependency() error = %v", err)
	}

	deps := graph.GetDependencies("/project/b.psx")
	if len(deps) != 1 || deps[0] != "/project/a.psx" {
		t.Errorf("expected dependency on a.psx, got %v", deps)
	}
}

func TestAddDependencyDuplicate(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)

	graph.AddDependency("/project/b.psx", "/project/a.psx")
	err := graph.AddDependency("/project/b.psx", "/project/a.psx")

	// Duplicate dependency should not error
	if err != nil {
		t.Errorf("duplicate AddDependency() should not error, got %v", err)
	}

	deps := graph.GetDependencies("/project/b.psx")
	if len(deps) != 1 {
		t.Errorf("duplicate dependency should not be added twice, got %d deps", len(deps))
	}
}

func TestAddDependencyToNonexistentFile(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)

	err := graph.AddDependency("/project/nonexistent.psx", "/project/a.psx")
	if err == nil {
		t.Error("AddDependency() should error when source file doesn't exist")
	}

	err = graph.AddDependency("/project/a.psx", "/project/nonexistent.psx")
	if err == nil {
		t.Error("AddDependency() should error when target file doesn't exist")
	}
}

func TestGetDependents(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()
	moduleC := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddFile("/project/c.psx", moduleC)

	// b and c both depend on a
	graph.AddDependency("/project/b.psx", "/project/a.psx")
	graph.AddDependency("/project/c.psx", "/project/a.psx")

	dependents := graph.GetDependents("/project/a.psx")
	if len(dependents) != 2 {
		t.Fatalf("expected 2 dependents, got %d", len(dependents))
	}

	// Check both b and c are in dependents
	hasB := false
	hasC := false
	for _, dep := range dependents {
		if dep == "/project/b.psx" {
			hasB = true
		}
		if dep == "/project/c.psx" {
			hasC = true
		}
	}

	if !hasB || !hasC {
		t.Errorf("expected both b.psx and c.psx as dependents, got %v", dependents)
	}
}

func TestClear(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.Clear()

	if graph.FileCount() != 0 {
		t.Errorf("after Clear(), expected 0 files, got %d", graph.FileCount())
	}

	if graph.HasFile("/project/a.psx") {
		t.Error("file should not exist after Clear()")
	}
}

// === Topological Sort Tests ===

func TestGetCompilationOrder_Empty(t *testing.T) {
	graph := NewGraph()

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	if len(order) != 0 {
		t.Errorf("empty graph should have empty order, got %v", order)
	}
}

func TestGetCompilationOrder_SingleFile(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	if len(order) != 1 || order[0] != "/project/a.psx" {
		t.Errorf("expected [a.psx], got %v", order)
	}
}

func TestGetCompilationOrder_NoDependencies(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()
	moduleC := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddFile("/project/c.psx", moduleC)

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	if len(order) != 3 {
		t.Errorf("expected 3 files, got %d", len(order))
	}
}

func TestGetCompilationOrder_LinearChain(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()
	moduleC := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddFile("/project/c.psx", moduleC)

	// Chain: a ← b ← c (b depends on a, c depends on b)
	graph.AddDependency("/project/b.psx", "/project/a.psx")
	graph.AddDependency("/project/c.psx", "/project/b.psx")

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	// a must come before b, b must come before c
	expectedOrder := []string{"/project/a.psx", "/project/b.psx", "/project/c.psx"}
	if !reflect.DeepEqual(order, expectedOrder) {
		t.Errorf("expected %v, got %v", expectedOrder, order)
	}
}

func TestGetCompilationOrder_Diamond(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()
	moduleC := createEmptyModule()
	moduleD := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddFile("/project/c.psx", moduleC)
	graph.AddFile("/project/d.psx", moduleD)

	// Diamond:
	//     a
	//    / \
	//   b   c
	//    \ /
	//     d
	graph.AddDependency("/project/b.psx", "/project/a.psx")
	graph.AddDependency("/project/c.psx", "/project/a.psx")
	graph.AddDependency("/project/d.psx", "/project/b.psx")
	graph.AddDependency("/project/d.psx", "/project/c.psx")

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	if len(order) != 4 {
		t.Fatalf("expected 4 files, got %d", len(order))
	}

	// Verify dependencies are satisfied
	positions := make(map[string]int)
	for i, file := range order {
		positions[file] = i
	}

	// a must come before b and c
	if positions["/project/a.psx"] > positions["/project/b.psx"] {
		t.Error("a should come before b")
	}
	if positions["/project/a.psx"] > positions["/project/c.psx"] {
		t.Error("a should come before c")
	}

	// b and c must come before d
	if positions["/project/b.psx"] > positions["/project/d.psx"] {
		t.Error("b should come before d")
	}
	if positions["/project/c.psx"] > positions["/project/d.psx"] {
		t.Error("c should come before d")
	}
}

func TestGetCompilationOrder_ComplexDAG(t *testing.T) {
	graph := NewGraph()
	files := []string{"a", "b", "c", "d", "e"}
	for _, file := range files {
		graph.AddFile("/project/"+file+".psx", createEmptyModule())
	}

	// Complex dependency structure
	graph.AddDependency("/project/b.psx", "/project/a.psx")
	graph.AddDependency("/project/c.psx", "/project/a.psx")
	graph.AddDependency("/project/d.psx", "/project/b.psx")
	graph.AddDependency("/project/d.psx", "/project/c.psx")
	graph.AddDependency("/project/e.psx", "/project/d.psx")

	order, err := graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("GetCompilationOrder() error = %v", err)
	}

	if len(order) != 5 {
		t.Fatalf("expected 5 files, got %d", len(order))
	}

	// Verify all dependencies are satisfied
	positions := make(map[string]int)
	for i, file := range order {
		positions[file] = i
	}

	verifyOrder := func(before, after string) {
		if positions[before] > positions[after] {
			t.Errorf("%s should come before %s in order %v", before, after, order)
		}
	}

	verifyOrder("/project/a.psx", "/project/b.psx")
	verifyOrder("/project/a.psx", "/project/c.psx")
	verifyOrder("/project/b.psx", "/project/d.psx")
	verifyOrder("/project/c.psx", "/project/d.psx")
	verifyOrder("/project/d.psx", "/project/e.psx")
}

// === Cycle Detection Tests ===

func TestDetectCycles_NoCycle(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddDependency("/project/b.psx", "/project/a.psx")

	cycles, err := graph.DetectCycles()
	if err != nil {
		t.Error("DetectCycles() should not error when no cycles exist")
	}

	if cycles != nil && len(cycles) > 0 {
		t.Errorf("expected no cycles, got %d", len(cycles))
	}
}

func TestDetectCycles_SimpleCircle(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)

	// Create cycle: a → b → a
	graph.AddDependency("/project/a.psx", "/project/b.psx")
	graph.AddDependency("/project/b.psx", "/project/a.psx")

	cycles, err := graph.DetectCycles()
	if err == nil {
		t.Fatal("DetectCycles() should return error when cycles exist")
	}

	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle")
	}

	// Verify cycle contains both files
	cycle := cycles[0]
	if len(cycle) < 2 {
		t.Errorf("cycle too short: %v", cycle)
	}
}

func TestDetectCycles_LongCycle(t *testing.T) {
	graph := NewGraph()
	files := []string{"a", "b", "c", "d"}
	for _, file := range files {
		graph.AddFile("/project/"+file+".psx", createEmptyModule())
	}

	// Create cycle: a → b → c → d → a
	graph.AddDependency("/project/a.psx", "/project/b.psx")
	graph.AddDependency("/project/b.psx", "/project/c.psx")
	graph.AddDependency("/project/c.psx", "/project/d.psx")
	graph.AddDependency("/project/d.psx", "/project/a.psx")

	cycles, err := graph.DetectCycles()
	if err == nil {
		t.Fatal("DetectCycles() should return error for cycle")
	}

	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle")
	}

	cycle := cycles[0]
	if len(cycle) < 4 {
		t.Errorf("expected cycle of length 4+, got %d: %v", len(cycle), cycle)
	}
}

func TestDetectCycles_SelfImport(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddDependency("/project/a.psx", "/project/a.psx")

	cycles, err := graph.DetectCycles()
	if err == nil {
		t.Fatal("DetectCycles() should return error for self-import")
	}

	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle")
	}

	cycle := cycles[0]
	if len(cycle) < 2 {
		t.Errorf("self-import cycle should have length 2, got %d: %v", len(cycle), cycle)
	}
}

func TestGetCompilationOrder_WithCycle(t *testing.T) {
	graph := NewGraph()
	moduleA := createEmptyModule()
	moduleB := createEmptyModule()

	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)

	// Create cycle
	graph.AddDependency("/project/a.psx", "/project/b.psx")
	graph.AddDependency("/project/b.psx", "/project/a.psx")

	_, err := graph.GetCompilationOrder()
	if err == nil {
		t.Fatal("GetCompilationOrder() should error on cycle")
	}

	cycleErr, ok := err.(*CycleError)
	if !ok {
		t.Fatalf("expected CycleError, got %T", err)
	}

	if len(cycleErr.Cycles) == 0 {
		t.Error("CycleError should contain cycle information")
	}
}

// === Import Extraction Tests ===

func TestExtractImports_NoImports(t *testing.T) {
	module := createEmptyModule()
	resolver := newMockResolver(map[string]string{})

	imports, err := ExtractImports(module, "/project/main.psx", resolver)
	if err != nil {
		t.Fatalf("ExtractImports() error = %v", err)
	}

	if len(imports) != 0 {
		t.Errorf("expected 0 imports, got %d", len(imports))
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
			name: "single name",
			input: &ast.DottedName{
				Names: []*ast.Name{
					{Token: lexer.Token{Lexeme: "utils"}},
				},
			},
			expected: "utils",
		},
		{
			name: "dotted name",
			input: &ast.DottedName{
				Names: []*ast.Name{
					{Token: lexer.Token{Lexeme: "os"}},
					{Token: lexer.Token{Lexeme: "path"}},
				},
			},
			expected: "os.path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDottedNameToPath(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// === Error Tests ===

func TestCycleError_Error(t *testing.T) {
	cycles := [][]string{
		{"/project/a.psx", "/project/b.psx", "/project/a.psx"},
	}

	err := NewCycleError(cycles)
	errMsg := err.Error()

	if !strings.Contains(errMsg, "circular dependencies detected") {
		t.Error("error message should mention circular dependencies")
	}

	if !strings.Contains(errMsg, "/project/a.psx") {
		t.Error("error message should contain file paths")
	}

	if !strings.Contains(errMsg, "imports") {
		t.Error("error message should show import arrows")
	}
}
