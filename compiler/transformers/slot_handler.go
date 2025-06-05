package transformers

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
)

// SourceOrderSlot represents slot content in the order it appears in source
type SourceOrderSlot struct {
	SlotName string
	Content  []ast.Stmt
}

// analyzeSlots recursively analyzes the view body to find all slot elements
func (vm *ViewTransformer) analyzeSlots(body []ast.Stmt) {
	for _, stmt := range body {
		vm.analyzeSlotInStatement(stmt)
	}
}

// analyzeSlotInStatement recursively looks for slot elements in a statement
func (vm *ViewTransformer) analyzeSlotInStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		if s.TagName.Lexeme == "slot" {
			// Found a slot element
			slotName := vm.getSlotName(s)

			// Only add to order if we haven't seen this slot before
			if _, exists := vm.slots[slotName]; !exists {
				vm.slotOrder = append(vm.slotOrder, slotName)
			}

			vm.slots[slotName] = &SlotInfo{
				Name:         slotName,
				FallbackHTML: s.Content,
				Element:      s,
			}
		} else {
			// Recursively check content of non-slot elements
			for _, contentStmt := range s.Content {
				vm.analyzeSlotInStatement(contentStmt)
			}
		}
	case *ast.For:
		for _, bodyStmt := range s.Body {
			vm.analyzeSlotInStatement(bodyStmt)
		}
		for _, elseStmt := range s.Else {
			vm.analyzeSlotInStatement(elseStmt)
		}
	case *ast.If:
		for _, bodyStmt := range s.Body {
			vm.analyzeSlotInStatement(bodyStmt)
		}
		for _, elseStmt := range s.Else {
			vm.analyzeSlotInStatement(elseStmt)
		}
	case *ast.While:
		for _, bodyStmt := range s.Body {
			vm.analyzeSlotInStatement(bodyStmt)
		}
		for _, elseStmt := range s.Else {
			vm.analyzeSlotInStatement(elseStmt)
		}
	}
}

// getSlotName extracts the slot name from a slot element (empty string for default slot)
func (vm *ViewTransformer) getSlotName(slotElement *ast.HTMLElement) string {
	for _, attr := range slotElement.Attributes {
		if attr.Name.Lexeme == "name" {
			if literal, ok := attr.Value.(*ast.Literal); ok && literal.Type == ast.LiteralTypeString {
				return literal.Value.(string)
			}
		}
	}
	return "" // Default slot
}

// createSlotTypeAnnotation creates a type annotation for slot parameters: Union[BaseView, str, None]
func (vm *ViewTransformer) createSlotTypeAnnotation() ast.Expr {
	// For now, return None as a placeholder
	// TODO: Implement proper Union[BaseView, str, None] type annotation
	return nil
}

// transformSlotElementToExpression transforms a slot element into a conditional expression
func (vm *ViewTransformer) transformSlotElementToExpression(slotElement *ast.HTMLElement) (ast.Expr, error) {
	slotName := vm.getSlotName(slotElement)

	// Determine the slot variable name
	var slotVarName string
	if slotName == "" {
		slotVarName = "children"
	} else {
		slotVarName = slotName
	}

	// Create the slot attribute access (self.slotName)
	slotAttr := &ast.Attribute{
		Object: &ast.Name{
			Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
			Span:  slotElement.Span,
		},
		Name: lexer.Token{Lexeme: slotVarName, Type: lexer.Identifier},
		Span: slotElement.Span,
	}

	// Create conditional: if self.slotName is not None
	condition := &ast.Binary{
		Left:     slotAttr,
		Operator: lexer.Token{Type: lexer.IsNot, Lexeme: "is not"},
		Right: &ast.Literal{
			Type:  ast.LiteralTypeNone,
			Value: nil,
			Span:  slotElement.Span,
		},
		Span: slotElement.Span,
	}

	// Create render_child call for provided content
	renderChildCall := &ast.Call{
		Callee: &ast.Name{
			Token: lexer.Token{Lexeme: "render_child", Type: lexer.Identifier},
			Span:  slotElement.Span,
		},
		Arguments: []*ast.Argument{{
			Value: slotAttr,
			Span:  slotElement.Span,
		}},
		Span: slotElement.Span,
	}

	// Create fallback content
	var fallbackExpr ast.Expr
	if len(slotElement.Content) > 0 {
		// Transform fallback content
		fallbackContentExpr, err := vm.transformHTMLContent(slotElement.Content)
		if err != nil {
			return nil, err
		}
		fallbackExpr = fallbackContentExpr
	} else {
		// Empty fallback
		fallbackExpr = &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  slotElement.Span,
		}
	}

	// Create ternary expression: render_child(self.slot) if self.slot is not None else fallback
	return &ast.TernaryExpr{
		Condition: condition,
		TrueExpr:  renderChildCall,
		FalseExpr: fallbackExpr,
		Span:      slotElement.Span,
	}, nil
}

// transformViewCallWithSlots creates a view instantiation call with slot content support
func (vm *ViewTransformer) transformViewCallWithSlots(viewStmt *ast.ViewStmt, element *ast.HTMLElement) (*ast.Call, error) {
	// Get the base call without slot content
	baseCall := vm.transformViewCall(viewStmt, element.Attributes)

	// Collect slot content from the element's children
	slotContent, err := vm.collectSlotContent(element.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid slot usage in view %s: %v", viewStmt.Name.Token.Lexeme, err)
	}

	// Add slot arguments to the call
	for slotName, content := range slotContent {
		// Skip empty content arrays
		if len(content) == 0 {
			continue
		}

		var paramName string
		if slotName == "" {
			paramName = "children"
		} else {
			paramName = slotName
		}

		// For slot content that might contain control structures,
		// we need to create an element that can handle complex statements
		var contentExpr ast.Expr

		// Use the proper view body transformation to handle control structures
		transformedContent, err := vm.transformViewBody(content)
		if err != nil {
			// If transformation fails, skip this slot
			continue
		}

		// If we have transformed content, we need to wrap it in a way that can be passed as an argument
		if len(transformedContent) == 1 {
			// Single statement - check if it's directly usable
			if exprStmt, ok := transformedContent[0].(*ast.ExprStmt); ok {
				contentExpr = exprStmt.Expr
			} else {
				// Complex statement - we need to wrap it in a lambda or similar
				// For now, fall back to fragment approach
				tempContext := vm.generateContextName("slot")
				fragmentCall := &ast.Call{
					Callee: &ast.Name{
						Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
						Span:  element.Span,
					},
					Arguments: []*ast.Argument{{
						Value: &ast.Name{
							Token: lexer.Token{Lexeme: tempContext, Type: lexer.Identifier},
							Span:  element.Span,
						},
						Span: element.Span,
					}},
					Span: element.Span,
				}
				contentExpr = fragmentCall
			}
		} else {
			// Multiple statements - use fragment approach
			tempContext := vm.generateContextName("slot")
			fragmentCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
					Span:  element.Span,
				},
				Arguments: []*ast.Argument{{
					Value: &ast.Name{
						Token: lexer.Token{Lexeme: tempContext, Type: lexer.Identifier},
						Span:  element.Span,
					},
					Span: element.Span,
				}},
				Span: element.Span,
			}
			contentExpr = fragmentCall
		}

		slotArg := &ast.Argument{
			Name: &ast.Name{
				Token: lexer.Token{
					Lexeme: paramName,
					Type:   lexer.Identifier,
				},
				Span: element.Span,
			},
			Value: contentExpr,
			Span:  element.Span,
		}

		baseCall.Arguments = append(baseCall.Arguments, slotArg)
	}

	return baseCall, nil
}

// collectSlotContent groups the element's content by slot name
func (vm *ViewTransformer) collectSlotContent(content []ast.Stmt) (map[string][]ast.Stmt, error) {
	slotContent := make(map[string][]ast.Stmt)

	for _, stmt := range content {
		if htmlElement, ok := stmt.(*ast.HTMLElement); ok {
			// Check if this element has a slot attribute
			if vm.hasSlotAttribute(htmlElement) {
				// Element has a slot attribute - get the slot name and place in that slot
				slotName := vm.getElementSlotName(htmlElement)
				// Remove the slot attribute from the element before adding to content
				filteredElement := vm.removeSlotAttribute(htmlElement)
				slotContent[slotName] = append(slotContent[slotName], filteredElement)
			} else {
				// Element without slot attribute - check for nested slots which should be invalid
				if vm.hasNestedSlotAttributes(htmlElement) {
					return nil, fmt.Errorf("slot attributes found nested inside HTML element <%s>. Slot attributes can only be used on direct children of view elements", htmlElement.TagName.Lexeme)
				}
				slotContent[""] = append(slotContent[""], stmt)
			}
		} else {
			// Check for nested slot attributes in control structures
			if vm.hasNestedSlotAttributesInStmt(stmt) {
				return nil, fmt.Errorf("slot attributes found inside control structures. Slot attributes can only be used on direct children of view elements, not within if/for/while statements")
			}
			// Non-HTML elements go to the default slot
			slotContent[""] = append(slotContent[""], stmt)
		}
	}

	return slotContent, nil
}

// collectSlotContentInSourceOrder groups the element's content by slot name in source order
func (vm *ViewTransformer) collectSlotContentInSourceOrder(content []ast.Stmt) []SourceOrderSlot {
	var sourceOrderSlots []SourceOrderSlot
	var currentSlotContent []ast.Stmt
	var currentSlotName *string // Use pointer to distinguish between "" and unset

	for _, stmt := range content {
		var stmtSlotName string
		var processedStmt ast.Stmt = stmt

		if htmlElement, ok := stmt.(*ast.HTMLElement); ok {
			// Check if this element has a slot attribute
			if vm.hasSlotAttribute(htmlElement) {
				// Element has a slot attribute
				stmtSlotName = vm.getElementSlotName(htmlElement)
				// Remove the slot attribute from the element before adding to content
				processedStmt = vm.removeSlotAttribute(htmlElement)
			} else {
				// Element without slot attribute - goes to default slot
				stmtSlotName = ""
			}
		} else {
			// Non-HTML elements go to default slot
			stmtSlotName = ""
		}

		// Check if we're starting a new slot or continuing the current one
		if currentSlotName == nil || *currentSlotName != stmtSlotName {
			// Starting a new slot - finish the current one if it exists
			if currentSlotName != nil && len(currentSlotContent) > 0 {
				sourceOrderSlots = append(sourceOrderSlots, SourceOrderSlot{
					SlotName: *currentSlotName,
					Content:  currentSlotContent,
				})
			}
			// Start new slot
			currentSlotName = &stmtSlotName
			currentSlotContent = []ast.Stmt{processedStmt}
		} else {
			// Continuing current slot
			currentSlotContent = append(currentSlotContent, processedStmt)
		}
	}

	// Add final slot if there's remaining content
	if currentSlotName != nil && len(currentSlotContent) > 0 {
		sourceOrderSlots = append(sourceOrderSlots, SourceOrderSlot{
			SlotName: *currentSlotName,
			Content:  currentSlotContent,
		})
	}

	return sourceOrderSlots
}

// hasSlotAttribute checks if an HTML element has a slot attribute
func (vm *ViewTransformer) hasSlotAttribute(element *ast.HTMLElement) bool {
	for _, attr := range element.Attributes {
		if attr.Name.Lexeme == "slot" {
			return true
		}
	}
	return false
}

// hasNestedSlotAttributes recursively checks if an HTML element contains nested elements with slot attributes
func (vm *ViewTransformer) hasNestedSlotAttributes(element *ast.HTMLElement) bool {
	for _, contentStmt := range element.Content {
		if vm.hasNestedSlotAttributesInStmt(contentStmt) {
			return true
		}
	}
	return false
}

// hasNestedSlotAttributesInStmt checks if a statement contains slot attributes (indicating invalid nesting)
func (vm *ViewTransformer) hasNestedSlotAttributesInStmt(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		// Check if this element has a slot attribute (invalid if nested)
		if vm.hasSlotAttribute(s) {
			return true
		}
		// Recursively check content
		return vm.hasNestedSlotAttributes(s)
	case *ast.For:
		// Check for slot attributes in for loop body
		for _, bodyStmt := range s.Body {
			if vm.hasNestedSlotAttributesInStmt(bodyStmt) {
				return true
			}
		}
		for _, elseStmt := range s.Else {
			if vm.hasNestedSlotAttributesInStmt(elseStmt) {
				return true
			}
		}
	case *ast.If:
		// Check for slot attributes in if statement body
		for _, bodyStmt := range s.Body {
			if vm.hasNestedSlotAttributesInStmt(bodyStmt) {
				return true
			}
		}
		for _, elseStmt := range s.Else {
			if vm.hasNestedSlotAttributesInStmt(elseStmt) {
				return true
			}
		}
	case *ast.While:
		// Check for slot attributes in while loop body
		for _, bodyStmt := range s.Body {
			if vm.hasNestedSlotAttributesInStmt(bodyStmt) {
				return true
			}
		}
		for _, elseStmt := range s.Else {
			if vm.hasNestedSlotAttributesInStmt(elseStmt) {
				return true
			}
		}
	}
	return false
}

// validateSlotUsage validates that slot attributes are only used on direct children of view elements
func (vm *ViewTransformer) validateSlotUsage(element *ast.HTMLElement) error {
	// Only validate if this is a view element
	if _, isView := vm.isViewElement(element); !isView {
		return nil
	}

	// Check each direct child - we only care about direct children that have invalid nested slots
	for _, stmt := range element.Content {
		if err := vm.validateDirectChild(stmt); err != nil {
			return fmt.Errorf("slot attributes can only be used on direct children of view elements, not nested within control structures or other elements. Found slot attribute nested within <%s> view: %v",
				element.TagName.Lexeme, err)
		}
	}

	return nil
}

// validateDirectChild validates that a direct child of a view element doesn't have improperly nested slot attributes
func (vm *ViewTransformer) validateDirectChild(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		// If this HTML element has a slot attribute, it's a valid direct child
		if vm.hasSlotAttribute(s) {
			// This is valid - it's a direct child with a slot attribute
			// But check that IT doesn't have nested slot attributes in its content
			return vm.validateSlotElementContent(s)
		} else {
			// This HTML element doesn't have a slot attribute, so it goes to the default slot
			// Check if it contains any nested slot attributes (which would be invalid)
			return vm.validateSlotElementContent(s)
		}
	case *ast.For, *ast.If, *ast.While:
		// Control structures are allowed, but they cannot contain slot attributes inside them
		// because slot attributes must be on direct children
		if vm.hasNestedSlotAttributesInStmt(s) {
			return fmt.Errorf("found slot attribute inside control structure")
		}
	}
	return nil
}

// validateSlotElementContent validates that an element doesn't contain nested slot attributes
func (vm *ViewTransformer) validateSlotElementContent(element *ast.HTMLElement) error {
	// Check if any content inside this element has slot attributes (which would be invalid)
	for _, contentStmt := range element.Content {
		if vm.hasNestedSlotAttributesInStmt(contentStmt) {
			return fmt.Errorf("found slot attribute nested inside HTML element")
		}
	}
	return nil
}

// removeSlotAttribute creates a copy of the HTML element without the slot attribute
func (vm *ViewTransformer) removeSlotAttribute(element *ast.HTMLElement) *ast.HTMLElement {
	var filteredAttrs []ast.HTMLAttribute

	for _, attr := range element.Attributes {
		if attr.Name.Lexeme != "slot" {
			filteredAttrs = append(filteredAttrs, attr)
		}
	}

	// Create a copy of the element with filtered attributes
	return &ast.HTMLElement{
		Type:       element.Type,
		TagName:    element.TagName,
		Attributes: filteredAttrs,
		Content:    element.Content,
		IsClosing:  element.IsClosing,
		Span:       element.Span,
	}
}

// getElementSlotName extracts the slot attribute value from an HTML element
func (vm *ViewTransformer) getElementSlotName(element *ast.HTMLElement) string {
	for _, attr := range element.Attributes {
		if attr.Name.Lexeme == "slot" {
			if literal, ok := attr.Value.(*ast.Literal); ok && literal.Type == ast.LiteralTypeString {
				return literal.Value.(string)
			}
		}
	}
	return "" // Default slot
}

// processSlotElement processes a slot element and generates the appropriate slot rendering code
func (vm *ViewTransformer) processSlotElement(slotElement *ast.HTMLElement) ([]ast.Stmt, error) {
	slotName := vm.getSlotName(slotElement)

	// Determine the slot variable name
	var slotVarName string
	if slotName == "" {
		slotVarName = "children"
	} else {
		slotVarName = slotName
	}

	// Create the slot attribute access (self.slotName)
	slotAttr := &ast.Attribute{
		Object: &ast.Name{
			Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
			Span:  slotElement.Span,
		},
		Name: lexer.Token{Lexeme: slotVarName, Type: lexer.Identifier},
		Span: slotElement.Span,
	}

	// Create conditional: if self.slotName is not None
	condition := &ast.Binary{
		Left:     slotAttr,
		Operator: lexer.Token{Type: lexer.IsNot, Lexeme: "is not"},
		Right: &ast.Literal{
			Type:  ast.LiteralTypeNone,
			Value: nil,
			Span:  slotElement.Span,
		},
		Span: slotElement.Span,
	}

	// Create render_child call for provided content
	renderChildCall := &ast.Call{
		Callee: &ast.Name{
			Token: lexer.Token{Lexeme: "render_child", Type: lexer.Identifier},
			Span:  slotElement.Span,
		},
		Arguments: []*ast.Argument{{
			Value: slotAttr,
			Span:  slotElement.Span,
		}},
		Span: slotElement.Span,
	}

	// Create fallback content
	var fallbackExpr ast.Expr
	if len(slotElement.Content) > 0 {
		// Transform fallback content
		fallbackContentExpr, err := vm.transformHTMLContent(slotElement.Content)
		if err != nil {
			return nil, err
		}
		fallbackExpr = fallbackContentExpr
	} else {
		// Empty fallback
		fallbackExpr = &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  slotElement.Span,
		}
	}

	// Create ternary expression: render_child(self.slot) if self.slot is not None else fallback
	slotExpr := &ast.TernaryExpr{
		Condition: condition,
		TrueExpr:  renderChildCall,
		FalseExpr: fallbackExpr,
		Span:      slotElement.Span,
	}

	// Handle based on whether we have a parent context
	if vm.currentContext != "" {
		// Append to parent context
		appendStmt := vm.createAppendStatement(vm.currentContext, slotExpr)
		return []ast.Stmt{appendStmt}, nil
	} else {
		// No parent context - this is a root element, return it directly
		returnStmt := &ast.ReturnStmt{
			Value: slotExpr,
			Span:  slotElement.Span,
		}
		return []ast.Stmt{returnStmt}, nil
	}
}

// getOrderedSlotNames returns slot names in the correct order:
// 1. children (default slot) first
// 2. named slots in the order they were defined in the view
func (vm *ViewTransformer) getOrderedSlotNames(slotContent map[string][]ast.Stmt) []string {
	var orderedNames []string

	// Add children (default slot) first if it exists in the content
	if _, hasDefault := slotContent[""]; hasDefault {
		orderedNames = append(orderedNames, "")
	}

	// Add named slots in the order they were defined in the view
	for _, slotName := range vm.slotOrder {
		// Skip default slot (already added) and slots not present in content
		if slotName != "" {
			if _, hasSlot := slotContent[slotName]; hasSlot {
				orderedNames = append(orderedNames, slotName)
			}
		}
	}

	return orderedNames
}

// getOrderedSlotNamesForView returns slot names in the correct order for a specific view:
// 1. children (default slot) first
// 2. named slots in the order they were defined in the target view
func (vm *ViewTransformer) getOrderedSlotNamesForView(slotContent map[string][]ast.Stmt, targetView *ast.ViewStmt) []string {
	var orderedNames []string

	// Add children (default slot) first if it exists in the content
	if _, hasDefault := slotContent[""]; hasDefault {
		orderedNames = append(orderedNames, "")
	}

	// Analyze the target view to get its slot order
	targetSlotOrder := vm.analyzeViewSlotOrder(targetView)

	// Add named slots in the order they were defined in the target view
	for _, slotName := range targetSlotOrder {
		// Skip default slot (already added) and slots not present in content
		if slotName != "" {
			if _, hasSlot := slotContent[slotName]; hasSlot {
				orderedNames = append(orderedNames, slotName)
			}
		}
	}

	return orderedNames
}

// analyzeViewSlotOrder analyzes a view's body to determine the order of slot definitions
func (vm *ViewTransformer) analyzeViewSlotOrder(viewStmt *ast.ViewStmt) []string {
	var slotOrder []string
	slots := make(map[string]bool)

	// Recursively analyze the view body to find slots in order
	vm.analyzeSlotOrderInStatements(viewStmt.Body, &slotOrder, slots)

	return slotOrder
}

// analyzeSlotOrderInStatements recursively analyzes statements to find slot elements in order
func (vm *ViewTransformer) analyzeSlotOrderInStatements(stmts []ast.Stmt, slotOrder *[]string, slots map[string]bool) {
	for _, stmt := range stmts {
		vm.analyzeSlotOrderInStatement(stmt, slotOrder, slots)
	}
}

// analyzeSlotOrderInStatement recursively looks for slot elements in a statement
func (vm *ViewTransformer) analyzeSlotOrderInStatement(stmt ast.Stmt, slotOrder *[]string, slots map[string]bool) {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		if s.TagName.Lexeme == "slot" {
			// Found a slot element
			slotName := vm.getSlotName(s)

			// Only add to order if we haven't seen this slot before
			if !slots[slotName] {
				*slotOrder = append(*slotOrder, slotName)
				slots[slotName] = true
			}
		} else {
			// Recursively check content of non-slot elements
			vm.analyzeSlotOrderInStatements(s.Content, slotOrder, slots)
		}
	case *ast.For:
		vm.analyzeSlotOrderInStatements(s.Body, slotOrder, slots)
		vm.analyzeSlotOrderInStatements(s.Else, slotOrder, slots)
	case *ast.If:
		vm.analyzeSlotOrderInStatements(s.Body, slotOrder, slots)
		vm.analyzeSlotOrderInStatements(s.Else, slotOrder, slots)
	case *ast.While:
		vm.analyzeSlotOrderInStatements(s.Body, slotOrder, slots)
		vm.analyzeSlotOrderInStatements(s.Else, slotOrder, slots)
	}
}
