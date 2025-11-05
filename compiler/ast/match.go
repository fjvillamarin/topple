package ast

import (
	"fmt"
	"strings"
	"topple/compiler/lexer"
)

// MatchStmt represents a match statement: match subject_expr: case_blocks
type MatchStmt struct {
	Subject Expr        // The expression to match against
	Cases   []CaseBlock // One or more case blocks

	Span lexer.Span
}

func (m *MatchStmt) isStmt() {}

func (m *MatchStmt) GetSpan() lexer.Span {
	return m.Span
}

func (m *MatchStmt) String() string {
	var cases []string
	for _, c := range m.Cases {
		cases = append(cases, c.String())
	}
	return fmt.Sprintf("match %s: %s", m.Subject.String(), strings.Join(cases, " "))
}

func (m *MatchStmt) Accept(visitor Visitor) {
	visitor.VisitMatch(m)
}

// CaseBlock represents a case block: case patterns [guard]: block
type CaseBlock struct {
	Patterns []Pattern // One or more patterns to match
	Guard    Expr      // Optional guard condition (if expression)
	Body     []Stmt    // Statements in the case block

	Span lexer.Span
}

func (c *CaseBlock) String() string {
	var patterns []string
	for _, p := range c.Patterns {
		patterns = append(patterns, p.String())
	}
	result := fmt.Sprintf("case %s", strings.Join(patterns, " | "))
	if c.Guard != nil {
		result += fmt.Sprintf(" if %s", c.Guard.String())
	}
	return result + ": ..."
}

// Pattern is the interface for all pattern nodes
type Pattern interface {
	Node
	isPattern()
}

// LiteralPattern represents literal values in patterns (numbers, strings, None, True, False)
type LiteralPattern struct {
	Value Expr // The literal expression

	Span lexer.Span
}

func (lp *LiteralPattern) isPattern() {}

func (lp *LiteralPattern) GetSpan() lexer.Span {
	return lp.Span
}

func (lp *LiteralPattern) String() string {
	return lp.Value.String()
}

func (lp *LiteralPattern) Accept(visitor Visitor) {
	visitor.VisitLiteralPattern(lp)
}

// CapturePattern represents a name pattern that captures the matched value
type CapturePattern struct {
	Name *Name // The name to bind the matched value to

	Span lexer.Span
}

func (cp *CapturePattern) isPattern() {}

func (cp *CapturePattern) GetSpan() lexer.Span {
	return cp.Span
}

func (cp *CapturePattern) String() string {
	return cp.Name.String()
}

func (cp *CapturePattern) Accept(visitor Visitor) {
	visitor.VisitCapturePattern(cp)
}

// WildcardPattern represents the wildcard pattern (_)
type WildcardPattern struct {
	Span lexer.Span
}

func (wp *WildcardPattern) isPattern() {}

func (wp *WildcardPattern) GetSpan() lexer.Span {
	return wp.Span
}

func (wp *WildcardPattern) String() string {
	return "_"
}

func (wp *WildcardPattern) Accept(visitor Visitor) {
	visitor.VisitWildcardPattern(wp)
}

// ValuePattern represents dotted names in patterns (for constants)
type ValuePattern struct {
	Value Expr // Attribute access or name for constant values

	Span lexer.Span
}

func (vp *ValuePattern) isPattern() {}

func (vp *ValuePattern) GetSpan() lexer.Span {
	return vp.Span
}

func (vp *ValuePattern) String() string {
	return vp.Value.String()
}

func (vp *ValuePattern) Accept(visitor Visitor) {
	visitor.VisitValuePattern(vp)
}

// GroupPattern represents a parenthesized pattern
type GroupPattern struct {
	Pattern Pattern // The pattern inside parentheses

	Span lexer.Span
}

func (gp *GroupPattern) isPattern() {}

func (gp *GroupPattern) GetSpan() lexer.Span {
	return gp.Span
}

func (gp *GroupPattern) String() string {
	return fmt.Sprintf("(%s)", gp.Pattern.String())
}

func (gp *GroupPattern) Accept(visitor Visitor) {
	visitor.VisitGroupPattern(gp)
}

// SequencePattern represents list or tuple patterns
type SequencePattern struct {
	Patterns []Pattern // Patterns in the sequence
	IsTuple  bool      // true for tuple patterns, false for list patterns

	Span lexer.Span
}

func (sp *SequencePattern) isPattern() {}

func (sp *SequencePattern) GetSpan() lexer.Span {
	return sp.Span
}

func (sp *SequencePattern) String() string {
	var patterns []string
	for _, p := range sp.Patterns {
		patterns = append(patterns, p.String())
	}
	if sp.IsTuple {
		return fmt.Sprintf("(%s)", strings.Join(patterns, ", "))
	}
	return fmt.Sprintf("[%s]", strings.Join(patterns, ", "))
}

func (sp *SequencePattern) Accept(visitor Visitor) {
	visitor.VisitSequencePattern(sp)
}

// StarPattern represents star patterns in sequences (*pattern)
type StarPattern struct {
	Pattern Pattern // The pattern after the star (can be wildcard or capture)

	Span lexer.Span
}

func (sp *StarPattern) isPattern() {}

func (sp *StarPattern) GetSpan() lexer.Span {
	return sp.Span
}

func (sp *StarPattern) String() string {
	return fmt.Sprintf("*%s", sp.Pattern.String())
}

func (sp *StarPattern) Accept(visitor Visitor) {
	visitor.VisitStarPattern(sp)
}

// MappingPattern represents dictionary patterns
type MappingPattern struct {
	Pairs      []MappingPatternPair // Key-value pattern pairs
	DoubleStar Pattern              // Optional **pattern for rest
	HasRest    bool                 // Whether there's a **pattern

	Span lexer.Span
}

func (mp *MappingPattern) isPattern() {}

func (mp *MappingPattern) GetSpan() lexer.Span {
	return mp.Span
}

func (mp *MappingPattern) String() string {
	var pairs []string
	for _, pair := range mp.Pairs {
		pairs = append(pairs, pair.String())
	}
	result := strings.Join(pairs, ", ")
	if mp.HasRest {
		if len(pairs) > 0 {
			result += ", "
		}
		result += fmt.Sprintf("**%s", mp.DoubleStar.String())
	}
	return fmt.Sprintf("{%s}", result)
}

func (mp *MappingPattern) Accept(visitor Visitor) {
	visitor.VisitMappingPattern(mp)
}

// MappingPatternPair represents a key-value pair in a mapping pattern
type MappingPatternPair struct {
	Key     Expr    // The key (literal or attribute)
	Pattern Pattern // The pattern to match the value

	Span lexer.Span
}

func (mpp *MappingPatternPair) String() string {
	return fmt.Sprintf("%s: %s", mpp.Key.String(), mpp.Pattern.String())
}

// ClassPattern represents class patterns for matching objects
type ClassPattern struct {
	Class       Expr             // The class name (possibly dotted)
	Patterns    []Pattern        // Positional patterns
	KwdPatterns []KwdPatternPair // Keyword patterns

	Span lexer.Span
}

func (cp *ClassPattern) isPattern() {}

func (cp *ClassPattern) GetSpan() lexer.Span {
	return cp.Span
}

func (cp *ClassPattern) String() string {
	var parts []string
	for _, p := range cp.Patterns {
		parts = append(parts, p.String())
	}
	for _, kp := range cp.KwdPatterns {
		parts = append(parts, kp.String())
	}
	return fmt.Sprintf("%s(%s)", cp.Class.String(), strings.Join(parts, ", "))
}

func (cp *ClassPattern) Accept(visitor Visitor) {
	visitor.VisitClassPattern(cp)
}

// KwdPatternPair represents a keyword pattern pair (name=pattern)
type KwdPatternPair struct {
	Name    *Name   // The keyword name
	Pattern Pattern // The pattern

	Span lexer.Span
}

func (kpp *KwdPatternPair) String() string {
	return fmt.Sprintf("%s=%s", kpp.Name.String(), kpp.Pattern.String())
}

// AsPattern represents patterns with 'as' binding (pattern as name)
type AsPattern struct {
	Pattern Pattern // The pattern to match
	Target  *Name   // The name to bind to

	Span lexer.Span
}

func (ap *AsPattern) isPattern() {}

func (ap *AsPattern) GetSpan() lexer.Span {
	return ap.Span
}

func (ap *AsPattern) String() string {
	return fmt.Sprintf("%s as %s", ap.Pattern.String(), ap.Target.String())
}

func (ap *AsPattern) Accept(visitor Visitor) {
	visitor.VisitAsPattern(ap)
}

// OrPattern represents alternative patterns (pattern1 | pattern2 | ...)
type OrPattern struct {
	Patterns []Pattern // Alternative patterns

	Span lexer.Span
}

func (op *OrPattern) isPattern() {}

func (op *OrPattern) GetSpan() lexer.Span {
	return op.Span
}

func (op *OrPattern) String() string {
	var patterns []string
	for _, p := range op.Patterns {
		patterns = append(patterns, p.String())
	}
	return strings.Join(patterns, " | ")
}

func (op *OrPattern) Accept(visitor Visitor) {
	visitor.VisitOrPattern(op)
}
