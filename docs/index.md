# Sylfie Documentation

Welcome to the Sylfie documentation! Sylfie is a modern templating language that seamlessly blends Python's power with HTML markup using the PSX (Python Sylfie eXtension) syntax.

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
- [Golden File Tests](GOLDEN_TESTS.md) - End-to-end testing guide

## Development Documentation

The `development/` folder contains internal documentation for contributors:

### Design Documents
- [Parser Design](development/parser_design.md) - Parser architecture and decisions
- [Resolver Design](development/resolver_design.md) - Symbol resolution system design
- [Filesystem Design](development/filesystem_design.md) - File system integration design
- [AST Reference](development/ast_reference.md) - Complete AST node reference

### Development Guides
- [Parser Test Improvements](development/parser_test_improvements.md) - Testing strategy for parser
- [Compilation Analysis](development/compilation_analysis.md) - Compilation process analysis
- [Memory State](development/memory.md) - Development context and memory state

### Test Plans
- [Transformers Test Plan](development/transformers_test_plan.md) - Test planning for transformers
- [Codegen Tests](development/codegen_tests.md) - Code generation test documentation
- [View Transformer Tests](development/view_transformer_tests.md) - View transformation testing

### Bug Tracking
- [Parser Bugs](development/parser_bugs.md) - Known parser issues
- [Codegen Bugs](development/codegen_bugs.md) - Known code generation issues

## Examples

See the [examples/](../examples/) directory for working Sylfie applications demonstrating various features.

## Contributing

1. Read the [Architecture Guide](architecture.md) to understand the system
2. Check the development documentation for the area you want to contribute to
3. Follow the testing guidelines in [Testing Strategy](tests.md)
4. Submit a pull request with tests

## Legacy Documentation

- [Legacy Grammar Reference](grammar_psx_legacy.md) - Original grammar documentation (historical reference)