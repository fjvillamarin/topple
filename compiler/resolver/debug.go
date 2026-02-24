package resolver

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/ast"
	"sort"
	"strings"
)

// DebugPrintResolutionTable outputs a comprehensive, formatted debug report of the given resolution table, including errors, variable states, view parameters, view composition, closure variables, and summary statistics.
func DebugPrintResolutionTable(table *ResolutionTable) {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    RESOLUTION TABLE DEBUG                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Print errors first if any
	if len(table.Errors) > 0 {
		printErrors(table.Errors)
		fmt.Println()
	}

	// Print variable resolution table
	printVariablesTable(table.Variables, table.ScopeDepths)
	fmt.Println()

	// Print view parameters
	if len(table.ViewParameters) > 0 {
		printViewParameters(table.ViewParameters)
		fmt.Println()
	}

	// Print view composition information
	printViewComposition(table.Views, table.ViewElements)
	fmt.Println()

	// Print closure information
	printClosureInfo(table.CellVars, table.FreeVars)
	fmt.Println()

	// Print summary statistics
	printSummary(table)
}

// printErrors displays resolution errors in a formatted way
func printErrors(errors []error) {
	fmt.Println("🔴 RESOLUTION ERRORS:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	for i, err := range errors {
		fmt.Printf("  %d. %s\n", i+1, err.Error())
	}
}

// printVariablesTable displays the main variables resolution table
func printVariablesTable(variables map[*ast.Name]*Variable, scopeDepths map[*ast.Name]int) {
	fmt.Println("📊 VARIABLES RESOLUTION TABLE:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(variables) == 0 {
		fmt.Println("  (No variables resolved)")
		return
	}

	// Group variables by their Variable object to avoid duplicates
	varMap := make(map[*Variable][]*ast.Name)
	for nameNode, variable := range variables {
		varMap[variable] = append(varMap[variable], nameNode)
	}

	// Sort variables by name for consistent output
	type varInfo struct {
		variable *Variable
		names    []*ast.Name
	}
	var sortedVars []varInfo
	for variable, names := range varMap {
		sortedVars = append(sortedVars, varInfo{variable, names})
	}
	sort.Slice(sortedVars, func(i, j int) bool {
		return sortedVars[i].variable.Name < sortedVars[j].variable.Name
	})

	// Print table header
	fmt.Printf("  %-20s %-10s %-15s %-20s %s\n", "NAME", "STATE", "TYPE", "DEPTH", "FLAGS")
	fmt.Printf("  %-20s %-10s %-15s %-15s %s\n",
		strings.Repeat("─", 20),
		strings.Repeat("─", 10),
		strings.Repeat("─", 15),
		strings.Repeat("─", 15),
		strings.Repeat("─", 20))

	// Print each variable
	for _, info := range sortedVars {
		variable := info.variable
		names := info.names

		// Get scope depth for the first occurrence
		depth := "unknown"
		if len(names) > 0 {
			if d, exists := scopeDepths[names[0]]; exists {
				switch d {
				case 0:
					depth = "0 (module)"
				default:
					depth = fmt.Sprintf("%d", d)
				}
			} else {
				depth = "untracked"
			}
		}

		// Format variable state
		state := formatVariableState(variable.State)

		// Format variable type/category
		varType := formatVariableType(variable)

		// Format flags
		flags := formatVariableFlags(variable)

		fmt.Printf("  %-20s %-10s %-15s %-15s %s\n",
			variable.Name, state, varType, depth, flags)

		// Show multiple references if any
		if len(names) > 1 {
			fmt.Printf("  %-20s └─ %d references\n", "", len(names))
		}
	}
}

// printViewParameters prints a formatted list of view parameters and their flags for debugging purposes.
// If no view parameters are present, it indicates this explicitly.
func printViewParameters(viewParams map[string]*Variable) {
	fmt.Println("🎯 VIEW PARAMETERS:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━")

	if len(viewParams) == 0 {
		fmt.Println("  (No view parameters)")
		return
	}

	// Sort by name
	var names []string
	for name := range viewParams {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		variable := viewParams[name]
		flags := formatVariableFlags(variable)
		fmt.Printf("  %-20s %s\n", name, flags)
	}
}

// printViewComposition prints information about defined views and their associated HTML element bindings for debugging purposes.
func printViewComposition(views map[string]*ast.ViewStmt, viewElements map[*ast.HTMLElement]*ast.ViewStmt) {
	fmt.Println("🔧 VIEW COMPOSITION:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━")
	if len(views) > 0 {
		fmt.Printf("  Views defined: %d\n", len(views))
		for viewName := range views {
			fmt.Printf("    • %s\n", viewName)
		}
		fmt.Printf("  View elements bound: %d\n", len(viewElements))
		for _, viewStmt := range viewElements {
			fmt.Printf("    • <HTML> → %s view\n", viewStmt.Name.Token.Lexeme)
		}
	} else {
		fmt.Println("  (No views defined)")
	}
	fmt.Println()
}

// printClosureInfo prints closure analysis details, listing cell variables captured by nested functions and free variables from enclosing scopes. If no closure variables are present, it indicates so.
func printClosureInfo(cellVars, freeVars map[string]bool) {
	fmt.Println("🔗 CLOSURE ANALYSIS:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━")

	hasCells := len(cellVars) > 0
	hasFree := len(freeVars) > 0

	if !hasCells && !hasFree {
		fmt.Println("  (No closure variables)")
		return
	}

	if hasCells {
		fmt.Println("  Cell Variables (captured by nested functions):")
		var cellNames []string
		for name := range cellVars {
			cellNames = append(cellNames, name)
		}
		sort.Strings(cellNames)
		for _, name := range cellNames {
			fmt.Printf("    • %s\n", name)
		}
	}

	if hasFree {
		fmt.Println("  Free Variables (from enclosing scopes):")
		var freeNames []string
		for name := range freeVars {
			freeNames = append(freeNames, name)
		}
		sort.Strings(freeNames)
		for _, name := range freeNames {
			fmt.Printf("    • %s\n", name)
		}
	}
}

// printSummary displays summary statistics
func printSummary(table *ResolutionTable) {
	fmt.Println("📈 SUMMARY:")
	fmt.Println("━━━━━━━━━━━")

	// Count unique variables
	uniqueVars := make(map[*Variable]bool)
	for _, variable := range table.Variables {
		uniqueVars[variable] = true
	}

	// Count by category
	var parameters, globals, nonlocals, regular int
	for variable := range uniqueVars {
		if variable.IsParameter {
			parameters++
		} else if variable.IsGlobal {
			globals++
		} else if variable.IsNonlocal {
			nonlocals++
		} else {
			regular++
		}
	}

	fmt.Printf("  Total unique variables: %d\n", len(uniqueVars))
	fmt.Printf("  Parameters:             %d\n", parameters)
	fmt.Printf("  Global declarations:    %d\n", globals)
	fmt.Printf("  Nonlocal declarations:  %d\n", nonlocals)
	fmt.Printf("  Regular variables:      %d\n", regular)
	fmt.Printf("  View parameters:        %d\n", len(table.ViewParameters))
	fmt.Printf("  Cell variables:         %d\n", len(table.CellVars))
	fmt.Printf("  Free variables:         %d\n", len(table.FreeVars))
	fmt.Printf("  Total name references:  %d\n", len(table.Variables))
	fmt.Printf("  Resolution errors:      %d\n", len(table.Errors))
}

// Helper functions for formatting

func formatVariableState(state VariableState) string {
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

func formatVariableType(variable *Variable) string {
	if variable.IsParameter {
		if variable.IsViewParameter {
			return "view-param"
		}
		return "parameter"
	}
	if variable.IsGlobal {
		return "global"
	}
	if variable.IsNonlocal {
		return "nonlocal"
	}
	if variable.IsImported {
		return "imported"
	}
	if variable.IsExceptionVar {
		return "exception"
	}
	return "local"
}

func formatVariableFlags(variable *Variable) string {
	var flags []string

	if variable.IsUsed {
		flags = append(flags, "used")
	}
	if variable.IsCaptured {
		flags = append(flags, "captured")
	}
	if variable.IsCell {
		flags = append(flags, "cell")
	}
	if variable.UsedBeforeDef {
		flags = append(flags, "early-use")
	}

	if len(flags) == 0 {
		return "—"
	}
	return strings.Join(flags, ", ")
}

// DebugPrintResolver prints the current state of the resolver during resolution
func DebugPrintResolver(r *Resolver) {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     RESOLVER STATE DEBUG                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Print current scope stack
	fmt.Println("📚 SCOPE CHAIN:")
	fmt.Println("━━━━━━━━━━━━━━")
	if r.ScopeChain == nil {
		fmt.Println("  (No active scopes)")
	} else {
		// Build list of scopes from current to module
		var scopes []*Scope
		for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
			scopes = append(scopes, scope)
		}

		// Print from module (bottom) to current (top)
		for i := len(scopes) - 1; i >= 0; i-- {
			scope := scopes[i]
			depth := len(scopes) - 1 - i
			scopeTypeStr := formatScopeType(scope.ScopeType)

			if i == 0 {
				fmt.Printf("  %d. %s (current)\n", depth, scopeTypeStr)
			} else {
				fmt.Printf("  %d. %s\n", depth, scopeTypeStr)
			}

			// Show variables in this scope
			if len(scope.Bindings) > 0 {
				var names []string
				for name := range scope.Bindings {
					names = append(names, name)
				}
				sort.Strings(names)
				fmt.Printf("     Variables: %s\n", strings.Join(names, ", "))
			}
		}
	}
	fmt.Println()

	// Print context information
	fmt.Println("📍 CONTEXT:")
	fmt.Println("━━━━━━━━━━━")
	fmt.Printf("  Function scope depth: %d\n", r.FunctionScopeDepth)
	fmt.Printf("  Class scope depth:    %d\n", r.ClassScopeDepth)
	fmt.Printf("  View scope depth:     %d\n", r.ViewScopeDepth)
	if r.CurrentFunction != nil && r.CurrentFunction.Name != nil {
		fmt.Printf("  Current function:     %s\n", r.CurrentFunction.Name.Token.Lexeme)
	}
	if r.CurrentView != nil && r.CurrentView.Name != nil {
		fmt.Printf("  Current view:         %s\n", r.CurrentView.Name.Token.Lexeme)
	}
	fmt.Println()

	// Print module globals
	if len(r.ModuleGlobals) > 0 {
		fmt.Println("🌍 MODULE GLOBALS:")
		fmt.Println("━━━━━━━━━━━━━━━━")
		var names []string
		for name := range r.ModuleGlobals {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			variable := r.ModuleGlobals[name]
			state := formatVariableState(variable.State)
			flags := formatVariableFlags(variable)
			fmt.Printf("  %-15s %-10s %s\n", name, state, flags)
		}
	}
}

func formatScopeType(scopeType ScopeType) string {
	switch scopeType {
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

// DebugPrintCurrentScope prints debug info about the current resolver state
func DebugPrintCurrentScope(r *Resolver, context string) {
	fmt.Printf("=== DEBUG: %s ===\n", context)
	if r.ScopeChain != nil {
		fmt.Printf("Current scope type: %s\n", formatScopeType(r.ScopeChain.ScopeType))
		// Calculate scope depth
		depth := 0
		for scope := r.ScopeChain; scope != nil; scope = scope.Parent {
			depth++
		}
		fmt.Printf("Scope depth: %d\n", depth)
	} else {
		fmt.Println("Current scope: nil")
		fmt.Println("Scope depth: 0")
	}
	fmt.Printf("View scope depth: %d\n", r.ViewScopeDepth)
	fmt.Printf("Function scope depth: %d\n", r.FunctionScopeDepth)

	if r.ScopeChain != nil && len(r.ScopeChain.Bindings) > 0 {
		fmt.Println("Variables in current scope:")
		for name, binding := range r.ScopeChain.Bindings {
			fmt.Printf("  %s: %s (%s)\n", name, formatVariableState(binding.Variable.State), formatVariableType(binding.Variable))
		}
	}

	fmt.Printf("Total resolved variables: %d\n", len(r.Variables))
	fmt.Println("===========================")
}
