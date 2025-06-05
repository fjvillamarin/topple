package transformers

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/codegen"
	"biscuit/compiler/lexer"
	"biscuit/compiler/resolver"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test case structure
type viewTestCase struct {
	name     string
	view     *ast.ViewStmt
	category string
	testFile string
}

// Test runner helper
func runViewTests(t *testing.T, tests []viewTestCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform to Class
			transformer := NewViewTransformer(resolver.NewResolutionTable())
			classNode, err := transformer.TransformViewToClass(tt.view)
			if err != nil {
				t.Fatalf("Transformation failed: %v", err)
			}

			// Generate Python code from the class
			gen := codegen.NewCodeGenerator()
			generated := gen.Generate(classNode)

			// Compare with expected output
			expectedPath := filepath.Join("testdata", "expected", tt.category, tt.testFile+".py")
			compareWithExpected(t, generated, expectedPath, tt.category, tt.testFile)
		})
	}
}

// compareWithExpected compares generated code with expected output
func compareWithExpected(t *testing.T, generated, expectedPath, category, testFile string) {
	// Read expected file
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		if os.IsNotExist(err) && os.Getenv("UPDATE_GOLDEN") == "1" {
			// Create expected file
			err = os.MkdirAll(filepath.Dir(expectedPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create expected file directory: %v", err)
			}
			err = os.WriteFile(expectedPath, []byte(generated), 0644)
			if err != nil {
				t.Fatalf("Failed to write expected file: %v", err)
			}
			t.Logf("Created expected file: %s", expectedPath)
			return
		}
		t.Fatalf("Failed to read expected file: %v", err)
	}

	// Write to generated directory
	generatedPath := filepath.Join("testdata", "generated", category, testFile+".py")
	err = os.MkdirAll(filepath.Dir(generatedPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create generated directory: %v", err)
	}
	err = os.WriteFile(generatedPath, []byte(generated), 0644)
	if err != nil {
		t.Fatalf("Failed to write generated file: %v", err)
	}

	// Normalize and compare
	expectedStr := normalizeWhitespace(string(expected))
	generatedStr := normalizeWhitespace(generated)

	if expectedStr != generatedStr {
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			// Update expected file
			err = os.WriteFile(expectedPath, []byte(generated), 0644)
			if err != nil {
				t.Fatalf("Failed to update expected file: %v", err)
			}
			t.Logf("Updated expected file: %s", expectedPath)
			return
		}

		// Show diff
		t.Errorf("Generated code does not match expected for %s/%s", category, testFile)
		t.Logf("\n=== Expected ===\n%s\n", expectedStr)
		t.Logf("\n=== Generated ===\n%s\n", generatedStr)
		t.Logf("\n=== Diff ===")
		showDiff(t, expectedStr, generatedStr)
	}
}

// normalizeWhitespace normalizes whitespace for comparison
func normalizeWhitespace(s string) string {
	// Trim trailing whitespace from each line and trailing newlines
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// showDiff shows the difference between expected and generated strings
func showDiff(t *testing.T, expected, generated string) {
	expectedLines := strings.Split(expected, "\n")
	generatedLines := strings.Split(generated, "\n")

	maxLines := len(expectedLines)
	if len(generatedLines) > maxLines {
		maxLines = len(generatedLines)
	}

	for i := 0; i < maxLines; i++ {
		var expectedLine, generatedLine string
		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(generatedLines) {
			generatedLine = generatedLines[i]
		}

		if expectedLine != generatedLine {
			t.Logf("Line %d:", i+1)
			t.Logf("  Expected:  %q", expectedLine)
			t.Logf("  Generated: %q", generatedLine)
		}
	}
}

// TestBasicViewTransformation tests basic view to class transformations
func TestBasicViewTransformation(t *testing.T) {
	tests := []viewTestCase{
		{
			name: "simple_view",
			view: ast.HView("HelloWorld", nil,
				ast.HElement("div", "Hello World"),
			),
			category: "basic",
			testFile: "simple_view",
		},
		{
			name: "parametrized_view",
			view: ast.HView("Greeting", []*ast.Parameter{
				ast.HParam("name", "str"),
				ast.HParamWithDefault("title", "str", ast.S("Mr.")),
			},
				ast.HElement("h1", ast.HFStr("Hello, ", ast.N("title"), " ", ast.N("name"))),
			),
			category: "basic",
			testFile: "parametrized_view",
		},
		{
			name: "typed_parameters",
			view: ast.HView("UserCard", []*ast.Parameter{
				ast.HParam("user", "User"),
				ast.HParamWithDefault("show_email", "bool", ast.B(false)),
			},
				ast.HElement("div",
					ast.HAttr("class", ast.S("user-card")),
					ast.HElement("h2", ast.HAttributeAccess(ast.N("user"), "name")),
				),
			),
			category: "basic",
			testFile: "typed_parameters",
		},
	}

	runViewTests(t, tests)
}

// TestHTMLElements tests HTML element transformations
func TestHTMLElements(t *testing.T) {
	tests := []viewTestCase{
		{
			name: "single_element",
			view: ast.HView("SingleElement", nil,
				ast.HElement("p", "Simple paragraph"),
			),
			category: "html",
			testFile: "single_element",
		},
		{
			name: "nested_elements",
			view: ast.HView("NestedElements", nil,
				ast.HElement("div",
					ast.HElement("h1", "Title"),
					ast.HElement("p", "Content"),
				),
			),
			category: "html",
			testFile: "nested_elements",
		},
		{
			name: "element_with_attributes",
			view: ast.HView("ElementWithAttributes", []*ast.Parameter{
				ast.HParam("css_class", "str"),
			},
				ast.HElement("div",
					ast.HAttr("class", ast.N("css_class")),
					ast.HAttr("id", ast.S("main-div")),
					"Content with attributes",
				),
			),
			category: "html",
			testFile: "attributes",
		},
	}

	runViewTests(t, tests)
}

// TestSlotSystem tests slot transformations
func TestSlotSystem(t *testing.T) {
	tests := []viewTestCase{
		{
			name: "default_slot",
			view: ast.HView("Card", nil,
				ast.HElement("div",
					ast.HAttr("class", ast.S("card")),
					ast.HElement("slot"),
				),
			),
			category: "slots",
			testFile: "default_slot",
		},
		{
			name: "named_slots",
			view: ast.HView("Layout", nil,
				ast.HElement("div",
					ast.HElement("header",
						ast.HElement("slot", ast.HAttr("name", ast.S("header"))),
					),
					ast.HElement("main",
						ast.HElement("slot"),
					),
					ast.HElement("footer",
						ast.HElement("slot", ast.HAttr("name", ast.S("footer"))),
					),
				),
			),
			category: "slots",
			testFile: "named_slots",
		},
	}

	runViewTests(t, tests)
}

// TestControlFlow tests control flow transformations in views
func TestControlFlow(t *testing.T) {
	tests := []viewTestCase{
		{
			name: "if_statement",
			view: ast.HView("ConditionalView", []*ast.Parameter{
				ast.HParam("show_message"),
				ast.HParam("message"),
			},
				ast.HIf(ast.N("show_message"),
					[]ast.Stmt{ast.HElement("div", ast.N("message"))},
					ast.HElement("div", "No message"),
				),
			),
			category: "control_flow",
			testFile: "if_statement",
		},
		{
			name: "for_loop",
			view: ast.HView("ItemList", []*ast.Parameter{
				ast.HParam("items"),
			},
				ast.HElement("ul",
					ast.HFor(ast.N("item"), ast.N("items"),
						[]ast.Stmt{ast.HElement("li", ast.N("item"))},
					),
				),
			),
			category: "control_flow",
			testFile: "for_loop",
		},
		{
			name: "while_loop",
			view: ast.HView("CounterView", []*ast.Parameter{
				ast.HParam("count"),
			},
				ast.HElement("div",
					ast.HWhile(
						ast.HBinary(ast.N("count"), lexer.Greater, ">", ast.I(0)),
						[]ast.Stmt{
							ast.HElement("span", ast.N("count")),
							ast.HAssign(
								[]ast.Expr{ast.N("count")},
								ast.HBinary(ast.N("count"), lexer.Minus, "-", ast.I(1)),
							),
						},
					),
				),
			),
			category: "control_flow",
			testFile: "while_loop",
		},
	}

	runViewTests(t, tests)
}

// TestExpressions tests expression transformations
func TestExpressions(t *testing.T) {
	tests := []viewTestCase{
		{
			name: "self_parameter_access",
			view: ast.HView("UserProfile", []*ast.Parameter{
				ast.HParam("username"),
				ast.HParam("email"),
			},
				ast.HElement("div",
					ast.HElement("h1", ast.N("username")),
					ast.HElement("p", ast.HBinary(ast.S("Email: "), lexer.Plus, "+", ast.N("email"))),
				),
			),
			category: "expressions",
			testFile: "self_parameter_access",
		},
	}

	runViewTests(t, tests)
}

// TestEdgeCases tests edge cases and corner scenarios
func TestEdgeCases(t *testing.T) {
	tests := []viewTestCase{
		{
			name:     "empty_view",
			view:     ast.HView("EmptyView", nil),
			category: "edge_cases",
			testFile: "empty_view",
		},
		{
			name: "multiple_roots",
			view: ast.HView("MultipleRoots", nil,
				ast.HElement("h1", "First"),
				ast.HElement("p", "Second"),
				ast.HElement("div", "Third"),
			),
			category: "edge_cases",
			testFile: "multiple_roots",
		},
	}

	runViewTests(t, tests)
}

// TODO: Tests to implement when needed
func TestViewComponents(t *testing.T) {
	t.Skip("View components not implemented yet")
}

func TestHTMLContent(t *testing.T) {
	t.Skip("HTML content tests not implemented yet")
}

func TestAdvancedControlFlow(t *testing.T) {
	t.Skip("Advanced control flow (if/elif/else, try/except, match, with) not implemented yet")
}

func TestAdvancedExpressions(t *testing.T) {
	t.Skip("Advanced expressions (complex expressions, string interpolation, collections, ternary) not implemented yet")
}