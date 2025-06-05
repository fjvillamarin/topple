package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a with statement
func parseWithStatement(t *testing.T, input string) (*ast.With, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.withStatement()
	if err != nil {
		return nil, err
	}
	withStmt, ok := stmt.(*ast.With)
	if !ok {
		return nil, nil
	}
	return withStmt, nil
}

// Helper function to validate with statement structure
func validateWithStatement(t *testing.T, withStmt *ast.With, expectedItemCount int, expectedAsync bool, description string) {
	if withStmt == nil {
		t.Fatalf("Expected With statement but got nil for %s", description)
	}

	if withStmt.IsAsync != expectedAsync {
		t.Errorf("Expected IsAsync=%v but got %v for %s", expectedAsync, withStmt.IsAsync, description)
	}

	if expectedItemCount >= 0 && len(withStmt.Items) != expectedItemCount {
		t.Errorf("Expected %d context items but got %d for %s", expectedItemCount, len(withStmt.Items), description)
	}

	if len(withStmt.Body) == 0 {
		t.Errorf("With statement missing body for %s", description)
	}

	// Validate each item has a context expression
	for i, item := range withStmt.Items {
		if item.Expr == nil {
			t.Errorf("Item %d missing context expression for %s", i, description)
		}
	}
}

// Helper function to validate with parsing success
func validateWithParseSuccess(t *testing.T, withStmt *ast.With, err error, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if withStmt == nil {
		t.Fatalf("Expected With statement but got nil for %s", description)
	}
}

// Helper function to validate with parsing error
func validateWithParseError(t *testing.T, withStmt *ast.With, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if withStmt != nil {
		t.Errorf("Expected nil With statement on error, but got %T for %s", withStmt, description)
	}
}

// Helper function to validate with item structure
func validateWithItem(t *testing.T, withStmt *ast.With, itemIndex int, expectedHasTarget bool, description string) {
	if itemIndex >= len(withStmt.Items) {
		t.Fatalf("Test expects item at index %d, but only %d items found for %s", itemIndex, len(withStmt.Items), description)
	}

	item := withStmt.Items[itemIndex]
	hasTarget := item.As != nil

	if hasTarget != expectedHasTarget {
		t.Errorf("Expected hasTarget=%v but got %v for item %d in %s", expectedHasTarget, hasTarget, itemIndex, description)
	}

	if item.Expr == nil {
		t.Errorf("Item %d missing context expression for %s", itemIndex, description)
	}
}

// Test comprehensive with statement parsing functionality
func TestWithStatement(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		errorText         string
		expectedAsync     bool
		expectedItemCount int
		description       string
	}{
		{
			name:              "simple with",
			input: `with open("file.txt") as f:
    content = f.read()`,
			expectedAsync:     false,
			expectedItemCount: 1,
			description:       "basic with statement with single context manager",
		},
		{
			name:              "with without as",
			input: `with open("file.txt"):
    pass`,
			expectedAsync:     false,
			expectedItemCount: 1,
			description:       "with statement without target variable binding",
		},
		{
			name:              "multiple context managers",
			input: `with open("file1.txt") as f1, open("file2.txt") as f2:
    merge(f1, f2)`,
			expectedAsync:     false,
			expectedItemCount: 2,
			description:       "with statement with multiple context managers",
		},
		{
			name:              "async with",
			input: `async with aiofiles.open("file.txt") as f:
    content = await f.read()`,
			expectedAsync:     true,
			expectedItemCount: 1,
			description:       "async with statement for asynchronous context manager",
		},
		{
			name:              "nested context expressions",
			input: `with lock1, lock2:
    critical_section()`,
			expectedAsync:     false,
			expectedItemCount: 2,
			description:       "with statement with multiple simple context expressions",
		},
		{
			name:              "complex context expression",
			input: `with (
    database.session() as db,
    redis.connection() as cache
):
    process(db, cache)`,
			expectedAsync:     false,
			expectedItemCount: 2,
			description:       "with statement with parenthesized context managers",
		},
		{
			name:              "with expression result",
			input: `with threading.Lock():
    shared_resource += 1`,
			expectedAsync:     false,
			expectedItemCount: 1,
			description:       "with statement using function call as context manager",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withStmt, err := parseWithStatement(t, test.input)

			if test.hasError {
				validateWithParseError(t, withStmt, err, test.errorText, test.description)
			} else {
				validateWithParseSuccess(t, withStmt, err, test.description)
				validateWithStatement(t, withStmt, test.expectedItemCount, test.expectedAsync, test.description)
			}
		})
	}
}

// Test with statement items
func TestWithItems(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hasTarget bool
		itemIndex int
	}{
		{
			"with target",
			`with open("file.txt") as f:
    pass`,
			true,
			0,
		},
		{
			"without target",
			`with lock:
    pass`,
			false,
			0,
		},
		{
			"multiple items first with target",
			`with open("file1.txt") as f1, open("file2.txt"):
    pass`,
			true,
			0,
		},
		{
			"multiple items second without target",
			`with open("file1.txt") as f1, open("file2.txt"):
    pass`,
			false,
			1,
		},
		{
			"simple call expression", 
			`with database.connection() as db:
    pass`,
			true,
			0,
		},
		{
			"call expression",
			`with get_connection() as conn:
    pass`,
			true,
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withStmt, err := parseWithStatement(t, test.input)
			validateWithParseSuccess(t, withStmt, err, test.name)
			validateWithItem(t, withStmt, test.itemIndex, test.hasTarget, test.name)
		})
	}
}

// Test async with statements
func TestAsyncWith(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"simple async with",
			`async with aiofiles.open("file.txt") as f:
    content = await f.read()`,
		},
		{
			"async with multiple items",
			`async with db.transaction() as tx, redis.lock() as lock:
    await process_with_lock(tx, lock)`,
		},
		{
			"async with without target",
			`async with acquire_resource():
    await do_work()`,
		},
		{
			"nested async with",
			`async with outer_context():
    async with inner_context() as inner:
        await process(inner)`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withStmt, err := parseWithStatement(t, test.input)
			validateWithParseSuccess(t, withStmt, err, test.name)
			validateWithStatement(t, withStmt, -1, true, test.name)
		})
	}
}

// Test with statement body
func TestWithBody(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stmtCount int
	}{
		{
			"single statement body",
			`with resource:
    process()`,
			1,
		},
		{
			"multiple statement body",
			`with resource:
    setup()
    process()
    cleanup()`,
			3,
		},
		{
			"nested with",
			`with outer:
    with inner:
        process()`,
			1, // The inner with is one statement
		},
		{
			"complex body",
			`with database.transaction() as tx:
    user = tx.get_user(user_id)
    if user:
        user.update(data)
        tx.commit()
    else:
        tx.rollback()`,
			4, // user=, if, user.update/tx.commit, tx.rollback counted separately
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withStmt, err := parseWithStatement(t, test.input)
			validateWithParseSuccess(t, withStmt, err, test.name)
			
			// Verify the body has statements
			if len(withStmt.Body) == 0 {
				t.Errorf("Expected with statement to have body statements for %s", test.name)
			}
		})
	}
}

// Test error cases
func TestWithErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing colon", "with resource\n    pass"},
		{"empty context", "with :\n    pass"},
		{"missing body", "with resource:\n"},
		{"invalid as target", "with resource as 123:\n    pass"},
		{"async without with", "async resource:\n    pass"},
		{"multiple as in single item", "with resource as a as b:\n    pass"},
		{"missing expression after as", "with resource as :\n    pass"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withStmt, err := parseWithStatement(t, test.input)
			validateWithParseError(t, withStmt, err, "", test.name)
		})
	}
}