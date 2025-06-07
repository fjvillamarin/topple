package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// EventType represents the type of filesystem event
type EventType int

const (
	Create EventType = iota
	Modify
	Delete
	Rename
)

// String returns the string representation of the event type
func (e EventType) String() string {
	switch e {
	case Create:
		return "CREATE"
	case Modify:
		return "MODIFY"
	case Delete:
		return "DELETE"
	case Rename:
		return "RENAME"
	default:
		return "UNKNOWN"
	}
}

// FileEvent represents a filesystem event
type FileEvent struct {
	Path      string    // Path to the affected file
	Type      EventType // Type of event
	Timestamp time.Time // When the event occurred
}

// FileSystem provides an interface for filesystem operations
type FileSystem interface {
	// File Operations
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Exists(path string) (bool, error)
	IsDir(path string) (bool, error)

	// Directory Operations
	ListFiles(dir string, recursive bool) ([]string, error)
	ListPSXFiles(dir string, recursive bool) ([]string, error)
	MkdirAll(path string, perm os.FileMode) error

	// Path Operations
	ResolvePath(path string) (string, error)
	RelativePath(base, target string) (string, error)
	AbsolutePath(path string) (string, error)
	JoinPaths(elem ...string) string

	// Path Transformations for Compilation
	GetOutputPath(inputPath, outputDir string) (string, error)

	// Watching File Changes
	WatchFiles(ctx context.Context, dirs []string, recursive bool) (<-chan FileEvent, error)
	StopWatching() error
}

// StandardFileSystem implements FileSystem using the standard library
type StandardFileSystem struct {
	watcher   *fsnotify.Watcher
	eventChan chan FileEvent
	mu        sync.Mutex
	watching  bool
	done      chan struct{}
	logger    *slog.Logger
}

// NewFileSystem creates a new FileSystem instance with a logger
func NewFileSystem(logger *slog.Logger) FileSystem {
	if logger == nil {
		logger = slog.Default()
	}

	return &StandardFileSystem{
		eventChan: make(chan FileEvent),
		done:      make(chan struct{}),
		logger:    logger,
	}
}

// ReadFile reads a file's contents
func (s *StandardFileSystem) ReadFile(path string) ([]byte, error) {
	s.logger.Debug("Reading file", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.Error("Failed to read file", "path", path, "error", err)
		return nil, err
	}
	s.logger.Debug("Successfully read file", "path", path, "size", len(data))
	return data, nil
}

// WriteFile writes data to a file
func (s *StandardFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	s.logger.Debug("Writing file", "path", path, "size", len(data), "permission", perm)

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.logger.Error("Failed to create directory", "directory", dir, "error", err)
		return err
	}

	err := os.WriteFile(path, data, perm)
	if err != nil {
		s.logger.Error("Failed to write file", "path", path, "error", err)
		return err
	}

	s.logger.Debug("Successfully wrote file", "path", path, "size", len(data))
	return nil
}

// Exists checks if a file or directory exists
func (s *StandardFileSystem) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	s.logger.Error("Failed to check existence", "path", path, "error", err)
	return false, err
}

// IsDir checks if a path is a directory
func (s *StandardFileSystem) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Debug("Path does not exist", "path", path)
		} else {
			s.logger.Error("Failed to check if path is directory", "path", path, "error", err)
		}
		return false, err
	}
	return info.IsDir(), nil
}

// ListFiles lists all files in a directory
func (s *StandardFileSystem) ListFiles(dir string, recursive bool) ([]string, error) {
	s.logger.Debug("Listing files", "directory", dir, "recursive", recursive)

	var files []string

	isDir, err := s.IsDir(dir)
	if err != nil {
		return nil, err
	}

	if !isDir {
		s.logger.Debug("Path is not a directory, returning as single file", "path", dir)
		return []string{dir}, nil
	}

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("Error while walking directory", "path", path, "error", err)
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		} else if path != dir && !recursive {
			s.logger.Debug("Skipping subdirectory in non-recursive mode", "path", path)
			return filepath.SkipDir
		}

		return nil
	}

	if err := filepath.Walk(dir, walkFunc); err != nil {
		s.logger.Error("Failed to walk directory", "directory", dir, "error", err)
		return nil, err
	}

	s.logger.Debug("Found files", "directory", dir, "count", len(files))
	return files, nil
}

// ListPSXFiles lists all .psx files in a directory
func (s *StandardFileSystem) ListPSXFiles(dir string, recursive bool) ([]string, error) {
	s.logger.Debug("Listing PSX files", "directory", dir, "recursive", recursive)

	files, err := s.ListFiles(dir, recursive)
	if err != nil {
		return nil, err
	}

	var psxFiles []string
	for _, file := range files {
		if strings.HasSuffix(file, ".psx") {
			psxFiles = append(psxFiles, file)
		}
	}

	s.logger.Debug("Found PSX files", "directory", dir, "count", len(psxFiles))
	return psxFiles, nil
}

// MkdirAll creates a directory and all necessary parent directories
func (s *StandardFileSystem) MkdirAll(path string, perm os.FileMode) error {
	s.logger.Debug("Creating directory", "path", path, "permission", perm)
	err := os.MkdirAll(path, perm)
	if err != nil {
		s.logger.Error("Failed to create directory", "path", path, "error", err)
	}
	return err
}

// ResolvePath resolves a path to its absolute form, handling symlinks
func (s *StandardFileSystem) ResolvePath(path string) (string, error) {
	s.logger.Debug("Resolving path", "path", path)

	absPath, err := filepath.Abs(path)
	if err != nil {
		s.logger.Error("Failed to get absolute path", "path", path, "error", err)
		return "", err
	}

	// Resolve symlinks if they exist
	evalPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If path doesn't exist yet, just return the absolute path
		if os.IsNotExist(err) {
			s.logger.Debug("Path does not exist, returning absolute path", "path", absPath)
			return absPath, nil
		}
		s.logger.Error("Failed to evaluate symlinks", "path", absPath, "error", err)
		return "", err
	}

	s.logger.Debug("Resolved path", "original", path, "resolved", evalPath)
	return evalPath, nil
}

// RelativePath returns a relative path from base to target
func (s *StandardFileSystem) RelativePath(base, target string) (string, error) {
	s.logger.Debug("Getting relative path", "base", base, "target", target)
	relPath, err := filepath.Rel(base, target)
	if err != nil {
		s.logger.Error("Failed to get relative path", "base", base, "target", target, "error", err)
	}
	return relPath, err
}

// AbsolutePath returns the absolute path for a given path
func (s *StandardFileSystem) AbsolutePath(path string) (string, error) {
	s.logger.Debug("Getting absolute path", "path", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		s.logger.Error("Failed to get absolute path", "path", path, "error", err)
	}
	return absPath, err
}

// JoinPaths joins path elements
func (s *StandardFileSystem) JoinPaths(elem ...string) string {
	joined := filepath.Join(elem...)
	s.logger.Debug("Joined paths", "elements", elem, "result", joined)
	return joined
}

// GetOutputPath transforms an input path to an output path
func (s *StandardFileSystem) GetOutputPath(inputPath, outputDir string) (string, error) {
	s.logger.Debug("Getting output path", "input", inputPath, "outputDir", outputDir)

	// Get the absolute path for input
	absInputPath, err := s.AbsolutePath(inputPath)
	if err != nil {
		return "", err
	}

	// Check if input is a PSX file
	if !strings.HasSuffix(absInputPath, ".psx") {
		s.logger.Error("Input file must be a .psx file", "path", inputPath)
		return "", fmt.Errorf("input file must be a .psx file: %s", inputPath)
	}

	// If no output directory is specified or output directory is empty,
	// write .py file to same directory as input file
	if outputDir == "" {
		inputDir := filepath.Dir(absInputPath)
		baseName := filepath.Base(absInputPath)
		pyName := strings.TrimSuffix(baseName, ".psx") + ".py"
		outputPath := filepath.Join(inputDir, pyName)
		s.logger.Debug("No output dir specified, writing to input directory",
			"input", absInputPath,
			"output", outputPath)
		return outputPath, nil
	}

	// Get the absolute path for output
	absOutputDir, err := s.AbsolutePath(outputDir)
	if err != nil {
		return "", err
	}

	// Determine if outputDir is a file or directory
	isDir, err := s.IsDir(absOutputDir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// If output doesn't exist, check if it ends with .py
	if os.IsNotExist(err) {
		if strings.HasSuffix(absOutputDir, ".py") {
			// Output is a direct file path
			s.logger.Debug("Output is a Python file path", "path", absOutputDir)
			return absOutputDir, nil
		}
		// Otherwise, treat as a directory
		s.logger.Debug("Output directory doesn't exist, will be created", "path", absOutputDir)
		isDir = true
	}

	if isDir {
		// Replace .psx extension with .py
		baseName := filepath.Base(absInputPath)
		pyName := strings.TrimSuffix(baseName, ".psx") + ".py"
		outputPath := filepath.Join(absOutputDir, pyName)
		s.logger.Debug("Generated output path", "input", absInputPath, "output", outputPath)
		return outputPath, nil
	}

	// Output is a specific file
	s.logger.Debug("Using direct output file path", "path", absOutputDir)
	return absOutputDir, nil
}

// WatchFiles watches directories for file changes
func (s *StandardFileSystem) WatchFiles(ctx context.Context, dirs []string, recursive bool) (<-chan FileEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Starting file watcher", "directories", dirs, "recursive", recursive)

	if s.watching {
		s.logger.Error("Already watching directories")
		return nil, fmt.Errorf("already watching directories")
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.logger.Error("Failed to create watcher", "error", err)
		return nil, err
	}
	s.watcher = watcher

	// Initialize event channels
	s.eventChan = make(chan FileEvent)
	s.done = make(chan struct{})
	s.watching = true

	// Add directories to watch
	for _, dir := range dirs {
		if recursive {
			s.logger.Debug("Adding directory recursively", "directory", dir)
			err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					s.logger.Error("Error walking directory", "path", path, "error", err)
					return err
				}
				if d.IsDir() {
					if err := watcher.Add(path); err != nil {
						s.logger.Error("Failed to add directory to watcher", "path", path, "error", err)
						return err
					}
					s.logger.Debug("Added directory to watcher", "path", path)
				}
				return nil
			})
			if err != nil {
				s.logger.Error("Failed to walk directory", "directory", dir, "error", err)
				s.StopWatching()
				return nil, err
			}
		} else {
			s.logger.Debug("Adding directory (non-recursive)", "directory", dir)
			if err := watcher.Add(dir); err != nil {
				s.logger.Error("Failed to add directory to watcher", "directory", dir, "error", err)
				s.StopWatching()
				return nil, err
			}
		}
	}

	// Start goroutine to handle events
	go func() {
		defer close(s.eventChan)
		s.logger.Debug("Starting watcher event loop")

		for {
			select {
			case <-s.done:
				s.logger.Debug("Watcher stopped by done signal")
				return
			case <-ctx.Done():
				s.logger.Debug("Watcher stopped by context cancellation")
				s.StopWatching()
				return
			case err, ok := <-watcher.Errors:
				if !ok {
					s.logger.Debug("Watcher error channel closed")
					return
				}
				s.logger.Error("Watcher error", "error", err)
			case event, ok := <-watcher.Events:
				if !ok {
					s.logger.Debug("Watcher event channel closed")
					return
				}

				// Convert fsnotify event to FileEvent
				fileEvent := FileEvent{
					Path:      event.Name,
					Timestamp: time.Now(),
				}

				// Determine event type
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					fileEvent.Type = Create
					s.logger.Debug("File created", "path", event.Name)

					// If a new directory is created and we're watching recursively,
					// add it to the watcher
					if recursive {
						info, err := os.Stat(event.Name)
						if err == nil && info.IsDir() {
							s.logger.Debug("Adding new directory to watcher", "path", event.Name)
							err := watcher.Add(event.Name)
							if err != nil {
								s.logger.Error("Failed to add new directory to watcher", "path", event.Name, "error", err)
							}
						}
					}
				case event.Op&fsnotify.Write == fsnotify.Write:
					fileEvent.Type = Modify
					s.logger.Debug("File modified", "path", event.Name)
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					fileEvent.Type = Delete
					s.logger.Debug("File deleted", "path", event.Name)
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					fileEvent.Type = Rename
					s.logger.Debug("File renamed", "path", event.Name)
				default:
					// Skip unhandled event types
					s.logger.Debug("Skipping unhandled event type", "operation", event.Op, "path", event.Name)
					continue
				}

				// Send the event to the channel
				select {
				case s.eventChan <- fileEvent:
					s.logger.Debug("Sent file event", "type", fileEvent.Type, "path", fileEvent.Path)
				case <-s.done:
					s.logger.Debug("Watcher stopped by done signal while sending event")
					return
				case <-ctx.Done():
					s.logger.Debug("Watcher stopped by context cancellation while sending event")
					s.StopWatching()
					return
				}
			}
		}
	}()

	s.logger.Info("File watcher started", "directories", dirs)
	return s.eventChan, nil
}

// StopWatching stops watching for file changes
func (s *StandardFileSystem) StopWatching() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.watching {
		s.logger.Debug("Not watching, nothing to stop")
		return nil
	}

	s.logger.Info("Stopping file watcher")

	// Signal the goroutine to stop
	close(s.done)

	// Close the watcher
	if s.watcher != nil {
		err := s.watcher.Close()
		if err != nil {
			s.logger.Error("Error closing watcher", "error", err)
		}
		s.watcher = nil
		s.watching = false
		return err
	}

	s.watching = false
	s.logger.Debug("File watcher stopped")
	return nil
}
