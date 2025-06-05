package parser

import (
	"strings"
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a function definition
func parseFunctionDef(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.functionDef()
}

// Helper function to validate function structure
func validateFunction(t *testing.T, stmt ast.Stmt, expectedName string, expectedAsync bool) *ast.Function {
	funcDef, ok := stmt.(*ast.Function)
	if !ok {
		t.Fatalf("Expected *ast.Function, got %T", stmt)
	}

	if funcDef.Name == nil {
		t.Fatal("Function missing name")
	}

	if funcDef.Name.Token.Lexeme != expectedName {
		t.Errorf("Expected function name %s, got %s", expectedName, funcDef.Name.Token.Lexeme)
	}

	if funcDef.IsAsync != expectedAsync {
		t.Errorf("Expected isAsync=%v, got %v", expectedAsync, funcDef.IsAsync)
	}

	if len(funcDef.Body) == 0 {
		t.Error("Function missing body")
	}

	return funcDef
}

// Helper function to validate parameter list
func validateParameterList(t *testing.T, params *ast.ParameterList, expectedCount int, expectedVarArg, expectedKwArg, expectedSlash bool) {
	if params == nil {
		if expectedCount > 0 {
			t.Error("Expected parameters but got nil")
		}
		return
	}

	if expectedCount >= 0 && len(params.Parameters) != expectedCount {
		t.Errorf("Expected %d parameters, got %d", expectedCount, len(params.Parameters))
	}

	if params.HasVarArg != expectedVarArg {
		t.Errorf("Expected hasVarArg=%v, got %v", expectedVarArg, params.HasVarArg)
	}

	if params.HasKwArg != expectedKwArg {
		t.Errorf("Expected hasKwArg=%v, got %v", expectedKwArg, params.HasKwArg)
	}

	if params.HasSlash != expectedSlash {
		t.Errorf("Expected hasSlash=%v, got %v", expectedSlash, params.HasSlash)
	}
}

// Helper function to count parameters by type
func countParameterTypes(params *ast.ParameterList) (regular, withDefaults, starred, doubleStarred int) {
	if params == nil {
		return 0, 0, 0, 0
	}

	for _, param := range params.Parameters {
		if param.IsDoubleStar {
			doubleStarred++
		} else if param.IsStar {
			starred++
		} else if param.Default != nil {
			withDefaults++
		} else {
			regular++
		}
	}
	return
}

// Helper function to validate parameter annotations
func validateParameterAnnotations(t *testing.T, params *ast.ParameterList, expectedAnnotated int) {
	if params == nil {
		return
	}

	annotatedCount := 0
	for _, param := range params.Parameters {
		if param.Annotation != nil {
			annotatedCount++
		}
	}

	if annotatedCount != expectedAnnotated {
		t.Errorf("Expected %d annotated parameters, got %d", expectedAnnotated, annotatedCount)
	}
}

// Test comprehensive function definition functionality
func TestFunctionDefinition(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		hasError          bool
		expectedName      string
		expectedAsync     bool
		expectedReturn    bool
		expectedParams    int
		expectedVarArg    bool
		expectedKwArg     bool
		expectedSlash     bool
		expectedAnnotated int
		description       string
	}{
		// Basic functions
		{
			name: "simple function",
			input: `def greet():
    print("Hello")`,
			expectedName:   "greet",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 0,
			description:    "function without parameters",
		},
		{
			name: "function with parameters",
			input: `def add(x, y):
    return x + y`,
			expectedName:   "add",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 2,
			description:    "function with simple parameters",
		},
		{
			name: "function with defaults",
			input: `def greet(name="World", greeting="Hello"):
    print(f"{greeting}, {name}")`,
			expectedName:   "greet",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 2,
			description:    "function with default parameter values",
		},
		{
			name: "function with type annotations",
			input: `def add(x: int, y: int) -> int:
    return x + y`,
			expectedName:      "add",
			expectedAsync:     false,
			expectedReturn:    true,
			expectedParams:    2,
			expectedAnnotated: 2,
			description:       "function with type hints",
		},
		{
			name: "async function",
			input: `async def fetch_data():
    data = await get_data()
    return data`,
			expectedName:   "fetch_data",
			expectedAsync:  true,
			expectedReturn: false,
			expectedParams: 0,
			description:    "asynchronous function",
		},
		{
			name: "function with varargs",
			input: `def func(*args, **kwargs):
    pass`,
			expectedName:   "func",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 2,
			expectedVarArg: true,
			expectedKwArg:  true,
			description:    "function with *args and **kwargs",
		},
		{
			name: "function with positional-only params",
			input: `def func(a, b, /, c, d):
    pass`,
			expectedName:   "func",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 4,
			expectedSlash:  true,
			description:    "function with positional-only parameters",
		},
		{
			name: "function with keyword-only params",
			input: `def func(a, b, *, c, d):
    pass`,
			expectedName:   "func",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 4,
			description:    "function with keyword-only parameters",
		},
		{
			name: "complex function signature",
			input: `def process(data: list[int], *, validate: bool = True) -> Optional[int]:
    if validate:
        check(data)
    return max(data) if data else None`,
			expectedName:      "process",
			expectedAsync:     false,
			expectedReturn:    true,
			expectedParams:    2,
			expectedAnnotated: 2,
			description:       "function with complex signature",
		},
		{
			name: "function with all parameter types",
			input: `def func(a, b, /, c, d, *args, e, f, **kwargs):
    pass`,
			expectedName:   "func",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 8,
			expectedVarArg: true,
			expectedKwArg:  true,
			expectedSlash:  true,
			description:    "function with all parameter types",
		},
		{
			name: "function with docstring",
			input: `def calculate(x, y):
    """Calculate something important."""
    return x * y + 42`,
			expectedName:   "calculate",
			expectedAsync:  false,
			expectedReturn: false,
			expectedParams: 2,
			description:    "function with docstring",
		},
		{
			name: "function with complex return type",
			input: `def factory() -> Callable[[int], str]:
    return str`,
			expectedName:   "factory",
			expectedAsync:  false,
			expectedReturn: true,
			expectedParams: 0,
			description:    "function with callable return type",
		},
		{
			name: "function with union return type",
			input: `def func() -> int | str:
    return 42`,
			expectedName:   "func",
			expectedAsync:  false,
			expectedReturn: true,
			expectedParams: 0,
			description:    "function with union return type",
		},
		{
			name: "async function with complex signature",
			input: `async def process_async(data: list[str], *, timeout: float = 5.0) -> AsyncIterator[str]:
    for item in data:
        await asyncio.sleep(0.1)
        yield item.upper()`,
			expectedName:      "process_async",
			expectedAsync:     true,
			expectedReturn:    true,
			expectedParams:    2,
			expectedAnnotated: 2,
			description:       "async function with complex signature",
		},

		// Error cases
		{
			name:        "missing function name",
			input:       "def ():\n    pass",
			hasError:    true,
			description: "function without name",
		},
		{
			name:        "missing colon",
			input:       "def func()\n    pass",
			hasError:    true,
			description: "function without colon",
		},
		{
			name:        "missing parentheses",
			input:       "def func:\n    pass",
			hasError:    true,
			description: "function without parentheses",
		},
		{
			name:        "invalid parameter",
			input:       "def func(123):\n    pass",
			hasError:    true,
			description: "function with invalid parameter",
		},
		{
			name:        "empty body",
			input:       "def func():",
			hasError:    true,
			description: "function with empty body",
		},
		{
			name:        "invalid default value",
			input:       "def func(a=):\n    pass",
			hasError:    true,
			description: "function with incomplete default value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionDef(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunction(t, stmt, test.expectedName, test.expectedAsync)

			// Validate return type
			hasReturn := funcDef.ReturnType != nil
			if hasReturn != test.expectedReturn {
				t.Errorf("Expected hasReturn=%v, got %v", test.expectedReturn, hasReturn)
			}

			// Validate parameters
			validateParameterList(t, funcDef.Parameters, test.expectedParams, test.expectedVarArg, test.expectedKwArg, test.expectedSlash)

			// Validate annotations if specified
			if test.expectedAnnotated > 0 {
				validateParameterAnnotations(t, funcDef.Parameters, test.expectedAnnotated)
			}
		})
	}
}

// Test specific parameter patterns and edge cases
func TestFunctionParameterPatterns(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		hasError           bool
		expectedRegular    int
		expectedDefaults   int
		expectedStarred    int
		expectedDblStarred int
		description        string
	}{
		{
			name: "only regular parameters",
			input: `def func(a, b, c):
    pass`,
			expectedRegular: 3,
			description:     "function with only regular parameters",
		},
		{
			name: "mixed regular and default parameters",
			input: `def func(a, b, c=1, d=2):
    pass`,
			expectedRegular:  2,
			expectedDefaults: 2,
			description:      "function with regular and default parameters",
		},
		{
			name: "only starred parameters",
			input: `def func(*args, **kwargs):
    pass`,
			expectedStarred:    1,
			expectedDblStarred: 1,
			description:        "function with only starred parameters",
		},
		{
			name: "complete parameter pattern",
			input: `def func(a, b=1, *args, c, d=2, **kwargs):
    pass`,
			expectedRegular:    2, // a and c (keyword-only regular parameter)
			expectedDefaults:   2, // b and d (b=1, d=2)
			expectedStarred:    1,
			expectedDblStarred: 1,
			description:        "function with complete parameter pattern",
		},
		{
			name: "positional-only with defaults",
			input: `def func(a, b=1, /, c=2, d=3):
    pass`,
			expectedRegular:  1,
			expectedDefaults: 3,
			description:      "function with positional-only and default parameters",
		},
		{
			name: "keyword-only parameters",
			input: `def func(a, *, b, c=1):
    pass`,
			expectedRegular:  2, // a and b (both are parameters, b is keyword-only regular)
			expectedDefaults: 1, // only c has default value
			description:      "function with keyword-only parameters",
		},
		{
			name: "annotated parameters with defaults",
			input: `def func(a: int, b: str = "hello", *args: int, **kwargs: Any):
    pass`,
			expectedRegular:    1,
			expectedDefaults:   1,
			expectedStarred:    1,
			expectedDblStarred: 1,
			description:        "function with annotated parameters and defaults",
		},

		// Error cases
		{
			name:        "multiple varargs",
			input:       "def func(*args1, *args2):\n    pass",
			hasError:    true,
			description: "function with multiple *args",
		},
		{
			name:        "multiple kwargs",
			input:       "def func(**kw1, **kw2):\n    pass",
			hasError:    true,
			description: "function with multiple **kwargs",
		},
		{
			name:        "kwargs before varargs",
			input:       "def func(**kwargs, *args):\n    pass",
			hasError:    true,
			description: "function with **kwargs before *args",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionDef(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef := validateFunction(t, stmt, "func", false)

			regular, defaults, starred, dblStarred := countParameterTypes(funcDef.Parameters)

			if regular != test.expectedRegular {
				t.Errorf("Expected %d regular parameters, got %d", test.expectedRegular, regular)
			}
			if defaults != test.expectedDefaults {
				t.Errorf("Expected %d parameters with defaults, got %d", test.expectedDefaults, defaults)
			}
			if starred != test.expectedStarred {
				t.Errorf("Expected %d starred parameters, got %d", test.expectedStarred, starred)
			}
			if dblStarred != test.expectedDblStarred {
				t.Errorf("Expected %d double-starred parameters, got %d", test.expectedDblStarred, dblStarred)
			}
		})
	}
}

// Test return type annotations
func TestFunctionReturnTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		hasReturn   bool
		description string
	}{
		{
			name: "simple return type",
			input: `def func() -> int:
    return 42`,
			hasReturn:   true,
			description: "function with simple return type",
		},
		{
			name: "complex generic return type",
			input: `def func() -> list[tuple[int, str]]:
    return [(1, "one")]`,
			hasReturn:   true,
			description: "function with complex generic return type",
		},
		{
			name: "optional return type",
			input: `def func() -> Optional[int]:
    return None`,
			hasReturn:   true,
			description: "function with optional return type",
		},
		{
			name: "union return type",
			input: `def func() -> int | str:
    return 42`,
			hasReturn:   true,
			description: "function with union return type",
		},
		{
			name: "callable return type",
			input: `def factory() -> Callable[[int], str]:
    return str`,
			hasReturn:   true,
			description: "function with callable return type",
		},
		{
			name: "no return type annotation",
			input: `def func():
    return 42`,
			hasReturn:   false,
			description: "function without return type annotation",
		},
		{
			name: "async function with return type",
			input: `async def func() -> Awaitable[int]:
    return 42`,
			hasReturn:   true,
			description: "async function with return type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionDef(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			funcDef, ok := stmt.(*ast.Function)
			if !ok {
				t.Fatalf("Expected *ast.Function, got %T", stmt)
			}

			hasReturn := funcDef.ReturnType != nil
			if hasReturn != test.hasReturn {
				t.Errorf("Expected hasReturn=%v, got %v", test.hasReturn, hasReturn)
			}
		})
	}
}

// Test function edge cases and error scenarios
func TestFunctionEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		{
			name: "function with complex nested body",
			input: `def complex_func(data: dict[str, Any]) -> bool:
    for key, value in data.items():
        if isinstance(value, dict):
            if not complex_func(value):
                return False
        elif not validate_value(value):
            return False
    return True`,
			hasError:    false,
			description: "function with complex nested logic",
		},
		{
			name: "function with lambda in default",
			input: `def func(processor=lambda x: x * 2):
    return processor(42)`,
			hasError:    false,
			description: "function with lambda as default value",
		},
		{
			name: "function with comprehension in default",
			input: `def func(items=[x for x in range(5)]):
    return sum(items)`,
			hasError:    false,
			description: "function with comprehension as default value",
		},
		{
			name: "function with complex annotation",
			input: `def func(data: dict[str, list[tuple[int, Optional[str]]]]) -> Iterator[str]:
    for key, values in data.items():
        for value, text in values:
            if text:
                yield f"{key}: {text}"`,
			hasError:    false,
			description: "function with complex type annotations",
		},

		// Error cases with specific error messages
		{
			name:          "invalid syntax in parameter list",
			input:         "def func(a,, b):\n    pass",
			hasError:      true,
			errorContains: "expected",
			description:   "function with invalid parameter syntax",
		},
		{
			name:          "invalid return type syntax",
			input:         "def func() -> :\n    pass",
			hasError:      true,
			errorContains: "expected",
			description:   "function with invalid return type syntax",
		},
		{
			name:        "function with invalid indentation",
			input:       "def func():\npass",
			hasError:    true,
			description: "function with incorrect indentation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseFunctionDef(t, test.input)

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

			// For successful cases, just verify we got a valid function
			// Don't validate specific name for edge cases as they vary
			if stmt == nil {
				t.Errorf("Expected valid function statement for %s", test.description)
			}
		})
	}
}
