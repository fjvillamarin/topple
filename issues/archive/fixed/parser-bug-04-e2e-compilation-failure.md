# 🐛 Parser Bug: E2E Compilation Failures - HTML Tags Not Recognized in Views

## Summary
The E2E compilation test for basic view files is failing because the scanner is not properly switching to HTML mode when parsing HTML elements inside view bodies. The scanner treats `<` as a less-than operator instead of a tag opener, causing parsing errors like `at '/': unexpected token`.

## Priority: **CRITICAL** 🔴🔴🔴
This completely breaks the core functionality of the framework - parsing views with HTML content.

## Bug Details

### Location
- **File**: `compiler/e2e_test.go:245`
- **Test**: `TestE2EBasicFeatures/Simple_hello_world_view_with_basic_HTML`
- **Input File**: `compiler/testdata/input/basic/hello_world.psx`
- **Scanner File**: `compiler/lexer/scanner.go`
- **Parser File**: `compiler/parser/view.go`

### Current Behavior
```python
# Input file (hello_world.psx)
view HelloWorld():
    <div>Hello World</div>

# Error
at '/': unexpected token

# Expected: Should compile to Python class with render method
```

### Expected Behavior
The view should compile to:
```python
from runtime import BaseView, el

class HelloWorld(BaseView):
    def __init__(self):
        super().__init__()
    
    def _render(self):
        return el("div", "Hello World")
```

## Root Cause Analysis

After thorough investigation, I found that the scanner has the necessary infrastructure but there's a timing/state issue:

1. **Scanner State Management**:
   - The scanner correctly increments `viewDepth` when it sees the `view` keyword (line 737)
   - The scanner has logic to switch to `HTMLTagMode` when `viewDepth > 0` and line starts with `<` (lines 283-285)
   - BUT the mode switching only happens at line start via `handleLineStart()`

2. **The Problem**:
   - After parsing `view HelloWorld():`, the scanner is in `PythonMode`
   - When it encounters the indented `<div>` line, it should switch to `HTMLTagMode`
   - However, the `<` is being scanned in `PythonMode` where it's treated as a less-than operator (lines 532-543)

3. **Mode Detection Issue**:
   - The `atLineStart` flag might not be set correctly after the colon and newline
   - Or the mode detection happens too late, after the `<` has already been tokenized as `Less`

## Action Plan

### 1. **Fix Line Start Detection**
Ensure that after a colon and newline in a view context, the scanner properly sets `atLineStart = true` so that `handleLineStart()` is called before scanning the `<`.

### 2. **Improve Mode Switching**
The scanner should check for HTML mode not just at line start but also when encountering `<` inside a view body.

### 3. **Debug Token Generation**
Add logging to trace:
- When `viewDepth` changes
- When `atLineStart` is set/cleared
- When `handleLineStart()` is called
- What mode is active when scanning `<`

## Implementation Steps

### Fix Option 1: Ensure Line Start Detection
```go
// In scanNewline() or similar
func (s *Scanner) scanNewline() {
    s.line++
    s.column = 0
    s.addToken(Newline)
    s.ctx.atLineStart = true  // Ensure this is set
}
```

### Fix Option 2: Context-Aware `<` Handling
```go
// In scanToken() for '<' case
case '<':
    // If we're in a view and could be starting an HTML tag
    if s.ctx.viewDepth > 0 && s.ctx.mode == PythonMode {
        // Check if this looks like an HTML tag
        if s.peekAlpha() || s.peek() == '/' {
            s.ctx.mode = HTMLTagMode
            s.addToken(TagOpen)
            return
        }
    }
    // Otherwise, treat as less-than operator
    if s.match('<') {
        // ... existing operator logic
    }
```

### Fix Option 3: Preprocessing Check
Before scanning each token in view context, check if we should switch modes:
```go
func (s *Scanner) scanToken() {
    // Add mode detection for view context
    if s.ctx.viewDepth > 0 && s.ctx.mode == PythonMode {
        if s.peek() == '<' && (s.peekNext() == '/' || isAlpha(s.peekNext())) {
            s.ctx.mode = HTMLTagMode
        }
    }
    
    // Continue with existing logic
    // ...
}
```

## Test Cases to Verify

```python
# Basic view with HTML
view HelloWorld():
    <div>Hello World</div>

# View with nested HTML
view Card(title: str):
    <div class="card">
        <h1>{title}</h1>
        <p>Content here</p>
    </div>

# View with Python and HTML mixed
view TodoList(items: list):
    <ul>
        for item in items:
            <li>{item}</li>
    </ul>

# Single-line HTML
view Badge(text: str): <span class="badge">{text}</span>
```

## Expected Impact

### Fixes
- All E2E view compilation tests will pass
- HTML elements will be properly recognized inside view bodies
- Mixed Python/HTML syntax will work correctly

### Risks
- Need to ensure `<` as less-than operator still works in Python expressions inside views
- Mode switching must be precise to avoid breaking Python code
- Performance impact of additional checks should be minimal

## Debugging Steps

1. **Add Debug Logging**:
```go
// In scanToken()
if s.ctx.viewDepth > 0 {
    fmt.Printf("Scanning in view: mode=%v, char=%c, atLineStart=%v\n", 
        s.ctx.mode, s.peek(), s.ctx.atLineStart)
}
```

2. **Trace Token Stream**:
Run the scanner on the failing test input and print all tokens to see where it goes wrong.

3. **Unit Test Scanner**:
Create specific scanner tests for view + HTML scenarios to isolate the issue.

## References
- Scanner modes defined at lines 44-49 in scanner.go
- View detection at line 737 in scanner.go  
- HTML mode switching at lines 283-285 in scanner.go
- `<` operator handling at lines 532-543 in scanner.go