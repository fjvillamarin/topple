package resolver

import (
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

func TestResolverBasic(t *testing.T) {
	nameNode := &ast.Name{
		Token: lexer.Token{
			Lexeme: "x",
			Type:   lexer.Identifier,
		},
		Span: lexer.Span{},
	}
	// Create a simple module with a name reference
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.ExprStmt{
				Expr: nameNode,
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	// Create resolver and test basic functionality
	resolver := NewResolver()

	// Test that resolver can be created
	if resolver == nil {
		t.Fatal("NewResolver() returned nil")
	}

	// Test that resolution table can be created
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if table == nil {
		t.Fatal("Resolve() returned nil table")
	}

	// Should have resolved the 'x' name
	if len(table.Variables) == 0 {
		t.Error("Expected at least one variable to be resolved")
	}

	if table.Variables[nameNode] == nil {
		t.Error("Expected variable 'x' to be resolved")
	}

	if table.Variables[nameNode].State != VariableUndefined {
		t.Errorf("Expected variable state %v, got %v", VariableUndefined, table.Variables[nameNode].State)
	}
}

func TestVariableDefinitionAndUsage(t *testing.T) {
	// Create nodes for: x = 42; print(x)
	assignTarget := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	usageRef := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{
			// x = 42
			&ast.AssignStmt{
				Targets: []ast.Expr{assignTarget},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeNumber,
					Value: 42,
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			// print(x)
			&ast.ExprStmt{
				Expr: &ast.Call{
					Callee: &ast.Name{
						Token: lexer.Token{Lexeme: "print", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Arguments: []*ast.Argument{
						{Value: usageRef, Span: lexer.Span{}},
					},
					Span: lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check assignment target
	assignVar := table.Variables[assignTarget]
	if assignVar == nil {
		t.Fatal("Assignment target not resolved")
	}
	if assignVar.State != VariableDefined {
		t.Errorf("Expected assignment target state %v, got %v", VariableDefined, assignVar.State)
	}

	// Check usage reference
	usageVar := table.Variables[usageRef]
	if usageVar == nil {
		t.Fatal("Usage reference not resolved")
	}
	if usageVar != assignVar {
		t.Error("Usage reference should point to same variable as assignment target")
	}
	if !usageVar.IsUsed {
		t.Error("Variable should be marked as used")
	}
}

func TestGlobalStatement(t *testing.T) {
	// Create nodes for: global x; x = 42
	globalName := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	assignTarget := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	// Function containing global statement
	function := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "func", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{
			Parameters: []*ast.Parameter{},
			Span:       lexer.Span{},
		},
		Body: []ast.Stmt{
			&ast.GlobalStmt{
				Names: []*ast.Name{globalName},
				Span:  lexer.Span{},
			},
			&ast.AssignStmt{
				Targets: []ast.Expr{assignTarget},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeNumber,
					Value: 42,
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{function},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check that assignment target resolves to global variable
	assignVar := table.Variables[assignTarget]
	if assignVar == nil {
		t.Fatal("Assignment target not resolved")
	}
	if !assignVar.IsGlobal {
		t.Error("Variable should be marked as global")
	}
}

func TestFunctionScoping(t *testing.T) {
	// Create nodes for:
	// def func(param):
	//     local_var = param
	//     return local_var

	paramName := &ast.Name{
		Token: lexer.Token{Lexeme: "param", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	localVarAssign := &ast.Name{
		Token: lexer.Token{Lexeme: "local_var", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	paramUsage := &ast.Name{
		Token: lexer.Token{Lexeme: "param", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	localVarUsage := &ast.Name{
		Token: lexer.Token{Lexeme: "local_var", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	function := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "func", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{
			Parameters: []*ast.Parameter{
				{
					Name: paramName,
					Span: lexer.Span{},
				},
			},
			Span: lexer.Span{},
		},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{localVarAssign},
				Value:   paramUsage,
				Span:    lexer.Span{},
			},
			&ast.ReturnStmt{
				Value: localVarUsage,
				Span:  lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{function},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check parameter
	paramVar := table.Variables[paramName]
	if paramVar == nil {
		t.Fatal("Parameter not resolved")
	}
	if !paramVar.IsParameter {
		t.Error("Variable should be marked as parameter")
	}
	if paramVar.State != VariableDefined {
		t.Errorf("Parameter should be defined, got %v", paramVar.State)
	}

	// Check parameter usage
	paramUsageVar := table.Variables[paramUsage]
	if paramUsageVar != paramVar {
		t.Error("Parameter usage should resolve to same variable")
	}

	// Check local variable
	localVar := table.Variables[localVarAssign]
	if localVar == nil {
		t.Fatal("Local variable not resolved")
	}
	if localVar.State != VariableDefined {
		t.Errorf("Local variable should be defined, got %v", localVar.State)
	}

	// Check scope depths (both should be in function scope = 1, since module is 0)
	if table.ScopeDepths[paramUsage] != 1 {
		t.Errorf("Parameter usage scope depth should be 1, got %d", table.ScopeDepths[paramUsage])
	}
	if table.ScopeDepths[localVarUsage] != 1 {
		t.Errorf("Local variable usage scope depth should be 1, got %d", table.ScopeDepths[localVarUsage])
	}
}

func TestViewScoping(t *testing.T) {
	// Create nodes for:
	// view MyView(name: str):
	//     message = f"Hello {name}"
	//     <div>{message}</div>

	viewParamName := &ast.Name{
		Token: lexer.Token{Lexeme: "name", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	messageAssign := &ast.Name{
		Token: lexer.Token{Lexeme: "message", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	nameUsage := &ast.Name{
		Token: lexer.Token{Lexeme: "name", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	view := &ast.ViewStmt{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "MyView", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Params: &ast.ParameterList{
			Parameters: []*ast.Parameter{
				{
					Name: viewParamName,
					Annotation: &ast.Name{
						Token: lexer.Token{Lexeme: "str", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Span: lexer.Span{},
				},
			},
			Span: lexer.Span{},
		},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{messageAssign},
				Value: &ast.FString{
					Parts: []ast.FStringPart{
						&ast.FStringMiddle{Value: "Hello "},
						&ast.FStringReplacementField{Expression: nameUsage},
					},
					Span: lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{view},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check view parameter
	viewParam := table.Variables[viewParamName]
	if viewParam == nil {
		t.Fatal("View parameter not resolved")
	}
	if !viewParam.IsParameter {
		t.Error("Should be marked as parameter")
	}
	if !viewParam.IsViewParameter {
		t.Error("Should be marked as view parameter")
	}

	// Check parameter usage
	nameUsageVar := table.Variables[nameUsage]
	if nameUsageVar != viewParam {
		t.Error("Parameter usage should resolve to same variable")
	}
}

func TestLEGBResolution(t *testing.T) {
	// Create nested function structure to test LEGB:
	// x = "global"
	// def outer():
	//     x = "enclosing"
	//     def inner():
	//         x = "local"
	//         return x
	//     return inner

	globalX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	enclosingX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	localX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	localXUsage := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	innerFunc := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "inner", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}, Span: lexer.Span{}},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{localX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "local",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			&ast.ReturnStmt{
				Value: localXUsage,
				Span:  lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	outerFunc := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "outer", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}, Span: lexer.Span{}},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{enclosingX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "enclosing",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			innerFunc,
			&ast.ReturnStmt{
				Value: &ast.Name{
					Token: lexer.Token{Lexeme: "inner", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{globalX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "global",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			outerFunc,
		},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check that localXUsage resolves to local variable (scope depth 0)
	localUsageVar := table.Variables[localXUsage]
	localDefVar := table.Variables[localX]

	if localUsageVar != localDefVar {
		t.Error("Local usage should resolve to local definition")
	}

	if table.ScopeDepths[localXUsage] != 2 {
		t.Errorf("Local variable usage should have scope depth 2, got %d", table.ScopeDepths[localXUsage])
	}

	// Verify the three variables are distinct
	globalVar := table.Variables[globalX]
	enclosingVar := table.Variables[enclosingX]

	if globalVar == enclosingVar || globalVar == localDefVar || enclosingVar == localDefVar {
		t.Error("Global, enclosing, and local variables should be distinct")
	}
}

func TestNonlocalStatement(t *testing.T) {
	// Create:
	// def outer():
	//     x = "enclosing"
	//     def inner():
	//         nonlocal x
	//         x = "modified"
	//     return inner

	enclosingX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	nonlocalName := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	nonlocalAssign := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	innerFunc := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "inner", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}, Span: lexer.Span{}},
		Body: []ast.Stmt{
			&ast.NonlocalStmt{
				Names: []*ast.Name{nonlocalName},
				Span:  lexer.Span{},
			},
			&ast.AssignStmt{
				Targets: []ast.Expr{nonlocalAssign},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "modified",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	outerFunc := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "outer", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}, Span: lexer.Span{}},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{enclosingX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "enclosing",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			innerFunc,
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{outerFunc},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check that nonlocal assignment resolves to enclosing variable
	enclosingVar := table.Variables[enclosingX]
	nonlocalAssignVar := table.Variables[nonlocalAssign]

	if nonlocalAssignVar != enclosingVar {
		t.Error("Nonlocal assignment should resolve to enclosing variable")
	}

	if !enclosingVar.IsCaptured {
		t.Error("Enclosing variable should be marked as captured")
	}
}

func TestScopeManagement(t *testing.T) {
	resolver := NewResolver()

	// Test initial state - should have module scope
	if resolver.ScopeChain == nil {
		t.Error("Expected scope chain to be initialized")
	}

	if resolver.ScopeChain.ScopeType != ModuleScopeType {
		t.Errorf("Expected module scope type, got %v", resolver.ScopeChain.ScopeType)
	}

	// Count initial scopes in chain
	scopeCount := 0
	for scope := resolver.ScopeChain; scope != nil; scope = scope.Parent {
		scopeCount++
	}
	if scopeCount != 1 {
		t.Errorf("Expected 1 initial scope, got %d", scopeCount)
	}

	// Test scope creation
	resolver.BeginScope(FunctionScopeType)
	scopeCount = 0
	for scope := resolver.ScopeChain; scope != nil; scope = scope.Parent {
		scopeCount++
	}
	if scopeCount != 2 {
		t.Errorf("Expected 2 scopes after BeginScope, got %d", scopeCount)
	}

	if resolver.ScopeChain.ScopeType != FunctionScopeType {
		t.Errorf("Expected function scope type, got %v", resolver.ScopeChain.ScopeType)
	}

	// Test scope destruction
	resolver.EndScope()
	scopeCount = 0
	for scope := resolver.ScopeChain; scope != nil; scope = scope.Parent {
		scopeCount++
	}
	if scopeCount != 1 {
		t.Errorf("Expected 1 scope after EndScope, got %d", scopeCount)
	}
}

func TestVariableDefinition(t *testing.T) {
	resolver := NewResolver()

	// Test variable definition
	variable := resolver.DefineVariable("test_var", lexer.Span{})

	if variable == nil {
		t.Fatal("DefineVariable returned nil")
	}

	if variable.Name != "test_var" {
		t.Errorf("Expected variable name 'test_var', got '%s'", variable.Name)
	}

	if variable.State != VariableDeclared {
		t.Errorf("Expected variable state %v, got %v", VariableDeclared, variable.State)
	}

	// Test that variable is in current scope via binding
	if binding, exists := resolver.ScopeChain.Bindings["test_var"]; !exists {
		t.Error("Variable binding not found in current scope")
	} else if binding.Variable != variable {
		t.Error("Binding does not point to the correct variable")
	}
}

func TestAssignmentTargetAnalysis(t *testing.T) {
	// Test tuple unpacking: a, b = (1, 2)
	aName := &ast.Name{
		Token: lexer.Token{Lexeme: "a", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}
	bName := &ast.Name{
		Token: lexer.Token{Lexeme: "b", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{
					&ast.TupleExpr{
						Elements: []ast.Expr{aName, bName},
						Span:     lexer.Span{},
					},
				},
				Value: &ast.TupleExpr{
					Elements: []ast.Expr{
						&ast.Literal{Type: ast.LiteralTypeNumber, Value: 1, Span: lexer.Span{}},
						&ast.Literal{Type: ast.LiteralTypeNumber, Value: 2, Span: lexer.Span{}},
					},
					Span: lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Check both variables were defined
	aVar := table.Variables[aName]
	bVar := table.Variables[bName]

	if aVar == nil {
		t.Error("Variable 'a' should be defined")
	}
	if bVar == nil {
		t.Error("Variable 'b' should be defined")
	}

	if aVar != nil && aVar.State != VariableDefined {
		t.Errorf("Variable 'a' should be defined, got %v", aVar.State)
	}
	if bVar != nil && bVar.State != VariableDefined {
		t.Errorf("Variable 'b' should be defined, got %v", bVar.State)
	}
}

func TestClassScopeIsolation(t *testing.T) {
	// Test that class variables don't interfere with nested function lookups
	// x = "global"
	// class MyClass:
	//     x = "class"
	//     def method(self):
	//         def inner():
	//             return x  # Should resolve to global, not class
	//         return inner()

	globalX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	classX := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	innerXUsage := &ast.Name{
		Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
		Span:  lexer.Span{},
	}

	innerFunc := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "inner", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}, Span: lexer.Span{}},
		Body: []ast.Stmt{
			&ast.ReturnStmt{
				Value: innerXUsage,
				Span:  lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	method := &ast.Function{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "method", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Parameters: &ast.ParameterList{
			Parameters: []*ast.Parameter{
				{
					Name: &ast.Name{
						Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Span: lexer.Span{},
				},
			},
			Span: lexer.Span{},
		},
		Body: []ast.Stmt{
			innerFunc,
			&ast.ReturnStmt{
				Value: &ast.Call{
					Callee: &ast.Name{
						Token: lexer.Token{Lexeme: "inner", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Arguments: []*ast.Argument{},
					Span:      lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	class := &ast.Class{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "MyClass", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{classX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "class",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			method,
		},
		Span: lexer.Span{},
	}

	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{globalX},
				Value: &ast.Literal{
					Type:  ast.LiteralTypeString,
					Value: "global",
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
			class,
		},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Inner function usage should resolve to global variable, not class variable
	globalVar := table.Variables[globalX]
	classVar := table.Variables[classX]
	innerUsageVar := table.Variables[innerXUsage]

	if innerUsageVar != globalVar {
		t.Error("Inner function should resolve to global variable, not class variable")
	}

	if innerUsageVar == classVar {
		t.Error("Inner function should NOT resolve to class variable")
	}
}

func TestErrorReporting(t *testing.T) {
	// Test global statement at module level (should be an error)
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.GlobalStmt{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "x", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	resolver := NewResolver()
	table, err := resolver.Resolve(module)

	// Should have error but still return table
	if err == nil {
		t.Error("Expected error for global statement at module level")
	}

	if table == nil {
		t.Error("Should still return resolution table even with errors")
	}

	if len(table.Errors) == 0 {
		t.Error("Should have errors in resolution table")
	}
}
