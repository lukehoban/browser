// Package js provides a simple JavaScript interpreter for executing scripts
// found in HTML documents. It supports a practical subset of ECMAScript 5/6:
// variables, functions, closures, control flow, and DOM manipulation.
package js

// Node is the base interface for all AST nodes.
type Node interface {
	nodeMarker()
}

// nodeBase is embedded in all AST nodes.
type nodeBase struct{}

func (nodeBase) nodeMarker() {}

// ---- Statements ----

// Program is the root of the AST.
type Program struct {
	nodeBase
	Body []Node
}

// ExpressionStatement wraps an expression used as a statement.
type ExpressionStatement struct {
	nodeBase
	Expr Node
}

// BlockStatement is a sequence of statements wrapped in braces.
type BlockStatement struct {
	nodeBase
	Body []Node
}

// VariableDeclarator is a single "id = init" within a var/let/const declaration.
type VariableDeclarator struct {
	ID   *Identifier
	Init Node // may be nil
}

// VariableDeclaration is var/let/const declaration with one or more declarators.
type VariableDeclaration struct {
	nodeBase
	Kind        string // "var", "let", or "const"
	Declarators []VariableDeclarator
}

// FunctionDeclaration is a named function declaration statement.
type FunctionDeclaration struct {
	nodeBase
	ID     *Identifier
	Params []*Identifier
	Body   *BlockStatement
}

// IfStatement is an if/else statement.
type IfStatement struct {
	nodeBase
	Test       Node
	Consequent Node
	Alternate  Node // may be nil
}

// WhileStatement is a while loop.
type WhileStatement struct {
	nodeBase
	Test Node
	Body Node
}

// DoWhileStatement is a do...while loop.
type DoWhileStatement struct {
	nodeBase
	Body Node
	Test Node
}

// ForStatement is a for loop (init; test; update).
type ForStatement struct {
	nodeBase
	Init   Node // VariableDeclaration, ExpressionStatement, or nil
	Test   Node // may be nil
	Update Node // may be nil
	Body   Node
}

// ForInStatement handles both for...in and for...of loops.
type ForInStatement struct {
	nodeBase
	Left  Node // VariableDeclaration or Identifier
	Right Node
	Body  Node
	Of    bool // true for for...of
}

// ReturnStatement returns a value from a function.
type ReturnStatement struct {
	nodeBase
	Argument Node // may be nil
}

// BreakStatement breaks out of a loop or switch.
type BreakStatement struct {
	nodeBase
}

// ContinueStatement continues to the next loop iteration.
type ContinueStatement struct {
	nodeBase
}

// ThrowStatement throws an exception.
type ThrowStatement struct {
	nodeBase
	Argument Node
}

// TryCatchStatement is a try/catch/finally block.
type TryCatchStatement struct {
	nodeBase
	Block   *BlockStatement
	Param   *Identifier     // catch parameter, may be nil
	Handler *BlockStatement // may be nil if no catch
	Finally *BlockStatement // may be nil
}

// SwitchStatement is a switch statement.
type SwitchStatement struct {
	nodeBase
	Discriminant Node
	Cases        []SwitchCase
}

// SwitchCase is a case clause within a switch statement.
type SwitchCase struct {
	Test       Node // nil for default
	Consequent []Node
}

// ---- Expressions ----

// Literal represents a literal value: number, string, bool, null, or undefined.
type Literal struct {
	nodeBase
	Value interface{} // float64, string, bool, nil (for null/undefined)
	Raw   string
}

// Identifier is a variable or property name.
type Identifier struct {
	nodeBase
	Name string
}

// BinaryExpression is a binary operation: +, -, *, /, %, ==, !=, <, >, etc.
type BinaryExpression struct {
	nodeBase
	Op    string
	Left  Node
	Right Node
}

// UnaryExpression is a unary operation: -, +, !, ~, typeof, void, delete.
type UnaryExpression struct {
	nodeBase
	Op       string
	Argument Node
	Prefix   bool
}

// UpdateExpression is ++ or -- (pre/post).
type UpdateExpression struct {
	nodeBase
	Op       string // "++" or "--"
	Argument Node
	Prefix   bool
}

// AssignmentExpression is an assignment: =, +=, -=, *=, /=, %=, etc.
type AssignmentExpression struct {
	nodeBase
	Op    string
	Left  Node
	Right Node
}

// LogicalExpression is && or || or ??.
type LogicalExpression struct {
	nodeBase
	Op    string
	Left  Node
	Right Node
}

// ConditionalExpression is the ternary operator: test ? consequent : alternate.
type ConditionalExpression struct {
	nodeBase
	Test       Node
	Consequent Node
	Alternate  Node
}

// CallExpression is a function call.
type CallExpression struct {
	nodeBase
	Callee    Node
	Arguments []Node
}

// MemberExpression is property access: obj.prop or obj[expr].
type MemberExpression struct {
	nodeBase
	Object   Node
	Property Node
	Computed bool // true for obj[expr], false for obj.prop
}

// Property is a key-value pair within an object literal.
type Property struct {
	Key      Node // Identifier or Literal
	Value    Node
	Shorthand bool
}

// ObjectExpression is an object literal: { key: value, ... }.
type ObjectExpression struct {
	nodeBase
	Properties []Property
}

// ArrayExpression is an array literal: [elem, elem, ...].
type ArrayExpression struct {
	nodeBase
	Elements []Node // may contain nil for holes/elision
}

// FunctionExpression is a function expression (possibly named).
type FunctionExpression struct {
	nodeBase
	ID     *Identifier // may be nil
	Params []*Identifier
	Body   *BlockStatement
}

// ArrowFunctionExpression is an arrow function: (params) => body.
type ArrowFunctionExpression struct {
	nodeBase
	Params []Node // *Identifier or destructuring (we support *Identifier)
	Body   Node   // BlockStatement or expression
}

// NewExpression is a constructor call: new Foo(args).
type NewExpression struct {
	nodeBase
	Callee    Node
	Arguments []Node
}

// ThisExpression is the `this` keyword.
type ThisExpression struct {
	nodeBase
}

// SequenceExpression is a comma-separated list of expressions.
type SequenceExpression struct {
	nodeBase
	Expressions []Node
}

// SpreadElement is ...expr inside a call/array.
type SpreadElement struct {
	nodeBase
	Argument Node
}

// TemplateLiteral is a template string with interpolated expressions.
type TemplateLiteral struct {
	nodeBase
	Quasis      []string // static string parts (len = len(Expressions) + 1)
	Expressions []Node
}

// StatementList is a sequence of statements that executes in the *current* scope
// (unlike BlockStatement, which always creates a new child scope).
// Used for class declarations where the constructor and methods must be visible
// in the enclosing scope.
type StatementList struct {
	nodeBase
	Body []Node
}
