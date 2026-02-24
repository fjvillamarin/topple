# Topple Documentation

Welcome to the Topple documentation! Topple is a modern templating language that seamlessly blends Python's power with HTML markup using the PSX (Python Syntax eXtension) syntax.

## Getting Started

- [README](../README.md) - Project overview and quick start
- [PSX Language Grammar](grammar_psx.md) - Complete language reference with examples
- [CLI Reference](cli.md) - Command-line tool documentation

## Architecture & Design

- [Architecture Guide](architecture.md) - Comprehensive system architecture
- [Runtime System](runtime.md) - Runtime library documentation and examples
- [Error Handling](error_handling.md) - Error reporting and recovery
- [Python Grammar](grammar_python.md) - Python PEG grammar reference

## Testing

- [Testing Strategy](tests.md) - Comprehensive testing documentation
- [Golden File Tests](golden_tests.md) - End-to-end testing guide

## Development Documentation

The `development/` folder contains internal documentation for contributors:

### Design Documents
- [Resolver Design](development/resolver_design.md) - Symbol resolution system design
- [AST Reference](development/ast_reference.md) - Complete AST node reference

### Bug Tracking
- [Parser Bugs](development/parser_bugs.md) - Known parser issues
- [Codegen Bugs](development/codegen_bugs.md) - Known code generation issues

## Examples

See the [examples/](../examples/) directory for working Topple applications demonstrating various features.

## Contributing

1. Read the [Architecture Guide](architecture.md) to understand the system
2. Check the development documentation for the area you want to contribute to
3. Follow the testing guidelines in [Testing Strategy](tests.md)
4. Submit a pull request with tests
