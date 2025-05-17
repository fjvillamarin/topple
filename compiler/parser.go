package compiler

// The scaffold parses only *one-line expression statements* so that you can
// start writing tests immediately and grow the grammar feature-by-feature.

type Parser struct {
	toks   []Token
	cur    int
	Errors []error
}

// NewParser returns a new parser instance.
func NewParser(tokens []Token) *Parser { return &Parser{toks: tokens} }

// Parse runs to EOF and produces a *Module* AST.
func (p *Parser) Parse() (*Module, []error) {
	mod := &Module{}

	for !p.isAtEnd() {
		if stmt := p.simpleStmt(); stmt != nil {
			mod.Body = append(mod.Body, stmt)
		} else { // panic-mode recovery: skip one token
			p.Errors = append(p.Errors, NewParseError(p.peek(), "invalid syntax"))
			p.advance()
		}
	}
	return mod, p.Errors
}

// ── phase-0 grammar: simple_stmt ::= expr NEWLINE | NEWLINE ──────────
func (p *Parser) simpleStmt() Stmt {
	// blank line?
	if p.match(Newline) {
		return nil
	}

	expr := p.atom()
	if expr == nil {
		return nil
	}

	p.expect(Newline) // consume required terminator; tolerate optional SEMI later
	return &ExprStmt{Value: expr, Tok: p.prev()}
}

// ── atoms: NAME | NUMBER | STRING ────────────────────────────────────
func (p *Parser) atom() Expr {
	switch {
	case p.match(Identifier):
		return &Name{Tok: p.prev()}
	case p.match(Number, String):
		return &Constant{Tok: p.prev(), Value: p.prev().Literal}
	default:
		p.Errors = append(p.Errors, NewParseError(p.peek(), "expected expression"))
		return nil
	}
}

// ── tiny token-level helpers (similar to your Lox version) ───────────
func (p *Parser) match(kinds ...TokenType) bool {
	for _, k := range kinds {
		if p.check(k) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(kind TokenType) {
	if !p.match(kind) {
		p.Errors = append(p.Errors, NewParseError(p.peek(), "expected '"+kind.String()+"'"))
	}
}

func (p *Parser) check(k TokenType) bool { return !p.isAtEnd() && p.peek().Type == k }
func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.cur++
	}
}
func (p *Parser) isAtEnd() bool { return p.peek().Type == EOF }
func (p *Parser) peek() Token   { return p.toks[p.cur] }
func (p *Parser) prev() Token   { return p.toks[p.cur-1] }
