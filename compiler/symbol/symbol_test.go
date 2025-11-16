package symbol

import (
	"testing"
)

func TestSymbolType_String(t *testing.T) {
	tests := []struct {
		symType SymbolType
		want    string
	}{
		{SymbolView, "view"},
		{SymbolFunction, "function"},
		{SymbolClass, "class"},
		{SymbolVariable, "variable"},
		{SymbolType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.symType.String()
			if got != tt.want {
				t.Errorf("SymbolType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVisibility_String(t *testing.T) {
	tests := []struct {
		vis  Visibility
		want string
	}{
		{Public, "public"},
		{Private, "private"},
		{Visibility(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.vis.String()
			if got != tt.want {
				t.Errorf("Visibility.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
