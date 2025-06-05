# Golden File Tests

This document describes the golden file testing system for the Sylfie compiler.

## Overview

Golden file tests are end-to-end tests that validate the complete compilation pipeline from `.psx` input files to `.py` output files. They use the "golden file" testing pattern where the expected output is stored in reference files and compared against the actual compiler output.

## Quick Start

```bash
# Show all available test categories and usage examples
mise run test-golden-list

# Run all golden file tests
mise run test-golden

# Update all golden files (use when compiler output changes intentionally)
mise run test-golden-update
```

## Test Organization

Tests are organized into categories in `compiler/testdata/`:

- **input/**: Contains `.psx` source files to compile
- **expected/**: Contains `.py` golden files with expected output  
- **generated/**: Contains `.py` files generated during test runs (local development only, not committed)

### Available Categories

- `basic` - Simple Python code and basic views
- `views` - View definitions and parameters  
- `control_flow` - If statements, loops, match, try/except
- `composition` - View composition and nesting
- `slots` - Slot functionality and templates
- `attributes` - HTML attribute handling
- `expressions` - Complex expressions and f-strings
- `python_integration` - Python-specific features
- `htmx` - HTMX integration examples
- `fastapi` - FastAPI integration examples
- `errors` - Error handling and edge cases

## Mise Tasks

### Running Tests

```bash
# Run all golden file tests
mise run test-golden

# Run tests for a specific category
mise run test-golden-category CATEGORY=basic

# Run a single test file
mise run test-golden-single TEST=basic/hello_world

# Show available categories and usage examples
mise run test-golden-list
```

### Updating Golden Files

⚠️ **Important**: Only update golden files when the compiler output has intentionally changed.

```bash
# Update all golden files
mise run test-golden-update

# Update golden files for a specific category
mise run test-golden-category-update CATEGORY=views

# Update a single golden file
mise run test-golden-single-update TEST=control_flow/loops
```

### Comparing Generated vs Expected Files

When tests run, generated output is automatically saved to `compiler/testdata/generated/` for easy comparison:

```bash
# Compare all files in a category
mise run test-golden-diff CATEGORY=basic

# Compare a single file
mise run test-golden-diff-single TEST=basic/hello_world

# Clean generated files (optional - they're gitignored)
mise run test-golden-clean
```

**File Locations for Manual Comparison:**
- Expected: `compiler/testdata/expected/<category>/<testname>.py`
- Generated: `compiler/testdata/generated/<category>/<testname>.py` 
- Input: `compiler/testdata/input/<category>/<testname>.psx`

## Adding New Tests

1. **Create input file**: Add a `.psx` file in `compiler/testdata/input/<category>/`
2. **Generate golden file**: Run `mise run test-golden-category-update CATEGORY=<category>`
3. **Verify output**: Check that the generated `.py` file in `compiler/testdata/expected/<category>/` is correct
4. **Run test**: Use `mise run test-golden-category CATEGORY=<category>` to verify

## Example Usage

```bash
# Test basic functionality
mise run test-golden-category CATEGORY=basic

# Test and update view-related tests
mise run test-golden-category CATEGORY=views
mise run test-golden-category-update CATEGORY=views

# Test a specific file that handles loops
mise run test-golden-single TEST=control_flow/loops

# Update golden file after fixing a bug in loop compilation
mise run test-golden-single-update TEST=control_flow/loops

# Compare generated vs expected files to debug differences
mise run test-golden-diff CATEGORY=basic
mise run test-golden-diff-single TEST=control_flow/loops
```

## Development Workflow

1. **Make changes to the compiler**
2. **Run tests**: `mise run test-golden-category CATEGORY=basic`
3. **If tests fail**: 
   - Check differences: `mise run test-golden-diff CATEGORY=basic`
   - View specific file: `compiler/testdata/generated/basic/hello_world.py`
   - Fix compiler issue or update golden files if change is intentional
4. **Clean up** (optional): `mise run test-golden-clean`

## Integration with CI/CD

Add to your CI pipeline:

```bash
# Run all tests (unit tests + golden file tests)
mise run test-all

# Or run just golden file tests
mise run test-golden
```

Golden file tests ensure that:
- The complete compilation pipeline works correctly
- Compiler output is consistent and predictable
- Regressions are caught automatically
- New features don't break existing functionality