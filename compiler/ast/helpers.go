package ast

import (
	"sylfie/compiler/lexer"
	"strings"
)

// Helper functions for constructing AST nodes in tests.
// All functions are prefixed with 'H' to avoid conflicts with struct names.

// N creates an identifier AST node
func N(id string) *Name {
	return &Name{Token: lexer.Token{Lexeme: id}}
}

// S creates a string literal AST node
func S(value string) *Literal {
	return &Literal{Value: value, Type: LiteralTypeString}
}

// I creates an integer literal AST node
func I(value int64) *Literal {
	return &Literal{Value: value, Type: LiteralTypeNumber}
}

// B creates a boolean literal AST node
func B(value bool) *Literal {
	return &Literal{Value: value, Type: LiteralTypeBool}
}

// Nil creates a None literal AST node
func Nil() *Literal {
	return &Literal{Value: nil, Type: LiteralTypeNone}
}

// Param creates a parameter with optional type annotation
func HParam(name string, annotation ...string) *Parameter {
	param := &Parameter{Name: N(name)}
	if len(annotation) > 0 {
		param.Annotation = N(annotation[0])
	}
	return param
}

// ParamWithDefault creates a parameter with type annotation and default value
func HParamWithDefault(name string, annotation string, defaultValue Expr) *Parameter {
	return &Parameter{
		Name:       N(name),
		Annotation: N(annotation),
		Default:    defaultValue,
	}
}

// View creates a ViewStmt AST node
func HView(name string, params []*Parameter, body ...Stmt) *ViewStmt {
	return &ViewStmt{
		Name:   N(name),
		Params: &ParameterList{Parameters: params},
		Body:   body,
	}
}

// Element creates an HTMLElement AST node with flexible content
func HElement(tag string, content ...interface{}) *HTMLElement {
	elem := &HTMLElement{
		TagName: lexer.Token{Lexeme: tag},
	}
	
	for _, item := range content {
		switch v := item.(type) {
		case HTMLAttribute:
			elem.Attributes = append(elem.Attributes, v)
		case []HTMLAttribute:
			elem.Attributes = append(elem.Attributes, v...)
		case Stmt:
			elem.Content = append(elem.Content, v)
		case []Stmt:
			elem.Content = append(elem.Content, v...)
		case string:
			// Convenience: string becomes literal expression statement
			elem.Content = append(elem.Content, &ExprStmt{Expr: S(v)})
		case Expr:
			// Convenience: expression becomes expression statement
			elem.Content = append(elem.Content, &ExprStmt{Expr: v})
		}
	}
	
	return elem
}

// Attr creates an HTML attribute
func HAttr(name string, value Expr) HTMLAttribute {
	return HTMLAttribute{
		Name:  lexer.Token{Lexeme: name},
		Value: value,
	}
}

// ExprStmt wraps an expression as a statement
func HExprStmt(expr Expr) *ExprStmt {
	return &ExprStmt{Expr: expr}
}

// FStr creates an f-string with mixed string and expression parts
func HFStr(parts ...interface{}) *FString {
	fstring := &FString{}
	for _, part := range parts {
		switch v := part.(type) {
		case string:
			fstring.Parts = append(fstring.Parts, &FStringMiddle{Value: v})
		case Expr:
			fstring.Parts = append(fstring.Parts, &FStringReplacementField{Expression: v})
		}
	}
	return fstring
}

// Binary creates a binary expression
func HBinary(left Expr, op lexer.TokenType, opStr string, right Expr) *Binary {
	return &Binary{
		Left:     left,
		Operator: lexer.Token{Type: op, Lexeme: opStr},
		Right:    right,
	}
}

// Call creates a function call expression
func HCall(callee Expr, args ...Expr) *Call {
	call := &Call{Callee: callee}
	for _, arg := range args {
		call.Arguments = append(call.Arguments, &Argument{Value: arg})
	}
	return call
}

// NamedCall creates a function call with named arguments
func NamedCall(callee Expr, args ...*Argument) *Call {
	return &Call{
		Callee:    callee,
		Arguments: args,
	}
}

// Arg creates a positional argument
func HArg(value Expr) *Argument {
	return &Argument{Value: value}
}

// NamedArg creates a named argument
func NamedArg(name string, value Expr) *Argument {
	return &Argument{
		Name:  N(name),
		Value: value,
	}
}

// List creates a list literal
func HList(elements ...Expr) *ListExpr {
	return &ListExpr{Elements: elements}
}

// Dict creates a dict literal
func HDict(pairs ...DictPair) *DictExpr {
	return &DictExpr{Pairs: pairs}
}

// KeyValue creates a key-value pair for a dict
func HKeyValue(key, value Expr) *KeyValuePair {
	return &KeyValuePair{Key: key, Value: value}
}

// DictUnpack creates a double-starred dict unpacking (**dict)
func HDictUnpack(value Expr) *DoubleStarredPair {
	return &DoubleStarredPair{Expr: value}
}

// Tuple creates a tuple literal
func HTuple(elements ...Expr) *TupleExpr {
	return &TupleExpr{Elements: elements}
}

// Set creates a set literal
func HSet(elements ...Expr) *SetExpr {
	return &SetExpr{Elements: elements}
}

// AttributeAccess creates an attribute access expression (e.g., obj.attr)
func HAttributeAccess(object Expr, attrName string) *Attribute {
	return &Attribute{
		Object: object,
		Name:   lexer.Token{Lexeme: attrName},
	}
}

// Subscript creates a subscript expression (e.g., obj[index])
func HSubscript(object Expr, indices ...Expr) *Subscript {
	return &Subscript{
		Object:  object,
		Indices: indices,
	}
}

// Slice creates a slice expression (e.g., obj[start:stop:step])
func HSlice(start, stop, step Expr) *Slice {
	return &Slice{
		StartIndex: start,
		EndIndex:   stop,
		Step:       step,
	}
}

// Assign creates an assignment statement
func HAssign(targets []Expr, value Expr) *AssignStmt {
	return &AssignStmt{
		Targets: targets,
		Value:   value,
	}
}

// If creates an if statement
func HIf(condition Expr, body []Stmt, elseBody ...Stmt) *If {
	return &If{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
	}
}

// For creates a for loop
func HFor(target Expr, iterable Expr, body []Stmt, elseBody ...Stmt) *For {
	return &For{
		Target:   target,
		Iterable: iterable,
		Body:     body,
		Else:     elseBody,
	}
}

// While creates a while loop
func HWhile(test Expr, body []Stmt, elseBody ...Stmt) *While {
	return &While{
		Test: test,
		Body: body,
		Else: elseBody,
	}
}

// Return creates a return statement
func HReturn(value Expr) *ReturnStmt {
	return &ReturnStmt{Value: value}
}

// Class creates a class definition
func HClass(name string, bases []Expr, body []Stmt) *Class {
	// Convert base expressions to Arguments
	var args []*Argument
	for _, base := range bases {
		args = append(args, &Argument{Value: base})
	}
	return &Class{
		Name: N(name),
		Args: args,
		Body: body,
	}
}

// Function creates a function definition
func HFunction(name string, params []*Parameter, body []Stmt, returnType Expr) *Function {
	return &Function{
		Name:       N(name),
		Parameters: &ParameterList{Parameters: params},
		Body:       body,
		ReturnType: returnType,
	}
}

// Lambda creates a lambda expression
func HLambda(params []*Parameter, body Expr) *Lambda {
	return &Lambda{
		Parameters: &ParameterList{Parameters: params},
		Body:       body,
	}
}

// Module creates a module (root AST node)
func HModule(body ...Stmt) *Module {
	return &Module{Body: body}
}

// Import creates an import statement
func HImport(names ...*ImportName) *ImportStmt {
	return &ImportStmt{Names: names}
}

// ImportName creates an import name with optional alias
func HImportN(name string, alias ...string) *ImportName {
	// Convert string to []*Name for DottedName
	parts := strings.Split(name, ".")
	var nameNodes []*Name
	for _, part := range parts {
		nameNodes = append(nameNodes, N(part))
	}
	
	item := &ImportName{DottedName: &DottedName{Names: nameNodes}}
	if len(alias) > 0 {
		item.AsName = N(alias[0])
	}
	return item
}

// ImportFrom creates an import from statement
func HImportFrom(module string, names []*ImportName, level ...int) *ImportFromStmt {
	// Convert module string to []*Name for DottedName
	parts := strings.Split(module, ".")
	var nameNodes []*Name
	for _, part := range parts {
		nameNodes = append(nameNodes, N(part))
	}
	
	stmt := &ImportFromStmt{
		DottedName: &DottedName{Names: nameNodes},
		Names:      names,
	}
	if len(level) > 0 {
		stmt.DotCount = level[0]
	}
	return stmt
}

// Try creates a try-except statement
func HTry(body []Stmt, handlers []Except, elseBody []Stmt, finallyBody []Stmt) *Try {
	return &Try{
		Body:    body,
		Excepts: handlers,
		Else:    elseBody,
		Finally: finallyBody,
	}
}

// Except creates an exception handler
func HExcept(exceptionType Expr, name string, body []Stmt) *Except {
	handler := &Except{Body: body}
	if exceptionType != nil {
		handler.Type = exceptionType
	}
	if name != "" {
		handler.Name = N(name)
	}
	return handler
}

// With creates a with statement
func HWith(items []WithItem, body []Stmt) *With {
	return &With{
		Items: items,
		Body:  body,
	}
}

// WithItem creates a with item
func HWithItem(contextExpr Expr, optionalVars Expr) *WithItem {
	return &WithItem{
		Expr: contextExpr,
		As:   optionalVars,
	}
}

// Match creates a match statement
func HMatch(subject Expr, cases []CaseBlock) *MatchStmt {
	return &MatchStmt{
		Subject: subject,
		Cases:   cases,
	}
}

// MatchCase creates a match case
func HMatchCase(patterns []Pattern, guard Expr, body []Stmt) *CaseBlock {
	return &CaseBlock{
		Patterns: patterns,
		Guard:    guard,
		Body:     body,
	}
}

// Global creates a global statement
func HGlobal(names ...string) *GlobalStmt {
	var nameNodes []*Name
	for _, name := range names {
		nameNodes = append(nameNodes, N(name))
	}
	return &GlobalStmt{Names: nameNodes}
}

// Nonlocal creates a nonlocal statement
func HNonlocal(names ...string) *NonlocalStmt {
	var nameNodes []*Name
	for _, name := range names {
		nameNodes = append(nameNodes, N(name))
	}
	return &NonlocalStmt{Names: nameNodes}
}

// Pass creates a pass statement
func HPass() *PassStmt {
	return &PassStmt{}
}

// Break creates a break statement
func HBreak() *BreakStmt {
	return &BreakStmt{}
}

// Continue creates a continue statement
func HContinue() *ContinueStmt {
	return &ContinueStmt{}
}

// Raise creates a raise statement
func HRaise(exc Expr, cause ...Expr) *RaiseStmt {
	stmt := &RaiseStmt{Exception: exc, HasException: exc != nil}
	if len(cause) > 0 {
		stmt.FromExpr = cause[0]
		stmt.HasFrom = true
	}
	return stmt
}

// Assert creates an assert statement
func HAssert(test Expr, msg ...Expr) *AssertStmt {
	stmt := &AssertStmt{Test: test}
	if len(msg) > 0 {
		stmt.Message = msg[0]
	}
	return stmt
}

// Yield creates a yield expression
func HYield(value ...Expr) *YieldExpr {
	y := &YieldExpr{}
	if len(value) > 0 {
		y.Value = value[0]
	}
	return y
}

// YieldFrom creates a yield from expression
func HYieldFrom(value Expr) *YieldExpr {
	return &YieldExpr{Value: value, IsFrom: true}
}

// Unary creates a unary expression
func HUnary(op lexer.TokenType, opStr string, operand Expr) *Unary {
	return &Unary{
		Operator: lexer.Token{Type: op, Lexeme: opStr},
		Right:    operand,
	}
}

// Ternary creates a ternary expression (a if condition else b)
func HTernary(condition Expr, trueExpr Expr, falseExpr Expr) *TernaryExpr {
	return &TernaryExpr{
		Condition: condition,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

// Star creates a starred expression
func HStar(value Expr) *StarExpr {
	return &StarExpr{Expr: value}
}

// ForIf creates a for-if clause for comprehensions
func HForIf(target Expr, iter Expr, ifs ...Expr) ForIfClause {
	return ForIfClause{
		Target: target,
		Iter:   iter,
		Ifs:    ifs,
	}
}

// ListComp creates a list comprehension
func HListComp(element Expr, generators ...ForIfClause) *ListComp {
	return &ListComp{
		Element: element,
		Clauses: generators,
	}
}

// DictComp creates a dict comprehension
func HDictComp(key Expr, value Expr, generators ...ForIfClause) *DictComp {
	return &DictComp{
		Key:     key,
		Value:   value,
		Clauses: generators,
	}
}

// SetComp creates a set comprehension
func HSetComp(element Expr, generators ...ForIfClause) *SetComp {
	return &SetComp{
		Element: element,
		Clauses: generators,
	}
}

// GeneratorExp creates a generator expression
func HGeneratorExp(element Expr, generators ...ForIfClause) *GenExpr {
	return &GenExpr{
		Element: element,
		Clauses: generators,
	}
}