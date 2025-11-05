package parser

import (
	"strings"
	"testing"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// Helper function to parse an import statement
func parseImportStatement(t *testing.T, input string) (ast.Stmt, error) {
	scanner := lexer.NewScanner([]byte(input))
	tokens := scanner.ScanTokens()

	// Check for lexer errors
	if len(scanner.Errors) > 0 {
		t.Fatalf("Lexer errors encountered: %v", scanner.Errors)
	}

	parser := NewParser(tokens)
	stmt, err := parser.importStatement()

	// Check for parser errors
	if len(parser.Errors) > 0 {
		t.Fatalf("Parser errors encountered: %v", parser.Errors)
	}

	// Verify that all tokens were consumed (only EOF or Newline should remain)
	if err == nil && !parser.isAtEnd() {
		nextToken := parser.peek()
		// Allow Newline or EOF
		if nextToken.Type != lexer.Newline && nextToken.Type != lexer.EOF {
			// Create an error for unconsumed tokens
			return stmt, parser.error(nextToken, "unexpected token after import statement")
		}
	}

	return stmt, err
}

// Helper function to validate import statement type and structure
func validateImportStatement(t *testing.T, stmt ast.Stmt, expectedType string, expectedCount int) {
	switch expectedType {
	case "import":
		importStmt, ok := stmt.(*ast.ImportStmt)
		if !ok {
			t.Fatalf("Expected *ast.ImportStmt, got %T", stmt)
		}
		if len(importStmt.Names) != expectedCount {
			t.Errorf("Expected %d import names, got %d", expectedCount, len(importStmt.Names))
		}
	case "from":
		fromImport, ok := stmt.(*ast.ImportFromStmt)
		if !ok {
			t.Fatalf("Expected *ast.ImportFromStmt, got %T", stmt)
		}
		if expectedCount >= 0 && len(fromImport.Names) != expectedCount {
			t.Errorf("Expected %d import names, got %d", expectedCount, len(fromImport.Names))
		}
	default:
		t.Fatalf("Unknown expected type: %s", expectedType)
	}
}

// Helper function to validate aliases in import statements
func validateImportAlias(t *testing.T, stmt ast.Stmt, expectedHasAlias bool, expectedAlias string) {
	var hasAlias bool
	var aliasName string

	switch s := stmt.(type) {
	case *ast.ImportStmt:
		if len(s.Names) > 0 {
			hasAlias = s.Names[0].AsName != nil
			if hasAlias && s.Names[0].AsName != nil {
				aliasName = s.Names[0].AsName.Token.Lexeme
			}
		}
	case *ast.ImportFromStmt:
		if len(s.Names) > 0 {
			hasAlias = s.Names[0].AsName != nil
			if hasAlias && s.Names[0].AsName != nil {
				aliasName = s.Names[0].AsName.Token.Lexeme
			}
		}
	}

	if hasAlias != expectedHasAlias {
		t.Errorf("Expected hasAlias=%v, got %v", expectedHasAlias, hasAlias)
	}

	if expectedHasAlias && aliasName != expectedAlias {
		t.Errorf("Expected alias name %s, got %s", expectedAlias, aliasName)
	}
}

// Helper function to validate relative imports
func validateRelativeImport(t *testing.T, stmt ast.Stmt, expectedDotCount int) {
	fromImport, ok := stmt.(*ast.ImportFromStmt)
	if !ok {
		t.Fatalf("Expected *ast.ImportFromStmt for relative import, got %T", stmt)
	}

	if fromImport.DotCount != expectedDotCount {
		t.Errorf("Expected dot count %d, got %d", expectedDotCount, fromImport.DotCount)
	}

	if expectedDotCount > 0 && fromImport.DotCount == 0 {
		t.Error("Expected relative import to have dots")
	}
}

// Helper function to validate wildcard imports
func validateWildcardImport(t *testing.T, stmt ast.Stmt, expectedWildcard bool) {
	fromImport, ok := stmt.(*ast.ImportFromStmt)
	if !ok {
		t.Fatalf("Expected *ast.ImportFromStmt for wildcard check, got %T", stmt)
	}

	if fromImport.IsWildcard != expectedWildcard {
		t.Errorf("Expected IsWildcard=%v, got %v", expectedWildcard, fromImport.IsWildcard)
	}

	if expectedWildcard && len(fromImport.Names) != 0 {
		t.Errorf("Expected 0 names for wildcard import, got %d", len(fromImport.Names))
	}
}

// Helper function to validate dotted name structure
func validateDottedName(t *testing.T, stmt ast.Stmt, expectedParts int) {
	switch s := stmt.(type) {
	case *ast.ImportStmt:
		if len(s.Names) > 0 && s.Names[0].DottedName != nil {
			parts := len(s.Names[0].DottedName.Names)
			if parts != expectedParts {
				t.Errorf("Expected %d dotted name parts, got %d", expectedParts, parts)
			}
		}
	case *ast.ImportFromStmt:
		if s.DottedName != nil {
			parts := len(s.DottedName.Names)
			if parts != expectedParts {
				t.Errorf("Expected %d dotted name parts, got %d", expectedParts, parts)
			}
		}
	}
}

// Test comprehensive import statement functionality
func TestImportStatements(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		statementType string
		expectedCount int
		hasAlias      bool
		aliasName     string
		dottedParts   int
		description   string
	}{
		// Basic import statements
		{
			name:          "simple import",
			input:         "import os",
			statementType: "import",
			expectedCount: 1,
			dottedParts:   1,
			description:   "basic single module import",
		},
		{
			name:          "multiple imports",
			input:         "import os, sys, json",
			statementType: "import",
			expectedCount: 3,
			description:   "multiple modules in single import",
		},
		{
			name:          "import with alias",
			input:         "import numpy as np",
			statementType: "import",
			expectedCount: 1,
			hasAlias:      true,
			aliasName:     "np",
			dottedParts:   1,
			description:   "import with alias",
		},
		{
			name:          "multiple imports with aliases",
			input:         "import numpy as np, pandas as pd",
			statementType: "import",
			expectedCount: 2,
			hasAlias:      true,
			aliasName:     "np",
			description:   "multiple imports with aliases",
		},
		{
			name:          "dotted module import",
			input:         "import os.path",
			statementType: "import",
			expectedCount: 1,
			dottedParts:   2,
			description:   "import dotted module name",
		},
		{
			name:          "complex dotted import",
			input:         "import collections.abc.Mapping",
			statementType: "import",
			expectedCount: 1,
			dottedParts:   3,
			description:   "import deeply nested module",
		},
		{
			name:          "long dotted import with alias",
			input:         "import xml.etree.ElementTree as ET",
			statementType: "import",
			expectedCount: 1,
			hasAlias:      true,
			aliasName:     "ET",
			dottedParts:   3,
			description:   "complex dotted import with alias",
		},

		// Basic from imports
		{
			name:          "simple from import",
			input:         "from os import path",
			statementType: "from",
			expectedCount: 1,
			dottedParts:   1,
			description:   "basic from import",
		},
		{
			name:          "from import multiple",
			input:         "from collections import Counter, defaultdict",
			statementType: "from",
			expectedCount: 2,
			description:   "from import multiple names",
		},
		{
			name:          "from import with alias",
			input:         "from collections import Counter as C",
			statementType: "from",
			expectedCount: 1,
			hasAlias:      true,
			aliasName:     "C",
			description:   "from import with alias",
		},
		{
			name:          "from dotted module import",
			input:         "from collections.abc import Mapping",
			statementType: "from",
			expectedCount: 1,
			dottedParts:   2,
			description:   "from import from dotted module",
		},
		{
			name:          "from import mixed aliases",
			input:         "from typing import List, Dict as D, Optional",
			statementType: "from",
			expectedCount: 3,
			description:   "from import with mixed aliases",
		},

		// Error cases
		{
			name:        "empty import",
			input:       "import",
			hasError:    true,
			description: "import without module name",
		},
		{
			name:        "empty from import",
			input:       "from module import",
			hasError:    true,
			description: "from import without names",
		},
		{
			name:        "invalid alias syntax",
			input:       "import os as",
			hasError:    true,
			description: "import with incomplete alias",
		},
		{
			name:        "invalid from alias",
			input:       "from os import path as",
			hasError:    true,
			description: "from import with incomplete alias",
		},
		{
			name:        "missing module in from",
			input:       "from import path",
			hasError:    true,
			description: "from import without module",
		},
		{
			name:        "star in regular import",
			input:       "import *",
			hasError:    true,
			description: "wildcard in regular import",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseImportStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateImportStatement(t, stmt, test.statementType, test.expectedCount)

			if test.hasAlias {
				validateImportAlias(t, stmt, test.hasAlias, test.aliasName)
			}

			if test.dottedParts > 0 {
				validateDottedName(t, stmt, test.dottedParts)
			}
		})
	}
}

// Test relative imports
func TestRelativeImports(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasError     bool
		expectedDots int
		hasModule    bool
		description  string
	}{
		{
			name:         "current package import",
			input:        "from . import module",
			expectedDots: 1,
			hasModule:    false,
			description:  "import from current package",
		},
		{
			name:         "parent package import",
			input:        "from .. import module",
			expectedDots: 2,
			hasModule:    false,
			description:  "import from parent package",
		},
		{
			name:         "grandparent package import",
			input:        "from ... import module",
			expectedDots: 3,
			hasModule:    false,
			description:  "import from grandparent package",
		},
		{
			name:         "relative with module path",
			input:        "from ..utils import helper",
			expectedDots: 2,
			hasModule:    true,
			description:  "relative import with module path",
		},
		{
			name:         "deep relative with module",
			input:        "from ...config.settings import DATABASE_URL",
			expectedDots: 3,
			hasModule:    true,
			description:  "deep relative import with nested module",
		},
		{
			name:         "relative with alias",
			input:        "from .helpers import utility as util",
			expectedDots: 1,
			hasModule:    true,
			description:  "relative import with alias",
		},
		{
			name:         "relative multiple imports",
			input:        "from ..models import User, Post, Comment",
			expectedDots: 2,
			hasModule:    true,
			description:  "relative import of multiple names",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseImportStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateRelativeImport(t, stmt, test.expectedDots)

			fromImport := stmt.(*ast.ImportFromStmt)
			hasModule := fromImport.DottedName != nil
			if hasModule != test.hasModule {
				t.Errorf("Expected hasModule=%v for %s, got %v", test.hasModule, test.description, hasModule)
			}
		})
	}
}

// Test wildcard and star imports
func TestWildcardImports(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		hasError    bool
		description string
	}{
		{
			name:        "simple star import",
			input:       "from math import *",
			description: "wildcard import from standard module",
		},
		{
			name:        "dotted star import",
			input:       "from collections.abc import *",
			description: "wildcard import from dotted module",
		},
		{
			name:        "relative star import",
			input:       "from . import *",
			description: "wildcard import from current package",
		},
		{
			name:        "parent relative star import",
			input:       "from .. import *",
			description: "wildcard import from parent package",
		},
		{
			name:        "complex relative star import",
			input:       "from ...utils.helpers import *",
			description: "wildcard import from nested relative module",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseImportStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateWildcardImport(t, stmt, true)
		})
	}
}

// Test parenthesized imports
func TestParenthesizedImports(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		expectedCount int
		description   string
	}{
		{
			name: "simple parenthesized import",
			input: `from collections import (
    Counter,
    defaultdict
)`,
			expectedCount: 2,
			description:   "parenthesized import with multiple names",
		},
		{
			name: "parenthesized with aliases",
			input: `from collections import (
    Counter as C,
    defaultdict as dd,
    OrderedDict as OD
)`,
			expectedCount: 3,
			description:   "parenthesized import with aliases",
		},
		{
			name: "parenthesized with trailing comma",
			input: `from typing import (
    List,
    Dict,
    Optional,
)`,
			expectedCount: 3,
			description:   "parenthesized import with trailing comma",
		},
		{
			name: "single item parenthesized",
			input: `from os import (
    path
)`,
			expectedCount: 1,
			description:   "single name in parentheses",
		},
		{
			name: "complex parenthesized import",
			input: `from typing import (
    List,
    Dict as D,
    Optional,
    Union as U,
    Callable
)`,
			expectedCount: 5,
			description:   "complex parenthesized import with mixed aliases",
		},
		{
			name: "multiline parenthesized",
			input: `from django.contrib.auth import (
    authenticate,
    login,
    logout,
    get_user_model,
)`,
			expectedCount: 4,
			description:   "multiline parenthesized import",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseImportStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			validateImportStatement(t, stmt, "from", test.expectedCount)
		})
	}
}

// Test import edge cases and complex scenarios
func TestImportEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		hasError      bool
		errorContains string
		description   string
	}{
		{
			name:        "very long dotted import",
			input:       "import deeply.nested.package.module.submodule.component",
			hasError:    false,
			description: "import with many dotted components",
		},
		{
			name:        "import with unicode module name",
			input:       "import módulo",
			hasError:    false,
			description: "import with unicode characters",
		},
		{
			name:        "from import with many names",
			input:       "from module import a, b, c, d, e, f, g, h, i, j",
			hasError:    false,
			description: "from import with many names",
		},
		{
			name:        "complex alias names",
			input:       "import very_long_module_name as short",
			hasError:    false,
			description: "import with descriptive alias",
		},
		{
			name:        "relative import at max depth",
			input:       "from ............. import module",
			hasError:    false,
			description: "very deep relative import",
		},

		// Error cases
		{
			name:          "invalid relative import syntax",
			input:         "import ..module",
			hasError:      true,
			errorContains: "relative",
			description:   "relative syntax in regular import",
		},
		{
			name:        "incomplete parenthesized import",
			input:       "from module import (",
			hasError:    true,
			description: "unclosed parenthesized import",
		},
		{
			name:        "invalid comma usage",
			input:       "from module import a,, b",
			hasError:    true,
			description: "double comma in import list",
		},
		{
			name:          "missing import keyword",
			input:         "from module",
			hasError:      true,
			errorContains: "import",
			description:   "from without import keyword",
		},
		{
			name:        "invalid alias keyword",
			input:       "from module import name is alias",
			hasError:    true,
			description: "wrong alias keyword",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stmt, err := parseImportStatement(t, test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", test.description)
					return
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", test.description, err)
			}

			// For successful cases, just verify we got a valid statement
			if stmt == nil {
				t.Error("Statement should not be nil")
			}
		})
	}
}
