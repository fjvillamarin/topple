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
	"strings"
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
	nestingLevel int  // Level of f-string nesting (for nested f-strings)
	isRaw        bool // Whether this is a raw f-string
	isTriple     bool // Whether this is a triple-quoted f-string
}

// LexMode represents the current lexing mode
type LexMode int

const (
	PythonMode            LexMode = iota // Normal Python tokenization
	HTMLTagMode                          // Inside <tag attributes> - handles tag names, attributes, and >
	HTMLContentMode                      // Inside tag content, can contain text
	HTMLInterpolationMode                // Inside {expression} within HTML
)

// LexerContext tracks the state for context-aware lexing
type LexerContext struct {
	viewDepth       int       // Depth of nested view functions (0 = not in view)
	mode            LexMode   // Current lexing mode
	htmlTagDepth    int       // Track HTML tag nesting
	atLineStart     bool      // Are we at the start of a logical line?
	pendingIndent   bool      // Waiting to process indentation?
	htmlTagName     string    // Current HTML tag being processed
	isClosingTag    bool      // Whether we're parsing a closing tag
	inHTMLAttribute bool      // Whether we're inside an HTML attribute
	modeStack       []LexMode // Stack to track mode before interpolations
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

	// Lexer context for HTML/Python mode switching
	ctx LexerContext
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
		ctx: LexerContext{
			mode:        PythonMode,
			atLineStart: true,
		},
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
		// Set proper position for dedent token at EOF
		s.start = s.cur
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

// ── context-aware lexing helpers ────────────────────────────────────

// handleLineStart determines the lexing mode based on the first non-whitespace character
func (s *Scanner) handleLineStart() {
	if s.ctx.viewDepth == 0 {
		// Outside view functions, always Python mode
		s.ctx.mode = PythonMode
		s.ctx.atLineStart = false
		return
	}

	// Skip whitespace to find first significant character
	firstChar := s.peekFirstNonWhitespace()

	switch {
	case firstChar == '<':
		// Line starts with < → HTML mode
		s.ctx.mode = HTMLTagMode
		// Skip to the '<' character
		s.skipToFirstNonWhitespace()

	case isIdentifierStart(firstChar):
		// Check if it's a keyword that indicates Python mode
		if s.isKeywordAtPosition() {
			s.ctx.mode = PythonMode
		} else {
			// Identifier at line start → Python mode (assignment, etc.)
			s.ctx.mode = PythonMode
		}

	case s.parenDepth > 0:
		// Continuation line with open parentheses → keep current mode
		// (don't change mode)

	default:
		// Default to Python mode for other cases
		s.ctx.mode = PythonMode
	}

	s.ctx.atLineStart = false
}

// skipToFirstNonWhitespace advances the cursor to the first non-whitespace character
func (s *Scanner) skipToFirstNonWhitespace() {
	for s.cur < len(s.src) {
		r, size := utf8.DecodeRune(s.src[s.cur:])
		if r != ' ' && r != '\t' && r != '\r' {
			break
		}
		s.cur += size
		s.col++
	}
}

// peekFirstNonWhitespace returns the first non-whitespace character from current position
func (s *Scanner) peekFirstNonWhitespace() rune {
	i := s.cur
	for i < len(s.src) {
		r, size := utf8.DecodeRune(s.src[i:])
		if r != ' ' && r != '\t' && r != '\r' {
			return r
		}
		i += size
	}
	return -1 // EOF or newline
}

// isKeywordAtPosition checks if there's a Python keyword at the current position
func (s *Scanner) isKeywordAtPosition() bool {
	// Save current position
	savedCur := s.cur

	// Scan identifier
	if !isIdentifierStart(s.peekFirstNonWhitespace()) {
		return false
	}

	// Skip to first non-whitespace
	for s.cur < len(s.src) {
		r, size := utf8.DecodeRune(s.src[s.cur:])
		if r != ' ' && r != '\t' && r != '\r' {
			break
		}
		s.cur += size
	}

	start := s.cur
	for s.cur < len(s.src) {
		r, size := utf8.DecodeRune(s.src[s.cur:])
		if !isIdentifierContinue(r) {
			break
		}
		s.cur += size
	}

	identifier := string(s.src[start:s.cur])
	isKeyword := IsKeyword(identifier)

	// Restore position
	s.cur = savedCur

	return isKeyword
}

// detectViewFunction checks if we're entering a view function
func (s *Scanner) detectViewFunction() {
	// Increment view depth for nested views
	s.ctx.viewDepth++
}

// ── main dispatcher ─────────────────────────────────────────────────

func (s *Scanner) scanToken() {
	// Handle line start mode detection
	if s.ctx.atLineStart {
		s.handleLineStart()
	}

	// Check if we're inside an f-string and should continue f-string scanning
	if len(s.fstringStack) > 0 {
		s.scanFStringContent()
		return
	}

	// Route to appropriate scanner based on current mode
	switch s.ctx.mode {
	case PythonMode:
		s.scanPythonToken()
	case HTMLTagMode:
		s.scanHTMLTag()
	case HTMLContentMode:
		s.scanHTMLContent()
	case HTMLInterpolationMode:
		s.scanPythonToken() // Python expressions inside {}
	}
}

// scanPythonToken handles Python tokenization
func (s *Scanner) scanPythonToken() {
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

		// Check if we're closing HTML interpolation
		if s.ctx.mode == HTMLInterpolationMode {
			s.addToken(HTMLInterpolationEnd)
			// Restore previous mode from stack
			if len(s.ctx.modeStack) > 0 {
				s.ctx.mode = s.ctx.modeStack[len(s.ctx.modeStack)-1]
				s.ctx.modeStack = s.ctx.modeStack[:len(s.ctx.modeStack)-1] // Pop from stack
			} else {
				// Fallback to content mode if stack is empty
				s.ctx.mode = HTMLContentMode
			}
		} else {
			s.addToken(RightBrace)
		}
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
		// If we're in a view and could be starting an HTML tag
		if s.ctx.viewDepth > 0 && s.ctx.mode == PythonMode {
			// Check if this looks like an HTML tag
			nextChar := s.peek()
			if isIdentifierStart(nextChar) || nextChar == '/' {
				s.ctx.mode = HTMLTagMode
				s.addToken(TagOpen)
				return
			}
		}
		// Otherwise, treat as less-than operator
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
		s.handleNewline()        // emits NEWLINE + {INDENT,DEDENT}*
		s.ctx.atLineStart = true // Mark that we're at line start after newline
	case '#':
		// consume comment until physical line break
		for !s.atEnd() && s.peek() != '\n' {
			s.advance()
		}
		// newline will be consumed on next loop

	// ── literals / identifiers ──
	case '"', '\'':
		// Check if this might be part of a nested f-string
		// We need to look back to see if there was an 'f' or 'r' prefix
		if s.start > 0 {
			// Check for f-string prefix
			prevStart := s.start - 1
			if prevStart > 0 && (s.src[prevStart] == 'f' || s.src[prevStart] == 'F') {
				// This is a nested f-string - handle it specially
				s.cur = s.start     // Reset to before the quote
				s.start = prevStart // Include the 'f' prefix
				s.fstring(r)
				return
			}
			// Check for raw f-string prefix (rf or fr)
			if prevStart > 0 && (s.src[prevStart] == 'r' || s.src[prevStart] == 'R') {
				prevPrevStart := prevStart - 1
				if prevPrevStart >= 0 && (s.src[prevPrevStart] == 'f' || s.src[prevPrevStart] == 'F') {
					// This is a raw f-string (fr")
					s.cur = s.start         // Reset to before the quote
					s.start = prevPrevStart // Include the 'fr' prefix
					s.fstring(r)
					return
				}
			}
			if prevStart > 0 && (s.src[prevStart] == 'f' || s.src[prevStart] == 'F') {
				prevPrevStart := prevStart - 1
				if prevPrevStart >= 0 && (s.src[prevPrevStart] == 'r' || s.src[prevPrevStart] == 'R') {
					// This is a raw f-string (rf")
					s.cur = s.start         // Reset to before the quote
					s.start = prevPrevStart // Include the 'rf' prefix
					s.fstring(r)
					return
				}
			}
		}
		s.string(r)
	default:
		switch {
		case r == 'f' || r == 'F':
			// Check if this is a nested f-string (f" or f')
			if s.peek() == '"' || s.peek() == '\'' {
				quote := s.peek()
				s.fstring(quote)
				return
			}
			// If not an f-string, treat as identifier
			s.identifier()
		case r == 'r' || r == 'R':
			// Check if this is a raw f-string (rf" or rf')
			next := s.peek()
			if next == 'f' || next == 'F' {
				s.advance() // consume 'f' or 'F'
				if s.peek() == '"' || s.peek() == '\'' {
					quote := s.peek()
					s.fstring(quote)
					return
				}
				// If not an f-string, backtrack and treat as identifier
				s.cur-- // backtrack the 'f'
				s.identifier()
				return
			}
			// Check if this is a raw string (r" or r')
			if next == '"' || next == '\'' {
				s.advance() // consume quote
				s.string(next)
				return
			}
			// If not a raw string, treat as identifier
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
		// Special handling for 'view' keyword
		if tok == View {
			s.detectViewFunction()
		}
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
	// Check if this is a triple-quoted f-string
	isTriple := s.peek() == quote && s.peekN(1) == quote

	// Determine if this is a raw f-string by checking if we consumed 'r' before 'f'
	isRaw := false
	if s.start > 0 {
		// Check if the character before 'f' was 'r' or 'R'
		prevChar, _ := utf8.DecodeLastRune(s.src[:s.start])
		if prevChar == 'r' || prevChar == 'R' {
			isRaw = true
		}
	}

	// Consume the opening quote(s)
	s.advance() // consume the first quote
	if isTriple {
		s.advance() // consume second quote
		s.advance() // consume third quote
	}

	// Emit FSTRING_START token
	s.addToken(FStringStart)

	// Determine nesting level
	nestingLevel := 0
	if len(s.fstringStack) > 0 {
		nestingLevel = s.fstringStack[len(s.fstringStack)-1].nestingLevel + 1
	}

	// Push f-string context onto stack
	s.fstringStack = append(s.fstringStack, fstringContext{
		quote:        quote,
		braceDepth:   0,
		inExpression: false,
		inFormatSpec: false,
		nestingLevel: nestingLevel,
		isRaw:        isRaw,
		isTriple:     isTriple,
	})

	// Start tracking content after the opening quote(s)
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
	// Check if we have a valid f-string context
	if len(s.fstringStack) == 0 {
		s.errorf("internal error: scanFStringText called without f-string context")
		return
	}

	ctx := &s.fstringStack[len(s.fstringStack)-1]

	for !s.atEnd() {
		r := s.peek()

		// Check for end of f-string
		if r == ctx.quote {
			if ctx.isTriple {
				// For triple-quoted f-strings, need to check for three consecutive quotes
				if s.peekN(1) == ctx.quote && s.peekN(2) == ctx.quote {
					// Emit any accumulated text
					if s.cur > s.start {
						text := string(s.src[s.start:s.cur])
						s.addTokenLit(FStringMiddle, text)
					}

					// Consume closing quotes and emit FSTRING_END
					s.start = s.cur
					s.advance() // first quote
					s.advance() // second quote
					s.advance() // third quote
					s.addToken(FStringEnd)

					// Pop f-string context
					s.fstringStack = s.fstringStack[:len(s.fstringStack)-1]

					// Reset start position
					s.start = s.cur
					return
				}
			} else {
				// Single quote - end of regular f-string
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

			// Set start position for the replacement field token
			s.start = s.cur
			s.advance() // consume '{'
			s.addToken(LeftBraceF)

			// Set f-string context to expression mode
			ctx.inExpression = true
			ctx.braceDepth = 1

			// Reset start position for the expression
			s.start = s.cur
			return
		}

		// Check for unmatched closing brace (should not happen in valid f-strings)
		if r == '}' {
			s.errorf("f-string: single '}' is not allowed")
			return
		}

		// Handle newlines in f-strings
		if r == '\n' {
			if !ctx.isTriple {
				// Newlines not allowed in single-quoted f-strings
				s.errorf("f-string: unterminated string literal (detected at line %d)", s.line)
				return
			}
			s.advance()
			continue
		}

		// Handle escape sequences
		if r == '\\' {
			if ctx.isRaw {
				// In raw f-strings, backslashes are literal
				s.advance()
			} else {
				// In regular f-strings, handle escape sequences
				s.advance() // consume backslash
				if !s.atEnd() {
					s.advance() // consume escaped character
				}
			}
			continue
		}

		s.advance()
	}

	// If we get here, the f-string was not terminated
	if ctx.isTriple {
		s.errorf("unterminated triple-quoted f-string")
	} else {
		s.errorf("unterminated f-string")
	}
}

// scanFStringExpression scans expressions inside f-string replacement fields
func (s *Scanner) scanFStringExpression() {
	// Check if we have a valid f-string context
	if len(s.fstringStack) == 0 {
		s.errorf("internal error: scanFStringExpression called without f-string context")
		return
	}

	ctx := &s.fstringStack[len(s.fstringStack)-1]

	// Additional safety check
	if !ctx.inExpression {
		s.errorf("internal error: scanFStringExpression called when not in expression mode")
		return
	}

	// We're in expression mode - scan tokens directly but watch for special characters
	for !s.atEnd() && ctx.inExpression && len(s.fstringStack) > 0 {
		s.lexLine, s.lexCol = s.line, s.col
		s.start = s.cur

		r := s.peek()

		// Handle braces for nesting
		if r == '{' {
			ctx.braceDepth++
			s.advance()
			// If we're in a format spec, this is a nested replacement field
			if ctx.inFormatSpec {
				s.addToken(LeftBraceF)
			} else {
				s.addToken(LeftBrace)
			}
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
				// If we're in a format spec, this closes a nested replacement field
				if ctx.inFormatSpec {
					s.addToken(RightBraceF)
					// After closing a nested replacement field in format spec,
					// we need to continue scanning the format spec
					s.start = s.cur
					s.scanFStringFormatSpec()
					return
				} else {
					s.addToken(RightBrace)
				}
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
		// Check if this might be part of a nested f-string
		// We need to look back to see if there was an 'f' or 'r' prefix
		if s.start > 0 {
			// Check for f-string prefix
			prevStart := s.start - 1
			if prevStart > 0 && (s.src[prevStart] == 'f' || s.src[prevStart] == 'F') {
				// This is a nested f-string - handle it specially
				s.cur = s.start     // Reset to before the quote
				s.start = prevStart // Include the 'f' prefix
				s.fstring(r)
				return
			}
			// Check for raw f-string prefix (rf or fr)
			if prevStart > 0 && (s.src[prevStart] == 'r' || s.src[prevStart] == 'R') {
				prevPrevStart := prevStart - 1
				if prevPrevStart >= 0 && (s.src[prevPrevStart] == 'f' || s.src[prevPrevStart] == 'F') {
					// This is a raw f-string (fr")
					s.cur = s.start         // Reset to before the quote
					s.start = prevPrevStart // Include the 'fr' prefix
					s.fstring(r)
					return
				}
			}
			if prevStart > 0 && (s.src[prevStart] == 'f' || s.src[prevStart] == 'F') {
				prevPrevStart := prevStart - 1
				if prevPrevStart >= 0 && (s.src[prevPrevStart] == 'r' || s.src[prevPrevStart] == 'R') {
					// This is a raw f-string (rf")
					s.cur = s.start         // Reset to before the quote
					s.start = prevPrevStart // Include the 'rf' prefix
					s.fstring(r)
					return
				}
			}
		}
		s.string(r)
	default:
		switch {
		case r == 'f' || r == 'F':
			// Check if this is a nested f-string (f" or f')
			if s.peek() == '"' || s.peek() == '\'' {
				quote := s.peek()
				s.fstring(quote)
				return
			}
			// If not an f-string, treat as identifier
			s.identifier()
		case r == 'r' || r == 'R':
			// Check if this is a raw f-string (rf" or rf')
			next := s.peek()
			if next == 'f' || next == 'F' {
				s.advance() // consume 'f' or 'F'
				if s.peek() == '"' || s.peek() == '\'' {
					quote := s.peek()
					s.fstring(quote)
					return
				}
				// If not an f-string, backtrack and treat as identifier
				s.cur-- // backtrack the 'f'
				s.identifier()
				return
			}
			// Check if this is a raw string (r" or r')
			if next == '"' || next == '\'' {
				s.advance() // consume quote
				s.string(next)
				return
			}
			// If not a raw string, treat as identifier
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

// scanFStringFormatSpec scans format specifications (after :)
func (s *Scanner) scanFStringFormatSpec() {
	// Check if we have a valid f-string context
	if len(s.fstringStack) == 0 {
		s.errorf("internal error: scanFStringFormatSpec called without f-string context")
		return
	}

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
			// Emit any accumulated format spec text before the nested field
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

			// Reset start position and continue scanning the expression
			s.start = s.cur

			// The nested replacement field will be handled by scanFStringExpression
			// When it returns, we need to continue scanning the format spec
			return
		}

		s.advance()
	}
}

// ── HTML scanning methods ──────────────────────────────────────────

// scanHTMLTag scans HTML tag content (tag name and attributes)
func (s *Scanner) scanHTMLTag() {
	// If we're at the start of a tag, we need to consume the '<'
	if s.peek() == '<' {
		s.advance() // consume '<'

		// Check for HTML comment first
		if s.peek() == '!' && s.peekN(1) == '-' && s.peekN(2) == '-' {
			s.scanHTMLComment()
			// After comment, we're back in content mode
			s.ctx.mode = HTMLContentMode
			return
		}

		// Check for closing tag
		if s.peek() == '/' {
			s.advance()               // consume '/'
			s.addToken(TagCloseStart) // Emit '</' token
			// Stay in tag mode to handle tag name
			return
		}

		// Emit '<' token for opening tag
		s.addToken(TagOpen)
		// Stay in tag mode to handle tag name and attributes
		return
	}

	// We're inside a tag, handle tag content
	for !s.atEnd() {
		s.start = s.cur

		switch r := s.advance(); r {
		case ' ', '\t', '\r':
			// Skip whitespace in tags
			continue
		case '\n':
			// Newlines in tags are treated as whitespace
			s.line++
			s.col = s.cfg.StartColumn
			continue
		case '>':
			// End of tag
			s.addToken(TagClose)
			s.ctx.mode = HTMLContentMode
			return
		case '/':
			// Check for self-closing tag
			if s.peek() == '>' {
				s.advance() // consume '>'
				s.addToken(TagSelfClose)
				s.ctx.mode = HTMLContentMode
				return
			}
			s.errorf("unexpected '/' in HTML tag")
			return
		case '=':
			s.addToken(Equal)
		case '"', '\'':
			s.string(r) // Use regular string parsing
		case '{':
			// Start of interpolation in attribute
			s.ctx.modeStack = append(s.ctx.modeStack, s.ctx.mode) // Push current mode
			s.addToken(HTMLInterpolationStart)
			s.ctx.mode = HTMLInterpolationMode
			return
		default:
			if isIdentifierStart(r) {
				s.scanHTMLIdentifier()
			} else {
				s.errorf("unexpected character %q in HTML tag", r)
				return
			}
		}
	}
}

// scanHTMLContent scans HTML content between tags
func (s *Scanner) scanHTMLContent() {
	textStart := s.cur

	for !s.atEnd() {
		r := s.peek()
		switch r {
		case '<':
			// Emit any accumulated text
			if s.cur > textStart {
				s.addHTMLText(textStart)
			}

			// Switch to tag mode to handle the '<'
			s.ctx.mode = HTMLTagMode
			return

		case '{':
			// Emit any accumulated text
			if s.cur > textStart {
				s.addHTMLText(textStart)
			}

			// Set start position for the interpolation token
			s.start = s.cur
			s.advance()                                           // consume '{'
			s.ctx.modeStack = append(s.ctx.modeStack, s.ctx.mode) // Push current mode
			s.addToken(HTMLInterpolationStart)
			s.ctx.mode = HTMLInterpolationMode
			return

		case '\n':
			// Emit any accumulated text first
			if s.cur > textStart {
				s.addHTMLText(textStart)
			}

			// Let the main scanner handle the newline properly
			// by returning to Python mode temporarily
			s.ctx.mode = PythonMode
			s.ctx.atLineStart = true
			return

		default:
			s.advance()
		}
	}

	// Emit any remaining text
	if s.cur > textStart {
		s.addHTMLText(textStart)
	}
}

// ── HTML helper methods ────────────────────────────────────────────

// isHTMLComment checks if we're at the start of an HTML comment
func (s *Scanner) isHTMLComment() bool {
	// We should be positioned after '<', check for '!--'
	return s.peek() == '!' && s.peekN(1) == '-' && s.peekN(2) == '-'
}

// scanHTMLComment scans and skips HTML comments
func (s *Scanner) scanHTMLComment() {
	// We've seen '<!' and confirmed next two chars are '--'
	s.advance() // consume '!'
	s.advance() // consume first '-'
	s.advance() // consume second '-'

	// Consume until we find '-->'
	for !s.atEnd() {
		if s.peek() == '-' && s.peekN(1) == '-' && s.peekN(2) == '>' {
			s.advance() // consume first '-'
			s.advance() // consume second '-'
			s.advance() // consume '>'
			return
		}
		r := s.advance()
		if r == '\n' {
			s.line++
			s.col = s.cfg.StartColumn
		}
	}

	s.errorf("unterminated HTML comment")
}

// isNextContentOnSameLine checks if content starts on the same line as the tag
func (s *Scanner) isNextContentOnSameLine() bool {
	// Skip whitespace
	i := s.cur
	for i < len(s.src) && (s.src[i] == ' ' || s.src[i] == '\t') {
		i++
	}

	// If we hit a newline or end of file, it's multiline
	if i >= len(s.src) || s.src[i] == '\n' || s.src[i] == '\r' {
		return false
	}

	return true
}

// scanHTMLIdentifier scans an identifier in HTML context (tag name or attribute name)
func (s *Scanner) scanHTMLIdentifier() {
	// We've already consumed the first character in scanHTMLTag
	// Continue scanning the rest of the identifier
	for isIdentifierContinue(s.peek()) || s.peek() == '-' {
		s.advance()
	}

	// This is a tag name or attribute name
	s.addToken(Identifier)
}

// addHTMLText adds an HTML text token from the given start position
func (s *Scanner) addHTMLText(textStart int) {
	text := string(s.src[textStart:s.cur])
	if len(text) > 0 {
		// Skip tokens that are ENTIRELY whitespace (like newlines/indentation)
		// But preserve PARTIAL whitespace (like "text: " or " text") as it's semantically significant
		if len(strings.TrimSpace(text)) == 0 {
			return // Skip entirely-whitespace tokens
		}
		// Preserve the text exactly as-is - whitespace is meaningful in HTML
		token := Token{
			Type:    HTMLTextInline,
			Lexeme:  text,
			Literal: text,
			Span: Span{
				Start: Position{Line: s.lexLine, Column: s.lexCol},
				End:   Position{Line: s.line, Column: s.col},
			},
		}
		s.tokens = append(s.tokens, token)
	}
}

// ── small utility ───────────────────────────────────────────────────

func isDigit(r rune) bool { return unicode.IsDigit(r) }
