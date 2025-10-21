package codegen

import (
	"strings"
	"topple/compiler/ast"
)

// F-string visitors

func (cg *CodeGenerator) VisitFString(f *ast.FString) ast.Visitor {
	cg.write("f\"")
	for _, part := range f.Parts {
		part.Accept(cg)
	}
	cg.write("\"")
	return cg
}

func (cg *CodeGenerator) VisitFStringMiddle(f *ast.FStringMiddle) ast.Visitor {
	// Escape special characters for f-string content
	value := f.Value
	// Escape backslashes first
	value = strings.ReplaceAll(value, "\\", "\\\\")
	// Escape curly braces
	value = strings.ReplaceAll(value, "{", "{{")
	value = strings.ReplaceAll(value, "}", "}}")
	// Escape newlines
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	cg.write(value)
	return cg
}

func (cg *CodeGenerator) VisitFStringReplacementField(f *ast.FStringReplacementField) ast.Visitor {
	cg.write("{")
	f.Expression.Accept(cg)
	if f.Conversion != nil {
		f.Conversion.Accept(cg)
	}
	if f.FormatSpec != nil {
		f.FormatSpec.Accept(cg)
	}
	cg.write("}")
	return cg
}

func (cg *CodeGenerator) VisitFStringConversion(f *ast.FStringConversion) ast.Visitor {
	cg.write("!")
	cg.write(f.Type)
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatSpec(f *ast.FStringFormatSpec) ast.Visitor {
	cg.write(":")
	for _, part := range f.Spec {
		part.Accept(cg)
	}
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatMiddle(f *ast.FStringFormatMiddle) ast.Visitor {
	// Escape special characters for f-string format spec content
	value := f.Value
	// Escape backslashes first
	value = strings.ReplaceAll(value, "\\", "\\\\")
	// Escape curly braces
	value = strings.ReplaceAll(value, "{", "{{")
	value = strings.ReplaceAll(value, "}", "}}")
	// Escape newlines
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	cg.write(value)
	return cg
}

func (cg *CodeGenerator) VisitFStringFormatReplacementField(f *ast.FStringFormatReplacementField) ast.Visitor {
	cg.write("{")
	f.Expression.Accept(cg)
	cg.write("}")
	return cg
}
