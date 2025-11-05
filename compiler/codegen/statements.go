package codegen

import (
	"strings"
	"topple/compiler/ast"
)

// Statement visitors

func (cg *CodeGenerator) VisitExprStmt(e *ast.ExprStmt) ast.Visitor {
	e.Expr.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitTypeAlias(t *ast.TypeAlias) ast.Visitor {
	cg.write("type ")
	cg.write(t.Name.Lexeme)
	if len(t.Params) > 0 {
		cg.write("[")
		for i, param := range t.Params {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}
	cg.write(" = ")
	t.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitReturnStmt(r *ast.ReturnStmt) ast.Visitor {
	cg.write("return")
	if r.Value != nil {
		cg.write(" ")
		r.Value.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitRaiseStmt(r *ast.RaiseStmt) ast.Visitor {
	cg.write("raise")
	if r.HasException {
		cg.write(" ")
		r.Exception.Accept(cg)
		if r.HasFrom {
			cg.write(" from ")
			r.FromExpr.Accept(cg)
		}
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitPassStmt(p *ast.PassStmt) ast.Visitor {
	cg.write("pass")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitBreakStmt(b *ast.BreakStmt) ast.Visitor {
	cg.write("break")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitContinueStmt(c *ast.ContinueStmt) ast.Visitor {
	cg.write("continue")
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitYieldStmt(y *ast.YieldStmt) ast.Visitor {
	y.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAssertStmt(a *ast.AssertStmt) ast.Visitor {
	cg.write("assert ")
	a.Test.Accept(cg)
	if a.Message != nil {
		cg.write(", ")
		a.Message.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitGlobalStmt(g *ast.GlobalStmt) ast.Visitor {
	cg.write("global ")
	for i, name := range g.Names {
		if i > 0 {
			cg.write(", ")
		}
		name.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitNonlocalStmt(n *ast.NonlocalStmt) ast.Visitor {
	cg.write("nonlocal ")
	for i, name := range n.Names {
		if i > 0 {
			cg.write(", ")
		}
		name.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitImportStmt(i *ast.ImportStmt) ast.Visitor {
	cg.write("import ")
	for idx, alias := range i.Names {
		if idx > 0 {
			cg.write(", ")
		}
		cg.writeImportName(alias)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitImportFromStmt(i *ast.ImportFromStmt) ast.Visitor {
	cg.write("from ")
	if i.DotCount > 0 {
		cg.write(strings.Repeat(".", i.DotCount))
	}
	if i.DottedName != nil {
		cg.writeDottedName(i.DottedName)
	}
	cg.write(" import ")
	if i.IsWildcard {
		cg.write("*")
	} else {
		for idx, alias := range i.Names {
			if idx > 0 {
				cg.write(", ")
			}
			cg.writeImportName(alias)
		}
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAssignStmt(a *ast.AssignStmt) ast.Visitor {
	for i, target := range a.Targets {
		if i > 0 {
			cg.write(" = ")
		}
		target.Accept(cg)
	}
	cg.write(" = ")
	a.Value.Accept(cg)
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitAnnotationStmt(a *ast.AnnotationStmt) ast.Visitor {
	a.Target.Accept(cg)
	cg.write(": ")
	a.Type.Accept(cg)
	if a.HasValue {
		cg.write(" = ")
		a.Value.Accept(cg)
	}
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitMultiStmt(m *ast.MultiStmt) ast.Visitor {
	// MultiStmt should never reach codegen as it's unwrapped in the parser
	panic("MultiStmt should not reach codegen - it should be unwrapped in the parser")
}
