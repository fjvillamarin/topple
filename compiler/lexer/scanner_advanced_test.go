package lexer

import (
	"strings"
	"testing"
)

// Test advanced f-string features
func TestAdvancedFStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, tokens []Token)
	}{
		{
			"raw f-string",
			`rf"raw\n{x}"`,
			func(t *testing.T, tokens []Token) {
				// Should preserve \n as literal
				if tokens[1].Type != FStringMiddle || !strings.Contains(tokens[1].Lexeme, "\\n") {
					t.Errorf("Raw f-string should preserve backslash sequences")
				}
			},
		},
		{
			"triple-quoted f-string",
			`f"""multi
line {x}
string"""`,
			func(t *testing.T, tokens []Token) {
				if tokens[0].Type != FStringStart {
					t.Error("Expected FStringStart for triple-quoted f-string")
				}
				// Should handle multiline content
				foundMultiline := false
				for _, tok := range tokens {
					if tok.Type == FStringMiddle && strings.Contains(tok.Lexeme, "\n") {
						foundMultiline = true
						break
					}
				}
				if !foundMultiline {
					t.Error("Triple-quoted f-string should preserve newlines")
				}
			},
		},
		{
			"f-string with conversion",
			`f"{x!r}"`,
			func(t *testing.T, tokens []Token) {
				// Should have conversion token
				hasConversion := false
				for _, tok := range tokens {
					if tok.Type == FStringConversionStart {
						hasConversion = true
						break
					}
				}
				if !hasConversion {
					t.Error("Expected FStringConversionStart for !r conversion")
				}
			},
		},
		{
			"deeply nested f-strings",
			`f"{f'{f"{x}"}'}"`,
			func(t *testing.T, tokens []Token) {
				// Count nesting levels
				fstringStarts := 0
				for _, tok := range tokens {
					if tok.Type == FStringStart {
						fstringStarts++
					}
				}
				if fstringStarts != 3 {
					t.Errorf("Expected 3 nested f-strings, got %d", fstringStarts)
				}
			},
		},
		{
			"f-string with complex expression",
			`f"{[x for x in range(10) if x % 2 == 0]}"`,
			func(t *testing.T, tokens []Token) {
				// Should tokenize the list comprehension correctly
				hasFor := false
				hasIf := false
				for _, tok := range tokens {
					if tok.Type == For {
						hasFor = true
					}
					if tok.Type == If {
						hasIf = true
					}
				}
				if !hasFor || !hasIf {
					t.Error("F-string should correctly tokenize complex expressions")
				}
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens := scanTokens(test.input)
			test.validate(t, tokens)
		})
	}
}

// Test match statement patterns
func TestMatchPatterns(t *testing.T) {
	input := `
match point:
    case (0, 0):
        return "Origin"
    case (0, y):
        return f"On Y-axis at {y}"
    case (x, 0):
        return f"On X-axis at {x}"
    case Point(x=0, y=y):
        return f"Complex pattern"
    case [*elements]:
        return "Sequence pattern"
    case _:
        return "Default"
`
	
	tokens := scanTokens(input)
	
	// Count case keywords
	caseCount := 0
	for _, tok := range tokens {
		if tok.Type == Case {
			caseCount++
		}
	}
	
	if caseCount != 6 {
		t.Errorf("Expected 6 case statements, got %d", caseCount)
	}
}

// Test type hints and annotations
func TestTypeAnnotations(t *testing.T) {
	input := `
def greet(name: str, age: int = 0) -> str:
    return f"Hello {name}"

x: int = 42
y: list[int] = [1, 2, 3]
z: dict[str, Any] = {}
w: tuple[int, ...] = (1, 2, 3)
`
	
	tokens := scanTokens(input)
	
	// Check for proper tokenization of type hints
	colonCount := 0
	arrowCount := 0
	for _, tok := range tokens {
		if tok.Type == Colon {
			colonCount++
		}
		if tok.Type == Arrow {
			arrowCount++
		}
	}
	
	if colonCount < 5 { // At least 5 type annotations
		t.Errorf("Expected at least 5 colons for type annotations, got %d", colonCount)
	}
	if arrowCount != 1 {
		t.Errorf("Expected 1 arrow for return type, got %d", arrowCount)
	}
}

// Test async/await syntax
func TestAsyncAwait(t *testing.T) {
	input := `
async def fetch_data():
    result = await get_response()
    async with session:
        data = await session.read()
    async for item in stream:
        process(item)
`
	
	tokens := scanTokens(input)
	
	asyncCount := 0
	awaitCount := 0
	for _, tok := range tokens {
		if tok.Type == Async {
			asyncCount++
		}
		if tok.Type == Await {
			awaitCount++
		}
	}
	
	if asyncCount != 3 { // async def, async with, async for
		t.Errorf("Expected 3 async keywords, got %d", asyncCount)
	}
	if awaitCount != 2 {
		t.Errorf("Expected 2 await keywords, got %d", awaitCount)
	}
}

// Test decorators
func TestDecorators(t *testing.T) {
	input := `
@property
@deprecated("Use new_method instead")
@functools.lru_cache(maxsize=128)
def my_method(self):
    pass
`
	
	tokens := scanTokens(input)
	
	atCount := 0
	for _, tok := range tokens {
		if tok.Type == At {
			atCount++
		}
	}
	
	if atCount != 3 {
		t.Errorf("Expected 3 @ symbols for decorators, got %d", atCount)
	}
}

// Test walrus operator
func TestWalrusOperator(t *testing.T) {
	input := `
if (n := len(data)) > 10:
    print(f"List is too long ({n} elements)")
while (line := file.readline()):
    process(line)
`
	
	tokens := scanTokens(input)
	
	walrusCount := 0
	for _, tok := range tokens {
		if tok.Type == Walrus {
			walrusCount++
		}
	}
	
	if walrusCount != 2 {
		t.Errorf("Expected 2 walrus operators, got %d", walrusCount)
	}
}

// Test complex indentation scenarios
func TestComplexIndentation(t *testing.T) {
	input := `
def outer():
    def inner():
        if True:
            while x:
                for i in range(10):
                    try:
                        pass
                    except:
                        pass
                    finally:
                        pass
    return inner
`
	
	scanner := NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	
	// Track maximum indentation depth
	maxDepth := 0
	currentDepth := 0
	for _, tok := range tokens {
		switch tok.Type {
		case Indent:
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		case Dedent:
			currentDepth--
		}
	}
	
	if maxDepth != 6 {
		t.Errorf("Expected maximum indentation depth of 6, got %d", maxDepth)
	}
	
	if currentDepth != 0 {
		t.Errorf("Indentation not balanced, final depth: %d", currentDepth)
	}
}

// Test error recovery
func TestErrorRecovery(t *testing.T) {
	// Scanner should continue after encountering errors
	input := `
x = 1
y = "unterminated string
z = 3
w = 'another unterminated
a = 5
`
	
	scanner := NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	
	// Should have multiple errors
	if len(scanner.Errors) < 2 {
		t.Errorf("Expected at least 2 errors, got %d", len(scanner.Errors))
	}
	
	// Should still tokenize valid parts
	identCount := 0
	numberCount := 0
	for _, tok := range tokens {
		if tok.Type == Identifier {
			identCount++
		}
		if tok.Type == Number {
			numberCount++
		}
	}
	
	if identCount < 3 { // At least x, z, a
		t.Errorf("Expected at least 3 identifiers after error recovery, got %d", identCount)
	}
	if numberCount < 3 { // At least 1, 3, 5
		t.Errorf("Expected at least 3 numbers after error recovery, got %d", numberCount)
	}
}

// Test HTML/View mode switching
func TestViewModeContextSwitching(t *testing.T) {
	input := `
@view
def MyComponent(name):
    <div class="container">
        <h1>Hello {name}</h1>
        {if logged_in:
            <span>Welcome back!</span>
        }
    </div>
    return None
`
	
	scanner := NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	
	// Should have HTML tokens
	hasHTMLTokens := false
	for _, tok := range tokens {
		if tok.Type == TagOpen || tok.Type == TagClose {
			hasHTMLTokens = true
			break
		}
	}
	
	if !hasHTMLTokens {
		t.Error("Expected HTML tokens in view function")
	}
}

// Test numbers with different bases and formats
func TestNumberFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		valid    bool
	}{
		// Valid numbers
		{"0b1111", 15, true},
		{"0o777", 511, true},
		{"0xDEADBEEF", 3735928559, true},
		{"1.23e-4", 0.000123, true},
		
		// Invalid numbers (should produce errors)
		{"0b2", 0, false},      // Invalid binary digit
		{"0o8", 0, false},      // Invalid octal digit  
		{"0xGHI", 0, false},    // Invalid hex digit
		{"1.2.3", 0, false},    // Multiple decimal points
		{"1ee3", 0, false},     // Invalid exponent
	}
	
	for _, test := range tests {
		scanner := NewScanner([]byte(test.input))
		tokens := scanner.ScanTokens()
		
		if test.valid {
			if len(scanner.Errors) > 0 {
				t.Errorf("Expected no errors for %s, got %v", test.input, scanner.Errors)
			}
			if len(tokens) >= 1 && tokens[0].Type == Number {
				if val, ok := tokens[0].Literal.(float64); ok {
					if val != test.expected {
						t.Errorf("Expected %f for %s, got %f", test.expected, test.input, val)
					}
				}
			}
		} else {
			if len(scanner.Errors) == 0 {
				t.Errorf("Expected error for invalid number %s", test.input)
			}
		}
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"only whitespace", "   \n\t  \n"},
		{"only comment", "# just a comment"},
		{"comment with no newline", "# comment"},
		{"nested parentheses", "(((())))"},
		{"mixed brackets", "([{[({})]}])"},
		{"lone backslash", "\\"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			
			// Should always end with EOF
			if len(tokens) == 0 || tokens[len(tokens)-1].Type != EOF {
				t.Error("Scanner should always produce EOF token")
			}
		})
	}
}