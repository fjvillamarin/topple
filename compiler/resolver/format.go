package resolver

import (
	"fmt"
	"strings"

	"topple/compiler/lexer"
)

// formatVariableState returns a human-readable string for a VariableState.
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

// formatVariableType returns a human-readable type classification for a Variable.
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

// formatVariableFlags returns a comma-separated list of variable usage flags.
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

// formatScopeType returns a human-readable string for a ScopeType.
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

// formatSpan returns a human-readable string for a source span (line:column).
func formatSpan(span lexer.Span) string {
	if span.Start.Line == 0 && span.Start.Column == 0 {
		return "?"
	}
	return fmt.Sprintf("%d:%d", span.Start.Line, span.Start.Column)
}
