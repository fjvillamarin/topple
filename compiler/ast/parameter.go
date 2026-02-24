package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// Parameter represents a function parameter
type Parameter struct {
	Name          *Name // Parameter name
	Annotation    Expr  // Optional type annotation (:Type)
	Default       Expr  // Optional default value (=default)
	IsStar        bool  // Whether this is a *args parameter
	IsDoubleStar  bool  // Whether this is a **kwargs parameter
	IsSlash       bool  // Whether this is a positional-only parameter (before /)
	IsKeywordOnly bool  // Whether this is a keyword-only parameter (after * or *args)

	Span lexer.Span
}

func (p *Parameter) isExpr() {}

func (p *Parameter) GetSpan() lexer.Span {
	return p.Span
}

// Accept implements the Node interface by delegating to Visit
func (p *Parameter) Accept(visitor Visitor) {
	visitor.VisitParameter(p)
}

func (p *Parameter) String() string {
	var prefix string
	if p.IsStar {
		prefix = "*"
	} else if p.IsDoubleStar {
		prefix = "**"
	}

	var annotation string
	if p.Annotation != nil {
		annotation = ": " + p.Annotation.String()
	}

	var defaultValue string
	if p.Default != nil {
		defaultValue = " = " + p.Default.String()
	}

	var suffix string
	if p.IsSlash {
		suffix = " /"
	}

	return fmt.Sprintf("%s%s%s%s%s", prefix, p.Name.String(), annotation, defaultValue, suffix)
}

// ParameterList represents a list of parameters in a function definition
type ParameterList struct {
	Parameters  []*Parameter // Regular parameters
	HasSlash    bool         // Whether the parameter list has a / (positional-only separator)
	SlashIndex  int          // Index of the / separator (-1 if not present)
	HasVarArg   bool         // Whether the parameter list has a *args parameter
	VarArgIndex int          // Index of the *args parameter (-1 if not present)
	HasKwArg    bool         // Whether the parameter list has a **kwargs parameter
	KwArgIndex  int          // Index of the **kwargs parameter (-1 if not present)

	Span lexer.Span
}

func (p *ParameterList) isExpr() {}

func (p *ParameterList) GetSpan() lexer.Span {
	return p.Span
}

// Accept implements the Node interface by delegating to Visit
func (p *ParameterList) Accept(visitor Visitor) {
	visitor.VisitParameterList(p)
}

func (p *ParameterList) String() string {
	var result string

	for i, param := range p.Parameters {
		if i > 0 {
			result += ", "
		}
		result += param.String()
	}

	return result
}
