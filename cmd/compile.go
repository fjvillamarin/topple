package main

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"topple/compiler"
	"topple/compiler/codegen"
	"topple/compiler/lexer"
	"topple/compiler/resolver"
	"topple/compiler/transformers"
	"topple/internal/filesystem"
)

// emitSet tracks which intermediate artifacts to emit during compilation.
type emitSet struct {
	Tokens         bool
	AST            bool
	Resolution     bool
	TransformedAST bool
}

// any returns true if any emit flag is set.
func (e emitSet) any() bool {
	return e.Tokens || e.AST || e.Resolution || e.TransformedAST
}

// parseEmit parses a comma-separated emit string into an emitSet.
// Valid values: tokens, ast, resolution, transformed-ast, all.
func parseEmit(raw string) (emitSet, error) {
	if raw == "" {
		return emitSet{}, nil
	}

	var es emitSet
	for _, part := range strings.Split(raw, ",") {
		switch strings.TrimSpace(part) {
		case "tokens":
			es.Tokens = true
		case "ast":
			es.AST = true
		case "resolution":
			es.Resolution = true
		case "transformed-ast":
			es.TransformedAST = true
		case "all":
			return emitSet{Tokens: true, AST: true, Resolution: true, TransformedAST: true}, nil
		default:
			return emitSet{}, fmt.Errorf("unknown emit value %q (valid: tokens, ast, resolution, transformed-ast, all)", part)
		}
	}
	return es, nil
}

// CompileCmd defines the "compile" command.
// It includes file inputs, output options, options to choose Python version,
// and various customization flags.
type CompileCmd struct {
	// Positional arguments
	Input  string `arg:"" required:"" help:"Path to a PSX file or directory"`
	Output string `arg:"" optional:"" help:"Output directory for compiled Python files (default: same as input)"`

	// Flags
	Emit string `help:"Emit intermediate artifacts (comma-separated: tokens,ast,resolution,transformed-ast,all)" short:"e" default:""`
}

func (c *CompileCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	// Typically you'd load your configuration file using CLI.Config here,
	// merge it with flag values, and create a final config struct.
	log.InfoContext(*ctx, "Running compile command with options")
	log.InfoContext(*ctx, "Debug mode", slog.Bool("enabled", globals.Debug))
	log.InfoContext(*ctx, "Recursive mode", slog.Bool("enabled", globals.Recursive))
	log.InfoContext(*ctx, "Input path", slog.String("path", c.Input))
	log.InfoContext(*ctx, "Output path", slog.String("path", c.Output))

	// Parse emit flags
	emit, err := parseEmit(c.Emit)
	if err != nil {
		return err
	}

	// Default behavior: if no output directory is provided, we'll output .py files in the same directory as the input files
	if c.Output == "" {
		log.DebugContext(*ctx, "No output path specified, will create .py files in the same directory as input files")
	} else {
		log.DebugContext(*ctx, "Output path specified", slog.String("path", c.Output))
	}

	// Initialize the filesystem service
	fs := filesystem.NewFileSystem(log)

	// Initialize the compiler service
	cmp := compiler.NewCompiler(log)

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

		// List all PSX files
		files, err := fs.ListPSXFiles(c.Input, globals.Recursive)
		if err != nil {
			return fmt.Errorf("error listing PSX files: %w", err)
		}

		log.InfoContext(*ctx, "Found PSX files", slog.Int("count", len(files)))

		for _, file := range files {
			if err := compileFile(fs, cmp, file, c.Output, emit, log, *ctx); err != nil {
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

		if err := compileFile(fs, cmp, c.Input, c.Output, emit, log, *ctx); err != nil {
			return err
		}
	}

	elapsed := time.Since(startTime)
	log.InfoContext(*ctx, "Compilation completed", slog.Duration("elapsed", elapsed))

	return nil
}

// compileFile compiles a single PSX file to a Python file.
// When emit flags are set, it runs the pipeline step-by-step and writes intermediate artifacts.
func compileFile(fs filesystem.FileSystem, cmp compiler.Compiler, inputPath, outputDir string, emit emitSet, log *slog.Logger, ctx context.Context) error {
	log.DebugContext(ctx, "Compiling file", slog.String("input", inputPath))

	// Read the input file
	content, err := fs.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}

	// Get the output path for the .py file
	outputPath, err := fs.GetOutputPath(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("error determining output path: %w", err)
	}

	// Ensure the output directory exists
	outputDirPath := filepath.Dir(outputPath)
	if err := fs.MkdirAll(outputDirPath, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Fast path: no emit flags, use the compiler directly
	if !emit.any() {
		file := compiler.File{
			Name:    filepath.Base(inputPath),
			Content: content,
		}
		pythonCode, errors := cmp.Compile(ctx, file)
		if len(errors) > 0 {
			for _, err := range errors {
				log.ErrorContext(ctx, "Error compiling file", slog.String("error", err.Error()))
			}
			return fmt.Errorf("error compiling file: %d errors", len(errors))
		}

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

	// Emit path: run pipeline step-by-step
	return compileFileWithEmit(fs, content, inputPath, outputDir, outputPath, emit, log, ctx)
}

// compileFileWithEmit runs the compilation pipeline step-by-step,
// writing intermediate artifacts at each stage based on emit flags.
func compileFileWithEmit(fs filesystem.FileSystem, content []byte, inputPath, outputDir, outputPath string, emit emitSet, log *slog.Logger, ctx context.Context) error {
	filename := filepath.Base(inputPath)

	// Step 1: Scan
	tokens, errors := compiler.Scan(content)
	if len(errors) > 0 {
		for _, err := range errors {
			log.ErrorContext(ctx, "Scan error", slog.String("error", err.Error()))
		}
		return fmt.Errorf("error scanning file: %d errors", len(errors))
	}

	if emit.Tokens {
		tokPath := getEmitOutputPath(inputPath, outputDir, ".tok")
		tokOutput := formatTokens(tokens, filename)
		if err := fs.WriteFile(tokPath, []byte(tokOutput), 0644); err != nil {
			return fmt.Errorf("error writing token file: %w", err)
		}
		log.InfoContext(ctx, "Wrote token file", slog.String("output", tokPath))
	}

	// Step 2: Parse
	module, errors := compiler.ParseTokens(tokens)
	if len(errors) > 0 {
		for _, err := range errors {
			log.ErrorContext(ctx, "Parse error", slog.String("error", err.Error()))
		}
		return fmt.Errorf("error parsing file: %d errors", len(errors))
	}

	if emit.AST {
		astPath := getEmitOutputPath(inputPath, outputDir, ".ast")
		printer := compiler.NewASTPrinter("  ")
		astOutput := fmt.Sprintf("=== %s ===\n\n%s\n", filename, printer.Print(module))
		if err := fs.WriteFile(astPath, []byte(astOutput), 0644); err != nil {
			return fmt.Errorf("error writing AST file: %w", err)
		}
		log.InfoContext(ctx, "Wrote AST file", slog.String("output", astPath))
	}

	// Step 3: Resolve
	r := resolver.NewResolver()
	resolutionTable, err := r.Resolve(module)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}
	if len(resolutionTable.Errors) > 0 {
		for _, err := range resolutionTable.Errors {
			log.ErrorContext(ctx, "Resolution error", slog.String("error", err.Error()))
		}
		return fmt.Errorf("error resolving file: %d errors", len(resolutionTable.Errors))
	}

	if emit.Resolution {
		resPath := getEmitOutputPath(inputPath, outputDir, ".res")
		if err := resolver.WriteResolutionText(resolutionTable, filename, resPath); err != nil {
			return fmt.Errorf("error writing resolution text file: %w", err)
		}
		log.InfoContext(ctx, "Wrote resolution text file", slog.String("output", resPath))

		jsonPath := getEmitOutputPath(inputPath, outputDir, ".res.json")
		if err := resolver.WriteResolutionJSON(resolutionTable, filename, jsonPath); err != nil {
			return fmt.Errorf("error writing resolution JSON file: %w", err)
		}
		log.InfoContext(ctx, "Wrote resolution JSON file", slog.String("output", jsonPath))
	}

	// Step 4: Transform
	transformerVisitor := transformers.NewTransformerVisitor()
	module, err = transformerVisitor.TransformModule(module, resolutionTable)
	if err != nil {
		return fmt.Errorf("error transforming file: %w", err)
	}

	if emit.TransformedAST {
		tastPath := getEmitOutputPath(inputPath, outputDir, ".tast")
		printer := compiler.NewASTPrinter("  ")
		tastOutput := fmt.Sprintf("=== %s (transformed) ===\n\n%s\n", filename, printer.Print(module))
		if err := fs.WriteFile(tastPath, []byte(tastOutput), 0644); err != nil {
			return fmt.Errorf("error writing transformed AST file: %w", err)
		}
		log.InfoContext(ctx, "Wrote transformed AST file", slog.String("output", tastPath))
	}

	// Step 5: Codegen (always)
	generator := codegen.NewCodeGenerator()
	result := generator.Generate(module)

	if err := fs.WriteFile(outputPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	log.InfoContext(ctx, "Compiled file",
		slog.String("input", inputPath),
		slog.String("output", outputPath),
		slog.Int("inputSize", len(content)),
		slog.Int("outputSize", len(result)))

	return nil
}

// getEmitOutputPath determines the output path for an intermediate artifact file.
func getEmitOutputPath(inputPath, outputDir, ext string) string {
	baseName := filepath.Base(inputPath)
	name := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ext

	if outputDir == "" {
		return filepath.Join(filepath.Dir(inputPath), name)
	}
	return filepath.Join(outputDir, name)
}

// formatTokens formats a token slice into the same text format used by the scan command.
func formatTokens(tokens []lexer.Token, filename string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n\n", filename))
	for i, tok := range tokens {
		sb.WriteString(fmt.Sprintf("%d: %s %d %q %v @ %s\n",
			i, tok.Type, int(tok.Type), tok.Lexeme, tok.Literal, tok.Span.String()))
	}
	return sb.String()
}
