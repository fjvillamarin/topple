package module

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fjvillamarin/topple/internal/filesystem"
)

// mockFileSystem implements filesystem.FileSystem for testing
type mockFileSystem struct {
	files map[string]bool // path -> exists
}

func newMockFS(files map[string]bool) *mockFileSystem {
	return &mockFileSystem{files: files}
}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) Exists(path string) (bool, error) {
	exists, ok := m.files[path]
	if !ok {
		return false, nil
	}
	return exists, nil
}

func (m *mockFileSystem) IsDir(path string) (bool, error) {
	// Check if any file starts with this path + "/"
	prefix := path + string(filepath.Separator)
	for p := range m.files {
		if strings.HasPrefix(p, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockFileSystem) ListFiles(dir string, recursive bool) ([]string, error) {
	return nil, nil
}

func (m *mockFileSystem) ListPSXFiles(dir string, recursive bool) ([]string, error) {
	return nil, nil
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) ResolvePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (m *mockFileSystem) RelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

func (m *mockFileSystem) AbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (m *mockFileSystem) JoinPaths(elem ...string) string {
	return filepath.Join(elem...)
}

func (m *mockFileSystem) GetOutputPath(inputPath, outputDir string) (string, error) {
	return "", nil
}

func (m *mockFileSystem) WatchFiles(ctx context.Context, dirs []string, recursive bool) (<-chan filesystem.FileEvent, error) {
	return nil, nil
}

func (m *mockFileSystem) StopWatching() error {
	return nil
}

func TestResolveAbsolute(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		rootDir    string
		files      map[string]bool
		want       string
		wantErr    bool
		errType    ErrorType
	}{
		{
			name:       "simple module file",
			modulePath: "my_module",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/my_module.psx": true,
			},
			want:    "/proj/my_module.psx",
			wantErr: false,
		},
		{
			name:       "package with __init__",
			modulePath: "components",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/components/__init__.psx": true,
			},
			want:    "/proj/components/__init__.psx",
			wantErr: false,
		},
		{
			name:       "prefer file over package",
			modulePath: "utils",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/utils.psx":          true,
				"/proj/utils/__init__.psx": true,
			},
			want:    "/proj/utils.psx",
			wantErr: false,
		},
		{
			name:       "dotted module path",
			modulePath: "lib.utils",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/lib/utils.psx": true,
			},
			want:    "/proj/lib/utils.psx",
			wantErr: false,
		},
		{
			name:       "nested package",
			modulePath: "lib.subpackage",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/lib/subpackage/__init__.psx": true,
			},
			want:    "/proj/lib/subpackage/__init__.psx",
			wantErr: false,
		},
		{
			name:       "module not found",
			modulePath: "missing",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    ModuleNotFound,
		},
		{
			name:       "deeply nested module",
			modulePath: "a.b.c.d",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/a/b/c/d.psx": true,
			},
			want:    "/proj/a/b/c/d.psx",
			wantErr: false,
		},
		{
			name:       "deeply nested package",
			modulePath: "a.b.c.d",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/a/b/c/d/__init__.psx": true,
			},
			want:    "/proj/a/b/c/d/__init__.psx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFS(tt.files)
			resolver := NewResolver(Config{
				RootDir:    tt.rootDir,
				FileSystem: fs,
			})

			got, err := resolver.ResolveAbsolute(context.Background(), tt.modulePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAbsolute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if resErr, ok := err.(*ResolutionError); ok {
					if resErr.ErrorType != tt.errType {
						t.Errorf("Error type = %v, want %v", resErr.ErrorType, tt.errType)
					}
				} else {
					t.Errorf("Expected ResolutionError, got %T", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("ResolveAbsolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveRelative(t *testing.T) {
	tests := []struct {
		name       string
		dotCount   int
		modulePath string
		sourceFile string
		rootDir    string
		files      map[string]bool
		want       string
		wantErr    bool
		errType    ErrorType
	}{
		{
			name:       "single dot - current package sibling",
			dotCount:   1,
			modulePath: "sibling",
			sourceFile: "/proj/app.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/sibling.psx": true,
			},
			want:    "/proj/sibling.psx",
			wantErr: false,
		},
		{
			name:       "double dot - parent level",
			dotCount:   2,
			modulePath: "utils",
			sourceFile: "/proj/views/app.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/utils.psx": true,
			},
			want:    "/proj/utils.psx",
			wantErr: false,
		},
		{
			name:       "triple dot - grandparent level",
			dotCount:   3,
			modulePath: "config",
			sourceFile: "/proj/a/b/file.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/config.psx": true,
			},
			want:    "/proj/config.psx",
			wantErr: false,
		},
		{
			name:       "single dot with package",
			dotCount:   1,
			modulePath: "components",
			sourceFile: "/proj/views/app.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/views/components/__init__.psx": true,
			},
			want:    "/proj/views/components/__init__.psx",
			wantErr: false,
		},
		{
			name:       "current package only (no module path)",
			dotCount:   1,
			modulePath: "",
			sourceFile: "/proj/pkg/file.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/pkg/__init__.psx": true,
			},
			want:    "/proj/pkg/__init__.psx",
			wantErr: false,
		},
		{
			name:       "too many dots",
			dotCount:   10,
			modulePath: "something",
			sourceFile: "/proj/file.psx",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
		{
			name:       "zero dots (invalid)",
			dotCount:   0,
			modulePath: "module",
			sourceFile: "/proj/file.psx",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    InvalidRelativeImport,
		},
		{
			name:       "relative module not found",
			dotCount:   1,
			modulePath: "missing",
			sourceFile: "/proj/file.psx",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    ModuleNotFound,
		},
		{
			name:       "double dot with dotted path",
			dotCount:   2,
			modulePath: "lib.utils",
			sourceFile: "/proj/views/app.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/lib/utils.psx": true,
			},
			want:    "/proj/lib/utils.psx",
			wantErr: false,
		},
		{
			name:       "relative to package",
			dotCount:   1,
			modulePath: "child",
			sourceFile: "/proj/pkg/file.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/pkg/child/__init__.psx": true,
			},
			want:    "/proj/pkg/child/__init__.psx",
			wantErr: false,
		},
		// Security tests: Prevent path traversal attacks
		{
			name:       "security: escape project root - 4 dots from nested file",
			dotCount:   4,
			modulePath: "malicious",
			sourceFile: "/home/user/proj/a/b/file.psx",
			rootDir:    "/home/user/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
		{
			name:       "security: exactly at project root boundary - should succeed",
			dotCount:   3,
			modulePath: "config",
			sourceFile: "/proj/a/b/file.psx",
			rootDir:    "/proj",
			files: map[string]bool{
				"/proj/config.psx": true,
			},
			want:    "/proj/config.psx",
			wantErr: false,
		},
		{
			name:       "security: escape by one level",
			dotCount:   4,
			modulePath: "etc.passwd",
			sourceFile: "/proj/a/b/file.psx",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
		{
			name:       "security: deep nesting - escape with many dots",
			dotCount:   8,
			modulePath: "attack",
			sourceFile: "/home/user/projects/myapp/src/components/views/page.psx",
			rootDir:    "/home/user/projects/myapp",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
		{
			name:       "security: single dot above root should fail",
			dotCount:   2,
			modulePath: "outside",
			sourceFile: "/proj/file.psx",
			rootDir:    "/proj",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
		{
			name:       "security: two levels deep - three dots escapes",
			dotCount:   3,
			modulePath: "escape",
			sourceFile: "/myproject/src/file.psx",
			rootDir:    "/myproject",
			files:      map[string]bool{},
			wantErr:    true,
			errType:    TooManyDots,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFS(tt.files)
			resolver := NewResolver(Config{
				RootDir:    tt.rootDir,
				FileSystem: fs,
			})

			got, err := resolver.ResolveRelative(
				context.Background(),
				tt.dotCount,
				tt.modulePath,
				tt.sourceFile,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRelative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if resErr, ok := err.(*ResolutionError); ok {
					if resErr.ErrorType != tt.errType {
						t.Errorf("Error type = %v, want %v", resErr.ErrorType, tt.errType)
					}
				} else {
					t.Errorf("Expected ResolutionError, got %T", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("ResolveRelative() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaching(t *testing.T) {
	fs := newMockFS(map[string]bool{
		"/proj/module.psx": true,
	})

	resolver := NewResolver(Config{
		RootDir:    "/proj",
		FileSystem: fs,
	})

	// First resolution
	path1, err := resolver.ResolveAbsolute(context.Background(), "module")
	if err != nil {
		t.Fatalf("First resolution failed: %v", err)
	}

	// Second resolution (should use cache)
	path2, err := resolver.ResolveAbsolute(context.Background(), "module")
	if err != nil {
		t.Fatalf("Second resolution failed: %v", err)
	}

	if path1 != path2 {
		t.Errorf("Cache returned different paths: %v vs %v", path1, path2)
	}

	// Verify cache is working by checking the cached value directly
	if cached, ok := resolver.cache["module"]; !ok {
		t.Error("Expected module to be cached")
	} else if cached != path1 {
		t.Errorf("Cached path = %v, want %v", cached, path1)
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		files      map[string]bool
		want       bool
	}{
		{
			name:       "module exists",
			modulePath: "module",
			files: map[string]bool{
				"/proj/module.psx": true,
			},
			want: true,
		},
		{
			name:       "module does not exist",
			modulePath: "missing",
			files:      map[string]bool{},
			want:       false,
		},
		{
			name:       "package exists",
			modulePath: "pkg",
			files: map[string]bool{
				"/proj/pkg/__init__.psx": true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMockFS(tt.files)
			resolver := NewResolver(Config{
				RootDir:    "/proj",
				FileSystem: fs,
			})

			got := resolver.Exists(context.Background(), tt.modulePath)
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchPaths(t *testing.T) {
	resolver := NewResolver(Config{
		RootDir:     "/proj",
		SearchPaths: []string{"/lib1", "/lib2"},
		FileSystem:  newMockFS(map[string]bool{}),
	})

	paths := resolver.SearchPaths()

	expected := []string{"/proj", "/lib1", "/lib2"}
	if len(paths) != len(expected) {
		t.Errorf("SearchPaths() length = %v, want %v", len(paths), len(expected))
	}

	for i, path := range paths {
		if path != expected[i] {
			t.Errorf("SearchPaths()[%d] = %v, want %v", i, path, expected[i])
		}
	}
}

func TestMultipleSearchPaths(t *testing.T) {
	fs := newMockFS(map[string]bool{
		"/lib1/module_a.psx": true,
		"/lib2/module_b.psx": true,
		"/proj/module_c.psx": true,
	})

	resolver := NewResolver(Config{
		RootDir:     "/proj",
		SearchPaths: []string{"/lib1", "/lib2"},
		FileSystem:  fs,
	})

	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{
			name:       "module in root",
			modulePath: "module_c",
			want:       "/proj/module_c.psx",
		},
		{
			name:       "module in first search path",
			modulePath: "module_a",
			want:       "/lib1/module_a.psx",
		},
		{
			name:       "module in second search path",
			modulePath: "module_b",
			want:       "/lib2/module_b.psx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ResolveAbsolute(context.Background(), tt.modulePath)
			if err != nil {
				t.Errorf("ResolveAbsolute() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveAbsolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	fs := newMockFS(map[string]bool{})
	resolver := NewResolver(Config{
		RootDir:    "/proj",
		FileSystem: fs,
	})

	t.Run("module not found includes searched paths", func(t *testing.T) {
		_, err := resolver.ResolveAbsolute(context.Background(), "missing")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "cannot resolve import") {
			t.Errorf("Error message should contain 'cannot resolve import': %v", errMsg)
		}
		if !strings.Contains(errMsg, "searched:") {
			t.Errorf("Error message should contain 'searched:': %v", errMsg)
		}
	})

	t.Run("relative import error includes source file", func(t *testing.T) {
		_, err := resolver.ResolveRelative(context.Background(), 1, "missing", "/proj/app.psx")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "/proj/app.psx") {
			t.Errorf("Error message should contain source file: %v", errMsg)
		}
	})

	t.Run("too many dots error is clear", func(t *testing.T) {
		_, err := resolver.ResolveRelative(context.Background(), 100, "something", "/file.psx")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "too many dots") {
			t.Errorf("Error message should mention 'too many dots': %v", errMsg)
		}
		if !strings.Contains(errMsg, "cannot navigate above root") {
			t.Errorf("Error message should mention root directory: %v", errMsg)
		}
	})
}
