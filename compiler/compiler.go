package compiler

import (
	"context"
	"log/slog"

	"sylfie/compiler/ast"
	"sylfie/compiler/codegen"
	"sylfie/compiler/lexer"
	"sylfie/compiler/parser"
	"sylfie/compiler/resolver"
	"sylfie/compiler/transformers"
)

// File represents a file in the Biscuit compiler
type File struct {
	Name    string // Name of the file
	Content []byte // Content of the file
}

// Compiler is the main interface for the Biscuit compiler
type Compiler interface {
	// Compile takes a Biscuit source code and compiles it to Python code
	Compile(ctx context.Context, file File) ([]byte, []error)
}

// StandardCompiler is the standard implementation of the Compiler interface
type StandardCompiler struct {
	logger *slog.Logger
}

// NewCompiler creates a new StandardCompiler with default options
func NewCompiler(logger *slog.Logger) *StandardCompiler {
	if logger == nil {
		logger = slog.Default()
	}

	return &StandardCompiler{
		logger: logger,
	}
}

// Compile takes a Biscuit source code and compiles it to Python code
func (c *StandardCompiler) Compile(ctx context.Context, file File) ([]byte, []error) {
	ast, errors := Parse(file.Content)
	if len(errors) > 0 {
		return nil, errors
	}

	// Variable resolution phase
	r := resolver.NewResolver()
	resolutionTable, err := r.Resolve(ast)
	if err != nil {
		return nil, []error{err}
	}
	if len(resolutionTable.Errors) > 0 {
		return nil, resolutionTable.Errors
	}

	resolver.DebugPrintResolutionTable(resolutionTable)

	// Transformation phase with resolution information
	transformerVisitor := transformers.NewTransformerVisitor()
	ast, err = transformerVisitor.TransformModule(ast, resolutionTable)
	if err != nil {
		return nil, []error{err}
	}

	generator := codegen.NewCodeGenerator()
	result := generator.Generate(ast)

	return []byte(result), nil
}

// Parse scans a source file and returns a parsed AST.
// It returns both the AST and a slice of any errors encountered during scanning and parsing.
func Parse(src []byte) (*ast.Module, []error) {
	scanner := lexer.NewScanner(src)
	tokens := scanner.ScanTokens()

	if len(scanner.Errors) > 0 {
		return nil, scanner.Errors
	}

	parser := parser.NewParser(tokens)
	program, errors := parser.Parse()

	if len(errors) > 0 {
		return nil, errors
	}

	return program, nil
}
