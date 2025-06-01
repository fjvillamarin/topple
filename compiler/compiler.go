package compiler

import (
	"context"
	"log/slog"

	"biscuit/compiler/ast"
	"biscuit/compiler/codegen"
	"biscuit/compiler/codegen/transformers"
	"biscuit/compiler/lexer"
	"biscuit/compiler/parser"
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

	transformerVisitor := transformers.NewTransformerVisitor(transformers.NewViewTransformer())
	ast, err := transformerVisitor.TransformModule(ast)
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

	transformerVisitor := transformers.NewTransformerVisitor(transformers.NewViewTransformer())
	program, err := transformerVisitor.TransformModule(program)
	if err != nil {
		return nil, []error{err}
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return program, nil
}
