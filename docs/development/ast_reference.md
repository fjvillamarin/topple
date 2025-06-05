# Sylfie AST Reference

This document provides a comprehensive reference for all AST (Abstract Syntax Tree) node types used in the Sylfie compiler.

## Overview

The Sylfie AST extends Python's AST with additional nodes for view definitions and HTML elements. All AST nodes implement the visitor pattern for traversal and transformation.

## Base Types

### Node Interface

All AST nodes implement:

```go
type Node interface {
    Accept(visitor Visitor) Visitor
    GetSpan() lexer.Span
    String() string
}
```

### Expression vs Statement

```go
type Expr interface {
    Node
    isExpr()  // Marker method
}

type Stmt interface {
    Node
    isStmt()  // Marker method
}
```

## Module and Top-Level

### Module

The root node of any parsed file:

```go
type Module struct {
    Body []Stmt
    Span lexer.Span
}
```

## Statements

### View Statement (Sylfie-specific)

```go
type ViewStmt struct {
    Name       *Name
    Params     *ParameterList
    ReturnType Expr  // Optional return type annotation
    Body       []Stmt
    Span       lexer.Span
}
```

Example:
```python
view HelloWorld(name: str = "World"):
    <div>Hello, {name}!</div>
```

### HTML Element (Sylfie-specific)

```go
type HTMLElement struct {
    TagName    lexer.Token      // Tag name token
    Attributes []HTMLAttribute  // HTML attributes
    Content    []Stmt          // Child elements/content
    IsSelfClosing bool
    Span       lexer.Span
}

type HTMLAttribute struct {
    Name  lexer.Token
    Value Expr  // Can be string literal, identifier, etc.
    Span  lexer.Span
}
```

### Assignment Statement

```go
type AssignStmt struct {
    Targets []Expr  // Can be Name, Subscript, Attribute, etc.
    Value   Expr
    Span    lexer.Span
}
```

Examples:
```python
x = 42
a, b = 1, 2
obj.attr = value
items[0] = "first"
```

### Expression Statement

```go
type ExprStmt struct {
    Expr Expr
    Span lexer.Span
}
```

### Function Definition

```go
type Function struct {
    Name       *Name
    Parameters *ParameterList
    ReturnType Expr         // Optional
    Body       []Stmt
    Decorators []*Decorator
    IsAsync    bool
    Span       lexer.Span
}

type ParameterList struct {
    Parameters   []*Parameter
    HasVarArg    bool  // Has *args
    HasKwArg     bool  // Has **kwargs
    HasSlash     bool  // Has / (positional-only)
    HasStar      bool  // Has * (keyword-only)
    SlashIndex   int   // Position of /
    StarIndex    int   // Position of *
    Span         lexer.Span
}

type Parameter struct {
    Name       *Name
    Annotation Expr   // Type annotation
    Default    Expr   // Default value
    IsStar     bool   // Is *args
    IsDoubleStar bool // Is **kwargs
    Span       lexer.Span
}
```

### Class Definition

```go
type Class struct {
    Name       *Name
    Args       []*Argument  // Base classes as arguments
    Body       []Stmt
    Decorators []*Decorator
    Span       lexer.Span
}
```

### Control Flow Statements

#### If Statement
```go
type If struct {
    Condition Expr
    Body      []Stmt
    Else      []Stmt  // Can contain elif chains
    Span      lexer.Span
}
```

#### For Loop
```go
type For struct {
    Target   Expr    // Loop variable(s)
    Iterable Expr    // What to iterate over
    Body     []Stmt
    Else     []Stmt  // Optional else clause
    IsAsync  bool
    Span     lexer.Span
}
```

#### While Loop
```go
type While struct {
    Test Expr
    Body []Stmt
    Else []Stmt
    Span lexer.Span
}
```

#### Match Statement (Python 3.10+)
```go
type MatchStmt struct {
    Subject Expr
    Cases   []CaseBlock
    Span    lexer.Span
}

type CaseBlock struct {
    Patterns []Pattern
    Guard    Expr    // Optional 'if' condition
    Body     []Stmt
    Span     lexer.Span
}
```

### Exception Handling

```go
type Try struct {
    Body     []Stmt
    Excepts  []Except
    Else     []Stmt
    Finally  []Stmt
    Span     lexer.Span
}

type Except struct {
    Type  Expr    // Exception type (optional)
    Name  *Name   // Variable name (optional)
    Body  []Stmt
    Span  lexer.Span
}
```

### Import Statements

```go
type ImportStmt struct {
    Names []*ImportName
    Span  lexer.Span
}

type ImportFromStmt struct {
    DottedName *DottedName  // Module path
    Names      []*ImportName
    DotCount   int  // Relative import level
    Span       lexer.Span
}

type ImportName struct {
    DottedName *DottedName
    AsName     *Name  // Optional alias
    Span       lexer.Span
}
```

### Other Statements

- `ReturnStmt` - Return from function
- `YieldStmt` - Yield expression as statement  
- `RaiseStmt` - Raise exception
- `BreakStmt` - Break from loop
- `ContinueStmt` - Continue to next iteration
- `PassStmt` - No-op statement
- `GlobalStmt` - Global variable declaration
- `NonlocalStmt` - Nonlocal variable declaration
- `AssertStmt` - Assertion
- `DelStmt` - Delete statement
- `With` - Context manager

## Expressions

### Literals

```go
type Literal struct {
    Type  LiteralType
    Value interface{}
    Span  lexer.Span
}

// LiteralType values:
// - LiteralTypeString
// - LiteralTypeNumber
// - LiteralTypeBool
// - LiteralTypeNone
// - LiteralTypeEllipsis
```

### Identifiers

```go
type Name struct {
    Token lexer.Token  // Token.Lexeme contains the identifier
    Span  lexer.Span
}
```

### Binary Operations

```go
type Binary struct {
    Left     Expr
    Operator lexer.Token
    Right    Expr
    Span     lexer.Span
}
```

Operators: `+`, `-`, `*`, `/`, `//`, `%`, `**`, `@`, `&`, `|`, `^`, `<<`, `>>`, `and`, `or`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `is`, `is not`, `in`, `not in`

### Unary Operations

```go
type Unary struct {
    Operator lexer.Token
    Right    Expr
    Span     lexer.Span
}
```

Operators: `not`, `-`, `+`, `~`

### Function Calls

```go
type Call struct {
    Callee    Expr
    Arguments []*Argument
    Span      lexer.Span
}

type Argument struct {
    Name  *Name  // For keyword arguments
    Value Expr
    IsStar bool  // For *args
    IsDoubleStar bool  // For **kwargs
    Span  lexer.Span
}
```

### Attribute Access

```go
type Attribute struct {
    Object Expr
    Name   lexer.Token  // Attribute name
    Span   lexer.Span
}
```

### Subscript/Indexing

```go
type Subscript struct {
    Object  Expr
    Indices []Expr  // Can include Slice nodes
    Span    lexer.Span
}

type Slice struct {
    StartIndex Expr
    EndIndex   Expr
    Step       Expr
    Span       lexer.Span
}
```

### Collections

#### List
```go
type ListExpr struct {
    Elements []Expr
    Span     lexer.Span
}
```

#### Tuple
```go
type TupleExpr struct {
    Elements []Expr
    Span     lexer.Span
}
```

#### Dict
```go
type DictExpr struct {
    Pairs []DictPair
    Span  lexer.Span
}

// DictPair can be either KeyValuePair or DoubleStarredPair
type KeyValuePair struct {
    Key   Expr
    Value Expr
    Span  lexer.Span
}

type DoubleStarredPair struct {
    Expr Expr  // For **dict unpacking
    Span lexer.Span
}
```

#### Set
```go
type SetExpr struct {
    Elements []Expr
    Span     lexer.Span
}
```

### F-Strings

```go
type FString struct {
    Parts []FStringPart
    Span  lexer.Span
}

// FStringPart types:
type FStringStart struct {
    Value string
    Span  lexer.Span
}

type FStringMiddle struct {
    Value string
    Span  lexer.Span
}

type FStringEnd struct {
    Value string
    Span  lexer.Span
}

type FStringReplacementField struct {
    Expression  Expr
    Debug       bool   // For {expr=} syntax
    Conversion  string // !r, !s, !a
    FormatSpec  *FString // Nested f-string for format
    Span        lexer.Span
}
```

### Comprehensions

```go
type ListComp struct {
    Element Expr
    Clauses []ForIfClause
    Span    lexer.Span
}

type DictComp struct {
    Key     Expr
    Value   Expr
    Clauses []ForIfClause
    Span    lexer.Span
}

type SetComp struct {
    Element Expr
    Clauses []ForIfClause
    Span    lexer.Span
}

type GenExpr struct {
    Element Expr
    Clauses []ForIfClause
    Span    lexer.Span
}

type ForIfClause struct {
    Target  Expr
    Iter    Expr
    Ifs     []Expr  // if conditions
    IsAsync bool
    Span    lexer.Span
}
```

### Other Expressions

- `Lambda` - Lambda expression
- `TernaryExpr` - Conditional expression (`a if b else c`)
- `YieldExpr` - Yield expression
- `AwaitExpr` - Await expression
- `StarExpr` - Starred expression (`*expr`)
- `NamedExpr` - Walrus operator (`:=`)

## Helper Functions

The AST package provides helper functions for constructing nodes:

```go
// Basic value constructors
ast.N("identifier")     // Name node
ast.S("string")         // String literal
ast.I(42)              // Integer literal
ast.B(true)            // Boolean literal
ast.Nil()              // None literal

// Complex constructors
ast.HView(name, params, body...)
ast.HElement(tag, content...)
ast.HAttr(name, value)
ast.HCall(callee, args...)
ast.HBinary(left, op, right)
// ... and many more
```

## Visitor Pattern

All nodes support the visitor pattern:

```go
type Visitor interface {
    // Expression visitors
    VisitName(n *Name) Visitor
    VisitLiteral(l *Literal) Visitor
    VisitBinary(b *Binary) Visitor
    // ... etc for all node types
    
    // Statement visitors  
    VisitAssignStmt(a *AssignStmt) Visitor
    VisitIf(i *If) Visitor
    // ... etc for all statement types
}
```

## Usage Example

```go
// Parse PSX code
module, err := parser.Parse(source)

// Create a visitor
visitor := &MyVisitor{}

// Traverse the AST
module.Accept(visitor)
```

## See Also

- [Architecture Guide](../architecture.md) - Compiler architecture
- [Parser Design](parser_design.md) - Parser implementation details
- [AST Helpers](../../compiler/ast/helpers.go) - Helper function source