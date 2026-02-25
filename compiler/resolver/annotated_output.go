package resolver

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// JSONAnnotatedOutput is the top-level JSON structure for annotated source output.
type JSONAnnotatedOutput struct {
	File  string              `json:"file"`
	Lines []JSONAnnotatedLine `json:"lines"`
}

// JSONAnnotatedLine represents a single source line with its annotations.
type JSONAnnotatedLine struct {
	Line        int              `json:"line"`
	Source      string           `json:"source"`
	Annotations []JSONAnnotation `json:"annotations,omitempty"`
}

// JSONAnnotation represents a single resolved name annotation on a source line.
type JSONAnnotation struct {
	StartCol int    `json:"start_col"`
	EndCol   int    `json:"end_col"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	State    string `json:"state"`
	ScopeID  int    `json:"scope_id"`
	Depth    int    `json:"depth"`
	Context  string `json:"context"`
}

// annotation is an internal type used during rendering.
type annotation struct {
	startCol int
	endCol   int
	name     string
	varType  string
	state    string
	scopeID  int
	depth    int
	context  string
	flags    string
}

// ToAnnotatedText renders source code with inline resolution annotations.
func (rt *ResolutionTable) ToAnnotatedText(filename string, source []byte) (string, error) {
	annotations := rt.buildAnnotations()
	lines := splitLines(source)

	var sb strings.Builder

	// Determine line number width for alignment
	lineWidth := len(fmt.Sprintf("%d", len(lines)))

	for i, line := range lines {
		lineNum := i + 1 // 1-based

		// Source line
		sb.WriteString(fmt.Sprintf("%*d | %s\n", lineWidth, lineNum, line))

		// Annotation lines for this source line
		lineAnnotations := annotations[lineNum]
		if len(lineAnnotations) == 0 {
			continue
		}

		// Sort by column
		sort.Slice(lineAnnotations, func(a, b int) bool {
			return lineAnnotations[a].startCol < lineAnnotations[b].startCol
		})

		padding := strings.Repeat(" ", lineWidth)
		for _, ann := range lineAnnotations {
			// Build the underline marker line
			// startCol is 1-based, we need 0-based for spacing
			col0 := ann.startCol - 1
			if col0 < 0 {
				col0 = 0
			}
			caretLen := ann.endCol - ann.startCol
			if caretLen < 1 {
				caretLen = 1
			}

			leadingSpaces := strings.Repeat(" ", col0)
			carets := strings.Repeat("^", caretLen)

			// Build metadata string
			meta := formatAnnotationMeta(ann)

			sb.WriteString(fmt.Sprintf("%s | %s%s %s\n", padding, leadingSpaces, carets, meta))
		}
	}

	return sb.String(), nil
}

// ToAnnotatedJSON returns a structured JSON representation of annotated source lines.
func (rt *ResolutionTable) ToAnnotatedJSON(filename string, source []byte) (*JSONAnnotatedOutput, error) {
	annotations := rt.buildAnnotations()
	lines := splitLines(source)

	result := &JSONAnnotatedOutput{
		File:  filename,
		Lines: make([]JSONAnnotatedLine, len(lines)),
	}

	for i, line := range lines {
		lineNum := i + 1
		jsonLine := JSONAnnotatedLine{
			Line:   lineNum,
			Source: line,
		}

		lineAnnotations := annotations[lineNum]
		if len(lineAnnotations) > 0 {
			sort.Slice(lineAnnotations, func(a, b int) bool {
				return lineAnnotations[a].startCol < lineAnnotations[b].startCol
			})

			for _, ann := range lineAnnotations {
				jsonLine.Annotations = append(jsonLine.Annotations, JSONAnnotation{
					StartCol: ann.startCol,
					EndCol:   ann.endCol,
					Name:     ann.name,
					Type:     ann.varType,
					State:    ann.state,
					ScopeID:  ann.scopeID,
					Depth:    ann.depth,
					Context:  ann.context,
				})
			}
		}

		result.Lines[i] = jsonLine
	}

	return result, nil
}

// WriteAnnotatedText writes annotated source output to a file.
func WriteAnnotatedText(table *ResolutionTable, filename string, source []byte, outputPath string) error {
	text, err := table.ToAnnotatedText(filename, source)
	if err != nil {
		return fmt.Errorf("failed to convert to annotated text: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// buildAnnotations extracts annotations from the resolution table, grouped by line number.
func (rt *ResolutionTable) buildAnnotations() map[int][]annotation {
	result := make(map[int][]annotation)

	for nameNode, variable := range rt.Variables {
		span := nameNode.GetSpan()
		lineNum := span.Start.Line
		if lineNum == 0 {
			continue
		}

		startCol := span.Start.Column
		endCol := span.End.Column

		// For multi-line spans, clamp to end of start line
		if span.End.Line != span.Start.Line {
			endCol = startCol + len(nameNode.Token.Lexeme)
		}

		// Determine scope ID
		scopeID := -1
		if binding, ok := rt.NameToBinding[nameNode]; ok && binding.Scope != nil {
			scopeID = binding.Scope.ID
		}

		// Determine context: definition or reference
		context := "reference"
		if span.Start.Line == variable.FirstDefSpan.Start.Line &&
			span.Start.Column == variable.FirstDefSpan.Start.Column {
			context = "definition"
		}

		flags := formatVariableFlags(variable)

		ann := annotation{
			startCol: startCol,
			endCol:   endCol,
			name:     variable.Name,
			varType:  formatVariableType(variable),
			state:    formatVariableState(variable.State),
			scopeID:  scopeID,
			depth:    variable.DefinitionDepth,
			context:  context,
			flags:    flags,
		}

		result[lineNum] = append(result[lineNum], ann)
	}

	return result
}

// formatAnnotationMeta builds the metadata string for a text annotation.
func formatAnnotationMeta(ann annotation) string {
	parts := []string{ann.varType}

	if ann.context == "definition" {
		parts = append(parts, "defined")
	} else {
		parts = append(parts, "ref")
	}

	parts = append(parts, fmt.Sprintf("scope #%d", ann.scopeID))
	parts = append(parts, fmt.Sprintf("depth=%d", ann.depth))

	if ann.flags != "—" {
		parts = append(parts, ann.flags)
	}

	return strings.Join(parts, ", ")
}

// splitLines splits source bytes into lines, stripping trailing newlines.
func splitLines(source []byte) []string {
	text := string(source)
	// Remove trailing newline to avoid ghost empty line
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
