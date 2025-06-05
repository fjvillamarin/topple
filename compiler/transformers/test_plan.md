# View Transformer Golden File Test Enhancement Plan

## Current Coverage
- **basic/** - Simple views with parameters
- **html/** - HTML element handling  
- **slots/** - Slot functionality

## Proposed Additional Test Categories

### 1. **control_flow/** - All control flow transformations
- `if_statement.py` - If/else in views
- `if_elif_else.py` - Multiple conditions
- `for_loop.py` - Basic for loops with HTML generation
- `for_loop_nested.py` - Nested loops
- `while_loop.py` - While loops in views
- `try_except.py` - Error handling in views
- `match_statement.py` - Pattern matching
- `with_statement.py` - Context managers

### 2. **expressions/** - Expression transformations
- `self_parameter_access.py` - View parameters become self.param
- `complex_expressions.py` - Binary, unary, calls, attributes
- `string_interpolation.py` - F-strings and string concatenation
- `collections.py` - Lists, dicts, sets in views
- `ternary_conditional.py` - Conditional expressions
- `lambda_expressions.py` - Lambda usage

### 3. **view_components/** - View composition
- `simple_component.py` - Using one view inside another
- `component_with_props.py` - Passing parameters to view components
- `component_with_slots.py` - View components with slot content
- `nested_components.py` - Multiple levels of view composition

### 4. **html_content/** - Content handling
- `text_content.py` - Plain text in elements
- `escaped_content.py` - HTML escaping with escape()
- `mixed_content.py` - Text and elements mixed
- `dynamic_content.py` - Expressions as content
- `whitespace_handling.py` - How whitespace is preserved/trimmed

### 5. **edge_cases/** - Corner cases and error conditions
- `empty_view.py` - View with no content (returns fragment)
- `multiple_roots.py` - Multiple top-level elements
- `no_parameters.py` - View without parameters
- `deeply_nested.py` - Very deep nesting
- `complex_attributes.py` - Dynamic and conditional attributes

### 6. **integration/** - Real-world patterns
- `todo_list.py` - A todo list component
- `form_with_validation.py` - Form handling
- `layout_with_slots.py` - Layout component pattern
- `conditional_rendering.py` - Show/hide logic
- `list_rendering.py` - Rendering lists of items

## Test File Structure

For each test case:
1. **AST Builder** in `view_ast_builders.go`
2. **Expected output** in `testdata/expected/category/test_name.py`
3. **Test registration** in `view_test.go`

## Benefits of This Structure

1. **Comprehensive Coverage** - Every code path in our refactored modules gets tested
2. **Easy Debugging** - When something breaks, we know exactly which feature is affected
3. **Documentation** - The tests serve as examples of how to use each feature
4. **Regression Prevention** - Any changes that break existing behavior are caught
5. **Visual Verification** - Easy to see the transformation by comparing input AST with output Python

## Implementation Priority

1. **High Priority** (Core functionality)
   - control_flow (especially if/for)
   - expressions (self parameter transformation)
   - view_components (basic composition)

2. **Medium Priority** (Common patterns)
   - html_content variations
   - more complex control flow (try/match)
   - component slots

3. **Low Priority** (Edge cases)
   - edge_cases
   - integration examples

## Example Test Case: If Statement

```go
// In view_ast_builders.go
"control_flow/if_statement": func() ast.Node {
    return &ast.ViewStmt{
        Name: &ast.Name{Token: lexer.Token{Lexeme: "ConditionalView"}},
        Params: &ast.ParameterList{
            Parameters: []*ast.Parameter{
                {Name: &ast.Name{Token: lexer.Token{Lexeme: "show_message"}}},
                {Name: &ast.Name{Token: lexer.Token{Lexeme: "message"}}},
            },
        },
        Body: []ast.Stmt{
            &ast.If{
                Condition: &ast.Name{Token: lexer.Token{Lexeme: "show_message"}},
                Body: []ast.Stmt{
                    &ast.HTMLElement{
                        TagName: &ast.Name{Token: lexer.Token{Lexeme: "div"}},
                        Content: []ast.Stmt{
                            &ast.ExprStmt{
                                Expr: &ast.Name{Token: lexer.Token{Lexeme: "message"}},
                            },
                        },
                    },
                },
                Else: []ast.Stmt{
                    &ast.HTMLElement{
                        TagName: &ast.Name{Token: lexer.Token{Lexeme: "div"}},
                        Content: []ast.Stmt{
                            &ast.ExprStmt{
                                Expr: &ast.Literal{Value: "No message"},
                            },
                        },
                    },
                },
            },
        },
    }
},
```

Expected output:
```python
class ConditionalView(BaseView):
    def __init__(self, show_message, message):
        super().__init__()
        self.show_message = show_message
        self.message = message
    
    def _render(self) -> Element:
        if self.show_message:
            return el("div", escape(self.message))
        else:
            return el("div", "No message")
```