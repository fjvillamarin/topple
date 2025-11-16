/*
Package depgraph implements dependency graph construction and analysis
for the Topple compiler's import system.

The dependency graph tracks dependencies between PSX files to determine
the correct compilation order and detect circular dependencies.

# Basic Usage

Create a new graph and add files:

	graph := depgraph.NewGraph()

	// Add files with their parsed ASTs
	graph.AddFile("/project/a.psx", moduleA)
	graph.AddFile("/project/b.psx", moduleB)
	graph.AddFile("/project/c.psx", moduleC)

	// Add dependencies (from file depends on to file)
	graph.AddDependency("/project/b.psx", "/project/a.psx") // b depends on a
	graph.AddDependency("/project/c.psx", "/project/b.psx") // c depends on b

	// Get compilation order using topological sort
	order, err := graph.GetCompilationOrder()
	if err != nil {
		// Handle cycle error
		cycleErr := err.(*depgraph.CycleError)
		fmt.Println(cycleErr.Error())
		return
	}

	// Compile files in dependency order
	for _, file := range order {
		// Compile file...
	}

# Extract Imports from AST

The package provides import extraction from parsed AST:

	// Extract imports from a module
	imports, err := depgraph.ExtractImports(module, "/project/main.psx", resolver)
	if err != nil {
		// Handle error
	}

	// Build graph from imports
	for _, imp := range imports {
		graph.AddDependency("/project/main.psx", imp.ModulePath)
	}

# Algorithms

The package uses two well-known graph algorithms:

Topological Sort (Kahn's Algorithm):
  - Determines valid compilation order for files
  - Time complexity: O(V + E) where V is files and E is dependencies
  - Detects cycles as a byproduct

Cycle Detection (Depth-First Search):
  - Finds all circular dependencies in the graph
  - Returns the actual paths forming cycles
  - Useful for detailed error reporting

# Error Handling

Circular dependencies are reported with detailed paths:

	err := graph.GetCompilationOrder()
	if cycleErr, ok := err.(*depgraph.CycleError); ok {
		// cycleErr.Cycles contains all detected cycles
		// Each cycle is a []string showing the dependency path
		fmt.Println(cycleErr.Error())
		// Output:
		// circular dependencies detected:
		//   Cycle 1:
		//     /project/a.psx
		//      ↓ imports
		//     /project/b.psx
		//      ↓ imports
		//     /project/a.psx
	}

# Integration

This package is designed to integrate with:
  - compiler/module: For resolving import paths to file paths
  - compiler/symbol: For tracking exported symbols
  - compiler/resolver: For name resolution with import context

Typical compilation pipeline:
 1. Parse all files → AST forest
 2. Extract imports → Build dependency graph
 3. Topological sort → Get compilation order
 4. For each file (in order):
    a. Resolve imports
    b. Transform AST
    c. Generate code
*/
package depgraph
