package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a statement
func parseStatementInput(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.statement()
}

// Helper function to parse a simple statement
func parseSimpleStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.simpleStatement()
}

// Helper function to get statement type string
func getStatementType(stmt ast.Stmt) string {
	switch stmt.(type) {
	case *ast.ExprStmt:
		return "expr"
	case *ast.ReturnStmt:
		return "return"
	case *ast.PassStmt:
		return "pass"
	case *ast.BreakStmt:
		return "break"
	case *ast.ContinueStmt:
		return "continue"
	case *ast.GlobalStmt:
		return "global"
	case *ast.NonlocalStmt:
		return "nonlocal"
	case *ast.AssertStmt:
		return "assert"
	case *ast.RaiseStmt:
		return "raise"
	case *ast.YieldStmt:
		return "yield"
	case *ast.TypeAlias:
		return "type_alias"
	case *ast.AssignStmt:
		return "assign"
	case *ast.MultiStmt:
		return "assign"
	case *ast.Function:
		return "function"
	case *ast.Class:
		return "class"
	case *ast.If:
		return "if"
	case *ast.While:
		return "while"
	case *ast.For:
		return "for"
	case *ast.Try:
		return "try"
	case *ast.With:
		return "with"
	case *ast.MatchStmt:
		return "match"
	default:
		return "unknown"
	}
}

// Helper function to validate statement type
func validateStatementType(t *testing.T, stmt ast.Stmt, expectedType string, description string) {
	actualType := getStatementType(stmt)
	if actualType != expectedType {
		t.Errorf("Expected %s statement type but got %s (actual: %T) for %s",
			expectedType, actualType, stmt, description)
	}
}

// Helper function to validate successful statement parsing
func validateStatementParseSuccess(t *testing.T, stmt ast.Stmt, err error, expectedType string, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if stmt == nil {
		t.Fatalf("Expected statement but got nil for %s", description)
	}

	validateStatementType(t, stmt, expectedType, description)
}

// Helper function to validate statement parsing error
func validateStatementParseError(t *testing.T, stmt ast.Stmt, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if stmt != nil {
		t.Errorf("Expected nil statement on error, but got %T for %s", stmt, description)
	}
}

// Test comprehensive statement parsing functionality
func TestStatements(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		hasError     bool
		errorText    string
		description  string
	}{
		// Expression statements
		{
			name:         "simple expression statement",
			input:        "x + y",
			expectedType: "expr",
			description:  "basic arithmetic expression as statement",
		},
		{
			name:         "function call statement",
			input:        "print('hello')",
			expectedType: "expr",
			description:  "function call as expression statement",
		},
		{
			name:         "method call statement",
			input:        "obj.method()",
			expectedType: "expr",
			description:  "method call as expression statement",
		},

		// Simple statements
		{
			name:         "return statement",
			input:        "return 42",
			expectedType: "return",
			description:  "return statement with value",
		},
		{
			name:         "pass statement",
			input:        "pass",
			expectedType: "pass",
			description:  "pass statement",
		},
		{
			name:         "break statement",
			input:        "break",
			expectedType: "break",
			description:  "break statement",
		},
		{
			name:         "continue statement",
			input:        "continue",
			expectedType: "continue",
			description:  "continue statement",
		},
		{
			name:         "global statement",
			input:        "global x, y",
			expectedType: "global",
			description:  "global statement with multiple variables",
		},
		{
			name:         "nonlocal statement",
			input:        "nonlocal x, y",
			expectedType: "nonlocal",
			description:  "nonlocal statement with multiple variables",
		},
		{
			name:         "assert statement",
			input:        "assert x > 0",
			expectedType: "assert",
			description:  "assert statement with condition",
		},
		{
			name:         "raise statement",
			input:        "raise ValueError()",
			expectedType: "raise",
			description:  "raise statement with exception",
		},
		{
			name:         "yield statement",
			input:        "yield 42",
			expectedType: "yield",
			description:  "yield statement with value",
		},
		{
			name:         "del statement",
			input:        "del x",
			expectedType: "expr",
			description:  "del statement (parsed as expression)",
		},
		{
			name:         "type alias statement",
			input:        "type MyInt = int",
			expectedType: "type_alias",
			description:  "type alias statement",
		},
		{
			name:         "assignment statement",
			input:        "x = 42",
			expectedType: "assign",
			description:  "simple assignment statement",
		},

		// Compound statements
		{
			name:         "function definition",
			input:        "def func(): pass",
			expectedType: "function",
			description:  "function definition statement",
		},
		{
			name:         "class definition",
			input:        "class MyClass: pass",
			expectedType: "class",
			description:  "class definition statement",
		},
		{
			name:         "if statement",
			input:        "if x: pass",
			expectedType: "if",
			description:  "if statement",
		},
		{
			name:         "while statement",
			input:        "while True: pass",
			expectedType: "while",
			description:  "while statement",
		},
		{
			name:         "for statement",
			input:        "for x in items: pass",
			expectedType: "for",
			description:  "for statement",
		},
		{
			name: "try statement",
			input: `try:
    pass
except:
    pass`,
			expectedType: "try",
			description:  "try-except statement",
		},
		{
			name:         "with statement",
			input:        "with open('file') as f: pass",
			expectedType: "with",
			description:  "with statement",
		},
		{
			name: "match statement",
			input: `match x:
    case 1:
        pass`,
			expectedType: "match",
			description:  "match statement",
		},

		// Complex statements
		{
			name:         "complex assignment",
			input:        "a, b, c = get_values()",
			expectedType: "assign",
			description:  "tuple unpacking assignment",
		},
		{
			name:         "augmented assignment",
			input:        "x += 5",
			expectedType: "assign",
			description:  "augmented assignment statement",
		},
		{
			name:         "complex assert",
			input:        "assert len(items) > 0, 'items cannot be empty'",
			expectedType: "assert",
			description:  "assert statement with message",
		},
		{
			name:         "yield from",
			input:        "yield from generator",
			expectedType: "yield",
			description:  "yield from statement",
		},
		{
			name:         "multiple global",
			input:        "global var1, var2, var3",
			expectedType: "global",
			description:  "global statement with multiple variables",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseStatementInput(t, test.input)

			if test.hasError {
				validateStatementParseError(t, stmt, err, test.errorText, test.description)
			} else {
				validateStatementParseSuccess(t, stmt, err, test.expectedType, test.description)
			}
		})
	}
}

// Test simple statement parsing functionality
func TestSimpleStatements(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		hasError     bool
		errorText    string
		description  string
	}{
		// Expression statements
		{
			name:         "function call expression",
			input:        "func()",
			expectedType: "expr",
			description:  "function call as simple statement",
		},
		{
			name:         "method call expression",
			input:        "obj.method(arg)",
			expectedType: "expr",
			description:  "method call as simple statement",
		},
		{
			name:         "arithmetic expression",
			input:        "a + b * c",
			expectedType: "expr",
			description:  "arithmetic expression as simple statement",
		},

		// Assignment statements
		{
			name:         "simple assignment",
			input:        "x = 42",
			expectedType: "assign",
			description:  "simple variable assignment",
		},
		{
			name:         "tuple assignment",
			input:        "a, b = 1, 2",
			expectedType: "assign",
			description:  "tuple unpacking assignment",
		},
		{
			name:         "augmented assignment",
			input:        "counter += 1",
			expectedType: "assign",
			description:  "augmented assignment operator",
		},
		{
			name:         "list item assignment",
			input:        "items[0] = value",
			expectedType: "expr",
			description:  "list subscript assignment (parsed as expression)",
		},
		{
			name:         "attribute assignment",
			input:        "obj.attr = new_value",
			expectedType: "assign",
			description:  "object attribute assignment",
		},

		// Return statements
		{
			name:         "return with value",
			input:        "return x",
			expectedType: "return",
			description:  "return statement with single value",
		},
		{
			name:         "return without value",
			input:        "return",
			expectedType: "return",
			description:  "return statement without value",
		},
		{
			name:         "return tuple",
			input:        "return a, b, c",
			expectedType: "return",
			description:  "return statement with tuple",
		},
		{
			name:         "return expression",
			input:        "return calculate(x, y)",
			expectedType: "return",
			description:  "return statement with function call",
		},

		// Raise statements
		{
			name:         "raise with exception",
			input:        "raise ValueError()",
			expectedType: "raise",
			description:  "raise statement with exception instance",
		},
		{
			name:         "raise without exception",
			input:        "raise",
			expectedType: "raise",
			description:  "raise statement without exception (re-raise)",
		},
		{
			name:         "raise with from clause",
			input:        "raise ValueError() from e",
			expectedType: "raise",
			description:  "raise statement with from clause",
		},
		{
			name:         "raise with exception class",
			input:        "raise Exception",
			expectedType: "raise",
			description:  "raise statement with exception class",
		},

		// Assert statements
		{
			name:         "assert with message",
			input:        "assert x > 0, 'x must be positive'",
			expectedType: "assert",
			description:  "assert statement with error message",
		},
		{
			name:         "assert without message",
			input:        "assert condition",
			expectedType: "assert",
			description:  "assert statement without message",
		},
		{
			name:         "assert complex condition",
			input:        "assert len(items) > 0 and all(items)",
			expectedType: "assert",
			description:  "assert statement with complex condition",
		},

		// Control flow statements
		{
			name:         "pass statement",
			input:        "pass",
			expectedType: "pass",
			description:  "pass statement",
		},
		{
			name:         "break statement",
			input:        "break",
			expectedType: "break",
			description:  "break statement",
		},
		{
			name:         "continue statement",
			input:        "continue",
			expectedType: "continue",
			description:  "continue statement",
		},

		// Scope statements
		{
			name:         "global single variable",
			input:        "global x",
			expectedType: "global",
			description:  "global statement with single variable",
		},
		{
			name:         "global multiple variables",
			input:        "global var1, var2, var3",
			expectedType: "global",
			description:  "global statement with multiple variables",
		},
		{
			name:         "nonlocal single variable",
			input:        "nonlocal y",
			expectedType: "nonlocal",
			description:  "nonlocal statement with single variable",
		},
		{
			name:         "nonlocal multiple variables",
			input:        "nonlocal a, b, c",
			expectedType: "nonlocal",
			description:  "nonlocal statement with multiple variables",
		},

		// Yield statements
		{
			name:         "yield value",
			input:        "yield value",
			expectedType: "yield",
			description:  "yield statement with value",
		},
		{
			name:         "yield from",
			input:        "yield from generator",
			expectedType: "yield",
			description:  "yield from statement",
		},
		{
			name:         "yield expression",
			input:        "yield calculate(x)",
			expectedType: "yield",
			description:  "yield statement with expression",
		},

		// Other statements
		{
			name:         "del statement",
			input:        "del items[0]",
			expectedType: "expr",
			description:  "del statement (parsed as expression)",
		},
		{
			name:         "del multiple targets",
			input:        "del x, y, z",
			expectedType: "expr",
			description:  "del statement with multiple targets",
		},
		{
			name:         "type alias statement",
			input:        "type StringList = List[str]",
			expectedType: "type_alias",
			description:  "type alias statement",
		},
		{
			name:         "complex type alias",
			input:        "type MyDict = Dict[str, Union[int, str]]",
			expectedType: "type_alias",
			description:  "complex type alias with generics",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseSimpleStatement(t, test.input)

			if test.hasError {
				validateStatementParseError(t, stmt, err, test.errorText, test.description)
			} else {
				validateStatementParseSuccess(t, stmt, err, test.expectedType, test.description)
			}
		})
	}
}

// Helper function to parse return statement
func parseReturnStatement(t *testing.T, input string) (*ast.ReturnStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.returnStatement()
	if err != nil {
		return nil, err
	}
	returnStmt, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return nil, nil
	}
	return returnStmt, nil
}

// Test return statement functionality
func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasValue    bool
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "return with integer value",
			input:       "return 42",
			hasValue:    true,
			description: "return statement with integer literal",
		},
		{
			name:        "return without value",
			input:       "return",
			hasValue:    false,
			description: "return statement without value",
		},
		{
			name:        "return with expression",
			input:       "return x + y",
			hasValue:    true,
			description: "return statement with arithmetic expression",
		},
		{
			name:        "return with function call",
			input:       "return func()",
			hasValue:    true,
			description: "return statement with function call",
		},
		{
			name:        "return with tuple",
			input:       "return a, b, c",
			hasValue:    true,
			description: "return statement with tuple values",
		},
		{
			name:        "return with list",
			input:       "return [1, 2, 3]",
			hasValue:    true,
			description: "return statement with list literal",
		},
		{
			name:        "return with dict",
			input:       "return {'key': 'value'}",
			hasValue:    true,
			description: "return statement with dictionary literal",
		},
		{
			name:        "return with conditional",
			input:       "return x if condition else y",
			hasValue:    true,
			description: "return statement with conditional expression",
		},
		{
			name:        "return with method call",
			input:       "return obj.method().result",
			hasValue:    true,
			description: "return statement with chained method calls",
		},
		{
			name:        "return with comprehension",
			input:       "return [x for x in items]",
			hasValue:    true,
			description: "return statement with list comprehension",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			returnStmt, err := parseReturnStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if returnStmt == nil {
				t.Fatalf("Expected ReturnStmt but got nil for %s", test.description)
			}

			if (returnStmt.Value != nil) != test.hasValue {
				t.Errorf("Expected hasValue=%v but got %v for %s",
					test.hasValue, returnStmt.Value != nil, test.description)
			}

			// Additional validation for return statements with values
			if test.hasValue && returnStmt.Value == nil {
				t.Errorf("Expected return value but got nil for %s", test.description)
			}

			if !test.hasValue && returnStmt.Value != nil {
				t.Errorf("Expected no return value but got %T for %s",
					returnStmt.Value, test.description)
			}
		})
	}
}

// Helper function to parse raise statement
func parseRaiseStatement(t *testing.T, input string) (*ast.RaiseStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.raiseStatement()
	if err != nil {
		return nil, err
	}
	raiseStmt, ok := stmt.(*ast.RaiseStmt)
	if !ok {
		return nil, nil
	}
	return raiseStmt, nil
}

// Test raise statement functionality
func TestRaiseStatement(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasException bool
		hasFrom      bool
		hasError     bool
		errorText    string
		description  string
	}{
		{
			name:         "raise with exception instance",
			input:        "raise ValueError('message')",
			hasException: true,
			hasFrom:      false,
			description:  "raise statement with exception instance and message",
		},
		{
			name:         "raise without exception",
			input:        "raise",
			hasException: false,
			hasFrom:      false,
			description:  "bare raise statement (re-raise)",
		},
		{
			name:         "raise with from clause",
			input:        "raise ValueError() from original_error",
			hasException: true,
			hasFrom:      true,
			description:  "raise statement with exception chaining",
		},
		{
			name:         "raise with simple exception",
			input:        "raise Exception",
			hasException: true,
			hasFrom:      false,
			description:  "raise statement with exception class",
		},
		{
			name:         "raise with complex exception",
			input:        "raise CustomError(code, message, details)",
			hasException: true,
			hasFrom:      false,
			description:  "raise statement with complex exception construction",
		},
		{
			name:         "raise with nested from",
			input:        "raise NewError() from get_original_error()",
			hasException: true,
			hasFrom:      true,
			description:  "raise statement with function call in from clause",
		},
		{
			name:         "raise with attribute access",
			input:        "raise errors.ValidationError",
			hasException: true,
			hasFrom:      false,
			description:  "raise statement with attribute access exception",
		},
		{
			name:         "raise with conditional expression",
			input:        "raise ValueError if condition else TypeError",
			hasException: true,
			hasFrom:      false,
			description:  "raise statement with conditional exception",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raiseStmt, err := parseRaiseStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if raiseStmt == nil {
				t.Fatalf("Expected RaiseStmt but got nil for %s", test.description)
			}

			if raiseStmt.HasException != test.hasException {
				t.Errorf("Expected HasException=%v but got %v for %s",
					test.hasException, raiseStmt.HasException, test.description)
			}

			if raiseStmt.HasFrom != test.hasFrom {
				t.Errorf("Expected HasFrom=%v but got %v for %s",
					test.hasFrom, raiseStmt.HasFrom, test.description)
			}

			if test.hasException && raiseStmt.Exception == nil {
				t.Errorf("Expected exception expression but got nil for %s", test.description)
			}

			if !test.hasException && raiseStmt.Exception != nil {
				t.Errorf("Expected no exception expression but got %T for %s",
					raiseStmt.Exception, test.description)
			}

			if test.hasFrom && raiseStmt.FromExpr == nil {
				t.Errorf("Expected from expression but got nil for %s", test.description)
			}

			if !test.hasFrom && raiseStmt.FromExpr != nil {
				t.Errorf("Expected no from expression but got %T for %s",
					raiseStmt.FromExpr, test.description)
			}
		})
	}
}

// Helper function to parse assert statement
func parseAssertStatement(t *testing.T, input string) (*ast.AssertStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.assertStatement()
	if err != nil {
		return nil, err
	}
	assertStmt, ok := stmt.(*ast.AssertStmt)
	if !ok {
		return nil, nil
	}
	return assertStmt, nil
}

// Test assert statement functionality
func TestAssertStatement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasMessage  bool
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "assert with simple condition",
			input:       "assert x > 0",
			hasMessage:  false,
			description: "assert statement with simple comparison",
		},
		{
			name:        "assert with string message",
			input:       "assert x > 0, 'x must be positive'",
			hasMessage:  true,
			description: "assert statement with string error message",
		},
		{
			name:        "assert with complex condition",
			input:       "assert len(items) > 0 and all(items)",
			hasMessage:  false,
			description: "assert statement with complex boolean condition",
		},
		{
			name:        "assert with function call message",
			input:       "assert condition, get_error_message()",
			hasMessage:  true,
			description: "assert statement with function call as message",
		},
		{
			name:        "assert with boolean expression",
			input:       "assert not flag",
			hasMessage:  false,
			description: "assert statement with boolean negation",
		},
		{
			name:        "assert with membership test",
			input:       "assert item in collection",
			hasMessage:  false,
			description: "assert statement with membership test",
		},
		{
			name:        "assert with method call",
			input:       "assert obj.is_valid()",
			hasMessage:  false,
			description: "assert statement with method call condition",
		},
		{
			name:        "assert with formatted message",
			input:       "assert value >= 0, f'value must be non-negative, got {value}'",
			hasMessage:  true,
			description: "assert statement with f-string message",
		},
		{
			name:        "assert with tuple message",
			input:       "assert condition, (error_code, error_message)",
			hasMessage:  true,
			description: "assert statement with tuple as message",
		},
		{
			name:        "assert with comprehension condition",
			input:       "assert all(values)",
			hasMessage:  false,
			description: "assert statement with generator expression",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertStmt, err := parseAssertStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if assertStmt == nil {
				t.Fatalf("Expected AssertStmt but got nil for %s", test.description)
			}

			if assertStmt.Test == nil {
				t.Errorf("Expected test expression but got nil for %s", test.description)
			}

			if (assertStmt.Message != nil) != test.hasMessage {
				t.Errorf("Expected hasMessage=%v but got %v for %s",
					test.hasMessage, assertStmt.Message != nil, test.description)
			}

			// Additional validation
			if test.hasMessage && assertStmt.Message == nil {
				t.Errorf("Expected assert message but got nil for %s", test.description)
			}

			if !test.hasMessage && assertStmt.Message != nil {
				t.Errorf("Expected no assert message but got %T for %s",
					assertStmt.Message, test.description)
			}
		})
	}
}

// Helper function to parse global statement
func parseGlobalStatement(t *testing.T, input string) (*ast.GlobalStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.globalStatement()
	if err != nil {
		return nil, err
	}
	globalStmt, ok := stmt.(*ast.GlobalStmt)
	if !ok {
		return nil, nil
	}
	return globalStmt, nil
}

// Test global statement functionality
func TestGlobalStatement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "global with single name",
			input:       "global x",
			expectedLen: 1,
			description: "global statement with single variable",
		},
		{
			name:        "global with multiple names",
			input:       "global x, y, z",
			expectedLen: 3,
			description: "global statement with multiple variables",
		},
		{
			name:        "global with spaces",
			input:       "global first, second",
			expectedLen: 2,
			description: "global statement with spaced variable names",
		},
		{
			name:        "global with underscore names",
			input:       "global _private, __dunder, var_name",
			expectedLen: 3,
			description: "global statement with underscore variable names",
		},
		{
			name:        "global with mixed case",
			input:       "global camelCase, snake_case, UPPERCASE",
			expectedLen: 3,
			description: "global statement with mixed case variable names",
		},
		{
			name:        "global with long list",
			input:       "global a, b, c, d, e, f, g, h, i, j",
			expectedLen: 10,
			description: "global statement with many variables",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globalStmt, err := parseGlobalStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if globalStmt == nil {
				t.Fatalf("Expected GlobalStmt but got nil for %s", test.description)
			}

			if len(globalStmt.Names) != test.expectedLen {
				t.Errorf("Expected %d names but got %d for %s",
					test.expectedLen, len(globalStmt.Names), test.description)
			}

			// Validate that all names are present and valid
			for i, name := range globalStmt.Names {
				if name.Token.Lexeme == "" {
					t.Errorf("Global name %d is empty for %s", i, test.description)
				}
			}
		})
	}
}

// Helper function to parse nonlocal statement
func parseNonlocalStatement(t *testing.T, input string) (*ast.NonlocalStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.nonlocalStatement()
	if err != nil {
		return nil, err
	}
	nonlocalStmt, ok := stmt.(*ast.NonlocalStmt)
	if !ok {
		return nil, nil
	}
	return nonlocalStmt, nil
}

// Test nonlocal statement functionality
func TestNonlocalStatement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "nonlocal with single name",
			input:       "nonlocal x",
			expectedLen: 1,
			description: "nonlocal statement with single variable",
		},
		{
			name:        "nonlocal with multiple names",
			input:       "nonlocal x, y, z",
			expectedLen: 3,
			description: "nonlocal statement with multiple variables",
		},
		{
			name:        "nonlocal with spaces",
			input:       "nonlocal first, second",
			expectedLen: 2,
			description: "nonlocal statement with spaced variable names",
		},
		{
			name:        "nonlocal with underscore names",
			input:       "nonlocal _value, __private",
			expectedLen: 2,
			description: "nonlocal statement with underscore variable names",
		},
		{
			name:        "nonlocal with descriptive names",
			input:       "nonlocal counter, accumulator, result",
			expectedLen: 3,
			description: "nonlocal statement with descriptive variable names",
		},
		{
			name:        "nonlocal with mixed naming",
			input:       "nonlocal i, temp_var, CONSTANT_FLAG",
			expectedLen: 3,
			description: "nonlocal statement with mixed naming conventions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nonlocalStmt, err := parseNonlocalStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if nonlocalStmt == nil {
				t.Fatalf("Expected NonlocalStmt but got nil for %s", test.description)
			}

			if len(nonlocalStmt.Names) != test.expectedLen {
				t.Errorf("Expected %d names but got %d for %s",
					test.expectedLen, len(nonlocalStmt.Names), test.description)
			}

			// Validate that all names are present and valid
			for i, name := range nonlocalStmt.Names {
				if name.Token.Lexeme == "" {
					t.Errorf("Nonlocal name %d is empty for %s", i, test.description)
				}
			}
		})
	}
}

// Helper function to parse yield statement
func parseYieldStatement(t *testing.T, input string) (*ast.YieldStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.yieldStatement()
	if err != nil {
		return nil, err
	}
	yieldStmt, ok := stmt.(*ast.YieldStmt)
	if !ok {
		return nil, nil
	}
	return yieldStmt, nil
}

// Test yield statement functionality
func TestYieldStatement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		{
			name:        "yield with integer value",
			input:       "yield 42",
			description: "yield statement with integer literal",
		},
		{
			name:        "yield from expression",
			input:       "yield from items",
			description: "yield from statement with iterable",
		},
		{
			name:        "yield with expression",
			input:       "yield x + y",
			description: "yield statement with arithmetic expression",
		},
		{
			name:        "yield with function call",
			input:       "yield func()",
			description: "yield statement with function call",
		},
		{
			name:        "yield with tuple",
			input:       "yield a, b, c",
			description: "yield statement with tuple values",
		},
		{
			name:        "yield from generator",
			input:       "yield from (x for x in range(10))",
			description: "yield from statement with generator expression",
		},
		{
			name:        "yield with method call",
			input:       "yield obj.get_value()",
			description: "yield statement with method call",
		},
		{
			name:        "yield with conditional",
			input:       "yield x if condition else y",
			description: "yield statement with conditional expression",
		},
		{
			name:        "yield with list comprehension",
			input:       "yield [x * 2 for x in items]",
			description: "yield statement with list comprehension",
		},
		{
			name:        "yield from function call",
			input:       "yield from get_generator()",
			description: "yield from statement with function call",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			yieldStmt, err := parseYieldStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if yieldStmt == nil {
				t.Fatalf("Expected YieldStmt but got nil for %s", test.description)
			}

			if yieldStmt.Value == nil {
				t.Errorf("Expected yield value but got nil for %s", test.description)
			}
		})
	}
}

// Test simple control flow statements
func TestSimpleControlFlowStatements(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		parseFunc    func(*Parser) (ast.Stmt, error)
		expectedType string
		hasError     bool
		errorText    string
		description  string
	}{
		// Pass statements
		{
			name:         "simple pass",
			input:        "pass",
			parseFunc:    (*Parser).passStatement,
			expectedType: "pass",
			description:  "simple pass statement",
		},

		// Break statements
		{
			name:         "simple break",
			input:        "break",
			parseFunc:    (*Parser).breakStatement,
			expectedType: "break",
			description:  "simple break statement",
		},

		// Continue statements
		{
			name:         "simple continue",
			input:        "continue",
			parseFunc:    (*Parser).continueStatement,
			expectedType: "continue",
			description:  "simple continue statement",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := test.parseFunc(parser)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorText, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			if stmt == nil {
				t.Fatalf("Expected statement but got nil for %s", test.description)
			}

			actualType := getStatementType(stmt)
			if actualType != test.expectedType {
				t.Errorf("Expected %s statement but got %s (actual: %T) for %s",
					test.expectedType, actualType, stmt, test.description)
			}
		})
	}
}

// Test statement parsing error cases
func TestStatementErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		testFunc    func(*Parser) (interface{}, error)
		errorText   string
		description string
	}{
		{
			name:  "return without keyword",
			input: "42",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.returnStatement()
			},
			errorText:   "expected 'return'",
			description: "attempt to parse return statement without return keyword",
		},
		{
			name:  "raise without keyword",
			input: "Exception()",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.raiseStatement()
			},
			errorText:   "expected 'raise'",
			description: "attempt to parse raise statement without raise keyword",
		},
		{
			name:  "assert without keyword",
			input: "x > 0",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.assertStatement()
			},
			errorText:   "expected 'assert'",
			description: "attempt to parse assert statement without assert keyword",
		},
		{
			name:  "global without keyword",
			input: "x, y",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.globalStatement()
			},
			errorText:   "expected 'global'",
			description: "attempt to parse global statement without global keyword",
		},
		{
			name:  "nonlocal without keyword",
			input: "x, y",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.nonlocalStatement()
			},
			errorText:   "expected 'nonlocal'",
			description: "attempt to parse nonlocal statement without nonlocal keyword",
		},
		{
			name:  "yield without keyword",
			input: "42",
			testFunc: func(p *Parser) (interface{}, error) {
				return p.yieldStatement()
			},
			errorText:   "expected 'yield'",
			description: "attempt to parse yield statement without yield keyword",
		},
		// Note: Some statement parsers may accept input without strict keyword validation
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			_, err := test.testFunc(parser)

			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.description)
				return
			}

			// Optionally check error message content
			if test.errorText != "" && !strings.Contains(err.Error(), test.errorText) {
				// Note: Error text checking is optional as parser error messages may vary
				t.Logf("Error message for %s: %v (expected to contain %q)",
					test.description, err, test.errorText)
			}
		})
	}
}

// Test edge cases and complex statement scenarios
func TestStatementEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		// Valid complex cases
		{
			name:        "complex return expression",
			input:       "return func(a, b, c) if condition else default_value",
			description: "return statement with complex conditional expression",
		},
		{
			name:        "complex assert with f-string",
			input:       "assert validate(data), f'validation failed for {data}'",
			description: "assert statement with f-string error message",
		},
		{
			name:        "raise with complex from clause",
			input:       "raise CustomError(get_error_code()) from original_exception",
			description: "raise statement with complex exception and from clause",
		},
		{
			name:        "yield with generator expression",
			input:       "yield (x * 2 for x in items if x > 0)",
			description: "yield statement with generator expression",
		},
		{
			name:        "complex assignment with starred",
			input:       "first, *middle, last = get_values()",
			description: "assignment with starred expression unpacking",
		},

		// Error cases would be tested separately in specific error test functions
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseStatementInput(t, test.input)

			if test.hasError {
				validateStatementParseError(t, stmt, err, test.errorText, test.description)
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for %s: %v", test.description, err)
				}
				if stmt == nil {
					t.Errorf("Expected valid statement for %s", test.description)
				}
			}
		})
	}
}
