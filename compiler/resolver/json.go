package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// JSON serialization types for resolution output

// JSONSpan represents a single position in source code
type JSONSpan struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// JSONSpanRange represents a range in source code (start and end)
type JSONSpanRange struct {
	Start JSONSpan `json:"start"`
	End   JSONSpan `json:"end"`
}

// JSONScope represents a lexical scope in the source code
type JSONScope struct {
	ID        int           `json:"id"`
	Type      string        `json:"type"`
	Span      JSONSpanRange `json:"span"`
	ParentID  *int          `json:"parent_id"`
	Bindings  []JSONBinding `json:"bindings"`
	Globals   []string      `json:"globals"`
	Nonlocals []string      `json:"nonlocals"`
	NodeType  string        `json:"node_type,omitempty"` // Type of AST node
}

// JSONBinding represents a variable binding in a scope
type JSONBinding struct {
	Name       string        `json:"name"`
	VariableID string        `json:"variable_id"`
	Span       JSONSpanRange `json:"span"`
	ShadowedBy *string       `json:"shadowed_by,omitempty"`
}

// JSONClassification holds variable classification flags
type JSONClassification struct {
	IsParameter     bool `json:"is_parameter"`
	IsGlobal        bool `json:"is_global"`
	IsNonlocal      bool `json:"is_nonlocal"`
	IsImported      bool `json:"is_imported"`
	IsViewParameter bool `json:"is_view_parameter"`
	IsExceptionVar  bool `json:"is_exception_var"`
}

// JSONUsage holds variable usage flags
type JSONUsage struct {
	IsUsed        bool `json:"is_used"`
	IsCaptured    bool `json:"is_captured"`
	IsCell        bool `json:"is_cell"`
	UsedBeforeDef bool `json:"used_before_def"`
}

// JSONVariableSpans holds source spans for a variable
type JSONVariableSpans struct {
	FirstDef *JSONSpanRange `json:"first_def,omitempty"`
	FirstUse *JSONSpanRange `json:"first_use,omitempty"`
}

// JSONReference represents a reference to a variable
type JSONReference struct {
	Span     JSONSpanRange `json:"span"`
	ScopeID  int           `json:"scope_id"`
	NodeType string        `json:"node_type"`
	Context  string        `json:"context"` // "definition", "reference", "assignment"
}

// JSONVariable represents variable metadata
type JSONVariable struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	DefinitionDepth int                `json:"definition_depth"`
	State           string             `json:"state"`
	Classification  JSONClassification `json:"classification"`
	Usage           JSONUsage          `json:"usage"`
	Spans           JSONVariableSpans  `json:"spans"`
	References      []JSONReference    `json:"references"`
}

// JSONView represents a view definition
type JSONView struct {
	Name    string        `json:"name"`
	ScopeID int           `json:"scope_id"`
	Span    JSONSpanRange `json:"span"`
}

// JSONViewReference represents a reference to a view (HTML element)
type JSONViewReference struct {
	ViewName string        `json:"view_name"`
	Span     JSONSpanRange `json:"span"`
	NodeType string        `json:"node_type"`
}

// JSONViews holds view composition information
type JSONViews struct {
	Defined    []JSONView          `json:"defined"`
	References []JSONViewReference `json:"references"`
}

// JSONClosure holds closure analysis information
type JSONClosure struct {
	CellVars []string `json:"cell_vars"`
	FreeVars []string `json:"free_vars"`
}

// JSONDiagnostic represents an error or warning
type JSONDiagnostic struct {
	Severity string         `json:"severity"` // "error" or "warning"
	Message  string         `json:"message"`
	Span     *JSONSpanRange `json:"span,omitempty"`
}

// JSONDiagnostics holds all diagnostics
type JSONDiagnostics struct {
	Errors   []JSONDiagnostic `json:"errors"`
	Warnings []JSONDiagnostic `json:"warnings"`
}

// JSONSummary holds summary statistics
type JSONSummary struct {
	TotalScopes     int `json:"total_scopes"`
	TotalVariables  int `json:"total_variables"`
	TotalReferences int `json:"total_references"`
	Parameters      int `json:"parameters"`
	ViewParameters  int `json:"view_parameters"`
	Globals         int `json:"globals"`
	Nonlocals       int `json:"nonlocals"`
	CellVars        int `json:"cell_vars"`
	FreeVars        int `json:"free_vars"`
}

// JSONResolution is the top-level structure for JSON output
type JSONResolution struct {
	File            string          `json:"file"`
	Scopes          []JSONScope     `json:"scopes"`
	Variables       []JSONVariable  `json:"variables"`
	Views           JSONViews       `json:"views"`
	ClosureAnalysis JSONClosure     `json:"closure_analysis"`
	Diagnostics     JSONDiagnostics `json:"diagnostics"`
	Summary         JSONSummary     `json:"summary"`
}

// ToJSON converts a ResolutionTable to JSON format
func (rt *ResolutionTable) ToJSON(filename string) (*JSONResolution, error) {
	result := &JSONResolution{
		File: filename,
	}

	// Build variable ID map (Variable pointer -> unique ID)
	varIDMap := make(map[*Variable]string)
	varIDCounter := 0
	for _, v := range rt.Variables {
		if _, exists := varIDMap[v]; !exists {
			varIDCounter++
			varIDMap[v] = fmt.Sprintf("var_%d", varIDCounter)
		}
	}

	// Build binding ID map for shadowing references
	bindingIDMap := make(map[*Binding]string)
	bindingIDCounter := 0
	for _, scope := range rt.Scopes {
		for _, binding := range scope.Bindings {
			bindingIDCounter++
			bindingIDMap[binding] = fmt.Sprintf("binding_%d", bindingIDCounter)
		}
	}

	// Convert scopes
	result.Scopes = convertScopes(rt.Scopes, varIDMap, bindingIDMap)

	// Convert variables with references
	result.Variables = convertVariables(rt, varIDMap)

	// Convert views
	result.Views = convertViews(rt, result.Scopes)

	// Convert closure analysis
	result.ClosureAnalysis = convertClosureAnalysis(rt)

	// Convert diagnostics
	result.Diagnostics = convertDiagnostics(rt.Errors)

	// Calculate summary
	result.Summary = calculateSummary(rt, result)

	return result, nil
}

// convertScopes converts all scopes to JSON format
func convertScopes(scopes map[int]*Scope, varIDMap map[*Variable]string, bindingIDMap map[*Binding]string) []JSONScope {
	var result []JSONScope

	// Sort by scope ID for consistent output
	var scopeIDs []int
	for id := range scopes {
		scopeIDs = append(scopeIDs, id)
	}
	sort.Ints(scopeIDs)

	for _, id := range scopeIDs {
		scope := scopes[id]
		jsonScope := JSONScope{
			ID:        scope.ID,
			Type:      scopeTypeToString(scope.ScopeType),
			Bindings:  convertBindings(scope.Bindings, varIDMap, bindingIDMap),
			Globals:   extractNames(scope.Globals),
			Nonlocals: extractNames(scope.Nonlocals),
		}

		// Set parent ID
		if scope.Parent != nil {
			parentID := scope.Parent.ID
			jsonScope.ParentID = &parentID
		}

		// Get span and node type from AST node if available
		if scope.ASTNode != nil {
			jsonScope.Span = spanToJSONRange(scope.ASTNode.GetSpan())
			jsonScope.NodeType = getNodeType(scope.ASTNode)
		}

		result = append(result, jsonScope)
	}

	return result
}

// convertBindings converts scope bindings to JSON format
func convertBindings(bindings map[string]*Binding, varIDMap map[*Variable]string, bindingIDMap map[*Binding]string) []JSONBinding {
	var result []JSONBinding

	// Sort by name for consistent output
	var names []string
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		binding := bindings[name]
		jsonBinding := JSONBinding{
			Name:       binding.Name,
			VariableID: varIDMap[binding.Variable],
			Span:       spanToJSONRange(binding.Variable.FirstDefSpan),
		}

		// Add shadowing reference
		if binding.ShadowedBy != nil {
			if shadowID, exists := bindingIDMap[binding.ShadowedBy]; exists {
				jsonBinding.ShadowedBy = &shadowID
			}
		}

		result = append(result, jsonBinding)
	}

	return result
}

// convertVariables converts all variables to JSON format with references
func convertVariables(rt *ResolutionTable, varIDMap map[*Variable]string) []JSONVariable {
	var result []JSONVariable

	// Group name nodes by variable
	varToNames := make(map[*Variable][]*ast.Name)
	for nameNode, variable := range rt.Variables {
		varToNames[variable] = append(varToNames[variable], nameNode)
	}

	// Sort variables by ID for consistent output
	type varEntry struct {
		variable *Variable
		id       string
	}
	var sortedVars []varEntry
	for v, id := range varIDMap {
		sortedVars = append(sortedVars, varEntry{v, id})
	}
	sort.Slice(sortedVars, func(i, j int) bool {
		return sortedVars[i].id < sortedVars[j].id
	})

	for _, entry := range sortedVars {
		variable := entry.variable
		jsonVar := JSONVariable{
			ID:              entry.id,
			Name:            variable.Name,
			DefinitionDepth: variable.DefinitionDepth,
			State:           variableStateToString(variable.State),
			Classification: JSONClassification{
				IsParameter:     variable.IsParameter,
				IsGlobal:        variable.IsGlobal,
				IsNonlocal:      variable.IsNonlocal,
				IsImported:      variable.IsImported,
				IsViewParameter: variable.IsViewParameter,
				IsExceptionVar:  variable.IsExceptionVar,
			},
			Usage: JSONUsage{
				IsUsed:        variable.IsUsed,
				IsCaptured:    variable.IsCaptured,
				IsCell:        variable.IsCell,
				UsedBeforeDef: variable.UsedBeforeDef,
			},
		}

		// Add spans
		if !isEmptySpan(variable.FirstDefSpan) {
			defSpan := spanToJSONRange(variable.FirstDefSpan)
			jsonVar.Spans.FirstDef = &defSpan
		}
		if !isEmptySpan(variable.FirstUseSpan) {
			useSpan := spanToJSONRange(variable.FirstUseSpan)
			jsonVar.Spans.FirstUse = &useSpan
		}

		// Add references
		if names, exists := varToNames[variable]; exists {
			jsonVar.References = convertReferences(names, rt)
		}

		result = append(result, jsonVar)
	}

	return result
}

// convertReferences converts name node references to JSON format
func convertReferences(names []*ast.Name, rt *ResolutionTable) []JSONReference {
	var result []JSONReference

	for _, name := range names {
		ref := JSONReference{
			Span:     spanToJSONRange(name.Span),
			NodeType: "Name",
			Context:  "reference", // Default to reference
		}

		// Determine scope ID
		if scope, exists := rt.NodeScopes[name]; exists {
			ref.ScopeID = scope.ID
		}

		// Try to determine context (definition vs reference vs assignment)
		// This is a best-effort based on available information
		if binding, exists := rt.NameToBinding[name]; exists {
			if binding.Variable.FirstDefSpan == name.Span {
				ref.Context = "definition"
			}
		}

		result = append(result, ref)
	}

	return result
}

// convertViews converts view composition data to JSON format
func convertViews(rt *ResolutionTable, scopes []JSONScope) JSONViews {
	views := JSONViews{
		Defined:    []JSONView{},
		References: []JSONViewReference{},
	}

	// Convert defined views
	for name, viewStmt := range rt.Views {
		// Find scope ID for this view
		scopeID := 0 // default to module scope
		for _, scope := range scopes {
			if scope.NodeType == "ViewStmt" {
				// Match by name from bindings
				for _, binding := range scope.Bindings {
					if binding.Name == name {
						scopeID = scope.ID
						break
					}
				}
			}
		}

		views.Defined = append(views.Defined, JSONView{
			Name:    name,
			ScopeID: scopeID,
			Span:    spanToJSONRange(viewStmt.Span),
		})
	}

	// Sort defined views by name
	sort.Slice(views.Defined, func(i, j int) bool {
		return views.Defined[i].Name < views.Defined[j].Name
	})

	// Convert view references (HTML elements that reference views)
	for htmlElement, viewStmt := range rt.ViewElements {
		views.References = append(views.References, JSONViewReference{
			ViewName: viewStmt.Name.Token.Lexeme,
			Span:     spanToJSONRange(htmlElement.Span),
			NodeType: "HTMLElement",
		})
	}

	return views
}

// convertClosureAnalysis converts closure data to JSON format
func convertClosureAnalysis(rt *ResolutionTable) JSONClosure {
	closure := JSONClosure{
		CellVars: []string{},
		FreeVars: []string{},
	}

	// Convert cell vars
	for name := range rt.CellVars {
		closure.CellVars = append(closure.CellVars, name)
	}
	sort.Strings(closure.CellVars)

	// Convert free vars
	for name := range rt.FreeVars {
		closure.FreeVars = append(closure.FreeVars, name)
	}
	sort.Strings(closure.FreeVars)

	return closure
}

// convertDiagnostics converts errors to JSON format
func convertDiagnostics(errors []error) JSONDiagnostics {
	diagnostics := JSONDiagnostics{
		Errors:   []JSONDiagnostic{},
		Warnings: []JSONDiagnostic{},
	}

	for _, err := range errors {
		diagnostic := JSONDiagnostic{
			Severity: "error",
			Message:  err.Error(),
		}
		diagnostics.Errors = append(diagnostics.Errors, diagnostic)
	}

	return diagnostics
}

// calculateSummary calculates summary statistics
func calculateSummary(rt *ResolutionTable, result *JSONResolution) JSONSummary {
	summary := JSONSummary{
		TotalScopes:     len(result.Scopes),
		TotalVariables:  len(result.Variables),
		TotalReferences: len(rt.Variables),
		ViewParameters:  len(rt.ViewParameters),
		CellVars:        len(rt.CellVars),
		FreeVars:        len(rt.FreeVars),
	}

	// Count variable types
	for _, v := range result.Variables {
		if v.Classification.IsParameter {
			summary.Parameters++
		}
		if v.Classification.IsGlobal {
			summary.Globals++
		}
		if v.Classification.IsNonlocal {
			summary.Nonlocals++
		}
	}

	return summary
}

// Helper functions

// spanToJSONRange converts a lexer.Span to JSONSpanRange
func spanToJSONRange(span lexer.Span) JSONSpanRange {
	return JSONSpanRange{
		Start: JSONSpan{
			Line:   span.Start.Line,
			Column: span.Start.Column,
		},
		End: JSONSpan{
			Line:   span.End.Line,
			Column: span.End.Column,
		},
	}
}

// isEmptySpan checks if a span is empty/uninitialized
func isEmptySpan(span lexer.Span) bool {
	return span.Start.Line == 0 && span.Start.Column == 0
}

// scopeTypeToString converts ScopeType to string
func scopeTypeToString(st ScopeType) string {
	switch st {
	case ModuleScopeType:
		return "module"
	case FunctionScopeType:
		return "function"
	case ClassScopeType:
		return "class"
	case ViewScopeType:
		return "view"
	case ComprehensionScopeType:
		return "comprehension"
	case ExceptScopeType:
		return "except"
	case WithScopeType:
		return "with"
	default:
		return "unknown"
	}
}

// variableStateToString converts VariableState to string
func variableStateToString(state VariableState) string {
	switch state {
	case VariableUndefined:
		return "undefined"
	case VariableDeclared:
		return "declared"
	case VariableDefined:
		return "defined"
	case VariableUsed:
		return "used"
	default:
		return "unknown"
	}
}

// getNodeType returns the type name of an AST node
func getNodeType(node ast.Node) string {
	if node == nil {
		return ""
	}
	t := reflect.TypeOf(node)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// extractNames extracts variable names from a binding map
func extractNames(bindings map[string]*Binding) []string {
	var names []string
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// WriteResolutionJSON writes a ResolutionTable to a JSON file
func WriteResolutionJSON(table *ResolutionTable, filename string, outputPath string) error {
	jsonData, err := table.ToJSON(filename)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
