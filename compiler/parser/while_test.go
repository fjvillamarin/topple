package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a while statement
func parseWhileStatement(t *testing.T, input string) (*ast.While, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.whileStatement()
	if err != nil {
		return nil, err
	}
	whileStmt, ok := stmt.(*ast.While)
	if !ok {
		return nil, nil
	}
	return whileStmt, nil
}

// Helper function to validate while statement structure
func validateWhileStatement(t *testing.T, whileStmt *ast.While, expectedElseCount int, description string) {
	if whileStmt == nil {
		t.Fatalf("Expected While statement but got nil for %s", description)
	}

	if whileStmt.Test == nil {
		t.Errorf("While statement missing test condition for %s", description)
	}

	if len(whileStmt.Body) == 0 {
		t.Errorf("While statement missing body for %s", description)
	}

	if expectedElseCount >= 0 && len(whileStmt.Else) != expectedElseCount {
		t.Errorf("Expected %d else statements but got %d for %s",
			expectedElseCount, len(whileStmt.Else), description)
	}
}

// Helper function to validate while parsing success
func validateWhileParseSuccess(t *testing.T, whileStmt *ast.While, err error, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if whileStmt == nil {
		t.Fatalf("Expected While statement but got nil for %s", description)
	}
}

// Helper function to validate while parsing error
func validateWhileParseError(t *testing.T, whileStmt *ast.While, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if whileStmt != nil {
		t.Errorf("Expected nil While statement on error, but got %T for %s", whileStmt, description)
	}
}

// Helper function to recursively check for break/continue statements in statement list
func findBreakContinueInStatements(stmts []ast.Stmt, hasBreak *bool, hasContinue *bool) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.BreakStmt:
			*hasBreak = true
		case *ast.ContinueStmt:
			*hasContinue = true
		case *ast.If:
			findBreakContinueInStatements(s.Body, hasBreak, hasContinue)
			findBreakContinueInStatements(s.Else, hasBreak, hasContinue)
		case *ast.While:
			findBreakContinueInStatements(s.Body, hasBreak, hasContinue)
		case *ast.For:
			findBreakContinueInStatements(s.Body, hasBreak, hasContinue)
		case *ast.MultiStmt:
			findBreakContinueInStatements(s.Stmts, hasBreak, hasContinue)
		}
	}
}

// Helper function to validate break/continue presence
func validateBreakContinue(t *testing.T, whileStmt *ast.While, expectedBreak, expectedContinue bool, description string) {
	hasBreak := false
	hasContinue := false
	findBreakContinueInStatements(whileStmt.Body, &hasBreak, &hasContinue)

	if hasBreak != expectedBreak {
		t.Errorf("Expected hasBreak=%v but got %v for %s", expectedBreak, hasBreak, description)
	}
	if hasContinue != expectedContinue {
		t.Errorf("Expected hasContinue=%v but got %v for %s", expectedContinue, hasContinue, description)
	}
}

// Test comprehensive while statement parsing functionality
func TestWhileStatement(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		errorText         string
		expectedElseCount int
		description       string
	}{
		{
			name: "simple while",
			input: `while x > 0:
    x -= 1`,
			expectedElseCount: 0,
			description:       "basic while loop with simple condition",
		},
		{
			name: "while with else",
			input: `while x > 0:
    x -= 1
else:
    print("done")`,
			expectedElseCount: 1,
			description:       "while loop with else clause",
		},
		{
			name: "while True",
			input: `while True:
    if done():
        break
    process()`,
			expectedElseCount: 0,
			description:       "infinite while loop with break condition",
		},
		{
			name: "while with complex condition",
			input: `while x > 0 and not finished:
    process(x)
    x = get_next()`,
			expectedElseCount: 0,
			description:       "while loop with complex boolean condition",
		},
		{
			name: "nested while",
			input: `while outer:
    while inner:
        do_work()
    update_outer()`,
			expectedElseCount: 0,
			description:       "nested while loops",
		},
		{
			name: "while with continue",
			input: `while items:
    item = items.pop()
    if skip(item):
        continue
    process(item)`,
			expectedElseCount: 0,
			description:       "while loop with continue statement",
		},
		{
			name: "while with walrus",
			input: `while (line := file.readline()):
    process(line)`,
			expectedElseCount: 0,
			description:       "while loop with walrus operator assignment",
		},
		{
			name: "empty while",
			input: `while waiting():
    pass`,
			expectedElseCount: 0,
			description:       "while loop with pass statement",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whileStmt, err := parseWhileStatement(t, test.input)

			if test.hasError {
				validateWhileParseError(t, whileStmt, err, test.errorText, test.description)
			} else {
				validateWhileParseSuccess(t, whileStmt, err, test.description)
				validateWhileStatement(t, whileStmt, test.expectedElseCount, test.description)
			}
		})
	}
}

// Test while loops with break/continue statements
func TestWhileBreakContinue(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		hasError         bool
		errorText        string
		expectedBreak    bool
		expectedContinue bool
		description      string
	}{
		{
			name: "with break",
			input: `while True:
    if condition:
        break`,
			expectedBreak:    true,
			expectedContinue: false,
			description:      "while loop containing break statement",
		},
		{
			name: "with continue",
			input: `while x < 10:
    if x % 2 == 0:
        continue
    process(x)`,
			expectedBreak:    false,
			expectedContinue: true,
			description:      "while loop containing continue statement",
		},
		{
			name: "with both",
			input: `while running:
    x = get_next()
    if x < 0:
        break
    if x == 0:
        continue
    process(x)`,
			expectedBreak:    true,
			expectedContinue: true,
			description:      "while loop containing both break and continue",
		},
		{
			name: "nested break",
			input: `while outer:
    while inner:
        if done:
            break
    if finished:
        break`,
			expectedBreak:    true,
			expectedContinue: false,
			description:      "nested while loops with break statements",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whileStmt, err := parseWhileStatement(t, test.input)

			if test.hasError {
				validateWhileParseError(t, whileStmt, err, test.errorText, test.description)
			} else {
				validateWhileParseSuccess(t, whileStmt, err, test.description)
				validateBreakContinue(t, whileStmt, test.expectedBreak, test.expectedContinue, test.description)
			}
		})
	}
}

// Test while statement error cases and edge conditions
func TestWhileErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorText   string
		description string
	}{
		{
			name:        "missing colon",
			input:       "while x > 0\n    x -= 1",
			errorText:   "expected",
			description: "while statement missing colon should fail",
		},
		{
			name:        "missing condition",
			input:       "while :\n    pass",
			errorText:   "expected",
			description: "while statement missing condition should fail",
		},
		{
			name:        "empty body",
			input:       "while True:\n",
			errorText:   "expected",
			description: "while statement with empty body should fail",
		},
		{
			name:        "invalid condition",
			input:       "while x = 1:\n    pass",
			errorText:   "expected",
			description: "while statement with assignment in condition should fail",
		},
		{
			name:        "else without body",
			input:       "while x:\n    pass\nelse:\n",
			errorText:   "expected",
			description: "while else clause without body should fail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whileStmt, err := parseWhileStatement(t, test.input)
			validateWhileParseError(t, whileStmt, err, test.errorText, test.description)
		})
	}
}

// Test complex while statement scenarios
func TestWhileComplexScenarios(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		errorText         string
		expectedElseCount int
		description       string
	}{
		{
			name: "while with complex condition expressions",
			input: `while (x > 0 and y < 100) or (flag and not done):
    process(x, y)
    x, y = update_values(x, y)`,
			expectedElseCount: 0,
			description:       "while loop with complex boolean condition",
		},
		{
			name: "while with function call condition",
			input: `while has_more_data() and not should_stop():
    data = get_next_data()
    process_data(data)`,
			expectedElseCount: 0,
			description:       "while loop with function calls in condition",
		},
		{
			name: "while with multiple else statements",
			input: `while queue:
    item = queue.pop()
    if process(item):
        continue
    handle_error(item)
else:
    finalize_processing()
    cleanup_resources()`,
			expectedElseCount: 2,
			description:       "while loop with multiple statements in else clause",
		},
		{
			name: "while with nested control structures",
			input: `while running:
    try:
        item = get_item()
        for i in range(item.count):
            if should_process(i):
                process_item(item, i)
            else:
                skip_item(item, i)
    except StopIteration:
        break
    except ProcessingError:
        continue`,
			expectedElseCount: 0,
			description:       "while loop with nested try-except and for loop",
		},
		{
			name: "while with comprehensions",
			input: `while items:
    processed = [transform(x) for x in items if is_valid(x)]
    results.extend(processed)
    items = get_next_batch()`,
			expectedElseCount: 0,
			description:       "while loop containing list comprehensions",
		},
		{
			name: "while with lambda expressions",
			input: `while data:
    filtered = filter(lambda x: x.is_active, data)
    processed = map(lambda x: x.process(), filtered)
    results.extend(list(processed))
    data = get_more_data()`,
			expectedElseCount: 0,
			description:       "while loop with lambda expressions",
		},
		{
			name: "deeply nested while loops",
			input: `while outer_condition:
    while middle_condition:
        while inner_condition:
            if deepest_check():
                process_deeply_nested()
            inner_condition = update_inner()
        middle_condition = update_middle()
    outer_condition = update_outer()`,
			expectedElseCount: 0,
			description:       "three levels of nested while loops",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whileStmt, err := parseWhileStatement(t, test.input)

			if test.hasError {
				validateWhileParseError(t, whileStmt, err, test.errorText, test.description)
			} else {
				validateWhileParseSuccess(t, whileStmt, err, test.description)
				validateWhileStatement(t, whileStmt, test.expectedElseCount, test.description)
			}
		})
	}
}

// Test while statement edge cases and boundary conditions
func TestWhileEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		{
			name: "while with very long condition",
			input: `while (condition1 and condition2 and condition3 and condition4 and condition5 and condition6 and condition7 and condition8):
    process_long_condition()`,
			description: "while loop with very long compound condition",
		},
		{
			name: "while with multiline condition",
			input: `while (check_first_condition() and
       check_second_condition() and
       check_third_condition()):
    process_multiline_condition()`,
			description: "while loop with multiline condition expression",
		},
		{
			name: "while with docstring",
			input: `while True:
    """
    This is a while loop with a docstring.
    It demonstrates docstring support in loops.
    """
    do_work()
    if should_exit():
        break`,
			description: "while loop containing docstring",
		},
		{
			name: "while with assignment and indexing",
			input: `while data:
    item = data[0]
    result = process(item)
    data = data[1:]`,
			description: "while loop with indexing and slicing",
		},
		{
			name: "while with invalid syntax in body",
			input: `while True:
    def invalid syntax here`,
			hasError:    true,
			errorText:   "expected",
			description: "while loop with invalid syntax in body should fail",
		},
		{
			name: "while with incomplete condition",
			input: `while x >:
    pass`,
			hasError:    true,
			errorText:   "expected",
			description: "while loop with incomplete condition should fail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			whileStmt, err := parseWhileStatement(t, test.input)

			if test.hasError {
				validateWhileParseError(t, whileStmt, err, test.errorText, test.description)
			} else {
				validateWhileParseSuccess(t, whileStmt, err, test.description)

				// For successful complex cases, just verify basic structure
				if len(whileStmt.Body) == 0 {
					t.Errorf("Expected while body to have statements for %s", test.description)
				}
			}
		})
	}
}
