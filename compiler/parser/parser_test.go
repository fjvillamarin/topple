package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to create a parser and get tokens from input
func createParserWithTokens(t *testing.T, input string) (*Parser, []lexer.Token) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser, tokens
}

// Helper function to validate parser initialization
func validateParserInit(t *testing.T, parser *Parser, expectedTokenCount int) {
	if parser == nil {
		t.Fatal("NewParser returned nil")
	}

	if len(parser.Tokens) != expectedTokenCount {
		t.Errorf("Expected %d tokens, got %d", expectedTokenCount, len(parser.Tokens))
	}

	if parser.Current != 0 {
		t.Errorf("Expected Current to be 0, got %d", parser.Current)
	}

	if len(parser.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(parser.Errors))
	}
}

// Helper function to parse input and return module with errors
func parseInput(t *testing.T, input string) (*ast.Module, []error) {
	parser := createParser(t, input)
	return parser.Parse()
}

// Helper function to validate successful parse result
func validateParseSuccess(t *testing.T, module *ast.Module, errors []error, expectedStatements int) {
	if len(errors) > 0 {
		t.Fatalf("Unexpected errors: %v", errors)
	}

	if module == nil {
		t.Fatal("Expected module but got nil")
	}

	if expectedStatements >= 0 && len(module.Body) != expectedStatements {
		t.Errorf("Expected %d statements, got %d", expectedStatements, len(module.Body))
	}
}

// Helper function to validate parse result errors
func validateParseResultError(t *testing.T, module *ast.Module, errors []error, expectedErrorText string) {
	if len(errors) == 0 {
		t.Fatal("Expected error but got none")
	}

	if module != nil {
		t.Error("Expected nil module on error but got module")
	}

	if expectedErrorText != "" {
		errorStr := errors[0].Error()
		if !strings.Contains(errorStr, expectedErrorText) {
			t.Errorf("Expected error to contain %q, got %q", expectedErrorText, errorStr)
		}
	}
}

// Helper function to validate parser error structure
func validateParseErrorStruct(t *testing.T, err *ParseError, expectedToken lexer.Token, expectedMessage string) {
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if err.Token != expectedToken {
		t.Errorf("Expected token %v, got %v", expectedToken, err.Token)
	}

	if err.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, err.Message)
	}

	errorStr := err.Error()
	if expectedToken.Type == lexer.EOF {
		if !strings.Contains(errorStr, "at end:") {
			t.Errorf("Expected EOF error to contain 'at end:', got %q", errorStr)
		}
	} else {
		if !strings.Contains(errorStr, "at '"+expectedToken.Lexeme+"':") {
			t.Errorf("Expected error to contain token lexeme, got %q", errorStr)
		}
	}

	span := err.Span()
	if span != expectedToken.Span {
		t.Errorf("Expected span %v, got %v", expectedToken.Span, span)
	}
}

// Test parser initialization
func TestNewParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tokenLen int
	}{
		{
			name:     "empty input",
			input:    "",
			tokenLen: 1, // EOF token
		},
		{
			name:     "single identifier",
			input:    "hello",
			tokenLen: 2, // identifier + EOF
		},
		{
			name:     "simple expression",
			input:    "x + y",
			tokenLen: 4, // x + y + EOF
		},
		{
			name:     "complex expression",
			input:    "func(a, b)",
			tokenLen: 7, // func ( a , b ) + EOF
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser, _ := createParserWithTokens(t, test.input)
			validateParserInit(t, parser, test.tokenLen)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedLen  int
		hasError     bool
		errorMessage string
	}{
		{
			name:        "empty input",
			input:       "",
			expectedLen: 0,
		},
		{
			name:        "single expression",
			input:       "x + y",
			expectedLen: 1,
		},
		{
			name:        "multiple statements",
			input:       "x = 1\ny = 2\nz = x + y",
			expectedLen: 3,
		},
		{
			name:        "expression with newlines",
			input:       "\n\nx + y\n\n",
			expectedLen: 1,
		},
		{
			name:        "function definition",
			input:       "def func():\n    pass",
			expectedLen: 1,
		},
		{
			name:        "class definition",
			input:       "class MyClass:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "if statement",
			input:       "if x:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "for loop",
			input:       "for x in items:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "while loop",
			input:       "while True:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "try statement",
			input:       "try:\n    pass\nexcept:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "with statement",
			input:       "with open('file') as f:\n    pass",
			expectedLen: 1,
		},
		{
			name:        "import statement",
			input:       "import os",
			expectedLen: 1,
		},
		{
			name:        "from import statement",
			input:       "from os import path",
			expectedLen: 1,
		},
		{
			name:        "return statement",
			input:       "return 42",
			expectedLen: 1,
		},
		{
			name:        "yield statement",
			input:       "yield value",
			expectedLen: 1,
		},
		{
			name:        "raise statement",
			input:       "raise Exception()",
			expectedLen: 1,
		},
		{
			name:        "assert statement",
			input:       "assert x > 0",
			expectedLen: 1,
		},
		{
			name:        "global statement",
			input:       "global x, y",
			expectedLen: 1,
		},
		{
			name:        "nonlocal statement",
			input:       "nonlocal x, y",
			expectedLen: 1,
		},
		{
			name:        "pass statement",
			input:       "pass",
			expectedLen: 1,
		},
		{
			name:        "break statement",
			input:       "break",
			expectedLen: 1,
		},
		{
			name:        "continue statement",
			input:       "continue",
			expectedLen: 1,
		},
		{
			name:        "del statement",
			input:       "del x",
			expectedLen: 1,
		},
		{
			name:        "type alias statement",
			input:       "type MyInt = int",
			expectedLen: 1,
		},
		{
			name:        "decorator statement",
			input:       "@decorator\ndef func():\n    pass",
			expectedLen: 1,
		},
		{
			name:        "match statement",
			input:       "match x:\n    case 1:\n        pass",
			expectedLen: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module, errors := parseInput(t, test.input)

			if test.hasError {
				validateParseResultError(t, module, errors, test.errorMessage)
			} else {
				validateParseSuccess(t, module, errors, test.expectedLen)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedMsg  string
		containsText string
		description  string
	}{
		{
			name:         "unterminated parentheses",
			input:        "(x + y",
			containsText: "expected ')'",
			description:  "missing closing parenthesis in expression",
		},
		{
			name:         "missing colon",
			input:        "if x\n    pass",
			containsText: "expected ':'",
			description:  "missing colon in if statement",
		},
		{
			name:         "invalid assignment",
			input:        "42 = x",
			containsText: "unexpected token",
			description:  "invalid assignment target",
		},
		{
			name:         "incomplete expression",
			input:        "x +",
			containsText: "unexpected",
			description:  "binary operator without right operand",
		},
		{
			name:         "mismatched brackets",
			input:        "[1, 2)",
			containsText: "expected",
			description:  "mismatched opening and closing brackets",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module, errors := parseInput(t, test.input)
			validateParseResultError(t, module, errors, test.containsText)
		})
	}
}

func TestParseErrorStruct(t *testing.T) {
	tests := []struct {
		name    string
		token   lexer.Token
		message string
	}{
		{
			name: "identifier error",
			token: lexer.Token{
				Type:   lexer.Identifier,
				Lexeme: "hello",
				Span:   lexer.Span{Start: lexer.Position{Line: 1, Column: 0}, End: lexer.Position{Line: 1, Column: 5}},
			},
			message: "unexpected identifier",
		},
		{
			name: "EOF error",
			token: lexer.Token{
				Type:   lexer.EOF,
				Lexeme: "",
				Span:   lexer.Span{Start: lexer.Position{Line: 1, Column: 5}, End: lexer.Position{Line: 1, Column: 5}},
			},
			message: "unexpected end of file",
		},
		{
			name: "operator error",
			token: lexer.Token{
				Type:   lexer.Plus,
				Lexeme: "+",
				Span:   lexer.Span{Start: lexer.Position{Line: 1, Column: 2}, End: lexer.Position{Line: 1, Column: 3}},
			},
			message: "unexpected operator",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewParseError(test.token, test.message)
			validateParseErrorStruct(t, err, test.token, test.message)
		})
	}
}

func TestParseRecovery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "invalid assignment",
			input:       "42 = x",
			expectError: true,
		},
		{
			name:        "unclosed parentheses",
			input:       "(x + y",
			expectError: true,
		},
		{
			name:        "invalid function call",
			input:       "func(",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module, errors := parseInput(t, test.input)

			if test.expectError {
				validateParseResultError(t, module, errors, "")
			} else {
				validateParseSuccess(t, module, errors, -1)
			}
		})
	}
}

func TestParseComplexPrograms(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
	}{
		{
			name: "simple function",
			input: `def add(a, b):
    return a + b`,
			expectedLen: 1,
		},
		{
			name: "class with method",
			input: `class Calculator:
    def add(self, a, b):
        return a + b`,
			expectedLen: 1,
		},
		{
			name: "if-else statement",
			input: `if x > 0:
    print("positive")
else:
    print("non-positive")`,
			expectedLen: 1,
		},
		{
			name: "for loop with list",
			input: `for item in items:
    process(item)`,
			expectedLen: 1,
		},
		{
			name: "try-except block",
			input: `try:
    risky_operation()
except Exception as e:
    handle_error(e)`,
			expectedLen: 1,
		},
		{
			name: "multiple imports",
			input: `import os
import sys
from pathlib import Path`,
			expectedLen: 3,
		},
		{
			name: "mixed statements",
			input: `x = 42
y = "hello"
z = [1, 2, 3]
print(x, y, z)`,
			expectedLen: 4,
		},
		{
			name: "function with decorators",
			input: `@decorator
@another_decorator
def func():
    pass`,
			expectedLen: 1,
		},
		{
			name: "async function",
			input: `async def fetch_data():
    return await api_call()`,
			expectedLen: 1,
		},
		{
			name: "with statement",
			input: `with open('file.txt') as f:
    content = f.read()`,
			expectedLen: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module, errors := parseInput(t, test.input)
			validateParseSuccess(t, module, errors, test.expectedLen)
		})
	}
}

func TestParseModule(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty module",
			input: "",
		},
		{
			name:  "module with only newlines",
			input: "\n\n\n",
		},
		{
			name: "simple module",
			input: `x = 1
y = 2`,
		},
		{
			name: "complex module",
			input: `import os
import sys

def main():
    print("Hello, World!")

if __name__ == "__main__":
    main()`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module, errors := parseInput(t, test.input)
			validateParseSuccess(t, module, errors, -1)

			// Verify that the module has a Body field
			if module.Body == nil {
				t.Error("Expected module body but got nil")
			}

			// For empty inputs or only newlines, body should be empty
			if test.input == "" || strings.TrimSpace(test.input) == "" {
				if len(module.Body) != 0 {
					t.Errorf("Expected empty body for empty input but got %d statements", len(module.Body))
				}
			}
		})
	}
}

func TestParserErrorAccumulation(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
	}{
		{
			name:          "invalid assignment",
			input:         "42 = x",
			expectedError: true,
		},
		{
			name:          "valid input",
			input:         "x = 42",
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)

			// Check initial state
			if len(parser.Errors) != 0 {
				t.Errorf("Expected no initial errors but got %d", len(parser.Errors))
			}

			module, errors := parser.Parse()

			if test.expectedError {
				validateParseResultError(t, module, errors, "")
			} else {
				validateParseSuccess(t, module, errors, -1)
			}
		})
	}
}
