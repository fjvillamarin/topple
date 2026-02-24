package symbol

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"testing"
)

func createTestName(name string) *ast.Name {
	return &ast.Name{
		Token: lexer.Token{Lexeme: name, Type: lexer.Identifier},
		Span:  lexer.Span{},
	}
}

func TestCollectViewSymbol(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.ViewStmt{
				Name: createTestName("MyView"),
				Params: &ast.ParameterList{
					Parameters: []*ast.Parameter{},
				},
				Body: []ast.Stmt{},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	if len(moduleSymbols.Symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(moduleSymbols.Symbols))
	}

	symbol, exists := moduleSymbols.LookupSymbol("MyView")
	if !exists {
		t.Fatal("MyView symbol not found")
	}

	if symbol.Type != SymbolView {
		t.Errorf("expected SymbolView, got %v", symbol.Type)
	}
	if symbol.Visibility != Public {
		t.Errorf("expected Public visibility, got %v", symbol.Visibility)
	}
}

func TestCollectFunctionSymbol(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.Function{
				Name: createTestName("my_function"),
				Parameters: &ast.ParameterList{
					Parameters: []*ast.Parameter{},
				},
				Body: []ast.Stmt{},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("my_function")
	if !exists {
		t.Fatal("my_function symbol not found")
	}

	if symbol.Type != SymbolFunction {
		t.Errorf("expected SymbolFunction, got %v", symbol.Type)
	}
}

func TestCollectClassSymbol(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.Class{
				Name: createTestName("MyClass"),
				Args: []*ast.Argument{},
				Body: []ast.Stmt{},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("MyClass")
	if !exists {
		t.Fatal("MyClass symbol not found")
	}

	if symbol.Type != SymbolClass {
		t.Errorf("expected SymbolClass, got %v", symbol.Type)
	}
}

func TestCollectVariableSymbol(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{
					createTestName("my_var"),
				},
				Value: &ast.Literal{
					Token: lexer.Token{Lexeme: "42", Type: lexer.Number},
					Span:  lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("my_var")
	if !exists {
		t.Fatal("my_var symbol not found")
	}

	if symbol.Type != SymbolVariable {
		t.Errorf("expected SymbolVariable, got %v", symbol.Type)
	}
}

func TestCollectMultipleSymbols(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.ViewStmt{
				Name:   createTestName("View1"),
				Params: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:   []ast.Stmt{},
				Span:   lexer.Span{},
			},
			&ast.Function{
				Name:       createTestName("func1"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
			&ast.Class{
				Name: createTestName("Class1"),
				Args: []*ast.Argument{},
				Body: []ast.Stmt{},
				Span: lexer.Span{},
			},
			&ast.AssignStmt{
				Targets: []ast.Expr{createTestName("var1")},
				Value:   &ast.Literal{Token: lexer.Token{Lexeme: "1", Type: lexer.Number}, Span: lexer.Span{}},
				Span:    lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	expectedSymbols := map[string]SymbolType{
		"View1":  SymbolView,
		"func1":  SymbolFunction,
		"Class1": SymbolClass,
		"var1":   SymbolVariable,
	}

	if len(moduleSymbols.Symbols) != len(expectedSymbols) {
		t.Fatalf("expected %d symbols, got %d", len(expectedSymbols), len(moduleSymbols.Symbols))
	}

	for name, expectedType := range expectedSymbols {
		symbol, exists := moduleSymbols.LookupSymbol(name)
		if !exists {
			t.Errorf("symbol %s not found", name)
			continue
		}
		if symbol.Type != expectedType {
			t.Errorf("symbol %s: expected type %v, got %v", name, expectedType, symbol.Type)
		}
	}
}

func TestVisibilityPrivate(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.Function{
				Name:       createTestName("_private_func"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("_private_func")
	if !exists {
		t.Fatal("_private_func symbol not found")
	}

	if symbol.Visibility != Private {
		t.Errorf("expected Private visibility, got %v", symbol.Visibility)
	}
}

func TestGetPublicSymbolsFiltering(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.Function{
				Name:       createTestName("public_func"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
			&ast.Function{
				Name:       createTestName("_private_func"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	publicSymbols := moduleSymbols.GetPublicSymbols()

	if len(publicSymbols) != 1 {
		t.Fatalf("expected 1 public symbol, got %d", len(publicSymbols))
	}

	if publicSymbols[0].Name != "public_func" {
		t.Errorf("expected public_func, got %s", publicSymbols[0].Name)
	}
}

func TestTupleUnpacking(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{
					&ast.TupleExpr{
						Elements: []ast.Expr{
							createTestName("a"),
							createTestName("b"),
							createTestName("c"),
						},
						Span: lexer.Span{},
					},
				},
				Value: &ast.TupleExpr{
					Elements: []ast.Expr{
						&ast.Literal{Token: lexer.Token{Lexeme: "1", Type: lexer.Number}, Span: lexer.Span{}},
						&ast.Literal{Token: lexer.Token{Lexeme: "2", Type: lexer.Number}, Span: lexer.Span{}},
						&ast.Literal{Token: lexer.Token{Lexeme: "3", Type: lexer.Number}, Span: lexer.Span{}},
					},
					Span: lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	expectedVars := []string{"a", "b", "c"}

	if len(moduleSymbols.Symbols) != len(expectedVars) {
		t.Fatalf("expected %d symbols, got %d", len(expectedVars), len(moduleSymbols.Symbols))
	}

	for _, varName := range expectedVars {
		symbol, exists := moduleSymbols.LookupSymbol(varName)
		if !exists {
			t.Errorf("symbol %s not found", varName)
			continue
		}
		if symbol.Type != SymbolVariable {
			t.Errorf("symbol %s: expected SymbolVariable, got %v", varName, symbol.Type)
		}
	}
}

func TestListUnpacking(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AssignStmt{
				Targets: []ast.Expr{
					&ast.ListExpr{
						Elements: []ast.Expr{
							createTestName("x"),
							createTestName("y"),
						},
						Span: lexer.Span{},
					},
				},
				Value: &ast.ListExpr{
					Elements: []ast.Expr{
						&ast.Literal{Token: lexer.Token{Lexeme: "10", Type: lexer.Number}, Span: lexer.Span{}},
						&ast.Literal{Token: lexer.Token{Lexeme: "20", Type: lexer.Number}, Span: lexer.Span{}},
					},
					Span: lexer.Span{},
				},
				Span: lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	expectedVars := []string{"x", "y"}

	if len(moduleSymbols.Symbols) != len(expectedVars) {
		t.Fatalf("expected %d symbols, got %d", len(expectedVars), len(moduleSymbols.Symbols))
	}

	for _, varName := range expectedVars {
		symbol, exists := moduleSymbols.LookupSymbol(varName)
		if !exists {
			t.Errorf("symbol %s not found", varName)
			continue
		}
		if symbol.Type != SymbolVariable {
			t.Errorf("symbol %s: expected SymbolVariable, got %v", varName, symbol.Type)
		}
	}
}

func TestAnnotationStmt(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.AnnotationStmt{
				Target: createTestName("typed_var"),
				Type: &ast.Name{
					Token: lexer.Token{Lexeme: "int", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
				HasValue: false,
				Span:     lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("typed_var")
	if !exists {
		t.Fatal("typed_var symbol not found")
	}

	if symbol.Type != SymbolVariable {
		t.Errorf("expected SymbolVariable, got %v", symbol.Type)
	}
}

func TestEmptyModule(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/file.psx")
	moduleSymbols := collector.CollectFromModule(module)

	if len(moduleSymbols.Symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(moduleSymbols.Symbols))
	}
}

func TestCollectorReset(t *testing.T) {
	collector := NewCollector("/test/file.psx")

	// First collection
	module1 := &ast.Module{
		Body: []ast.Stmt{
			&ast.Function{
				Name:       createTestName("func1"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	moduleSymbols1 := collector.CollectFromModule(module1)
	if len(moduleSymbols1.Symbols) != 1 {
		t.Fatalf("expected 1 symbol in first collection, got %d", len(moduleSymbols1.Symbols))
	}

	// Second collection should reset
	module2 := &ast.Module{
		Body: []ast.Stmt{
			&ast.Function{
				Name:       createTestName("func2"),
				Parameters: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:       []ast.Stmt{},
				Span:       lexer.Span{},
			},
		},
		Span: lexer.Span{},
	}

	moduleSymbols2 := collector.CollectFromModule(module2)
	if len(moduleSymbols2.Symbols) != 1 {
		t.Fatalf("expected 1 symbol in second collection, got %d", len(moduleSymbols2.Symbols))
	}

	// Should only have func2, not func1
	if _, exists := moduleSymbols2.LookupSymbol("func1"); exists {
		t.Error("func1 should not exist in second collection (collector should reset)")
	}

	if _, exists := moduleSymbols2.LookupSymbol("func2"); !exists {
		t.Error("func2 should exist in second collection")
	}
}

func TestSymbolLocation(t *testing.T) {
	module := &ast.Module{
		Body: []ast.Stmt{
			&ast.ViewStmt{
				Name:   createTestName("TestView"),
				Params: &ast.ParameterList{Parameters: []*ast.Parameter{}},
				Body:   []ast.Stmt{},
				Span: lexer.Span{
					Start: lexer.Position{Line: 5, Column: 1},
				},
			},
		},
		Span: lexer.Span{},
	}

	collector := NewCollector("/test/module.psx")
	moduleSymbols := collector.CollectFromModule(module)

	symbol, exists := moduleSymbols.LookupSymbol("TestView")
	if !exists {
		t.Fatal("TestView symbol not found")
	}

	if symbol.Location.File != "/test/module.psx" {
		t.Errorf("expected file '/test/module.psx', got '%s'", symbol.Location.File)
	}
	if symbol.Location.Line != 5 {
		t.Errorf("expected line 5, got %d", symbol.Location.Line)
	}
	if symbol.Location.Column != 1 {
		t.Errorf("expected column 1, got %d", symbol.Location.Column)
	}
}
