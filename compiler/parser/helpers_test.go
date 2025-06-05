package parser

import (
	"sylfie/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to create a parser from input string
func createParser(t *testing.T, input string) *Parser {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	return NewParser(tokens)
}

// Helper function to validate parser token position
func validateParserPosition(t *testing.T, parser *Parser, expectedPosition int, operation string) {
	if parser.Current != expectedPosition {
		t.Errorf("%s: Expected parser position %d, got %d", operation, expectedPosition, parser.Current)
	}
}

// Helper function to validate token type and advance behavior
func validateTokenAndAdvance(t *testing.T, parser *Parser, expectedType lexer.TokenType, shouldAdvance bool, operation string) lexer.Token {
	initialPos := parser.Current
	var token lexer.Token
	
	switch operation {
	case "peek":
		token = parser.peek()
	case "advance":
		token = parser.advance()
	case "previous":
		token = parser.previous()
	}
	
	if token.Type != expectedType {
		t.Errorf("%s: Expected token type %s, got %s", operation, expectedType, token.Type)
	}
	
	expectedPos := initialPos
	if shouldAdvance {
		expectedPos++
	}
	
	validateParserPosition(t, parser, expectedPos, operation)
	return token
}

// Helper function to test error creation and validation
func validateParseError(t *testing.T, err error, expectedMessage string, expectedTokenType lexer.TokenType) {
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	
	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	
	if parseErr.Message != expectedMessage {
		t.Errorf("Expected error message %q, got %q", expectedMessage, parseErr.Message)
	}
	
	if parseErr.Token.Type != expectedTokenType {
		t.Errorf("Expected error token type %s, got %s", expectedTokenType, parseErr.Token.Type)
	}
}

// Test parser helper methods comprehensively
func TestParserHelpers(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "basic token operations",
			input:       "hello world 42",
			description: "test basic parser operations with multiple tokens",
		},
		{
			name:        "single token",
			input:       "identifier",
			description: "test operations with single token",
		},
		{
			name:        "empty input",
			input:       "",
			description: "test operations with empty input",
		},
		{
			name:        "operators and punctuation",
			input:       "+ - * / ( ) { }",
			description: "test operations with various operators",
		},
		{
			name:        "keywords and literals",
			input:       "def if else True False None",
			description: "test operations with keywords and literals",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			
			// Test initial state
			if test.input == "" {
				if !parser.isAtEnd() {
					t.Error("Parser should be at end for empty input")
				}
				return
			}
			
			if parser.isAtEnd() {
				t.Error("Parser should not be at end for non-empty input")
			}
			
			// Test peek without advancing
			initialPos := parser.Current
			firstToken := parser.peek()
			validateParserPosition(t, parser, initialPos, "peek")
			
			// Test advance
			advancedToken := parser.advance()
			if advancedToken.Type != firstToken.Type {
				t.Errorf("Advanced token should match peeked token: expected %s, got %s", 
					firstToken.Type, advancedToken.Type)
			}
			validateParserPosition(t, parser, initialPos+1, "advance")
			
			// Test previous
			if parser.Current > 0 {
				previousToken := parser.previous()
				if previousToken.Type != advancedToken.Type {
					t.Errorf("Previous token should match advanced token: expected %s, got %s", 
						advancedToken.Type, previousToken.Type)
				}
			}
		})
	}
}

// Test consume method with various scenarios
func TestConsume(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		tokenType     lexer.TokenType
		message       string
		hasError      bool
		errorContains string
		description   string
	}{
		{
			name:        "consume expected identifier",
			input:       "hello",
			tokenType:   lexer.Identifier,
			message:     "expected identifier",
			description: "consume matching token type",
		},
		{
			name:        "consume expected number",
			input:       "42",
			tokenType:   lexer.Number,
			message:     "expected number",
			description: "consume numeric token",
		},
		{
			name:        "consume expected string",
			input:       `"hello world"`,
			tokenType:   lexer.String,
			message:     "expected string",
			description: "consume string literal",
		},
		{
			name:        "consume expected operator",
			input:       "+",
			tokenType:   lexer.Plus,
			message:     "expected plus",
			description: "consume operator token",
		},
		{
			name:        "consume expected punctuation",
			input:       "(",
			tokenType:   lexer.LeftParen,
			message:     "expected left paren",
			description: "consume punctuation token",
		},
		
		// Error cases
		{
			name:          "wrong token type",
			input:         "hello",
			tokenType:     lexer.Number,
			message:       "expected number",
			hasError:      true,
			errorContains: "expected number",
			description:   "attempt to consume wrong token type",
		},
		{
			name:          "consume at end",
			input:         "",
			tokenType:     lexer.Identifier,
			message:       "expected identifier",
			hasError:      true,
			errorContains: "expected identifier",
			description:   "attempt to consume when at end",
		},
		{
			name:          "consume unexpected token",
			input:         "hello world",
			tokenType:     lexer.Number,
			message:       "expected number",
			hasError:      true,
			errorContains: "expected number",
			description:   "consume unexpected token in sequence",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			token, err := parser.consume(test.tokenType, test.message)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if token.Type != test.tokenType {
				t.Errorf("Expected consumed token type %s, got %s", test.tokenType, token.Type)
			}
		})
	}
}

// Test match method with multiple token types
func TestMatch(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		tokenTypes   []lexer.TokenType
		shouldMatch  bool
		shouldAdvance bool
		description  string
	}{
		{
			name:         "match single type",
			input:        "hello",
			tokenTypes:   []lexer.TokenType{lexer.Identifier},
			shouldMatch:  true,
			shouldAdvance: true,
			description:  "match single expected token type",
		},
		{
			name:         "match multiple types",
			input:        "42",
			tokenTypes:   []lexer.TokenType{lexer.Identifier, lexer.Number, lexer.String},
			shouldMatch:  true,
			shouldAdvance: true,
			description:  "match from multiple possible types",
		},
		{
			name:         "match operators",
			input:        "*",
			tokenTypes:   []lexer.TokenType{lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash},
			shouldMatch:  true,
			shouldAdvance: true,
			description:  "match operator from set",
		},
		{
			name:         "match punctuation",
			input:        "{",
			tokenTypes:   []lexer.TokenType{lexer.LeftParen, lexer.LeftBrace, lexer.LeftBracket},
			shouldMatch:  true,
			shouldAdvance: true,
			description:  "match punctuation from set",
		},
		{
			name:         "no match single type",
			input:        "hello",
			tokenTypes:   []lexer.TokenType{lexer.Number},
			shouldMatch:  false,
			shouldAdvance: false,
			description:  "no match for single wrong type",
		},
		{
			name:         "no match multiple types",
			input:        "hello",
			tokenTypes:   []lexer.TokenType{lexer.Number, lexer.String, lexer.Plus},
			shouldMatch:  false,
			shouldAdvance: false,
			description:  "no match for multiple wrong types",
		},
		{
			name:         "no match at end",
			input:        "",
			tokenTypes:   []lexer.TokenType{lexer.Identifier, lexer.Number},
			shouldMatch:  false,
			shouldAdvance: false,
			description:  "no match when at end of input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			initialPos := parser.Current
			result := parser.match(test.tokenTypes...)

			if result != test.shouldMatch {
				t.Errorf("Expected match result %v for %s, got %v", test.shouldMatch, test.description, result)
			}

			expectedPos := initialPos
			if test.shouldAdvance {
				expectedPos++
			}
			validateParserPosition(t, parser, expectedPos, "match")
		})
	}
}

// Test check and checkNext methods
func TestCheckMethods(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		checkType   lexer.TokenType
		checkNext   bool
		expected    bool
		description string
	}{
		// check() tests
		{
			name:        "check current identifier",
			input:       "hello world",
			checkType:   lexer.Identifier,
			expected:    true,
			description: "check current token is identifier",
		},
		{
			name:        "check current number",
			input:       "42 hello",
			checkType:   lexer.Number,
			expected:    true,
			description: "check current token is number",
		},
		{
			name:        "check current wrong type",
			input:       "hello world",
			checkType:   lexer.Number,
			expected:    false,
			description: "check current token is wrong type",
		},
		{
			name:        "check at end",
			input:       "",
			checkType:   lexer.Identifier,
			expected:    false,
			description: "check when at end of input",
		},

		// checkNext() tests
		{
			name:        "check next identifier",
			input:       "hello world",
			checkType:   lexer.Identifier,
			checkNext:   true,
			expected:    true,
			description: "check next token is identifier",
		},
		{
			name:        "check next number",
			input:       "hello 42",
			checkType:   lexer.Number,
			checkNext:   true,
			expected:    true,
			description: "check next token is number",
		},
		{
			name:        "check next wrong type",
			input:       "hello world",
			checkType:   lexer.Number,
			checkNext:   true,
			expected:    false,
			description: "check next token is wrong type",
		},
		{
			name:        "check next at end",
			input:       "hello",
			checkType:   lexer.Identifier,
			checkNext:   true,
			expected:    false,
			description: "check next when only one token",
		},
		{
			name:        "check next EOF",
			input:       "hello",
			checkType:   lexer.EOF,
			checkNext:   true,
			expected:    true,
			description: "check next is EOF",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			initialPos := parser.Current
			
			var result bool
			if test.checkNext {
				result = parser.checkNext(test.checkType)
			} else {
				result = parser.check(test.checkType)
			}

			if result != test.expected {
				t.Errorf("Expected %v for %s, got %v", test.expected, test.description, result)
			}

			// Both check methods should not advance position
			validateParserPosition(t, parser, initialPos, "check methods")
		})
	}
}

// Test error creation and formatting
func TestErrorCreation(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		message      string
		expectToken  lexer.TokenType
		description  string
	}{
		{
			name:        "error with identifier",
			input:       "hello",
			message:     "unexpected identifier",
			expectToken: lexer.Identifier,
			description: "create error for identifier token",
		},
		{
			name:        "error with number",
			input:       "42",
			message:     "unexpected number",
			expectToken: lexer.Number,
			description: "create error for number token",
		},
		{
			name:        "error with operator",
			input:       "+",
			message:     "unexpected operator",
			expectToken: lexer.Plus,
			description: "create error for operator token",
		},
		{
			name:        "error at end",
			input:       "",
			message:     "unexpected end of input",
			expectToken: lexer.EOF,
			description: "create error at end of input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			token := parser.peek()
			err := parser.error(token, test.message)

			validateParseError(t, err, test.message, test.expectToken)
		})
	}
}

// Test navigation and positioning methods
func TestParserNavigation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		operations  []string
		description string
	}{
		{
			name:        "basic navigation",
			input:       "a b c d",
			operations:  []string{"advance", "advance", "previous", "peek"},
			description: "test basic forward and backward navigation",
		},
		{
			name:        "peek operations",
			input:       "x + y - z",
			operations:  []string{"peek", "peek", "advance", "peek"},
			description: "test multiple peek operations",
		},
		{
			name:        "advance to end",
			input:       "hello world",
			operations:  []string{"advance", "advance", "isAtEnd"},
			description: "test advancing to end of input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			
			for i, operation := range test.operations {
				switch operation {
				case "advance":
					if !parser.isAtEnd() {
						parser.advance()
					}
				case "peek":
					parser.peek()
				case "previous":
					if parser.Current > 0 {
						parser.previous()
					}
				case "isAtEnd":
					if !parser.isAtEnd() {
						t.Errorf("Step %d: Expected to be at end but wasn't", i)
					}
				}
			}
		})
	}
}

// Test integration of multiple helper methods
func TestHelperMethodIntegration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "expression parsing simulation",
			input:       "x + y * 2",
			description: "simulate parsing expression with helpers",
		},
		{
			name:        "function call simulation",
			input:       "func(a, b, c)",
			description: "simulate parsing function call with helpers",
		},
		{
			name:        "nested structure simulation",
			input:       "obj.method()[0].attr",
			description: "simulate parsing nested structure with helpers",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := createParser(t, test.input)
			
			// Simulate parsing by using various helper methods
			tokenCount := 0
			for !parser.isAtEnd() {
				// Peek at current token
				current := parser.peek()
				if current.Type == lexer.EOF {
					break
				}
				
				// Check if it matches expected types
				if parser.check(current.Type) {
					// Advance past the token
					advanced := parser.advance()
					if advanced.Type != current.Type {
						t.Errorf("Token type mismatch: expected %s, got %s", current.Type, advanced.Type)
					}
					tokenCount++
					
					// Test previous() works
					if parser.Current > 0 {
						prev := parser.previous()
						if prev.Type != advanced.Type {
							t.Errorf("Previous token mismatch: expected %s, got %s", advanced.Type, prev.Type)
						}
					}
				} else {
					t.Errorf("Check failed for token type %s", current.Type)
				}
			}

			if tokenCount == 0 {
				t.Errorf("No tokens were processed for %s", test.description)
			}

			// Verify we're at the end
			if !parser.isAtEnd() {
				t.Errorf("Should be at end after processing all tokens for %s", test.description)
			}
		})
	}
}