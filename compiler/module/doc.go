// Package module provides module path resolution for the Topple compiler.
//
// The module resolver translates import paths from PSX source code into
// filesystem paths to .psx files. It handles:
//
//   - Absolute imports: "my_module" -> "./my_module.psx"
//   - Relative imports: ".sibling" -> "../sibling.psx"
//   - Package imports: "pkg" -> "./pkg/__init__.psx"
//
// The resolver respects Python's import semantics while working with
// .psx file extensions instead of .py.
//
// # Example Usage
//
//	config := module.Config{
//		RootDir:    ".",
//		FileSystem: filesystem.NewOSFileSystem(),
//	}
//	resolver := module.NewResolver(config)
//
//	// Resolve absolute import
//	path, err := resolver.ResolveAbsolute(ctx, "components.button")
//	// Returns: "./components/button.psx"
//
//	// Resolve relative import
//	path, err := resolver.ResolveRelative(ctx, 1, "sibling", "/proj/app.psx")
//	// Returns: "/proj/sibling.psx"
package module
