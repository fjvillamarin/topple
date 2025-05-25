package lexer

// scanner.go
//
// A hand-written scanner for Python 3.12, following the style of
// "Crafting Interpreters" but implemented in Go.
//
// It produces a flat slice of Token values; the parser drives
// consumption.  Any diagnostics are placed in the public Errors slice.

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// ── configuration helper ──────────────────────────────────────────────

type ScannerConfig struct {
	StartLine   int // usually 1
	StartColumn int // usually 1
}

func DefaultScannerConfig() ScannerConfig {
	return ScannerConfig{StartLine: 1, StartColumn: 1}
}

// fstringContext tracks the state of f-string parsing
type fstringContext struct {
	quote        rune // The quote character (' or ")
	braceDepth   int  // Depth of nested braces in expressions
	inExpression bool // Whether we're currently parsing an expression inside {}
	inFormatSpec bool // Whether we're currently parsing a format specification
}

// ── scanner object ───────────────────────────────────────────────────

type Scanner struct {
	src        []byte
	start, cur int // byte offsets into src
	line, col  int // current location (1-based)
	// location of *start* of current lexeme:
	lexLine, lexCol int

	tokens []Token
	Errors []error

	indentStack []int // stack[0] == 0  (invariant)
	parenDepth  int   // (),[],{} nesting ⇒ lines may continue
	cfg         ScannerConfig

	// F-string state management
	fstringStack []fstringContext // Stack of f-string contexts for nested f-strings
}

// NewScanner returns a default-configured scanner.
func NewScanner(src []byte) *Scanner {
	return NewScannerWithConfig(src, DefaultScannerConfig())
}

func NewScannerWithConfig(src []byte, cfg ScannerConfig) *Scanner {
	sc := &Scanner{
		src:         src,
		line:        cfg.StartLine,
		col:         cfg.StartColumn,
		lexLine:     cfg.StartLine,
		lexCol:      cfg.StartColumn,
		cfg:         cfg,
		indentStack: []int{0}, // invariant bottom = 0
	}
	return sc
}

// ── public entrypoint ────────────────────────────────────────────────

func (s *Scanner) ScanTokens() []Token {
	for !s.atEnd() {
		s.lexLine, s.lexCol = s.line, s.col
		s.start = s.cur
		s.scanToken()
	}

	// flush pending dedents (PEP Tokenizer rule 3)
	for len(s.indentStack) > 1 {
		s.indentStack = s.indentStack[:len(s.indentStack)-1]
		s.addToken(Dedent)
	}

	s.tokens = append(s.tokens, Token{
		Type: EOF,
		Span: Span{
			Start: Position{Line: s.line, Column: s.col},
			End:   Position{Line: s.line, Column: s.col},
		},
	})

	// Post-process tokens to detect composite tokens
	s.processCompositeTokens()

	return s.tokens
}

// Process tokens to detect and generate composite tokens like "is not" and "not in"
func (s *Scanner) processCompositeTokens() {
	if len(s.tokens) < 2 {
		return
	}

	processed := make([]Token, 0, len(s.tokens))
	i := 0

	for i < len(s.tokens) {
		// Check for "is not" sequence
		if i+1 < len(s.tokens) &&
			s.tokens[i].Type == Is && s.tokens[i+1].Type == Not {
			// Create a composite "is not" token
			isNotToken := Token{
				Type:   IsNot,
				Lexeme: "is not",
				Span: Span{
					Start: s.tokens[i].Start(),
					End:   s.tokens[i+1].End(),
				},
			}
			processed = append(processed, isNotToken)
			i += 2 // Skip both tokens
			continue
		}

		// Check for "not in" sequence
		if i+1 < len(s.tokens) &&
			s.tokens[i].Type == Not && s.tokens[i+1].Type == In {
			// Create a composite "not in" token
			notInToken := Token{
				Type:   NotIn,
				Lexeme: "not in",
				Span: Span{
					Start: s.tokens[i].Start(),
					End:   s.tokens[i+1].End(),
				},
			}
			processed = append(processed, notInToken)
			i += 2 // Skip both tokens
			continue
		}

		// Regular token, just add it
		processed = append(processed, s.tokens[i])
		i++
	}

	s.tokens = processed
}

// ── low-level helpers ────────────────────────────────────────────────

func (s *Scanner) atEnd() bool { return s.cur >= len(s.src) }

func (s *Scanner) peek() rune {
	if s.atEnd() {
		return -1
	}
	r, _ := utf8.DecodeRune(s.src[s.cur:])
	return r
}

func (s *Scanner) peekN(n int) rune {
	i := s.cur
	for n > 0 && i < len(s.src) {
		_, size := utf8.DecodeRune(s.src[i:])
		i += size
		n--
	}
	if i >= len(s.src) {
		return -1
	}
	r, _ := utf8.DecodeRune(s.src[i:])
	return r
}

func (s *Scanner) advance() rune {
	if s.atEnd() {
		return -1
	}
	r, size := utf8.DecodeRune(s.src[s.cur:])
	s.cur += size
	s.col += 1 // count characters not bytes for columns
	if r == '\n' {
		s.line++
		s.col = s.cfg.StartColumn
	}
	return r
}

func (s *Scanner) match(expect rune) bool {
	if s.peek() != expect {
		return false
	}
	s.advance()
	return true
}

func (s *Scanner) addToken(tt TokenType) {
	s.tokens = append(s.tokens, Token{
		Type: tt,
		// lexeme is *bytes* slice – OK even for UTF-8, we store raw input:
		Lexeme: string(s.src[s.start:s.cur]),
		Span: Span{
			Start: Position{Line: s.lexLine, Column: s.lexCol},
			End:   Position{Line: s.line, Column: s.col},
		},
	})
}

func (s *Scanner) addTokenLit(tt TokenType, lit any) {
	s.tokens = append(s.tokens, Token{
		Type:    tt,
		Lexeme:  string(s.src[s.start:s.cur]),
		Literal: lit,
		Span: Span{
			Start: Position{Line: s.lexLine, Column: s.lexCol},
			End:   Position{Line: s.line, Column: s.col},
		},
	})
}

func (s *Scanner) errorf(format string, args ...any) {
	s.Errors = append(s.Errors, NewScannerError(fmt.Sprintf(format, args...), s.lexLine, s.lexCol))
}

// ── main dispatcher ─────────────────────────────────────────────────

func (s *Scanner) scanToken() {
	// Check if we're inside an f-string and should continue f-string scanning
	if len(s.fstringStack) > 0 {
		s.scanFStringContent()
		return
	}

	switch r := s.advance(); r {
	// ── single char punctuation ──
	case '(':
		s.parenDepth++
		s.addToken(LeftParen)
	case ')':
		if s.parenDepth > 0 {
			s.parenDepth--
		}
		s.addToken(RightParen)
	case '[':
		s.parenDepth++
		s.addToken(LeftBracket)
	case ']':
		if s.parenDepth > 0 {
			s.parenDepth--
		}
		s.addToken(RightBracket)
	case '{':
		s.parenDepth++
		s.addToken(LeftBrace)
	case '}':
		if s.parenDepth > 0 {
			s.parenDepth--
		}
		s.addToken(RightBrace)
	case ',':
		s.addToken(Comma)
	case ':':
		if s.match('=') {
			s.addToken(Walrus)
		} else {
			s.addToken(Colon)
		}
	case ';':
		s.addToken(Semicolon)
	case '~':
		s.addToken(Tilde)
	case '.':
		//lint:ignore SA4000 // Intentional consecutive matches for ellipsis
		if s.match('.') && s.match('.') {
			s.addToken(Ellipsis)
		} else if isDigit(s.peek()) {
			s.number()
		} else {
			s.addToken(Dot)
		}

	// ── operator and assignment combos (longest-match first) ──
	case '+':
		if s.match('=') {
			s.addToken(PlusEqual)
		} else {
			s.addToken(Plus)
		}
	case '-':
		if s.match('=') {
			s.addToken(MinusEqual)
		} else if s.match('>') {
			s.addToken(Arrow)
		} else {
			s.addToken(Minus)
		}
	case '*':
		if s.match('*') {
			if s.match('=') {
				s.addToken(StarStarEqual)
			} else {
				s.addToken(StarStar)
			}
		} else if s.match('=') {
			s.addToken(StarEqual)
		} else {
			s.addToken(Star)
		}
	case '/':
		if s.match('/') {
			if s.match('=') {
				s.addToken(SlashSlashEqual)
			} else {
				s.addToken(SlashSlash)
			}
		} else if s.match('=') {
			s.addToken(SlashEqual)
		} else {
			s.addToken(Slash)
		}
	case '%':
		if s.match('=') {
			s.addToken(PercentEqual)
		} else {
			s.addToken(Percent)
		}
	case '|':
		if s.match('=') {
			s.addToken(PipeEqual)
		} else {
			s.addToken(Pipe)
		}
	case '&':
		if s.match('=') {
			s.addToken(AmpEqual)
		} else {
			s.addToken(Ampersand)
		}
	case '^':
		if s.match('=') {
			s.addToken(CaretEqual)
		} else {
			s.addToken(Caret)
		}
	case '<':
		if s.match('<') {
			if s.match('=') {
				s.addToken(LessLessEqual)
			} else {
				s.addToken(LessLess)
			}
		} else if s.match('=') {
			s.addToken(LessEqual)
		} else {
			s.addToken(Less)
		}
	case '>':
		if s.match('>') {
			if s.match('=') {
				s.addToken(GreaterGreaterEqual)
			} else {
				s.addToken(GreaterGreater)
			}
		} else if s.match('=') {
			s.addToken(GreaterEqual)
		} else {
			s.addToken(Greater)
		}
	case '=':
		if s.match('=') {
			s.addToken(EqualEqual)
		} else {
			s.addToken(Equal)
		}
	case '!':
		if s.match('=') {
			s.addToken(BangEqual)
		} else {
			s.errorf("unexpected '!' – only '!=' is valid in Python")
		}
	case '@':
		if s.match('=') {
			s.addToken(AtEqual)
		} else {
			s.addToken(At)
		}

	// ── whitespace / comments / newlines ──
	case ' ', '\t', '\r':
		// just skip – indentation handled after NEWLINE
	case '\n':
		s.handleNewline() // emits NEWLINE + {INDENT,DEDENT}*
	case '#':
		// consume comment until physical line break
		for !s.atEnd() && s.peek() != '\n' {
			s.advance()
		}
		// newline will be consumed on next loop

	// ── literals / identifiers ──
	case '"', '\'':
		s.string(r)
	default:
		switch {
		case r == 'f' || r == 'F':
			// Check if this is an f-string (f" or f')
			if s.peek() == '"' || s.peek() == '\'' {
				quote := s.peek()
				s.fstring(quote)
				return
			}
			// If not an f-string, treat as identifier
			s.identifier()
		case isIdentifierStart(r):
			s.identifier()
		case unicode.IsDigit(r):
			s.number()
		default:
			s.errorf("unexpected character %q", r)
		}
	}
}

// ── NEWLINE + INDENT / DEDENT handling ──────────────────────────────

func (s *Scanner) handleNewline() {
	// Physical newline was already consumed by advance().
	if s.parenDepth > 0 {
		// inside (), [], {} ⇒ newline is whitespace
		return
	}
	s.addToken(Newline)

	// Measure indentation on following line:
	indent := 0
	for {
		switch s.peek() {
		case ' ':
			s.advance()
			indent++
		case '\t':
			s.advance()
			indent += 8 - (indent % 8) // tab stops
		case '\f':
			s.advance() // form-feed counts as 0 indent in CPython
			indent = 0
		case '\r': // CRLF normalisation
			s.advance()
		default:
			goto doneIndent
		}
	}
doneIndent:

	if s.peek() == '\n' || s.peek() == '#' || s.atEnd() {
		// blank line – do *not* change indent level
		return
	}

	// Compare to stack top
	top := s.indentStack[len(s.indentStack)-1]
	switch {
	case indent == top:
		// nothing to emit
	case indent > top:
		s.indentStack = append(s.indentStack, indent)
		s.addToken(Indent)
	case indent < top:
		for indent < top {
			s.indentStack = s.indentStack[:len(s.indentStack)-1]
			top = s.indentStack[len(s.indentStack)-1]
			s.addToken(Dedent)
		}
		if indent != top {
			s.errorf("inconsistent indentation")
		}
	}
}

// ── identifier / keyword ────────────────────────────────────────────

func (s *Scanner) identifier() {
	for isIdentifierContinue(s.peek()) {
		s.advance()
	}

	lexeme := string(s.src[s.start:s.cur])

	if tok, ok := Keywords[lexeme]; ok {
		s.addToken(tok)
		return
	}
	s.addToken(Identifier)
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierContinue(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ── numeric literal (decimal, binary, octal, hex & float) ───────────

func (s *Scanner) number() {
	// Check for special prefixes after initial '0'
	if s.src[s.start] == '0' && !s.atEnd() {
		next := s.peek()
		switch next {
		case 'b', 'B':
			s.binaryNumber()
			return
		case 'o', 'O':
			s.octalNumber()
			return
		case 'x', 'X':
			s.hexNumber()
			return
		}
	}

	// Regular decimal number (possibly with decimal point and exponent)
	for unicode.IsDigit(s.peek()) {
		s.advance()
	}

	isFloat := false
	if s.peek() == '.' && unicode.IsDigit(s.peekN(1)) {
		isFloat = true
		s.advance() // consume '.'
		for unicode.IsDigit(s.peek()) {
			s.advance()
		}
	}
	// exponent part
	if p := s.peek(); p == 'e' || p == 'E' {
		isFloat = true
		s.advance()
		if s.peek() == '+' || s.peek() == '-' {
			s.advance()
		}
		for unicode.IsDigit(s.peek()) {
			s.advance()
		}
	}

	// Check for imaginary unit (j or J suffix)
	if p := s.peek(); p == 'j' || p == 'J' {
		s.advance() // consume 'j' or 'J'
		s.complexNumber(isFloat)
		return
	}

	lit := string(s.src[s.start:s.cur])
	if isFloat {
		val, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			s.errorf("invalid float literal: %v", err)
			return
		}
		s.addTokenLit(Number, val)
	} else {
		val, err := strconv.ParseInt(lit, 10, 64)
		if err != nil {
			s.errorf("invalid int literal: %v", err)
			return
		}
		s.addTokenLit(Number, val)
	}
}

// complexNumber handles imaginary number literals (ending with j/J)
func (s *Scanner) complexNumber(isFloat bool) {
	// Get the numeric part (excluding the 'j' suffix)
	numStr := string(s.src[s.start : s.cur-1])

	if isFloat {
		imagPart, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			s.errorf("invalid complex literal: %v", err)
			return
		}
		// Create complex number with 0 real part and parsed imaginary part
		complexVal := complex(0, imagPart)
		s.addTokenLit(Number, complexVal)
	} else {
		imagPart, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			s.errorf("invalid complex literal: %v", err)
			return
		}
		// Create complex number with 0 real part and parsed imaginary part
		complexVal := complex(0, float64(imagPart))
		s.addTokenLit(Number, complexVal)
	}
}

// binaryNumber handles binary literals like 0b1010 or 0B1010
func (s *Scanner) binaryNumber() {
	s.advance() // consume 'b' or 'B'

	start := s.cur
	for {
		r := s.peek()
		if r == '0' || r == '1' {
			s.advance()
		} else {
			break
		}
	}

	if s.cur == start {
		s.errorf("invalid binary literal: no digits after 0b/0B")
		return
	}

	// Check for imaginary unit after binary number
	if p := s.peek(); p == 'j' || p == 'J' {
		s.advance() // consume 'j' or 'J'
		// Parse the binary digits (skip the "0b" prefix)
		binaryStr := string(s.src[s.start+2 : s.cur-1])
		val, err := strconv.ParseInt(binaryStr, 2, 64)
		if err != nil {
			s.errorf("invalid binary complex literal: %v", err)
			return
		}
		complexVal := complex(0, float64(val))
		s.addTokenLit(Number, complexVal)
		return
	}

	// Parse the binary digits (skip the "0b" prefix)
	binaryStr := string(s.src[s.start+2 : s.cur])
	val, err := strconv.ParseInt(binaryStr, 2, 64)
	if err != nil {
		s.errorf("invalid binary literal: %v", err)
		return
	}
	s.addTokenLit(Number, val)
}

// octalNumber handles octal literals like 0o755 or 0O755
func (s *Scanner) octalNumber() {
	s.advance() // consume 'o' or 'O'

	start := s.cur
	for {
		r := s.peek()
		if r >= '0' && r <= '7' {
			s.advance()
		} else {
			break
		}
	}

	if s.cur == start {
		s.errorf("invalid octal literal: no digits after 0o/0O")
		return
	}

	// Check for imaginary unit after octal number
	if p := s.peek(); p == 'j' || p == 'J' {
		s.advance() // consume 'j' or 'J'
		// Parse the octal digits (skip the "0o" prefix)
		octalStr := string(s.src[s.start+2 : s.cur-1])
		val, err := strconv.ParseInt(octalStr, 8, 64)
		if err != nil {
			s.errorf("invalid octal complex literal: %v", err)
			return
		}
		complexVal := complex(0, float64(val))
		s.addTokenLit(Number, complexVal)
		return
	}

	// Parse the octal digits (skip the "0o" prefix)
	octalStr := string(s.src[s.start+2 : s.cur])
	val, err := strconv.ParseInt(octalStr, 8, 64)
	if err != nil {
		s.errorf("invalid octal literal: %v", err)
		return
	}
	s.addTokenLit(Number, val)
}

// hexNumber handles hexadecimal literals like 0x123 or 0X123
func (s *Scanner) hexNumber() {
	s.advance() // consume 'x' or 'X'

	start := s.cur
	for {
		r := s.peek()
		if unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			s.advance()
		} else {
			break
		}
	}

	if s.cur == start {
		s.errorf("invalid hexadecimal literal: no digits after 0x/0X")
		return
	}

	// Check for imaginary unit after hex number
	if p := s.peek(); p == 'j' || p == 'J' {
		s.advance() // consume 'j' or 'J'
		// Parse the hex digits (skip the "0x" prefix)
		hexStr := string(s.src[s.start+2 : s.cur-1])
		val, err := strconv.ParseInt(hexStr, 16, 64)
		if err != nil {
			s.errorf("invalid hexadecimal complex literal: %v", err)
			return
		}
		complexVal := complex(0, float64(val))
		s.addTokenLit(Number, complexVal)
		return
	}

	// Parse the hex digits (skip the "0x" prefix)
	hexStr := string(s.src[s.start+2 : s.cur])
	val, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		s.errorf("invalid hexadecimal literal: %v", err)
		return
	}
	s.addTokenLit(Number, val)
}

// ── string literal (single / double; no prefixes yet) ───────────────

func (s *Scanner) string(quote rune) {
	isTriple := s.peek() == quote && s.peekN(1) == quote
	if isTriple {
		// consume the two additional quotes
		s.advance()
		s.advance()
		for {
			if s.atEnd() {
				s.errorf("unterminated triple-quoted string")
				return
			}
			if s.peek() == quote && s.peekN(1) == quote && s.peekN(2) == quote {
				// closing """
				s.advance()
				s.advance()
				s.advance()
				break
			}
			s.advance()
		}
		// For triple-quoted strings, skip 3 characters at start and end
		body := s.src[s.start+3 : s.cur-3]
		s.addTokenLit(String, string(body))
	} else {
		for {
			if s.atEnd() {
				s.errorf("unterminated string")
				return
			}
			r := s.peek()
			if r == '\n' {
				s.errorf("string literal cannot span newline")
				return
			}
			if r == '\\' { // escape
				s.advance()
				s.advance()
				continue
			}
			if r == quote {
				s.advance()
				break
			}
			s.advance()
		}
		// For regular strings, skip 1 character at start and end
		body := s.src[s.start+1 : s.cur-1]
		s.addTokenLit(String, string(body))
	}
}

// ── f-string literal ──────────────────────────────────────────────

func (s *Scanner) fstring(quote rune) {
	// We've already consumed 'f', now consume the opening quote
	s.advance() // consume the quote

	// Emit FSTRING_START token (includes 'f"' or "f'")
	s.addToken(FStringStart)

	// Push f-string context onto stack
	s.fstringStack = append(s.fstringStack, fstringContext{
		quote:        quote,
		braceDepth:   0,
		inExpression: false,
		inFormatSpec: false,
	})

	// Start tracking content after the opening quote
	s.start = s.cur

	// Continue scanning in f-string mode
	s.scanFStringContent()
}

// scanFStringContent handles the main f-string content scanning
func (s *Scanner) scanFStringContent() {
	for !s.atEnd() && len(s.fstringStack) > 0 {
		ctx := &s.fstringStack[len(s.fstringStack)-1]

		if ctx.inExpression {
			s.scanFStringExpression()
		} else {
			s.scanFStringText()
		}
	}
}

// scanFStringText scans literal text portions of f-strings
func (s *Scanner) scanFStringText() {
	ctx := &s.fstringStack[len(s.fstringStack)-1]

	for !s.atEnd() {
		r := s.peek()

		// Check for end of f-string
		if r == ctx.quote {
			// Emit any accumulated text
			if s.cur > s.start {
				text := string(s.src[s.start:s.cur])
				s.addTokenLit(FStringMiddle, text)
			}

			// Consume closing quote and emit FSTRING_END
			s.start = s.cur
			s.advance()
			s.addToken(FStringEnd)

			// Pop f-string context
			s.fstringStack = s.fstringStack[:len(s.fstringStack)-1]

			// Reset start position
			s.start = s.cur
			return
		}

		// Check for start of replacement field
		if r == '{' {
			// Check for escaped brace {{
			if s.peekN(1) == '{' {
				// Include the escaped brace in the text
				s.advance() // consume first {
				s.advance() // consume second {
				continue
			}

			// Emit any accumulated text before the replacement field
			if s.cur > s.start {
				text := string(s.src[s.start:s.cur])
				s.addTokenLit(FStringMiddle, text)
			}

			// Start replacement field
			s.start = s.cur
			s.advance() // consume '{'
			s.addToken(LeftBraceF)

			// Switch to expression mode
			ctx.inExpression = true
			ctx.braceDepth = 1

			// Reset start for expression parsing
			s.start = s.cur
			return
		}

		// Check for unmatched closing brace
		if r == '}' {
			// Check for escaped brace }}
			if s.peekN(1) == '}' {
				// Include the escaped brace in the text
				s.advance() // consume first }
				s.advance() // consume second }
				continue
			}

			// Unmatched closing brace
			s.errorf("f-string: single '}' is not allowed")
			return
		}

		// Handle newlines in f-strings (only allowed in triple-quoted)
		if r == '\n' {
			// For now, allow newlines (proper handling would check for triple quotes)
			s.advance()
			continue
		}

		// Handle escape sequences
		if r == '\\' {
			s.advance() // consume backslash
			if !s.atEnd() {
				s.advance() // consume escaped character
			}
			continue
		}

		s.advance()
	}

	// If we get here, the f-string was not terminated
	s.errorf("unterminated f-string")
}

// scanFStringExpression scans expressions inside f-string replacement fields
func (s *Scanner) scanFStringExpression() {
	ctx := &s.fstringStack[len(s.fstringStack)-1]

	// We're in expression mode - scan tokens directly but watch for special characters
	for !s.atEnd() && ctx.inExpression {
		s.lexLine, s.lexCol = s.line, s.col
		s.start = s.cur

		r := s.peek()

		// Handle braces for nesting
		if r == '{' {
			ctx.braceDepth++
			s.advance()
			s.addToken(LeftBrace)
			continue
		}

		if r == '}' {
			ctx.braceDepth--
			if ctx.braceDepth == 0 {
				// End of replacement field
				s.advance()
				s.addToken(RightBraceF)
				ctx.inExpression = false
				ctx.inFormatSpec = false
				s.start = s.cur
				return
			} else {
				s.advance()
				s.addToken(RightBrace)
				continue
			}
		}

		// Handle debugging equals (=)
		if r == '=' && !ctx.inFormatSpec {
			s.advance()
			s.addToken(FStringEqual)
			continue
		}

		// Handle conversion specifier (!)
		if r == '!' && !ctx.inFormatSpec {
			s.advance()
			s.addToken(FStringConversionStart)
			// Next should be a name (r, s, a)
			s.start = s.cur
			if isIdentifierStart(s.peek()) {
				s.identifier()
			}
			continue
		}

		// Handle format specification (:)
		if r == ':' && !ctx.inFormatSpec {
			s.advance()
			s.addToken(Colon)
			ctx.inFormatSpec = true
			s.start = s.cur
			s.scanFStringFormatSpec()
			continue
		}

		// For all other characters, use direct token scanning (not scanToken to avoid recursion)
		s.scanExpressionToken()
	}
}

// scanExpressionToken scans a single token for expressions (used in f-strings)
// This is like scanToken but doesn't check for f-string context to avoid recursion
func (s *Scanner) scanExpressionToken() {
	switch r := s.advance(); r {
	// ── single char punctuation ──
	case '(':
		s.parenDepth++
		s.addToken(LeftParen)
	case ')':
		if s.parenDepth > 0 {
			s.parenDepth--
		}
		s.addToken(RightParen)
	case '[':
		s.parenDepth++
		s.addToken(LeftBracket)
	case ']':
		if s.parenDepth > 0 {
			s.parenDepth--
		}
		s.addToken(RightBracket)
	case ',':
		s.addToken(Comma)
	case ';':
		s.addToken(Semicolon)
	case '~':
		s.addToken(Tilde)
	case '.':
		if s.match('.') && s.match('.') {
			s.addToken(Ellipsis)
		} else if isDigit(s.peek()) {
			s.number()
		} else {
			s.addToken(Dot)
		}

	// ── operator and assignment combos ──
	case '+':
		if s.match('=') {
			s.addToken(PlusEqual)
		} else {
			s.addToken(Plus)
		}
	case '-':
		if s.match('=') {
			s.addToken(MinusEqual)
		} else if s.match('>') {
			s.addToken(Arrow)
		} else {
			s.addToken(Minus)
		}
	case '*':
		if s.match('*') {
			if s.match('=') {
				s.addToken(StarStarEqual)
			} else {
				s.addToken(StarStar)
			}
		} else if s.match('=') {
			s.addToken(StarEqual)
		} else {
			s.addToken(Star)
		}
	case '/':
		if s.match('/') {
			if s.match('=') {
				s.addToken(SlashSlashEqual)
			} else {
				s.addToken(SlashSlash)
			}
		} else if s.match('=') {
			s.addToken(SlashEqual)
		} else {
			s.addToken(Slash)
		}
	case '%':
		if s.match('=') {
			s.addToken(PercentEqual)
		} else {
			s.addToken(Percent)
		}
	case '|':
		if s.match('=') {
			s.addToken(PipeEqual)
		} else {
			s.addToken(Pipe)
		}
	case '&':
		if s.match('=') {
			s.addToken(AmpEqual)
		} else {
			s.addToken(Ampersand)
		}
	case '^':
		if s.match('=') {
			s.addToken(CaretEqual)
		} else {
			s.addToken(Caret)
		}
	case '<':
		if s.match('<') {
			if s.match('=') {
				s.addToken(LessLessEqual)
			} else {
				s.addToken(LessLess)
			}
		} else if s.match('=') {
			s.addToken(LessEqual)
		} else {
			s.addToken(Less)
		}
	case '>':
		if s.match('>') {
			if s.match('=') {
				s.addToken(GreaterGreaterEqual)
			} else {
				s.addToken(GreaterGreater)
			}
		} else if s.match('=') {
			s.addToken(GreaterEqual)
		} else {
			s.addToken(Greater)
		}
	case '!':
		if s.match('=') {
			s.addToken(BangEqual)
		} else {
			s.errorf("unexpected '!' – only '!=' is valid in Python")
		}
	case '@':
		if s.match('=') {
			s.addToken(AtEqual)
		} else {
			s.addToken(At)
		}

	// ── whitespace / comments / newlines ──
	case ' ', '\t', '\r':
		// just skip
	case '\n':
		// In expressions, newlines are usually ignored
	case '#':
		// consume comment until physical line break
		for !s.atEnd() && s.peek() != '\n' {
			s.advance()
		}

	// ── literals / identifiers ──
	case '"', '\'':
		s.string(r)
	default:
		switch {
		case isIdentifierStart(r):
			s.identifier()
		case unicode.IsDigit(r):
			s.number()
		default:
			s.errorf("unexpected character %q", r)
		}
	}
}

// scanFStringFormatSpec scans format specifications (after :)
func (s *Scanner) scanFStringFormatSpec() {
	ctx := &s.fstringStack[len(s.fstringStack)-1]

	for !s.atEnd() && ctx.inFormatSpec {
		r := s.peek()

		// End of replacement field
		if r == '}' {
			// Emit any accumulated format spec text
			if s.cur > s.start {
				text := string(s.src[s.start:s.cur])
				s.addTokenLit(FStringMiddle, text)
			}
			ctx.inFormatSpec = false
			return
		}

		// Nested replacement field in format spec
		if r == '{' {
			// Emit any accumulated format spec text
			if s.cur > s.start {
				text := string(s.src[s.start:s.cur])
				s.addTokenLit(FStringMiddle, text)
			}

			// Start nested replacement field
			s.start = s.cur
			s.advance()
			s.addToken(LeftBraceF)

			// Increase brace depth for nested expression
			ctx.braceDepth++

			s.start = s.cur
			return
		}

		s.advance()
	}
}

// ── small utility ───────────────────────────────────────────────────

func isDigit(r rune) bool { return unicode.IsDigit(r) }
