package lexer

import (
	"testing"
)

// Helper function to create a scanner and scan tokens
func scanTokens(input string) []Token {
	scanner := NewScanner([]byte(input))
	return scanner.ScanTokens()
}

// Helper function to check if tokens match expected types
func assertTokenTypes(t *testing.T, tokens []Token, expected []TokenType) {
	t.Helper()
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %s, got %s", i, expected[i], tok.Type)
		}
	}
}

// Helper function to check token with value
func assertToken(t *testing.T, tok Token, expectedType TokenType, expectedLexeme string) {
	t.Helper()
	if tok.Type != expectedType {
		t.Errorf("Expected token type %s, got %s", expectedType, tok.Type)
	}
	if tok.Lexeme != expectedLexeme {
		t.Errorf("Expected lexeme %q, got %q", expectedLexeme, tok.Lexeme)
	}
}

// Test basic single-character punctuation tokens
func TestSingleCharacterTokens(t *testing.T) {
	input := "()[]{},:;+-*/%|&^~@."
	tokens := scanTokens(input)
	
	expected := []TokenType{
		LeftParen, RightParen,
		LeftBracket, RightBracket,
		LeftBrace, RightBrace,
		Comma, Colon, Semicolon,
		Plus, Minus, Star, Slash, Percent,
		Pipe, Ampersand, Caret, Tilde, At, Dot,
		EOF,
	}
	
	assertTokenTypes(t, tokens, expected)
}

// Test multi-character operators
func TestMultiCharacterOperators(t *testing.T) {
	input := "+= -= *= /= %= |= &= ^= @= //= **= <<= >>= != == <= >= := -> // ** << >> ="
	tokens := scanTokens(input)
	
	expected := []TokenType{
		PlusEqual, MinusEqual, StarEqual, SlashEqual, PercentEqual,
		PipeEqual, AmpEqual, CaretEqual, AtEqual,
		SlashSlashEqual, StarStarEqual, LessLessEqual, GreaterGreaterEqual,
		BangEqual, EqualEqual, LessEqual, GreaterEqual,
		Walrus, Arrow, SlashSlash, StarStar, LessLess, GreaterGreater, Equal,
		EOF,
	}
	
	assertTokenTypes(t, tokens, expected)
}

// Test keywords
func TestKeywords(t *testing.T) {
	input := "and as assert async await break class continue def del elif else except False finally for from global if import in is lambda match None nonlocal not or pass raise return True try while with yield case type view component"
	tokens := scanTokens(input)
	
	expected := []TokenType{
		And, As, Assert, Async, Await, Break, Class, Continue, Def, Del,
		Elif, Else, Except, False, Finally, For, From, Global, If, Import,
		In, Is, Lambda, Match, None, Nonlocal, Not, Or, Pass, Raise,
		Return, True, Try, While, With, Yield, Case, Type, View, Component,
		EOF,
	}
	
	assertTokenTypes(t, tokens, expected)
}

// Test identifiers
func TestIdentifiers(t *testing.T) {
	input := "x _private __dunder__ CamelCase snake_case CONSTANT_NAME café 变量"
	tokens := scanTokens(input)
	
	for i := 0; i < len(tokens)-1; i++ { // -1 to skip EOF
		if tokens[i].Type != Identifier {
			t.Errorf("Token %d: expected Identifier, got %s", i, tokens[i].Type)
		}
	}
	
	// Check specific identifiers
	assertToken(t, tokens[0], Identifier, "x")
	assertToken(t, tokens[1], Identifier, "_private")
	assertToken(t, tokens[2], Identifier, "__dunder__")
	assertToken(t, tokens[3], Identifier, "CamelCase")
	assertToken(t, tokens[4], Identifier, "snake_case")
	assertToken(t, tokens[5], Identifier, "CONSTANT_NAME")
	assertToken(t, tokens[6], Identifier, "café")
	assertToken(t, tokens[7], Identifier, "变量")
}

// Test numbers
func TestNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"42", int64(42)},
		{"3.14", 3.14},
		{"0b1010", int64(10)},
		{"0o12", int64(10)},
		{"0xA", int64(10)},
		{"1e10", 1e10},
		{"3.14e-2", 3.14e-2},
	}
	
	for _, test := range tests {
		tokens := scanTokens(test.input)
		if len(tokens) != 2 || tokens[0].Type != Number {
			t.Errorf("Expected Number token for %s", test.input)
			continue
		}
		
		// Handle both int64 and float64 literals
		switch expected := test.expected.(type) {
		case int64:
			if val, ok := tokens[0].Literal.(int64); !ok {
				t.Errorf("Expected int64 literal for %s, got %T", test.input, tokens[0].Literal)
			} else if val != expected {
				t.Errorf("Expected %d for %s, got %d", expected, test.input, val)
			}
		case float64:
			if val, ok := tokens[0].Literal.(float64); !ok {
				t.Errorf("Expected float64 literal for %s, got %T", test.input, tokens[0].Literal)
			} else if val != expected {
				t.Errorf("Expected %g for %s, got %g", expected, test.input, val)
			}
		}
	}
}

// Test strings
func TestStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"hello\nworld"`, `hello\nworld`},
		{`"hello\tworld"`, `hello\tworld`},
		{`"hello\\world"`, `hello\\world`},
		{`"hello\"world"`, `hello\"world`},
		{`'''triple
single'''`, "triple\nsingle"},
		{`"""triple
double"""`, "triple\ndouble"},
		{`r"raw\nstring"`, `"raw\nstring`},
	}
	
	for _, test := range tests {
		tokens := scanTokens(test.input)
		if len(tokens) != 2 || tokens[0].Type != String {
			t.Errorf("Expected String token for %s, got %v", test.input, tokens)
			continue
		}
		if val, ok := tokens[0].Literal.(string); !ok {
			t.Errorf("Expected string literal for %s, got %T: %v", test.input, tokens[0].Literal, tokens[0].Literal)
		} else if val != test.expected {
			t.Errorf("Expected %q for %s, got %q", test.expected, test.input, val)
		}
	}
}

// Test indentation and dedentation
func TestIndentation(t *testing.T) {
	input := `def foo():
    x = 1
    if True:
        y = 2
    z = 3
`
	tokens := scanTokens(input)
	
	// Find all INDENT and DEDENT tokens
	var indents, dedents int
	for _, tok := range tokens {
		switch tok.Type {
		case Indent:
			indents++
		case Dedent:
			dedents++
		}
	}
	
	if indents != 2 {
		t.Errorf("Expected 2 INDENT tokens, got %d", indents)
	}
	if dedents != 2 {
		t.Errorf("Expected 2 DEDENT tokens, got %d", dedents)
	}
}

// Test composite tokens
func TestCompositeTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{"x is not None", []TokenType{Identifier, IsNot, None, EOF}},
		{"y not in list", []TokenType{Identifier, NotIn, Identifier, EOF}},
		{"a is b", []TokenType{Identifier, Is, Identifier, EOF}},
		{"c not d", []TokenType{Identifier, Not, Identifier, EOF}},
	}
	
	for _, test := range tests {
		tokens := scanTokens(test.input)
		assertTokenTypes(t, tokens, test.expected)
	}
}

// Test f-strings
func TestFStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			"simple f-string",
			`f"hello {name}"`,
			[]TokenType{FStringStart, FStringMiddle, LeftBraceF, Identifier, RightBraceF, FStringEnd, EOF},
		},
		{
			"f-string with expression",
			`f"result: {x + y}"`,
			[]TokenType{FStringStart, FStringMiddle, LeftBraceF, Identifier, Plus, Identifier, RightBraceF, FStringEnd, EOF},
		},
		{
			"f-string with format spec",
			`f"{value:.2f}"`,
			[]TokenType{FStringStart, LeftBraceF, Identifier, Colon, FStringMiddle, RightBraceF, FStringEnd, EOF},
		},
		{
			"f-string with debug",
			`f"{x=}"`,
			[]TokenType{FStringStart, LeftBraceF, Identifier, FStringEqual, RightBraceF, FStringEnd, EOF},
		},
		{
			"nested f-string",
			`f"outer {f'inner {x}'}"`,
			[]TokenType{
				FStringStart, FStringMiddle, LeftBraceF,
				FStringStart, FStringMiddle, LeftBraceF, Identifier, RightBraceF, FStringEnd,
				RightBraceF, FStringEnd, EOF,
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := scanTokens(test.input)
			assertTokenTypes(t, tokens, test.expected)
		})
	}
}

// Test error handling
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
		hasError bool
	}{
		{"unterminated string", `"hello`, true},
		{"invalid character", "x = 1 $ y", true},
		{"mixed tabs and spaces", "if True:\n\tx = 1\n    y = 2", true},
		{"valid code", "x = 1 + 2", false},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := NewScanner([]byte(test.input))
			scanner.ScanTokens()
			
			if test.hasError && len(scanner.Errors) == 0 {
				t.Error("Expected error but got none")
			}
			if !test.hasError && len(scanner.Errors) > 0 {
				t.Errorf("Expected no error but got: %v", scanner.Errors)
			}
		})
	}
}

// Test line continuations
func TestLineContinuations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasNewline bool
	}{
		{
			"implicit continuation in parentheses",
			"x = (1 +\n     2)",
			false,
		},
		{
			"no continuation",
			"x = 1\ny = 2",
			true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := scanTokens(test.input)
			hasNewline := false
			for _, tok := range tokens {
				if tok.Type == Newline {
					hasNewline = true
					break
				}
			}
			if hasNewline != test.hasNewline {
				t.Errorf("Expected hasNewline=%v, got %v", test.hasNewline, hasNewline)
			}
		})
	}
}

// Test ellipsis
func TestEllipsis(t *testing.T) {
	input := "x = ... # ellipsis"
	tokens := scanTokens(input)
	
	expected := []TokenType{Identifier, Equal, Ellipsis, EOF}
	assertTokenTypes(t, tokens, expected)
}

// Test position tracking
func TestPositionTracking(t *testing.T) {
	input := "x = 1\ny = 2"
	tokens := scanTokens(input)
	
	// Check first token (x) position
	if tokens[0].Start().Line != 1 || tokens[0].Start().Column != 1 {
		t.Errorf("Expected token 'x' at line 1, column 1, got %v", tokens[0].Start())
	}
	
	// Check token on second line (y)
	yToken := tokens[4] // After x, =, 1, newline
	if yToken.Start().Line != 2 || yToken.Start().Column != 1 {
		t.Errorf("Expected token 'y' at line 2, column 1, got %v", yToken.Start())
	}
}

// Test HTML/View mode (basic)
func TestHTMLMode(t *testing.T) {
	// This tests basic HTML tokenization in view context
	scanner := NewScanner([]byte("<div>"))
	scanner.ctx.viewDepth = 1 // Simulate being inside a view
	tokens := scanner.ScanTokens()
	
	// Should produce: TagOpen, Identifier (div), TagClose, EOF
	expected := []TokenType{TagOpen, Identifier, TagClose, EOF}
	assertTokenTypes(t, tokens, expected)
}

// Test Unicode handling
func TestUnicodeHandling(t *testing.T) {
	input := `"你好" + "🌍" # Unicode comment 中文`
	tokens := scanTokens(input)
	
	// Check that Unicode strings are properly tokenized
	if tokens[0].Type != String || tokens[0].Literal != "你好" {
		t.Errorf("Expected Chinese string, got %v", tokens[0])
	}
	if tokens[2].Type != String || tokens[2].Literal != "🌍" {
		t.Errorf("Expected emoji string, got %v", tokens[2])
	}
}

// Benchmark tokenization
func BenchmarkScanner(b *testing.B) {
	// Sample Python code for benchmarking
	input := `
def fibonacci(n):
    """Calculate fibonacci number"""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Test the function
for i in range(10):
    print(f"fib({i}) = {fibonacci(i)}")
`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner([]byte(input))
		scanner.ScanTokens()
	}
}