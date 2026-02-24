package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
	"testing"
)

// Helper function to parse a view statement
func parseViewStatement(t *testing.T, input string) (*ast.ViewStmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()
	parser := NewParser(tokens)
	stmt, err := parser.statement()
	if err != nil {
		return nil, err
	}
	viewStmt, ok := stmt.(*ast.ViewStmt)
	if !ok {
		return nil, nil
	}
	return viewStmt, nil
}

// Helper function to validate view statement structure
func validateViewStatement(t *testing.T, viewStmt *ast.ViewStmt, expectedName string, expectedParamCount int, expectedBodyCount int, description string) {
	if viewStmt == nil {
		t.Fatalf("Expected ViewStmt but got nil for %s", description)
	}

	if viewStmt.Name == nil {
		t.Errorf("View name should not be nil for %s", description)
	} else if expectedName != "" && viewStmt.Name.Token.Lexeme != expectedName {
		t.Errorf("Expected view name %s but got %s for %s", expectedName, viewStmt.Name.Token.Lexeme, description)
	}

	if viewStmt.Params == nil {
		t.Errorf("View params should not be nil for %s", description)
	} else if expectedParamCount >= 0 && len(viewStmt.Params.Parameters) != expectedParamCount {
		t.Errorf("Expected %d parameters but got %d for %s", expectedParamCount, len(viewStmt.Params.Parameters), description)
	}

	if expectedBodyCount >= 0 && len(viewStmt.Body) != expectedBodyCount {
		t.Errorf("Expected %d body statements but got %d for %s", expectedBodyCount, len(viewStmt.Body), description)
	}

	if viewStmt.Kind != ast.ViewKindServerView {
		t.Errorf("Expected ViewKindServerView but got %v for %s", viewStmt.Kind, description)
	}
}

// Helper function to validate view parsing success
func validateViewParseSuccess(t *testing.T, viewStmt *ast.ViewStmt, err error, description string) {
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %v", description, err)
	}

	if viewStmt == nil {
		t.Fatalf("Expected ViewStmt but got nil for %s", description)
	}
}

// Helper function to validate view parsing error
func validateViewParseError(t *testing.T, viewStmt *ast.ViewStmt, err error, expectedErrorText string, description string) {
	if err == nil {
		t.Errorf("Expected error for %s, but got none", description)
		return
	}

	if expectedErrorText != "" && !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q for %s",
			expectedErrorText, err.Error(), description)
	}

	if viewStmt != nil {
		t.Errorf("Expected nil ViewStmt on error, but got %T for %s", viewStmt, description)
	}
}

// Helper function to check for specific statement types in view body
func findStatementInViewBody(viewStmt *ast.ViewStmt, stmtType string) bool {
	return findStatementInList(viewStmt.Body, stmtType)
}

// Helper function to recursively search for statement types in a list of statements
func findStatementInList(stmts []ast.Stmt, stmtType string) bool {
	for _, stmt := range stmts {
		// Check direct match
		switch stmtType {
		case "HTMLElement":
			if _, ok := stmt.(*ast.HTMLElement); ok {
				return true
			}
		case "If":
			if _, ok := stmt.(*ast.If); ok {
				return true
			}
		case "For":
			if _, ok := stmt.(*ast.For); ok {
				return true
			}
		case "Assignment":
			if _, ok := stmt.(*ast.AssignStmt); ok {
				return true
			}
		case "Expression":
			if _, ok := stmt.(*ast.ExprStmt); ok {
				return true
			}
		}

		// Check inside MultiStmt containers
		if multiStmt, ok := stmt.(*ast.MultiStmt); ok {
			if findStatementInList(multiStmt.Stmts, stmtType) {
				return true
			}
		}
	}
	return false
}

// Helper function to validate HTML element in view
func validateHTMLElementInView(t *testing.T, viewStmt *ast.ViewStmt, expectedTag string, description string) {
	if !findStatementInViewBody(viewStmt, "HTMLElement") {
		t.Errorf("Expected to find HTMLElement in view body for %s", description)
		return
	}

	for _, stmt := range viewStmt.Body {
		if htmlElem, ok := stmt.(*ast.HTMLElement); ok {
			if expectedTag != "" && htmlElem.TagName.Lexeme != expectedTag {
				t.Errorf("Expected HTML tag %s but got %s for %s", expectedTag, htmlElem.TagName.Lexeme, description)
			}
			return
		}
	}
}

// Test comprehensive view statement parsing functionality
func TestViewStatement(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		hasError           bool
		errorText          string
		expectedName       string
		expectedParamCount int
		expectedBodyCount  int
		description        string
	}{
		{
			name: "simple view without parameters",
			input: `view hello():
    <div>Hello</div>`,
			expectedName:       "hello",
			expectedParamCount: 0,
			expectedBodyCount:  1,
			description:        "view without parameters containing single HTML element",
		},
		{
			name: "view with parameters",
			input: `view greet(name: str):
    <div>Hello {name}</div>`,
			expectedName:       "greet",
			expectedParamCount: 1,
			expectedBodyCount:  1,
			description:        "view with typed parameter containing HTML with interpolation",
		},
		{
			name: "view with return type",
			input: `view mycomponent() -> str:
    <div>Content</div>`,
			expectedName:       "mycomponent",
			expectedParamCount: 0,
			expectedBodyCount:  1,
			description:        "view with return type annotation",
		},
		{
			name: "multiline view",
			input: `view layout():
    <div>
        <h1>Title</h1>
        <p>Content</p>
    </div>`,
			expectedName:       "layout",
			expectedParamCount: 0,
			expectedBodyCount:  1,
			description:        "view with multiline HTML content",
		},
		{
			name: "view with Python statements",
			input: `view conditional(show: bool):
    if show:
        <div>Visible</div>`,
			expectedName:       "conditional",
			expectedParamCount: 1,
			expectedBodyCount:  1,
			description:        "view containing Python control flow statements",
		},
		{
			name:        "view without body",
			input:       "view empty():",
			hasError:    true,
			errorText:   "expected",
			description: "view missing body should fail",
		},
		{
			name:        "view without colon",
			input:       "view invalid() <div></div>",
			hasError:    true,
			errorText:   "expected",
			description: "view missing colon should fail",
		},
		{
			name:        "view without closing parenthesis",
			input:       "view broken(name: str: <div></div>",
			hasError:    true,
			errorText:   "expected",
			description: "view with malformed parameters should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewStmt, err := parseViewStatement(t, tt.input)

			if tt.hasError {
				validateViewParseError(t, viewStmt, err, tt.errorText, tt.description)
			} else {
				validateViewParseSuccess(t, viewStmt, err, tt.description)
				validateViewStatement(t, viewStmt, tt.expectedName, tt.expectedParamCount, tt.expectedBodyCount, tt.description)
			}
		})
	}
}

// Test view statements containing HTML elements
func TestViewWithHTMLElements(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		expectedTag string
		description string
	}{
		{
			name: "self-closing tag",
			input: `view test():
    <br/>`,
			expectedTag: "br",
			description: "view with self-closing HTML tag",
		},
		{
			name: "empty tag",
			input: `view test():
    <div></div>`,
			expectedTag: "div",
			description: "view with empty HTML tag",
		},
		{
			name: "tag with text content",
			input: `view test():
    <p>Hello world</p>`,
			expectedTag: "p",
			description: "view with HTML tag containing text content",
		},
		{
			name: "tag with interpolation",
			input: `view test(name: str):
    <span>{name}</span>`,
			expectedTag: "span",
			description: "view with HTML tag containing variable interpolation",
		},
		{
			name: "tag with attributes",
			input: `view test():
    <div class="container" id="main"></div>`,
			expectedTag: "div",
			description: "view with HTML tag containing multiple attributes",
		},
		{
			name: "tag with boolean attribute",
			input: `view test():
    <input disabled checked/>`,
			expectedTag: "input",
			description: "view with HTML tag containing boolean attributes",
		},
		{
			name: "tag with expression attribute",
			input: `view test(variant: str):
    <div class={variant}></div>`,
			expectedTag: "div",
			description: "view with HTML tag containing expression-based attributes",
		},
		{
			name: "nested elements",
			input: `view test():
    <div><span>nested</span></div>`,
			expectedTag: "div",
			description: "view with nested HTML elements",
		},
		{
			name: "unclosed tag",
			input: `view test():
    <div>content`,
			hasError:    true,
			errorText:   "expected",
			description: "view with unclosed HTML tag should fail",
		},
		{
			name: "mismatched closing tag",
			input: `view test():
    <div></span>`,
			hasError:    true,
			errorText:   "closing tag name doesn't match",
			description: "view with mismatched HTML tag should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewStmt, err := parseViewStatement(t, tt.input)

			if tt.hasError {
				validateViewParseError(t, viewStmt, err, tt.errorText, tt.description)
			} else {
				validateViewParseSuccess(t, viewStmt, err, tt.description)
				validateHTMLElementInView(t, viewStmt, tt.expectedTag, tt.description)
			}
		})
	}
}

// Test view statements containing Python code
func TestViewWithPythonStatements(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		hasError         bool
		errorText        string
		expectedStmtType string
		description      string
	}{
		{
			name: "python assignment",
			input: `view test():
    x = 5
    <div>{x}</div>`,
			expectedStmtType: "Assignment",
			description:      "view with Python assignment statement",
		},
		{
			name: "python if statement",
			input: `view test(show: bool):
    if show:
        <div>Visible</div>`,
			expectedStmtType: "If",
			description:      "view with Python conditional statement",
		},
		{
			name: "python for loop",
			input: `view test(items: list):
    for item in items:
        <div>{item}</div>`,
			expectedStmtType: "For",
			description:      "view with Python for loop statement",
		},
		{
			name: "mixed statements",
			input: `view test():
    <h1>Title</h1>
    name = "World"
    <p>Hello {name}</p>`,
			expectedStmtType: "HTMLElement", // First statement is HTML
			description:      "view with mixed HTML and Python statements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewStmt, err := parseViewStatement(t, tt.input)

			if tt.hasError {
				validateViewParseError(t, viewStmt, err, tt.errorText, tt.description)
			} else {
				validateViewParseSuccess(t, viewStmt, err, tt.description)

				if len(viewStmt.Body) == 0 {
					t.Errorf("Expected view body to have at least one statement for %s", tt.description)
					return
				}

				// Validate the expected statement type is present
				if tt.expectedStmtType != "" && !findStatementInViewBody(viewStmt, tt.expectedStmtType) {
					t.Errorf("Expected to find %s statement in view body for %s", tt.expectedStmtType, tt.description)
				}
			}
		})
	}
}

// Test view statements with complex parameter patterns
func TestViewParameterPatterns(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		hasError           bool
		errorText          string
		expectedParamCount int
		description        string
	}{
		{
			name: "view with typed parameters",
			input: `view test_view(title: str, count: int, active: bool):
    <div class={"active" if active else ""}>
        <h1>{title}</h1>
        <span>{count}</span>
    </div>`,
			expectedParamCount: 3,
			description:        "view with multiple typed parameters",
		},
		{
			name: "view with default parameters",
			input: `view button(text: str = "Click me", variant: str = "primary"):
    <button class={variant}>{text}</button>`,
			expectedParamCount: 2,
			description:        "view with default parameter values",
		},
		{
			name: "view with complex type annotations",
			input: `view list_view(items: list, formatter: str):
    for item in items:
        <div>{item}</div>`,
			expectedParamCount: 2,
			description:        "view with complex type annotations",
		},
		{
			name: "view with invalid parameter syntax",
			input: `view broken(123invalid: str):
    <div></div>`,
			hasError:    true,
			errorText:   "expected",
			description: "view with invalid parameter identifier should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewStmt, err := parseViewStatement(t, tt.input)

			if tt.hasError {
				validateViewParseError(t, viewStmt, err, tt.errorText, tt.description)
			} else {
				validateViewParseSuccess(t, viewStmt, err, tt.description)
				validateViewStatement(t, viewStmt, "", tt.expectedParamCount, -1, tt.description)
			}
		})
	}
}

// Test view statements with edge cases and error conditions
func TestViewEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		errorText   string
		description string
	}{
		{
			name: "view with complex nested content",
			input: `view dashboard(user: User, sections: list[Section]):
    <div class="dashboard">
        <header>
            <h1>Welcome, {user.name}</h1>
            if user.is_admin:
                <span class="admin-badge">Admin</span>
        </header>
        
        <main>
            for section in sections:
                <section class="dashboard-section">
                    <h2>{section.title}</h2>
                    
                    try:
                        content = section.render()
                        <div class="content">{content}</div>
                    except RenderError as e:
                        <div class="error">Failed to render: {e}</div>
                </section>
        </main>
    </div>`,
			description: "view with deeply nested HTML and Python control flow",
		},
		{
			name: "view with docstring",
			input: `view documented():
    """
    This is a documented view component.
    It demonstrates docstring support.
    """
    <div>Documented content</div>`,
			description: "view with docstring documentation",
		},
		{
			name: "view with missing view keyword",
			input: `component():
    <div>Missing view keyword</div>`,
			hasError:    true,
			errorText:   "expected",
			description: "statement without view keyword should fail",
		},
		{
			name: "view with empty name",
			input: `view ():
    <div>Empty name</div>`,
			hasError:    true,
			errorText:   "expected",
			description: "view with empty name should fail",
		},
		{
			name: "view with reserved keyword name",
			input: `view def():
    <div>Reserved keyword</div>`,
			hasError:    true,
			errorText:   "expected",
			description: "view with reserved keyword as name should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewStmt, err := parseViewStatement(t, tt.input)

			if tt.hasError {
				validateViewParseError(t, viewStmt, err, tt.errorText, tt.description)
			} else {
				validateViewParseSuccess(t, viewStmt, err, tt.description)

				// For successful complex cases, just verify basic structure
				if len(viewStmt.Body) == 0 {
					t.Errorf("Expected view body to have statements for %s", tt.description)
				}
			}
		})
	}
}
