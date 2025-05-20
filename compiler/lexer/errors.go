package lexer

import "fmt"

// ScannerError is an error that occurs in the scanner.
// Line and Column use the same indexing system as configured in the Scanner.
type ScannerError struct {
	Message string
	Line    int
	Column  int
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("Error: %s at position %s", e.Message, e.Span())
}

// Span returns a string representation of the error's position.
func (e *ScannerError) Span() string {
	return fmt.Sprintf("L%d:%d", e.Line, e.Column)
}

// NewScannerError creates a new ScannerError.
func NewScannerError(message string, line int, column int) *ScannerError {
	return &ScannerError{Message: message, Line: line, Column: column}
}
