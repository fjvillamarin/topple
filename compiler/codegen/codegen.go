package codegen

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/ast"
	"strings"
)

type CodeGenerator struct {
	builder strings.Builder
	indent  int

	// Additional fields for proper code generation
	needsNewline bool
	atLineStart  bool

	ast.Visitor
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		atLineStart: true,
	}
}

// Generate generates Python code from the given AST node
func (cg *CodeGenerator) Generate(node ast.Node) string {
	cg.builder.Reset()
	cg.indent = 0
	cg.needsNewline = false
	cg.atLineStart = true

	node.Accept(cg)
	return cg.builder.String()
}

// Helper methods for formatting
func (cg *CodeGenerator) write(s string) {
	if cg.atLineStart && cg.indent > 0 && s != "\n" {
		cg.builder.WriteString(strings.Repeat("    ", cg.indent))
		cg.atLineStart = false
	}
	cg.builder.WriteString(s)
	if s == "\n" {
		cg.atLineStart = true
	}
}

func (cg *CodeGenerator) writef(format string, args ...interface{}) {
	cg.write(fmt.Sprintf(format, args...))
}

func (cg *CodeGenerator) newline() {
	cg.write("\n")
}

func (cg *CodeGenerator) increaseIndent() {
	cg.indent++
}

func (cg *CodeGenerator) decreaseIndent() {
	cg.indent--
}

func (cg *CodeGenerator) writeStmts(stmts []ast.Stmt) {
	for _, stmt := range stmts {
		stmt.Accept(cg)
	}
}

// Generic visit method - delegate to specific visitor
func (cg *CodeGenerator) Visit(node ast.Node) ast.Visitor {
	node.Accept(cg)
	return cg
}

// Module visitor
func (cg *CodeGenerator) VisitModule(m *ast.Module) ast.Visitor {
	cg.writeStmts(m.Body)
	return cg
}

// Helper method for writing import names
func (cg *CodeGenerator) writeImportName(name *ast.ImportName) {
	cg.writeDottedName(name.DottedName)
	if name.AsName != nil {
		cg.write(" as ")
		name.AsName.Accept(cg)
	}
}

// Helper method for writing dotted names
func (cg *CodeGenerator) writeDottedName(name *ast.DottedName) {
	for i, part := range name.Names {
		if i > 0 {
			cg.write(".")
		}
		part.Accept(cg)
	}
}

// Helper method for writing ForIfClause
func (cg *CodeGenerator) writeForIfClause(clause ast.ForIfClause) {
	if clause.IsAsync {
		cg.write("async ")
	}
	cg.write("for ")
	clause.Target.Accept(cg)
	cg.write(" in ")
	clause.Iter.Accept(cg)
	for _, ifCond := range clause.Ifs {
		cg.write(" if ")
		ifCond.Accept(cg)
	}
}
