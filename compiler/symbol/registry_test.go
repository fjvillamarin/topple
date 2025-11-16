package symbol

import (
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if len(registry.modules) != 0 {
		t.Errorf("expected empty registry, got %d modules", len(registry.modules))
	}
}

func TestRegisterModule(t *testing.T) {
	registry := NewRegistry()

	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "MyView",
		Type:       SymbolView,
		Visibility: Public,
	})

	registry.RegisterModule("/test/module.psx", moduleSymbols)

	if !registry.HasModule("/test/module.psx") {
		t.Error("module not registered")
	}
}

func TestLookupSymbol(t *testing.T) {
	registry := NewRegistry()

	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "MyView",
		Type:       SymbolView,
		Visibility: Public,
	})
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "_private_func",
		Type:       SymbolFunction,
		Visibility: Private,
	})

	registry.RegisterModule("/test/module.psx", moduleSymbols)

	tests := []struct {
		name       string
		modulePath string
		symbolName string
		wantErr    bool
		wantType   SymbolType
	}{
		{
			name:       "lookup existing public symbol",
			modulePath: "/test/module.psx",
			symbolName: "MyView",
			wantErr:    false,
			wantType:   SymbolView,
		},
		{
			name:       "lookup existing private symbol",
			modulePath: "/test/module.psx",
			symbolName: "_private_func",
			wantErr:    false,
			wantType:   SymbolFunction,
		},
		{
			name:       "lookup non-existent symbol",
			modulePath: "/test/module.psx",
			symbolName: "NonExistent",
			wantErr:    true,
		},
		{
			name:       "lookup in non-existent module",
			modulePath: "/test/nonexistent.psx",
			symbolName: "MyView",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, err := registry.LookupSymbol(tt.modulePath, tt.symbolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupSymbol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if symbol.Name != tt.symbolName {
					t.Errorf("expected symbol name %s, got %s", tt.symbolName, symbol.Name)
				}
				if symbol.Type != tt.wantType {
					t.Errorf("expected symbol type %s, got %s", tt.wantType, symbol.Type)
				}
			}
		})
	}
}

func TestGetPublicSymbols(t *testing.T) {
	registry := NewRegistry()

	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "PublicView",
		Type:       SymbolView,
		Visibility: Public,
	})
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "PublicFunc",
		Type:       SymbolFunction,
		Visibility: Public,
	})
	moduleSymbols.AddSymbol(&Symbol{
		Name:       "_private_var",
		Type:       SymbolVariable,
		Visibility: Private,
	})

	registry.RegisterModule("/test/module.psx", moduleSymbols)

	publicSymbols, err := registry.GetPublicSymbols("/test/module.psx")
	if err != nil {
		t.Fatalf("GetPublicSymbols() error = %v", err)
	}

	if len(publicSymbols) != 2 {
		t.Errorf("expected 2 public symbols, got %d", len(publicSymbols))
	}

	for _, symbol := range publicSymbols {
		if symbol.Visibility != Public {
			t.Errorf("expected public symbol, got %s", symbol.Visibility)
		}
	}
}

func TestClear(t *testing.T) {
	registry := NewRegistry()

	moduleSymbols := NewModuleSymbols("/test/module.psx")
	registry.RegisterModule("/test/module.psx", moduleSymbols)

	if !registry.HasModule("/test/module.psx") {
		t.Error("module not registered")
	}

	registry.Clear()

	if registry.HasModule("/test/module.psx") {
		t.Error("module still registered after Clear()")
	}
}

func TestGetStats(t *testing.T) {
	registry := NewRegistry()

	module1 := NewModuleSymbols("/test/module1.psx")
	module1.AddSymbol(&Symbol{Name: "View1", Type: SymbolView, Visibility: Public})
	module1.AddSymbol(&Symbol{Name: "func1", Type: SymbolFunction, Visibility: Public})

	module2 := NewModuleSymbols("/test/module2.psx")
	module2.AddSymbol(&Symbol{Name: "View2", Type: SymbolView, Visibility: Public})
	module2.AddSymbol(&Symbol{Name: "Class1", Type: SymbolClass, Visibility: Public})
	module2.AddSymbol(&Symbol{Name: "var1", Type: SymbolVariable, Visibility: Public})

	registry.RegisterModule("/test/module1.psx", module1)
	registry.RegisterModule("/test/module2.psx", module2)

	stats := registry.GetStats()

	if stats.ModuleCount != 2 {
		t.Errorf("expected 2 modules, got %d", stats.ModuleCount)
	}
	if stats.SymbolCount != 5 {
		t.Errorf("expected 5 symbols, got %d", stats.SymbolCount)
	}
	if stats.ViewCount != 2 {
		t.Errorf("expected 2 views, got %d", stats.ViewCount)
	}
	if stats.FunctionCount != 1 {
		t.Errorf("expected 1 function, got %d", stats.FunctionCount)
	}
	if stats.ClassCount != 1 {
		t.Errorf("expected 1 class, got %d", stats.ClassCount)
	}
	if stats.VariableCount != 1 {
		t.Errorf("expected 1 variable, got %d", stats.VariableCount)
	}
}

func TestGetAllModules(t *testing.T) {
	registry := NewRegistry()

	module1 := NewModuleSymbols("/test/module1.psx")
	module2 := NewModuleSymbols("/test/module2.psx")

	registry.RegisterModule("/test/module1.psx", module1)
	registry.RegisterModule("/test/module2.psx", module2)

	modules := registry.GetAllModules()

	if len(modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(modules))
	}

	// Check that both modules are in the list
	found1 := false
	found2 := false
	for _, path := range modules {
		if path == "/test/module1.psx" {
			found1 = true
		}
		if path == "/test/module2.psx" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("module1.psx not found in GetAllModules()")
	}
	if !found2 {
		t.Error("module2.psx not found in GetAllModules()")
	}
}

func TestGetModuleSymbols(t *testing.T) {
	registry := NewRegistry()

	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{Name: "Symbol1", Type: SymbolView, Visibility: Public})

	registry.RegisterModule("/test/module.psx", moduleSymbols)

	t.Run("existing module", func(t *testing.T) {
		symbols, err := registry.GetModuleSymbols("/test/module.psx")
		if err != nil {
			t.Fatalf("GetModuleSymbols() error = %v", err)
		}
		if len(symbols.Symbols) != 1 {
			t.Errorf("expected 1 symbol, got %d", len(symbols.Symbols))
		}
	})

	t.Run("non-existent module", func(t *testing.T) {
		_, err := registry.GetModuleSymbols("/test/missing.psx")
		if err == nil {
			t.Error("expected error for non-existent module, got nil")
		}

		if regErr, ok := err.(*RegistryError); ok {
			if regErr.Type != ModuleNotRegistered {
				t.Errorf("expected ModuleNotRegistered error, got %v", regErr.Type)
			}
		} else {
			t.Errorf("expected RegistryError, got %T", err)
		}
	})
}

func TestModuleSymbols_LookupSymbol(t *testing.T) {
	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{Name: "Test", Type: SymbolView, Visibility: Public})

	symbol, exists := moduleSymbols.LookupSymbol("Test")
	if !exists {
		t.Error("expected symbol to exist")
	}
	if symbol.Name != "Test" {
		t.Errorf("expected symbol name 'Test', got '%s'", symbol.Name)
	}

	_, exists = moduleSymbols.LookupSymbol("Missing")
	if exists {
		t.Error("expected symbol to not exist")
	}
}

func TestModuleSymbols_GetAllSymbols(t *testing.T) {
	moduleSymbols := NewModuleSymbols("/test/module.psx")
	moduleSymbols.AddSymbol(&Symbol{Name: "Public", Type: SymbolView, Visibility: Public})
	moduleSymbols.AddSymbol(&Symbol{Name: "_Private", Type: SymbolFunction, Visibility: Private})

	allSymbols := moduleSymbols.GetAllSymbols()
	if len(allSymbols) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(allSymbols))
	}
}

func TestDetermineVisibility(t *testing.T) {
	tests := []struct {
		name string
		want Visibility
	}{
		{"public_name", Public},
		{"PublicName", Public},
		{"_private_name", Private},
		{"__dunder__", Private},
		{"", Public},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineVisibility(tt.name)
			if got != tt.want {
				t.Errorf("determineVisibility(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestExtractLocation(t *testing.T) {
	node := &ast.ViewStmt{
		Name: &ast.Name{
			Token: lexer.Token{Lexeme: "Test"},
		},
		Span: lexer.Span{
			Start: lexer.Position{Line: 10, Column: 5},
		},
	}

	location := extractLocation(node, "/test/file.psx")

	if location.File != "/test/file.psx" {
		t.Errorf("expected file '/test/file.psx', got '%s'", location.File)
	}
	if location.Line != 10 {
		t.Errorf("expected line 10, got %d", location.Line)
	}
	if location.Column != 5 {
		t.Errorf("expected column 5, got %d", location.Column)
	}
}
