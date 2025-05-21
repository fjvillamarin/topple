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
	if tok, ok := Keywords[string(s.src[s.start:s.cur])]; ok {
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

// ── numeric literal (decimal & float only – extend later) ───────────

func (s *Scanner) number() {
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
	}

	body := s.src[s.start+1 : s.cur-1]
	s.addTokenLit(String, string(body)) // raw; real unescape can be deferred
}

// ── small utility ───────────────────────────────────────────────────

func isDigit(r rune) bool { return unicode.IsDigit(r) }
