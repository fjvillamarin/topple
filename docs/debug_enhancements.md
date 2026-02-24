# Debug Enhancements

This document describes planned enhancements to Topple's debug tooling. Each proposal addresses gaps identified in the current debug infrastructure and aims to make the compiler more observable and developer-friendly.

## Current State

The compiler exposes debug information through three CLI commands:

| Stage    | Command               | Output              | Formats          |
|----------|-----------------------|---------------------|------------------|
| Lexer    | `topple scan`         | Token stream        | `.tok` / console |
| Parser   | `topple parse`        | AST tree            | `.ast` / console |
| Resolver | `topple parse --format` | Scopes + variables | `.res` / `.res.json` |
| Codegen  | _(none)_              | _(none)_            | _(none)_         |

### Known Problems

1. **Duplicate debug systems** -- `debug.go` and `text_output.go` both format resolution data for human consumption with duplicated logic for sorting, variable formatting, flag rendering, and scope type mapping.
2. **Unconditional debug output** -- `compiler.go:60` always calls `DebugPrintResolutionTable` during compilation, polluting stdout with no way to disable it.
3. **No intermediate output from `compile`** -- The compile command is a black box. Debugging requires switching to `topple parse` and mentally tracing through transformation and codegen.
4. **Token output is raw** -- Scan output shows numeric token codes (`Identifier 72 "view"`) which aren't meaningful. Tokens aren't grouped by source line.
5. **No source-annotated views** -- There is no way to see resolution info overlaid on the original source code.
6. **No warning system** -- The resolver tracks `IsUsed`, `UsedBeforeDef`, and shadowing but never generates warnings from them.
7. **No pipeline stage control** -- Cannot stop the pipeline at a specific stage or dump intermediate artifacts selectively.
8. **Scattered format helpers** -- `formatScopeType` in `debug.go` and `scopeTypeToString` in `json.go` do the same thing. Same for variable state formatting.

---

## Output Format Strategy

All debug commands and new features will support two output modes:

- **Text** -- Human-readable, formatted for terminal consumption. This is the default.
- **JSON** -- Machine-readable, structured for programmatic processing, piping into `jq`, editor integrations, web visualizers, and CI tooling.

Each feature section below specifies both formats. The JSON schemas are designed to be stable and composable -- downstream tools can rely on the structure.

A global `--json` flag will be the standard way to request JSON output across all commands:

```bash
topple scan file.psx --json          # JSON token stream
topple parse file.psx --json         # JSON AST
topple inspect file.psx --json       # JSON annotated output
topple compile file.psx --emit ast --json  # JSON intermediate artifacts
```

---

## A. Consolidate Debug Output

**Type:** Refactor
**Priority:** 1 (do first -- unblocks other work)
**Effort:** Low

### Goal

Eliminate the duplicate debug system in the resolver. Remove `debug.go` entirely and consolidate all formatting logic into well-organized shared code.

### Changes

1. **Delete `compiler/resolver/debug.go`** -- All three functions (`DebugPrintResolutionTable`, `DebugPrintResolver`, `DebugPrintCurrentScope`) are replaced by `text_output.go` and `json.go`.

2. **Create `compiler/resolver/format.go`** -- Extract shared helpers used by both text and JSON output:
   - `FormatVariableState(state) string`
   - `FormatVariableType(variable) string`
   - `FormatVariableFlags(variable) string`
   - `FormatScopeType(scopeType) string`
   - `FormatSpan(span) string`

   These become the single source of truth for formatting resolver data. Both `text_output.go` and `json.go` import from here instead of maintaining their own copies.

3. **Remove the unconditional call in `compiler.go:60`** -- `DebugPrintResolutionTable` is deleted. Resolution output is now only produced when explicitly requested via CLI flags.

4. **Update `text_output.go`** -- Import helpers from `format.go` instead of defining its own.

5. **Update `json.go`** -- Import helpers from `format.go` for `scopeTypeToString`, `variableStateToString`, etc.

### Result

- ~200 lines of duplicated code removed
- Single set of formatting functions
- No more unwanted stdout pollution during `topple compile`

---

## B. `--emit` Flag on `compile`

**Type:** Feature
**Priority:** 2
**Effort:** Low

### Goal

Let users dump intermediate compilation artifacts alongside the compiled output, without switching to a different command.

### CLI Interface

```bash
topple compile file.psx --emit tokens              # also write .tok
topple compile file.psx --emit ast                  # also write .ast
topple compile file.psx --emit resolution           # also write .res
topple compile file.psx --emit transformed-ast      # AST after view transform
topple compile file.psx --emit all                  # write everything
topple compile file.psx --emit ast,resolution       # comma-separated selection
```

Each `--emit` stage writes a file alongside the `.py` output using the corresponding extension (`.tok`, `.ast`, `.res`, `.res.json`).

### Flag Definition

```go
type CompileCmd struct {
    Input  string `arg:"" required:""`
    Output string `arg:"" optional:""`
    Emit   string `help:"Emit intermediate artifacts: tokens, ast, resolution, transformed-ast, all" default:""`
}
```

### Emit Stages

| Value              | File Extension | Contents                               |
|--------------------|----------------|----------------------------------------|
| `tokens`           | `.tok`         | Token stream from lexer                |
| `ast`              | `.ast`         | AST after parsing                      |
| `resolution`       | `.res`         | Resolution table (text)                |
| `resolution`       | `.res.json`    | Resolution table (JSON) -- always written alongside `.res` |
| `transformed-ast`  | `.tast`        | AST after view transformation          |
| `all`              | all of above   | Every intermediate artifact            |

### JSON Mode

When combined with `--json`, emit files use JSON format where applicable:

```bash
topple compile file.psx --emit tokens --json   # .tok.json instead of .tok
topple compile file.psx --emit ast --json      # .ast.json instead of .ast
```

### Implementation

The `Compile` method in `compiler.go` already runs each pipeline stage sequentially. The change captures each intermediate result and writes it out based on the `--emit` flag. No changes to the pipeline itself.

---

## C. `topple inspect` Command

**Type:** Feature
**Priority:** 5
**Effort:** Medium

### Goal

A unified debug command that provides a single entry point for inspecting any compilation stage. Replaces the mental overhead of choosing between `scan`, `parse`, and `compile` for debugging.

### CLI Interface

```bash
topple inspect file.psx                        # default: source-annotated summary
topple inspect file.psx --stage tokens         # stop at lexer
topple inspect file.psx --stage ast            # stop at parser
topple inspect file.psx --stage resolution     # stop at resolver
topple inspect file.psx --stage transform      # stop after view transformation
topple inspect file.psx --stage codegen        # show final generated code
```

### Flag Definition

```go
type InspectCmd struct {
    Input string `arg:"" required:"" help:"Path to a PSX file"`
    Stage string `help:"Pipeline stage to inspect: tokens, ast, resolution, transform, codegen" default:"summary" enum:"summary,tokens,ast,resolution,transform,codegen"`
}
```

### Default Mode (`summary`)

When no `--stage` is given, show a compact overview:

```
file: views.psx
tokens: 142
ast nodes: 38
scopes: 3 (module, view, function)
variables: 7 (3 params, 4 local)
views: 1 (HelloWorld)
warnings: 1
errors: 0
```

### JSON Mode

```bash
topple inspect file.psx --json
```

```json
{
  "file": "views.psx",
  "tokens": 142,
  "ast_nodes": 38,
  "scopes": {
    "total": 3,
    "types": ["module", "view", "function"]
  },
  "variables": {
    "total": 7,
    "parameters": 3,
    "local": 4
  },
  "views": ["HelloWorld"],
  "diagnostics": {
    "warnings": 1,
    "errors": 0
  }
}
```

Stage-specific JSON output follows the schemas defined in each stage's respective section.

### Relationship to `scan` and `parse`

`inspect` does not replace `scan` and `parse` -- those remain for file-writing workflows (`-w` flag). `inspect` is the interactive, console-first debug tool.

---

## D. Source-Annotated Resolution Output

**Type:** Feature
**Priority:** 4
**Effort:** Medium

### Goal

The most practically useful debug format: the original source code with resolution annotations overlaid inline. This is what developers actually want when debugging "why isn't this variable resolving correctly?"

### Text Format

```
  1 | view HelloWorld(name: str = "World"):
    |      ^^^^^^^^^^                        view, defined, scope #0
    |                  ^^^^                  view-param, str, used, depth=1
    |                                ^^^^^^^ literal, str
  2 |     <div>Hello, {name}!</div>
    |                  ^^^^                  view-param, ref -> scope #1, depth=1
  3 |
  4 |     x = 42
    |     ^                                  local, defined, scope #1, unused
```

Each source line is followed by zero or more annotation lines pointing to specific spans with resolution metadata.

### JSON Format

```json
{
  "file": "views.psx",
  "lines": [
    {
      "line": 1,
      "source": "view HelloWorld(name: str = \"World\"):",
      "annotations": [
        {
          "start_col": 5,
          "end_col": 15,
          "name": "HelloWorld",
          "type": "view",
          "state": "defined",
          "scope_id": 0,
          "depth": 0
        },
        {
          "start_col": 16,
          "end_col": 20,
          "name": "name",
          "type": "view-param",
          "state": "used",
          "scope_id": 1,
          "depth": 1
        }
      ]
    },
    {
      "line": 2,
      "source": "    <div>Hello, {name}!</div>",
      "annotations": [
        {
          "start_col": 17,
          "end_col": 21,
          "name": "name",
          "type": "view-param",
          "state": "used",
          "scope_id": 1,
          "depth": 1,
          "context": "reference"
        }
      ]
    }
  ]
}
```

### CLI Integration

Available as a format option in `parse` and as the default for `inspect`:

```bash
topple parse file.psx --format annotated          # text annotated view
topple parse file.psx --format annotated --json   # JSON annotated view
topple inspect file.psx                            # default shows annotated view
```

### Implementation

Requires access to the original source text alongside the resolution table. The annotator:

1. Splits source into lines
2. Iterates all resolved variables and their references
3. Maps each reference's `Span` to a source line and column range
4. Groups annotations by line
5. Sorts annotations by column within each line
6. Renders the underline markers and metadata

---

## E. Resolver Warnings

**Type:** Feature
**Priority:** 3
**Effort:** Medium

### Goal

Wire up existing variable metadata to produce actionable warnings. The resolver already tracks everything needed -- it just doesn't report it.

### Warning Types

| Warning | Condition | Severity |
|---------|-----------|----------|
| Unused variable | `IsUsed == false` and not `_`-prefixed | warning |
| Unused parameter | `IsParameter && !IsUsed` and not `_`-prefixed | info |
| Used before definition | `UsedBeforeDef == true` | warning |
| Variable shadows outer | `Binding.ShadowedBy != nil` | info |
| Unused import | `IsImported && !IsUsed` | warning |

### Data Model

Add a `Warnings` field to `ResolutionTable`:

```go
type DiagnosticSeverity int

const (
    SeverityError DiagnosticSeverity = iota
    SeverityWarning
    SeverityInfo
)

type Diagnostic struct {
    Severity DiagnosticSeverity
    Message  string
    Span     lexer.Span
    Code     string // e.g., "W001", "W002" for stable references
}
```

```go
type ResolutionTable struct {
    // ... existing fields ...
    Warnings []Diagnostic
}
```

### Warning Codes

| Code | Name | Message Template |
|------|------|-----------------|
| W001 | unused-variable | `variable '{name}' is defined but never used` |
| W002 | unused-parameter | `parameter '{name}' is never used` |
| W003 | used-before-def | `variable '{name}' is used before definition` |
| W004 | shadowed-variable | `variable '{name}' shadows variable from outer scope` |
| W005 | unused-import | `imported name '{name}' is never used` |

### Text Output

Warnings appear in the DIAGNOSTICS section of `.res` files:

```
DIAGNOSTICS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Errors: 0
  Warnings: 2
    W001 variable 'x' is defined but never used (4:4)
    W004 variable 'name' shadows variable from outer scope (7:8)
```

### JSON Output

```json
{
  "diagnostics": {
    "errors": [],
    "warnings": [
      {
        "code": "W001",
        "severity": "warning",
        "message": "variable 'x' is defined but never used",
        "span": {
          "start": {"line": 4, "column": 4},
          "end": {"line": 4, "column": 5}
        }
      },
      {
        "code": "W004",
        "severity": "info",
        "message": "variable 'name' shadows variable from outer scope",
        "span": {
          "start": {"line": 7, "column": 8},
          "end": {"line": 7, "column": 12}
        }
      }
    ]
  }
}
```

### Implementation

Add a `generateWarnings()` method called at the end of `Resolve()`, after all scopes are closed. It iterates all tracked variables and checks the conditions above. Warnings are appended to the new `Warnings` field. Both `text_output.go` and `json.go` are updated to render them.

---

## F. Improved Token Output

**Type:** Refactor
**Priority:** 6
**Effort:** Low

### Goal

Make token output useful for understanding lexer behavior by grouping tokens by source line and showing meaningful categories instead of raw numeric codes.

### Text Format

Current:
```
0: Identifier 72 "view" <nil> @ 1:0-1:4
1: Identifier 72 "HelloWorld" <nil> @ 1:5-1:15
2: LeftParen 0 "(" <nil> @ 1:15-1:16
```

Proposed:
```
Line 1:
  KEYWORD    view           1:0-1:4
  IDENT      HelloWorld     1:5-1:15
  PUNCT      (              1:15-1:16
  IDENT      name           1:16-1:20
  PUNCT      :              1:20-1:21
  IDENT      str            1:22-1:25
  OP         =              1:26-1:27
  STRING     "World"        1:28-1:35
  PUNCT      )              1:35-1:36
  PUNCT      :              1:36-1:37

Line 2:
  HTML_TAG   <div>          2:4-2:9
  HTML_TEXT  Hello,         2:9-2:16
  INTERP     {              2:16-2:17
  IDENT      name           2:17-2:21
  INTERP     }              2:21-2:22
  HTML_TEXT  !              2:22-2:23
  HTML_TAG   </div>         2:23-2:29
```

Token categories:
- `KEYWORD` -- Language keywords (`view`, `def`, `class`, `if`, etc.)
- `IDENT` -- Identifiers
- `OP` -- Operators (`+`, `-`, `==`, `**`, etc.)
- `PUNCT` -- Punctuation (`(`, `)`, `:`, `,`, etc.)
- `STRING` -- String literals
- `NUMBER` -- Numeric literals
- `HTML_TAG` -- HTML tag tokens
- `HTML_TEXT` -- HTML text content
- `INTERP` -- Interpolation delimiters
- `FSTRING` -- F-string parts
- `SPECIAL` -- Indent, dedent, newline, EOF

### JSON Format

```json
{
  "file": "views.psx",
  "total_tokens": 42,
  "lines": [
    {
      "line": 1,
      "tokens": [
        {
          "index": 0,
          "type": "Identifier",
          "category": "KEYWORD",
          "lexeme": "view",
          "literal": null,
          "span": {
            "start": {"line": 1, "column": 0},
            "end": {"line": 1, "column": 4}
          }
        },
        {
          "index": 1,
          "type": "Identifier",
          "category": "IDENT",
          "lexeme": "HelloWorld",
          "literal": null,
          "span": {
            "start": {"line": 1, "column": 5},
            "end": {"line": 1, "column": 15}
          }
        }
      ]
    }
  ],
  "errors": []
}
```

### Implementation

Add a `TokenCategory(token) string` function that maps token types to categories. Update the scan command's output formatter. The existing `.tok` format remains available for backwards compatibility via a flag if needed, but the new format becomes the default.

---

## G. JSON AST Output

**Type:** Feature
**Priority:** 3 (alongside E)
**Effort:** Medium

### Goal

The AST printer currently only outputs the tree-sitter-like text format. Add a JSON serialization so the AST can be consumed by external tools, diffed programmatically, and used in test assertions.

### JSON Format

```json
{
  "type": "Module",
  "span": {"start": {"line": 1, "column": 0}, "end": {"line": 5, "column": 0}},
  "body": [
    {
      "type": "ViewStmt",
      "span": {"start": {"line": 1, "column": 0}, "end": {"line": 3, "column": 0}},
      "name": {
        "type": "Name",
        "value": "HelloWorld",
        "span": {"start": {"line": 1, "column": 5}, "end": {"line": 1, "column": 15}}
      },
      "params": [
        {
          "type": "Parameter",
          "name": "name",
          "annotation": {
            "type": "Name",
            "value": "str"
          },
          "default": {
            "type": "StringLiteral",
            "value": "World"
          }
        }
      ],
      "body": [
        {
          "type": "HTMLElement",
          "tag": "div",
          "children": [
            {
              "type": "HTMLText",
              "value": "Hello, "
            },
            {
              "type": "HTMLInterpolation",
              "expression": {
                "type": "Name",
                "value": "name"
              }
            },
            {
              "type": "HTMLText",
              "value": "!"
            }
          ]
        }
      ]
    }
  ]
}
```

### CLI Integration

```bash
topple parse file.psx --json           # JSON AST to console
topple parse file.psx -w --json        # write .ast.json file
```

### Implementation

Add a `ToJSON()` method on `ast.Module` (or an `ASTSerializer` visitor) that walks the tree and produces a nested JSON structure. Each node type maps to a JSON object with a `type` field and node-specific fields. The visitor pattern is already in place -- this is a new visitor implementation alongside the existing `ASTPrinter`.

---

## Implementation Order

```
1. A  Consolidate debug output        (unblocks everything, removes noise)
2. B  --emit flag on compile           (quick win, high value)
3. E  Resolver warnings                (uses existing data, high value)
4. G  JSON AST output                  (enables JSON across the full pipeline)
5. D  Source-annotated output          (builds on resolution + source access)
6. F  Improved token output            (polish)
7. C  inspect command                  (ties everything together)
```

Each step is independently shippable. Later steps build on earlier ones but don't strictly require them.
