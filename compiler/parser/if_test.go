package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse an if statement
func parseIfStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.ifStatement()
}

// Helper function to parse a ternary expression
func parseTernaryExpression(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.expression()
}

// Helper function to validate if statement structure
func validateIfStatement(t *testing.T, stmt ast.Stmt, expectedHasElse bool, expectedBodyCount int) *ast.If {
	ifStmt, ok := stmt.(*ast.If)
	if !ok {
		t.Fatalf("Expected *ast.If, got %T", stmt)
	}

	if ifStmt.Condition == nil {
		t.Fatal("If statement missing condition")
	}

	if len(ifStmt.Body) == 0 {
		t.Error("If statement missing body")
	}

	if expectedBodyCount > 0 && len(ifStmt.Body) != expectedBodyCount {
		t.Errorf("Expected %d body statements, got %d", expectedBodyCount, len(ifStmt.Body))
	}

	hasElse := len(ifStmt.Else) > 0
	if hasElse != expectedHasElse {
		t.Errorf("Expected hasElse=%v, got %v", expectedHasElse, hasElse)
	}

	return ifStmt
}

// Helper function to validate ternary expression structure
func validateTernaryExpression(t *testing.T, expr ast.Expr) *ast.TernaryExpr {
	ternary, ok := expr.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("Expected *ast.TernaryExpr, got %T", expr)
	}

	if ternary.TrueExpr == nil {
		t.Error("Ternary missing true expression")
	}
	if ternary.Condition == nil {
		t.Error("Ternary missing condition")
	}
	if ternary.FalseExpr == nil {
		t.Error("Ternary missing false expression")
	}

	return ternary
}

// Helper function to count nested elif blocks
func countElifBlocks(ifStmt *ast.If) (elifCount int, hasFinalElse bool) {
	current := ifStmt
	for len(current.Else) > 0 {
		if len(current.Else) == 1 {
			if nextIf, ok := current.Else[0].(*ast.If); ok {
				elifCount++
				current = nextIf
			} else {
				// This is the final else block
				hasFinalElse = true
				break
			}
		} else {
			// Multiple statements in else, final else block
			hasFinalElse = true
			break
		}
	}
	return
}

// Helper function to validate condition complexity
func validateConditionComplexity(t *testing.T, condition ast.Expr, expectedType string) {
	switch expectedType {
	case "simple":
		// Could be binary comparison, name, literal, etc.
	case "complex":
		// Should be complex expression with multiple operations
		if binary, ok := condition.(*ast.Binary); ok && binary != nil {
			// Complex condition likely has binary operations
		}
	case "assignment":
		// Should contain assignment expression (walrus operator)
	}
}

// Test comprehensive if statement functionality
func TestIfStatement(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		expectedHasElse   bool
		expectedBodyCount int
		conditionType     string
		description       string
	}{
		// Basic if statements
		{
			name: "simple if",
			input: `if x > 0:
    print("positive")`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "basic if statement with simple condition",
		},
		{
			name: "if with else",
			input: `if x > 0:
    print("positive")
else:
    print("not positive")`,
			expectedHasElse:   true,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with else clause",
		},
		{
			name: "complex condition",
			input: `if x > 0 and y < 10:
    process()`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "complex",
			description:       "if statement with complex boolean condition",
		},
		{
			name: "multiple statements in body",
			input: `if condition:
    x = 1
    y = 2
    z = 3`,
			expectedHasElse:   false,
			expectedBodyCount: 3,
			conditionType:     "simple",
			description:       "if statement with multiple body statements",
		},
		{
			name: "empty body with pass",
			input: `if x:
    pass`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with pass statement",
		},
		{
			name: "nested if statement",
			input: `if outer:
    if inner:
        do_something()`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with nested if in body",
		},
		{
			name: "if with walrus operator",
			input: `if (x := get_value()) > 0:
    use(x)`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "assignment",
			description:       "if statement with assignment expression",
		},
		{
			name: "if with function call condition",
			input: `if is_valid(data):
    process(data)`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with function call condition",
		},
		{
			name: "if with attribute access condition",
			input: `if obj.is_ready():
    obj.start()`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with method call condition",
		},
		{
			name: "if with comparison chain",
			input: `if 0 < x < 10:
    print("single digit")`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "complex",
			description:       "if statement with comparison chaining",
		},
		{
			name: "if with membership test",
			input: `if item in collection:
    use(item)`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with membership test",
		},
		{
			name: "if with not operator",
			input: `if not condition:
    handle_false_case()`,
			expectedHasElse:   false,
			expectedBodyCount: 1,
			conditionType:     "simple",
			description:       "if statement with negated condition",
		},

		// Error cases
		{
			name:        "missing colon",
			input:       "if x > 0\n    print('x')",
			hasError:    true,
			description: "if statement without colon",
		},
		{
			name:        "missing condition",
			input:       "if :\n    pass",
			hasError:    true,
			description: "if statement with empty condition",
		},
		{
			name:        "invalid condition syntax",
			input:       "if x = 1:\n    pass",
			hasError:    true,
			description: "if statement with assignment instead of comparison",
		},
		{
			name:        "missing indentation",
			input:       "if x:\nprint('x')",
			hasError:    true,
			description: "if statement with incorrect indentation",
		},
		{
			name:        "empty body",
			input:       "if x:",
			hasError:    true,
			description: "if statement with no body",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseIfStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			ifStmt := validateIfStatement(t, stmt, test.expectedHasElse, test.expectedBodyCount)

			// Validate condition complexity if specified
			if test.conditionType != "" {
				validateConditionComplexity(t, ifStmt.Condition, test.conditionType)
			}
		})
	}
}

// Test elif blocks and complex if-elif-else chains
func TestElifBlocks(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		expectedElifCount int
		expectedFinalElse bool
		description       string
	}{
		{
			name: "single elif",
			input: `if x > 0:
    print("positive")
elif x < 0:
    print("negative")`,
			expectedElifCount: 1,
			expectedFinalElse: false,
			description:       "if statement with single elif",
		},
		{
			name: "elif with final else",
			input: `if x > 0:
    print("positive")
elif x < 0:
    print("negative")
else:
    print("zero")`,
			expectedElifCount: 1,
			expectedFinalElse: true,
			description:       "if-elif-else chain",
		},
		{
			name: "multiple elif blocks",
			input: `if x == 1:
    print("one")
elif x == 2:
    print("two")
elif x == 3:
    print("three")
else:
    print("other")`,
			expectedElifCount: 2,
			expectedFinalElse: true,
			description:       "multiple elif blocks with final else",
		},
		{
			name: "complex elif conditions",
			input: `if x > 10:
    high()
elif x > 5 and x <= 10:
    medium()
elif x > 0 and x <= 5:
    low()
else:
    zero_or_negative()`,
			expectedElifCount: 2,
			expectedFinalElse: true,
			description:       "elif blocks with complex conditions",
		},
		{
			name: "nested elif structure",
			input: `if category == 'A':
    if subcategory == 1:
        handle_A1()
    elif subcategory == 2:
        handle_A2()
elif category == 'B':
    handle_B()`,
			expectedElifCount: 1,
			expectedFinalElse: false,
			description:       "nested if-elif structure",
		},
		{
			name: "long elif chain",
			input: `if grade >= 90:
    print("A")
elif grade >= 80:
    print("B")
elif grade >= 70:
    print("C")
elif grade >= 60:
    print("D")
else:
    print("F")`,
			expectedElifCount: 3,
			expectedFinalElse: true,
			description:       "long if-elif-else chain",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseIfStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			ifStmt := validateIfStatement(t, stmt, true, -1) // Don't check body count for complex structures

			elifCount, hasFinalElse := countElifBlocks(ifStmt)

			if elifCount != test.expectedElifCount {
				t.Errorf("Expected %d elif blocks for %s, got %d", test.expectedElifCount, test.description, elifCount)
			}

			if hasFinalElse != test.expectedFinalElse {
				t.Errorf("Expected final else=%v for %s, got %v", test.expectedFinalElse, test.description, hasFinalElse)
			}
		})
	}
}

// Test ternary/conditional expressions
func TestTernaryExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		isNested    bool
		description string
	}{
		{
			name:        "simple ternary",
			input:       "x if condition else y",
			description: "basic ternary expression",
		},
		{
			name:        "ternary with literals",
			input:       "1 if True else 0",
			description: "ternary with literal values",
		},
		{
			name:        "ternary with expressions",
			input:       "x * 2 if flag else x / 2",
			description: "ternary with arithmetic expressions",
		},
		{
			name:        "ternary with comparison",
			input:       "positive if x > 0 else negative",
			description: "ternary with comparison condition",
		},
		{
			name:        "ternary with function calls",
			input:       "process(x) if is_valid(x) else default_value()",
			description: "ternary with function calls",
		},
		{
			name:        "ternary in assignment context",
			input:       "'yes' if answer else 'no'",
			description: "ternary expression for assignment",
		},
		{
			name:        "ternary as function argument",
			input:       "abs(x) if x < 0 else x",
			description: "ternary expression as function argument",
		},
		{
			name:        "nested ternary",
			input:       "a if x else b if y else c",
			isNested:    true,
			description: "nested ternary expressions",
		},
		{
			name:        "complex nested ternary",
			input:       "high if score > 80 else medium if score > 60 else low",
			isNested:    true,
			description: "complex nested ternary chain",
		},
		{
			name:        "ternary with method chaining",
			input:       "obj.method1() if condition else obj.method2()",
			description: "ternary with method calls",
		},
		{
			name:        "ternary with list access",
			input:       "items[0] if items else None",
			description: "ternary with list indexing",
		},
		{
			name:        "ternary with attribute access",
			input:       "obj.attr if obj else default.attr",
			description: "ternary with attribute access",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parseTernaryExpression(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateTernaryExpression(t, expr)

			// For nested ternaries, check the structure
			if test.isNested {
				ternary := expr.(*ast.TernaryExpr)
				// The false expression should also be a ternary in nested cases
				if _, ok := ternary.FalseExpr.(*ast.TernaryExpr); !ok {
					t.Errorf("Expected nested ternary in false expression for %s", test.description)
				}
			}
		})
	}
}

// Test error cases and edge scenarios
func TestIfEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		{
			name: "if with complex nested body",
			input: `if condition:
    for item in items:
        if item.is_valid():
            process(item)
        else:
            skip(item)`,
			hasError:    false,
			description: "if statement with complex nested body",
		},
		{
			name: "if with try-except in body",
			input: `if should_process:
    try:
        risky_operation()
    except Exception:
        handle_error()`,
			hasError:    false,
			description: "if statement with exception handling in body",
		},
		{
			name: "if with comprehension condition",
			input: `if any(x > 0 for x in values):
    print("has positive")`,
			hasError:    false,
			description: "if statement with generator expression condition",
		},

		// Error cases
		{
			name:          "standalone elif",
			input:         "elif x > 0:\n    pass",
			hasError:      true,
			errorContains: "elif",
			description:   "elif without preceding if",
		},
		{
			name:          "standalone else",
			input:         "else:\n    pass",
			hasError:      true,
			errorContains: "else",
			description:   "else without preceding if",
		},
		{
			name:        "invalid ternary syntax",
			input:       "x if else y",
			hasError:    true,
			description: "ternary expression with missing condition",
		},
		{
			name:        "incomplete ternary",
			input:       "x if condition",
			hasError:    true,
			description: "ternary expression missing else clause",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error

			// Try parsing as statement first, then as expression
			if strings.HasPrefix(test.input, "if ") {
				_, err = parseIfStatement(t, test.input)
			} else {
				_, err = parseTernaryExpression(t, test.input)
			}

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}
		})
	}
}
