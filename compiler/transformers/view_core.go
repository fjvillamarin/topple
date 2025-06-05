package transformers

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
	"sylfie/compiler/resolver"
	"fmt"
)

// ViewTransformer transforms PSX view statements into Python classes
type ViewTransformer struct {
	// Track if we need to add psx_runtime imports
	needsRuntimeImports bool

	// Resolution table for parameter transformation
	resolutionTable *resolver.ResolutionTable

	// Context tracking for hierarchical HTML generation
	contextStack   []string // Stack of current children array names
	currentContext string   // Current children array name
	nextContextId  int      // Counter for generating unique context names

	// Slot information
	slots     map[string]*SlotInfo // Map of slot name to slot info (empty string for default slot)
	slotOrder []string             // Order of slot names as they appear in view definition
}

// SlotInfo contains information about a slot in a view
type SlotInfo struct {
	Name         string           // Slot name (empty for default slot)
	FallbackHTML []ast.Stmt       // Fallback content for the slot
	Element      *ast.HTMLElement // The slot element itself
}

// SlotContent represents content designated for a specific slot
type SlotContent struct {
	SlotName string     // Target slot name (empty for default slot)
	Content  []ast.Stmt // Content for the slot
}


// HTMLContext represents a context for collecting HTML children
type HTMLContext struct {
	ChildrenVarName string     // Name of the children array variable
	Children        []ast.Stmt // Statements that build the children
}

// NewViewTransformer creates a new ViewTransformer with the given resolution table
func NewViewTransformer(resolutionTable *resolver.ResolutionTable) *ViewTransformer {
	return &ViewTransformer{
		needsRuntimeImports: false,
		resolutionTable:     resolutionTable,
		contextStack:        []string{},
		currentContext:      "",
		nextContextId:       1000,
		slots:               make(map[string]*SlotInfo),
		slotOrder:           []string{},
	}
}

// generateContextName generates a unique name for a children array
func (vm *ViewTransformer) generateContextName(prefix string) string {
	name := fmt.Sprintf("_%s_children_%d", prefix, vm.nextContextId)
	vm.nextContextId += 1000 // Increment by fixed amount for deterministic output
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
	// Reset slots for each view transformation
	vm.slots = make(map[string]*SlotInfo)
	vm.slotOrder = []string{}

	// Analyze slots in the view body
	vm.analyzeSlots(viewStmt.Body)

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