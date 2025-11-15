package depgraph

// DetectCycles finds all cycles in the dependency graph using depth-first search.
// Returns a list of cycles, where each cycle is a path of files forming a loop.
// Returns nil if no cycles are found.
func (g *DependencyGraph) DetectCycles() ([][]string, error) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cycles := [][]string{}

	// Try starting DFS from each unvisited node
	for file := range g.nodes {
		if !visited[file] {
			path := []string{}
			g.dfsCycle(file, visited, recStack, path, &cycles)
		}
	}

	if len(cycles) > 0 {
		return cycles, NewCycleError(cycles)
	}
	return nil, nil
}

// dfsCycle performs depth-first search to detect cycles
func (g *DependencyGraph) dfsCycle(
	file string,
	visited map[string]bool,
	recStack map[string]bool,
	path []string,
	cycles *[][]string,
) {
	visited[file] = true
	recStack[file] = true
	path = append(path, file)

	// Visit all dependencies
	for _, dep := range g.edges[file] {
		if !visited[dep] {
			// Continue DFS
			g.dfsCycle(dep, visited, recStack, path, cycles)
		} else if recStack[dep] {
			// Found a back edge - cycle detected
			cycle := extractCycle(path, dep)
			*cycles = append(*cycles, cycle)
		}
	}

	recStack[file] = false
}

// extractCycle extracts the cycle path from the current DFS path
func extractCycle(path []string, cycleTo string) []string {
	// Find where the cycle starts in the path
	for i, file := range path {
		if file == cycleTo {
			// Return the cycle portion and close it
			cycleLen := len(path) - i + 1
			cycle := make([]string, cycleLen)
			copy(cycle, path[i:])
			cycle[cycleLen-1] = cycleTo // Close the cycle
			return cycle
		}
	}
	// If not found in path (shouldn't happen), return the full path
	return append(path, cycleTo)
}
