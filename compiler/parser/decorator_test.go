package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"testing"
)

// Helper function to parse a statement
func parseStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	return parser.statement()
}

// Helper function to validate decorator structure
func validateDecorator(t *testing.T, stmt ast.Stmt, expectedDecoratorCount int, expectedStmtType string) {
	decorator, ok := stmt.(*ast.Decorator)
	if !ok {
		t.Fatalf("Expected *ast.Decorator, got %T", stmt)
	}

	if decorator.Expr == nil {
		t.Error("Decorator expression should not be nil")
	}

	if decorator.Stmt == nil {
		t.Error("Decorated statement should not be nil")
	}

	// Count chained decorators if expected
	if expectedDecoratorCount > 0 {
		count := countChainedDecorators(decorator)
		if count != expectedDecoratorCount {
			t.Errorf("Expected %d decorators, got %d", expectedDecoratorCount, count)
		}
	}

	// Validate innermost statement type if specified
	if expectedStmtType != "" {
		innerStmt := getInnermostDecoratedStatement(decorator)
		actualType := getStmtTypeName(innerStmt)
		if actualType != expectedStmtType {
			t.Errorf("Expected decorated statement type %s, got %s", expectedStmtType, actualType)
		}
	}
}

// Helper function to count chained decorators
func countChainedDecorators(decorator *ast.Decorator) int {
	count := 1
	current := decorator.Stmt

	for {
		if innerDecorator, ok := current.(*ast.Decorator); ok {
			count++
			current = innerDecorator.Stmt
		} else {
			break
		}
	}

	return count
}

// Helper function to get the innermost decorated statement
func getInnermostDecoratedStatement(decorator *ast.Decorator) ast.Stmt {
	current := decorator.Stmt

	for {
		if innerDecorator, ok := current.(*ast.Decorator); ok {
			current = innerDecorator.Stmt
		} else {
			break
		}
	}

	return current
}

// Helper function to get expression type name
func getExprTypeName(expr ast.Expr) string {
	if expr == nil {
		return "nil"
	}
	
	switch expr.(type) {
	case *ast.Name:
		return "*ast.Name"
	case *ast.Attribute:
		return "*ast.Attribute"
	case *ast.Call:
		return "*ast.Call"
	case *ast.Literal:
		return "*ast.Literal"
	case *ast.Binary:
		return "*ast.Binary"
	case *ast.Lambda:
		return "*ast.Lambda"
	default:
		return "unknown"
	}
}

// Helper function to get statement type name
func getStmtTypeName(stmt ast.Stmt) string {
	if stmt == nil {
		return "nil"
	}
	
	switch stmt.(type) {
	case *ast.Function:
		return "*ast.Function"
	case *ast.Class:
		return "*ast.Class"
	case *ast.Decorator:
		return "*ast.Decorator"
	case *ast.AssignStmt:
		return "*ast.AssignStmt"
	case *ast.ExprStmt:
		return "*ast.ExprStmt"
	default:
		return "unknown"
	}
}

// Test comprehensive decorator functionality
func TestDecorators(t *testing.T) {
	tests := []struct {
		name                  string
		input                 string
		hasError              bool
		expectedDecoratorCount int
		expectedStmtType      string
		expectedExprType      string
	}{
		// Simple decorators
		{
			name: "simple function decorator",
			input: `@property
def get_value(self):
    return self._value`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Name",
		},
		{
			name: "simple class decorator", 
			input: `@dataclass
class Point:
    pass`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Class",
			expectedExprType:       "*ast.Name",
		},
		{
			name: "decorator with attribute access",
			input: `@app.route
def func():
    pass`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Attribute",
		},
		{
			name: "decorator with function call",
			input: `@cache(timeout=300)
def func():
    pass`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Call",
		},
		{
			name: "complex decorator expression",
			input: `@my_module.decorators.validate_input(schema=user_schema)
def create_user(data):
    return User(data)`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Call",
		},

		// Multiple decorators
		{
			name: "multiple decorators on function",
			input: `@staticmethod
@property
def get_class_value():
    return cls._value`,
			expectedDecoratorCount: 2,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Name",
		},
		{
			name: "multiple decorators on class",
			input: `@dataclass
@frozen
@validate
class ImmutablePoint:
    x: int
    y: int`,
			expectedDecoratorCount: 3,
			expectedStmtType:       "*ast.Class",
			expectedExprType:       "*ast.Name",
		},
		{
			name: "three decorators with different styles",
			input: `@app.route("/users/<int:user_id>")
@require_auth
@cache(timeout=600)
def get_user(user_id):
    return User.get(user_id)`,
			expectedDecoratorCount: 3,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Call",
		},

		// Decorated async function
		{
			name: "decorated async function",
			input: `@async_cache
async def fetch_data():
    return await api_call()`,
			expectedDecoratorCount: 1,
			expectedStmtType:       "*ast.Function",
			expectedExprType:       "*ast.Name",
		},

		// Error cases
		{
			name: "decorator without newline",
			input: `@property def get_value(self): return self._value`,
			hasError: true,
		},
		{
			name: "decorator on invalid statement",
			input: `@decorator
x = 5`,
			hasError: true,
		},
		{
			name: "decorator without expression",
			input: `@
def func():
    pass`,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			validateDecorator(t, stmt, test.expectedDecoratorCount, test.expectedStmtType)

			// Validate expression type if specified
			if test.expectedExprType != "" {
				decorator := stmt.(*ast.Decorator)
				actualExprType := getExprTypeName(decorator.Expr)
				if actualExprType != test.expectedExprType {
					t.Errorf("Expected decorator expression type %s, got %s", test.expectedExprType, actualExprType)
				}
			}
		})
	}
}

// Test decorator arguments and complex expressions
func TestDecoratorArguments(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		validateCall bool
	}{
		{
			name: "decorator with positional args",
			input: `@cache("users", 300)
def get_users():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with keyword args",
			input: `@validate(schema=user_schema, strict=True)
def create_user():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with mixed args",
			input: `@app.route("/api/users", methods=["GET", "POST"])
def users_endpoint():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with complex expressions",
			input: `@rate_limit(requests=get_config().max_requests, per="minute")
def api_call():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with lambda",
			input: `@validate_input(validator=lambda x: len(x) > 0)
def process_input():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with list comprehension",
			input: `@requires_permissions([p for p in all_perms if p.level > 5])
def admin_function():
    pass`,
			validateCall: true,
		},
		{
			name: "decorator with nested attribute access",
			input: `@config.app.middleware.auth.require_login
def protected_view():
    pass`,
			validateCall: false,
		},
		{
			name: "decorator with subscript",
			input: `@decorators["auth"]
def secure_endpoint():
    pass`,
			validateCall: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			decorator, ok := stmt.(*ast.Decorator)
			if !ok {
				t.Fatalf("Expected *ast.Decorator, got %T", stmt)
			}

			if decorator.Expr == nil {
				t.Error("Decorator expression should not be nil")
			}

			// Validate that call decorators have proper structure
			if test.validateCall {
				if _, ok := decorator.Expr.(*ast.Call); !ok {
					t.Errorf("Expected decorator expression to be *ast.Call, got %T", decorator.Expr)
				}
			}
		})
	}
}

// Test edge cases and error conditions
func TestDecoratorEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "decorator on nested function",
			input: `def outer():
    @inner_decorator
    def inner():
        pass
    return inner`,
			hasError: false,
		},
		{
			name: "multiple decorators with complex expressions",
			input: `@permission_required("admin")
@rate_limit(requests=100, per="hour")
@log_function_calls(level="INFO")
def admin_function():
    pass`,
			hasError: false,
		},
		// NOTE: Generator expressions in function calls not yet supported by parser
		// {
		// 	name: "decorator with generator expression",
		// 	input: `@requires_all(perm for perm in permissions if perm.active)
		// def protected_function():
		//     pass`,
		// 	hasError: false,
		// },
		{
			name: "empty decorator list (invalid)",
			input: `@
@property
def func():
    pass`,
			hasError: true,
		},
		{
			name: "decorator on import statement (invalid)",
			input: `@decorator
import module`,
			hasError: true,
		},
		{
			name: "decorator on return statement (invalid)",
			input: `@decorator
return value`,
			hasError: true,
		},
		{
			name: "nested decorators with calls",
			input: `@outer_decorator(inner_decorator(value))
def complex_function():
    pass`,
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// For valid cases, just verify we got a valid statement
			if stmt == nil {
				t.Error("Expected non-nil statement")
			}
		})
	}
}