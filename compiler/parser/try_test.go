package parser

import (
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse a try statement
func parseTryStatement(t *testing.T, input string) (*ast.Try, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.tryStatement()
	if err != nil {
		return nil, err
	}
	tryStmt, ok := stmt.(*ast.Try)
	if !ok {
		return nil, nil
	}
	return tryStmt, nil
}

// Helper function to validate try statement structure
func validateTryStatement(t *testing.T, tryStmt *ast.Try, expectedExcepts, expectedElse, expectedFinally int, description string) {
	if tryStmt == nil {
		t.Fatalf("Expected Try statement but got nil for %s", description)
	}

	if len(tryStmt.Body) == 0 {
		t.Errorf("Try statement missing body for %s", description)
	}

	if len(tryStmt.Excepts) != expectedExcepts {
		t.Errorf("Expected %d except handlers but got %d for %s",
			expectedExcepts, len(tryStmt.Excepts), description)
	}

	if len(tryStmt.Else) != expectedElse {
		t.Errorf("Expected %d else statements but got %d for %s",
			expectedElse, len(tryStmt.Else), description)
	}

	if len(tryStmt.Finally) != expectedFinally {
		t.Errorf("Expected %d finally statements but got %d for %s",
			expectedFinally, len(tryStmt.Finally), description)
	}
}

// Helper function to validate exception handler properties
func validateExceptionHandler(t *testing.T, handler *ast.Except, hasType, hasName bool, description string) {
	if handler == nil {
		t.Fatalf("Expected exception handler but got nil for %s", description)
	}

	actualHasType := handler.Type != nil
	if actualHasType != hasType {
		t.Errorf("Expected hasType=%v but got %v for %s", hasType, actualHasType, description)
	}

	actualHasName := handler.Name != nil
	if actualHasName != hasName {
		t.Errorf("Expected hasName=%v but got %v for %s", hasName, actualHasName, description)
	}

	if len(handler.Body) == 0 {
		t.Errorf("Exception handler missing body for %s", description)
	}
}

// Helper function to validate try parsing success
func validateTryParseSuccess(t *testing.T, tryStmt *ast.Try, err error, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if tryStmt == nil {
		t.Fatalf("Expected Try statement but got nil for %s", description)
	}
}

// Helper function to validate try parsing error
func validateTryParseError(t *testing.T, tryStmt *ast.Try, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if tryStmt != nil {
		t.Errorf("Expected nil Try statement on error, but got %T for %s", tryStmt, description)
	}
}

// Helper function to check exception type structure
func isExceptionTypeTuple(expr ast.Expr) bool {
	_, ok := expr.(*ast.TupleExpr)
	return ok
}

// Helper function to get exception handler count
func getExceptionHandlerCount(handlers []ast.Except) int {
	return len(handlers)
}

// Helper function to check if exception has specific properties
func analyzeExceptionHandler(handler *ast.Except) (hasType, hasName, isTuple bool) {
	hasType = handler.Type != nil
	hasName = handler.Name != nil
	if hasType {
		isTuple = isExceptionTypeTuple(handler.Type)
	}
	return
}

// Test comprehensive try statement parsing functionality
func TestTryStatement(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedExcepts int
		expectedElse    int
		expectedFinally int
		hasError        bool
		errorText       string
		description     string
	}{
		// Basic try-except patterns
		{
			name: "simple try-except",
			input: `try:
    risky_operation()
except:
    handle_error()`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "basic try-except block",
		},
		{
			name: "try with specific exception",
			input: `try:
    int(user_input)
except ValueError:
    print("Invalid number")`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try-except with specific exception type",
		},
		{
			name: "try with exception variable",
			input: `try:
    process()
except Exception as e:
    log_error(e)`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try-except with exception variable binding",
		},
		{
			name: "multiple except blocks",
			input: `try:
    risky_operation()
except ValueError:
    handle_value_error()
except KeyError:
    handle_key_error()
except Exception as e:
    handle_other(e)`,
			expectedExcepts: 3,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try with multiple except handlers",
		},

		// Try with else clause
		{
			name: "try with else",
			input: `try:
    result = compute()
except ComputeError:
    result = None
else:
    cache_result(result)`,
			expectedExcepts: 1,
			expectedElse:    1,
			expectedFinally: 0,
			description:     "try-except-else block",
		},
		{
			name: "try with else and multiple excepts",
			input: `try:
    result = process()
except ValueError:
    result = default_value()
except TypeError:
    result = alternative_value()
else:
    validate_result(result)`,
			expectedExcepts: 2,
			expectedElse:    1,
			expectedFinally: 0,
			description:     "try with multiple except handlers and else",
		},

		// Try with finally clause
		{
			name: "try with finally",
			input: `try:
    file = open("data.txt")
    process(file)
finally:
    file.close()`,
			expectedExcepts: 0,
			expectedElse:    0,
			expectedFinally: 1,
			description:     "try-finally block without except",
		},
		{
			name: "try-except with finally",
			input: `try:
    connection = connect()
    send_data(connection)
except NetworkError:
    log_error("Network failed")
finally:
    close_connection(connection)`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 1,
			description:     "try-except-finally block",
		},

		// Complete try statement
		{
			name: "complete try statement",
			input: `try:
    result = risky_operation()
except ValueError as ve:
    handle_value_error(ve)
except Exception as e:
    handle_generic(e)
else:
    use_result(result)
finally:
    cleanup()`,
			expectedExcepts: 2,
			expectedElse:    1,
			expectedFinally: 1,
			description:     "complete try-except-else-finally block",
		},

		// Nested try statements
		{
			name: "nested try",
			input: `try:
    try:
        inner_operation()
    except InnerError:
        handle_inner()
except OuterError:
    handle_outer()`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "nested try statements",
		},

		// Exception type variations
		{
			name: "try with tuple of exceptions",
			input: `try:
    operation()
except (ValueError, TypeError, KeyError) as e:
    handle_multiple(e)`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try-except with tuple of exception types",
		},
		{
			name: "try with dotted exception name",
			input: `try:
    request()
except requests.exceptions.HTTPError as e:
    handle_http_error(e)`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try-except with dotted exception name",
		},

		// Complex try statements
		{
			name: "try with complex body",
			input: `try:
    for item in items:
        if validate(item):
            process(item)
        else:
            skip(item)
except ProcessError as e:
    log_process_error(e)
except ValidationError as e:
    log_validation_error(e)
else:
    finalize_processing()
finally:
    cleanup_resources()`,
			expectedExcepts: 2,
			expectedElse:    1,
			expectedFinally: 1,
			description:     "try statement with complex body and handlers",
		},
		{
			name: "try with star exception",
			input: `try:
    dynamic_operation()
except *ExceptionGroup as eg:
    handle_exception_group(eg)`,
			expectedExcepts: 1,
			expectedElse:    0,
			expectedFinally: 0,
			description:     "try-except with star exception syntax",
		},

		// Error cases
		{
			name:        "try without except or finally",
			input:       "try:\n    pass",
			hasError:    true,
			errorText:   "expected",
			description: "try without except or finally should fail",
		},
		{
			name: "else without except",
			input: `try:
    pass
else:
    pass`,
			hasError:    true,
			errorText:   "'else' clause requires at least one 'except' clause",
			description: "else clause without except should fail",
		},
		{
			name:        "empty try body",
			input:       "try:\nexcept:\n    pass",
			hasError:    true,
			errorText:   "expected",
			description: "empty try body should fail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)

			if test.hasError {
				validateTryParseError(t, tryStmt, err, test.errorText, test.description)
			} else {
				validateTryParseSuccess(t, tryStmt, err, test.description)
				validateTryStatement(t, tryStmt, test.expectedExcepts, test.expectedElse, test.expectedFinally, test.description)
			}
		})
	}
}

// Test exception handler parsing functionality
func TestExceptionHandlers(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasType     bool
		hasName     bool
		isTuple     bool
		hasError    bool
		errorText   string
		description string
	}{
		// Basic exception handlers
		{
			name: "bare except",
			input: `try:
    pass
except:
    pass`,
			hasType:     false,
			hasName:     false,
			isTuple:     false,
			description: "bare except handler without type or name",
		},
		{
			name: "except with type only",
			input: `try:
    pass
except ValueError:
    pass`,
			hasType:     true,
			hasName:     false,
			isTuple:     false,
			description: "except handler with exception type only",
		},
		{
			name: "except with type and name",
			input: `try:
    pass
except ValueError as e:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     false,
			description: "except handler with exception type and variable",
		},

		// Tuple exception handlers
		{
			name: "except with tuple type",
			input: `try:
    pass
except (ValueError, TypeError):
    pass`,
			hasType:     true,
			hasName:     false,
			isTuple:     true,
			description: "except handler with tuple of exception types",
		},
		{
			name: "except tuple with name",
			input: `try:
    pass
except (IOError, OSError) as e:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     true,
			description: "except handler with tuple of types and variable",
		},
		{
			name: "except complex tuple",
			input: `try:
    pass
except (ValueError, TypeError, KeyError, AttributeError) as e:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     true,
			description: "except handler with large tuple of exception types",
		},

		// Dotted exception names
		{
			name: "except with dotted name",
			input: `try:
    pass
except requests.exceptions.HTTPError as e:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     false,
			description: "except handler with dotted exception name",
		},
		{
			name: "except with complex dotted name",
			input: `try:
    pass
except package.module.submodule.CustomError:
    pass`,
			hasType:     true,
			hasName:     false,
			isTuple:     false,
			description: "except handler with complex dotted exception name",
		},

		// Advanced exception patterns
		{
			name: "except with attribute access",
			input: `try:
    pass
except errors.ValidationError as ve:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     false,
			description: "except handler with attribute access exception",
		},
		{
			name: "except with subscript",
			input: `try:
    pass
except EXCEPTIONS['validation'] as e:
    pass`,
			hasType:     true,
			hasName:     true,
			isTuple:     false,
			description: "except handler with subscript exception reference",
		},

		// Error cases
		{
			name: "invalid except syntax",
			input: `try:
    pass
except as e:
    pass`,
			hasError:    true,
			errorText:   "expected",
			description: "except with 'as' but no exception type",
		},
		{
			name: "missing colon after except",
			input: `try:
    pass
except ValueError
    pass`,
			hasError:    true,
			errorText:   "expected",
			description: "except missing colon",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)

			if test.hasError {
				validateTryParseError(t, tryStmt, err, test.errorText, test.description)
				return
			}

			validateTryParseSuccess(t, tryStmt, err, test.description)

			if len(tryStmt.Excepts) == 0 {
				t.Fatalf("Expected at least one exception handler for %s", test.description)
			}

			handler := &tryStmt.Excepts[0]
			validateExceptionHandler(t, handler, test.hasType, test.hasName, test.description)

			// Additional tuple validation
			if test.hasType && test.isTuple && !isExceptionTypeTuple(handler.Type) {
				t.Errorf("Expected exception type to be tuple for %s", test.description)
			}

			if test.hasType && !test.isTuple && isExceptionTypeTuple(handler.Type) {
				t.Errorf("Expected exception type not to be tuple for %s", test.description)
			}
		})
	}
}

// Test try statement combinations and validity
func TestTryCombinations(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		valid       bool
		errorText   string
		description string
	}{
		// Valid combinations
		{
			name: "try-except",
			input: `try:
    pass
except:
    pass`,
			valid:       true,
			description: "basic try-except combination",
		},
		{
			name: "try-finally",
			input: `try:
    pass
finally:
    pass`,
			valid:       true,
			description: "try-finally combination",
		},
		{
			name: "try-except-else",
			input: `try:
    pass
except:
    pass
else:
    pass`,
			valid:       true,
			description: "try-except-else combination",
		},
		{
			name: "try-except-finally",
			input: `try:
    pass
except:
    pass
finally:
    pass`,
			valid:       true,
			description: "try-except-finally combination",
		},
		{
			name: "try-except-else-finally",
			input: `try:
    pass
except:
    pass
else:
    pass
finally:
    pass`,
			valid:       true,
			description: "complete try-except-else-finally combination",
		},
		{
			name: "multiple excepts with else and finally",
			input: `try:
    pass
except ValueError:
    pass
except TypeError:
    pass
else:
    pass
finally:
    pass`,
			valid:       true,
			description: "multiple except handlers with else and finally",
		},

		// Invalid combinations
		{
			name: "try alone",
			input: `try:
    pass`,
			valid:       false,
			errorText:   "expected",
			description: "try without except or finally",
		},
		{
			name: "else without except",
			input: `try:
    pass
else:
    pass`,
			valid:       false,
			errorText:   "'else' clause requires at least one 'except' clause",
			description: "else clause without except handlers",
		},
		{
			name: "else before except",
			input: `try:
    pass
else:
    pass
except:
    pass`,
			valid:       false,
			errorText:   "'else' clause requires at least one 'except' clause",
			description: "else clause before except handlers",
		},
		{
			name: "finally before except",
			input: `try:
    pass
finally:
    pass
except:
    pass`,
			valid:       false,
			errorText:   "'except' clause cannot appear after 'finally'",
			description: "finally clause before except handlers",
		},
		{
			name: "finally before else",
			input: `try:
    pass
except:
    pass
finally:
    pass
else:
    pass`,
			valid:       false,
			errorText:   "'else' clause cannot appear after 'finally'",
			description: "finally clause before else clause",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)

			if test.valid {
				validateTryParseSuccess(t, tryStmt, err, test.description)
			} else {
				validateTryParseError(t, tryStmt, err, test.errorText, test.description)
			}
		})
	}
}

// Test nested exception handling scenarios
func TestNestedExceptions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		outerExcepts int
		outerFinally int
		hasInnerTry  bool
		description  string
	}{
		{
			name: "simple nested try",
			input: `try:
    try:
        inner_operation()
    except InnerError:
        handle_inner()
except OuterError:
    handle_outer()`,
			outerExcepts: 1,
			outerFinally: 0,
			hasInnerTry:  true,
			description:  "simple nested try-except structure",
		},
		{
			name: "nested try with finally",
			input: `try:
    try:
        risky_inner()
    except InnerError as ie:
        log_inner(ie)
    finally:
        cleanup_inner()
except OuterError as oe:
    handle_outer(oe)
finally:
    cleanup_outer()`,
			outerExcepts: 1,
			outerFinally: 1,
			hasInnerTry:  true,
			description:  "nested try with finally blocks",
		},
		{
			name: "deeply nested try",
			input: `try:
    try:
        try:
            deep_operation()
        except DeepError:
            handle_deep()
    except MiddleError:
        handle_middle()
except OuterError:
    handle_outer()`,
			outerExcepts: 1,
			outerFinally: 0,
			hasInnerTry:  true,
			description:  "deeply nested try statements",
		},
		{
			name: "nested try in except handler",
			input: `try:
    risky_operation()
except MainError as e:
    try:
        recovery_operation()
    except RecoveryError:
        fallback_operation()`,
			outerExcepts: 1,
			outerFinally: 0,
			hasInnerTry:  false, // Inner try is in except block
			description:  "nested try within except handler",
		},
		{
			name: "nested try in finally",
			input: `try:
    main_operation()
except MainError:
    handle_main_error()
finally:
    try:
        cleanup_operation()
    except CleanupError:
        log_cleanup_error()`,
			outerExcepts: 1,
			outerFinally: 1,
			hasInnerTry:  false, // Inner try is in finally block
			description:  "nested try within finally block",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)
			validateTryParseSuccess(t, tryStmt, err, test.description)

			// Validate outer structure
			if len(tryStmt.Excepts) != test.outerExcepts {
				t.Errorf("Expected %d outer except handlers but got %d for %s",
					test.outerExcepts, len(tryStmt.Excepts), test.description)
			}

			if len(tryStmt.Finally) != test.outerFinally {
				t.Errorf("Expected %d outer finally statements but got %d for %s",
					test.outerFinally, len(tryStmt.Finally), test.description)
			}

			// Basic validation that we have a valid nested structure
			if len(tryStmt.Body) == 0 {
				t.Errorf("Try statement missing body for %s", test.description)
			}
		})
	}
}

// Test try statement error cases and edge conditions
func TestTryErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorText   string
		description string
	}{
		{
			name:        "missing colon after try",
			input:       "try\n    pass",
			errorText:   "expected",
			description: "try keyword without colon",
		},
		{
			name:        "missing colon after except",
			input:       "try:\n    pass\nexcept ValueError\n    pass",
			errorText:   "expected",
			description: "except clause without colon",
		},
		{
			name:        "missing colon after else",
			input:       "try:\n    pass\nexcept:\n    pass\nelse\n    pass",
			errorText:   "expected",
			description: "else clause without colon",
		},
		{
			name:        "missing colon after finally",
			input:       "try:\n    pass\nexcept:\n    pass\nfinally\n    pass",
			errorText:   "expected",
			description: "finally clause without colon",
		},
		{
			name:        "empty try body",
			input:       "try:\nexcept:\n    pass",
			errorText:   "expected",
			description: "try with empty body",
		},
		{
			name:        "empty except body",
			input:       "try:\n    pass\nexcept:",
			errorText:   "expected",
			description: "except with empty body",
		},
		{
			name:        "invalid except syntax",
			input:       "try:\n    pass\nexcept as e:\n    pass",
			errorText:   "expected",
			description: "except with 'as' but no exception type",
		},
		// Note: (ValueError,) is actually valid Python syntax - single element tuple
		// Removed this test as it was incorrectly expecting an error
		{
			name:        "invalid variable name",
			input:       "try:\n    pass\nexcept ValueError as 123:\n    pass",
			errorText:   "expected",
			description: "except with invalid variable name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)
			validateTryParseError(t, tryStmt, err, test.errorText, test.description)
		})
	}
}

// Test edge cases and complex try scenarios
func TestTryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		// Complex valid cases
		{
			name: "try with complex exception expressions",
			input: `try:
    operation()
except getattr(module, 'CustomError') as e:
    handle_error(e)`,
			description: "try-except with complex exception expression",
		},
		{
			name: "try with lambda in body",
			input: `try:
    result = (lambda x: x * 2)(value)
except:
    result = None`,
			description: "try with lambda expression in body",
		},
		{
			name: "try with comprehension in body",
			input: `try:
    results = [process(item) for item in items]
except ProcessingError:
    results = []`,
			description: "try with list comprehension in body",
		},
		{
			name: "try with async operations",
			input: `try:
    await async_operation()
except AsyncError:
    await handle_async_error()`,
			description: "try with async/await operations",
		},
		{
			name: "try with star expressions",
			input: `try:
    result = func(*args, **kwargs)
except TypeError:
    result = None`,
			description: "try with star expressions",
		},

		// Edge error cases
		{
			name: "except with invalid expression",
			input: `try:
    pass
except 123:
    pass`,
			hasError:    true,
			errorText:   "expected",
			description: "except with invalid exception expression",
		},
		{
			name: "nested syntax error",
			input: `try:
    try:
        pass
    except
except:
    pass`,
			hasError:    true,
			errorText:   "expected",
			description: "syntax error in nested try",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tryStmt, err := parseTryStatement(t, test.input)

			if test.hasError {
				validateTryParseError(t, tryStmt, err, test.errorText, test.description)
			} else {
				validateTryParseSuccess(t, tryStmt, err, test.description)

				// For successful cases, verify basic structure
				if len(tryStmt.Body) == 0 {
					t.Errorf("Expected non-empty try body for %s", test.description)
				}
			}
		})
	}
}
