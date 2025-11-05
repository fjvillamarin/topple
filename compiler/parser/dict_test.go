package parser

import (
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse a dictionary expression
func parseDict(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.dict()
}

// Helper function to parse dict or set (for comprehensions)
func parseDictOrSet(t *testing.T, input string) (ast.Expr, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.dictOrSet()
}

// Helper function to count different types of dict pairs
func countDictPairs(pairs []ast.DictPair) (kvCount, unpackCount int) {
	for _, pair := range pairs {
		switch pair.(type) {
		case *ast.KeyValuePair:
			kvCount++
		case *ast.DoubleStarredPair:
			unpackCount++
		}
	}
	return
}

// Helper function to validate dict expression structure
func validateDictExpr(t *testing.T, expr ast.Expr, expectedKV, expectedUnpack int) {
	dict, ok := expr.(*ast.DictExpr)
	if !ok {
		t.Fatalf("Expected *ast.DictExpr, got %T", expr)
	}

	kvCount, unpackCount := countDictPairs(dict.Pairs)

	if kvCount != expectedKV {
		t.Errorf("Expected %d key-value pairs, got %d", expectedKV, kvCount)
	}
	if expectedUnpack >= 0 && unpackCount != expectedUnpack {
		t.Errorf("Expected %d unpack expressions, got %d", expectedUnpack, unpackCount)
	}
}

// Helper function to validate dict comprehension
func validateDictComp(t *testing.T, expr ast.Expr, expectedGenerators int) {
	comp, ok := expr.(*ast.DictComp)
	if !ok {
		t.Fatalf("Expected *ast.DictComp, got %T", expr)
	}

	if comp.Key == nil {
		t.Error("Dict comprehension missing key expression")
	}
	if comp.Value == nil {
		t.Error("Dict comprehension missing value expression")
	}
	if len(comp.Clauses) == 0 {
		t.Error("Dict comprehension missing for clauses")
	}

	genCount := len(comp.Clauses)
	if genCount != expectedGenerators {
		t.Errorf("Expected %d generators, got %d", expectedGenerators, genCount)
	}
}

// Test comprehensive dictionary functionality
func TestDictionary(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		hasError         bool
		expectedKV       int
		expectedUnpack   int
		isDictComp       bool
		expectedGenCount int
	}{
		// Basic dictionary expressions
		{"empty dict", "{}", false, 0, 0, false, 0},
		{"single pair", "{'key': 'value'}", false, 1, 0, false, 0},
		{"multiple pairs", "{'a': 1, 'b': 2, 'c': 3}", false, 3, 0, false, 0},
		{"numeric keys", "{1: 'one', 2: 'two'}", false, 2, 0, false, 0},
		{"expression keys", "{x + 1: y * 2}", false, 1, 0, false, 0},
		{"nested dict", "{'outer': {'inner': 'value'}}", false, 1, 0, false, 0},
		{"trailing comma", "{'a': 1, 'b': 2,}", false, 2, 0, false, 0},
		{"mixed key types", "{1: 'int', 'str': 2, (1, 2): 'tuple'}", false, 3, 0, false, 0},

		// Dictionary unpacking
		{"single unpack", "{**other}", false, 0, 1, false, 0},
		{"unpack with pairs", "{'a': 1, **other, 'b': 2}", false, 2, 1, false, 0},
		{"multiple unpacks", "{**dict1, **dict2}", false, 0, 2, false, 0},
		{"mixed unpacking", "{'x': 1, **a, 'y': 2, **b}", false, 2, 2, false, 0},
		{"unpack expression", "{**get_dict()}", false, 0, 1, false, 0},

		// Dictionary comprehensions
		{"simple comprehension", "{x: x**2 for x in range(5)}", false, 0, -1, true, 1},
		{"comp with condition", "{k: v for k, v in items if v > 0}", false, 0, -1, true, 1},
		{"expression transform", "{k.upper(): v * 2 for k, v in data.items()}", false, 0, -1, true, 1},
		{"nested generators", "{x: y for x in range(3) for y in range(3)}", false, 0, -1, true, 2},
		{"complex comprehension", "{k: process(v) for k, v in items if validate(k)}", false, 0, -1, true, 1},

		// Error cases
		{"unclosed dict", "{'key': 'value'", true, 0, 0, false, 0},
		{"missing colon", "{'key' 'value'}", true, 0, 0, false, 0},
		{"missing value", "{'key': }", true, 0, 0, false, 0},
		{"missing key", "{: 'value'}", true, 0, 0, false, 0},
		{"invalid unpack", "{***other}", true, 0, 0, false, 0},
		{"empty element", "{'a': 1, , 'b': 2}", true, 0, 0, false, 0},
		{"invalid comprehension", "{x: for x in items}", true, 0, 0, false, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var expr ast.Expr
			var err error

			if test.isDictComp {
				expr, err = parseDictOrSet(t, test.input)
			} else {
				expr, err = parseDict(t, test.input)
			}

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.input, err)
			}

			if test.isDictComp {
				validateDictComp(t, expr, test.expectedGenCount)
			} else {
				validateDictExpr(t, expr, test.expectedKV, test.expectedUnpack)
			}
		})
	}
}

// Test dictionary edge cases and complex scenarios
func TestDictionaryEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"function call as key", "{func(): 'value'}", false},
		{"list comprehension as value", "{'key': [x for x in range(3)]}", false},
		{"nested comprehensions", "{k: {v: v*2 for v in values} for k in keys}", false},
		{"unpack with expression", "{**{x: x*2 for x in range(3)}}", false},
		{"complex key expressions", "{(a, b): c for a in range(2) for b in range(2) for c in range(2)}", false},
		{"mixed literals and expressions", "{1: 'one', variable: computed_value, **rest}", false},
		{"empty dict in comprehension", "{k: {} for k in keys}", false},
		{"starred expression as value", "{'key': *values}", true}, // Invalid syntax
		{"double colon", "{'key':: 'value'}", true},
		{"comprehension missing colon", "{x x for x in items}", true},
		{"unpack in key position", "{**other: 'value'}", true},
		{"partial unpacking", "{**}", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Try both parsing methods since some syntax might work in different contexts
			expr1, err1 := parseDict(t, test.input)
			expr2, err2 := parseDictOrSet(t, test.input)

			hasAnyError := err1 != nil && err2 != nil
			hasAnySuccess := err1 == nil || err2 == nil

			if test.hasError {
				if !hasAnyError {
					t.Errorf("Expected error for %s, but parsing succeeded", test.input)
				}
			} else {
				if !hasAnySuccess {
					t.Errorf("Expected success for %s, but both parsers failed: %v, %v", test.input, err1, err2)
				}
				// At least one should succeed and return a valid expression
				if hasAnySuccess && expr1 == nil && expr2 == nil {
					t.Errorf("Parsing succeeded but returned nil expressions")
				}
			}
		})
	}
}
