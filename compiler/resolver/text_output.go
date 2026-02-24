package resolver

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"topple/compiler/ast"
)

// ToText converts a ResolutionTable to a human-readable text format
func (rt *ResolutionTable) ToText(filename string) (string, error) {
	var sb strings.Builder

	// Header
	writeHeader(&sb, filename)

	// Scopes section
	writeScopesSection(&sb, rt)

	// Variables section
	writeVariablesSection(&sb, rt)

	// View composition section
	writeViewCompositionSection(&sb, rt)

	// Closure analysis section
	writeClosureSection(&sb, rt)

	// Diagnostics section
	writeDiagnosticsSection(&sb, rt)

	// Summary section
	writeSummarySection(&sb, rt)

	return sb.String(), nil
}

// WriteResolutionText writes a ResolutionTable to a text file
func WriteResolutionText(table *ResolutionTable, filename string, outputPath string) error {
	text, err := table.ToText(filename)
	if err != nil {
		return fmt.Errorf("failed to convert to text: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Header functions

func writeHeader(sb *strings.Builder, filename string) {
	separator := strings.Repeat("=", 80)
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("RESOLUTION OUTPUT: %s\n", filename))
	sb.WriteString(separator)
	sb.WriteString("\n\n")
}

func writeSectionHeader(sb *strings.Builder, title string) {
	separator := strings.Repeat("━", 80)
	sb.WriteString(fmt.Sprintf("%s:\n", strings.ToUpper(title)))
	sb.WriteString(separator)
	sb.WriteString("\n")
}

// Scopes section

func writeScopesSection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "scopes")

	if len(rt.Scopes) == 0 {
		sb.WriteString("  (No scopes)\n\n")
		return
	}

	// Sort scopes by ID
	var scopeIDs []int
	for id := range rt.Scopes {
		scopeIDs = append(scopeIDs, id)
	}
	sort.Ints(scopeIDs)

	for _, id := range scopeIDs {
		scope := rt.Scopes[id]
		writeScopeInfo(sb, scope, rt)
		sb.WriteString("\n")
	}
}

func writeScopeInfo(sb *strings.Builder, scope *Scope, rt *ResolutionTable) {
	// Scope header
	scopeType := formatScopeType(scope.ScopeType)
	if scope.Parent != nil {
		sb.WriteString(fmt.Sprintf("Scope #%d [%s] → parent: #%d\n",
			scope.ID, scopeType, scope.Parent.ID))
	} else {
		sb.WriteString(fmt.Sprintf("Scope #%d [%s]\n", scope.ID, scopeType))
	}

	// Bindings
	if len(scope.Bindings) > 0 {
		sb.WriteString("  Bindings:\n")

		// Sort bindings by name
		var names []string
		for name := range scope.Bindings {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			binding := scope.Bindings[name]
			variable := binding.Variable

			// Format: name | type | state | location | flags
			varType := formatVariableType(variable)
			state := formatVariableState(variable.State)
			location := formatSpan(variable.FirstDefSpan)
			flags := formatVariableFlags(variable)

			flagsStr := ""
			if flags != "—" {
				flagsStr = fmt.Sprintf("  [%s]", flags)
			}

			sb.WriteString(fmt.Sprintf("    %-20s %-12s %-10s (%s)%s\n",
				name, varType, state, location, flagsStr))
		}
	} else {
		sb.WriteString("  (No bindings)\n")
	}

	// Global declarations
	if len(scope.Globals) > 0 {
		var globalNames []string
		for name := range scope.Globals {
			globalNames = append(globalNames, name)
		}
		sort.Strings(globalNames)
		sb.WriteString(fmt.Sprintf("  Global declarations: %s\n", strings.Join(globalNames, ", ")))
	}

	// Nonlocal declarations
	if len(scope.Nonlocals) > 0 {
		var nonlocalNames []string
		for name := range scope.Nonlocals {
			nonlocalNames = append(nonlocalNames, name)
		}
		sort.Strings(nonlocalNames)
		sb.WriteString(fmt.Sprintf("  Nonlocal declarations: %s\n", strings.Join(nonlocalNames, ", ")))
	}
}

// Variables section

func writeVariablesSection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "variables")

	if len(rt.Variables) == 0 {
		sb.WriteString("  (No variables)\n\n")
		return
	}

	// Group variables by their Variable object to avoid duplicates
	varMap := make(map[*Variable][]*ast.Name)
	for nameNode, variable := range rt.Variables {
		varMap[variable] = append(varMap[variable], nameNode)
	}

	// Sort variables by name
	type varInfo struct {
		variable *Variable
		names    []*ast.Name
	}
	var sortedVars []varInfo
	for variable, names := range varMap {
		sortedVars = append(sortedVars, varInfo{variable, names})
	}
	sort.Slice(sortedVars, func(i, j int) bool {
		vi, vj := sortedVars[i].variable, sortedVars[j].variable
		if vi.Name != vj.Name {
			return vi.Name < vj.Name
		}
		if vi.DefinitionDepth != vj.DefinitionDepth {
			return vi.DefinitionDepth < vj.DefinitionDepth
		}
		if vi.FirstDefSpan.Start.Line != vj.FirstDefSpan.Start.Line {
			return vi.FirstDefSpan.Start.Line < vj.FirstDefSpan.Start.Line
		}
		return vi.FirstDefSpan.Start.Column < vj.FirstDefSpan.Start.Column
	})

	// Table header
	sb.WriteString(fmt.Sprintf("  %-20s %-12s %-10s %-8s %-20s %s\n",
		"NAME", "TYPE", "STATE", "DEPTH", "FLAGS", "REFS"))
	sb.WriteString(fmt.Sprintf("  %s\n",
		strings.Repeat("─", 78)))

	// Print each variable
	for _, info := range sortedVars {
		variable := info.variable
		names := info.names

		depth := fmt.Sprintf("%d", variable.DefinitionDepth)

		// Format variable info
		varType := formatVariableType(variable)
		state := formatVariableState(variable.State)
		flags := formatVariableFlags(variable)
		refs := fmt.Sprintf("%d", len(names))

		sb.WriteString(fmt.Sprintf("  %-20s %-12s %-10s %-8s %-20s %s\n",
			variable.Name, varType, state, depth, flags, refs))
	}

	sb.WriteString("\n")
}

// View composition section

func writeViewCompositionSection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "view composition")

	hasViews := len(rt.Views) > 0
	hasRefs := len(rt.ViewElements) > 0

	if !hasViews && !hasRefs {
		sb.WriteString("  (No views)\n\n")
		return
	}

	// Defined views
	if hasViews {
		sb.WriteString("  Defined Views:\n")
		var viewNames []string
		for name := range rt.Views {
			viewNames = append(viewNames, name)
		}
		sort.Strings(viewNames)

		for _, name := range viewNames {
			viewStmt := rt.Views[name]
			location := formatSpan(viewStmt.Span)
			sb.WriteString(fmt.Sprintf("    • %s (%s)\n", name, location))
		}
	} else {
		sb.WriteString("  Defined Views: (none)\n")
	}

	sb.WriteString("\n")

	// View references
	if hasRefs {
		sb.WriteString("  View References:\n")
		var refNames []string
		for _, viewStmt := range rt.ViewElements {
			refNames = append(refNames, viewStmt.Name.Token.Lexeme)
		}
		sort.Strings(refNames)
		for _, name := range refNames {
			sb.WriteString(fmt.Sprintf("    • <HTML> → %s\n", name))
		}
	} else {
		sb.WriteString("  View References: (none)\n")
	}

	sb.WriteString("\n")
}

// Closure section

func writeClosureSection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "closure analysis")

	hasCells := len(rt.CellVars) > 0
	hasFree := len(rt.FreeVars) > 0

	if !hasCells && !hasFree {
		sb.WriteString("  (No closure variables)\n\n")
		return
	}

	// Cell variables
	if hasCells {
		var cellNames []string
		for name := range rt.CellVars {
			cellNames = append(cellNames, name)
		}
		sort.Strings(cellNames)
		sb.WriteString(fmt.Sprintf("  Cell Variables: %s\n", strings.Join(cellNames, ", ")))
	} else {
		sb.WriteString("  Cell Variables: (none)\n")
	}

	// Free variables
	if hasFree {
		var freeNames []string
		for name := range rt.FreeVars {
			freeNames = append(freeNames, name)
		}
		sort.Strings(freeNames)
		sb.WriteString(fmt.Sprintf("  Free Variables: %s\n", strings.Join(freeNames, ", ")))
	} else {
		sb.WriteString("  Free Variables: (none)\n")
	}

	sb.WriteString("\n")
}

// Diagnostics section

func writeDiagnosticsSection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "diagnostics")

	if len(rt.Errors) == 0 {
		sb.WriteString("  ✓ No errors\n\n")
		return
	}

	sb.WriteString(fmt.Sprintf("  Errors: %d\n", len(rt.Errors)))
	for i, err := range rt.Errors {
		sb.WriteString(fmt.Sprintf("    %d. %s\n", i+1, err.Error()))
	}

	sb.WriteString("\n")
}

// Summary section

func writeSummarySection(sb *strings.Builder, rt *ResolutionTable) {
	writeSectionHeader(sb, "summary")

	// Count unique variables
	uniqueVars := make(map[*Variable]bool)
	for _, variable := range rt.Variables {
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

	sb.WriteString(fmt.Sprintf("  Total Scopes:           %d\n", len(rt.Scopes)))
	sb.WriteString(fmt.Sprintf("  Total Variables:        %d\n", len(uniqueVars)))
	sb.WriteString(fmt.Sprintf("  Total References:       %d\n", len(rt.Variables)))
	sb.WriteString(fmt.Sprintf("  Parameters:             %d\n", parameters))
	sb.WriteString(fmt.Sprintf("  View Parameters:        %d\n", len(rt.ViewParameters)))
	sb.WriteString(fmt.Sprintf("  Global Declarations:    %d\n", globals))
	sb.WriteString(fmt.Sprintf("  Nonlocal Declarations:  %d\n", nonlocals))
	sb.WriteString(fmt.Sprintf("  Regular Variables:      %d\n", regular))
	sb.WriteString(fmt.Sprintf("  Cell Variables:         %d\n", len(rt.CellVars)))
	sb.WriteString(fmt.Sprintf("  Free Variables:         %d\n", len(rt.FreeVars)))
	sb.WriteString(fmt.Sprintf("  Resolution Errors:      %d\n", len(rt.Errors)))

	sb.WriteString("\n")
}
