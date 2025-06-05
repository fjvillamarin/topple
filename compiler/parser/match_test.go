package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a match statement
func parseMatchStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.statement()
}

// Helper function to validate match statement structure
func validateMatchStatement(t *testing.T, stmt ast.Stmt, expectedCases int, hasGuards bool) *ast.MatchStmt {
	matchStmt, ok := stmt.(*ast.MatchStmt)
	if !ok {
		t.Fatalf("Expected *ast.MatchStmt, got %T", stmt)
	}

	if matchStmt.Subject == nil {
		t.Fatal("Match statement should have a subject")
	}

	if len(matchStmt.Cases) == 0 {
		t.Fatal("Match statement should have at least one case")
	}

	if expectedCases >= 0 && len(matchStmt.Cases) != expectedCases {
		t.Errorf("Expected %d cases, got %d", expectedCases, len(matchStmt.Cases))
	}

	// Check for guards if specified
	if hasGuards {
		hasGuardedCase := false
		for _, caseBlock := range matchStmt.Cases {
			if caseBlock.Guard != nil {
				hasGuardedCase = true
				break
			}
		}
		if !hasGuardedCase {
			t.Error("Expected at least one case with a guard")
		}
	}

	return matchStmt
}

// Helper function to count different pattern types in match statement
func countPatternTypes(matchStmt *ast.MatchStmt) (literals, captures, wildcards, sequences, mappings int) {
	for _, caseBlock := range matchStmt.Cases {
		for _, pattern := range caseBlock.Patterns {
			switch pattern.(type) {
			case *ast.LiteralPattern:
				literals++
			case *ast.CapturePattern:
				captures++
			case *ast.WildcardPattern:
				wildcards++
			case *ast.SequencePattern:
				sequences++
			case *ast.MappingPattern:
				mappings++
			}
		}
	}
	return
}

// Helper function to check if match has specific pattern features
func hasPatternFeature(matchStmt *ast.MatchStmt, feature string) bool {
	for _, caseBlock := range matchStmt.Cases {
		switch feature {
		case "guard":
			if caseBlock.Guard != nil {
				return true
			}
		case "wildcard":
			for _, pattern := range caseBlock.Patterns {
				if _, ok := pattern.(*ast.WildcardPattern); ok {
					return true
				}
			}
		case "sequence":
			for _, pattern := range caseBlock.Patterns {
				if _, ok := pattern.(*ast.SequencePattern); ok {
					return true
				}
			}
		case "mapping":
			for _, pattern := range caseBlock.Patterns {
				if _, ok := pattern.(*ast.MappingPattern); ok {
					return true
				}
			}
		}
	}
	return false
}

// Test basic match statement functionality
func TestMatchStatements(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCases int
		hasGuards     bool
		description   string
	}{
		// Basic match statements
		{
			name: "simple literal match",
			input: `match value:
    case 1:
        print("one")
    case 2:
        print("two")`,
			expectedCases: 2,
			description:   "basic match with literal patterns",
		},
		{
			name: "match with wildcard",
			input: `match x:
    case 1:
        return "one"
    case _:
        return "other"`,
			expectedCases: 2,
			description:   "match with wildcard pattern",
		},
		{
			name: "match with capture pattern",
			input: `match data:
    case name:
        print(f"Hello {name}")`,
			expectedCases: 1,
			description:   "match with variable capture",
		},
		{
			name: "match with multiple cases",
			input: `match status:
    case "success":
        handle_success()
    case "error":
        handle_error()
    case "pending":
        handle_pending()`,
			expectedCases: 3,
			description:   "match with multiple string literals",
		},

		// Error cases
		{
			name:        "match without cases",
			input:       `match value:`,
			hasError:    true,
			description: "match statement missing case blocks",
		},
		{
			name:        "match without colon",
			input:       `match value case 1: pass`,
			hasError:    true,
			description: "match statement missing colon",
		},
		{
			name:        "match without subject",
			input:       `match: case 1: pass`,
			hasError:    true,
			description: "match statement missing subject expression",
		},
		{
			name:        "incomplete match",
			input:       `match`,
			hasError:    true,
			description: "incomplete match statement",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateMatchStatement(t, stmt, test.expectedCases, test.hasGuards)
		})
	}
}

// Test match statements with guards
func TestMatchWithGuards(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCases int
		description   string
	}{
		{
			name: "simple guard",
			input: `match x:
    case n if n > 0:
        print("positive")
    case n if n < 0:
        print("negative")
    case 0:
        print("zero")`,
			expectedCases: 3,
			description:   "match with conditional guards",
		},
		{
			name: "complex guard expression",
			input: `match point:
    case (x, y) if x == y:
        print("on diagonal")
    case (x, y) if x > y:
        print("above diagonal")
    case (x, y):
        print("below diagonal")`,
			expectedCases: 3,
			description:   "match with complex guard expressions",
		},
		{
			name: "guard with function call",
			input: `match data:
    case value if is_valid(value):
        process(value)
    case _:
        reject()`,
			expectedCases: 2,
			description:   "match with function call in guard",
		},
		{
			name: "multiple guards on same pattern",
			input: `match item:
    case x if x > 100:
        handle_large(x)
    case x if x > 10:
        handle_medium(x)
    case x:
        handle_small(x)`,
			expectedCases: 3,
			description:   "match with multiple guards on similar patterns",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateMatchStatement(t, stmt, test.expectedCases, true)
		})
	}
}

// Test sequence patterns in match statements
func TestMatchSequencePatterns(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCases int
		description   string
	}{
		{
			name: "tuple pattern",
			input: `match point:
    case (x, y):
        print(f"Point at {x}, {y}")`,
			expectedCases: 1,
			description:   "match with tuple destructuring",
		},
		{
			name: "list pattern",
			input: `match items:
    case []:
        print("empty")
    case [x]:
        print(f"single item: {x}")
    case [x, y]:
        print(f"two items: {x}, {y}")`,
			expectedCases: 3,
			description:   "match with list patterns of different lengths",
		},
		{
			name: "star pattern",
			input: `match items:
    case [first, *rest]:
        print(f"first: {first}, rest: {rest}")
    case [*head, last]:
        print(f"head: {head}, last: {last}")`,
			expectedCases: 2,
			description:   "match with starred patterns for variable length",
		},
		{
			name: "nested sequence patterns",
			input: `match data:
    case [(x, y), (a, b)]:
        print("two points")
    case [first, *middle, last]:
        print("at least two items")`,
			expectedCases: 2,
			description:   "match with nested sequence patterns",
		},
		{
			name: "complex sequence pattern",
			input: `match matrix:
    case [[a, b], [c, d]]:
        print("2x2 matrix")
    case [row] if len(row) == 3:
        print("single row with 3 elements")
    case []:
        print("empty matrix")`,
			expectedCases: 3,
			description:   "match with complex nested sequences and guards",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateMatchStatement(t, stmt, test.expectedCases, false)
		})
	}
}

// Test mapping patterns in match statements
func TestMatchMappingPatterns(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCases int
		description   string
	}{
		{
			name: "simple mapping pattern",
			input: `match data:
    case {"name": name}:
        print(f"Name: {name}")
    case {"age": age}:
        print(f"Age: {age}")`,
			expectedCases: 2,
			description:   "match with dictionary key extraction",
		},
		{
			name: "mapping with multiple keys",
			input: `match person:
    case {"name": name, "age": age}:
        print(f"{name} is {age} years old")`,
			expectedCases: 1,
			description:   "match with multiple dictionary keys",
		},
		{
			name: "mapping with rest pattern",
			input: `match config:
    case {"type": "user", **rest}:
        handle_user_config(rest)
    case {"type": "admin", **rest}:
        handle_admin_config(rest)`,
			expectedCases: 2,
			description:   "match with dictionary rest pattern",
		},
		{
			name: "empty mapping",
			input: `match data:
    case {}:
        print("empty dict")
    case {"key": value}:
        print(f"has key: {value}")`,
			expectedCases: 2,
			description:   "match with empty dictionary pattern",
		},
		{
			name: "nested mapping pattern",
			input: `match response:
    case {"data": {"user": {"name": name}}}:
        print(f"User name: {name}")
    case {"error": {"message": msg}}:
        print(f"Error: {msg}")`,
			expectedCases: 2,
			description:   "match with nested dictionary patterns",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateMatchStatement(t, stmt, test.expectedCases, false)
		})
	}
}

// Test advanced match patterns
func TestMatchAdvancedPatterns(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCases int
		description   string
	}{
		{
			name: "or patterns",
			input: `match value:
    case 1 | 2 | 3:
        print("small number")
    case 10 | 20 | 30:
        print("large number")`,
			expectedCases: 2,
			description:   "match with OR patterns",
		},
		{
			name: "as patterns",
			input: `match data:
    case [1, 2, 3] as numbers:
        print(f"got numbers: {numbers}")
    case (x, y) as coords:
        print(f"coordinates: {coords}")`,
			expectedCases: 2,
			description:   "match with AS patterns for capture",
		},
		{
			name: "value patterns",
			input: `match status:
    case Status.SUCCESS:
        handle_success()
    case Status.ERROR:
        handle_error()`,
			expectedCases: 2,
			description:   "match with constant value patterns",
		},
		{
			name: "class patterns",
			input: `match obj:
    case Point(x=0, y=0):
        print("origin")
    case Point(x=x, y=y):
        print(f"point at {x}, {y}")`,
			expectedCases: 2,
			description:   "match with class/constructor patterns",
		},
		{
			name: "mixed complex patterns",
			input: `match data:
    case {"type": "user", "data": [name, age]} if age >= 18:
        print(f"Adult: {name}")
    case {"type": "user", "data": user_info} as full_record:
        print(f"User record: {full_record}")
    case [*items] if len(items) > 5:
        print("many items")
    case _:
        print("unknown format")`,
			expectedCases: 4,
			description:   "match with mixed complex pattern types",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateMatchStatement(t, stmt, test.expectedCases, false)
		})
	}
}

// Test match edge cases and error scenarios
func TestMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		// Complex valid cases
		{
			name: "deeply nested patterns",
			input: `match data:
    case {"users": [{"profile": {"name": name, "settings": {"theme": theme}}}]}:
        print(f"User {name} prefers {theme}")
    case {"config": {"database": {"host": host, "port": port}}}:
        print(f"DB at {host}:{port}")`,
			description: "match with deeply nested structural patterns",
		},
		{
			name: "complex subject expression",
			input: `match func(get_data().transform()):
    case result if validate(result):
        use(result)
    case _:
        fallback()`,
			description: "match with complex subject expression",
		},
		{
			name: "match with function call in guard",
			input: `match items:
    case data if validate_all(data):
        print("all valid")
    case data if has_invalid(data):
        print("has invalid")`,
			description: "match with function calls in guard",
		},

		// Error cases
		{
			name:          "invalid pattern syntax",
			input:         `match x: case 1 2: pass`,
			hasError:      true,
			errorContains: "case",
			description:   "invalid case pattern syntax",
		},
		{
			name:        "missing case keyword",
			input:       `match x: 1: pass`,
			hasError:    true,
			description: "case block missing 'case' keyword",
		},
		{
			name:        "invalid guard syntax",
			input:       `match x: case n if: pass`,
			hasError:    true,
			description: "guard with incomplete condition",
		},
		{
			name:        "malformed mapping pattern",
			input:       `match x: case {key value}: pass`,
			hasError:    true,
			description: "malformed dictionary pattern",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseMatchStatement(t, test.input)

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

			// For successful cases, just verify we got a valid match statement
			if stmt == nil {
				t.Errorf("Expected valid statement but got nil for %s", test.description)
				return
			}

			validateMatchStatement(t, stmt, -1, false)
		})
	}
}
