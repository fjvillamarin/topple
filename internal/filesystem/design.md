# Filesystem Service Design

## Overview

The filesystem service provides a unified interface for file operations needed by the Biscuit compiler and watch services. It abstracts away the details of file I/O, watching for changes, and path resolution to simplify the rest of the application code.

## Core Interfaces

### `FileSystem` Interface

```go
type FileSystem interface {
    // File Operations
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Exists(path string) (bool, error)
    IsDir(path string) (bool, error)
    
    // Directory Operations
    ListFiles(dir string, recursive bool) ([]string, error)
    ListBiscuitFiles(dir string, recursive bool) ([]string, error)
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
```

### `FileEvent` Type

```go
type EventType int

const (
    Create EventType = iota
    Modify
    Delete
    Rename
)

type FileEvent struct {
    Path      string    // Path to the affected file
    Type      EventType // Type of event
    Timestamp time.Time // When the event occurred
}
```

## Implementation Details

### Standard Filesystem Implementation

```go
type StandardFileSystem struct {
    watcher   *fsnotify.Watcher
    eventChan chan FileEvent
    mu        sync.Mutex
    watching  bool
}

func NewFileSystem() FileSystem {
    return &StandardFileSystem{}
}
```

### Mocked Filesystem for Testing

```go
type MockFileSystem struct {
    Files     map[string][]byte
    Dirs      map[string]bool
    Events    chan FileEvent
    WatchDirs []string
    Recursive bool
}

func NewMockFileSystem() *MockFileSystem {
    return &MockFileSystem{
        Files:  make(map[string][]byte),
        Dirs:   make(map[string]bool),
        Events: make(chan FileEvent),
    }
}
```

## Key Implementation Features

### Path Resolution and Manipulation

- Handle absolute and relative paths consistently
- Convert between paths in different formats
- Resolve symlinks when necessary
- Generate output paths for compiled files

### File Operations with Error Handling

- Abstract error handling for common file operations
- Provide clear error messages for common failures
- Implement retries for transient failures

### Watching File System Changes

- Use fsnotify for efficient file system monitoring
- Implement debouncing to prevent multiple rapid triggers
- Filter events to focus on .bsct files and related files
- Support recursive watching of directories
- Clean up watchers when they're no longer needed

### Biscuit File Discovery

- Find all .bsct files in a directory tree
- Sort files in a consistent order for predictable compilation
- Filter out files that should be excluded

## Usage Examples

### Compiling Files

```go
func CompileDirectory(fs filesystem.FileSystem, inputDir, outputDir string, recursive bool) error {
    // List all biscuit files
    biscuitFiles, err := fs.ListBiscuitFiles(inputDir, recursive)
    if err != nil {
        return err
    }
    
    // Process each file
    for _, file := range biscuitFiles {
        data, err := fs.ReadFile(file)
        if err != nil {
            return err
        }
        
        // Compile the file content
        compiled := compileContent(data)
        
        // Determine output path
        outputPath, err := fs.GetOutputPath(file, outputDir)
        if err != nil {
            return err
        }
        
        // Create parent directories if needed
        parentDir := filepath.Dir(outputPath)
        if err := fs.MkdirAll(parentDir, 0755); err != nil {
            return err
        }
        
        // Write compiled content
        if err := fs.WriteFile(outputPath, compiled, 0644); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Watching for Changes

```go
func WatchAndCompile(ctx context.Context, fs filesystem.FileSystem, 
                     inputDir, outputDir string, recursive bool, delay time.Duration) error {
    // Start watching the directory
    eventChan, err := fs.WatchFiles(ctx, []string{inputDir}, recursive)
    if err != nil {
        return err
    }
    
    // Create a debounce timer
    timer := time.NewTimer(delay)
    timer.Stop()
    
    // Process events
    for {
        select {
        case <-ctx.Done():
            return fs.StopWatching()
            
        case event := <-eventChan:
            // Check if it's a biscuit file
            if !strings.HasSuffix(event.Path, ".bsct") {
                continue
            }
            
            // Reset debounce timer
            timer.Reset(delay)
            
        case <-timer.C:
            // Compile the directory after delay
            if err := CompileDirectory(fs, inputDir, outputDir, recursive); err != nil {
                log.Printf("Compilation error: %v", err)
            }
        }
    }
}
```

## Extension Points

- Support for virtual file systems (e.g., embedded files, remote files)
- Configuration for file encoding and line endings
- Handling of include/import paths for compilation
- Custom file filtering for the watcher
- Progress reporting for long-running operations
- Integration with version control systems

## Testing Strategy

- Unit tests for individual functions
- Integration tests for file operations
- Mocked file system for testing compilation logic
- Simulated file events for testing the watcher
- Benchmark tests for performance-critical operations