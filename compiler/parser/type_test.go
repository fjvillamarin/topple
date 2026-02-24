package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"testing"
)

func TestTypeAliasStatements(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		hasParams    bool
		paramCount   int
		hasError     bool
	}{
		{
			name:         "simple type alias",
			input:        "type StringList = List[str]",
			expectedName: "StringList",
			hasParams:    false,
		},
		{
			name:         "type alias with single parameter",
			input:        "type MyList[T] = List[T]",
			expectedName: "MyList",
			hasParams:    true,
			paramCount:   1,
		},
		{
			name:         "type alias with multiple parameters",
			input:        "type Dict[K, V] = dict[K, V]",
			expectedName: "Dict",
			hasParams:    true,
			paramCount:   2,
		},
		{
			name:         "type alias with complex expression",
			input:        "type IntOrStr = int | str",
			expectedName: "IntOrStr",
			hasParams:    false,
		},
		{
			name:         "type alias with callable",
			input:        "type Handler = Callable[[Request], Response]",
			expectedName: "Handler",
			hasParams:    false,
		},
		{
			name:     "type alias without name",
			input:    "type = List[str]",
			hasError: true,
		},
		{
			name:     "type alias without equals",
			input:    "type MyType List[str]",
			hasError: true,
		},
		{
			name:     "type alias without value",
			input:    "type MyType =",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			if typeAlias.Name.Lexeme != tt.expectedName {
				t.Errorf("Expected name %q but got %q", tt.expectedName, typeAlias.Name.Lexeme)
			}

			if tt.hasParams {
				if typeAlias.Params == nil {
					t.Errorf("Expected type parameters but got none")
				} else if len(typeAlias.Params) != tt.paramCount {
					t.Errorf("Expected %d type parameters but got %d", tt.paramCount, len(typeAlias.Params))
				}
			} else {
				if typeAlias.Params != nil && len(typeAlias.Params) > 0 {
					t.Errorf("Expected no type parameters but got %d", len(typeAlias.Params))
				}
			}

			if typeAlias.Value == nil {
				t.Errorf("Type alias should have a value")
			}
		})
	}
}

func TestTypeParameters(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		paramIndex   int
		expectedName string
		hasBound     bool
		hasDefault   bool
		isStar       bool
		isDoubleStar bool
		hasError     bool
	}{
		{
			name:         "simple type parameter",
			input:        "type Generic[T] = List[T]",
			paramIndex:   0,
			expectedName: "T",
		},
		{
			name:         "type parameter with bound",
			input:        "type Comparable[T: Ord] = T",
			paramIndex:   0,
			expectedName: "T",
			hasBound:     true,
		},
		{
			name:         "type parameter with default",
			input:        "type Optional[T = None] = T | None",
			paramIndex:   0,
			expectedName: "T",
			hasDefault:   true,
		},
		{
			name:         "type parameter with bound and default",
			input:        "type Numeric[T: Number = int] = T",
			paramIndex:   0,
			expectedName: "T",
			hasBound:     true,
			hasDefault:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			if tt.paramIndex >= len(typeAlias.Params) {
				t.Errorf("Parameter index %d out of range (have %d parameters)", tt.paramIndex, len(typeAlias.Params))
				return
			}

			param, ok := typeAlias.Params[tt.paramIndex].(*ast.TypeParam)
			if !ok {
				t.Errorf("Expected *ast.TypeParam but got %T", typeAlias.Params[tt.paramIndex])
				return
			}

			if param.Name.Lexeme != tt.expectedName {
				t.Errorf("Expected parameter name %q but got %q", tt.expectedName, param.Name.Lexeme)
			}

			if (param.Bound != nil) != tt.hasBound {
				t.Errorf("Expected hasBound=%v but got %v", tt.hasBound, param.Bound != nil)
			}

			if (param.Default != nil) != tt.hasDefault {
				t.Errorf("Expected hasDefault=%v but got %v", tt.hasDefault, param.Default != nil)
			}

			if param.IsStar != tt.isStar {
				t.Errorf("Expected IsStar=%v but got %v", tt.isStar, param.IsStar)
			}

			if param.IsDoubleStar != tt.isDoubleStar {
				t.Errorf("Expected IsDoubleStar=%v but got %v", tt.isDoubleStar, param.IsDoubleStar)
			}
		})
	}
}

func TestTypeParameterBounds(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		paramIndex   int
		expectedType string // "name", "subscript", "binary", etc.
		hasError     bool
	}{
		{
			name:         "simple bound",
			input:        "type Ordered[T: Comparable] = T",
			paramIndex:   0,
			expectedType: "name",
		},
		{
			name:         "generic bound",
			input:        "type Container[T: Iterable[Any]] = T",
			paramIndex:   0,
			expectedType: "subscript",
		},
		{
			name:         "union bound",
			input:        "type Number[T: int | float] = T",
			paramIndex:   0,
			expectedType: "binary",
		},
		{
			name:         "complex bound",
			input:        "type Callable[T: Callable[[int], str]] = T",
			paramIndex:   0,
			expectedType: "subscript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			param := typeAlias.Params[tt.paramIndex].(*ast.TypeParam)
			if param.Bound == nil {
				t.Errorf("Expected parameter to have bound")
				return
			}

			var actualType string
			switch param.Bound.(type) {
			case *ast.Name:
				actualType = "name"
			case *ast.Subscript:
				actualType = "subscript"
			case *ast.Binary:
				actualType = "binary"
			case *ast.Attribute:
				actualType = "attribute"
			default:
				actualType = "unknown"
			}

			if actualType != tt.expectedType {
				t.Errorf("Expected bound type %s but got %s (actual: %T)",
					tt.expectedType, actualType, param.Bound)
			}
		})
	}
}

func TestTypeParameterDefaults(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		paramIndex   int
		expectedType string // "name", "literal", "binary", etc.
		hasError     bool
	}{
		{
			name:         "literal default",
			input:        "type Optional[T = None] = T | None",
			paramIndex:   0,
			expectedType: "literal", // None is parsed as a literal
		},
		{
			name:         "type default",
			input:        "type Numeric[T = int] = T",
			paramIndex:   0,
			expectedType: "name",
		},
		{
			name:         "generic default",
			input:        "type Container[T = List[Any]] = T",
			paramIndex:   0,
			expectedType: "subscript",
		},
		{
			name:         "union default",
			input:        "type Flexible[T = int | str] = T",
			paramIndex:   0,
			expectedType: "binary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			param := typeAlias.Params[tt.paramIndex].(*ast.TypeParam)
			if param.Default == nil {
				t.Errorf("Expected parameter to have default")
				return
			}

			var actualType string
			switch param.Default.(type) {
			case *ast.Name:
				actualType = "name"
			case *ast.Literal:
				actualType = "literal"
			case *ast.Subscript:
				actualType = "subscript"
			case *ast.Binary:
				actualType = "binary"
			case *ast.StarExpr:
				actualType = "star"
			default:
				actualType = "unknown"
			}

			if actualType != tt.expectedType {
				t.Errorf("Expected default type %s but got %s (actual: %T)",
					tt.expectedType, actualType, param.Default)
			}
		})
	}
}

func TestMultipleTypeParameters(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		paramInfo []struct {
			name         string
			isStar       bool
			isDoubleStar bool
			hasBound     bool
			hasDefault   bool
		}
		hasError bool
	}{
		{
			name:  "two simple parameters",
			input: "type Pair[T, U] = tuple[T, U]",
			paramInfo: []struct {
				name         string
				isStar       bool
				isDoubleStar bool
				hasBound     bool
				hasDefault   bool
			}{
				{name: "T"},
				{name: "U"},
			},
		},
		{
			name:  "parameters with bounds",
			input: "type Ordered[T: Comparable, U: Hashable] = dict[T, U]",
			paramInfo: []struct {
				name         string
				isStar       bool
				isDoubleStar bool
				hasBound     bool
				hasDefault   bool
			}{
				{name: "T", hasBound: true},
				{name: "U", hasBound: true},
			},
		},
		{
			name:  "multiple simple parameters",
			input: "type Triple[T, U, V] = tuple[T, U, V]",
			paramInfo: []struct {
				name         string
				isStar       bool
				isDoubleStar bool
				hasBound     bool
				hasDefault   bool
			}{
				{name: "T"},
				{name: "U"},
				{name: "V"},
			},
		},
		{
			name:  "parameters with defaults",
			input: "type Optional[T, U = None] = T | U",
			paramInfo: []struct {
				name         string
				isStar       bool
				isDoubleStar bool
				hasBound     bool
				hasDefault   bool
			}{
				{name: "T"},
				{name: "U", hasDefault: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			if len(typeAlias.Params) != len(tt.paramInfo) {
				t.Errorf("Expected %d parameters but got %d", len(tt.paramInfo), len(typeAlias.Params))
				return
			}

			for i, expectedInfo := range tt.paramInfo {
				param := typeAlias.Params[i].(*ast.TypeParam)

				if param.Name.Lexeme != expectedInfo.name {
					t.Errorf("Parameter %d: expected name %q but got %q", i, expectedInfo.name, param.Name.Lexeme)
				}

				if param.IsStar != expectedInfo.isStar {
					t.Errorf("Parameter %d: expected IsStar=%v but got %v", i, expectedInfo.isStar, param.IsStar)
				}

				if param.IsDoubleStar != expectedInfo.isDoubleStar {
					t.Errorf("Parameter %d: expected IsDoubleStar=%v but got %v", i, expectedInfo.isDoubleStar, param.IsDoubleStar)
				}

				if (param.Bound != nil) != expectedInfo.hasBound {
					t.Errorf("Parameter %d: expected hasBound=%v but got %v", i, expectedInfo.hasBound, param.Bound != nil)
				}

				if (param.Default != nil) != expectedInfo.hasDefault {
					t.Errorf("Parameter %d: expected hasDefault=%v but got %v", i, expectedInfo.hasDefault, param.Default != nil)
				}
			}
		})
	}
}

func TestTypeAliasExpressions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string // "name", "subscript", "binary", etc.
		hasError     bool
	}{
		{
			name:         "simple type name",
			input:        "type MyInt = int",
			expectedType: "name",
		},
		{
			name:         "generic type",
			input:        "type MyList = List[int]",
			expectedType: "subscript",
		},
		{
			name:         "union type",
			input:        "type IntOrStr = int | str",
			expectedType: "binary",
		},
		{
			name:         "callable type",
			input:        "type Handler = Callable[[int], str]",
			expectedType: "subscript",
		},
		{
			name:         "complex union",
			input:        "type Flexible = int | str | List[Any]",
			expectedType: "binary",
		},
		{
			name:         "nested generic",
			input:        "type NestedDict = Dict[str, List[int]]",
			expectedType: "subscript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			stmt, err := parser.statement()

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			typeAlias, ok := stmt.(*ast.TypeAlias)
			if !ok {
				t.Errorf("Expected *ast.TypeAlias but got %T", stmt)
				return
			}

			var actualType string
			switch typeAlias.Value.(type) {
			case *ast.Name:
				actualType = "name"
			case *ast.Subscript:
				actualType = "subscript"
			case *ast.Binary:
				actualType = "binary"
			case *ast.Attribute:
				actualType = "attribute"
			default:
				actualType = "unknown"
			}

			if actualType != tt.expectedType {
				t.Errorf("Expected value type %s but got %s (actual: %T)",
					tt.expectedType, actualType, typeAlias.Value)
			}
		})
	}
}

func TestTypeAliasErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing type name",
			input: "type = int",
		},
		{
			name:  "missing equals",
			input: "type MyType int",
		},
		{
			name:  "missing value",
			input: "type MyType =",
		},
		{
			name:  "unclosed type parameters",
			input: "type Generic[T = int",
		},
		{
			name:  "invalid type parameter",
			input: "type Bad[123] = int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.NewScanner([]byte(tt.input))
			tokens := scanner.ScanTokens()
			parser := NewParser(tokens)
			_, err := parser.statement()

			if err == nil {
				t.Errorf("Expected error but got none for input: %s", tt.input)
			}
		})
	}
}
