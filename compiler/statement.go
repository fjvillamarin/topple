package compiler

import (
	"fmt"
	"strings"
)

// StmtVisitor is the interface for visitors that traverse statements.
type StmtVisitor interface {
	VisitModule(m *Module) Visitor
	VisitExprStmt(e *ExprStmt) Visitor
	VisitTypeAlias(t *TypeAlias) Visitor
	VisitReturnStmt(r *ReturnStmt) Visitor
	VisitRaiseStmt(r *RaiseStmt) Visitor
	VisitPassStmt(p *PassStmt) Visitor
	VisitBreakStmt(b *BreakStmt) Visitor
	VisitContinueStmt(c *ContinueStmt) Visitor
	VisitYieldStmt(y *YieldStmt) Visitor
	VisitAssertStmt(a *AssertStmt) Visitor
	VisitGlobalStmt(g *GlobalStmt) Visitor
	VisitNonlocalStmt(n *NonlocalStmt) Visitor
	VisitImportStmt(i *ImportStmt) Visitor
	VisitImportFromStmt(i *ImportFromStmt) Visitor
	VisitAssignStmt(a *AssignStmt) Visitor
	VisitAugAssignStmt(a *AugAssignStmt) Visitor
	VisitAnnotationStmt(a *AnnotationStmt) Visitor
	VisitMultiStmt(m *MultiStmt) Visitor
}

// Module is the root node of a program, containing a list of statements.
type Module struct {
	BaseNode
	Body []Stmt
}

func NewModule(body []Stmt, startPos Position, endPos Position) *Module {
	return &Module{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Body: body,
	}
}

func (m *Module) isStmt() {}

// Accept calls the VisitModule method on the visitor
func (m *Module) Accept(visitor Visitor) {
	visitor.VisitModule(m)
}

func (m *Module) String() string {
	return fmt.Sprintf("Module(%d stmts)", len(m.Body))
}

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	BaseNode
	Value Expr
}

func NewExprStmt(value Expr, startPos Position, endPos Position) *ExprStmt {
	return &ExprStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Value: value,
	}
}

func (e *ExprStmt) isStmt() {}

// Accept calls the VisitExprStmt method on the visitor
func (e *ExprStmt) Accept(visitor Visitor) {
	visitor.VisitExprStmt(e)
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%s)", e.Value)
}

// TypeAlias represents a 'type' statement.
type TypeAlias struct {
	BaseNode
	Name   Token
	Params []Expr // Type parameters as expressions
	Value  Expr   // The type value expression
}

func NewTypeAlias(name Token, params []Expr, value Expr, startPos Position, endPos Position) *TypeAlias {
	return &TypeAlias{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Name:   name,
		Params: params,
		Value:  value,
	}
}

func (t *TypeAlias) isStmt() {}

// Accept calls the VisitTypeAlias method on the visitor
func (t *TypeAlias) Accept(visitor Visitor) {
	visitor.VisitTypeAlias(t)
}

func (t *TypeAlias) String() string {
	return fmt.Sprintf("TypeAlias(%s)", t.Name.Lexeme)
}

// ReturnStmt represents a 'return' statement.
type ReturnStmt struct {
	BaseNode
	Value Expr
}

func NewReturnStmt(value Expr, startPos Position, endPos Position) *ReturnStmt {
	return &ReturnStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Value: value,
	}
}

func (r *ReturnStmt) isStmt() {}

func (r *ReturnStmt) Accept(visitor Visitor) {
	visitor.VisitReturnStmt(r)
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt(%s)", r.Value)
}

// RaiseStmt represents a 'raise' statement.
type RaiseStmt struct {
	BaseNode
	Exception    Expr
	FromExpr     Expr
	HasException bool
	HasFrom      bool
}

func NewRaiseStmt(exception Expr, fromExpr Expr, hasException bool, hasFrom bool, startPos Position, endPos Position) *RaiseStmt {
	return &RaiseStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Exception:    exception,
		FromExpr:     fromExpr,
		HasException: hasException,
		HasFrom:      hasFrom,
	}
}

func (r *RaiseStmt) isStmt() {}

func (r *RaiseStmt) Accept(visitor Visitor) {
	visitor.VisitRaiseStmt(r)
}

func (r *RaiseStmt) String() string {
	if r.HasException {
		if r.HasFrom {
			return fmt.Sprintf("RaiseStmt(%s from %s)", r.Exception, r.FromExpr)
		}
		return fmt.Sprintf("RaiseStmt(%s)", r.Exception)
	}
	return "RaiseStmt()"
}

// PassStmt represents a 'pass' statement, which is a no-op.
type PassStmt struct {
	BaseNode
}

func NewPassStmt(startPos Position, endPos Position) *PassStmt {
	return &PassStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (p *PassStmt) isStmt() {}

func (p *PassStmt) Accept(visitor Visitor) {
	visitor.VisitPassStmt(p)
}

func (p *PassStmt) String() string {
	return "PassStmt()"
}

// BreakStmt represents a 'break' statement.
type BreakStmt struct {
	BaseNode
}

func NewBreakStmt(startPos Position, endPos Position) *BreakStmt {
	return &BreakStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (b *BreakStmt) isStmt() {}

func (b *BreakStmt) Accept(visitor Visitor) {
	visitor.VisitBreakStmt(b)
}

func (b *BreakStmt) String() string {
	return "BreakStmt()"
}

// ContinueStmt represents a 'continue' statement.
type ContinueStmt struct {
	BaseNode
}

func NewContinueStmt(startPos Position, endPos Position) *ContinueStmt {
	return &ContinueStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
	}
}

func (c *ContinueStmt) isStmt() {}

func (c *ContinueStmt) Accept(visitor Visitor) {
	visitor.VisitContinueStmt(c)
}

func (c *ContinueStmt) String() string {
	return "ContinueStmt()"
}

// YieldStmt represents a 'yield' statement.
type YieldStmt struct {
	BaseNode
	Value Expr
}

func NewYieldStmt(value Expr, startPos Position, endPos Position) *YieldStmt {
	return &YieldStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Value: value,
	}
}

func (y *YieldStmt) isStmt() {}

func (y *YieldStmt) Accept(visitor Visitor) {
	visitor.VisitYieldStmt(y)
}

func (y *YieldStmt) String() string {
	return fmt.Sprintf("YieldStmt(%s)", y.Value)
}

// AssertStmt represents an 'assert' statement.
type AssertStmt struct {
	BaseNode
	Test    Expr
	Message Expr // Optional error message
}

func NewAssertStmt(test Expr, message Expr, startPos Position, endPos Position) *AssertStmt {
	return &AssertStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Test:    test,
		Message: message,
	}
}

func (a *AssertStmt) isStmt() {}

func (a *AssertStmt) Accept(visitor Visitor) {
	visitor.VisitAssertStmt(a)
}

func (a *AssertStmt) String() string {
	if a.Message != nil {
		return fmt.Sprintf("AssertStmt(%s, %s)", a.Test, a.Message)
	}
	return fmt.Sprintf("AssertStmt(%s)", a.Test)
}

// GlobalStmt represents a 'global' statement.
type GlobalStmt struct {
	BaseNode
	Names []*Name // The identifiers declared as global
}

func NewGlobalStmt(names []*Name, startPos Position, endPos Position) *GlobalStmt {
	return &GlobalStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Names: names,
	}
}

func (g *GlobalStmt) isStmt() {}

func (g *GlobalStmt) Accept(visitor Visitor) {
	visitor.VisitGlobalStmt(g)
}

func (g *GlobalStmt) String() string {
	names := make([]string, len(g.Names))
	for i, name := range g.Names {
		names[i] = name.String()
	}
	return fmt.Sprintf("GlobalStmt(%s)", strings.Join(names, ", "))
}

// NonlocalStmt represents a 'nonlocal' statement.
type NonlocalStmt struct {
	BaseNode
	Names []*Name // The identifiers declared as nonlocal
}

func NewNonlocalStmt(names []*Name, startPos Position, endPos Position) *NonlocalStmt {
	return &NonlocalStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Names: names,
	}
}

func (n *NonlocalStmt) isStmt() {}

func (n *NonlocalStmt) Accept(visitor Visitor) {
	visitor.VisitNonlocalStmt(n)
}

func (n *NonlocalStmt) String() string {
	names := make([]string, len(n.Names))
	for i, name := range n.Names {
		names[i] = name.String()
	}
	return fmt.Sprintf("NonlocalStmt(%s)", strings.Join(names, ", "))
}

// ImportStmt represents an 'import' statement.
type ImportStmt struct {
	BaseNode
	Names []*ImportName // List of imported modules with optional aliases
}

func NewImportStmt(names []*ImportName, startPos Position, endPos Position) *ImportStmt {
	return &ImportStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Names: names,
	}
}

func (i *ImportStmt) isStmt() {}

func (i *ImportStmt) Accept(visitor Visitor) {
	visitor.VisitImportStmt(i)
}

func (i *ImportStmt) String() string {
	names := make([]string, len(i.Names))
	for j, name := range i.Names {
		names[j] = name.String()
	}
	return fmt.Sprintf("ImportStmt(%s)", strings.Join(names, ", "))
}

// ImportName represents a single module import with optional alias.
type ImportName struct {
	BaseNode
	DottedName *DottedName // The module path
	AsName     *Name       // Optional alias name
}

func NewImportName(dottedName *DottedName, asName *Name, startPos Position, endPos Position) *ImportName {
	return &ImportName{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		DottedName: dottedName,
		AsName:     asName,
	}
}

func (i *ImportName) String() string {
	if i.AsName != nil {
		return fmt.Sprintf("%s as %s", i.DottedName, i.AsName)
	}
	return i.DottedName.String()
}

// DottedName represents a dotted module path.
type DottedName struct {
	BaseNode
	Names []*Name // Parts of the dotted path
}

func NewDottedName(names []*Name, startPos Position, endPos Position) *DottedName {
	return &DottedName{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Names: names,
	}
}

func (d *DottedName) String() string {
	parts := make([]string, len(d.Names))
	for i, name := range d.Names {
		parts[i] = name.String()
	}
	return strings.Join(parts, ".")
}

// ImportFromStmt represents a 'from ... import ...' statement.
type ImportFromStmt struct {
	BaseNode
	DottedName *DottedName   // The module path to import from (may be nil for relative imports)
	DotCount   int           // Number of leading dots for relative imports
	Names      []*ImportName // List of imported names with optional aliases
	IsWildcard bool          // True if importing '*'
}

func NewImportFromStmt(dottedName *DottedName, dotCount int, names []*ImportName, isWildcard bool, startPos Position, endPos Position) *ImportFromStmt {
	return &ImportFromStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		DottedName: dottedName,
		DotCount:   dotCount,
		Names:      names,
		IsWildcard: isWildcard,
	}
}

func (i *ImportFromStmt) isStmt() {}

func (i *ImportFromStmt) Accept(visitor Visitor) {
	visitor.VisitImportFromStmt(i)
}

func (i *ImportFromStmt) String() string {
	var module string
	if i.DottedName != nil {
		module = i.DottedName.String()
	} else {
		module = strings.Repeat(".", i.DotCount)
	}

	if i.IsWildcard {
		return fmt.Sprintf("ImportFromStmt(%s, *)", module)
	}

	names := make([]string, len(i.Names))
	for j, name := range i.Names {
		names[j] = name.String()
	}
	return fmt.Sprintf("ImportFromStmt(%s, [%s])", module, strings.Join(names, ", "))
}

// AssignStmt represents an assignment statement like 'x = y' or 'a, b = c, d'.
type AssignStmt struct {
	BaseNode
	Targets []Expr // Left-hand side targets (can be multiple for unpacking)
	Value   Expr   // Right-hand side expression
}

// NewAssignStmt creates a new assignment statement.
func NewAssignStmt(targets []Expr, value Expr, startPos Position, endPos Position) *AssignStmt {
	return &AssignStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Targets: targets,
		Value:   value,
	}
}

func (a *AssignStmt) isStmt() {}

func (a *AssignStmt) Accept(visitor Visitor) {
	visitor.VisitAssignStmt(a)
}

func (a *AssignStmt) String() string {
	var targetStrs []string
	for _, target := range a.Targets {
		targetStrs = append(targetStrs, target.String())
	}
	return fmt.Sprintf("AssignStmt(%s = %s)", strings.Join(targetStrs, ", "), a.Value)
}

// AugAssignStmt represents an augmented assignment statement like 'x += y'.
type AugAssignStmt struct {
	BaseNode
	Target   Expr  // Left-hand side target
	Operator Token // The augmented assignment operator
	Value    Expr  // Right-hand side expression
}

// NewAugAssignStmt creates a new augmented assignment statement.
func NewAugAssignStmt(target Expr, operator Token, value Expr, startPos Position, endPos Position) *AugAssignStmt {
	return &AugAssignStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Target:   target,
		Operator: operator,
		Value:    value,
	}
}

func (a *AugAssignStmt) isStmt() {}

func (a *AugAssignStmt) Accept(visitor Visitor) {
	visitor.VisitAugAssignStmt(a)
}

func (a *AugAssignStmt) String() string {
	return fmt.Sprintf("AugAssignStmt(%s %s %s)", a.Target, a.Operator.Lexeme, a.Value)
}

// AnnotationStmt represents a variable annotation like 'x: int' or 'x: int = 5'.
type AnnotationStmt struct {
	BaseNode
	Target   Expr // The variable or target being annotated
	Type     Expr // The type annotation expression
	Value    Expr // Optional initializer value (can be nil)
	HasValue bool // Whether an initializer value is present
}

// NewAnnotationStmt creates a new annotation statement.
func NewAnnotationStmt(target Expr, typeExpr Expr, value Expr, hasValue bool, startPos Position, endPos Position) *AnnotationStmt {
	return &AnnotationStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Target:   target,
		Type:     typeExpr,
		Value:    value,
		HasValue: hasValue,
	}
}

func (a *AnnotationStmt) isStmt() {}

func (a *AnnotationStmt) Accept(visitor Visitor) {
	visitor.VisitAnnotationStmt(a)
}

func (a *AnnotationStmt) String() string {
	if a.HasValue {
		return fmt.Sprintf("AnnotationStmt(%s: %s = %s)", a.Target, a.Type, a.Value)
	}
	return fmt.Sprintf("AnnotationStmt(%s: %s)", a.Target, a.Type)
}

// MultiStmt represents a list of statements.
// It doesn't have semantic meaning, it's just a container for a list of statements.
type MultiStmt struct {
	BaseNode
	Stmts []Stmt
}

func NewMultiStmt(stmts []Stmt, startPos Position, endPos Position) *MultiStmt {
	return &MultiStmt{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Stmts: stmts,
	}
}

func (m *MultiStmt) isStmt() {}

func (m *MultiStmt) Accept(visitor Visitor) {
	visitor.VisitMultiStmt(m)
}

func (m *MultiStmt) String() string {
	return fmt.Sprintf("MultiStmt(%d stmts)", len(m.Stmts))
}
