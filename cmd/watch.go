package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fjvillamarin/topple/compiler"
	"github.com/fjvillamarin/topple/internal/filesystem"
)

// WatchCmd defines the "watch" command.
type WatchCmd struct {
	// Positional argument
	Directory string `arg:"" required:"" help:"Directory to watch for changes"`

	// Options (shared with compile)
	Delay int  `help:"Debounce delay in milliseconds" default:"300"`
	Clear bool `help:"Clear terminal on each compilation" default:"false"`

	// Options for output
	Output string `help:"Output directory for compiled Python files (default: same as input)" default:""`
}

func (w *WatchCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	// We'll only use the Output directory if explicitly set
	// Otherwise files will be created in the same directory as source files

	// Configure logging
	log.InfoContext(*ctx, "Watching directory",
		slog.String("directory", w.Directory),
		slog.Bool("recursive", globals.Recursive),
		slog.Int("delay", w.Delay),
		slog.String("output", w.Output))

	// Initialize filesystem service
	fs := filesystem.NewFileSystem(log)

	// Initialize the compiler service
	cmp := compiler.NewCompiler(log)

	// Check if directory exists
	exists, err := fs.Exists(w.Directory)
	if err != nil {
		return fmt.Errorf("error checking directory: %w", err)
	}
	if !exists {
		return fmt.Errorf("directory does not exist: %s", w.Directory)
	}

	// Check if it's actually a directory
	isDir, err := fs.IsDir(w.Directory)
	if err != nil {
		return fmt.Errorf("error checking if path is a directory: %w", err)
	}
	if !isDir {
		return fmt.Errorf("path is not a directory: %s", w.Directory)
	}

	// Initial compilation
	log.InfoContext(*ctx, "Performing initial compilation")
	if err := compileDirectory(fs, cmp, w.Directory, w.Output, globals.Recursive, log, *ctx); err != nil {
		return fmt.Errorf("initial compilation failed: %w", err)
	}

	// Start watching
	log.InfoContext(*ctx, "Starting file watcher")

	// Create a cancellable context for the watcher
	watchCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start watching the directory
	events, err := fs.WatchFiles(watchCtx, []string{w.Directory}, globals.Recursive)
	if err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	// Create a timer for debouncing
	timer := time.NewTimer(time.Duration(w.Delay) * time.Millisecond)
	timer.Stop()

	// Track when we need to recompile
	needsRecompile := false

	// Print watching message
	fmt.Printf("Watching '%s' for changes...\n", w.Directory)

	// Watch loop

	for {
		select {
		case <-(*ctx).Done():
			// Context was cancelled (Ctrl+C or similar)
			log.InfoContext(*ctx, "Stopping watch due to context cancellation")
			return nil

		case event, ok := <-events:
			if !ok {
				// Channel was closed
				log.InfoContext(*ctx, "Event channel closed, stopping watch")
				return nil
			}

			log.DebugContext(*ctx, "File change detected",
				slog.String("path", event.Path),
				slog.String("event", event.Type.String()),
				slog.Time("timestamp", event.Timestamp))

			// Check if this is a .psx file or a Python file generated from a .psx file
			if !isPSXRelatedFile(event.Path) {
				log.DebugContext(*ctx, "Ignoring non-PSX file", slog.String("path", event.Path))
				continue
			}

			// Reset debounce timer
			timer.Reset(time.Duration(w.Delay) * time.Millisecond)
			needsRecompile = true

		case <-timer.C:
			if needsRecompile {
				// Clear terminal if requested
				if w.Clear {
					clearTerminal()
				}

				// Recompile
				log.InfoContext(*ctx, "Recompiling after file changes")
				if err := compileDirectory(fs, cmp, w.Directory, w.Output, globals.Recursive, log, *ctx); err != nil {
					log.ErrorContext(*ctx, "Compilation failed", slog.String("error", err.Error()))
					fmt.Printf("Compilation error: %v\n", err)
				} else {
					log.InfoContext(*ctx, "Compilation successful")
					fmt.Println("Compilation successful")
				}

				needsRecompile = false
			}
		}
	}
}

// compileDirectory compiles all PSX files in a directory
func compileDirectory(fs filesystem.FileSystem, cmp compiler.Compiler, inputDir, outputDir string, recursive bool, log *slog.Logger, ctx context.Context) error {
	// List all PSX files
	files, err := fs.ListPSXFiles(inputDir, recursive)
	if err != nil {
		return fmt.Errorf("error listing PSX files: %w", err)
	}

	log.InfoContext(ctx, "Found PSX files to compile", slog.Int("count", len(files)))

	// Track compilation time
	startTime := time.Now()

	// Compile each file
	for _, file := range files {
		log.DebugContext(ctx, "Compiling file", slog.String("file", file))
		if err := compileFile(fs, cmp, file, outputDir, log, ctx); err != nil {
			return fmt.Errorf("error compiling %s: %w", file, err)
		}
	}

	elapsed := time.Since(startTime)
	log.InfoContext(ctx, "Directory compilation completed",
		slog.Duration("elapsed", elapsed),
		slog.Int("fileCount", len(files)))

	return nil
}

// clearTerminal clears the terminal screen
func clearTerminal() {
	switch runtime := os.Getenv("TERM"); runtime {
	case "linux", "xterm", "xterm-256color", "screen":
		// ANSI escape sequence to clear screen and move cursor to top-left
		fmt.Print("\033[H\033[2J")
	default:
		// For windows or unknown terminals, print newlines
		fmt.Print("\n\n\n\n\n")
	}
}

// isPSXRelatedFile checks if a file is a .psx file or a .py file that was generated from a .psx file
func isPSXRelatedFile(path string) bool {
	ext := filepath.Ext(path)
	if ext == ".psx" {
		return true
	}
	// if ext == ".py" {
	// 	// Check if there's a corresponding .psx file
	// 	psxFile := strings.TrimSuffix(path, ".py") + ".psx"
	// 	if _, err := os.Stat(psxFile); err == nil {
	// 		return true
	// 	}
	// }
	return false
}
