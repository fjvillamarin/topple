package codegen

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/ast"
	"strconv"
)

// Expression visitors

func (cg *CodeGenerator) VisitName(n *ast.Name) ast.Visitor {
	cg.write(n.Token.Lexeme)
	return cg
}

func (cg *CodeGenerator) VisitLiteral(l *ast.Literal) ast.Visitor {
	// Handle the case where literal type might be wrong - check actual value type
	switch v := l.Value.(type) {
	case string:
		if l.Type == ast.LiteralTypeString {
			// Check if this is a raw string by looking at the lexeme
			lexeme := l.Token.Lexeme
			if len(lexeme) > 0 && (lexeme[0] == 'r' || lexeme[0] == 'R') {
				// Raw string - output the lexeme directly (includes 'r' prefix)
				cg.write(lexeme)
			} else {
				// Normal string - use strconv.Quote to properly escape all special characters
				cg.write(strconv.Quote(v))
			}
		} else {
			// String value but wrong type - treat as string anyway
			cg.write(strconv.Quote(v))
		}
	case int:
		cg.write(fmt.Sprintf("%d", v))
	case int64:
		cg.write(fmt.Sprintf("%d", v))
	case float64:
		cg.write(fmt.Sprintf("%g", v))
	case bool:
		if v {
			cg.write("True")
		} else {
			cg.write("False")
		}
	default:
		// Fallback for other types
		if l.Type == ast.LiteralTypeNone {
			cg.write("None")
		} else {
			cg.write(fmt.Sprintf("%v", v))
		}
	}
	return cg
}

func (cg *CodeGenerator) VisitAttribute(a *ast.Attribute) ast.Visitor {
	a.Object.Accept(cg)
	cg.write(".")
	cg.write(a.Name.Lexeme)
	return cg
}

func (cg *CodeGenerator) VisitCall(c *ast.Call) ast.Visitor {
	c.Callee.Accept(cg)
	cg.write("(")
	for i, arg := range c.Arguments {
		if i > 0 {
			cg.write(", ")
		}
		arg.Accept(cg)
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSubscript(s *ast.Subscript) ast.Visitor {
	s.Object.Accept(cg)
	cg.write("[")
	for i, index := range s.Indices {
		if i > 0 {
			cg.write(", ")
		}
		index.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitBinary(b *ast.Binary) ast.Visitor {
	b.Left.Accept(cg)
	cg.writef(" %s ", b.Operator.Lexeme)
	b.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitUnary(u *ast.Unary) ast.Visitor {
	cg.write(u.Operator.Lexeme)
	if u.Operator.Lexeme == "not" {
		cg.write(" ")
	}
	u.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitAssignExpr(a *ast.AssignExpr) ast.Visitor {
	a.Left.Accept(cg)
	cg.write(" := ")
	a.Right.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitStarExpr(s *ast.StarExpr) ast.Visitor {
	cg.write("*")
	s.Expr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitTernaryExpr(t *ast.TernaryExpr) ast.Visitor {
	t.TrueExpr.Accept(cg)
	cg.write(" if ")
	t.Condition.Accept(cg)
	cg.write(" else ")
	t.FalseExpr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitListExpr(l *ast.ListExpr) ast.Visitor {
	cg.write("[")
	for i, elem := range l.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitTupleExpr(t *ast.TupleExpr) ast.Visitor {
	cg.write("(")
	for i, elem := range t.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	if len(t.Elements) == 1 {
		cg.write(",")
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSetExpr(s *ast.SetExpr) ast.Visitor {
	// Special case: empty set must use set() constructor
	if len(s.Elements) == 0 {
		cg.write("set()")
		return cg
	}

	// Non-empty sets use {} syntax
	cg.write("{")
	for i, elem := range s.Elements {
		if i > 0 {
			cg.write(", ")
		}
		elem.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitDictExpr(d *ast.DictExpr) ast.Visitor {
	cg.write("{")
	for i, pair := range d.Pairs {
		if i > 0 {
			cg.write(", ")
		}
		switch p := pair.(type) {
		case *ast.KeyValuePair:
			p.Key.Accept(cg)
			cg.write(": ")
			p.Value.Accept(cg)
		case *ast.DoubleStarredPair:
			cg.write("**")
			p.Expr.Accept(cg)
		}
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitListComp(lc *ast.ListComp) ast.Visitor {
	cg.write("[")
	lc.Element.Accept(cg)
	for _, clause := range lc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitSetComp(sc *ast.SetComp) ast.Visitor {
	cg.write("{")
	sc.Element.Accept(cg)
	for _, clause := range sc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitDictComp(dc *ast.DictComp) ast.Visitor {
	cg.write("{")
	dc.Key.Accept(cg)
	cg.write(": ")
	dc.Value.Accept(cg)
	for _, clause := range dc.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitGenExpr(ge *ast.GenExpr) ast.Visitor {
	cg.write("(")
	ge.Element.Accept(cg)
	for _, clause := range ge.Clauses {
		cg.write(" ")
		cg.writeForIfClause(clause)
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitYieldExpr(y *ast.YieldExpr) ast.Visitor {
	cg.write("yield")
	if y.Value != nil {
		cg.write(" ")
		y.Value.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitGroupExpr(g *ast.GroupExpr) ast.Visitor {
	cg.write("(")
	g.Expression.Accept(cg)
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitTypeParamExpr(t *ast.TypeParam) ast.Visitor {
	if t.IsStar {
		cg.write("*")
	} else if t.IsDoubleStar {
		cg.write("**")
	}
	cg.write(t.Name.Lexeme)
	if t.Bound != nil {
		cg.write(": ")
		t.Bound.Accept(cg)
	}
	if t.Default != nil {
		cg.write(" = ")
		t.Default.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitSlice(s *ast.Slice) ast.Visitor {
	if s.StartIndex != nil {
		s.StartIndex.Accept(cg)
	}
	cg.write(":")
	if s.EndIndex != nil {
		s.EndIndex.Accept(cg)
	}
	if s.Step != nil {
		cg.write(":")
		s.Step.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitAwaitExpr(a *ast.AwaitExpr) ast.Visitor {
	cg.write("await ")
	a.Expr.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitArgument(a *ast.Argument) ast.Visitor {
	if a.Name != nil {
		a.Name.Accept(cg)
		cg.write("=")
	}
	if a.IsStar {
		cg.write("*")
	} else if a.IsDoubleStar {
		cg.write("**")
	}
	a.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitLambda(l *ast.Lambda) ast.Visitor {
	cg.write("lambda")
	if l.Parameters != nil && len(l.Parameters.Parameters) > 0 {
		cg.write(" ")
		l.Parameters.Accept(cg)
	}
	cg.write(": ")
	l.Body.Accept(cg)
	return cg
}
