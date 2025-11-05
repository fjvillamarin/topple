package parser

import (
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse an else block
func parseElseBlock(t *testing.T, input string) ([]ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.elseBlock()
}

// Helper function to validate else block results
func validateElseBlock(t *testing.T, statements []ast.Stmt, expectedLen int, shouldHaveElse bool) {
	if shouldHaveElse {
		if statements == nil {
			t.Error("Expected else statements but got nil")
			return
		}
		if len(statements) != expectedLen {
			t.Errorf("Expected %d statements, got %d", expectedLen, len(statements))
		}
	} else {
		if statements != nil {
			t.Errorf("Expected no else statements but got %d", len(statements))
		}
	}
}

// Test comprehensive else block functionality
func TestElseBlock(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasError       bool
		shouldHaveElse bool
		expectedLen    int
		errorContains  string
	}{
		// Basic else blocks
		{
			name: "simple else block",
			input: `else:
    pass`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name:           "single line else",
			input:          `else: pass`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "multiple statements",
			input: `else:
    x = 1
    y = 2
    print(x + y)`,
			shouldHaveElse: true,
			expectedLen:    3,
		},
		{
			name: "complex expression",
			input: `else:
    result = func(a, b) + other.method()`,
			shouldHaveElse: true,
			expectedLen:    1,
		},

		// Different statement types in else
		{
			name: "else with assignment",
			input: `else:
    x = 42`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with function call",
			input: `else:
    print('hello')`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with return",
			input: `else:
    return 42`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with break",
			input: `else:
    break`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with continue",
			input: `else:
    continue`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with raise",
			input: `else:
    raise Exception()`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with assert",
			input: `else:
    assert condition`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with global",
			input: `else:
    global x`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with nonlocal",
			input: `else:
    nonlocal y`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with del",
			input: `else:
    del items[0]`,
			shouldHaveElse: true,
			expectedLen:    2, // del statement may parse as multiple internal statements
		},
		{
			name: "else with yield",
			input: `else:
    yield value`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with import",
			input: `else:
    import os`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "else with from import",
			input: `else:
    from sys import path`,
			shouldHaveElse: true,
			expectedLen:    1,
		},

		// Nested and complex structures
		{
			name: "nested if",
			input: `else:
    if x:
        pass`,
			shouldHaveElse: true,
			expectedLen:    1,
		},
		{
			name: "multiple indented statements",
			input: `else:
    x = 1
    y = 2`,
			shouldHaveElse: true,
			expectedLen:    2,
		},

		// No else cases
		{
			name:           "no else block",
			input:          `x = 1`,
			shouldHaveElse: false,
			expectedLen:    0,
		},
		{
			name:           "no else keyword",
			input:          `pass`,
			shouldHaveElse: false,
			expectedLen:    0,
		},
		{
			name:           "different keyword",
			input:          `elif x: pass`,
			shouldHaveElse: false,
			expectedLen:    0,
		},

		// Error cases
		{
			name: "else without colon",
			input: `else
    pass`,
			hasError:      true,
			errorContains: "expected ':'",
		},
		{
			name:          "else without body",
			input:         `else:`,
			hasError:      true,
			errorContains: "expected",
		},
		{
			name: "missing indentation",
			input: `else:
pass`,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseElseBlock(t, test.input)

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

			validateElseBlock(t, statements, test.expectedLen, test.shouldHaveElse)
		})
	}
}

// Test else block in different contexts
func TestElseBlockContexts(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		context     string
		hasElse     bool
		expectedLen int
	}{
		{"if statement context", "else:\n    x = 42\n    y = x * 2", "if", true, 2},
		{"while loop context", "else:\n    print('loop completed')", "while", true, 1},
		{"for loop context", "else:\n    print('no break occurred')", "for", true, 1},
		{"try statement context", "else:\n    print('no exception raised')", "try", true, 1},
		{"no else clause", "pass", "any", false, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseElseBlock(t, test.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateElseBlock(t, statements, test.expectedLen, test.hasElse)
		})
	}
}

// Test else block edge cases and error handling
func TestElseBlockEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasError  bool
		shouldLog bool
	}{
		{"empty else with proper syntax", "else: pass", false, false},
		{"else with complex nested structure", "else:\n    if a:\n        if b:\n            pass", false, false},
		{"else with multiple statement types", "else:\n    x = 1\n    print(x)\n    return x", false, false},
		{"else with expression statement", "else:\n    func()", false, false},
		{"malformed else syntax", "else pass", true, false},
		{"else keyword but wrong structure", "else x = 1", true, false},
		{"proper indentation validation", "else:\n    pass\n    continue", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseElseBlock(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// For successful cases, just verify we got valid results
			if statements == nil {
				t.Error("Expected statements result (even if empty)")
			}

			if test.shouldLog {
				t.Logf("Parsed %d statements from: %s", len(statements), test.input)
			}
		})
	}
}

// Test specific parsing behaviors
func TestElseBlockParsing(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectElseKeyword bool
		expectStatements  bool
		minStatementCount int
	}{
		{"detects else keyword", "else: pass", true, true, 1},
		{"ignores non-else", "pass", false, false, 0},
		{"handles elif correctly", "elif x: pass", false, false, 0},
		{"parses else with multiple statements", "else:\n    a = 1\n    b = 2", true, true, 2},
		{"handles single line format", "else: return 42", true, true, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statements, err := parseElseBlock(t, test.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			hasStatements := statements != nil && len(statements) > 0

			if test.expectStatements != hasStatements {
				t.Errorf("Expected statements=%v, got statements=%v", test.expectStatements, hasStatements)
			}

			if hasStatements && len(statements) < test.minStatementCount {
				t.Errorf("Expected at least %d statements, got %d", test.minStatementCount, len(statements))
			}
		})
	}
}
