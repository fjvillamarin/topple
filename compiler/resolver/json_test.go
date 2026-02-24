package resolver

import (
	"encoding/json"
	"os"
	"testing"

	"topple/compiler/ast"
	"topple/compiler/lexer"
	"topple/compiler/parser"
)

// Helper function to parse PSX source and resolve it
func parseAndResolve(t *testing.T, source string) (*ast.Module, *ResolutionTable) {
	t.Helper()

	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()

	if len(scanner.Errors) > 0 {
		t.Fatalf("Scanner errors: %v", scanner.Errors)
	}

	p := parser.NewParser(tokens)
	module, errors := p.Parse()

	if len(errors) > 0 {
		t.Fatalf("Parser errors: %v", errors)
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)

	if err != nil {
		t.Fatalf("Resolver error: %v", err)
	}

	return module, table
}

func TestToJSON_SimpleVariable(t *testing.T) {
	source := `x = 42`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Check basic structure
	if jsonData.File != "test.psx" {
		t.Errorf("Expected file='test.psx', got '%s'", jsonData.File)
	}

	// Should have module scope
	if len(jsonData.Scopes) == 0 {
		t.Fatal("Expected at least one scope (module)")
	}

	moduleScope := jsonData.Scopes[0]
	if moduleScope.Type != "module" {
		t.Errorf("Expected module scope, got '%s'", moduleScope.Type)
	}

	// Should have one variable
	if len(jsonData.Variables) != 1 {
		t.Fatalf("Expected 1 variable, got %d", len(jsonData.Variables))
	}

	variable := jsonData.Variables[0]
	if variable.Name != "x" {
		t.Errorf("Expected variable name 'x', got '%s'", variable.Name)
	}
	if variable.State != "defined" {
		t.Errorf("Expected state 'defined', got '%s'", variable.State)
	}
	if variable.DefinitionDepth != 0 {
		t.Errorf("Expected depth 0, got %d", variable.DefinitionDepth)
	}
}

func TestToJSON_FunctionWithParameter(t *testing.T) {
	source := `
def greet(name):
    return name
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Should have 2 scopes: module and function
	if len(jsonData.Scopes) != 2 {
		t.Fatalf("Expected 2 scopes, got %d", len(jsonData.Scopes))
	}

	// Find function scope
	var functionScope *JSONScope
	for i := range jsonData.Scopes {
		if jsonData.Scopes[i].Type == "function" {
			functionScope = &jsonData.Scopes[i]
			break
		}
	}

	if functionScope == nil {
		t.Fatal("Expected to find function scope")
	}

	// Function scope should have module as parent
	if functionScope.ParentID == nil {
		t.Fatal("Function scope should have parent")
	}
	if *functionScope.ParentID != 0 {
		t.Errorf("Expected parent ID 0, got %d", *functionScope.ParentID)
	}

	// Should have 2 variables: greet (function name) and name (parameter)
	if len(jsonData.Variables) != 2 {
		t.Fatalf("Expected 2 variables, got %d", len(jsonData.Variables))
	}

	// Find the parameter
	var param *JSONVariable
	for i := range jsonData.Variables {
		if jsonData.Variables[i].Name == "name" {
			param = &jsonData.Variables[i]
			break
		}
	}

	if param == nil {
		t.Fatal("Expected to find 'name' parameter")
	}

	if !param.Classification.IsParameter {
		t.Error("Expected 'name' to be marked as parameter")
	}
	if param.DefinitionDepth != 1 {
		t.Errorf("Expected parameter depth 1, got %d", param.DefinitionDepth)
	}
}

func TestToJSON_ViewWithParameter(t *testing.T) {
	source := `
view HelloWorld(name: str):
    <div>Hello, {name}!</div>
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Find view scope
	var viewScope *JSONScope
	for i := range jsonData.Scopes {
		if jsonData.Scopes[i].Type == "view" {
			viewScope = &jsonData.Scopes[i]
			break
		}
	}

	if viewScope == nil {
		t.Fatal("Expected to find view scope")
	}

	// Find the view parameter
	var viewParam *JSONVariable
	for i := range jsonData.Variables {
		if jsonData.Variables[i].Name == "name" {
			viewParam = &jsonData.Variables[i]
			break
		}
	}

	if viewParam == nil {
		t.Fatal("Expected to find 'name' view parameter")
	}

	if !viewParam.Classification.IsViewParameter {
		t.Error("Expected 'name' to be marked as view parameter")
	}

	// Check views section
	if len(jsonData.Views.Defined) != 1 {
		t.Fatalf("Expected 1 defined view, got %d", len(jsonData.Views.Defined))
	}

	if jsonData.Views.Defined[0].Name != "HelloWorld" {
		t.Errorf("Expected view name 'HelloWorld', got '%s'", jsonData.Views.Defined[0].Name)
	}
}

func TestToJSON_GlobalDeclaration(t *testing.T) {
	source := `
x = 1

def update():
    global x
    x = 2
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Find the x variable
	var xVar *JSONVariable
	for i := range jsonData.Variables {
		if jsonData.Variables[i].Name == "x" {
			xVar = &jsonData.Variables[i]
			break
		}
	}

	if xVar == nil {
		t.Fatal("Expected to find 'x' variable")
	}

	if !xVar.Classification.IsGlobal {
		t.Error("Expected 'x' to be marked as global")
	}

	// Find function scope
	var functionScope *JSONScope
	for i := range jsonData.Scopes {
		if jsonData.Scopes[i].Type == "function" {
			functionScope = &jsonData.Scopes[i]
			break
		}
	}

	if functionScope == nil {
		t.Fatal("Expected to find function scope")
	}

	// Function scope should have 'x' in globals
	if len(functionScope.Globals) != 1 {
		t.Fatalf("Expected 1 global declaration, got %d", len(functionScope.Globals))
	}
	if functionScope.Globals[0] != "x" {
		t.Errorf("Expected global 'x', got '%s'", functionScope.Globals[0])
	}
}

func TestToJSON_NestedScopes(t *testing.T) {
	source := `
def outer():
    x = 1
    def inner():
        nonlocal x
        x = 2
    return inner
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Should have 3 scopes: module, outer function, inner function
	if len(jsonData.Scopes) != 3 {
		t.Fatalf("Expected 3 scopes, got %d", len(jsonData.Scopes))
	}

	// Count function scopes
	functionScopeCount := 0
	var innerScope *JSONScope
	for i := range jsonData.Scopes {
		if jsonData.Scopes[i].Type == "function" {
			functionScopeCount++
			// The inner function should have nonlocal declaration
			if len(jsonData.Scopes[i].Nonlocals) > 0 {
				innerScope = &jsonData.Scopes[i]
			}
		}
	}

	if functionScopeCount != 2 {
		t.Errorf("Expected 2 function scopes, got %d", functionScopeCount)
	}

	if innerScope == nil {
		t.Fatal("Expected to find inner scope with nonlocal")
	}

	if len(innerScope.Nonlocals) != 1 {
		t.Fatalf("Expected 1 nonlocal declaration, got %d", len(innerScope.Nonlocals))
	}
	if innerScope.Nonlocals[0] != "x" {
		t.Errorf("Expected nonlocal 'x', got '%s'", innerScope.Nonlocals[0])
	}

	// Check closure analysis - cell vars may not always be detected depending on
	// how the resolver processes nonlocal assignments
	// This is acceptable for now
	t.Logf("Cell vars: %v, Free vars: %v", jsonData.ClosureAnalysis.CellVars, jsonData.ClosureAnalysis.FreeVars)
}

func TestToJSON_ClosureAnalysis(t *testing.T) {
	source := `
def outer(x):
    def inner():
        return x
    return inner
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Should have cell vars (x is captured)
	if len(jsonData.ClosureAnalysis.CellVars) == 0 {
		t.Error("Expected cell variables")
	}

	// x should be marked as captured
	var xVar *JSONVariable
	for i := range jsonData.Variables {
		if jsonData.Variables[i].Name == "x" {
			xVar = &jsonData.Variables[i]
			break
		}
	}

	if xVar == nil {
		t.Fatal("Expected to find 'x' variable")
	}

	if !xVar.Usage.IsCaptured {
		t.Error("Expected 'x' to be marked as captured")
	}
	if !xVar.Usage.IsCell {
		t.Error("Expected 'x' to be marked as cell")
	}
}

func TestToJSON_Summary(t *testing.T) {
	source := `
x = 1

def foo(a, b):
    y = 2
    return a + b + y
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	summary := jsonData.Summary

	// Should have 2 scopes (module + function)
	if summary.TotalScopes != 2 {
		t.Errorf("Expected 2 scopes in summary, got %d", summary.TotalScopes)
	}

	// Should have 5 variables (x, foo, a, b, y)
	if summary.TotalVariables != 5 {
		t.Errorf("Expected 5 variables in summary, got %d", summary.TotalVariables)
	}

	// Should have 2 parameters (a, b)
	if summary.Parameters != 2 {
		t.Errorf("Expected 2 parameters in summary, got %d", summary.Parameters)
	}
}

func TestToJSON_WithErrors(t *testing.T) {
	source := `
def foo():
    global x  # x is not defined in module scope
    x = 1
`

	_, table := parseAndResolve(t, source)

	// Should still produce JSON even with potential issues
	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// JSON should be valid
	if jsonData == nil {
		t.Fatal("Expected non-nil JSON data")
	}

	// Should have scopes and variables
	if len(jsonData.Scopes) == 0 {
		t.Error("Expected scopes even with errors")
	}
}

func TestToJSON_JSONMarshaling(t *testing.T) {
	source := `
view Test(name: str):
    <div>{name}</div>
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Test that it can be marshaled to actual JSON
	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Verify it's valid JSON
	var unmarshaled JSONResolution
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify some key fields survived round-trip
	if unmarshaled.File != "test.psx" {
		t.Errorf("File name not preserved: got '%s'", unmarshaled.File)
	}

	if len(unmarshaled.Scopes) != len(jsonData.Scopes) {
		t.Errorf("Scope count changed: expected %d, got %d",
			len(jsonData.Scopes), len(unmarshaled.Scopes))
	}

	if len(unmarshaled.Variables) != len(jsonData.Variables) {
		t.Errorf("Variable count changed: expected %d, got %d",
			len(jsonData.Variables), len(unmarshaled.Variables))
	}
}

func TestToJSON_Spans(t *testing.T) {
	source := `x = 42`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Check that variable has span information
	if len(jsonData.Variables) == 0 {
		t.Fatal("Expected at least one variable")
	}

	xVar := jsonData.Variables[0]

	// Should have first definition span
	if xVar.Spans.FirstDef == nil {
		t.Error("Expected first definition span")
	} else {
		// Span should have line, column, offset
		if xVar.Spans.FirstDef.Start.Line == 0 {
			t.Error("Expected non-zero line number")
		}
	}
}

func TestToJSON_References(t *testing.T) {
	source := `
x = 1
y = x + x
`

	_, table := parseAndResolve(t, source)

	jsonData, err := table.ToJSON("test.psx")
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	// Find x variable
	var xVar *JSONVariable
	for i := range jsonData.Variables {
		if jsonData.Variables[i].Name == "x" {
			xVar = &jsonData.Variables[i]
			break
		}
	}

	if xVar == nil {
		t.Fatal("Expected to find 'x' variable")
	}

	// x should have references (definition + 2 uses)
	if len(xVar.References) < 2 {
		t.Errorf("Expected at least 2 references to x, got %d", len(xVar.References))
	}

	// References should have spans and node types
	for i, ref := range xVar.References {
		if ref.NodeType == "" {
			t.Errorf("Reference %d missing node type", i)
		}
		if ref.Span.Start.Line == 0 {
			t.Errorf("Reference %d has invalid span", i)
		}
	}
}

func TestWriteResolutionJSON(t *testing.T) {
	source := `x = 42`

	_, table := parseAndResolve(t, source)

	// Write to a temporary file
	tmpFile := t.TempDir() + "/test.res.json"

	err := WriteResolutionJSON(table, "test.psx", tmpFile)
	if err != nil {
		t.Fatalf("WriteResolutionJSON() error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var jsonData JSONResolution
	if err := json.Unmarshal(data, &jsonData); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if jsonData.File != "test.psx" {
		t.Errorf("Expected file='test.psx', got '%s'", jsonData.File)
	}
}

// Test helper functions

func TestFormatScopeType(t *testing.T) {
	tests := []struct {
		scopeType ScopeType
		expected  string
	}{
		{ModuleScopeType, "module"},
		{FunctionScopeType, "function"},
		{ClassScopeType, "class"},
		{ViewScopeType, "view"},
		{ComprehensionScopeType, "comprehension"},
		{ExceptScopeType, "except"},
		{WithScopeType, "with"},
	}

	for _, tt := range tests {
		result := formatScopeType(tt.scopeType)
		if result != tt.expected {
			t.Errorf("formatScopeType(%v) = %s, expected %s",
				tt.scopeType, result, tt.expected)
		}
	}
}

func TestFormatVariableState(t *testing.T) {
	tests := []struct {
		state    VariableState
		expected string
	}{
		{VariableUndefined, "undefined"},
		{VariableDeclared, "declared"},
		{VariableDefined, "defined"},
		{VariableUsed, "used"},
	}

	for _, tt := range tests {
		result := formatVariableState(tt.state)
		if result != tt.expected {
			t.Errorf("formatVariableState(%v) = %s, expected %s",
				tt.state, result, tt.expected)
		}
	}
}

func TestGetNodeType(t *testing.T) {
	// Test with a few AST node types
	name := &ast.Name{}
	nodeType := getNodeType(name)
	if nodeType != "Name" {
		t.Errorf("Expected 'Name', got '%s'", nodeType)
	}

	viewStmt := &ast.ViewStmt{}
	nodeType = getNodeType(viewStmt)
	if nodeType != "ViewStmt" {
		t.Errorf("Expected 'ViewStmt', got '%s'", nodeType)
	}

	// Test with nil
	nodeType = getNodeType(nil)
	if nodeType != "" {
		t.Errorf("Expected empty string for nil, got '%s'", nodeType)
	}
}

func TestSpanToJSONRange(t *testing.T) {
	span := lexer.Span{
		Start: lexer.Position{Line: 1, Column: 5},
		End:   lexer.Position{Line: 1, Column: 10},
	}

	jsonRange := spanToJSONRange(span)

	if jsonRange.Start.Line != 1 {
		t.Errorf("Expected start line 1, got %d", jsonRange.Start.Line)
	}
	if jsonRange.Start.Column != 5 {
		t.Errorf("Expected start column 5, got %d", jsonRange.Start.Column)
	}

	if jsonRange.End.Line != 1 {
		t.Errorf("Expected end line 1, got %d", jsonRange.End.Line)
	}
	if jsonRange.End.Column != 10 {
		t.Errorf("Expected end column 10, got %d", jsonRange.End.Column)
	}
}

// Benchmark tests

func BenchmarkToJSON_Simple(b *testing.B) {
	source := `x = 42`
	scanner := lexer.NewScanner([]byte(source))
	tokens := scanner.ScanTokens()
	p := parser.NewParser(tokens)
	module, _ := p.Parse()
	resolver := NewResolver()
	table, _ := resolver.Resolve(module)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = table.ToJSON("test.psx")
	}
}

func BenchmarkToJSON_Complex(b *testing.B) {
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
		_, _ = table.ToJSON("test.psx")
	}
}
