package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"topple/compiler/lexer"
	"topple/compiler/parser"
)

func TestToText_SimpleVariable(t *testing.T) {
	source := `x = 42`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Check that text contains expected sections
	if !strings.Contains(text, "RESOLUTION OUTPUT: test.psx") {
		t.Error("Expected header with filename")
	}

	if !strings.Contains(text, "SCOPES:") {
		t.Error("Expected SCOPES section")
	}

	if !strings.Contains(text, "VARIABLES:") {
		t.Error("Expected VARIABLES section")
	}

	if !strings.Contains(text, "SUMMARY:") {
		t.Error("Expected SUMMARY section")
	}

	// Check variable appears
	if !strings.Contains(text, "x") {
		t.Error("Expected variable 'x' in output")
	}
}

func TestToText_FunctionWithParameter(t *testing.T) {
	source := `
def greet(name):
    return name
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should mention both function and parameter
	if !strings.Contains(text, "greet") {
		t.Error("Expected function 'greet' in output")
	}

	if !strings.Contains(text, "name") {
		t.Error("Expected parameter 'name' in output")
	}

	// Should show scope hierarchy
	if !strings.Contains(text, "parent:") {
		t.Error("Expected parent scope reference")
	}

	// Should show parameter type
	if !strings.Contains(text, "parameter") {
		t.Error("Expected parameter type classification")
	}
}

func TestToText_ViewWithParameter(t *testing.T) {
	source := `
view HelloWorld(name: str):
    <div>Hello, {name}!</div>
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Check view appears in views section
	if !strings.Contains(text, "VIEW COMPOSITION:") {
		t.Error("Expected VIEW COMPOSITION section")
	}

	if !strings.Contains(text, "HelloWorld") {
		t.Error("Expected view 'HelloWorld' in output")
	}

	// Check view parameter is marked
	if !strings.Contains(text, "view-param") {
		t.Error("Expected view parameter classification")
	}
}

func TestToText_GlobalDeclaration(t *testing.T) {
	source := `
x = 1

def update():
    global x
    x = 2
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should show global declaration in function scope
	if !strings.Contains(text, "Global declarations:") {
		t.Error("Expected global declarations section in scope")
	}

	if !strings.Contains(text, "global") {
		t.Error("Expected 'global' type classification")
	}
}

func TestToText_NestedScopes(t *testing.T) {
	source := `
def outer():
    x = 1
    def inner():
        nonlocal x
        x = 2
    return inner
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should show nested scopes
	if !strings.Contains(text, "Scope #0") {
		t.Error("Expected Scope #0")
	}

	if !strings.Contains(text, "Scope #1") {
		t.Error("Expected Scope #1")
	}

	if !strings.Contains(text, "Scope #2") {
		t.Error("Expected Scope #2")
	}

	// Should show nonlocal declaration
	if !strings.Contains(text, "Nonlocal declarations:") {
		t.Error("Expected nonlocal declarations section")
	}
}

func TestToText_ClosureAnalysis(t *testing.T) {
	source := `
def outer(x):
    def inner():
        return x
    return inner
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should have closure analysis section
	if !strings.Contains(text, "CLOSURE ANALYSIS:") {
		t.Error("Expected CLOSURE ANALYSIS section")
	}

	// Should mention cell or free variables
	if !strings.Contains(text, "Cell Variables:") {
		t.Error("Expected Cell Variables subsection")
	}

	if !strings.Contains(text, "Free Variables:") {
		t.Error("Expected Free Variables subsection")
	}
}

func TestToText_WithErrors(t *testing.T) {
	// 'global' at module level triggers a resolver error
	source := `
global x
x = 1
`

	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()

	resolver := NewResolver()
	table, _ := resolver.Resolve(module)

	if len(table.Errors) == 0 {
		t.Fatal("Expected resolver errors but got none")
	}

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should have diagnostics section with actual errors
	if !strings.Contains(text, "DIAGNOSTICS:") {
		t.Error("Expected DIAGNOSTICS section")
	}

	if !strings.Contains(text, "Errors: ") {
		t.Error("Expected error count in diagnostics section")
	}

	// Should show non-zero error count in summary
	if strings.Contains(text, "Resolution Errors:      0") {
		t.Error("Expected non-zero Resolution Errors count in summary")
	}
}

func TestToText_Summary(t *testing.T) {
	source := `
x = 1

def foo(a, b):
    y = 2
    return a + b + y
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Check summary section exists
	if !strings.Contains(text, "SUMMARY:") {
		t.Error("Expected SUMMARY section")
	}

	// Check key statistics are present
	if !strings.Contains(text, "Total Scopes:") {
		t.Error("Expected Total Scopes in summary")
	}

	if !strings.Contains(text, "Total Variables:") {
		t.Error("Expected Total Variables in summary")
	}

	if !strings.Contains(text, "Parameters:") {
		t.Error("Expected Parameters count in summary")
	}
}

func TestToText_FormattingConsistency(t *testing.T) {
	source := `
view Test(name: str):
    <div>{name}</div>
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Check that section headers use consistent formatting
	lines := strings.Split(text, "\n")

	sectionCount := 0
	for _, line := range lines {
		if strings.Contains(line, "━") {
			sectionCount++
		}
	}

	// Should have multiple sections
	if sectionCount < 5 {
		t.Errorf("Expected at least 5 section separators, got %d", sectionCount)
	}

	// Informational: log lines exceeding 100 chars (not a hard failure, just advisory)
	for i, line := range lines {
		// Skip separator lines and data lines
		if !strings.Contains(line, "━") && !strings.Contains(line, "─") {
			if len(line) > 100 {
				t.Logf("Advisory: line %d exceeds 100 chars: %s", i+1, line[:50])
			}
		}
	}
}

func TestWriteResolutionText(t *testing.T) {
	source := `x = 42`

	_, table := parseAndResolve(t, source)

	// Write to a temporary file
	tmpFile := filepath.Join(t.TempDir(), "test.res")

	err := WriteResolutionText(table, "test.psx", tmpFile)
	if err != nil {
		t.Fatalf("WriteResolutionText() error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read text file: %v", err)
	}

	text := string(data)

	// Basic validation
	if !strings.Contains(text, "test.psx") {
		t.Error("Expected filename in output")
	}

	if !strings.Contains(text, "x") {
		t.Error("Expected variable 'x' in output")
	}

	// Check file ends with newline
	if !strings.HasSuffix(text, "\n") {
		t.Error("Expected file to end with newline")
	}
}

func TestToText_EmptyModule(t *testing.T) {
	source := ``

	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()

	resolver := NewResolver()
	table, _ := resolver.Resolve(module)

	text, err := table.ToText("empty.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Should still have all sections
	if !strings.Contains(text, "SCOPES:") {
		t.Error("Expected SCOPES section even for empty module")
	}

	if !strings.Contains(text, "SUMMARY:") {
		t.Error("Expected SUMMARY section even for empty module")
	}

	// Should show 0 counts appropriately
	if !strings.Contains(text, "Total Scopes:           1") {
		t.Error("Expected 1 scope (module) for empty module")
	}
}

func TestToText_VariableTable(t *testing.T) {
	source := `
x = 1
y = "hello"

def foo(a, b):
    z = a + b
    return z
`

	_, table := parseAndResolve(t, source)

	text, err := table.ToText("test.psx")
	if err != nil {
		t.Fatalf("ToText() error: %v", err)
	}

	// Check variables table header
	if !strings.Contains(text, "NAME") {
		t.Error("Expected NAME column in variables table")
	}

	if !strings.Contains(text, "TYPE") {
		t.Error("Expected TYPE column in variables table")
	}

	if !strings.Contains(text, "STATE") {
		t.Error("Expected STATE column in variables table")
	}

	if !strings.Contains(text, "DEPTH") {
		t.Error("Expected DEPTH column in variables table")
	}

	if !strings.Contains(text, "FLAGS") {
		t.Error("Expected FLAGS column in variables table")
	}

	if !strings.Contains(text, "REFS") {
		t.Error("Expected REFS column in variables table")
	}

	// Check table separator
	if !strings.Contains(text, "─") {
		t.Error("Expected table separator line")
	}
}

// Benchmark tests

func BenchmarkToText_Simple(b *testing.B) {
	source := `x = 42`
	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()
	resolver := NewResolver()
	table, _ := resolver.Resolve(module)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = table.ToText("test.psx")
	}
}

func BenchmarkToText_Complex(b *testing.B) {
	source := `
def outer(x):
    y = 1
    def inner(z):
        nonlocal y
        global g
        return x + y + z + g
    return inner

view Test(name: str):
    <div>{name}</div>
`
	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()
	resolver := NewResolver()
	table, _ := resolver.Resolve(module)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = table.ToText("test.psx")
	}
}
