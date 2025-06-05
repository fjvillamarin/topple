package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

func TestSimpleSlices(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		hasError        bool
		isSlice         bool
		hasStart        bool
		hasEnd          bool
		hasStep         bool
		expectedIndices int
	}{
		{
			name:            "simple index",
			input:           "arr[0]",
			isSlice:         false,
			expectedIndices: 1,
		},
		{
			name:            "slice with start and end",
			input:           "arr[1:3]",
			isSlice:         true,
			hasStart:        true,
			hasEnd:          true,
			hasStep:         false,
			expectedIndices: 1,
		},
		{
			name:            "slice with start only",
			input:           "arr[1:]",
			isSlice:         true,
			hasStart:        true,
			hasEnd:          false,
			hasStep:         false,
			expectedIndices: 1,
		},
		{
			name:            "slice with end only",
			input:           "arr[:3]",
			isSlice:         true,
			hasStart:        false,
			hasEnd:          true,
			hasStep:         false,
			expectedIndices: 1,
		},
		{
			name:            "slice with step",
			input:           "arr[::2]",
			isSlice:         true,
			hasStart:        false,
			hasEnd:          false,
			hasStep:         true,
			expectedIndices: 1,
		},
		{
			name:            "slice with all parts",
			input:           "arr[1:5:2]",
			isSlice:         true,
			hasStart:        true,
			hasEnd:          true,
			hasStep:         true,
			expectedIndices: 1,
		},
		{
			name:            "full slice",
			input:           "arr[:]",
			isSlice:         true,
			hasStart:        false,
			hasEnd:          false,
			hasStep:         false,
			expectedIndices: 1,
		},
		{
			name:            "negative indices",
			input:           "arr[-1]",
			isSlice:         false,
			expectedIndices: 1,
		},
		{
			name:            "negative slice",
			input:           "arr[-3:-1]",
			isSlice:         true,
			hasStart:        true,
			hasEnd:          true,
			hasStep:         false,
			expectedIndices: 1,
		},
		{
			name:            "expression indices",
			input:           "arr[i + 1]",
			isSlice:         false,
			expectedIndices: 1,
		},
		{
			name:            "expression slice",
			input:           "arr[start:end]",
			isSlice:         true,
			hasStart:        true,
			hasEnd:          true,
			hasStep:         false,
			expectedIndices: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			subscript, ok := expr.(*ast.Subscript)
			if !ok {
				t.Errorf("Expected *ast.Subscript but got %T", expr)
				return
			}

			if subscript.Object == nil {
				t.Errorf("Subscript should have an object")
			}

			if len(subscript.Indices) != tt.expectedIndices {
				t.Errorf("Expected %d indices but got %d", tt.expectedIndices, len(subscript.Indices))
				return
			}

			if len(subscript.Indices) == 0 {
				t.Errorf("Subscript should have at least one index")
				return
			}

			// Validate slice vs index based on test expectations
			if tt.isSlice {
				slice, ok := subscript.Indices[0].(*ast.Slice)
				if !ok {
					t.Errorf("Expected *ast.Slice but got %T", subscript.Indices[0])
					return
				}

				// Validate slice components based on test expectations
				if tt.hasStart {
					if slice.StartIndex == nil {
						t.Errorf("Expected slice to have start index")
					}
				} else {
					if slice.StartIndex != nil {
						t.Errorf("Expected slice to have no start index")
					}
				}

				if tt.hasEnd {
					if slice.EndIndex == nil {
						t.Errorf("Expected slice to have end index")
					}
				} else {
					if slice.EndIndex != nil {
						t.Errorf("Expected slice to have no end index")
					}
				}

				if tt.hasStep {
					if slice.Step == nil {
						t.Errorf("Expected slice to have step")
					}
				} else {
					if slice.Step != nil {
						t.Errorf("Expected slice to have no step")
					}
				}
			} else {
				// Should be a simple index, not a slice
				if _, ok := subscript.Indices[0].(*ast.Slice); ok {
					t.Errorf("Expected simple index, not slice")
				}
			}
		})
	}
}

func TestMultiDimensionalSlices(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		hasError        bool
		expectedIndices int
	}{
		{
			name:            "two dimensional index",
			input:           "matrix[0, 1]",
			expectedIndices: 2,
		},
		{
			name:            "two dimensional slice",
			input:           "matrix[1:3, 2:4]",
			expectedIndices: 2,
		},
		{
			name:            "mixed index and slice",
			input:           "matrix[0, 1:3]",
			expectedIndices: 2,
		},
		{
			name:            "slice with step in multi-dim",
			input:           "matrix[::2, 1:]",
			expectedIndices: 2,
		},
		{
			name:            "complex multi-dimensional",
			input:           "tensor[i, j:k, ::2]",
			expectedIndices: 3,
		},
		{
			name:            "three dimensional",
			input:           "cube[x, y, z]",
			expectedIndices: 3,
		},
		{
			name:            "trailing comma allowed",
			input:           "matrix[0, 1,]",
			expectedIndices: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			subscript, ok := expr.(*ast.Subscript)
			if !ok {
				t.Errorf("Expected *ast.Subscript but got %T", expr)
				return
			}

			if subscript.Object == nil {
				t.Errorf("Subscript should have an object")
			}

			if len(subscript.Indices) != tt.expectedIndices {
				t.Errorf("Expected %d indices but got %d", tt.expectedIndices, len(subscript.Indices))
			}
		})
	}
}

func TestSliceExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:  "expression in slice start",
			input: "arr[i + 1:j]",
		},
		{
			name:  "expression in slice end",
			input: "arr[i:j - 1]",
		},
		{
			name:  "expression in slice step",
			input: "arr[::step * 2]",
		},
		{
			name:  "function call in index",
			input: "arr[len(data)]",
		},
		{
			name:  "method call in slice",
			input: "arr[obj.get_start():obj.get_end()]",
		},
		{
			name:  "complex expression slice",
			input: "arr[(x + y) * 2:(a - b) / 2]",
		},
		{
			name:  "nested subscript in slice",
			input: "arr[data[i]:data[j]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			subscript, ok := expr.(*ast.Subscript)
			if !ok {
				t.Errorf("Expected *ast.Subscript but got %T", expr)
				return
			}

			if subscript.Object == nil {
				t.Errorf("Subscript should have an object")
			}

			if len(subscript.Indices) == 0 {
				t.Errorf("Subscript should have at least one index")
			}
		})
	}
}

func TestEmptySlices(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
		hasStart bool
		hasEnd   bool
		hasStep  bool
	}{
		{
			name:     "empty slice",
			input:    "arr[:]",
			hasStart: false,
			hasEnd:   false,
			hasStep:  false,
		},
		{
			name:     "empty start",
			input:    "arr[:5]",
			hasStart: false,
			hasEnd:   true,
			hasStep:  false,
		},
		{
			name:     "empty end",
			input:    "arr[2:]",
			hasStart: true,
			hasEnd:   false,
			hasStep:  false,
		},
		{
			name:     "empty start and end",
			input:    "arr[::2]",
			hasStart: false,
			hasEnd:   false,
			hasStep:  true,
		},
		{
			name:     "empty step",
			input:    "arr[1:5:]",
			hasStart: true,
			hasEnd:   true,
			hasStep:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			expr, err := parser.expression()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			subscript, ok := expr.(*ast.Subscript)
			if !ok {
				t.Errorf("Expected *ast.Subscript but got %T", expr)
				return
			}

			if len(subscript.Indices) != 1 {
				t.Errorf("Expected 1 index but got %d", len(subscript.Indices))
				return
			}

			slice, ok := subscript.Indices[0].(*ast.Slice)
			if !ok {
				t.Errorf("Expected *ast.Slice but got %T", subscript.Indices[0])
				return
			}

			// Validate slice components based on test expectations
			if tt.hasStart {
				if slice.StartIndex == nil {
					t.Errorf("Expected slice to have start index")
				}
			} else {
				if slice.StartIndex != nil {
					t.Errorf("Expected slice to have no start index")
				}
			}

			if tt.hasEnd {
				if slice.EndIndex == nil {
					t.Errorf("Expected slice to have end index")
				}
			} else {
				if slice.EndIndex != nil {
					t.Errorf("Expected slice to have no end index")
				}
			}

			if tt.hasStep {
				if slice.Step == nil {
					t.Errorf("Expected slice to have step")
				}
			} else {
				if slice.Step != nil {
					t.Errorf("Expected slice to have no step")
				}
			}
		})
	}
}

func TestSliceErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed bracket",
			input: "arr[1:3",
		},
		{
			name:  "missing object",
			input: "[1:3]",
		},
		{
			name:  "invalid slice syntax",
			input: "arr[1:::]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			_, err := parser.expression()

			if err == nil {
				t.Errorf("Expected error but got none for input: %s", tt.input)
			}
		})
	}
}
