package depgraph

import (
	"fmt"
	"strings"
)

// CycleError represents a circular dependency error
type CycleError struct {
	Cycles [][]string // List of cycles, each cycle is a path of files
}

// Error returns a formatted error message showing all detected cycles
func (e *CycleError) Error() string {
	var sb strings.Builder
	sb.WriteString("circular dependencies detected:\n")

	for i, cycle := range e.Cycles {
		sb.WriteString(fmt.Sprintf("  Cycle %d:\n", i+1))
		for j, file := range cycle {
			sb.WriteString(fmt.Sprintf("    %s\n", file))
			if j < len(cycle)-1 {
				sb.WriteString("     ↓ imports\n")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// NewCycleError creates a new cycle error
func NewCycleError(cycles [][]string) *CycleError {
	return &CycleError{Cycles: cycles}
}
