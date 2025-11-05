package parser

import (
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse an f-string expression
func parseFString(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to validate f-string structure
func validateFString(t *testing.T, expr ast.Expr, expectedParts int) *ast.FString {
	fstring, ok := expr.(*ast.FString)
	if !ok {
		t.Fatalf("Expected *ast.FString, got %T", expr)
	}

	if fstring == nil {
		t.Fatal("F-string should not be nil")
	}

	if expectedParts >= 0 && len(fstring.Parts) != expectedParts {
		t.Errorf("Expected %d parts, got %d", expectedParts, len(fstring.Parts))
	}

	return fstring
}

// Helper function to find replacement field in f-string
func findReplacementField(fstring *ast.FString, index int) *ast.FStringReplacementField {
	if index >= len(fstring.Parts) {
		return nil
	}

	partIndex := 0
	for _, part := range fstring.Parts {
		if rf, ok := part.(*ast.FStringReplacementField); ok {
			if partIndex == index {
				return rf
			}
			partIndex++
		}
	}
	return nil
}

// Helper function to count replacement fields in f-string
func countReplacementFields(fstring *ast.FString) int {
	count := 0
	for _, part := range fstring.Parts {
		if _, ok := part.(*ast.FStringReplacementField); ok {
			count++
		}
	}
	return count
}

// Helper function to validate replacement field features
func validateReplacementField(t *testing.T, rf *ast.FStringReplacementField, expectedEqual bool, expectedConversion, expectedFormatSpec bool) {
	if rf.Expression == nil {
		t.Error("Replacement field should have an expression")
	}

	if rf.Equal != expectedEqual {
		t.Errorf("Expected equal=%v, got %v", expectedEqual, rf.Equal)
	}

	hasConversion := rf.Conversion != nil
	if hasConversion != expectedConversion {
		t.Errorf("Expected conversion=%v, got %v", expectedConversion, hasConversion)
	}

	hasFormatSpec := rf.FormatSpec != nil
	if hasFormatSpec != expectedFormatSpec {
		t.Errorf("Expected format spec=%v, got %v", expectedFormatSpec, hasFormatSpec)
	}
}

// Test comprehensive f-string functionality
func TestFString(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		hasError             bool
		expectedParts        int
		expectedReplacements int
		description          string
	}{
		// Basic f-strings
		{
			name:                 "simple variable",
			input:                `f"Hello {name}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "basic variable interpolation",
		},
		{
			name:                 "multiple expressions",
			input:                `f"Hello {name}, you are {age} years old"`,
			expectedParts:        5,
			expectedReplacements: 2,
			description:          "multiple variable interpolations",
		},
		{
			name:                 "plain string",
			input:                `f"Just a plain string"`,
			expectedParts:        1,
			expectedReplacements: 0,
			description:          "f-string without expressions",
		},
		{
			name:                 "consecutive expressions",
			input:                `f"{first}{second}{third}"`,
			expectedParts:        3,
			expectedReplacements: 3,
			description:          "consecutive replacement fields",
		},
		{
			name:                 "expression with spaces",
			input:                `f"  {value}  "`,
			expectedParts:        3,
			expectedReplacements: 1,
			description:          "replacement field with surrounding literals",
		},
		{
			name:                 "starting with expression",
			input:                `f"{greeting} world"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "f-string starting with replacement field",
		},
		{
			name:                 "ending with expression",
			input:                `f"Hello {name}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "f-string ending with replacement field",
		},

		// Complex expressions
		{
			name:                 "arithmetic expression",
			input:                `f"Result: {x + y}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "arithmetic in replacement field",
		},
		{
			name:                 "method call",
			input:                `f"Upper: {text.upper()}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "method call in replacement field",
		},
		{
			name:                 "attribute access",
			input:                `f"Name: {user.name}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "attribute access in replacement field",
		},
		{
			name:                 "indexing",
			input:                `f"First: {items[0]}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "indexing in replacement field",
		},
		{
			name:                 "function call",
			input:                `f"Length: {len(items)}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "function call in replacement field",
		},
		{
			name:                 "conditional expression",
			input:                `f"Status: {'active' if is_active else 'inactive'}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "ternary expression in replacement field",
		},
		{
			name:                 "list comprehension",
			input:                `f"Squares: {[x**2 for x in range(5)]}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "list comprehension in replacement field",
		},
		{
			name:                 "dictionary access",
			input:                `f"Value: {data['key']}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "dictionary access in replacement field",
		},
		{
			name:                 "complex arithmetic",
			input:                `f"Calculation: {(a + b) * c / d}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "complex arithmetic in replacement field",
		},
		{
			name:                 "boolean expression",
			input:                `f"Check: {x > 0 and y < 10}"`,
			expectedParts:        2,
			expectedReplacements: 1,
			description:          "boolean logic in replacement field",
		},

		// Error cases
		{
			name:        "unterminated expression",
			input:       `f"Hello {name"`,
			hasError:    true,
			description: "missing closing brace",
		},
		{
			name:        "empty expression",
			input:       `f"Hello {}"`,
			hasError:    true,
			description: "empty replacement field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseFString(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			fstring := validateFString(t, expr, test.expectedParts)

			actualReplacements := countReplacementFields(fstring)
			if actualReplacements != test.expectedReplacements {
				t.Errorf("Expected %d replacement fields, got %d", test.expectedReplacements, actualReplacements)
			}

			// Verify all parts are valid types
			for i, part := range fstring.Parts {
				switch part.(type) {
				case *ast.FStringMiddle, *ast.FStringReplacementField:
					// Valid part types
				default:
					t.Errorf("Unexpected f-string part type at index %d: %T", i, part)
				}
			}
		})
	}
}

// Test f-string conversions (!r, !s, !a)
func TestFStringConversions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedType  string
		errorContains string
	}{
		{"repr conversion", `f"{value!r}"`, false, "r", ""},
		{"str conversion", `f"{value!s}"`, false, "s", ""},
		{"ascii conversion", `f"{value!a}"`, false, "a", ""},
		{"invalid conversion type", `f"{value!x}"`, true, "", "invalid conversion"},
		{"conversion without type", `f"{value!}"`, true, "", "expected conversion"},
		{"multiple conversions", `f"{value!r!s}"`, true, "", "expected '}'"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseFString(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			fstring := validateFString(t, expr, -1)
			rf := findReplacementField(fstring, 0)
			if rf == nil {
				t.Fatal("Expected replacement field")
			}

			if rf.Conversion == nil {
				t.Error("Expected conversion but got none")
				return
			}

			if rf.Conversion.Type != test.expectedType {
				t.Errorf("Expected conversion type %s, got %s", test.expectedType, rf.Conversion.Type)
			}
		})
	}
}

// Test f-string format specifications
func TestFStringFormatSpecs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasError  bool
		hasFormat bool
	}{
		{"simple format spec", `f"{value:10}"`, false, true},
		{"precision format", `f"{value:.2f}"`, false, true},
		{"padding format", `f"{value:>10}"`, false, true},
		{"fill character format", `f"{value:0>10}"`, false, true},
		{"complex format", `f"{value:+#0=10.2f}"`, false, true},
		{"dynamic width", `f"{value:{width}}"`, false, true},
		{"dynamic precision", `f"{value:.{precision}f}"`, false, true},
		{"nested format", `f"{value:{width}.{precision}f}"`, false, true},
		{"no format spec", `f"{value}"`, false, false},
		{"empty format spec", `f"{value:}"`, false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseFString(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			fstring := validateFString(t, expr, -1)
			rf := findReplacementField(fstring, 0)
			if rf == nil {
				t.Fatal("Expected replacement field")
			}

			hasFormatSpec := rf.FormatSpec != nil
			if hasFormatSpec != test.hasFormat {
				t.Errorf("Expected format spec=%v, got %v", test.hasFormat, hasFormatSpec)
			}
		})
	}
}

// Test f-string debug expressions (=)
func TestFStringDebugExpressions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectEqual bool
		hasError    bool
	}{
		{"debug with equals", `f"{value=}"`, true, false},
		{"debug complex expression", `f"{x + y=}"`, true, false},
		{"debug with format", `f"{value=:.2f}"`, true, false},
		{"debug with conversion", `f"{value=!r}"`, true, false},
		{"debug with both", `f"{value=!r:.2f}"`, true, false},
		{"regular expression", `f"{value}"`, false, false},
		{"equals in expression", `f"{x == y}"`, false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseFString(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			fstring := validateFString(t, expr, -1)
			rf := findReplacementField(fstring, 0)
			if rf == nil {
				t.Fatal("Expected replacement field")
			}

			if rf.Equal != test.expectEqual {
				t.Errorf("Expected equal=%v, got %v", test.expectEqual, rf.Equal)
			}
		})
	}
}

// Test f-string edge cases and complex scenarios
func TestFStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"nested function calls", `f"Result: {func(nested(value))}"`, false},
		{"multiple method chains", `f"Chain: {obj.method().attr.call()}"`, false},
		{"list literal", `f"List: {[1, 2, 3]}"`, false},
		{"function call", `f"Call: {len(items)}"`, false},
		{"simple boolean", `f"Bool: {a and b}"`, false},
		{"arithmetic", `f"Math: {x + y * 2}"`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseFString(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.name, err)
			}

			// For successful cases, just verify we got a valid f-string
			validateFString(t, expr, -1)
		})
	}
}
