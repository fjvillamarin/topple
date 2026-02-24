package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"testing"
)

func TestExpressions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		parser string // which parser method to test
		want   struct {
			nodeType  string
			hasError  bool
			errorMsg  string
			hasValue  bool   // for yield, return statements
			isFrom    bool   // for yield from
			isStar    bool   // for star expressions
			elemCount int    // for tuples, lists
			nameCount int    // for name lists
			operator  string // for binary expressions
		}
	}{
		// Basic expression parsing
		{
			name:   "expression/simple_identifier",
			input:  "x",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "name"},
		},
		{
			name:   "expression/lambda_expression",
			input:  "lambda x: x + 1",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "lambda"},
		},
		{
			name:   "expression/ternary_expression",
			input:  "x if condition else y",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "ternary"},
		},
		{
			name:   "expression/binary_or",
			input:  "a or b",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "or"},
		},
		{
			name:   "expression/binary_and",
			input:  "a and b",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "and"},
		},
		{
			name:   "expression/unary_not",
			input:  "not x",
			parser: "expression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "unary", operator: "not"},
		},

		// Disjunction tests
		{
			name:   "disjunction/simple_or",
			input:  "a or b",
			parser: "disjunction",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "or"},
		},
		{
			name:   "disjunction/multiple_or",
			input:  "a or b or c",
			parser: "disjunction",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "or"},
		},
		{
			name:   "disjunction/single_expression",
			input:  "x",
			parser: "disjunction",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "name"},
		},

		// Conjunction tests
		{
			name:   "conjunction/simple_and",
			input:  "a and b",
			parser: "conjunction",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "and"},
		},
		{
			name:   "conjunction/multiple_and",
			input:  "a and b and c",
			parser: "conjunction",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "and"},
		},

		// Inversion tests
		{
			name:   "inversion/simple_not",
			input:  "not x",
			parser: "inversion",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "unary", operator: "not"},
		},
		{
			name:   "inversion/nested_not",
			input:  "not not x",
			parser: "inversion",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "unary", operator: "not"},
		},
		{
			name:   "inversion/not_with_comparison",
			input:  "not x < y",
			parser: "inversion",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "unary", operator: "not"},
		},
		{
			name:   "inversion/expression_without_not",
			input:  "x < y",
			parser: "inversion",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "<"},
		},

		// Yield expression tests
		{
			name:   "yieldExpression/simple_yield",
			input:  "yield",
			parser: "yieldExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "yield", hasValue: false, isFrom: false},
		},
		{
			name:   "yieldExpression/yield_with_value",
			input:  "yield 42",
			parser: "yieldExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "yield", hasValue: true, isFrom: false},
		},
		{
			name:   "yieldExpression/yield_from",
			input:  "yield from items",
			parser: "yieldExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "yield", hasValue: true, isFrom: true},
		},

		// Star expression tests
		{
			name:   "starExpression/starred_expression",
			input:  "*args",
			parser: "starExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "star", isStar: true},
		},
		{
			name:   "starExpression/regular_expression",
			input:  "x",
			parser: "starExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "name", isStar: false},
		},
		{
			name:   "starExpression/starred_complex",
			input:  "*func()",
			parser: "starExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "star", isStar: true},
		},

		// Star expressions tests
		{
			name:   "starExpressions/single_expression",
			input:  "x",
			parser: "starExpressions",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "name"},
		},
		{
			name:   "starExpressions/single_starred",
			input:  "*args",
			parser: "starExpressions",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "star", isStar: true},
		},
		{
			name:   "starExpressions/tuple_with_trailing_comma",
			input:  "x,",
			parser: "starExpressions",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "tuple", elemCount: 1},
		},
		{
			name:   "starExpressions/tuple_of_expressions",
			input:  "x, y, z",
			parser: "starExpressions",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "tuple", elemCount: 3},
		},

		// Named expression tests
		{
			name:   "namedExpression/walrus_operator",
			input:  "x := 42",
			parser: "namedExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "assign"},
		},
		{
			name:   "namedExpression/walrus_with_expression",
			input:  "result := func()",
			parser: "namedExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "assign"},
		},
		{
			name:   "namedExpression/regular_expression",
			input:  "x + y",
			parser: "namedExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "binary", operator: "+"},
		},

		// Star named expression tests
		{
			name:   "starNamedExpression/starred_expression",
			input:  "*args",
			parser: "starNamedExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "star", isStar: true},
		},
		{
			name:   "starNamedExpression/named_with_walrus",
			input:  "x := 42",
			parser: "starNamedExpression",
			want: struct {
				nodeType  string
				hasError  bool
				errorMsg  string
				hasValue  bool
				isFrom    bool
				isStar    bool
				elemCount int
				nameCount int
				operator  string
			}{nodeType: "assign"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			var result interface{}
			var err error

			// Call the appropriate parser method
			switch tt.parser {
			case "expression":
				result, err = parser.expression()
			case "disjunction":
				result, err = parser.disjunction()
			case "conjunction":
				result, err = parser.conjunction()
			case "inversion":
				result, err = parser.inversion()
			case "yieldExpression":
				result, err = parser.yieldExpression()
			case "starExpression":
				result, err = parser.starExpression()
			case "starExpressions":
				result, err = parser.starExpressions()
			case "namedExpression":
				result, err = parser.namedExpression()
			case "starNamedExpression":
				result, err = parser.starNamedExpression()
			default:
				t.Fatalf("Unknown parser method: %s", tt.parser)
			}

			// Check error expectations
			if tt.want.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.want.errorMsg != "" && !containsString(err.Error(), tt.want.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.want.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			expr, ok := result.(ast.Expr)
			if !ok {
				t.Errorf("Expected ast.Expr, got %T", result)
				return
			}

			// Check node type
			actualType := getExpressionType(expr)
			if actualType != tt.want.nodeType {
				t.Errorf("Expected node type %q, got %q (actual: %T)", tt.want.nodeType, actualType, expr)
			}

			// Check specific properties based on node type
			switch tt.want.nodeType {
			case "yield":
				yieldExpr := expr.(*ast.YieldExpr)
				if (yieldExpr.Value != nil) != tt.want.hasValue {
					t.Errorf("Expected hasValue=%v, got %v", tt.want.hasValue, yieldExpr.Value != nil)
				}
				if yieldExpr.IsFrom != tt.want.isFrom {
					t.Errorf("Expected isFrom=%v, got %v", tt.want.isFrom, yieldExpr.IsFrom)
				}

			case "star":
				if !tt.want.isStar {
					t.Errorf("Expected isStar=true for star node type")
				}

			case "binary":
				if tt.want.operator != "" {
					binaryExpr := expr.(*ast.Binary)
					if binaryExpr.Operator.Lexeme != tt.want.operator {
						t.Errorf("Expected operator %q, got %q", tt.want.operator, binaryExpr.Operator.Lexeme)
					}
				}

			case "unary":
				if tt.want.operator != "" {
					unaryExpr := expr.(*ast.Unary)
					if unaryExpr.Operator.Lexeme != tt.want.operator {
						t.Errorf("Expected operator %q, got %q", tt.want.operator, unaryExpr.Operator.Lexeme)
					}
				}

			case "tuple":
				tupleExpr := expr.(*ast.TupleExpr)
				if len(tupleExpr.Elements) != tt.want.elemCount {
					t.Errorf("Expected %d elements, got %d", tt.want.elemCount, len(tupleExpr.Elements))
				}
			}

			// Check star flag for non-star nodes
			if tt.want.nodeType != "star" && tt.want.isStar {
				t.Errorf("Expected star node but got %s", actualType)
			}
		})
	}
}

func TestParseNameList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  struct {
			count    int
			hasError bool
			errorMsg string
		}
	}{
		{
			name:  "single_name",
			input: "x",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{count: 1},
		},
		{
			name:  "two_names",
			input: "x, y",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{count: 2},
		},
		{
			name:  "three_names",
			input: "a, b, c",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{count: 3},
		},
		{
			name:  "names_with_spaces",
			input: "first, second, third",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{count: 3},
		},
		{
			name:  "empty_input_error",
			input: "",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{hasError: true, errorMsg: "expected identifier"},
		},
		{
			name:  "trailing_comma_error",
			input: "x,",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{hasError: true, errorMsg: "expected identifier after ','"},
		},
		{
			name:  "non_identifier_error",
			input: "123",
			want: struct {
				count    int
				hasError bool
				errorMsg string
			}{hasError: true, errorMsg: "expected identifier"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			names, err := parser.parseNameList()

			if tt.want.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.want.errorMsg != "" && !containsString(err.Error(), tt.want.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.want.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(names) != tt.want.count {
				t.Errorf("Expected %d names, got %d", tt.want.count, len(names))
			}

			// Verify all results are Name nodes
			for i, name := range names {
				if name == nil {
					t.Errorf("Name %d is nil", i)
				}
			}
		})
	}
}

// Helper functions for better test maintainability

func getExpressionType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.Name:
		return "name"
	case *ast.Lambda:
		return "lambda"
	case *ast.TernaryExpr:
		return "ternary"
	case *ast.Binary:
		return "binary"
	case *ast.Unary:
		return "unary"
	case *ast.Call:
		return "call"
	case *ast.Attribute:
		return "attribute"
	case *ast.Subscript:
		return "subscript"
	case *ast.GroupExpr:
		return "group"
	case *ast.Literal:
		return "literal"
	case *ast.ListExpr:
		return "list"
	case *ast.DictExpr:
		return "dict"
	case *ast.SetExpr:
		return "set"
	case *ast.TupleExpr:
		return "tuple"
	case *ast.YieldExpr:
		return "yield"
	case *ast.StarExpr:
		return "star"
	case *ast.AssignExpr:
		return "assign"
	default:
		return "unknown"
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
