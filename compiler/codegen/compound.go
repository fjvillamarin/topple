package codegen

import (
	"topple/compiler/ast"
)

// Compound statement visitors

func (cg *CodeGenerator) VisitIf(i *ast.If) ast.Visitor {
	cg.write("if ")
	i.Condition.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(i.Body)
	cg.decreaseIndent()

	if len(i.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(i.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitWhile(w *ast.While) ast.Visitor {
	cg.write("while ")
	w.Test.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(w.Body)
	cg.decreaseIndent()

	if len(w.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(w.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitFor(f *ast.For) ast.Visitor {
	if f.IsAsync {
		cg.write("async ")
	}
	cg.write("for ")
	f.Target.Accept(cg)
	cg.write(" in ")
	f.Iterable.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(f.Body)
	cg.decreaseIndent()

	if len(f.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(f.Else)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitWith(w *ast.With) ast.Visitor {
	if w.IsAsync {
		cg.write("async ")
	}
	cg.write("with ")
	for i, item := range w.Items {
		if i > 0 {
			cg.write(", ")
		}
		item.Expr.Accept(cg)
		if item.As != nil {
			cg.write(" as ")
			item.As.Accept(cg)
		}
	}
	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(w.Body)
	cg.decreaseIndent()
	return cg
}

func (cg *CodeGenerator) VisitTry(t *ast.Try) ast.Visitor {
	cg.write("try:")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(t.Body)
	cg.decreaseIndent()

	for _, handler := range t.Excepts {
		cg.write("except")
		if handler.Type != nil {
			cg.write(" ")
			handler.Type.Accept(cg)
			if handler.Name != nil {
				cg.write(" as ")
				handler.Name.Accept(cg)
			}
		}
		cg.write(":")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(handler.Body)
		cg.decreaseIndent()
	}

	if len(t.Else) > 0 {
		cg.write("else:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(t.Else)
		cg.decreaseIndent()
	}

	if len(t.Finally) > 0 {
		cg.write("finally:")
		cg.newline()
		cg.increaseIndent()
		cg.writeStmts(t.Finally)
		cg.decreaseIndent()
	}
	return cg
}

func (cg *CodeGenerator) VisitDecorator(d *ast.Decorator) ast.Visitor {
	cg.write("@")
	d.Expr.Accept(cg)
	cg.newline()
	d.Stmt.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitClass(c *ast.Class) ast.Visitor {
	cg.write("class ")
	c.Name.Accept(cg)

	if len(c.TypeParams) > 0 {
		cg.write("[")
		for i, param := range c.TypeParams {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}

	if len(c.Args) > 0 {
		cg.write("(")
		for i, base := range c.Args {
			if i > 0 {
				cg.write(", ")
			}
			base.Accept(cg)
		}
		cg.write(")")
	}

	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(c.Body)
	cg.decreaseIndent()
	return cg
}

func (cg *CodeGenerator) VisitFunction(f *ast.Function) ast.Visitor {
	if f.IsAsync {
		cg.write("async ")
	}
	cg.write("def ")
	f.Name.Accept(cg)

	if len(f.TypeParameters) > 0 {
		cg.write("[")
		for i, param := range f.TypeParameters {
			if i > 0 {
				cg.write(", ")
			}
			param.Accept(cg)
		}
		cg.write("]")
	}

	cg.write("(")
	if f.Parameters != nil {
		f.Parameters.Accept(cg)
	}
	cg.write(")")

	if f.ReturnType != nil {
		cg.write(" -> ")
		f.ReturnType.Accept(cg)
	}

	cg.write(":")
	cg.newline()
	cg.increaseIndent()
	cg.writeStmts(f.Body)
	cg.decreaseIndent()
	cg.newline()
	return cg
}

func (cg *CodeGenerator) VisitParameterList(p *ast.ParameterList) ast.Visitor {
	for i, param := range p.Parameters {
		if i > 0 {
			cg.write(", ")
		}
		param.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitParameter(p *ast.Parameter) ast.Visitor {
	if p.IsStar {
		cg.write("*")
	} else if p.IsDoubleStar {
		cg.write("**")
	}

	if p.Name != nil {
		p.Name.Accept(cg)
	}

	if p.Annotation != nil {
		cg.write(": ")
		p.Annotation.Accept(cg)
	}

	if p.Default != nil {
		cg.write("=")
		p.Default.Accept(cg)
	}

	return cg
}

func (cg *CodeGenerator) VisitMatch(m *ast.MatchStmt) ast.Visitor {
	cg.write("match ")
	m.Subject.Accept(cg)
	cg.write(":")
	cg.newline()
	cg.increaseIndent()

	for _, caseStmt := range m.Cases {
		for _, pattern := range caseStmt.Patterns {
			cg.write("case ")
			pattern.Accept(cg)
			if caseStmt.Guard != nil {
				cg.write(" if ")
				caseStmt.Guard.Accept(cg)
			}
			cg.write(":")
			cg.newline()
			cg.increaseIndent()
			cg.writeStmts(caseStmt.Body)
			cg.decreaseIndent()
		}
	}

	cg.decreaseIndent()
	return cg
}
