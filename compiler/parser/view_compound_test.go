package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"testing"
)

func TestViewIfStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple if in view",
			input: `view test(show: bool):
    if show:
        <div>Visible</div>`,
		},
		{
			name: "if-else in view",
			input: `view test(show: bool):
    if show:
        <div>Visible</div>
    else:
        <div>Hidden</div>`,
		},
		{
			name: "if-elif-else in view",
			input: `view test(status: str):
    if status == "success":
        <div class="success">Success!</div>
    elif status == "error":
        <div class="error">Error!</div>
    else:
        <div class="info">Info</div>`,
		},
		{
			name: "nested if in view",
			input: `view test(x: int, y: int):
    if x > 0:
        if y > 0:
            <div>Both positive</div>
        else:
            <div>X positive, Y not</div>`,
		},
		{
			name: "if with mixed content",
			input: `view test(items: list):
    if items:
        <h1>Items:</h1>
        for item in items:
            <div>{item}</div>
    else:
        <p>No items</p>`,
		},
		{
			name:     "if without colon",
			input:    `view test(): if True <div></div>`,
			hasError: true,
		},
		{
			name:     "if without condition",
			input:    `view test(): if: <div></div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Check that we have statements in the body
			if len(viewStmt.Body) == 0 {
				t.Errorf("Expected view body to have at least one statement")
			}

			// First statement should be an If
			if len(viewStmt.Body) > 0 {
				_, ok := viewStmt.Body[0].(*ast.If)
				if !ok {
					t.Errorf("Expected first body statement to be *ast.If, got %T", viewStmt.Body[0])
				}
			}
		})
	}
}

func TestViewForStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple for loop in view",
			input: `view test(items: list):
    for item in items:
        <div>{item}</div>`,
		},
		{
			name: "for loop with index",
			input: `view test(items: list):
    for i, item in enumerate(items):
        <div>Item {i}: {item}</div>`,
		},
		{
			name: "async for loop in view",
			input: `view test(items: list):
    async for item in items:
        <div>{item}</div>`,
		},
		{
			name: "for loop with else",
			input: `view test(items: list):
    for item in items:
        <div>{item}</div>
    else:
        <div>No items</div>`,
		},
		{
			name: "nested for loops in view",
			input: `view test(matrix: list):
    for row in matrix:
        <div class="row">
            for cell in row:
                <span>{cell}</span>
        </div>`,
		},
		{
			name: "for loop with mixed content",
			input: `view test(users: list):
    <h1>Users:</h1>
    for user in users:
        name = user.name
        <div class="user">{name}</div>
    <footer>End of users</footer>`,
		},
		{
			name:     "for without in",
			input:    `view test(): for item items: <div></div>`,
			hasError: true,
		},
		{
			name:     "for without colon",
			input:    `view test(): for item in items <div></div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Check that we have statements in the body
			if len(viewStmt.Body) == 0 {
				t.Errorf("Expected view body to have at least one statement")
			}

			// Look for For statement in the body
			foundFor := false
			for _, stmt := range viewStmt.Body {
				if _, ok := stmt.(*ast.For); ok {
					foundFor = true
					break
				}
			}

			if !foundFor {
				t.Errorf("Expected to find *ast.For in view body")
			}
		})
	}
}

func TestViewWhileStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple while loop in view",
			input: `view test(count: int):
    while count > 0:
        <div>Count: {count}</div>
        count -= 1`,
		},
		{
			name: "while loop with else",
			input: `view test(items: list):
    while items:
        item = items.pop()
        <div>{item}</div>
    else:
        <div>All done</div>`,
		},
		{
			name: "nested while in view",
			input: `view test(x: int):
    while x > 0:
        y = x
        while y > 0:
            <span>{y}</span>
            y -= 1
        x -= 1`,
		},
		{
			name:     "while without colon",
			input:    `view test(): while True <div></div>`,
			hasError: true,
		},
		{
			name:     "while without condition",
			input:    `view test(): while: <div></div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Look for While statement in the body
			foundWhile := false
			for _, stmt := range viewStmt.Body {
				if _, ok := stmt.(*ast.While); ok {
					foundWhile = true
					break
				}
			}

			if !foundWhile {
				t.Errorf("Expected to find *ast.While in view body")
			}
		})
	}
}

func TestViewTryStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple try-except in view",
			input: `view test(data: dict):
    try:
        value = data["key"]
        <div>{value}</div>
    except KeyError:
        <div>Key not found</div>`,
		},
		{
			name: "try-except-else-finally in view",
			input: `view test(data: dict):
    try:
        value = data["key"]
        <div>{value}</div>
    except KeyError as e:
        <div>Error: {e}</div>
    else:
        <div>Success</div>
    finally:
        <div>Cleanup</div>`,
		},
		{
			name: "try with multiple except blocks",
			input: `view test(data: str):
    try:
        num = int(data)
        <div>Number: {num}</div>
    except ValueError:
        <div>Invalid number</div>
    except TypeError:
        <div>Wrong type</div>
    except:
        <div>Unknown error</div>`,
		},
		{
			name: "try-finally without except",
			input: `view test():
    try:
        <div>Trying something</div>
    finally:
        <div>Always runs</div>`,
		},
		{
			name:     "try without except or finally",
			input:    `view test(): try: <div></div>`,
			hasError: true,
		},
		{
			name:     "try without colon",
			input:    `view test(): try <div></div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Look for Try statement in the body
			foundTry := false
			for _, stmt := range viewStmt.Body {
				if _, ok := stmt.(*ast.Try); ok {
					foundTry = true
					break
				}
			}

			if !foundTry {
				t.Errorf("Expected to find *ast.Try in view body")
			}
		})
	}
}

func TestViewMatchStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple match in view",
			input: `view test(value: int):
    match value:
        case 1:
            <div>One</div>
        case 2:
            <div>Two</div>
        case _:
            <div>Other</div>`,
		},
		{
			name: "match with guards in view",
			input: `view test(x: int):
    match x:
        case n if n > 0:
            <div>Positive: {n}</div>
        case n if n < 0:
            <div>Negative: {n}</div>
        case 0:
            <div>Zero</div>`,
		},
		{
			name: "match with complex patterns",
			input: `view test(data: tuple):
    match data:
        case (x, y) if x > y:
            <div>X is greater</div>
        case (x, y) if x < y:
            <div>Y is greater</div>
        case (x, x):
            <div>Equal values</div>`,
		},
		{
			name:     "match without cases",
			input:    `view test(): match value:`,
			hasError: true,
		},
		{
			name:     "match without colon",
			input:    `view test(): match value case 1: <div></div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Look for MatchStmt in the body
			foundMatch := false
			for _, stmt := range viewStmt.Body {
				if _, ok := stmt.(*ast.MatchStmt); ok {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				t.Errorf("Expected to find *ast.MatchStmt in view body")
			}
		})
	}
}

func TestViewFunctionDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "simple function in view",
			input: `view test():
    def helper():
        return "Hello"
    
    result = helper()
    <div>{result}</div>`,
		},
		{
			name: "function with parameters in view",
			input: `view test():
    def format_name(first: str, last: str) -> str:
        return f"{first} {last}"
    
    name = format_name("John", "Doe")
    <div>{name}</div>`,
		},
		{
			name: "async function in view",
			input: `view test():
    async def fetch_data():
        return await some_api_call()
    
    data = await fetch_data()
    <div>{data}</div>`,
		},
		{
			name: "nested function in view",
			input: `view test():
    def outer():
        def inner():
            return "nested"
        return inner()
    
    result = outer()
    <div>{result}</div>`,
		},
		{
			name:     "function without colon",
			input:    `view test(): def helper() return "test"`,
			hasError: true,
		},
		{
			name:     "function without name",
			input:    `view test(): def (): pass`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Look for Function statement in the body
			foundFunction := false
			for _, stmt := range viewStmt.Body {
				if _, ok := stmt.(*ast.Function); ok {
					foundFunction = true
					break
				}
			}

			if !foundFunction {
				t.Errorf("Expected to find *ast.Function in view body")
			}
		})
	}
}

func TestViewMixedCompoundStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name: "mixed control flow in view",
			input: `view dashboard(user: User):
    if user.is_admin:
        <h1>Admin Dashboard</h1>
        
        for section in admin_sections:
            <div class="section">
                <h2>{section.title}</h2>
                
                try:
                    data = section.get_data()
                    if data:
                        for item in data:
                            <div class="item">{item}</div>
                    else:
                        <p>No data available</p>
                except Exception as e:
                    <div class="error">Error: {e}</div>
            </div>
    else:
        <h1>User Dashboard</h1>
        <p>Welcome, {user.name}!</p>`,
		},
		{
			name: "complex nested view logic",
			input: `view report(data: dict):
    def format_value(val):
        if isinstance(val, float):
            return f"{val:.2f}"
        return str(val)
    
    <div class="report">
        <h1>Data Report</h1>
        
        for category, items in data.items():
            <section class="category">
                <h2>{category}</h2>
                
                if items:
                    <table>
                        for item in items:
                            <tr>
                                for key, value in item.items():
                                    formatted = format_value(value)
                                    <td>{formatted}</td>
                            </tr>
                    </table>
                else:
                    <p>No items in this category</p>
            </section>
    </div>`,
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

			viewStmt, ok := stmt.(*ast.ViewStmt)
			if !ok {
				t.Errorf("Expected *ast.ViewStmt but got %T", stmt)
				return
			}

			// Basic validation - should have multiple statements
			if len(viewStmt.Body) == 0 {
				t.Errorf("Expected view body to have statements")
			}
		})
	}
}