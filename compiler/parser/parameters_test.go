package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a function statement
func parseFunctionStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.statement()
}

// Helper function to validate function definition structure
func validateFunctionDefinition(t *testing.T, stmt ast.Stmt, expectedParams int) *ast.Function {
	funcDef, ok := stmt.(*ast.Function)
	if !ok {
		t.Fatalf("Expected *ast.Function, got %T", stmt)
	}

	if funcDef.Parameters == nil {
		t.Fatal("Function should have parameters object")
	}

	if expectedParams >= 0 && len(funcDef.Parameters.Parameters) != expectedParams {
		t.Errorf("Expected %d parameters, got %d", expectedParams, len(funcDef.Parameters.Parameters))
	}

	return funcDef
}

// Helper function to validate parameter characteristics
func validateParameterListCharacteristics(t *testing.T, params *ast.ParameterList, hasSlash, hasVarArg, hasKwArg bool) {
	if params.HasSlash != hasSlash {
		t.Errorf("Expected HasSlash=%v, got %v", hasSlash, params.HasSlash)
	}

	if params.HasVarArg != hasVarArg {
		t.Errorf("Expected HasVarArg=%v, got %v", hasVarArg, params.HasVarArg)
	}

	if params.HasKwArg != hasKwArg {
		t.Errorf("Expected HasKwArg=%v, got %v", hasKwArg, params.HasKwArg)
	}
}

// Helper function to validate individual parameter properties
func validateParameter(t *testing.T, param *ast.Parameter, expectedName string, hasDefault, hasAnnotation, isStar, isDoubleStar bool) {
	if param.Name.Token.Lexeme != expectedName {
		t.Errorf("Expected parameter name %q, got %q", expectedName, param.Name.Token.Lexeme)
	}

	if (param.Default != nil) != hasDefault {
		t.Errorf("Expected hasDefault=%v, got %v", hasDefault, param.Default != nil)
	}

	if (param.Annotation != nil) != hasAnnotation {
		t.Errorf("Expected hasAnnotation=%v, got %v", hasAnnotation, param.Annotation != nil)
	}

	if param.IsStar != isStar {
		t.Errorf("Expected IsStar=%v, got %v", isStar, param.IsStar)
	}

	if param.IsDoubleStar != isDoubleStar {
		t.Errorf("Expected IsDoubleStar=%v, got %v", isDoubleStar, param.IsDoubleStar)
	}
}

// Helper function to get expression type string
func getExpressionTypeString(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.Literal:
		return "literal"
	case *ast.Name:
		return "name"
	case *ast.Binary:
		return "binary"
	case *ast.Call:
		return "call"
	case *ast.ListExpr:
		return "list"
	case *ast.DictExpr:
		return "dict"
	case *ast.Subscript:
		return "subscript"
	case *ast.Attribute:
		return "attribute"
	default:
		return "unknown"
	}
}

// Helper function to categorize parameter type
func getParameterCategory(param *ast.Parameter) string {
	switch {
	case param.IsDoubleStar:
		return "doublestar"
	case param.IsStar:
		return "star"
	case param.Default != nil:
		return "pos_default"
	default:
		return "pos"
	}
}

// Helper function to count parameter categories
func countParameterCategories(params *ast.ParameterList) (pos, defaults, star, doublestar int) {
	for _, param := range params.Parameters {
		switch {
		case param.IsDoubleStar:
			doublestar++
		case param.IsStar:
			star++
		case param.Default != nil:
			defaults++
		default:
			pos++
		}
	}
	return
}

// Test comprehensive parameter functionality
func TestParameterParsing(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedParams int
		hasSlash       bool
		hasVarArg      bool
		hasKwArg       bool
		hasError       bool
		description    string
	}{
		// Basic parameter patterns
		{
			name:           "no parameters",
			input:          "def func(): pass",
			expectedParams: 0,
			description:    "function with no parameters",
		},
		{
			name:           "single parameter",
			input:          "def func(x): pass",
			expectedParams: 1,
			description:    "function with single parameter",
		},
		{
			name:           "multiple parameters",
			input:          "def func(x, y, z): pass",
			expectedParams: 3,
			description:    "function with multiple parameters",
		},
		{
			name:           "parameter with default",
			input:          "def func(x=42): pass",
			expectedParams: 1,
			description:    "function with parameter having default value",
		},
		{
			name:           "mixed parameters with defaults",
			input:          "def func(x, y=2, z=3): pass",
			expectedParams: 3,
			description:    "function with mixed regular and default parameters",
		},

		// Special parameter types
		{
			name:           "varargs parameter",
			input:          "def func(*args): pass",
			expectedParams: 1,
			hasVarArg:      true,
			description:    "function with *args parameter",
		},
		{
			name:           "kwargs parameter",
			input:          "def func(**kwargs): pass",
			expectedParams: 1,
			hasKwArg:       true,
			description:    "function with **kwargs parameter",
		},
		{
			name:           "all parameter types",
			input:          "def func(a, b=2, *args, c, d=4, **kwargs): pass",
			expectedParams: 6,
			hasVarArg:      true,
			hasKwArg:       true,
			description:    "function with all parameter types combined",
		},

		// Positional-only parameters
		{
			name:           "positional-only parameters",
			input:          "def func(a, b, /): pass",
			expectedParams: 2,
			hasSlash:       true,
			description:    "function with positional-only parameters",
		},
		{
			name:           "positional-only with regular",
			input:          "def func(a, /, b, c): pass",
			expectedParams: 3,
			hasSlash:       true,
			description:    "function mixing positional-only and regular parameters",
		},
		{
			name:           "keyword-only parameters",
			input:          "def func(*, a, b): pass",
			expectedParams: 2,
			description:    "function with keyword-only parameters",
		},

		// Complex parameter combinations
		{
			name:           "complex parameter ordering",
			input:          "def func(a, /, b, c=1, *args, d, e=2, **kwargs): pass",
			expectedParams: 7,
			hasSlash:       true,
			hasVarArg:      true,
			hasKwArg:       true,
			description:    "function with complex parameter ordering",
		},
		{
			name:           "annotated parameters",
			input:          "def func(x: int, y: str = 'default'): pass",
			expectedParams: 2,
			description:    "function with type-annotated parameters",
		},

		// Error cases
		{
			name:        "default before non-default",
			input:       "def func(x=1, y): pass",
			hasError:    true,
			description: "invalid: default parameter before non-default",
		},
		{
			name:        "multiple slashes",
			input:       "def func(a, /, b, /): pass",
			hasError:    true,
			description: "invalid: multiple positional-only separators",
		},
		{
			name:        "multiple varargs",
			input:       "def func(*args1, *args2): pass",
			hasError:    true,
			description: "invalid: multiple *args parameters",
		},
		{
			name:        "kwargs not last",
			input:       "def func(**kwargs, x): pass",
			hasError:    true,
			description: "invalid: **kwargs not as last parameter",
		},
		{
			name:        "slash after varargs",
			input:       "def func(*args, /): pass",
			hasError:    true,
			description: "invalid: slash after *args",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunctionDefinition(t, stmt, test.expectedParams)
			validateParameterListCharacteristics(t, funcDef.Parameters, test.hasSlash, test.hasVarArg, test.hasKwArg)
		})
	}
}

// Test parameter type characteristics
func TestParameterTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		paramIndex    int
		expectedName  string
		hasDefault    bool
		hasAnnotation bool
		isStar        bool
		isDoubleStar  bool
		hasError      bool
		description   string
	}{
		{
			name:         "simple parameter",
			input:        "def func(x): pass",
			paramIndex:   0,
			expectedName: "x",
			description:  "basic parameter without annotations or defaults",
		},
		{
			name:         "parameter with default",
			input:        "def func(x=42): pass",
			paramIndex:   0,
			expectedName: "x",
			hasDefault:   true,
			description:  "parameter with default value",
		},
		{
			name:          "parameter with annotation",
			input:         "def func(x: int): pass",
			paramIndex:    0,
			expectedName:  "x",
			hasAnnotation: true,
			description:   "parameter with type annotation",
		},
		{
			name:          "parameter with annotation and default",
			input:         "def func(x: int = 42): pass",
			paramIndex:    0,
			expectedName:  "x",
			hasAnnotation: true,
			hasDefault:    true,
			description:   "parameter with both type annotation and default",
		},
		{
			name:         "star parameter",
			input:        "def func(*args): pass",
			paramIndex:   0,
			expectedName: "args",
			isStar:       true,
			description:  "*args parameter",
		},
		{
			name:         "double star parameter",
			input:        "def func(**kwargs): pass",
			paramIndex:   0,
			expectedName: "kwargs",
			isDoubleStar: true,
			description:  "**kwargs parameter",
		},
		{
			name:          "annotated star parameter",
			input:         "def func(*args: list): pass",
			paramIndex:    0,
			expectedName:  "args",
			isStar:        true,
			hasAnnotation: true,
			description:   "*args parameter with type annotation",
		},
		{
			name:          "annotated kwargs parameter",
			input:         "def func(**kwargs: dict): pass",
			paramIndex:    0,
			expectedName:  "kwargs",
			isDoubleStar:  true,
			hasAnnotation: true,
			description:   "**kwargs parameter with type annotation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunctionDefinition(t, stmt, -1)

			if test.paramIndex >= len(funcDef.Parameters.Parameters) {
				t.Errorf("Parameter index %d out of range (have %d parameters)",
					test.paramIndex, len(funcDef.Parameters.Parameters))
				return
			}

			param := funcDef.Parameters.Parameters[test.paramIndex]
			validateParameter(t, param, test.expectedName, test.hasDefault,
				test.hasAnnotation, test.isStar, test.isDoubleStar)
		})
	}
}

// Test parameter default value expressions
func TestParameterDefaults(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		paramIndex    int
		expectedType  string
		expectedValue interface{}
		hasError      bool
		description   string
	}{
		{
			name:          "number default",
			input:         "def func(x=42): pass",
			paramIndex:    0,
			expectedType:  "literal",
			expectedValue: 42,
			description:   "parameter with numeric literal default",
		},
		{
			name:          "string default",
			input:         `def func(x="hello"): pass`,
			paramIndex:    0,
			expectedType:  "literal",
			expectedValue: "hello",
			description:   "parameter with string literal default",
		},
		{
			name:         "boolean default",
			input:        "def func(x=True): pass",
			paramIndex:   0,
			expectedType: "literal",
			description:  "parameter with boolean literal default",
		},
		{
			name:         "none default",
			input:        "def func(x=None): pass",
			paramIndex:   0,
			expectedType: "literal",
			description:  "parameter with None literal default",
		},
		{
			name:         "variable default",
			input:        "def func(x=default_value): pass",
			paramIndex:   0,
			expectedType: "name",
			description:  "parameter with variable name as default",
		},
		{
			name:         "expression default",
			input:        "def func(x=a + b): pass",
			paramIndex:   0,
			expectedType: "binary",
			description:  "parameter with binary expression default",
		},
		{
			name:         "function call default",
			input:        "def func(x=get_default()): pass",
			paramIndex:   0,
			expectedType: "call",
			description:  "parameter with function call default",
		},
		{
			name:         "list default",
			input:        "def func(x=[]): pass",
			paramIndex:   0,
			expectedType: "list",
			description:  "parameter with empty list default",
		},
		{
			name:         "dict default",
			input:        "def func(x={}): pass",
			paramIndex:   0,
			expectedType: "dict",
			description:  "parameter with empty dict default",
		},
		{
			name:         "complex expression default",
			input:        "def func(x=obj.method().value): pass",
			paramIndex:   0,
			expectedType: "attribute",
			description:  "parameter with complex chained expression default",
		},
		{
			name:         "conditional default",
			input:        "def func(x=a if condition else b): pass",
			paramIndex:   0,
			expectedType: "unknown",
			description:  "parameter with ternary expression default",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunctionDefinition(t, stmt, -1)

			if test.paramIndex >= len(funcDef.Parameters.Parameters) {
				t.Errorf("Parameter index %d out of range", test.paramIndex)
				return
			}

			param := funcDef.Parameters.Parameters[test.paramIndex]
			if param.Default == nil {
				t.Errorf("Expected parameter to have default value for %s", test.description)
				return
			}

			actualType := getExpressionTypeString(param.Default)
			if actualType != test.expectedType {
				t.Errorf("Expected default type %s, got %s (actual: %T) for %s",
					test.expectedType, actualType, param.Default, test.description)
			}
		})
	}
}

// Test parameter type annotations
func TestParameterAnnotations(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		paramIndex   int
		expectedType string
		hasError     bool
		description  string
	}{
		{
			name:         "simple type annotation",
			input:        "def func(x: int): pass",
			paramIndex:   0,
			expectedType: "name",
			description:  "parameter with simple type annotation",
		},
		{
			name:         "generic type annotation",
			input:        "def func(x: List[int]): pass",
			paramIndex:   0,
			expectedType: "subscript",
			description:  "parameter with generic type annotation",
		},
		{
			name:         "union type annotation",
			input:        "def func(x: int | str): pass",
			paramIndex:   0,
			expectedType: "binary",
			description:  "parameter with union type annotation",
		},
		{
			name:         "complex type annotation",
			input:        "def func(x: Dict[str, List[int]]): pass",
			paramIndex:   0,
			expectedType: "subscript",
			description:  "parameter with complex nested type annotation",
		},
		{
			name:         "optional type annotation",
			input:        "def func(x: Optional[str]): pass",
			paramIndex:   0,
			expectedType: "subscript",
			description:  "parameter with Optional type annotation",
		},
		{
			name:         "callable annotation",
			input:        "def func(f: Callable[[int], str]): pass",
			paramIndex:   0,
			expectedType: "subscript",
			description:  "parameter with Callable type annotation",
		},
		{
			name:         "qualified type annotation",
			input:        "def func(x: typing.Union[int, str]): pass",
			paramIndex:   0,
			expectedType: "subscript",
			description:  "parameter with qualified type annotation",
		},
		{
			name:         "string literal annotation",
			input:        `def func(x: "ForwardRef"): pass`,
			paramIndex:   0,
			expectedType: "literal",
			description:  "parameter with string literal type annotation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunctionDefinition(t, stmt, -1)

			if test.paramIndex >= len(funcDef.Parameters.Parameters) {
				t.Errorf("Parameter index %d out of range", test.paramIndex)
				return
			}

			param := funcDef.Parameters.Parameters[test.paramIndex]
			if param.Annotation == nil {
				t.Errorf("Expected parameter to have type annotation for %s", test.description)
				return
			}

			actualType := getExpressionTypeString(param.Annotation)
			if actualType != test.expectedType {
				t.Errorf("Expected annotation type %s, got %s (actual: %T) for %s",
					test.expectedType, actualType, param.Annotation, test.description)
			}
		})
	}
}

// Test parameter ordering and positioning rules
func TestParameterOrdering(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		positions   []string
		hasError    bool
		description string
	}{
		{
			name:        "correct ordering: positional then defaults",
			input:       "def func(a, b, c=1, d=2): pass",
			positions:   []string{"pos", "pos", "pos_default", "pos_default"},
			description: "valid parameter ordering with defaults after positional",
		},
		{
			name:        "correct ordering with varargs",
			input:       "def func(a, b=1, *args, c, d=2, **kwargs): pass",
			positions:   []string{"pos", "pos_default", "star", "pos", "pos_default", "doublestar"},
			description: "valid complex parameter ordering with *args and **kwargs",
		},
		{
			name:        "positional-only parameters",
			input:       "def func(a, b, /, c, d=1): pass",
			positions:   []string{"pos", "pos", "pos", "pos_default"},
			description: "valid positional-only parameter ordering",
		},
		{
			name:        "keyword-only parameters",
			input:       "def func(*, a, b=1): pass",
			positions:   []string{"pos", "pos_default"},
			description: "valid keyword-only parameter ordering",
		},
		{
			name:        "complex parameter ordering",
			input:       "def func(a, /, b, c=1, *args, d, e=2, **kwargs): pass",
			positions:   []string{"pos", "pos", "pos_default", "star", "pos", "pos_default", "doublestar"},
			description: "valid complex parameter ordering with all types",
		},
		{
			name:        "only varargs",
			input:       "def func(*args): pass",
			positions:   []string{"star"},
			description: "function with only *args parameter",
		},
		{
			name:        "only kwargs",
			input:       "def func(**kwargs): pass",
			positions:   []string{"doublestar"},
			description: "function with only **kwargs parameter",
		},
		{
			name:        "bare star for keyword-only",
			input:       "def func(a, *, b): pass",
			positions:   []string{"pos", "pos"},
			description: "function with bare * for keyword-only parameters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunctionDefinition(t, stmt, len(test.positions))

			for i, expectedType := range test.positions {
				if i >= len(funcDef.Parameters.Parameters) {
					t.Errorf("Parameter index %d out of range for %s", i, test.description)
					continue
				}

				param := funcDef.Parameters.Parameters[i]
				actualType := getParameterCategory(param)

				if actualType != expectedType {
					t.Errorf("Parameter %d: expected %s, got %s for %s",
						i, expectedType, actualType, test.description)
				}
			}
		})
	}
}

// Test parameter error cases
func TestParameterErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		description   string
	}{
		{
			name:        "default before non-default",
			input:       "def func(x=1, y): pass",
			description: "invalid: default parameter before non-default parameter",
		},
		{
			name:        "multiple varargs",
			input:       "def func(*args1, *args2): pass",
			description: "invalid: multiple *args parameters",
		},
		{
			name:        "multiple kwargs",
			input:       "def func(**kwargs1, **kwargs2): pass",
			description: "invalid: multiple **kwargs parameters",
		},
		{
			name:        "multiple slashes",
			input:       "def func(a, /, b, /): pass",
			description: "invalid: multiple positional-only separators",
		},
		{
			name:        "slash after varargs",
			input:       "def func(*args, /): pass",
			description: "invalid: positional-only separator after *args",
		},
		{
			name:        "slash without parameters",
			input:       "def func(/): pass",
			description: "invalid: positional-only separator without parameters",
		},
		{
			name:        "kwargs not last",
			input:       "def func(**kwargs, x): pass",
			description: "invalid: **kwargs parameter not in last position",
		},
		{
			name:        "parameter after kwargs",
			input:       "def func(a, **kwargs, b): pass",
			description: "invalid: parameter after **kwargs",
		},
		{
			name:        "star without name",
			input:       "def func(*): pass",
			description: "bare * without following keyword-only parameters",
		},
		{
			name:        "doublestar without name",
			input:       "def func(**): pass",
			description: "invalid: ** without parameter name",
		},
		{
			name:        "invalid parameter name",
			input:       "def func(123): pass",
			description: "invalid: numeric parameter name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseFunctionStatement(t, test.input)

			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.description)
				return
			}

			if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
				t.Errorf("Expected error to contain %q, got %q for %s",
					test.errorContains, err.Error(), test.description)
			}
		})
	}
}

// Test parameter edge cases and complex scenarios
func TestParameterEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		// Complex valid cases
		{
			name:        "comprehensive parameter types",
			input:       `def func(a, b=1, /, c, d=2, *args, e, f=3, **kwargs): pass`,
			description: "function with all parameter types in correct order",
		},
		{
			name:        "complex annotations and defaults",
			input:       `def func(x: List[Dict[str, int]] = [], y: Optional[Callable[[int], str]] = None): pass`,
			description: "parameters with complex type annotations and defaults",
		},
		{
			name:        "nested generic annotations",
			input:       `def func(data: Dict[str, List[Tuple[int, str, bool]]]): pass`,
			description: "parameter with deeply nested generic type annotation",
		},
		{
			name:        "annotated special parameters",
			input:       `def func(*args: Tuple[int, ...], **kwargs: Any): pass`,
			description: "annotated *args and **kwargs parameters",
		},

		// Error cases
		{
			name:        "invalid default syntax",
			input:       "def func(x=): pass",
			hasError:    true,
			description: "parameter with incomplete default value",
		},
		{
			name:        "missing parameter name",
			input:       "def func(: int): pass",
			hasError:    true,
			description: "type annotation without parameter name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q for %s",
						test.errorContains, err.Error(), test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			// For successful cases, just verify we got a valid function
			if stmt == nil {
				t.Errorf("Expected valid statement but got nil for %s", test.description)
				return
			}

			validateFunctionDefinition(t, stmt, -1)
		})
	}
}
