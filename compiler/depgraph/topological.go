package depgraph

// GetCompilationOrder returns files in topological order using Kahn's algorithm.
// Returns an error if circular dependencies are detected.
//
// The algorithm works as follows:
//  1. Calculate in-degrees (number of dependencies) for all files
//  2. Queue all files with in-degree 0 (no dependencies)
//  3. While queue is not empty:
//     a. Dequeue a file and add to result
//     b. For each file that depends on it:
//     - Decrement its in-degree
//     - If in-degree becomes 0, add to queue
//  4. If result length < total files, circular dependency detected
func (g *DependencyGraph) GetCompilationOrder() ([]string, error) {
	// 1. Calculate in-degrees (number of dependencies for each file)
	// in-degree[file] = number of files that file depends on
	inDegree := make(map[string]int)
	for file, deps := range g.edges {
		inDegree[file] = len(deps)
	}

	// Ensure all files have an entry (even if no dependencies)
	for file := range g.nodes {
		if _, exists := inDegree[file]; !exists {
			inDegree[file] = 0
		}
	}

	// 2. Queue files with no dependencies (in-degree 0)
	queue := []string{}
	for file, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, file)
		}
	}

	// 3. Process files in dependency order
	result := []string{}
	for len(queue) > 0 {
		// Dequeue
		file := queue[0]
		queue = queue[1:]
		result = append(result, file)

		// For each file that this file depends on (reverse direction)
		// We need to find files that have this file as a dependency
		for f, deps := range g.edges {
			for _, dep := range deps {
				if dep == file {
					// File f depends on current file
					// Decrement in-degree of f
					inDegree[f]--
					if inDegree[f] == 0 {
						queue = append(queue, f)
					}
				}
			}
		}
	}

	// 4. Check for cycles
	if len(result) != len(g.nodes) {
		// Cycle detected - use cycle detection to get details
		cycles, _ := g.DetectCycles()
		return nil, NewCycleError(cycles)
	}

	return result, nil
}
