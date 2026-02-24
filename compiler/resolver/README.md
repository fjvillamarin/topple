# Topple Variable Resolver

This package implements variable resolution for the Topple compiler, handling Python-like scoping rules with extensions for PSX-specific constructs.

## Overview

The resolver performs variable binding and scope analysis on the parsed AST before code generation. It implements Python's LEGB (Local, Enclosing, Global, Built-in) scoping rules and tracks additional metadata needed for proper code generation.

## Architecture

### Core Components

- **`types.go`**: Data structures for variables, scopes, and resolution results
- **`resolver.go`**: Main resolver logic and LEGB implementation
- **`visitor.go`**: AST visitor implementation for traversing and analyzing nodes

### Key Data Structures

- **`Variable`**: Metadata about a variable (scope depth, usage flags, etc.)
- **`Environment`**: Represents a scope with variable bindings
- **`Resolver`**: Main resolver state with scope stack and resolution results
- **`ResolutionTable`**: Final output containing variable→binding mappings

## Usage

```go
// Create a resolver
resolver := NewResolver()

// Resolve variables in a module
table, err := resolver.Resolve(module)
if err != nil {
    // Handle resolution errors
}

// Use resolution table in subsequent compiler phases
generator := codegen.NewCodeGeneratorWithResolution(table)
```

## Features Implemented

### Scoping
- Module-level scope
- Function scope
- Class scope (with isolation rules)
- View scope (PSX extension)
- Comprehension scoping (list/dict/set comprehensions with their own scope)
- Exception handler scopes
- With statement scopes
- Lambda expressions

### Variable Tracking
- Variable definition and usage
- Scope depth calculation
- Assignment target analysis
- Closure variable analysis
- Import statement handling

### Python Compatibility
- Global/nonlocal declaration handling
- LEGB resolution order
- Class scope isolation from nested functions
- Control flow analysis (if/for/while/try)

### Error Detection
- Undefined variable errors
- Invalid global/nonlocal usage
- Binding conflicts

## Integration Points

The resolver integrates with:

1. **Parser**: Takes parsed AST as input
2. **Transformers**: Provides resolution table for view parameter transformation
3. **Code Generator**: Uses resolution info for optimization and correctness

## Testing

Run tests with:
```bash
go test -v
```

The test suite covers:
- Basic resolver functionality
- Scope management
- Variable definition and lookup
- Resolution table generation

## Design Patterns

The resolver follows the **Visitor Pattern** to traverse the AST, with specialized logic for:

- **Expression visitors**: Handle variable references and complex expressions
- **Statement visitors**: Handle variable definitions and scope-creating constructs
- **Assignment analysis**: Special handling for assignment targets (unpacking, etc.)
