package compiler

import (
	"fmt"
	"topple/compiler/lexer"
)

// RuntimeError is an error that occurs in the runtime.
type RuntimeError struct {
	Token   lexer.Token
	Message string
}

// Error returns a string representation of the RuntimeError.
func (e *RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error: %s at position %s", e.Message, e.Token.Span)
}

// Span returns the span of the token that caused the error.
func (e *RuntimeError) Span() lexer.Span {
	return e.Token.Span
}

// NewRuntimeError creates a new RuntimeError.
func NewRuntimeError(token lexer.Token, message string) *RuntimeError {
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
	return fmt.Sprintf("Resolver error: %s at position %s", e.Message, e.Span())
}

// Span returns a string representation of the error's position.
func (e *ResolverError) Span() string {
	return fmt.Sprintf("L%d:%d", e.Line, e.Column)
}

// NewResolverError creates a new ResolverError.
func NewResolverError(message string, line int, column int) *ResolverError {
	return &ResolverError{Message: message, Line: line, Column: column}
}
