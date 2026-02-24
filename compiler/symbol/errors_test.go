package symbol

import (
	"strings"
	"testing"
)

func TestRegistryError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *RegistryError
		contains []string
	}{
		{
			name: "ModuleNotRegistered",
			err: &RegistryError{
				Type:       ModuleNotRegistered,
				ModulePath: "/test/module.psx",
			},
			contains: []string{"module not registered", "/test/module.psx"},
		},
		{
			name: "SymbolNotFound",
			err: &RegistryError{
				Type:       SymbolNotFound,
				ModulePath: "/test/module.psx",
				SymbolName: "MySymbol",
			},
			contains: []string{"symbol", "MySymbol", "not found", "/test/module.psx"},
		},
		{
			name: "DuplicateSymbol with location",
			err: &RegistryError{
				Type:       DuplicateSymbol,
				ModulePath: "/test/module.psx",
				SymbolName: "Duplicate",
				Location:   &Location{File: "/test/file.psx", Line: 10, Column: 5},
			},
			contains: []string{"duplicate symbol", "Duplicate", "/test/module.psx", "/test/file.psx:10:5"},
		},
		{
			name: "InvalidSymbol",
			err: &RegistryError{
				Type:    InvalidSymbol,
				Message: "test message",
			},
			contains: []string{"invalid symbol", "test message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("error message %q should contain %q", errMsg, substr)
				}
			}
		})
	}
}

func TestCollectionError_Error(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		err := &CollectionError{
			FilePath: "/test/file.psx",
			Errors: []error{
				&RegistryError{Type: ModuleNotRegistered, ModulePath: "/test/mod.psx"},
				&RegistryError{Type: SymbolNotFound, ModulePath: "/test/mod.psx", SymbolName: "Sym"},
			},
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "/test/file.psx") {
			t.Errorf("error should contain file path")
		}
		if !strings.Contains(errMsg, "module not registered") {
			t.Errorf("error should contain first error message")
		}
		if !strings.Contains(errMsg, "not found") {
			t.Errorf("error should contain second error message")
		}
	})

	t.Run("without errors", func(t *testing.T) {
		err := &CollectionError{
			FilePath: "/test/file.psx",
			Errors:   []error{},
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "/test/file.psx") {
			t.Errorf("error should contain file path")
		}
		if !strings.Contains(errMsg, "unknown error") {
			t.Errorf("error should mention unknown error")
		}
	})
}

func TestNewErrors(t *testing.T) {
	t.Run("newModuleNotRegisteredError", func(t *testing.T) {
		err := newModuleNotRegisteredError("/test/module.psx")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if regErr, ok := err.(*RegistryError); ok {
			if regErr.Type != ModuleNotRegistered {
				t.Errorf("expected ModuleNotRegistered, got %v", regErr.Type)
			}
		} else {
			t.Errorf("expected RegistryError, got %T", err)
		}
	})

	t.Run("newSymbolNotFoundError", func(t *testing.T) {
		err := newSymbolNotFoundError("/test/module.psx", "Symbol")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if regErr, ok := err.(*RegistryError); ok {
			if regErr.Type != SymbolNotFound {
				t.Errorf("expected SymbolNotFound, got %v", regErr.Type)
			}
		} else {
			t.Errorf("expected *RegistryError, got %T", err)
		}
	})

	t.Run("newDuplicateSymbolError", func(t *testing.T) {
		loc := &Location{File: "/test/file.psx", Line: 1, Column: 1}
		err := newDuplicateSymbolError("/test/module.psx", "Dup", loc)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if regErr, ok := err.(*RegistryError); ok {
			if regErr.Type != DuplicateSymbol {
				t.Errorf("expected DuplicateSymbol, got %v", regErr.Type)
			}
			if regErr.Location == nil {
				t.Error("expected location to be set")
			}
		}
	})
}
