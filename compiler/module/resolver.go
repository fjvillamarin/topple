package module

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/fjvillamarin/topple/internal/filesystem"
)

// Resolver translates import paths to filesystem paths
type Resolver interface {
	// ResolveAbsolute resolves an absolute import path
	// Example: "my_module" -> "/path/to/my_module.psx"
	ResolveAbsolute(ctx context.Context, modulePath string) (string, error)

	// ResolveRelative resolves a relative import from a source file
	// Example: dotCount=1, modulePath="sibling", sourceFile="/proj/app.psx"
	//   -> "/proj/sibling.psx"
	ResolveRelative(ctx context.Context, dotCount int, modulePath string, sourceFile string) (string, error)

	// Exists checks if a module exists
	Exists(ctx context.Context, modulePath string) bool

	// SearchPaths returns the configured search paths
	SearchPaths() []string
}

// Config holds configuration for module resolution
type Config struct {
	// RootDir is the base directory for resolution (usually cwd or project root)
	RootDir string

	// SearchPaths are additional directories to search (for future use)
	SearchPaths []string

	// FileSystem abstraction for testing
	FileSystem filesystem.FileSystem
}

// StandardResolver implements Resolver
type StandardResolver struct {
	config Config
	cache  map[string]string // Import path -> resolved file path
}

// NewResolver creates a new StandardResolver
func NewResolver(config Config) *StandardResolver {
	return &StandardResolver{
		config: config,
		cache:  make(map[string]string),
	}
}

// ResolveAbsolute resolves an absolute import path to a file
func (r *StandardResolver) ResolveAbsolute(ctx context.Context, modulePath string) (string, error) {
	// Check cache first
	if cached, ok := r.cache[modulePath]; ok {
		return cached, nil
	}

	// Build search paths: root dir first, then additional search paths
	searchPaths := []string{r.config.RootDir}
	searchPaths = append(searchPaths, r.config.SearchPaths...)

	var attemptedPaths []string

	// Try each search path
	for _, searchPath := range searchPaths {
		// Convert dotted path to filesystem path
		// "my.module" -> "my/module"
		fsPath := strings.ReplaceAll(modulePath, ".", string(filepath.Separator))

		// Try as file: searchPath/my/module.psx
		filePath := r.config.FileSystem.JoinPaths(searchPath, fsPath+".psx")
		absFilePath, err := r.config.FileSystem.AbsolutePath(filePath)
		if err == nil {
			attemptedPaths = append(attemptedPaths, absFilePath)
			exists, _ := r.config.FileSystem.Exists(absFilePath)
			if exists {
				r.cache[modulePath] = absFilePath
				return absFilePath, nil
			}
		}

		// Try as package: searchPath/my/module/__init__.psx
		pkgPath := r.config.FileSystem.JoinPaths(searchPath, fsPath, "__init__.psx")
		absPkgPath, err := r.config.FileSystem.AbsolutePath(pkgPath)
		if err == nil {
			attemptedPaths = append(attemptedPaths, absPkgPath)
			exists, _ := r.config.FileSystem.Exists(absPkgPath)
			if exists {
				r.cache[modulePath] = absPkgPath
				return absPkgPath, nil
			}
		}
	}

	return "", newModuleNotFoundError(modulePath, "", attemptedPaths)
}

// ResolveRelative resolves a relative import from a source file
func (r *StandardResolver) ResolveRelative(ctx context.Context, dotCount int, modulePath string, sourceFile string) (string, error) {
	if dotCount == 0 {
		return "", newInvalidRelativeImportError(
			modulePath,
			sourceFile,
			"dotCount must be > 0 for relative imports",
		)
	}

	// Get absolute path of source file
	absSourceFile, err := r.config.FileSystem.AbsolutePath(sourceFile)
	if err != nil {
		return "", err
	}

	// Get source directory
	sourceDir := filepath.Dir(absSourceFile)

	// Navigate up by (dotCount - 1) levels
	// dotCount=1 (.) means current directory
	// dotCount=2 (..) means parent directory
	targetDir := sourceDir
	for i := 1; i < dotCount; i++ {
		parent := filepath.Dir(targetDir)
		if parent == targetDir {
			// Reached filesystem root
			return "", newTooManyDotsError(
				strings.Repeat(".", dotCount)+modulePath,
				sourceFile,
			)
		}
		targetDir = parent
	}

	// Security check: Ensure targetDir is within project root
	// This prevents path traversal attacks where imports escape the configured RootDir
	absRootDir, err := r.config.FileSystem.AbsolutePath(r.config.RootDir)
	if err != nil {
		return "", err
	}

	absTargetDir, err := r.config.FileSystem.AbsolutePath(targetDir)
	if err != nil {
		return "", err
	}

	// Check if target directory escaped the project root
	relPath, err := r.config.FileSystem.RelativePath(absRootDir, absTargetDir)
	if err != nil || strings.HasPrefix(relPath, "..") {
		// Target directory is outside the project root
		return "", newTooManyDotsError(
			strings.Repeat(".", dotCount)+modulePath,
			sourceFile,
		)
	}

	// If no modulePath, we're importing from a package itself
	// e.g., "from . import x" means import from current package's __init__.psx
	if modulePath == "" {
		initPath := r.config.FileSystem.JoinPaths(targetDir, "__init__.psx")
		absInitPath, err := r.config.FileSystem.AbsolutePath(initPath)
		if err != nil {
			return "", err
		}

		exists, _ := r.config.FileSystem.Exists(absInitPath)
		if exists {
			return absInitPath, nil
		}

		return "", newModuleNotFoundError(
			strings.Repeat(".", dotCount),
			sourceFile,
			[]string{absInitPath},
		)
	}

	// Convert dotted module path to filesystem path
	fsPath := strings.ReplaceAll(modulePath, ".", string(filepath.Separator))

	var attemptedPaths []string

	// Try as file: targetDir/module.psx
	filePath := r.config.FileSystem.JoinPaths(targetDir, fsPath+".psx")
	absFilePath, err := r.config.FileSystem.AbsolutePath(filePath)
	if err == nil {
		attemptedPaths = append(attemptedPaths, absFilePath)
		exists, _ := r.config.FileSystem.Exists(absFilePath)
		if exists {
			return absFilePath, nil
		}
	}

	// Try as package: targetDir/module/__init__.psx
	pkgPath := r.config.FileSystem.JoinPaths(targetDir, fsPath, "__init__.psx")
	absPkgPath, err := r.config.FileSystem.AbsolutePath(pkgPath)
	if err == nil {
		attemptedPaths = append(attemptedPaths, absPkgPath)
		exists, _ := r.config.FileSystem.Exists(absPkgPath)
		if exists {
			return absPkgPath, nil
		}
	}

	importPath := strings.Repeat(".", dotCount) + modulePath
	return "", newModuleNotFoundError(importPath, sourceFile, attemptedPaths)
}

// Exists checks if a module exists (without caching)
func (r *StandardResolver) Exists(ctx context.Context, modulePath string) bool {
	_, err := r.ResolveAbsolute(ctx, modulePath)
	return err == nil
}

// SearchPaths returns the configured search paths
func (r *StandardResolver) SearchPaths() []string {
	paths := []string{r.config.RootDir}
	paths = append(paths, r.config.SearchPaths...)
	return paths
}
