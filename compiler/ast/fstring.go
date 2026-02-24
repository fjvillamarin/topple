package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
)

// FString represents an f-string literal: f"text {expr} more text"
type FString struct {
	Parts []FStringPart // The parts of the f-string (middle text and replacement fields)

	Span lexer.Span
}

func (f *FString) isExpr() {}

func (f *FString) GetSpan() lexer.Span {
	return f.Span
}

func (f *FString) String() string {
	var parts []string
	for _, part := range f.Parts {
		parts = append(parts, part.String())
	}
	return fmt.Sprintf("f\"%s\"", strings.Join(parts, ""))
}

func (f *FString) Accept(visitor Visitor) {
	visitor.VisitFString(f)
}

// FStringPart is the interface for parts of an f-string
type FStringPart interface {
	Node
	isFStringPart()
}

// FStringMiddle represents literal text in an f-string
type FStringMiddle struct {
	Value string // The literal text

	Span lexer.Span
}

func (f *FStringMiddle) isFStringPart() {}

func (f *FStringMiddle) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringMiddle) String() string {
	return f.Value
}

func (f *FStringMiddle) Accept(visitor Visitor) {
	visitor.VisitFStringMiddle(f)
}

// FStringReplacementField represents a replacement field in an f-string: {expr!conv:format}
type FStringReplacementField struct {
	Expression Expr               // The expression to evaluate
	Equal      bool               // Whether there's an = after the expression (for debugging)
	Conversion *FStringConversion // Optional conversion (!r, !s, !a)
	FormatSpec *FStringFormatSpec // Optional format specification

	Span lexer.Span
}

func (f *FStringReplacementField) isFStringPart() {}

func (f *FStringReplacementField) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringReplacementField) String() string {
	result := "{" + f.Expression.String()
	if f.Equal {
		result += "="
	}
	if f.Conversion != nil {
		result += f.Conversion.String()
	}
	if f.FormatSpec != nil {
		result += f.FormatSpec.String()
	}
	result += "}"
	return result
}

func (f *FStringReplacementField) Accept(visitor Visitor) {
	visitor.VisitFStringReplacementField(f)
}

// FStringConversion represents a conversion in an f-string: !r, !s, or !a
type FStringConversion struct {
	Type string // "r", "s", or "a"

	Span lexer.Span
}

func (f *FStringConversion) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringConversion) String() string {
	return "!" + f.Type
}

func (f *FStringConversion) Accept(visitor Visitor) {
	visitor.VisitFStringConversion(f)
}

// FStringFormatSpec represents a format specification in an f-string: :format
type FStringFormatSpec struct {
	Spec []FStringFormatPart // The format specification parts

	Span lexer.Span
}

func (f *FStringFormatSpec) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringFormatSpec) String() string {
	var parts []string
	for _, part := range f.Spec {
		parts = append(parts, part.String())
	}
	return ":" + strings.Join(parts, "")
}

func (f *FStringFormatSpec) Accept(visitor Visitor) {
	visitor.VisitFStringFormatSpec(f)
}

// FStringFormatPart is the interface for parts of an f-string format specification
type FStringFormatPart interface {
	Node
	isFStringFormatPart()
}

// FStringFormatMiddle represents literal text in a format specification
type FStringFormatMiddle struct {
	Value string // The literal text

	Span lexer.Span
}

func (f *FStringFormatMiddle) isFStringFormatPart() {}

func (f *FStringFormatMiddle) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringFormatMiddle) String() string {
	return f.Value
}

func (f *FStringFormatMiddle) Accept(visitor Visitor) {
	visitor.VisitFStringFormatMiddle(f)
}

// FStringFormatReplacementField represents a replacement field in a format specification
type FStringFormatReplacementField struct {
	Expression Expr               // The expression to evaluate
	Equal      bool               // Whether there's an = after the expression (for debugging)
	Conversion *FStringConversion // Optional conversion (!r, !s, !a)
	FormatSpec *FStringFormatSpec // Optional format specification

	Span lexer.Span
}

func (f *FStringFormatReplacementField) isFStringFormatPart() {}

func (f *FStringFormatReplacementField) GetSpan() lexer.Span {
	return f.Span
}

func (f *FStringFormatReplacementField) String() string {
	result := "{" + f.Expression.String()
	if f.Equal {
		result += "="
	}
	if f.Conversion != nil {
		result += f.Conversion.String()
	}
	if f.FormatSpec != nil {
		result += f.FormatSpec.String()
	}
	result += "}"
	return result
}

func (f *FStringFormatReplacementField) Accept(visitor Visitor) {
	visitor.VisitFStringFormatReplacementField(f)
}
