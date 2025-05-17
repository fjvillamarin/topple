package compiler

import "fmt"

// ScannerError is an error that occurs in the scanner.
// Line and Column use the same indexing system as configured in the Scanner.
type ScannerError struct {
	Message string
	Line    int
	Column  int
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("Error: %s at position %d:%d", e.Message, e.Line, e.Column)
}

// NewScannerError creates a new ScannerError.
func NewScannerError(message string, line int, column int) *ScannerError {
	return &ScannerError{Message: message, Line: line, Column: column}
}

// ParseError is an error that occurs in the parser.
type ParseError struct {
	Token   Token
	Message string
}

// Error returns a string representation of the ParseError.
func (e *ParseError) Error() string {
	if e.Token.Type == EOF {
		return fmt.Sprintf("at end: %s", e.Message)
	}
	return fmt.Sprintf("at '%s': %s", e.Token.Lexeme, e.Message)
}

// NewParseError creates a new ParseError.
func NewParseError(token Token, message string) *ParseError {
	return &ParseError{Token: token, Message: message}
}

// RuntimeError is an error that occurs in the runtime.
type RuntimeError struct {
	Token   Token
	Message string
}

// Error returns a string representation of the RuntimeError.
func (e *RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error: %s at position %d:%d", e.Message, e.Token.Line, e.Token.Column)
}

// NewRuntimeError creates a new RuntimeError.
func NewRuntimeError(token Token, message string) *RuntimeError {
	return &RuntimeError{Token: token, Message: message}
}

// ReturnValue is a special error type that carries a return value.
// It's not actually an error, but we use Go's error mechanism to
// propagate the return value up the call stack.
type ReturnValue struct {
	Value any
}

func (r *ReturnValue) Error() string {
	return "return value (this error should be caught)"
}

// NewReturnValue creates a new ReturnValue.
func NewReturnValue(value any) *ReturnValue {
	return &ReturnValue{
		Value: value,
	}
}

// ResolverError is an error that occurs in the resolver.
type ResolverError struct {
	Message string
	Line    int
	Column  int
}

func (e *ResolverError) Error() string {
	return fmt.Sprintf("Resolver error: %s at position %d:%d", e.Message, e.Line, e.Column)
}

// NewResolverError creates a new ResolverError.
func NewResolverError(message string, line int, column int) *ResolverError {
	return &ResolverError{Message: message, Line: line, Column: column}
}
