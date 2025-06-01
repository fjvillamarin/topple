package transformers

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"biscuit/compiler/resolver"
	"fmt"
	"math/rand"
	"time"
)

type ViewTransformer struct {
	// Track if we need to add psx_runtime imports
	needsRuntimeImports bool

	// Resolution table for parameter transformation
	resolutionTable *resolver.ResolutionTable

	// Context tracking for hierarchical HTML generation
	contextStack   []string // Stack of current children array names
	currentContext string   // Current children array name
	nextContextId  int      // Counter for generating unique context names
}

// HTMLContext represents a context for collecting HTML children
type HTMLContext struct {
	ChildrenVarName string     // Name of the children array variable
	Children        []ast.Stmt // Statements that build the children
}

// NewViewTransformer creates a new ViewTransformer with the given resolution table
func NewViewTransformer(resolutionTable *resolver.ResolutionTable) *ViewTransformer {
	// Initialize random seed for context ID generation
	rand.Seed(time.Now().UnixNano())

	return &ViewTransformer{
		needsRuntimeImports: false,
		resolutionTable:     resolutionTable,
		contextStack:        []string{},
		currentContext:      "",
		nextContextId:       1000,
	}
}

// generateContextName generates a unique name for a children array
func (vm *ViewTransformer) generateContextName(prefix string) string {
	name := fmt.Sprintf("_%s_children_%d", prefix, vm.nextContextId)
	vm.nextContextId += rand.Intn(9000) + 1000 // Generate random IDs to avoid conflicts
	return name
}

// pushContext creates a new HTML context and pushes it onto the stack
func (vm *ViewTransformer) pushContext(prefix string) string {
	contextName := vm.generateContextName(prefix)
	vm.contextStack = append(vm.contextStack, vm.currentContext)
	vm.currentContext = contextName
	return contextName
}

// popContext restores the previous HTML context
func (vm *ViewTransformer) popContext() string {
	if len(vm.contextStack) > 0 {
		vm.currentContext = vm.contextStack[len(vm.contextStack)-1]
		vm.contextStack = vm.contextStack[:len(vm.contextStack)-1]
	} else {
		vm.currentContext = ""
	}
	return vm.currentContext
}

// TransformViewToClass transforms a ViewStmt into a Class that inherits from BaseView
func (vm *ViewTransformer) TransformViewToClass(viewStmt *ast.ViewStmt) (*ast.Class, error) {
	// Resolution table is already stored during construction

	// Create the class name (same as view name)
	className := viewStmt.Name

	// Create base class reference: BaseView
	baseViewName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "BaseView",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create __init__ method
	initMethod, err := vm.createInitMethod(viewStmt)
	if err != nil {
		return nil, err
	}

	// Create _render method
	renderMethod, err := vm.createRenderMethod(viewStmt)
	if err != nil {
		return nil, err
	}

	// Create the class body with both methods
	classBody := []ast.Stmt{initMethod, renderMethod}

	// Convert TypeParams from []*TypeParam to []TypeParam
	var typeParams []ast.TypeParam
	for _, tp := range viewStmt.TypeParams {
		typeParams = append(typeParams, *tp)
	}

	// Create base class argument
	baseViewArg := &ast.Argument{
		Value: baseViewName,
		Span:  viewStmt.Span,
	}

	// Create the class node
	classNode := &ast.Class{
		Name:       className,
		TypeParams: typeParams,
		Args:       []*ast.Argument{baseViewArg}, // inherits from BaseView
		Body:       classBody,
		Span:       viewStmt.Span,
	}

	vm.needsRuntimeImports = true
	return classNode, nil
}

// isViewParameter checks if a name is a view parameter using the resolution table
func (vm *ViewTransformer) isViewParameter(name *ast.Name) bool {
	if vm.resolutionTable == nil {
		return false
	}

	// Look up the variable in the resolution table
	if variable, exists := vm.resolutionTable.Variables[name]; exists {
		return variable.IsViewParameter
	}

	return false
}

// transformNameToSelfAttribute transforms a view parameter name to self.param
func (vm *ViewTransformer) transformNameToSelfAttribute(name *ast.Name) *ast.Attribute {
	return &ast.Attribute{
		Object: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: name.Span,
		},
		Name: name.Token,
		Span: name.Span,
	}
}

// createInitMethod creates the __init__ method for the view class
func (vm *ViewTransformer) createInitMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create __init__ method name
	initName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "__init__",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Span,
	}

	// Start with self parameter
	initParams := []*ast.Parameter{selfParam}

	// Add view parameters to __init__
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}
			initParams = append(initParams, param)
		}
	}

	// Create parameter list
	paramList := &ast.ParameterList{
		Parameters:  initParams,
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Span,
	}

	// Create assignment statements for each view parameter
	var initBody []ast.Stmt
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}

			// Create self.param_name = param_name
			selfAttr := &ast.Attribute{
				Object: &ast.Name{
					Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
					Span:  viewStmt.Span,
				},
				Name: param.Name.Token,
				Span: viewStmt.Span,
			}

			assignment := &ast.AssignStmt{
				Targets: []ast.Expr{selfAttr},
				Value:   param.Name,
				Span:    viewStmt.Span,
			}

			initBody = append(initBody, assignment)
		}
	}

	// If no body, add pass statement
	if len(initBody) == 0 {
		passStmt := &ast.PassStmt{
			Span: viewStmt.Span,
		}
		initBody = append(initBody, passStmt)
	}

	return &ast.Function{
		Name:           initName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     nil,
		Body:           initBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}

// createRenderMethod creates the _render method for the view class
func (vm *ViewTransformer) createRenderMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create _render method name
	renderName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "_render",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Params.Span,
	}

	// Create parameter list with just self
	paramList := &ast.ParameterList{
		Parameters:  []*ast.Parameter{selfParam},
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Params.Span,
	}

	// Create return type annotation: Element
	elementType := &ast.Name{
		Token: lexer.Token{
			Lexeme: "Element",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Transform view body into _render method body
	renderBody, err := vm.transformViewBody(viewStmt.Body)
	if err != nil {
		return nil, err
	}

	return &ast.Function{
		Name:           renderName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     elementType,
		Body:           renderBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}

// transformViewBody transforms the view body statements into _render method statements
func (vm *ViewTransformer) transformViewBody(body []ast.Stmt) ([]ast.Stmt, error) {
	var transformedBody []ast.Stmt

	// Handle empty body
	if len(body) == 0 {
		// Return empty fragment for empty views
		returnValue := &ast.Call{
			Callee: &ast.Name{
				Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
				Span:  lexer.Span{},
			},
			Arguments: []*ast.Argument{{
				Value: &ast.ListExpr{Elements: []ast.Expr{}, Span: lexer.Span{}},
				Span:  lexer.Span{},
			}},
			Span: lexer.Span{},
		}

		returnStmt := &ast.ReturnStmt{
			Value: returnValue,
			Span:  lexer.Span{},
		}

		return []ast.Stmt{returnStmt}, nil
	}

	// Check if we have only one root element (regardless of its internal complexity)
	if len(body) == 1 {
		if htmlElement, ok := body[0].(*ast.HTMLElement); ok {
			// Single root HTML element - process it directly without view children array
			// The processHTMLElement method will handle returning it directly when no parent context
			processedStmts, err := vm.processViewStatement(htmlElement)
			if err != nil {
				return nil, err
			}
			return processedStmts, nil
		}
	}

	// Multiple elements or non-HTML root elements - use fragment approach
	viewChildrenName := vm.pushContext("view")
	defer vm.popContext()

	// Create the children array initialization
	childrenArray := &ast.AssignStmt{
		Targets: []ast.Expr{&ast.Name{
			Token: lexer.Token{Lexeme: viewChildrenName, Type: lexer.Identifier},
			Span:  lexer.Span{},
		}},
		Value: &ast.ListExpr{Elements: []ast.Expr{}, Span: lexer.Span{}},
		Span:  lexer.Span{},
	}
	transformedBody = append(transformedBody, childrenArray)

	// Process each statement in the view body
	for _, stmt := range body {
		if stmt == nil {
			continue
		}

		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Create the return statement using fragment
	returnValue := &ast.Call{
		Callee: &ast.Name{
			Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
			Span:  lexer.Span{},
		},
		Arguments: []*ast.Argument{{
			Value: &ast.Name{
				Token: lexer.Token{Lexeme: viewChildrenName, Type: lexer.Identifier},
				Span:  lexer.Span{},
			},
			Span: lexer.Span{},
		}},
		Span: lexer.Span{},
	}

	returnStmt := &ast.ReturnStmt{
		Value: returnValue,
		Span:  body[0].GetSpan(),
	}

	transformedBody = append(transformedBody, returnStmt)
	return transformedBody, nil
}

// processViewStatement processes a single statement in the view body
func (vm *ViewTransformer) processViewStatement(stmt ast.Stmt) ([]ast.Stmt, error) {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		// Transform HTML element and append to current context
		return vm.processHTMLElement(s)

	case *ast.HTMLContent:
		// Transform HTML content (text + interpolations) and append to current context
		return vm.processHTMLContent(s)

	case *ast.For:
		// Handle for loops in HTML context
		return vm.processForLoop(s)

	case *ast.If:
		// Handle if statements in HTML context
		return vm.processIfStatement(s)

	default:
		// For non-HTML statements, transform them and keep as Python code
		transformedStmt := vm.transformStatement(stmt)
		return []ast.Stmt{transformedStmt}, nil
	}
}

// needsHierarchicalProcessing analyzes content to determine if it requires children arrays
func (vm *ViewTransformer) needsHierarchicalProcessing(content []ast.Stmt) bool {
	for _, stmt := range content {
		switch stmt.(type) {
		case *ast.For, *ast.If, *ast.While, *ast.With, *ast.Try:
			// Compound statements that can generate multiple elements need hierarchical processing
			return true
		case *ast.HTMLElement:
			// Nested HTML elements might need hierarchical processing
			if htmlElement, ok := stmt.(*ast.HTMLElement); ok {
				if vm.needsHierarchicalProcessing(htmlElement.Content) {
					return true
				}
			}
		}
	}
	return false
}

// processHTMLElement processes an HTMLElement and returns the transformed statements
func (vm *ViewTransformer) processHTMLElement(element *ast.HTMLElement) ([]ast.Stmt, error) {
	var statements []ast.Stmt

	// Extract the actual tag name
	tagName := element.TagName.Lexeme

	// Store the parent context before potentially creating a new one
	parentContext := vm.currentContext
	var elCall ast.Expr

	// Check if this element needs hierarchical processing
	if len(element.Content) > 0 && vm.needsHierarchicalProcessing(element.Content) {
		// Complex content with compound statements - use children array approach
		elementChildrenName := vm.pushContext(tagName)

		// Initialize the children array
		childrenArray := &ast.AssignStmt{
			Targets: []ast.Expr{&ast.Name{
				Token: lexer.Token{Lexeme: elementChildrenName, Type: lexer.Identifier},
				Span:  lexer.Span{},
			}},
			Value: &ast.ListExpr{Elements: []ast.Expr{}, Span: lexer.Span{}},
			Span:  lexer.Span{},
		}
		statements = append(statements, childrenArray)

		// Process all content items, appending to this element's children array
		for _, contentItem := range element.Content {
			processedStmts, err := vm.processViewStatement(contentItem)
			if err != nil {
				return nil, err
			}
			statements = append(statements, processedStmts...)
		}

		// Restore parent context
		vm.popContext()

		// Create the element with the children array
		elCall = vm.createElCall(tagName, &ast.Name{
			Token: lexer.Token{Lexeme: elementChildrenName, Type: lexer.Identifier},
			Span:  lexer.Span{},
		}, nil)

	} else if len(element.Content) > 0 {
		// Simple content - transform directly without children arrays
		contentExpr, err := vm.transformHTMLContent(element.Content)
		if err != nil {
			return nil, err
		}

		elCall = vm.createElCall(tagName, contentExpr, nil)

	} else {
		// Empty element - create directly
		elCall = vm.createElCall(tagName, &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil)
	}

	// Handle the element based on whether we have a parent context
	if parentContext != "" {
		// Append to parent context
		appendStmt := vm.createAppendStatement(parentContext, elCall)
		statements = append(statements, appendStmt)
	} else {
		// No parent context - this is a root element, return it directly
		returnStmt := &ast.ReturnStmt{
			Value: elCall,
			Span:  element.Span,
		}
		statements = append(statements, returnStmt)
	}

	return statements, nil
}

// processForLoop processes a for loop in the context of an HTML context
func (vm *ViewTransformer) processForLoop(loop *ast.For) ([]ast.Stmt, error) {
	// Transform the iterable and target
	transformedIterable := vm.transformExpression(loop.Iterable)
	transformedTarget := vm.transformExpression(loop.Target)

	// Process the loop body
	var transformedBody []ast.Stmt
	for _, stmt := range loop.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Create the transformed for loop
	transformedLoop := &ast.For{
		Target:   transformedTarget,
		Iterable: transformedIterable,
		Body:     transformedBody,
		Else:     []ast.Stmt{}, // TODO: Handle else clause if needed
		IsAsync:  loop.IsAsync,
		Span:     loop.Span,
	}

	return []ast.Stmt{transformedLoop}, nil
}

// processIfStatement processes an if statement in the context of an HTML context
func (vm *ViewTransformer) processIfStatement(ifStmt *ast.If) ([]ast.Stmt, error) {
	// Transform the condition
	transformedCondition := vm.transformExpression(ifStmt.Condition)

	// Process the if body
	var transformedBody []ast.Stmt
	for _, stmt := range ifStmt.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Process the else body
	var transformedElse []ast.Stmt
	for _, stmt := range ifStmt.Else {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedElse = append(transformedElse, processedStmts...)
	}

	// Create the transformed if statement
	transformedIf := &ast.If{
		Condition: transformedCondition,
		Body:      transformedBody,
		Else:      transformedElse,
		Span:      ifStmt.Span,
	}

	return []ast.Stmt{transformedIf}, nil
}

// createAppendStatement creates a statement that appends an element to a children array
func (vm *ViewTransformer) createAppendStatement(arrayName string, element ast.Expr) ast.Stmt {
	// Create: arrayName.append(element)
	appendCall := &ast.Call{
		Callee: &ast.Attribute{
			Object: &ast.Name{
				Token: lexer.Token{Lexeme: arrayName, Type: lexer.Identifier},
				Span:  lexer.Span{},
			},
			Name: lexer.Token{Lexeme: "append", Type: lexer.Identifier},
			Span: lexer.Span{},
		},
		Arguments: []*ast.Argument{{
			Value: element,
			Span:  lexer.Span{},
		}},
		Span: lexer.Span{},
	}

	return &ast.ExprStmt{
		Expr: appendCall,
		Span: lexer.Span{},
	}
}

// transformHTMLElement transforms an HTMLElement into an el() call
func (vm *ViewTransformer) transformHTMLElement(element *ast.HTMLElement) (ast.Expr, error) {
	// Extract the actual tag name
	tagName := element.TagName.Lexeme

	// Transform the content
	contentExpr, err := vm.transformHTMLContent(element.Content)
	if err != nil {
		return nil, err
	}

	// TODO: Handle attributes when needed
	// For now, pass nil for attributes

	return vm.createElCall(tagName, contentExpr, nil), nil
}

// transformHTMLContent transforms HTML content (nested elements, text, etc.) into appropriate expressions
func (vm *ViewTransformer) transformHTMLContent(content []ast.Stmt) (ast.Expr, error) {
	if len(content) == 0 {
		// Empty content
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil
	}

	if len(content) == 1 {
		// Single content item
		return vm.transformHTMLContentItem(content[0])
	}

	// Multiple content items - create a list
	var contentExprs []ast.Expr
	for _, item := range content {
		expr, err := vm.transformHTMLContentItem(item)
		if err != nil {
			return nil, err
		}
		contentExprs = append(contentExprs, expr)
	}

	return &ast.ListExpr{
		Elements: contentExprs,
		Span:     lexer.Span{},
	}, nil
}

// transformHTMLContentItem transforms a single HTML content item
func (vm *ViewTransformer) transformHTMLContentItem(item ast.Stmt) (ast.Expr, error) {
	switch content := item.(type) {
	case *ast.HTMLElement:
		// Nested HTML element - recursively transform
		return vm.transformHTMLElement(content)

	case *ast.HTMLContent:
		// HTML content with text and interpolations
		return vm.transformHTMLContentParts(content.Parts)

	default:
		// Other statements (if/for blocks, etc.) - for now, skip them
		// TODO: Handle control flow statements properly
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "", // placeholder
			Span:  lexer.Span{},
		}, nil
	}
}

// transformHTMLContentParts transforms HTML content parts (text + interpolations)
func (vm *ViewTransformer) transformHTMLContentParts(parts []ast.HTMLContentPart) (ast.Expr, error) {
	if len(parts) == 0 {
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil
	}

	if len(parts) == 1 {
		// Single part
		switch part := parts[0].(type) {
		case *ast.HTMLText:
			// Pure text
			return &ast.Literal{
				Type:  ast.LiteralTypeString,
				Value: part.Value,
				Span:  part.Span,
			}, nil

		case *ast.HTMLInterpolation:
			// Expression interpolation - transform the expression for view parameters
			transformedExpr := vm.transformExpression(part.Expression)
			escapeCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  part.Span,
				},
				Arguments: []*ast.Argument{{Value: transformedExpr, Span: part.Span}},
				Span:      part.Span,
			}
			return escapeCall, nil
		}
	}

	// Multiple parts - build an f-string expression
	var fStringParts []ast.FStringPart

	for _, part := range parts {
		switch p := part.(type) {
		case *ast.HTMLText:
			// Add text as an f-string middle part
			fStringParts = append(fStringParts, &ast.FStringMiddle{
				Value: p.Value,
				Span:  p.Span,
			})

		case *ast.HTMLInterpolation:
			// Transform the expression for view parameters and add as replacement field
			transformedExpr := vm.transformExpression(p.Expression)
			escapeCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  p.Span,
				},
				Arguments: []*ast.Argument{{Value: transformedExpr, Span: p.Span}},
				Span:      p.Span,
			}

			replacementField := &ast.FStringReplacementField{
				Expression: escapeCall,
				Equal:      false,
				Conversion: nil,
				FormatSpec: nil,
				Span:       p.Span,
			}
			fStringParts = append(fStringParts, replacementField)
		}
	}

	// Build the f-string
	return &ast.FString{
		Parts: fStringParts,
		Span:  lexer.Span{},
	}, nil
}

// createElCall creates an el() function call
func (vm *ViewTransformer) createElCall(tag string, content ast.Expr, attrs ast.Expr) *ast.Call {
	// Create el function reference
	elFunc := &ast.Name{
		Token: lexer.Token{
			Lexeme: "el",
			Type:   lexer.Identifier,
		},
		Span: content.GetSpan(),
	}

	// Create tag argument
	tagLiteral := &ast.Argument{
		Value: &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: tag,
			Span:  content.GetSpan(),
		},
		Span: content.GetSpan(),
	}

	contentArg := &ast.Argument{
		Value: content,
		Span:  content.GetSpan(),
	}

	args := []*ast.Argument{tagLiteral, contentArg}

	// Add attrs argument if provided
	if attrs != nil {
		attrsArg := &ast.Argument{
			Value: attrs,
			Span:  content.GetSpan(),
		}
		args = append(args, attrsArg)
	}

	return &ast.Call{
		Callee:    elFunc,
		Arguments: args,
		Span:      content.GetSpan(),
	}
}

// GetRequiredImports returns the imports needed for the transformed code
func (vm *ViewTransformer) GetRequiredImports() []*ast.ImportFromStmt {
	if !vm.needsRuntimeImports {
		return nil
	}

	// Create import from runtime import BaseView, el, escape, Element, fragment
	importNames := []*ast.ImportName{
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "BaseView", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "el", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "Element", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "FragmentElement", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
	}

	importStmt := &ast.ImportFromStmt{
		DottedName: &ast.DottedName{
			Names: []*ast.Name{
				{
					Token: lexer.Token{Lexeme: "runtime", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
			},
			Span: lexer.Span{},
		},
		Names:      importNames,
		IsWildcard: false,
		DotCount:   0,
		Span:       lexer.Span{},
	}

	return []*ast.ImportFromStmt{importStmt}
}

// transformExpression recursively transforms expressions, replacing view parameters with self.param
func (vm *ViewTransformer) transformExpression(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Name:
		// Check if this is a view parameter
		if vm.isViewParameter(e) {
			return vm.transformNameToSelfAttribute(e)
		}
		return e

	case *ast.Attribute:
		return &ast.Attribute{
			Object: vm.transformExpression(e.Object),
			Name:   e.Name,
			Span:   e.Span,
		}

	case *ast.Call:
		// Transform callee and arguments
		transformedArgs := make([]*ast.Argument, len(e.Arguments))
		for i, arg := range e.Arguments {
			transformedArgs[i] = &ast.Argument{
				Name:  arg.Name,
				Value: vm.transformExpression(arg.Value),
				Span:  arg.Span,
			}
		}
		return &ast.Call{
			Callee:    vm.transformExpression(e.Callee),
			Arguments: transformedArgs,
			Span:      e.Span,
		}

	case *ast.Subscript:
		transformedIndices := make([]ast.Expr, len(e.Indices))
		for i, index := range e.Indices {
			transformedIndices[i] = vm.transformExpression(index)
		}
		return &ast.Subscript{
			Object:  vm.transformExpression(e.Object),
			Indices: transformedIndices,
			Span:    e.Span,
		}

	case *ast.Binary:
		return &ast.Binary{
			Left:     vm.transformExpression(e.Left),
			Operator: e.Operator,
			Right:    vm.transformExpression(e.Right),
			Span:     e.Span,
		}

	case *ast.Unary:
		return &ast.Unary{
			Operator: e.Operator,
			Right:    vm.transformExpression(e.Right),
			Span:     e.Span,
		}

	case *ast.TernaryExpr:
		return &ast.TernaryExpr{
			Condition: vm.transformExpression(e.Condition),
			TrueExpr:  vm.transformExpression(e.TrueExpr),
			FalseExpr: vm.transformExpression(e.FalseExpr),
			Span:      e.Span,
		}

	case *ast.ListExpr:
		transformedElements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			transformedElements[i] = vm.transformExpression(elem)
		}
		return &ast.ListExpr{
			Elements: transformedElements,
			Span:     e.Span,
		}

	case *ast.TupleExpr:
		transformedElements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			transformedElements[i] = vm.transformExpression(elem)
		}
		return &ast.TupleExpr{
			Elements: transformedElements,
			Span:     e.Span,
		}

	case *ast.DictExpr:
		transformedPairs := make([]ast.DictPair, len(e.Pairs))
		for i, pair := range e.Pairs {
			switch p := pair.(type) {
			case *ast.KeyValuePair:
				transformedPairs[i] = &ast.KeyValuePair{
					Key:   vm.transformExpression(p.Key),
					Value: vm.transformExpression(p.Value),
					Span:  p.Span,
				}
			case *ast.DoubleStarredPair:
				transformedPairs[i] = &ast.DoubleStarredPair{
					Expr: vm.transformExpression(p.Expr),
					Span: p.Span,
				}
			default:
				transformedPairs[i] = pair
			}
		}
		return &ast.DictExpr{
			Pairs: transformedPairs,
			Span:  e.Span,
		}

	case *ast.FString:
		transformedParts := make([]ast.FStringPart, len(e.Parts))
		for i, part := range e.Parts {
			switch p := part.(type) {
			case *ast.FStringReplacementField:
				transformedParts[i] = &ast.FStringReplacementField{
					Expression: vm.transformExpression(p.Expression),
					Conversion: p.Conversion,
					FormatSpec: p.FormatSpec,
					Span:       p.Span,
				}
			default:
				transformedParts[i] = part
			}
		}
		return &ast.FString{
			Parts: transformedParts,
			Span:  e.Span,
		}

	case *ast.GroupExpr:
		return &ast.GroupExpr{
			Expression: vm.transformExpression(e.Expression),
			Span:       e.Span,
		}

	case *ast.StarExpr:
		return &ast.StarExpr{
			Expr: vm.transformExpression(e.Expr),
			Span: e.Span,
		}

	// For literals and other expressions that don't contain references, return as-is
	default:
		return e
	}
}

// transformStatement recursively transforms statements, replacing view parameters with self.param
func (vm *ViewTransformer) transformStatement(stmt ast.Stmt) ast.Stmt {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return &ast.ExprStmt{
			Expr: vm.transformExpression(s.Expr),
			Span: s.Span,
		}

	case *ast.AssignStmt:
		// Transform the value but not the targets (assignments create new bindings)
		transformedTargets := make([]ast.Expr, len(s.Targets))
		for i, target := range s.Targets {
			transformedTargets[i] = vm.transformExpression(target)
		}
		return &ast.AssignStmt{
			Targets: transformedTargets,
			Value:   vm.transformExpression(s.Value),
			Span:    s.Span,
		}

	case *ast.ReturnStmt:
		return &ast.ReturnStmt{
			Value: vm.transformExpression(s.Value),
			Span:  s.Span,
		}

	case *ast.If:
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		return &ast.If{
			Condition: vm.transformExpression(s.Condition),
			Body:      transformedBody,
			Else:      transformedElse,
			Span:      s.Span,
		}

	case *ast.For:
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		return &ast.For{
			Target:   vm.transformExpression(s.Target),
			Iterable: vm.transformExpression(s.Iterable),
			Body:     transformedBody,
			Else:     transformedElse,
			IsAsync:  s.IsAsync,
			Span:     s.Span,
		}

	case *ast.Function:
		// Transform function body
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		// Note: Parameters are not transformed as they create new local scope
		return &ast.Function{
			Name:           s.Name,
			TypeParameters: s.TypeParameters,
			Parameters:     s.Parameters, // Keep parameters as-is
			ReturnType:     vm.transformExpression(s.ReturnType),
			Body:           transformedBody,
			IsAsync:        s.IsAsync,
			Span:           s.Span,
		}

	case *ast.MultiStmt:
		transformedStmts := make([]ast.Stmt, len(s.Stmts))
		for i, stmt := range s.Stmts {
			transformedStmts[i] = vm.transformStatement(stmt)
		}
		return &ast.MultiStmt{
			Stmts: transformedStmts,
			Span:  s.Span,
		}

	// For other statements, return as-is for now
	default:
		return s
	}
}

// processHTMLContent processes HTMLContent and returns the transformed statements
func (vm *ViewTransformer) processHTMLContent(content *ast.HTMLContent) ([]ast.Stmt, error) {
	// Transform HTML content parts (text + interpolations)
	contentExpr, err := vm.transformHTMLContentParts(content.Parts)
	if err != nil {
		return nil, err
	}

	// Append to current context if we have one
	if vm.currentContext != "" {
		appendStmt := vm.createAppendStatement(vm.currentContext, contentExpr)
		return []ast.Stmt{appendStmt}, nil
	}

	// If no current context, return as expression statement
	return []ast.Stmt{&ast.ExprStmt{
		Expr: contentExpr,
		Span: content.Span,
	}}, nil
}
