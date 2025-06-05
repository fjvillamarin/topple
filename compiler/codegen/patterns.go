package codegen

import (
	"sylfie/compiler/ast"
)

// Pattern visitors for match statements

func (cg *CodeGenerator) VisitLiteralPattern(lp *ast.LiteralPattern) ast.Visitor {
	lp.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitCapturePattern(cp *ast.CapturePattern) ast.Visitor {
	cp.Name.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitWildcardPattern(wp *ast.WildcardPattern) ast.Visitor {
	cg.write("_")
	return cg
}

func (cg *CodeGenerator) VisitValuePattern(vp *ast.ValuePattern) ast.Visitor {
	vp.Value.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitGroupPattern(gp *ast.GroupPattern) ast.Visitor {
	cg.write("(")
	gp.Pattern.Accept(cg)
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitSequencePattern(sp *ast.SequencePattern) ast.Visitor {
	cg.write("[")
	for i, pattern := range sp.Patterns {
		if i > 0 {
			cg.write(", ")
		}
		pattern.Accept(cg)
	}
	cg.write("]")
	return cg
}

func (cg *CodeGenerator) VisitStarPattern(sp *ast.StarPattern) ast.Visitor {
	cg.write("*")
	if sp.Pattern != nil {
		sp.Pattern.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitMappingPattern(mp *ast.MappingPattern) ast.Visitor {
	cg.write("{")
	for i, pair := range mp.Pairs {
		if i > 0 {
			cg.write(", ")
		}
		pair.Key.Accept(cg)
		cg.write(": ")
		pair.Pattern.Accept(cg)
	}
	if mp.DoubleStar != nil {
		if len(mp.Pairs) > 0 {
			cg.write(", ")
		}
		cg.write("**")
		mp.DoubleStar.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitClassPattern(cp *ast.ClassPattern) ast.Visitor {
	cp.Class.Accept(cg)
	cg.write("(")
	for i, pattern := range cp.Patterns {
		if i > 0 {
			cg.write(", ")
		}
		pattern.Accept(cg)
	}
	for i, kwPattern := range cp.KwdPatterns {
		if i > 0 || len(cp.Patterns) > 0 {
			cg.write(", ")
		}
		// cg.write(kwPattern.Arg)
		kwPattern.Name.Accept(cg)
		cg.write("=")
		kwPattern.Pattern.Accept(cg)
	}
	cg.write(")")
	return cg
}

func (cg *CodeGenerator) VisitAsPattern(ap *ast.AsPattern) ast.Visitor {
	if ap.Pattern != nil {
		ap.Pattern.Accept(cg)
		cg.write(" as ")
	}
	ap.Target.Accept(cg)
	return cg
}

func (cg *CodeGenerator) VisitOrPattern(op *ast.OrPattern) ast.Visitor {
	for i, pattern := range op.Patterns {
		if i > 0 {
			cg.write(" | ")
		}
		pattern.Accept(cg)
	}
	return cg
}
