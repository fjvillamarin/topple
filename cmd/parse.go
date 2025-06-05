package main

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"sylfie/compiler"
	"sylfie/internal/filesystem"
)

// ParseCmd defines the "parse" command which runs the scanner and parser
// and prints the AST or writes it to a file.
type ParseCmd struct {
	// Positional arguments – kept identical to CompileCmd for convenience
	Input  string `arg:"" required:"" help:"Path to a Python/Biscuit file or directory"`
	Output string `arg:"" optional:"" help:"Output directory for AST files (default: none)"`

	// Whether to write output files
	WriteAST bool `help:"Write AST to .ast files" short:"w" default:"false"`
}

// Run executes the parse command.
func (p *ParseCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	log.InfoContext(*ctx, "Running parse command")

	// Initialise helpers
	fs := filesystem.NewFileSystem(log)

	// Validate input path exists
	exists, err := fs.Exists(p.Input)
	if err != nil {
		return fmt.Errorf("error checking input path: %w", err)
	}
	if !exists {
		return fmt.Errorf("input path does not exist: %s", p.Input)
	}

	isDir, err := fs.IsDir(p.Input)
	if err != nil {
		return fmt.Errorf("error determining if input is directory: %w", err)
	}

	start := time.Now()

	if isDir {
		// Use the specialized method to get Biscuit (.psx) files
		sources, err := fs.ListBiscuitFiles(p.Input, globals.Recursive)
		if err != nil {
			return fmt.Errorf("error listing biscuit files: %w", err)
		}

		log.InfoContext(*ctx, "Parsing files in directory", slog.Int("fileCount", len(sources)))
		for _, file := range sources {
			if err := parseFile(fs, file, p.Output, p.WriteAST, log, *ctx); err != nil {
				return err
			}
		}
	} else {
		// Single file
		if err := parseFile(fs, p.Input, p.Output, p.WriteAST, log, *ctx); err != nil {
			return err
		}
	}

	log.InfoContext(*ctx, "Parsing completed", slog.Duration("elapsed", time.Since(start)))
	return nil
}

// parseFile runs the parser on a single file, prints AST to console,
// and optionally writes AST to a .ast file
func parseFile(fs filesystem.FileSystem, path, outputDir string, writeAST bool, log *slog.Logger, ctx context.Context) error {
	log.DebugContext(ctx, "Parsing file", slog.String("file", path))

	content, err := fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	program, errors := compiler.Parse(content)

	// Format AST into a string
	filename := filepath.Base(path)
	var output strings.Builder

	output.WriteString(fmt.Sprintf("=== %s ===\n\n", filename))

	if program != nil {
		printer := compiler.NewASTPrinter("  ")
		output.WriteString(printer.Print(program))
		output.WriteString("\n")
	}

	if len(errors) > 0 {
		output.WriteString(fmt.Sprintf("\n-- Errors (%d) --\n", len(errors)))
		for i, e := range errors {
			output.WriteString(fmt.Sprintf("%d: %v\n", i+1, e))
		}
	}

	if !writeAST {
		// Print to console if not writing to file
		fmt.Println()
		fmt.Print(output.String())
	}

	// Write to file if requested
	if writeAST {
		outputPath := getASTOutputPath(fs, path, outputDir)
		if err := fs.WriteFile(outputPath, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("error writing AST file: %w", err)
		}
		log.InfoContext(ctx, "Wrote AST file",
			slog.String("input", path),
			slog.String("output", outputPath))
	}

	return nil
}

// getASTOutputPath determines the output path for an AST file
func getASTOutputPath(fs filesystem.FileSystem, inputPath, outputDir string) string {
	var outputPath string

	if outputDir == "" {
		// Use same directory as input
		inputDir := filepath.Dir(inputPath)
		baseName := filepath.Base(inputPath)
		// Replace existing extension with .ast
		astName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".ast"
		outputPath = filepath.Join(inputDir, astName)
	} else {
		// Ensure output directory exists
		fs.MkdirAll(outputDir, 0755) // Ignore error, will fail on write if needed

		// Create .ast file in output directory
		baseName := filepath.Base(inputPath)
		astName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".ast"
		outputPath = filepath.Join(outputDir, astName)
	}

	return outputPath
}
