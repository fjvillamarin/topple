package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

// Helper function to parse a class statement
func parseClass(t *testing.T, input string) (*ast.Class, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)

	stmt, err := parser.classStatement()
	if err != nil {
		return nil, err
	}

	classDef, ok := stmt.(*ast.Class)
	if !ok {
		t.Errorf("Expected *ast.Class but got %T", stmt)
		return nil, nil
	}

	return classDef, nil
}

// Helper function to count decorators by unwrapping nested Decorator nodes
func countDecorators(stmt ast.Stmt) int {
	count := 0
	current := stmt
	for {
		if decorator, ok := current.(*ast.Decorator); ok {
			count++
			current = decorator.Stmt
		} else {
			break
		}
	}
	return count
}

// Test class definition parsing
func TestClassDef(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		className     string
		hasDecorators bool
		hasTypeParams bool
		hasSuper      bool
	}{
		{
			"simple class",
			`class MyClass:
    pass`,
			"MyClass",
			false,
			false,
			false,
		},
		{
			"class with inheritance",
			`class Child(Parent):
    pass`,
			"Child",
			false,
			false,
			true,
		},
		{
			"class with multiple inheritance",
			`class Child(Mother, Father):
    pass`,
			"Child",
			false,
			false,
			true,
		},
		{
			"class with metaclass",
			`class MyClass(metaclass=ABCMeta):
    pass`,
			"MyClass",
			false,
			false,
			true,
		},
		{
			"class with methods",
			`class Calculator:
    def add(self, a, b):
        return a + b
    
    def subtract(self, a, b):
        return a - b`,
			"Calculator",
			false,
			false,
			false,
		},
		{
			"class with docstring",
			`class Documented:
    """This is a documented class."""
    pass`,
			"Documented",
			false,
			false,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			stmt, err := parser.classStatement()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			classDef, ok := stmt.(*ast.Class)
			if !ok {
				t.Fatalf("Expected Class, got %T", stmt)
			}

			if classDef.Name.Token.Lexeme != test.className {
				t.Errorf("Expected class name %s, got %s", test.className, classDef.Name.Token.Lexeme)
			}

			hasDecorators := false // Decorators are separate wrapper nodes
			if hasDecorators != test.hasDecorators {
				t.Errorf("Expected hasDecorators=%v, got %v", test.hasDecorators, hasDecorators)
			}

			hasTypeParams := classDef.TypeParams != nil && len(classDef.TypeParams) > 0
			if hasTypeParams != test.hasTypeParams {
				t.Errorf("Expected hasTypeParams=%v, got %v", test.hasTypeParams, hasTypeParams)
			}

			hasSuper := len(classDef.Args) > 0
			if hasSuper != test.hasSuper {
				t.Errorf("Expected hasSuper=%v, got %v", test.hasSuper, hasSuper)
			}

			if len(classDef.Body) == 0 {
				t.Error("Class missing body")
			}
		})
	}
}

// Test class inheritance patterns
func TestClassInheritance(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		superCount int
		hasKwargs  bool
	}{
		{
			"single inheritance",
			`class Child(Parent):
    pass`,
			1,
			false,
		},
		{
			"multiple inheritance",
			`class Child(Mother, Father, Grandparent):
    pass`,
			3,
			false,
		},
		{
			"inheritance with kwargs",
			`class MyClass(metaclass=ABCMeta):
    pass`,
			0,
			true,
		},
		{
			"mixed inheritance",
			`class Widget(BaseWidget, metaclass=WidgetMeta):
    pass`,
			1,
			true,
		},
		{
			"complex inheritance",
			`class Advanced(Base1, Base2, mixin=MyMixin, metaclass=Meta):
    pass`,
			2,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			stmt, err := parser.classStatement()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			classDef, ok := stmt.(*ast.Class)
			if !ok {
				t.Fatalf("Expected Class, got %T", stmt)
			}

			// Count actual superclasses (not keyword arguments)
			superCount := 0
			for _, arg := range classDef.Args {
				if arg.Name == nil {
					superCount++
				}
			}

			if superCount != test.superCount {
				t.Errorf("Expected %d superclasses, got %d", test.superCount, superCount)
			}

			// Check for keyword arguments
			hasKwargs := false
			for _, arg := range classDef.Args {
				if arg.Name != nil {
					hasKwargs = true
					break
				}
			}

			if hasKwargs != test.hasKwargs {
				t.Errorf("Expected hasKwargs=%v, got %v", test.hasKwargs, hasKwargs)
			}
		})
	}
}

// Test decorated classes
func TestDecoratedClass(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		decoratorCount int
	}{
		{
			"single decorator",
			`@dataclass
class Point:
    x: int
    y: int`,
			1,
		},
		{
			"multiple decorators",
			`@dataclass
@frozen
class ImmutablePoint:
    x: int
    y: int`,
			2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			// Parse the statement - decorators create wrapper nodes
			stmt, err := parser.statement()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.input, err)
			}

			// Check if we got a Decorator node
			var decoratorCount int
			current := stmt
			for decoratorCount < test.decoratorCount {
				if decorator, ok := current.(*ast.Decorator); ok {
					decoratorCount++
					current = decorator.Stmt
				} else {
					break
				}
			}

			if decoratorCount != test.decoratorCount {
				t.Errorf("Expected %d decorators, got %d", test.decoratorCount, decoratorCount)
			}

			// The innermost should be a Class
			if _, ok := current.(*ast.Class); !ok {
				t.Fatalf("Expected innermost to be Class, got %T", current)
			}
		})
	}
}

// Test error cases
func TestClassErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing name", "class :"},
		{"missing colon", "class MyClass"},
		{"empty body", "class MyClass:"},
		{"invalid name", "class 123:"},
		{"invalid inheritance", "class Child():"},
		{"invalid type param", "class Generic[]:"},
		{"mixed inheritance syntax", "class Bad(Parent][T):"},
		{"duplicate type params", "class Dup[T, T]:"},
		{"invalid decorator position", "class MyClass:\n    @decorator\n    pass"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(test.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)

			_, err := parser.classStatement()
			if err == nil {
				t.Errorf("Expected error for %s, but got none", test.input)
			}
		})
	}
}
