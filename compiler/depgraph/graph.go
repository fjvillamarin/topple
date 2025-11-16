package depgraph

import (
	"fmt"
	"topple/compiler/ast"
)

// DependencyGraph represents dependencies between files
type DependencyGraph struct {
	nodes map[string]*FileNode // File path -> node
	edges map[string][]string  // File path -> dependencies
}

// FileNode represents a single file in the graph
type FileNode struct {
	FilePath string      // Absolute file path
	AST      *ast.Module // Parsed AST
	Imports  []*Import   // All imports in this file
}

// Import represents a single import statement
type Import struct {
	Statement  ast.Node // ImportStmt or ImportFromStmt
	ModulePath string   // Resolved file path
	Names      []string // Imported names (empty for "import x")
	IsWildcard bool     // True for "from x import *"
	Location   Location // Source location
}

// Location represents a position in source code
type Location struct {
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed)
}

// NewGraph creates a new dependency graph
func NewGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*FileNode),
		edges: make(map[string][]string),
	}
}

// AddFile adds a file to the graph
func (g *DependencyGraph) AddFile(filePath string, module *ast.Module) error {
	if _, exists := g.nodes[filePath]; exists {
		return fmt.Errorf("file already added to graph: %s", filePath)
	}

	g.nodes[filePath] = &FileNode{
		FilePath: filePath,
		AST:      module,
		Imports:  []*Import{},
	}
	g.edges[filePath] = []string{}

	return nil
}

// AddDependency adds a dependency edge (from depends on to)
func (g *DependencyGraph) AddDependency(from, to string) error {
	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("source file not in graph: %s", from)
	}
	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("target file not in graph: %s", to)
	}

	// Avoid duplicate edges
	for _, dep := range g.edges[from] {
		if dep == to {
			return nil // Already exists, not an error
		}
	}

	g.edges[from] = append(g.edges[from], to)
	return nil
}

// GetDependencies returns files that the given file depends on
func (g *DependencyGraph) GetDependencies(filePath string) []string {
	deps, exists := g.edges[filePath]
	if !exists {
		return []string{}
	}
	// Return a copy to prevent external modification
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetDependents returns files that depend on the given file
func (g *DependencyGraph) GetDependents(filePath string) []string {
	var dependents []string
	for file, deps := range g.edges {
		for _, dep := range deps {
			if dep == filePath {
				dependents = append(dependents, file)
				break
			}
		}
	}
	return dependents
}

// HasFile checks if a file is in the graph
func (g *DependencyGraph) HasFile(filePath string) bool {
	_, exists := g.nodes[filePath]
	return exists
}

// FileCount returns the number of files in the graph
func (g *DependencyGraph) FileCount() int {
	return len(g.nodes)
}

// Clear removes all files and dependencies
func (g *DependencyGraph) Clear() {
	g.nodes = make(map[string]*FileNode)
	g.edges = make(map[string][]string)
}

// GetFileNode returns the FileNode for a given path
func (g *DependencyGraph) GetFileNode(filePath string) (*FileNode, bool) {
	node, exists := g.nodes[filePath]
	return node, exists
}

// GetAllFiles returns all file paths in the graph
func (g *DependencyGraph) GetAllFiles() []string {
	files := make([]string, 0, len(g.nodes))
	for file := range g.nodes {
		files = append(files, file)
	}
	return files
}
