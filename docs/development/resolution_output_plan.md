# Resolution Output Implementation Plan

## Overview

This document outlines the plan to add resolution output capabilities to the Topple compiler, enabling developers to inspect variable resolution, scope analysis, and symbol tables for debugging and tooling purposes.

## Goals

1. **JSON Output Format** (`.res.json`) - Machine-readable resolution data for tooling
2. **Human-Readable Output** (`.res`) - Text-based summary for quick inspection
3. **Parse Command Integration** - Add resolution output to `topple parse` command
4. **Foundation for Future Tooling** - Enable web visualization and LSP integration

## Output Format Design

### 1. JSON Format (`.res.json`)

The JSON format should be comprehensive and machine-readable, containing all resolution information needed for tooling.

```json
{
  "file": "example.psx",
  "version": "1.0",
  "scopes": [
    {
      "id": 0,
      "type": "module",
      "span": {
        "start": {"line": 1, "column": 0, "offset": 0},
        "end": {"line": 10, "column": 0, "offset": 200}
      },
      "parent_id": null,
      "bindings": [
        {
          "name": "HelloWorld",
          "variable_id": "var_1",
          "span": {
            "start": {"line": 1, "column": 5, "offset": 5},
            "end": {"line": 1, "column": 15, "offset": 15}
          },
          "shadowed_by": null
        }
      ],
      "globals": [],
      "nonlocals": []
    },
    {
      "id": 1,
      "type": "view",
      "span": {...},
      "parent_id": 0,
      "bindings": [
        {
          "name": "name",
          "variable_id": "var_2",
          "span": {...},
          "shadowed_by": null
        }
      ],
      "globals": [],
      "nonlocals": []
    }
  ],
  "variables": [
    {
      "id": "var_1",
      "name": "HelloWorld",
      "definition_depth": 0,
      "state": "defined",
      "classification": {
        "is_parameter": false,
        "is_global": false,
        "is_nonlocal": false,
        "is_imported": false,
        "is_view_parameter": false,
        "is_exception_var": false
      },
      "usage": {
        "is_used": true,
        "is_captured": false,
        "is_cell": false,
        "used_before_def": false
      },
      "spans": {
        "first_def": {...},
        "first_use": {...}
      },
      "references": [
        {
          "span": {...},
          "scope_id": 0,
          "context": "definition"
        }
      ]
    },
    {
      "id": "var_2",
      "name": "name",
      "definition_depth": 1,
      "state": "defined",
      "classification": {
        "is_parameter": true,
        "is_global": false,
        "is_nonlocal": false,
        "is_imported": false,
        "is_view_parameter": true,
        "is_exception_var": false
      },
      "usage": {
        "is_used": true,
        "is_captured": false,
        "is_cell": false,
        "used_before_def": false
      },
      "spans": {
        "first_def": {...},
        "first_use": {...}
      },
      "references": [...]
    }
  ],
  "imports": [],
  "views": {
    "defined": [
      {
        "name": "HelloWorld",
        "scope_id": 0,
        "span": {...}
      }
    ],
    "references": []
  },
  "closure_analysis": {
    "cell_vars": [],
    "free_vars": []
  },
  "diagnostics": {
    "errors": [],
    "warnings": []
  },
  "summary": {
    "total_scopes": 2,
    "total_variables": 2,
    "total_references": 3,
    "parameters": 1,
    "view_parameters": 1,
    "globals": 0,
    "nonlocals": 0,
    "cell_vars": 0,
    "free_vars": 0
  }
}
```

**Key Design Principles:**
- **Unique IDs**: Each scope and variable has a unique ID for cross-referencing
- **Complete Spans**: All source locations include line, column, and offset
- **Flat Structure**: Top-level arrays for scopes and variables (easier to query)
- **References**: Each variable tracks all its references with context
- **Hierarchical Scopes**: Parent pointers enable scope chain traversal
- **Rich Metadata**: All flags, states, and classifications explicitly represented

### 2. Human-Readable Format (`.res`)

The text format should be concise and easy to read, suitable for quick inspection.

```
=== example.psx ===

SCOPES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Scope #0 [module] (1:0 - 10:0)
  Bindings:
    HelloWorld    view        defined    (1:5)

Scope #1 [view] (1:0 - 5:0) → parent: #0
  Bindings:
    name          parameter   defined    (1:16)  [view-param, used]

VARIABLES:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

NAME              TYPE            STATE       DEPTH    FLAGS               REFS
────────────────────────────────────────────────────────────────────────────────
HelloWorld        local           defined     0        used                1
name              view-param      defined     1        used                2

VIEW COMPOSITION:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Defined Views:
  • HelloWorld (1:5)

View References:
  (none)

CLOSURE ANALYSIS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Cell Variables: (none)
Free Variables: (none)

DIAGNOSTICS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Errors: 0
Warnings: 0

SUMMARY:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Scopes:           2
Total Variables:        2
Total References:       3
Parameters:             1
View Parameters:        1
Global Declarations:    0
Nonlocal Declarations:  0
Cell Variables:         0
Free Variables:         0
```

**Key Design Principles:**
- **Structured Sections**: Clear section headers with visual separation
- **Hierarchical Display**: Scopes shown with parent relationships
- **Compact Tables**: Variable information in aligned columns
- **Summary Stats**: Quick overview of resolution results
- **Error Highlighting**: Diagnostics prominently displayed

## Implementation Plan

### Phase 1: JSON Serialization (Core Infrastructure)

**Goal**: Add JSON marshaling support to resolution data structures

**Files to Modify/Create:**
1. `compiler/resolver/json.go` - New file for JSON serialization
2. `compiler/resolver/types.go` - Add JSON tags to structs

**Implementation Steps:**

1. **Add JSON Tags to Core Types**
   ```go
   // In types.go
   type Variable struct {
       Name            string        `json:"name"`
       DefinitionDepth int           `json:"definition_depth"`
       State           VariableState `json:"state"`
       // ... add JSON tags to all fields
   }
   ```

2. **Create Serialization Types**
   ```go
   // In json.go
   type JSONResolution struct {
       File            string          `json:"file"`
       Version         string          `json:"version"`
       Scopes          []JSONScope     `json:"scopes"`
       Variables       []JSONVariable  `json:"variables"`
       Imports         []JSONImport    `json:"imports"`
       Views           JSONViews       `json:"views"`
       ClosureAnalysis JSONClosure     `json:"closure_analysis"`
       Diagnostics     JSONDiagnostics `json:"diagnostics"`
       Summary         JSONSummary     `json:"summary"`
   }
   ```

3. **Implement Conversion Function**
   ```go
   func (rt *ResolutionTable) ToJSON(filename string) (*JSONResolution, error)
   ```

4. **Handle Special Cases**
   - AST node pointers (convert to spans only)
   - Variable/Binding circular references (use IDs)
   - Lexer.Span conversion to JSON-friendly format

**Testing:**
- Unit tests for JSON serialization
- Round-trip testing (serialize → deserialize → verify)
- Test with various PSX files (simple, complex, nested scopes, closures)

### Phase 2: Text Output Formatter

**Goal**: Generate human-readable text output from resolution table

**Files to Create:**
1. `compiler/resolver/text_output.go` - Text formatting functions

**Implementation Steps:**

1. **Reuse Existing Debug Code**
   - Refactor `debug.go` to use a writer interface
   - Extract formatting functions for reuse

2. **Create Text Output Function**
   ```go
   func (rt *ResolutionTable) ToText(filename string) (string, error)
   ```

3. **Implement Section Formatters**
   - `formatScopesSection()`
   - `formatVariablesSection()`
   - `formatViewCompositionSection()`
   - `formatClosureSection()`
   - `formatDiagnosticsSection()`
   - `formatSummarySection()`

**Testing:**
- Output consistency tests
- Visual inspection with sample files
- Compare with debug output for accuracy

### Phase 3: Parse Command Integration

**Goal**: Add resolution output to `topple parse` command

**Files to Modify:**
1. `cmd/parse.go` - Add resolution logic and flags
2. `.mise.toml` - Update parse task if needed

**Command Interface:**
```bash
# Parse with resolution output (default: both formats)
topple parse file.psx -w

# This generates:
# - file.ast      (existing AST output)
# - file.res      (new: human-readable resolution)
# - file.res.json (new: JSON resolution)

# Parse with only JSON output
topple parse file.psx -w --format=json

# Parse with only text output
topple parse file.psx -w --format=text

# Parse directory recursively
topple parse examples/ -r -w
```

**Implementation Steps:**

1. **Add Resolution to Parse Command**
   ```go
   // In cmd/parse.go
   func runParse(cmd *cobra.Command, args []string) error {
       // ... existing parsing logic ...

       // Run resolution
       resolver := resolver.NewResolver()
       resTable, err := resolver.Resolve(module)

       // Generate outputs based on flags
       if writeFiles {
           if format == "json" || format == "all" {
               writeResolutionJSON(resTable, outputPath)
           }
           if format == "text" || format == "all" {
               writeResolutionText(resTable, outputPath)
           }
       } else {
           // Print to console
           printResolutionToConsole(resTable)
       }
   }
   ```

2. **Add Flags**
   - `--format` - Output format (text, json, all)
   - Reuse existing `-w` flag for writing to files
   - Reuse existing `-r` flag for recursive processing

3. **File Output Logic**
   ```go
   func writeResolutionJSON(table *ResolutionTable, basePath string) error {
       jsonData, err := table.ToJSON(basePath)
       if err != nil {
           return err
       }

       outputPath := strings.TrimSuffix(basePath, ".psx") + ".res.json"
       data, err := json.MarshalIndent(jsonData, "", "  ")
       if err != nil {
           return err
       }

       return os.WriteFile(outputPath, data, 0644)
   }
   ```

**Testing:**
- Test with single files
- Test with directories (recursive)
- Test with different format flags
- Test error handling (invalid files, parse errors)

### Phase 4: Error and Warning Enhancements

**Goal**: Improve diagnostic reporting in resolution output

**Implementation Steps:**

1. **Structured Error Type**
   ```go
   type ResolutionDiagnostic struct {
       Severity string      `json:"severity"` // "error" or "warning"
       Message  string      `json:"message"`
       Span     lexer.Span  `json:"span"`
       Code     string      `json:"code"` // e.g., "undefined-var", "invalid-global"
   }
   ```

2. **Add Warning Detection**
   - Unused variables
   - Variables used before definition (in same scope)
   - Shadowed variables
   - Unused imports

3. **Update ResolutionTable**
   ```go
   type ResolutionTable struct {
       // ... existing fields ...
       Diagnostics []ResolutionDiagnostic `json:"diagnostics"`
   }
   ```

**Testing:**
- Test error cases (undefined vars, invalid global/nonlocal)
- Test warning cases (unused vars, shadowing)
- Verify diagnostic output in both JSON and text formats

## Testing Strategy

### Unit Tests

1. **JSON Serialization Tests** (`compiler/resolver/json_test.go`)
   - Test all data structure serialization
   - Test special cases (empty scopes, complex closures)
   - Test error handling

2. **Text Output Tests** (`compiler/resolver/text_output_test.go`)
   - Test section formatting
   - Test alignment and layout
   - Test edge cases (long names, deep nesting)

3. **Integration Tests** (`cmd/parse_test.go`)
   - Test command with various flags
   - Test file writing
   - Test recursive directory processing

### Golden File Tests

Create golden file tests for resolution output:
```
compiler/testdata/resolution/
  input/
    simple.psx
    closures.psx
    views.psx
  expected/
    simple.res
    simple.res.json
    closures.res
    closures.res.json
    views.res
    views.res.json
```

Test command:
```bash
mise run test-resolution
mise run test-resolution-update  # Update golden files
```

### Manual Testing

Test with real PSX files:
```bash
# Test with examples
topple parse examples/homepage/views.psx -w

# Verify outputs
cat examples/homepage/views.res
cat examples/homepage/views.res.json | jq .
```

## Implementation Order

1. **Week 1**: JSON Serialization Infrastructure
   - Add JSON tags to types
   - Create serialization types
   - Implement `ToJSON()` method
   - Write unit tests

2. **Week 2**: Text Output Formatter
   - Refactor debug code
   - Implement `ToText()` method
   - Write unit tests
   - Create golden files

3. **Week 3**: Parse Command Integration
   - Add flags and logic to parse command
   - Implement file writing
   - Test with various files
   - Update documentation

4. **Week 4**: Polish and Documentation
   - Add warning detection
   - Improve error messages
   - Write user documentation
   - Create examples

## Future Enhancements

### Web Visualization Tool

Build a web-based tool to visualize resolution data:
- Interactive scope tree navigation
- Variable reference highlighting
- Closure relationship visualization
- Search and filter capabilities

**Technology Stack:**
- Frontend: React + TailwindCSS
- Data: Load `.res.json` files
- Deployment: Static site (GitHub Pages)

### LSP Integration

Use resolution data for Language Server Protocol features:
- Go to definition
- Find references
- Symbol search
- Hover information
- Rename refactoring

**Integration Points:**
- LSP server loads `.res.json` on file changes
- Maps LSP positions to resolution data
- Provides symbol information from resolution table

### IDE Extensions

- VS Code extension for PSX/Topple
- Syntax highlighting
- Resolution data integration
- Inline diagnostics
- Scope visualization

## Success Criteria

✅ **JSON output** contains all resolution information in a structured format
✅ **Text output** is human-readable and informative
✅ **Parse command** supports resolution output with appropriate flags
✅ **Tests** cover all major functionality and edge cases
✅ **Documentation** explains how to use the feature
✅ **Foundation** is laid for web visualization and LSP integration

## References

- Current implementation: `compiler/resolver/`
- Debug output: `compiler/resolver/debug.go`
- Parser output: `cmd/parse.go`
- AST printer: `compiler/printer.go`
