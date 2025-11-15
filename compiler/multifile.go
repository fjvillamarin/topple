package compiler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"topple/compiler/ast"
	"topple/compiler/codegen"
	"topple/compiler/depgraph"
	"topple/compiler/lexer"
	"topple/compiler/module"
	"topple/compiler/parser"
	"topple/compiler/resolver"
	"topple/compiler/symbol"
	"topple/compiler/transformers"
	"topple/internal/filesystem"
)

// MultiFileOptions configures multi-file compilation
type MultiFileOptions struct {
	RootDir     string   // Project root for module resolution
	Files       []string // Explicit file list (absolute paths)
	SearchPaths []string // Additional search paths for imports
}

// CompilationError represents an error during multi-file compilation
type CompilationError struct {
	File    string // File where error occurred
	Stage   string // Compilation stage: "parse", "resolve", "transform", "codegen"
	Message string // Error message
	Details error  // Underlying error
}

func (e *CompilationError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s [%s]: %s: %v", e.File, e.Stage, e.Message, e.Details)
	}
	return fmt.Sprintf("%s [%s]: %s", e.File, e.Stage, e.Message)
}

// MultiFileOutput contains the results of multi-file compilation
type MultiFileOutput struct {
	CompiledFiles map[string][]byte         // filepath -> generated Python code
	Registry      *symbol.Registry          // Symbol registry with all exports
	Graph         *depgraph.DependencyGraph // Dependency graph
	Errors        []*CompilationError       // All compilation errors
}

// MultiFileCompiler compiles multiple interdependent PSX files
type MultiFileCompiler struct {
	logger         *slog.Logger
	fs             filesystem.FileSystem
	moduleResolver *module.StandardResolver
	symbolRegistry *symbol.Registry
	depGraph       *depgraph.DependencyGraph
}

// NewMultiFileCompiler creates a new multi-file compiler
func NewMultiFileCompiler(logger *slog.Logger) *MultiFileCompiler {
	return &MultiFileCompiler{
		logger:         logger,
		fs:             nil, // Will be initialized in CompileProject
		moduleResolver: nil, // Will be initialized in CompileProject
		symbolRegistry: symbol.NewRegistry(),
		depGraph:       depgraph.NewGraph(),
	}
}

// CompileProject compiles multiple PSX files with import resolution
func (c *MultiFileCompiler) CompileProject(ctx context.Context, opts MultiFileOptions) (*MultiFileOutput, error) {
	output := &MultiFileOutput{
		CompiledFiles: make(map[string][]byte),
		Registry:      c.symbolRegistry,
		Graph:         c.depGraph,
		Errors:        []*CompilationError{},
	}

	// Initialize filesystem and module resolver with project root
	if opts.RootDir == "" {
		return nil, fmt.Errorf("RootDir is required")
	}

	c.fs = filesystem.NewFileSystem(c.logger)

	// Create module resolver config
	resolverConfig := module.Config{
		RootDir:     opts.RootDir,
		SearchPaths: opts.SearchPaths,
		FileSystem:  c.fs,
	}
	c.moduleResolver = module.NewResolver(resolverConfig)

	// Stage 1: Collect all files
	c.logger.Info("Stage 1: Collecting files")
	files, err := c.collectAllFiles(opts.Files)
	if err != nil {
		return nil, fmt.Errorf("file collection failed: %w", err)
	}
	c.logger.Info("Collected files", "count", len(files))

	// Stage 2: Parse all files to AST
	c.logger.Info("Stage 2: Parsing all files")
	astMap, parseErrs := c.parseAllFiles(ctx, files)
	if len(parseErrs) > 0 {
		output.Errors = append(output.Errors, parseErrs...)
		return output, fmt.Errorf("parsing failed with %d errors", len(parseErrs))
	}
	c.logger.Info("Parsed all files", "count", len(astMap))

	// Stage 3: Build dependency graph
	c.logger.Info("Stage 3: Building dependency graph")
	graphErrs := c.buildDependencyGraph(ctx, astMap)
	if len(graphErrs) > 0 {
		output.Errors = append(output.Errors, graphErrs...)
		// Continue - missing imports will be caught during resolution
	}
	c.logger.Info("Dependency graph built", "files", c.depGraph.FileCount())

	// Stage 4: Get compilation order (topological sort)
	c.logger.Info("Stage 4: Computing compilation order")
	compilationOrder, err := c.depGraph.GetCompilationOrder()
	if err != nil {
		// Circular dependency is fatal
		return output, fmt.Errorf("circular dependency detected: %w", err)
	}
	c.logger.Info("Compilation order computed", "order", compilationOrder)

	// Stage 5: Collect symbols from all files (first pass)
	c.logger.Info("Stage 5: Collecting symbols")
	symbolErrs := c.collectSymbols(ctx, astMap, compilationOrder)
	if len(symbolErrs) > 0 {
		output.Errors = append(output.Errors, symbolErrs...)
		// Continue - some symbols may still be available
	}
	c.logger.Info("Symbols collected")

	// Stage 6: Resolve and generate code for each file (second pass)
	c.logger.Info("Stage 6: Resolving and generating code")
	compileErrs := c.resolveAndGenerate(ctx, astMap, compilationOrder, output.CompiledFiles)
	if len(compileErrs) > 0 {
		output.Errors = append(output.Errors, compileErrs...)
	}
	c.logger.Info("Code generation complete", "files", len(output.CompiledFiles))

	if len(output.Errors) > 0 {
		return output, fmt.Errorf("compilation completed with %d errors", len(output.Errors))
	}

	return output, nil
}

// collectAllFiles expands file paths and directories to a list of .psx files
func (c *MultiFileCompiler) collectAllFiles(files []string) ([]string, error) {
	result := []string{}
	seen := make(map[string]bool)

	for _, path := range files {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path %s: %w", path, err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", absPath, err)
		}

		if info.IsDir() {
			// Walk directory to find .psx files
			err := filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(p, ".psx") {
					if !seen[p] {
						result = append(result, p)
						seen[p] = true
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking directory %s: %w", absPath, err)
			}
		} else if strings.HasSuffix(absPath, ".psx") {
			if !seen[absPath] {
				result = append(result, absPath)
				seen[absPath] = true
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no .psx files found")
	}

	return result, nil
}

// parseAllFiles parses all files to AST
func (c *MultiFileCompiler) parseAllFiles(ctx context.Context, files []string) (map[string]*ast.Module, []*CompilationError) {
	astMap := make(map[string]*ast.Module)
	errors := []*CompilationError{}

	for _, filePath := range files {
		// Read file
		content, err := os.ReadFile(filePath)
		if err != nil {
			errors = append(errors, &CompilationError{
				File:    filePath,
				Stage:   "parse",
				Message: "failed to read file",
				Details: err,
			})
			continue
		}

		// Lex
		scanner := lexer.NewScanner(content)
		tokens := scanner.ScanTokens()
		if len(scanner.Errors) > 0 {
			for _, lexErr := range scanner.Errors {
				errors = append(errors, &CompilationError{
					File:    filePath,
					Stage:   "parse",
					Message: "lexer error",
					Details: lexErr,
				})
			}
			continue
		}

		// Parse
		p := parser.NewParser(tokens)
		module, parseErrors := p.Parse()
		if len(parseErrors) > 0 {
			for _, parseErr := range parseErrors {
				errors = append(errors, &CompilationError{
					File:    filePath,
					Stage:   "parse",
					Message: "parser error",
					Details: parseErr,
				})
			}
			continue
		}

		// Add to graph
		err = c.depGraph.AddFile(filePath, module)
		if err != nil {
			errors = append(errors, &CompilationError{
				File:    filePath,
				Stage:   "parse",
				Message: "failed to add to dependency graph",
				Details: err,
			})
			continue
		}

		astMap[filePath] = module
	}

	return astMap, errors
}

// buildDependencyGraph extracts imports and builds the dependency graph
func (c *MultiFileCompiler) buildDependencyGraph(ctx context.Context, astMap map[string]*ast.Module) []*CompilationError {
	errors := []*CompilationError{}

	for filePath, module := range astMap {
		// Extract imports from AST
		imports, err := depgraph.ExtractImports(module, filePath, c.moduleResolver)
		if err != nil {
			errors = append(errors, &CompilationError{
				File:    filePath,
				Stage:   "dependency",
				Message: "failed to extract imports",
				Details: err,
			})
			continue
		}

		// Add dependencies to graph
		for _, imp := range imports {
			err := c.depGraph.AddDependency(filePath, imp.ModulePath)
			if err != nil {
				errors = append(errors, &CompilationError{
					File:    filePath,
					Stage:   "dependency",
					Message: fmt.Sprintf("failed to add dependency to %s", imp.ModulePath),
					Details: err,
				})
			}
		}

		// Store imports in file node
		node, exists := c.depGraph.GetFileNode(filePath)
		if exists {
			node.Imports = imports
		}
	}

	return errors
}

// collectSymbols collects exported symbols from all files
func (c *MultiFileCompiler) collectSymbols(ctx context.Context, astMap map[string]*ast.Module, compilationOrder []string) []*CompilationError {
	errors := []*CompilationError{}

	for _, filePath := range compilationOrder {
		module, exists := astMap[filePath]
		if !exists {
			continue
		}

		// Collect symbols from module (with dependencies for re-export support)
		collector := symbol.NewCollectorWithDeps(filePath, c.symbolRegistry, c.moduleResolver)
		symbols := collector.CollectFromModule(module)

		// Register symbols
		c.symbolRegistry.RegisterModule(filePath, symbols)
	}

	return errors
}

// resolveAndGenerate resolves and generates code for each file
func (c *MultiFileCompiler) resolveAndGenerate(
	ctx context.Context,
	astMap map[string]*ast.Module,
	compilationOrder []string,
	output map[string][]byte,
) []*CompilationError {
	errors := []*CompilationError{}

	for _, filePath := range compilationOrder {
		module, exists := astMap[filePath]
		if !exists {
			continue
		}

		// Compile this file with full import context
		code, err := c.compileFile(ctx, filePath, module)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		output[filePath] = code
	}

	return errors
}

// compileFile compiles a single file with full import context
func (c *MultiFileCompiler) compileFile(ctx context.Context, filePath string, module *ast.Module) ([]byte, *CompilationError) {
	// Create resolver with import context
	res := resolver.NewResolverWithDeps(c.moduleResolver, c.symbolRegistry, filePath)

	// Resolve
	resolutionTable, err := res.Resolve(module)
	if err != nil || (resolutionTable != nil && len(resolutionTable.Errors) > 0) {
		// Aggregate all resolution errors
		if resolutionTable != nil && len(resolutionTable.Errors) > 0 {
			var errMsg strings.Builder
			for i, resErr := range resolutionTable.Errors {
				if i > 0 {
					errMsg.WriteString("; ")
				}
				errMsg.WriteString(resErr.Error())
			}
			return nil, &CompilationError{
				File:    filePath,
				Stage:   "resolve",
				Message: fmt.Sprintf("resolution failed with %d errors", len(resolutionTable.Errors)),
				Details: fmt.Errorf("%s", errMsg.String()),
			}
		}
		return nil, &CompilationError{
			File:    filePath,
			Stage:   "resolve",
			Message: "resolution failed",
			Details: err,
		}
	}

	// Transform
	transformer := transformers.NewTransformerVisitor()
	transformedModule, err := transformer.TransformModule(module, resolutionTable)
	if err != nil {
		return nil, &CompilationError{
			File:    filePath,
			Stage:   "transform",
			Message: "transformation failed",
			Details: err,
		}
	}

	// Generate code
	generator := codegen.NewCodeGenerator()
	code := generator.Generate(transformedModule)

	return []byte(code), nil
}
