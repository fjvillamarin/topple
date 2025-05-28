package main

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"biscuit/compiler"
	"biscuit/internal/filesystem"
)

// CompileCmd defines the "compile" command.
// It includes file inputs, output options, options to choose Python version,
// and various customization flags.
type CompileCmd struct {
	// Positional arguments
	Input  string `arg:"" required:"" help:"Path to a Biscuit file or directory"`
	Output string `arg:"" optional:"" help:"Output directory for compiled Python files (default: same as input)"`
}

func (c *CompileCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	// Typically you'd load your configuration file using CLI.Config here,
	// merge it with flag values, and create a final config struct.
	log.InfoContext(*ctx, "Running compile command with options")
	log.InfoContext(*ctx, "Debug mode", slog.Bool("enabled", globals.Debug))
	log.InfoContext(*ctx, "Recursive mode", slog.Bool("enabled", globals.Recursive))
	log.InfoContext(*ctx, "Input path", slog.String("path", c.Input))
	log.InfoContext(*ctx, "Output path", slog.String("path", c.Output))

	// Default behavior: if no output directory is provided, we'll output .py files in the same directory as the input files
	if c.Output == "" {
		log.DebugContext(*ctx, "No output path specified, will create .py files in the same directory as input files")
	} else {
		log.DebugContext(*ctx, "Output path specified", slog.String("path", c.Output))
	}

	// Initialize the filesystem service
	fs := filesystem.NewFileSystem(log)

	// Initialize the compiler service
	compiler := compiler.NewCompiler(log)

	// Check if input exists
	exists, err := fs.Exists(c.Input)
	if err != nil {
		return fmt.Errorf("error checking input path: %w", err)
	}
	if !exists {
		return fmt.Errorf("input path does not exist: %s", c.Input)
	}

	// Check if input is a directory or a file
	isDir, err := fs.IsDir(c.Input)
	if err != nil {
		return fmt.Errorf("error checking if input is a directory: %w", err)
	}

	startTime := time.Now()
	log.InfoContext(*ctx, "Starting compilation")

	if isDir {
		// Process directory
		log.DebugContext(*ctx, "Input is a directory", slog.String("path", c.Input))

		// List all biscuit files
		files, err := fs.ListBiscuitFiles(c.Input, globals.Recursive)
		if err != nil {
			return fmt.Errorf("error listing biscuit files: %w", err)
		}

		log.InfoContext(*ctx, "Found biscuit files", slog.Int("count", len(files)))

		for _, file := range files {
			if err := compileFile(fs, compiler, file, c.Output, log, *ctx); err != nil {
				return err
			}
		}
	} else {
		// Process single file
		log.DebugContext(*ctx, "Input is a file", slog.String("path", c.Input))

		// Verify it's a .psx file
		if filepath.Ext(c.Input) != ".psx" {
			return fmt.Errorf("input file is not a .psx file: %s", c.Input)
		}

		if err := compileFile(fs, compiler, c.Input, c.Output, log, *ctx); err != nil {
			return err
		}
	}

	elapsed := time.Since(startTime)
	log.InfoContext(*ctx, "Compilation completed", slog.Duration("elapsed", elapsed))

	return nil
}

// compileFile compiles a single biscuit file to a Python file
func compileFile(fs filesystem.FileSystem, cmp compiler.Compiler, inputPath, outputDir string, log *slog.Logger, ctx context.Context) error {
	log.DebugContext(ctx, "Compiling file", slog.String("input", inputPath))

	// Read the input file
	content, err := fs.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}

	// Get the output path
	outputPath, err := fs.GetOutputPath(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("error determining output path: %w", err)
	}

	// Ensure the output directory exists
	outputDirPath := filepath.Dir(outputPath)
	if err := fs.MkdirAll(outputDirPath, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// For now, we'll just write a placeholder Python file
	file := compiler.File{
		Name:    filepath.Base(inputPath),
		Content: content,
	}
	pythonCode, errors := cmp.Compile(ctx, file)
	if len(errors) > 0 {
		return fmt.Errorf("error compiling file: %d errors", len(errors))
	}

	// Write the output file
	if err := fs.WriteFile(outputPath, pythonCode, 0644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	log.InfoContext(ctx, "Compiled file",
		slog.String("input", inputPath),
		slog.String("output", outputPath),
		slog.Int("inputSize", len(content)),
		slog.Int("outputSize", len(pythonCode)))

	return nil
}
