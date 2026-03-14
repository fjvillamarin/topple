package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string // returns startDir
		wantDir  string                    // relative to tmpDir root, or "" for not found
		wantErr  bool
		absolute bool // if true, wantDir is the exact expected value
	}{
		{
			name: "marker in start directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte{}, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantDir: ".", // same as startDir
		},
		{
			name: "marker in parent directory",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte{}, 0644); err != nil {
					t.Fatal(err)
				}
				sub := filepath.Join(root, "sub")
				if err := os.MkdirAll(sub, 0755); err != nil {
					t.Fatal(err)
				}
				return sub
			},
			wantDir: "..", // parent of startDir
		},
		{
			name: "marker in grandparent directory",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte{}, 0644); err != nil {
					t.Fatal(err)
				}
				sub := filepath.Join(root, "a", "b")
				if err := os.MkdirAll(sub, 0755); err != nil {
					t.Fatal(err)
				}
				return sub
			},
			wantDir: "../..", // grandparent of startDir
		},
		{
			name: "no marker found",
			setup: func(t *testing.T) string {
				return t.TempDir() // no pyproject.toml anywhere
			},
			wantDir:  "",
			absolute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDir := tt.setup(t)

			got, err := FindProjectRoot(startDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FindProjectRoot() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.absolute {
				if got != tt.wantDir {
					t.Errorf("FindProjectRoot() = %q, want %q", got, tt.wantDir)
				}
				return
			}

			// Resolve expected path relative to startDir
			want, err := filepath.Abs(filepath.Join(startDir, tt.wantDir))
			if err != nil {
				t.Fatal(err)
			}

			if got != want {
				t.Errorf("FindProjectRoot() = %q, want %q", got, want)
			}
		})
	}
}
