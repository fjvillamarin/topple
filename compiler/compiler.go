package compiler

import (
	"context"
	"log/slog"
)

// File represents a file in the Biscuit compiler
type File struct {
	Name    string // Name of the file
	Content []byte // Content of the file
}

// Compiler is the main interface for the Biscuit compiler
type Compiler interface {
	// Compile takes a Biscuit source code and compiles it to Python code
	Compile(ctx context.Context, file File) ([]byte, error)
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
func (c *StandardCompiler) Compile(ctx context.Context, file File) ([]byte, error) {
	return nil, nil
}
