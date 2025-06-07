package main

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"sylfie/compiler/lexer"
	"sylfie/internal/filesystem"
)

// ScanCmd defines the "scan" command which runs just the scanner
// and prints the tokens or writes them to a file.
type ScanCmd struct {
	// Positional arguments – kept identical to CompileCmd for convenience
	Input  string `arg:"" required:"" help:"Path to a Python/PSX file or directory"`
	Output string `arg:"" optional:"" help:"Output directory for token files (default: none)"`

	// Whether to write output files
	WriteTokens bool `help:"Write tokens to .tok files" short:"w" default:"false"`
}

// Run executes the scan command.
func (s *ScanCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	log.InfoContext(*ctx, "Running scan command")

	// Initialise helpers
	fs := filesystem.NewFileSystem(log)

	// Validate input path exists
	exists, err := fs.Exists(s.Input)
	if err != nil {
		return fmt.Errorf("error checking input path: %w", err)
	}
	if !exists {
		return fmt.Errorf("input path does not exist: %s", s.Input)
	}

	isDir, err := fs.IsDir(s.Input)
	if err != nil {
		return fmt.Errorf("error determining if input is directory: %w", err)
	}

	start := time.Now()

	if isDir {
		// Include both PSX (.psx) and Python (.py) source files so the
		// scanner can be run against either.
		files, err := fs.ListFiles(s.Input, globals.Recursive)
		if err != nil {
			return fmt.Errorf("error listing files: %w", err)
		}

		var sources []string
		psxFiles := make(map[string]bool) // Track .psx files to avoid .py duplicates

		// First pass: collect all .psx files
		for _, f := range files {
			ext := filepath.Ext(f)
			if ext == ".psx" {
				sources = append(sources, f)
				// Store base name without extension to track .psx files
				baseName := strings.TrimSuffix(filepath.Base(f), ext)
				dirPath := filepath.Dir(f)
				psxFiles[filepath.Join(dirPath, baseName)] = true
			}
		}

		// Second pass: collect .py files only if no corresponding .psx exists
		for _, f := range files {
			ext := filepath.Ext(f)
			if ext == ".py" {
				baseName := strings.TrimSuffix(filepath.Base(f), ext)
				dirPath := filepath.Dir(f)
				baseKey := filepath.Join(dirPath, baseName)

				// Only include .py file if no corresponding .psx file exists
				if !psxFiles[baseKey] {
					sources = append(sources, f)
				}
			}
		}

		log.InfoContext(*ctx, "Scanning files in directory", slog.Int("fileCount", len(sources)))
		for _, file := range sources {
			if err := scanFile(fs, file, s.Output, s.WriteTokens, log, *ctx); err != nil {
				return err
			}
		}
	} else {
		// Single file
		if err := scanFile(fs, s.Input, s.Output, s.WriteTokens, log, *ctx); err != nil {
			return err
		}
	}

	log.InfoContext(*ctx, "Scanning completed", slog.Duration("elapsed", time.Since(start)))
	return nil
}

// scanFile runs the scanner on a single file, prints tokens to console,
// and optionally writes tokens to a .tok file
func scanFile(fs filesystem.FileSystem, path, outputDir string, writeTokens bool, log *slog.Logger, ctx context.Context) error {
	log.DebugContext(ctx, "Scanning file", slog.String("file", path))

	content, err := fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	scanner := lexer.NewScanner(content)
	tokens := scanner.ScanTokens()

	// Format tokens into a string
	filename := filepath.Base(path)
	var output strings.Builder

	output.WriteString(fmt.Sprintf("=== %s ===\n\n", filename))

	for i, tok := range tokens {
		output.WriteString(fmt.Sprintf("%d: %s %d %q %v @ %s\n",
			i, tok.Type, int(tok.Type), tok.Lexeme, tok.Literal, tok.Span.String()))
	}

	if len(scanner.Errors) > 0 {
		output.WriteString(fmt.Sprintf("\n-- Errors (%d) --\n", len(scanner.Errors)))
		for i, e := range scanner.Errors {
			output.WriteString(fmt.Sprintf("%d: %v\n", i+1, e))
		}
	}

	if !writeTokens {
		// Print to console if not writing to file
		fmt.Println()
		fmt.Print(output.String())
	}

	// Write to file if requested
	if writeTokens {
		outputPath := getTokenOutputPath(fs, path, outputDir)
		if err := fs.WriteFile(outputPath, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("error writing token file: %w", err)
		}
		log.InfoContext(ctx, "Wrote token file",
			slog.String("input", path),
			slog.String("output", outputPath))
	}

	return nil
}

// getTokenOutputPath determines the output path for a token file
func getTokenOutputPath(fs filesystem.FileSystem, inputPath, outputDir string) string {
	var outputPath string

	if outputDir == "" {
		// Use same directory as input
		inputDir := filepath.Dir(inputPath)
		baseName := filepath.Base(inputPath)
		// Replace existing extension with .tok
		tokName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tok"
		outputPath = filepath.Join(inputDir, tokName)
	} else {
		// Ensure output directory exists
		fs.MkdirAll(outputDir, 0755) // Ignore error, will fail on write if needed

		// Create .tok file in output directory
		baseName := filepath.Base(inputPath)
		tokName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tok"
		outputPath = filepath.Join(outputDir, tokName)
	}

	return outputPath
}
