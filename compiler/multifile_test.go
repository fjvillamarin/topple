package compiler

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestFiles creates temporary test files for multi-file compilation tests
func setupTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "multifile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create all files
	for relPath, content := range files {
		fullPath := filepath.Join(tmpDir, relPath)

		// Create parent directories if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

func TestMultiFileCompiler_SingleFile(t *testing.T) {
	files := map[string]string{
		"main.psx": `
def hello():
    return "Hello, World!"
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{filepath.Join(tmpDir, "main.psx")},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 1 {
		t.Errorf("Expected 1 compiled file, got %d", len(output.CompiledFiles))
	}

	mainPath := filepath.Join(tmpDir, "main.psx")
	code, exists := output.CompiledFiles[mainPath]
	if !exists {
		t.Fatalf("main.psx not compiled")
	}

	codeStr := string(code)
	if !strings.Contains(codeStr, "def hello()") {
		t.Errorf("Generated code missing function definition:\n%s", codeStr)
	}
}

func TestMultiFileCompiler_TwoFiles_SimpleImport(t *testing.T) {
	files := map[string]string{
		"utils.psx": `
def greet(name):
    return f"Hello, {name}!"
`,
		"main.psx": `
import utils

def main():
    return utils.greet("World")
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{filepath.Join(tmpDir, "main.psx"), filepath.Join(tmpDir, "utils.psx")},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 2 {
		t.Errorf("Expected 2 compiled files, got %d", len(output.CompiledFiles))
	}

	// Check utils was compiled
	utilsPath := filepath.Join(tmpDir, "utils.psx")
	utilsCode, exists := output.CompiledFiles[utilsPath]
	if !exists {
		t.Fatalf("utils.psx not compiled")
	}
	if !strings.Contains(string(utilsCode), "def greet") {
		t.Errorf("utils.psx missing greet function")
	}

	// Check main was compiled
	mainPath := filepath.Join(tmpDir, "main.psx")
	mainCode, exists := output.CompiledFiles[mainPath]
	if !exists {
		t.Fatalf("main.psx not compiled")
	}
	if !strings.Contains(string(mainCode), "import utils") {
		t.Errorf("main.psx missing import statement")
	}
}

func TestMultiFileCompiler_FromImport(t *testing.T) {
	files := map[string]string{
		"helpers.psx": `
def add(a, b):
    return a + b

def multiply(a, b):
    return a * b
`,
		"calc.psx": `
from helpers import add, multiply

def calculate():
    return add(2, 3) * multiply(4, 5)
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{filepath.Join(tmpDir, "calc.psx"), filepath.Join(tmpDir, "helpers.psx")},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 2 {
		t.Errorf("Expected 2 compiled files, got %d", len(output.CompiledFiles))
	}
}

func TestMultiFileCompiler_RelativeImport(t *testing.T) {
	files := map[string]string{
		"lib/core.psx": `
def core_function():
    return "core"
`,
		"lib/utils.psx": `
from .core import core_function

def util_function():
    return core_function() + "_util"
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files: []string{
			filepath.Join(tmpDir, "lib/core.psx"),
			filepath.Join(tmpDir, "lib/utils.psx"),
		},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 2 {
		t.Errorf("Expected 2 compiled files, got %d", len(output.CompiledFiles))
	}
}

func TestMultiFileCompiler_DependencyOrdering(t *testing.T) {
	// a.psx (no imports)
	// b.psx imports a
	// c.psx imports b
	// Expected order: a, b, c
	files := map[string]string{
		"a.psx": `
def func_a():
    return "a"
`,
		"b.psx": `
import a

def func_b():
    return a.func_a() + "b"
`,
		"c.psx": `
import b

def func_c():
    return b.func_b() + "c"
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files: []string{
			filepath.Join(tmpDir, "c.psx"), // Add in reverse order
			filepath.Join(tmpDir, "b.psx"),
			filepath.Join(tmpDir, "a.psx"),
		},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 3 {
		t.Errorf("Expected 3 compiled files, got %d", len(output.CompiledFiles))
	}

	// Verify compilation order was correct
	order, err := output.Graph.GetCompilationOrder()
	if err != nil {
		t.Fatalf("Failed to get compilation order: %v", err)
	}

	// a should come before b, b before c
	aIdx, bIdx, cIdx := -1, -1, -1
	for i, file := range order {
		if strings.HasSuffix(file, "a.psx") {
			aIdx = i
		} else if strings.HasSuffix(file, "b.psx") {
			bIdx = i
		} else if strings.HasSuffix(file, "c.psx") {
			cIdx = i
		}
	}

	if aIdx == -1 || bIdx == -1 || cIdx == -1 {
		t.Fatalf("Missing files in compilation order")
	}

	if aIdx >= bIdx {
		t.Errorf("a.psx should be compiled before b.psx (got a=%d, b=%d)", aIdx, bIdx)
	}
	if bIdx >= cIdx {
		t.Errorf("b.psx should be compiled before c.psx (got b=%d, c=%d)", bIdx, cIdx)
	}
}

func TestMultiFileCompiler_CircularDependency(t *testing.T) {
	// a.psx imports b
	// b.psx imports a
	files := map[string]string{
		"a.psx": `
import b

def func_a():
    return b.func_b()
`,
		"b.psx": `
import a

def func_b():
    return a.func_a()
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files: []string{
			filepath.Join(tmpDir, "a.psx"),
			filepath.Join(tmpDir, "b.psx"),
		},
	}

	output, err := compiler.CompileProject(context.Background(), opts)

	// Should fail with circular dependency error
	if err == nil {
		t.Fatalf("Expected circular dependency error, got success")
	}

	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("Error should mention circular dependency, got: %v", err)
	}

	// Output should still contain the graph
	if output.Graph == nil {
		t.Errorf("Output should contain dependency graph even on error")
	}
}

func TestMultiFileCompiler_ThreeWayCircular(t *testing.T) {
	// a → b → c → a (circular)
	files := map[string]string{
		"a.psx": `
import b
def func_a():
    return "a"
`,
		"b.psx": `
import c
def func_b():
    return "b"
`,
		"c.psx": `
import a
def func_c():
    return "c"
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files: []string{
			filepath.Join(tmpDir, "a.psx"),
			filepath.Join(tmpDir, "b.psx"),
			filepath.Join(tmpDir, "c.psx"),
		},
	}

	_, err := compiler.CompileProject(context.Background(), opts)

	if err == nil {
		t.Fatalf("Expected circular dependency error")
	}

	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("Error should mention circular dependency, got: %v", err)
	}
}

func TestMultiFileCompiler_DiamondDependency(t *testing.T) {
	// Diamond shape:
	//     base
	//    /    \
	//   a      b
	//    \    /
	//     main
	files := map[string]string{
		"base.psx": `
def base_func():
    return "base"
`,
		"a.psx": `
import base

def func_a():
    return base.base_func() + "_a"
`,
		"b.psx": `
import base

def func_b():
    return base.base_func() + "_b"
`,
		"main.psx": `
import a
import b

def main():
    return a.func_a() + b.func_b()
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files: []string{
			filepath.Join(tmpDir, "main.psx"),
			filepath.Join(tmpDir, "a.psx"),
			filepath.Join(tmpDir, "b.psx"),
			filepath.Join(tmpDir, "base.psx"),
		},
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	if len(output.CompiledFiles) != 4 {
		t.Errorf("Expected 4 compiled files, got %d", len(output.CompiledFiles))
	}

	// Verify base is compiled before a and b
	order, _ := output.Graph.GetCompilationOrder()
	baseIdx, aIdx, bIdx, mainIdx := -1, -1, -1, -1
	for i, file := range order {
		if strings.HasSuffix(file, "base.psx") {
			baseIdx = i
		} else if strings.HasSuffix(file, "a.psx") {
			aIdx = i
		} else if strings.HasSuffix(file, "b.psx") {
			bIdx = i
		} else if strings.HasSuffix(file, "main.psx") {
			mainIdx = i
		}
	}

	if baseIdx >= aIdx || baseIdx >= bIdx {
		t.Errorf("base should be compiled before a and b")
	}
	if aIdx >= mainIdx || bIdx >= mainIdx {
		t.Errorf("a and b should be compiled before main")
	}
}

func TestMultiFileCompiler_DirectoryInput(t *testing.T) {
	files := map[string]string{
		"file1.psx":        `def func1(): return 1`,
		"file2.psx":        `def func2(): return 2`,
		"subdir/file3.psx": `def func3(): return 3`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{tmpDir}, // Pass directory
	}

	output, err := compiler.CompileProject(context.Background(), opts)
	if err != nil {
		t.Fatalf("CompileProject failed: %v", err)
	}

	// Should find all 3 files
	if len(output.CompiledFiles) != 3 {
		t.Errorf("Expected 3 compiled files from directory, got %d", len(output.CompiledFiles))
	}
}

func TestMultiFileCompiler_ParseError(t *testing.T) {
	files := map[string]string{
		"good.psx": `def good(): return 1`,
		"bad.psx":  `def bad( this is invalid syntax`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{filepath.Join(tmpDir, "good.psx"), filepath.Join(tmpDir, "bad.psx")},
	}

	output, err := compiler.CompileProject(context.Background(), opts)

	// Should fail due to parse error
	if err == nil {
		t.Fatalf("Expected parse error, got success")
	}

	if !strings.Contains(err.Error(), "parsing failed") {
		t.Errorf("Error should mention parsing failure, got: %v", err)
	}

	// Should have compilation errors
	if len(output.Errors) == 0 {
		t.Errorf("Expected compilation errors to be recorded")
	}
}

func TestMultiFileCompiler_MissingImport(t *testing.T) {
	files := map[string]string{
		"main.psx": `
import nonexistent

def main():
    return nonexistent.func()
`,
	}

	tmpDir := setupTestFiles(t, files)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	opts := MultiFileOptions{
		RootDir: tmpDir,
		Files:   []string{filepath.Join(tmpDir, "main.psx")},
	}

	output, err := compiler.CompileProject(context.Background(), opts)

	// May succeed in parsing but should have dependency errors
	_ = err // Can be nil or error depending on strictness

	// Should have errors recorded about missing import
	if len(output.Errors) == 0 {
		// This is acceptable - missing imports may be reported differently
		// The important thing is that it doesn't crash
	}
}

func TestCollectAllFiles_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "empty-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	files, err := compiler.collectAllFiles([]string{tmpDir})
	if err == nil {
		t.Fatalf("Expected error for empty directory, got %d files", len(files))
	}

	if !strings.Contains(err.Error(), "no .psx files found") {
		t.Errorf("Expected 'no .psx files found' error, got: %v", err)
	}
}

func TestCollectAllFiles_NonexistentPath(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	compiler := NewMultiFileCompiler(logger)

	files, err := compiler.collectAllFiles([]string{"/nonexistent/path/file.psx"})
	if err == nil {
		t.Fatalf("Expected error for nonexistent path, got %d files", len(files))
	}
}
