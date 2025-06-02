package compiler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	testCategories := []string{
		"basic",
		"views", 
		"control_flow",
		"composition",
		"slots",
		"attributes",
		"expressions",
		"python_integration",
		"htmx",
		"fastapi",
		"errors",
	}
	
	updateGolden := os.Getenv("UPDATE_GOLDEN") == "1"
	
	for _, category := range testCategories {
		t.Run(category, func(t *testing.T) {
			inputDir := filepath.Join("testdata/input", category)
			expectedDir := filepath.Join("testdata/expected", category)
			
			// Skip if category directory doesn't exist
			if _, err := os.Stat(inputDir); os.IsNotExist(err) {
				t.Skipf("Category directory %s not found", inputDir)
				return
			}
			
			// Find all .psx files in category directory
			entries, err := os.ReadDir(inputDir)
			if err != nil {
				t.Fatalf("Failed to read input directory %s: %v", inputDir, err)
			}
			
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".psx") {
					continue
				}
				
				testName := strings.TrimSuffix(entry.Name(), ".psx")
				t.Run(testName, func(t *testing.T) {
					inputFile := filepath.Join(inputDir, entry.Name())
					expectedFile := filepath.Join(expectedDir, testName+".py")
					generatedDir := filepath.Join("testdata/generated", category)
					generatedFile := filepath.Join(generatedDir, testName+".py")
					
					// Compile the input file
					result := compileFile(inputFile)
					
					// Always write generated output to testdata/generated for comparison
					if err := os.MkdirAll(generatedDir, 0755); err != nil {
						t.Fatalf("Failed to create generated directory: %v", err)
					}
					err := os.WriteFile(generatedFile, []byte(result), 0644)
					if err != nil {
						t.Fatalf("Failed to write generated file: %v", err)
					}
					
					if updateGolden {
						// Ensure expected directory exists
						if err := os.MkdirAll(expectedDir, 0755); err != nil {
							t.Fatalf("Failed to create expected directory: %v", err)
						}
						
						// Update golden file
						err := os.WriteFile(expectedFile, []byte(result), 0644)
						if err != nil {
							t.Fatalf("Failed to write golden file: %v", err)
						}
						t.Logf("Updated golden file: %s", expectedFile)
						t.Logf("Generated file: %s", generatedFile)
						return
					}
					
					// Compare with expected output
					expected, err := os.ReadFile(expectedFile)
					if err != nil {
						if os.IsNotExist(err) {
							// Create golden file if it doesn't exist
							if err := os.MkdirAll(expectedDir, 0755); err != nil {
								t.Fatalf("Failed to create expected directory: %v", err)
							}
							err := os.WriteFile(expectedFile, []byte(result), 0644)
							if err != nil {
								t.Fatalf("Failed to create golden file: %v", err)
							}
							t.Logf("Created golden file: %s", expectedFile)
							t.Logf("Generated file: %s", generatedFile)
							return
						}
						t.Fatalf("Failed to read expected file: %v", err)
					}
					
					if result != string(expected) {
						t.Errorf("Output differs from expected:\nExpected file: %s\nGenerated file: %s\n\nDifferences:\nGot:\n%s\nExpected:\n%s", expectedFile, generatedFile, result, string(expected))
					} else {
						t.Logf("✅ Test passed - Generated file: %s", generatedFile)
					}
				})
			}
		})
	}
}

func compileFile(inputPath string) string {
	// Read the input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Sprintf("ERROR: Failed to read file: %v", err)
	}

	// Use the actual compiler with all the proper steps
	compiler := NewCompiler(slog.Default())
	file := File{
		Name:    inputPath,
		Content: content,
	}
	
	result, errors := compiler.Compile(context.Background(), file)
	if len(errors) > 0 {
		return fmt.Sprintf("COMPILATION_ERRORS: %v", errors)
	}
	
	return string(result)
}