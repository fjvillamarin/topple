package codegen

import (
	"topple/compiler/ast"
)

// PSX-specific visitors that should not be called after mutation
// These nodes should be transformed before code generation

func (cg *CodeGenerator) VisitViewStmt(v *ast.ViewStmt) ast.Visitor {
	// This should never be called since ViewStmt nodes are transformed to Class nodes by the mutator
	panic("ViewStmt nodes should be transformed to Class nodes before code generation")
}

func (cg *CodeGenerator) VisitHTMLElement(h *ast.HTMLElement) ast.Visitor {
	// This should never be called since HTML elements are transformed by the mutator
	panic("HTMLElement nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLContent(h *ast.HTMLContent) ast.Visitor {
	// This should never be called since HTML content is transformed by the mutator
	panic("HTMLContent nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLText(h *ast.HTMLText) ast.Visitor {
	// This should never be called since HTML text is transformed by the mutator
	panic("HTMLText nodes should be transformed before code generation")
}

func (cg *CodeGenerator) VisitHTMLInterpolation(h *ast.HTMLInterpolation) ast.Visitor {
	// This should never be called since HTML interpolation is transformed by the mutator
	panic("HTMLInterpolation nodes should be transformed before code generation")
}
