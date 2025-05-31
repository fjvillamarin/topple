package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
)

// fstring parses an f-string literal according to the grammar:
// fstring: FSTRING_START fstring_middle* FSTRING_END
func (p *Parser) fstring() (ast.Expr, error) {
	startToken, err := p.consume(lexer.FStringStart, "expected f-string start")
	if err != nil {
		return nil, err
	}

	var parts []ast.FStringPart

	// Parse f-string middle parts
	for !p.check(lexer.FStringEnd) && !p.isAtEnd() {
		if p.check(lexer.FStringMiddle) {
			// Parse literal text part
			middleToken, err := p.consume(lexer.FStringMiddle, "expected f-string middle")
			if err != nil {
				return nil, err
			}

			middle := &ast.FStringMiddle{
				Value: middleToken.Literal.(string),
				Span:  lexer.Span{Start: middleToken.Start(), End: middleToken.End()},
			}
			parts = append(parts, middle)
		} else if p.check(lexer.LeftBraceF) {
			// Parse replacement field
			replacementField, err := p.fstringReplacementField()
			if err != nil {
				return nil, err
			}
			parts = append(parts, replacementField)
		} else {
			// Unexpected token
			break
		}
	}

	endToken, err := p.consume(lexer.FStringEnd, "expected f-string end")
	if err != nil {
		return nil, err
	}

	return &ast.FString{
		Parts: parts,
		Span:  lexer.Span{Start: startToken.Start(), End: endToken.End()},
	}, nil
}

// fstringReplacementField parses a replacement field in an f-string: {expr!conv:format}
// Grammar: '{' annotated_rhs '='? [fstring_conversion] [fstring_full_format_spec] '}'
func (p *Parser) fstringReplacementField() (ast.FStringPart, error) {
	fmt.Printf("[DEBUG] fstringReplacementField() entry, current token: %d %s at pos %d\n", p.Current, p.peek().Type, p.Current)

	startBrace, err := p.consume(lexer.LeftBraceF, "expected '{'")
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] fstringReplacementField() calling annotatedRhs(), pos %d\n", p.Current)
	// Parse the expression (annotated_rhs) - using existing method from assignment.go
	expr, err := p.annotatedRhs()
	if err != nil {
		return nil, err
	}
	fmt.Printf("[DEBUG] fstringReplacementField() annotatedRhs() returned, pos %d\n", p.Current)

	// Check for optional debugging equals (=)
	var hasEqual bool
	if p.match(lexer.FStringEqual) {
		hasEqual = true
	}

	// Parse optional conversion (!r, !s, !a)
	var conversion *ast.FStringConversion
	if p.match(lexer.FStringConversionStart) {
		// Next should be an identifier (r, s, or a)
		convToken, err := p.consume(lexer.Identifier, "expected conversion type after '!'")
		if err != nil {
			return nil, err
		}

		// Validate conversion type
		convType := convToken.Lexeme
		if convType != "r" && convType != "s" && convType != "a" {
			return nil, p.error(convToken, "invalid conversion type, must be 'r', 's', or 'a'")
		}

		conversion = &ast.FStringConversion{
			Type: convType,
			Span: lexer.Span{Start: convToken.Start(), End: convToken.End()},
		}
	}

	// Parse optional format specification (:format)
	var formatSpec *ast.FStringFormatSpec
	if p.match(lexer.Colon) {
		formatSpec, err = p.fstringFormatSpec()
		if err != nil {
			return nil, err
		}
	}

	endBrace, err := p.consume(lexer.RightBraceF, "expected '}'")
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] fstringReplacementField() exit, pos %d\n", p.Current)
	return &ast.FStringReplacementField{
		Expression: expr,
		Equal:      hasEqual,
		Conversion: conversion,
		FormatSpec: formatSpec,
		Span:       lexer.Span{Start: startBrace.Start(), End: endBrace.End()},
	}, nil
}

// fstringFormatSpec parses format specifications after ':'
// Grammar: fstring_format_spec*
func (p *Parser) fstringFormatSpec() (*ast.FStringFormatSpec, error) {
	var parts []ast.FStringFormatPart

	for !p.check(lexer.RightBraceF) && !p.isAtEnd() {
		if p.check(lexer.FStringMiddle) {
			// Parse literal format text
			middleToken, err := p.consume(lexer.FStringMiddle, "expected format spec text")
			if err != nil {
				return nil, err
			}

			middle := &ast.FStringFormatMiddle{
				Value: middleToken.Literal.(string),
				Span:  lexer.Span{Start: middleToken.Start(), End: middleToken.End()},
			}
			parts = append(parts, middle)
		} else if p.check(lexer.LeftBraceF) {
			// Nested replacement field in format spec
			replacementField, err := p.fstringReplacementField()
			if err != nil {
				return nil, err
			}

			// Convert FStringReplacementField to FStringFormatReplacementField
			formatReplacementField := &ast.FStringFormatReplacementField{
				Expression: replacementField.(*ast.FStringReplacementField).Expression,
				Equal:      replacementField.(*ast.FStringReplacementField).Equal,
				Conversion: replacementField.(*ast.FStringReplacementField).Conversion,
				FormatSpec: replacementField.(*ast.FStringReplacementField).FormatSpec,
				Span:       replacementField.GetSpan(),
			}
			parts = append(parts, formatReplacementField)
		} else {
			break
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &ast.FStringFormatSpec{
		Spec: parts,
		Span: lexer.Span{Start: parts[0].GetSpan().Start, End: parts[len(parts)-1].GetSpan().End},
	}, nil
}
