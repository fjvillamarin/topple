package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fjvillamarin/topple/compiler"
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/codegen"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"github.com/fjvillamarin/topple/compiler/resolver"
	"github.com/fjvillamarin/topple/compiler/transformers"
	"github.com/fjvillamarin/topple/internal/filesystem"
)

// InspectCmd defines the "inspect" command which provides a unified
// entry point for inspecting any compilation stage.
type InspectCmd struct {
	Input string `arg:"" required:"" help:"Path to a PSX file"`
	Stage string `help:"Pipeline stage to inspect: summary, tokens, ast, resolution, transform, codegen" default:"summary" enum:"summary,tokens,ast,resolution,transform,codegen"`
	JSON  bool   `help:"Output in JSON format" default:"false"`
}

// Run executes the inspect command.
func (c *InspectCmd) Run(globals *Globals, ctx *context.Context, log *slog.Logger) error {
	fs := filesystem.NewFileSystem(log)

	exists, err := fs.Exists(c.Input)
	if err != nil {
		return fmt.Errorf("error checking input path: %w", err)
	}
	if !exists {
		return fmt.Errorf("input path does not exist: %s", c.Input)
	}

	content, err := fs.ReadFile(c.Input)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", c.Input, err)
	}

	filename := filepath.Base(c.Input)

	switch c.Stage {
	case "summary":
		return c.inspectSummary(content, filename)
	case "tokens":
		return c.inspectTokens(content, filename)
	case "ast":
		return c.inspectAST(content, filename)
	case "resolution":
		return c.inspectResolution(content, filename)
	case "transform":
		return c.inspectTransform(content, filename)
	case "codegen":
		return c.inspectCodegen(content, filename)
	default:
		return fmt.Errorf("unknown stage: %s", c.Stage)
	}
}

// inspectSummary shows a compact overview of all pipeline stages.
func (c *InspectCmd) inspectSummary(content []byte, filename string) error {
	// Scan
	tokens, scanErrs := compiler.Scan(content)
	tokenCount := 0
	if tokens != nil {
		tokenCount = len(tokens)
	}

	// Parse
	var module *ast.Module
	var parseErrs []error
	if tokens != nil {
		module, parseErrs = compiler.ParseTokens(tokens)
	}

	// Count AST nodes
	astNodeCount := 0
	if module != nil {
		printer := compiler.NewASTPrinter("  ")
		output := printer.Print(module)
		astNodeCount = countASTNodes(output)
	}

	// Resolve
	var table *resolver.ResolutionTable
	if module != nil {
		res := resolver.NewResolver()
		table, _ = res.Resolve(module)
	}

	// Collect all errors
	var allErrors []error
	allErrors = append(allErrors, scanErrs...)
	allErrors = append(allErrors, parseErrs...)
	if table != nil {
		allErrors = append(allErrors, table.Errors...)
	}

	// Gather stats
	scopeCount := 0
	var scopeTypes []string
	varTotal := 0
	paramCount := 0
	localCount := 0
	var viewNames []string

	if table != nil {
		scopeCount = len(table.Scopes)

		// Collect scope types
		scopeTypeSet := make(map[string]bool)
		for _, scope := range table.Scopes {
			scopeTypeSet[formatScopeType(scope.ScopeType)] = true
		}
		for t := range scopeTypeSet {
			scopeTypes = append(scopeTypes, t)
		}
		sort.Strings(scopeTypes)

		// Count unique variables
		uniqueVars := make(map[*resolver.Variable]bool)
		for _, v := range table.Variables {
			uniqueVars[v] = true
		}
		varTotal = len(uniqueVars)
		for v := range uniqueVars {
			if v.IsParameter {
				paramCount++
			} else {
				localCount++
			}
		}

		// View names
		for name := range table.Views {
			viewNames = append(viewNames, name)
		}
		sort.Strings(viewNames)
	}

	if c.JSON {
		return c.printSummaryJSON(filename, tokenCount, astNodeCount, scopeCount, scopeTypes, varTotal, paramCount, localCount, viewNames, len(allErrors))
	}

	// Text output
	fmt.Printf("file: %s\n", filename)
	fmt.Printf("tokens: %d\n", tokenCount)
	fmt.Printf("ast nodes: %d\n", astNodeCount)
	if len(scopeTypes) > 0 {
		fmt.Printf("scopes: %d (%s)\n", scopeCount, strings.Join(scopeTypes, ", "))
	} else {
		fmt.Printf("scopes: %d\n", scopeCount)
	}
	fmt.Printf("variables: %d (%d params, %d local)\n", varTotal, paramCount, localCount)
	if len(viewNames) > 0 {
		fmt.Printf("views: %d (%s)\n", len(viewNames), strings.Join(viewNames, ", "))
	} else {
		fmt.Printf("views: 0\n")
	}
	fmt.Printf("errors: %d\n", len(allErrors))
	return nil
}

// printSummaryJSON outputs the summary in JSON format.
func (c *InspectCmd) printSummaryJSON(filename string, tokens, astNodes, scopes int, scopeTypes []string, vars, params, local int, views []string, errors int) error {
	summary := map[string]any{
		"file":      filename,
		"tokens":    tokens,
		"ast_nodes": astNodes,
		"scopes": map[string]any{
			"total": scopes,
			"types": scopeTypes,
		},
		"variables": map[string]any{
			"total":      vars,
			"parameters": params,
			"local":      local,
		},
		"views": views,
		"diagnostics": map[string]any{
			"errors": errors,
		},
	}
	return printJSON(summary)
}

// inspectTokens shows the token stream.
func (c *InspectCmd) inspectTokens(content []byte, filename string) error {
	scanner := lexer.NewScanner(content)
	tokens := scanner.ScanTokens()

	if c.JSON {
		type jsonToken struct {
			Index   int    `json:"index"`
			Type    string `json:"type"`
			TypeNum int    `json:"type_num"`
			Lexeme  string `json:"lexeme"`
			Literal any    `json:"literal"`
			Span    string `json:"span"`
		}
		var out []jsonToken
		for i, tok := range tokens {
			out = append(out, jsonToken{
				Index:   i,
				Type:    tok.Type.String(),
				TypeNum: int(tok.Type),
				Lexeme:  tok.Lexeme,
				Literal: tok.Literal,
				Span:    tok.Span.String(),
			})
		}
		return printJSON(out)
	}

	// Text output — same format as topple scan
	fmt.Printf("=== %s ===\n\n", filename)
	for i, tok := range tokens {
		fmt.Printf("%d: %s %d %q %v @ %s\n",
			i, tok.Type, int(tok.Type), tok.Lexeme, tok.Literal, tok.Span.String())
	}

	if len(scanner.Errors) > 0 {
		fmt.Printf("\n-- Errors (%d) --\n", len(scanner.Errors))
		for i, e := range scanner.Errors {
			fmt.Printf("%d: %v\n", i+1, e)
		}
	}
	return nil
}

// inspectAST shows the parsed AST.
func (c *InspectCmd) inspectAST(content []byte, filename string) error {
	module, errors := compiler.Parse(content)

	if c.JSON {
		result := map[string]any{
			"file": filename,
		}
		if module != nil {
			printer := compiler.NewASTPrinter("  ")
			result["ast"] = printer.Print(module)
		}
		if len(errors) > 0 {
			var errStrings []string
			for _, e := range errors {
				errStrings = append(errStrings, e.Error())
			}
			result["errors"] = errStrings
		}
		return printJSON(result)
	}

	// Text output — same format as topple parse
	fmt.Printf("=== %s ===\n\n", filename)
	if module != nil {
		printer := compiler.NewASTPrinter("  ")
		fmt.Print(printer.Print(module))
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Printf("\n-- Errors (%d) --\n", len(errors))
		for i, e := range errors {
			fmt.Printf("%d: %v\n", i+1, e)
		}
	}
	return nil
}

// inspectResolution shows the resolution table.
func (c *InspectCmd) inspectResolution(content []byte, filename string) error {
	module, errors := compiler.Parse(content)
	if module == nil {
		return formatParseErrors(errors)
	}

	res := resolver.NewResolver()
	table, err := res.Resolve(module)
	if err != nil {
		return fmt.Errorf("resolution failed: %w", err)
	}

	if c.JSON {
		jsonRes, err := table.ToJSON(filename)
		if err != nil {
			return fmt.Errorf("JSON conversion failed: %w", err)
		}
		return printJSON(jsonRes)
	}

	text, err := table.ToText(filename)
	if err != nil {
		return fmt.Errorf("text conversion failed: %w", err)
	}
	fmt.Print(text)
	return nil
}

// inspectTransform shows the AST after view transformation.
func (c *InspectCmd) inspectTransform(content []byte, filename string) error {
	module, errors := compiler.Parse(content)
	if module == nil {
		return formatParseErrors(errors)
	}

	res := resolver.NewResolver()
	table, err := res.Resolve(module)
	if err != nil {
		return fmt.Errorf("resolution failed: %w", err)
	}
	if len(table.Errors) > 0 {
		return fmt.Errorf("resolution errors: %v", table.Errors[0])
	}

	tv := transformers.NewTransformerVisitor()
	transformed, err := tv.TransformModule(module, table)
	if err != nil {
		return fmt.Errorf("transformation failed: %w", err)
	}

	printer := compiler.NewASTPrinter("  ")
	output := printer.Print(transformed)

	if c.JSON {
		result := map[string]any{
			"file": filename,
			"ast":  output,
		}
		return printJSON(result)
	}

	fmt.Printf("=== %s (transformed) ===\n\n", filename)
	fmt.Println(output)
	return nil
}

// inspectCodegen shows the generated Python code.
func (c *InspectCmd) inspectCodegen(content []byte, filename string) error {
	module, errors := compiler.Parse(content)
	if module == nil {
		return formatParseErrors(errors)
	}

	res := resolver.NewResolver()
	table, err := res.Resolve(module)
	if err != nil {
		return fmt.Errorf("resolution failed: %w", err)
	}
	if len(table.Errors) > 0 {
		return fmt.Errorf("resolution errors: %v", table.Errors[0])
	}

	tv := transformers.NewTransformerVisitor()
	module, err = tv.TransformModule(module, table)
	if err != nil {
		return fmt.Errorf("transformation failed: %w", err)
	}

	generator := codegen.NewCodeGenerator()
	code := generator.Generate(module)

	if c.JSON {
		result := map[string]any{
			"file": filename,
			"code": code,
		}
		return printJSON(result)
	}

	fmt.Print(code)
	return nil
}

// countASTNodes counts the number of nodes in ASTPrinter output.
// Each node in the AST produces exactly one line in the printer output.
func countASTNodes(printerOutput string) int {
	trimmed := strings.TrimSpace(printerOutput)
	if trimmed == "" {
		return 0
	}
	return len(strings.Split(trimmed, "\n"))
}

// formatScopeType returns a human-readable name for a scope type.
func formatScopeType(scopeType resolver.ScopeType) string {
	switch scopeType {
	case resolver.ModuleScopeType:
		return "module"
	case resolver.FunctionScopeType:
		return "function"
	case resolver.ClassScopeType:
		return "class"
	case resolver.ViewScopeType:
		return "view"
	case resolver.ComprehensionScopeType:
		return "comprehension"
	case resolver.ExceptScopeType:
		return "except"
	case resolver.WithScopeType:
		return "with"
	default:
		return "unknown"
	}
}

// formatParseErrors returns a combined error from parse errors.
func formatParseErrors(errors []error) error {
	if len(errors) == 0 {
		return fmt.Errorf("parsing failed with no error details")
	}
	var msgs []string
	for _, e := range errors {
		msgs = append(msgs, e.Error())
	}
	return fmt.Errorf("parse errors:\n  %s", strings.Join(msgs, "\n  "))
}

// printJSON marshals v as indented JSON and prints it to stdout.
func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON encoding failed: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
