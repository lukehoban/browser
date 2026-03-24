package js

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// control flow signals
type returnSignal struct{ val *Value }
type breakSignal struct{}
type continueSignal struct{}
type throwSignal struct{ val *Value }

// Interpreter evaluates a JavaScript AST.
type Interpreter struct {
	global *Env
}

// NewInterpreter creates a new interpreter with a global environment.
func NewInterpreter() *Interpreter {
	interp := &Interpreter{}
	interp.global = NewEnv()
	interp.setupBuiltins()
	return interp
}

// Eval evaluates a program node in the global environment.
func (interp *Interpreter) Eval(prog *Program) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if ts, ok := r.(throwSignal); ok {
				err = fmt.Errorf("uncaught exception: %s", ts.val.ToString())
			} else if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("runtime error: %v", r)
			}
		}
	}()
	// Hoist function declarations first
	interp.hoistFunctions(prog.Body, interp.global)
	interp.execBlock(prog.Body, interp.global)
	return
}

// hoistFunctions hoists function declarations to the top of the current scope.
func (interp *Interpreter) hoistFunctions(stmts []Node, env *Env) {
	for _, stmt := range stmts {
		if fd, ok := stmt.(*FunctionDeclaration); ok && fd.ID != nil {
			fn := interp.makeFunction(fd.ID.Name, fd.Params, fd.Body, env)
			env.Define(fd.ID.Name, fn)
		}
	}
}

// execBlock executes a list of statements.
func (interp *Interpreter) execBlock(stmts []Node, env *Env) {
	for _, stmt := range stmts {
		interp.execStmt(stmt, env)
	}
}

// execStmt executes a single statement.
func (interp *Interpreter) execStmt(node Node, env *Env) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ExpressionStatement:
		interp.evalExpr(n.Expr, env)

	case *BlockStatement:
		child := NewChildEnv(env)
		interp.hoistFunctions(n.Body, child)
		interp.execBlock(n.Body, child)

	case *StatementList:
		// Execute in the current scope (used for class declarations)
		interp.hoistFunctions(n.Body, env)
		interp.execBlock(n.Body, env)

	case *VariableDeclaration:
		for _, d := range n.Declarators {
			var val *Value
			if d.Init != nil {
				val = interp.evalExpr(d.Init, env)
			} else {
				val = Undefined
			}
			if n.Kind == "var" {
				// var: hoist to function scope (simplified: define in current env)
				env.SetOrDefine(d.ID.Name, val)
			} else {
				env.Define(d.ID.Name, val)
			}
		}

	case *FunctionDeclaration:
		// Already hoisted; skip if already defined, but update to allow redefinition
		if n.ID != nil {
			fn := interp.makeFunction(n.ID.Name, n.Params, n.Body, env)
			env.Define(n.ID.Name, fn)
		}

	case *IfStatement:
		cond := interp.evalExpr(n.Test, env)
		if cond.ToBoolean() {
			interp.execStmt(n.Consequent, env)
		} else if n.Alternate != nil {
			interp.execStmt(n.Alternate, env)
		}

	case *WhileStatement:
		for {
			cond := interp.evalExpr(n.Test, env)
			if !cond.ToBoolean() {
				break
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(breakSignal); ok {
							return
						}
						if _, ok := r.(continueSignal); ok {
							return
						}
						panic(r)
					}
				}()
				interp.execStmt(n.Body, env)
			}()
			// Re-check break signal
		}

	case *DoWhileStatement:
		for {
			broken := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(breakSignal); ok {
							broken = true
							return
						}
						if _, ok := r.(continueSignal); ok {
							return
						}
						panic(r)
					}
				}()
				interp.execStmt(n.Body, env)
			}()
			if broken {
				break
			}
			cond := interp.evalExpr(n.Test, env)
			if !cond.ToBoolean() {
				break
			}
		}

	case *ForStatement:
		child := NewChildEnv(env)
		if n.Init != nil {
			interp.execStmt(n.Init, child)
		}
		for {
			if n.Test != nil {
				cond := interp.evalExpr(n.Test, child)
				if !cond.ToBoolean() {
					break
				}
			}
			broken := false
			continued := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(breakSignal); ok {
							broken = true
							return
						}
						if _, ok := r.(continueSignal); ok {
							continued = true
							return
						}
						panic(r)
					}
				}()
				interp.execStmt(n.Body, child)
			}()
			if broken {
				break
			}
			_ = continued
			if n.Update != nil {
				interp.evalExpr(n.Update, child)
			}
		}

	case *ForInStatement:
		iterVal := interp.evalExpr(n.Right, env)
		child := NewChildEnv(env)

		// Determine variable name for the loop variable
		var loopVar string
		switch lv := n.Left.(type) {
		case *VariableDeclaration:
			if len(lv.Declarators) > 0 {
				loopVar = lv.Declarators[0].ID.Name
			}
		case *Identifier:
			loopVar = lv.Name
		}
		child.Define(loopVar, Undefined)

		if n.Of {
			// for...of: iterate over array elements or string chars
			interp.forOfLoop(loopVar, iterVal, n.Body, child)
		} else {
			// for...in: iterate over object keys
			interp.forInLoop(loopVar, iterVal, n.Body, child)
		}

	case *ReturnStatement:
		var val *Value
		if n.Argument != nil {
			val = interp.evalExpr(n.Argument, env)
		} else {
			val = Undefined
		}
		panic(returnSignal{val: val})

	case *BreakStatement:
		panic(breakSignal{})

	case *ContinueStatement:
		panic(continueSignal{})

	case *ThrowStatement:
		val := interp.evalExpr(n.Argument, env)
		panic(throwSignal{val: val})

	case *TryCatchStatement:
		interp.execTryCatch(n, env)

	case *SwitchStatement:
		interp.execSwitch(n, env)
	}
}

func (interp *Interpreter) forOfLoop(loopVar string, iterVal *Value, body Node, env *Env) {
	if iterVal.typ == TypeString {
		for _, ch := range iterVal.strVal {
			env.Set(loopVar, StringVal(string(ch)))
			broken := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(breakSignal); ok {
							broken = true
						} else if _, ok := r.(continueSignal); ok {
							// continue
						} else {
							panic(r)
						}
					}
				}()
				interp.execStmt(body, env)
			}()
			if broken {
				break
			}
		}
		return
	}
	if iterVal.typ == TypeObject {
		obj := iterVal.objVal
		if obj.isArray {
			elems := obj.ArrayElements()
			for _, elem := range elems {
				env.Set(loopVar, elem)
				broken := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							if _, ok := r.(breakSignal); ok {
								broken = true
							} else if _, ok := r.(continueSignal); ok {
								// continue
							} else {
								panic(r)
							}
						}
					}()
					interp.execStmt(body, env)
				}()
				if broken {
					break
				}
			}
			return
		}
	}
}

func (interp *Interpreter) forInLoop(loopVar string, iterVal *Value, body Node, env *Env) {
	if iterVal.typ == TypeObject {
		obj := iterVal.objVal
		keys := obj.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			env.Set(loopVar, StringVal(key))
			broken := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(breakSignal); ok {
							broken = true
						} else if _, ok := r.(continueSignal); ok {
						} else {
							panic(r)
						}
					}
				}()
				interp.execStmt(body, env)
			}()
			if broken {
				break
			}
		}
	}
}

func (interp *Interpreter) execTryCatch(n *TryCatchStatement, env *Env) {
	var caughtErr interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(returnSignal); ok {
					panic(r) // re-throw return
				}
				if _, ok := r.(breakSignal); ok {
					panic(r)
				}
				if _, ok := r.(continueSignal); ok {
					panic(r)
				}
				caughtErr = r
			}
		}()
		child := NewChildEnv(env)
		interp.hoistFunctions(n.Block.Body, child)
		interp.execBlock(n.Block.Body, child)
	}()

	if caughtErr != nil && n.Handler != nil {
		child := NewChildEnv(env)
		if n.Param != nil {
			var errVal *Value
			if ts, ok := caughtErr.(throwSignal); ok {
				errVal = ts.val
			} else if e, ok := caughtErr.(error); ok {
				errVal = StringVal(e.Error())
			} else {
				errVal = StringVal(fmt.Sprintf("%v", caughtErr))
			}
			child.Define(n.Param.Name, errVal)
		}
		interp.hoistFunctions(n.Handler.Body, child)
		interp.execBlock(n.Handler.Body, child)
	}

	if n.Finally != nil {
		child := NewChildEnv(env)
		interp.hoistFunctions(n.Finally.Body, child)
		interp.execBlock(n.Finally.Body, child)
	}
}

func (interp *Interpreter) execSwitch(n *SwitchStatement, env *Env) {
	disc := interp.evalExpr(n.Discriminant, env)
	child := NewChildEnv(env)

	// Find matching case
	matched := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(breakSignal); ok {
					return
				}
				panic(r)
			}
		}()
		for _, c := range n.Cases {
			if !matched {
				if c.Test == nil {
					// default - skip for now, handle below
					continue
				}
				testVal := interp.evalExpr(c.Test, child)
				if strictEquals(disc, testVal) {
					matched = true
				}
			}
			if matched {
				for _, stmt := range c.Consequent {
					interp.execStmt(stmt, child)
				}
			}
		}
		// If no match, try default
		if !matched {
			for _, c := range n.Cases {
				if c.Test == nil {
					for _, stmt := range c.Consequent {
						interp.execStmt(stmt, child)
					}
					break
				}
			}
		}
	}()
}

// ---- Expression evaluation ----

func (interp *Interpreter) evalExpr(node Node, env *Env) *Value {
	if node == nil {
		return Undefined
	}
	switch n := node.(type) {
	case *Literal:
		return interp.evalLiteral(n)

	case *Identifier:
		return interp.evalIdentifier(n, env)

	case *BinaryExpression:
		return interp.evalBinary(n, env)

	case *UnaryExpression:
		return interp.evalUnary(n, env)

	case *UpdateExpression:
		return interp.evalUpdate(n, env)

	case *AssignmentExpression:
		return interp.evalAssignment(n, env)

	case *LogicalExpression:
		return interp.evalLogical(n, env)

	case *ConditionalExpression:
		cond := interp.evalExpr(n.Test, env)
		if cond.ToBoolean() {
			return interp.evalExpr(n.Consequent, env)
		}
		return interp.evalExpr(n.Alternate, env)

	case *CallExpression:
		return interp.evalCall(n, env)

	case *MemberExpression:
		obj, _ := interp.evalMember(n, env)
		return obj

	case *ObjectExpression:
		return interp.evalObjectLiteral(n, env)

	case *ArrayExpression:
		return interp.evalArrayLiteral(n, env)

	case *FunctionExpression:
		return interp.evalFunctionExpr(n, env)

	case *ArrowFunctionExpression:
		return interp.evalArrowFunc(n, env)

	case *NewExpression:
		return interp.evalNew(n, env)

	case *ThisExpression:
		if v, ok := env.Get("this"); ok {
			return v
		}
		return Undefined

	case *SequenceExpression:
		var result *Value = Undefined
		for _, expr := range n.Expressions {
			result = interp.evalExpr(expr, env)
		}
		return result

	case *TemplateLiteral:
		return interp.evalTemplate(n, env)

	case *SpreadElement:
		return interp.evalExpr(n.Argument, env)
	}
	return Undefined
}

func (interp *Interpreter) evalLiteral(n *Literal) *Value {
	switch v := n.Value.(type) {
	case float64:
		return NumberVal(v)
	case string:
		return StringVal(v)
	case bool:
		return BoolVal(v)
	case nil:
		if n.Raw == "undefined" {
			return Undefined
		}
		return Null
	case struct{}:
		return Undefined
	}
	return Undefined
}

func (interp *Interpreter) evalIdentifier(n *Identifier, env *Env) *Value {
	if v, ok := env.Get(n.Name); ok {
		return v
	}
	return Undefined
}

// evalMember evaluates a MemberExpression and returns (value, receiver, propertyKey).
func (interp *Interpreter) evalMember(n *MemberExpression, env *Env) (*Value, *Value) {
	obj := interp.evalExpr(n.Object, env)

	var key string
	if n.Computed {
		keyVal := interp.evalExpr(n.Property, env)
		key = keyVal.ToString()
	} else {
		if id, ok := n.Property.(*Identifier); ok {
			key = id.Name
		} else {
			key = interp.evalExpr(n.Property, env).ToString()
		}
	}

	return interp.getProperty(obj, key), obj
}

func (interp *Interpreter) getProperty(obj *Value, key string) *Value {
	// Handle primitive string properties
	if obj.typ == TypeString {
		switch key {
		case "length":
			return NumberVal(float64(len([]rune(obj.strVal))))
		default:
			// String indexing
			runes := []rune(obj.strVal)
			var idx int
			if _, err := fmt.Sscanf(key, "%d", &idx); err == nil && idx >= 0 && idx < len(runes) {
				return StringVal(string(runes[idx]))
			}
			// String prototype methods
			if v := interp.stringMethod(obj.strVal, key); v != nil {
				return v
			}
		}
		return Undefined
	}

	if obj.typ == TypeNumber {
		if v := interp.numberMethod(obj.numVal, key); v != nil {
			return v
		}
		return Undefined
	}

	if obj.typ == TypeObject || obj.typ == TypeFunction {
		o := obj.objVal
		if o == nil {
			return Undefined
		}
		// Array-specific methods
		if o.isArray {
			if v := interp.arrayMethod(o, key); v != nil {
				return v
			}
		}
		return o.Get(key)
	}

	return Undefined
}

func (interp *Interpreter) setProperty(obj *Value, key string, val *Value) {
	if obj.typ == TypeObject || obj.typ == TypeFunction {
		if obj.objVal != nil {
			obj.objVal.Set(key, val)
		}
	}
}

func (interp *Interpreter) evalBinary(n *BinaryExpression, env *Env) *Value {
	left := interp.evalExpr(n.Left, env)
	right := interp.evalExpr(n.Right, env)

	switch n.Op {
	case "+":
		// If either is a string, concatenate
		if left.typ == TypeString || right.typ == TypeString {
			return StringVal(left.ToString() + right.ToString())
		}
		if left.typ == TypeObject {
			lp := left.objVal.toPrimitive("default")
			if lp.typ == TypeString || right.typ == TypeString {
				return StringVal(lp.ToString() + right.ToString())
			}
			return NumberVal(lp.ToNumber() + right.ToNumber())
		}
		return NumberVal(left.ToNumber() + right.ToNumber())
	case "-":
		return NumberVal(left.ToNumber() - right.ToNumber())
	case "*":
		return NumberVal(left.ToNumber() * right.ToNumber())
	case "/":
		return NumberVal(left.ToNumber() / right.ToNumber())
	case "%":
		r := right.ToNumber()
		if r == 0 {
			return NumberVal(math.NaN())
		}
		return NumberVal(math.Mod(left.ToNumber(), r))
	case "**":
		return NumberVal(math.Pow(left.ToNumber(), right.ToNumber()))
	case "==":
		return BoolVal(looseEquals(left, right))
	case "!=":
		return BoolVal(!looseEquals(left, right))
	case "===":
		return BoolVal(strictEquals(left, right))
	case "!==":
		return BoolVal(!strictEquals(left, right))
	case "<":
		return BoolVal(lessThan(left, right))
	case ">":
		return BoolVal(lessThan(right, left))
	case "<=":
		return BoolVal(!lessThan(right, left))
	case ">=":
		return BoolVal(!lessThan(left, right))
	case "&":
		return NumberVal(float64(int32(left.ToNumber()) & int32(right.ToNumber())))
	case "|":
		return NumberVal(float64(int32(left.ToNumber()) | int32(right.ToNumber())))
	case "^":
		return NumberVal(float64(int32(left.ToNumber()) ^ int32(right.ToNumber())))
	case "<<":
		return NumberVal(float64(int32(left.ToNumber()) << uint(int32(right.ToNumber()))))
	case ">>":
		return NumberVal(float64(int32(left.ToNumber()) >> uint(int32(right.ToNumber()))))
	case "instanceof":
		if right.typ == TypeFunction && left.typ == TypeObject {
			// Simplified instanceof: check proto chain
			proto := right.objVal.Get("prototype")
			if proto.typ == TypeObject {
				p := left.objVal.proto
				for p != nil {
					if p == proto.objVal {
						return BoolVal(true)
					}
					p = p.proto
				}
			}
		}
		return BoolVal(false)
	case "in":
		if right.typ == TypeObject {
			key := left.ToString()
			v := right.objVal.Get(key)
			return BoolVal(!v.IsUndefined())
		}
		return BoolVal(false)
	}
	return Undefined
}

func (interp *Interpreter) evalUnary(n *UnaryExpression, env *Env) *Value {
	switch n.Op {
	case "!":
		v := interp.evalExpr(n.Argument, env)
		return BoolVal(!v.ToBoolean())
	case "-":
		v := interp.evalExpr(n.Argument, env)
		return NumberVal(-v.ToNumber())
	case "+":
		v := interp.evalExpr(n.Argument, env)
		return NumberVal(v.ToNumber())
	case "~":
		v := interp.evalExpr(n.Argument, env)
		return NumberVal(float64(^int32(v.ToNumber())))
	case "typeof":
		v := interp.evalExpr(n.Argument, env)
		switch v.typ {
		case TypeUndefined:
			return StringVal("undefined")
		case TypeNull:
			return StringVal("object")
		case TypeBoolean:
			return StringVal("boolean")
		case TypeNumber:
			return StringVal("number")
		case TypeString:
			return StringVal("string")
		case TypeFunction:
			return StringVal("function")
		case TypeObject:
			return StringVal("object")
		}
	case "void":
		interp.evalExpr(n.Argument, env)
		return Undefined
	case "delete":
		if mem, ok := n.Argument.(*MemberExpression); ok {
			obj := interp.evalExpr(mem.Object, env)
			var key string
			if mem.Computed {
				key = interp.evalExpr(mem.Property, env).ToString()
			} else if id, ok := mem.Property.(*Identifier); ok {
				key = id.Name
			}
			if obj.typ == TypeObject {
				obj.objVal.Delete(key)
			}
		}
		return BoolVal(true)
	}
	return Undefined
}

func (interp *Interpreter) evalUpdate(n *UpdateExpression, env *Env) *Value {
	current := interp.evalExpr(n.Argument, env)
	num := current.ToNumber()
	var newNum float64
	if n.Op == "++" {
		newNum = num + 1
	} else {
		newNum = num - 1
	}
	interp.assignTo(n.Argument, NumberVal(newNum), env)
	if n.Prefix {
		return NumberVal(newNum)
	}
	return NumberVal(num)
}

func (interp *Interpreter) evalAssignment(n *AssignmentExpression, env *Env) *Value {
	right := interp.evalExpr(n.Right, env)

	if n.Op == "=" {
		interp.assignTo(n.Left, right, env)
		return right
	}

	// Compound assignment
	left := interp.evalExpr(n.Left, env)
	var result *Value
	switch n.Op {
	case "+=":
		if left.typ == TypeString || right.typ == TypeString {
			result = StringVal(left.ToString() + right.ToString())
		} else {
			result = NumberVal(left.ToNumber() + right.ToNumber())
		}
	case "-=":
		result = NumberVal(left.ToNumber() - right.ToNumber())
	case "*=":
		result = NumberVal(left.ToNumber() * right.ToNumber())
	case "/=":
		result = NumberVal(left.ToNumber() / right.ToNumber())
	case "%=":
		result = NumberVal(math.Mod(left.ToNumber(), right.ToNumber()))
	case "&=":
		result = NumberVal(float64(int32(left.ToNumber()) & int32(right.ToNumber())))
	case "|=":
		result = NumberVal(float64(int32(left.ToNumber()) | int32(right.ToNumber())))
	case "^=":
		result = NumberVal(float64(int32(left.ToNumber()) ^ int32(right.ToNumber())))
	default:
		result = right
	}
	interp.assignTo(n.Left, result, env)
	return result
}

// assignTo assigns a value to a left-hand side expression.
func (interp *Interpreter) assignTo(lhs Node, val *Value, env *Env) {
	switch n := lhs.(type) {
	case *Identifier:
		if !env.Set(n.Name, val) {
			env.Define(n.Name, val)
		}
	case *MemberExpression:
		obj := interp.evalExpr(n.Object, env)
		var key string
		if n.Computed {
			key = interp.evalExpr(n.Property, env).ToString()
		} else if id, ok := n.Property.(*Identifier); ok {
			key = id.Name
		}
		interp.setProperty(obj, key, val)
	}
}

func (interp *Interpreter) evalLogical(n *LogicalExpression, env *Env) *Value {
	left := interp.evalExpr(n.Left, env)
	switch n.Op {
	case "&&":
		if !left.ToBoolean() {
			return left
		}
		return interp.evalExpr(n.Right, env)
	case "||":
		if left.ToBoolean() {
			return left
		}
		return interp.evalExpr(n.Right, env)
	case "??":
		if left.IsNullish() {
			return interp.evalExpr(n.Right, env)
		}
		return left
	}
	return Undefined
}

func (interp *Interpreter) evalCall(n *CallExpression, env *Env) *Value {
	// Evaluate callee and determine `this`
	var callee *Value
	var thisVal *Value

	switch c := n.Callee.(type) {
	case *MemberExpression:
		thisVal = interp.evalExpr(c.Object, env)
		var key string
		if c.Computed {
			key = interp.evalExpr(c.Property, env).ToString()
		} else if id, ok := c.Property.(*Identifier); ok {
			key = id.Name
		}
		callee = interp.getProperty(thisVal, key)
	default:
		callee = interp.evalExpr(n.Callee, env)
		thisVal = Undefined
	}

	// Evaluate arguments
	args := interp.evalArgs(n.Arguments, env)

	return interp.callFunction(callee, thisVal, args)
}

func (interp *Interpreter) evalArgs(argNodes []Node, env *Env) []*Value {
	var args []*Value
	for _, argNode := range argNodes {
		if spread, ok := argNode.(*SpreadElement); ok {
			v := interp.evalExpr(spread.Argument, env)
			if v.typ == TypeObject && v.objVal.isArray {
				for _, elem := range v.objVal.ArrayElements() {
					args = append(args, elem)
				}
			} else {
				args = append(args, v)
			}
		} else {
			args = append(args, interp.evalExpr(argNode, env))
		}
	}
	return args
}

func (interp *Interpreter) callFunction(callee *Value, thisVal *Value, args []*Value) *Value {
	if callee == nil || callee.IsUndefined() {
		panic(throwSignal{val: StringVal("TypeError: callee is not a function")})
	}
	if callee.typ != TypeFunction {
		panic(throwSignal{val: StringVal(fmt.Sprintf("TypeError: %s is not a function", callee.ToString()))})
	}

	obj := callee.objVal
	if obj.goFunc != nil {
		return obj.goFunc(thisVal, args)
	}

	// JavaScript function
	return interp.callJSFunction(obj, thisVal, args)
}

func (interp *Interpreter) callJSFunction(obj *Object, thisVal *Value, args []*Value) *Value {
	var params []*Identifier
	var body *BlockStatement
	var closureEnv *Env

	switch fn := obj.funcNode.(type) {
	case *FunctionDeclaration:
		params = fn.Params
		body = fn.Body
		closureEnv = obj.funcEnv
	case *FunctionExpression:
		params = fn.Params
		body = fn.Body
		closureEnv = obj.funcEnv
	case *ArrowFunctionExpression:
		// Arrow functions capture `this` from outer scope
		if thisVal == nil || thisVal.IsUndefined() {
			if v, ok := obj.funcEnv.Get("this"); ok {
				thisVal = v
			}
		}
		// params are []Node (identifiers)
		for _, p := range fn.Params {
			if id, ok := p.(*Identifier); ok {
				params = append(params, id)
			}
		}
		if b, ok := fn.Body.(*BlockStatement); ok {
			body = b
		} else {
			// Expression body
			fnEnv := NewChildEnv(obj.funcEnv)
			fnEnv.Define("this", thisVal)
			fnEnv.Define("arguments", makeArguments(args))
			for i, param := range params {
				if i < len(args) {
					fnEnv.Define(param.Name, args[i])
				} else {
					fnEnv.Define(param.Name, Undefined)
				}
			}
			return interp.evalExpr(fn.Body, fnEnv)
		}
		closureEnv = obj.funcEnv
	default:
		return Undefined
	}

	if closureEnv == nil {
		closureEnv = interp.global
	}

	fnEnv := NewChildEnv(closureEnv)
	fnEnv.Define("this", thisVal)
	fnEnv.Define("arguments", makeArguments(args))

	// Bind parameters
	for i, param := range params {
		if i < len(args) {
			fnEnv.Define(param.Name, args[i])
		} else {
			fnEnv.Define(param.Name, Undefined)
		}
	}

	// Hoist function declarations
	interp.hoistFunctions(body.Body, fnEnv)

	// Execute body
	var result *Value = Undefined
	func() {
		defer func() {
			if r := recover(); r != nil {
				if rs, ok := r.(returnSignal); ok {
					result = rs.val
				} else {
					panic(r)
				}
			}
		}()
		interp.execBlock(body.Body, fnEnv)
	}()
	return result
}

func makeArguments(args []*Value) *Value {
	arr := NewArray()
	for i, a := range args {
		arr.Set(fmt.Sprintf("%d", i), a)
	}
	arr.props["length"] = NumberVal(float64(len(args)))
	return ObjectVal(arr)
}

func (interp *Interpreter) evalObjectLiteral(n *ObjectExpression, env *Env) *Value {
	obj := NewObject()
	for _, prop := range n.Properties {
		var key string
		switch k := prop.Key.(type) {
		case *Identifier:
			key = k.Name
		case *Literal:
			key = k.Value.(string)
		default:
			key = interp.evalExpr(prop.Key, env).ToString()
		}
		val := interp.evalExpr(prop.Value, env)
		obj.Set(key, val)
	}
	return ObjectVal(obj)
}

func (interp *Interpreter) evalArrayLiteral(n *ArrayExpression, env *Env) *Value {
	arr := NewArray()
	for i, elem := range n.Elements {
		if elem == nil {
			arr.Set(fmt.Sprintf("%d", i), Undefined)
		} else if spread, ok := elem.(*SpreadElement); ok {
			v := interp.evalExpr(spread.Argument, env)
			if v.typ == TypeObject && v.objVal.isArray {
				for _, el := range v.objVal.ArrayElements() {
					arr.Push(el)
				}
			}
		} else {
			arr.Set(fmt.Sprintf("%d", i), interp.evalExpr(elem, env))
		}
	}
	if len(n.Elements) > 0 {
		arr.props["length"] = NumberVal(float64(len(n.Elements)))
	}
	return ObjectVal(arr)
}

func (interp *Interpreter) evalFunctionExpr(n *FunctionExpression, env *Env) *Value {
	return interp.makeFunction("", n.Params, n.Body, env)
}

func (interp *Interpreter) evalArrowFunc(n *ArrowFunctionExpression, env *Env) *Value {
	obj := &Object{
		props:    make(map[string]*Value),
		funcNode: n,
		funcEnv:  env,
	}
	return &Value{typ: TypeFunction, objVal: obj}
}

func (interp *Interpreter) evalNew(n *NewExpression, env *Env) *Value {
	callee := interp.evalExpr(n.Callee, env)
	args := interp.evalArgs(n.Arguments, env)

	if callee.typ != TypeFunction {
		panic(throwSignal{val: StringVal("TypeError: callee is not a constructor")})
	}

	// Create new object
	obj := NewObject()
	// Set prototype from callee.prototype
	proto := callee.objVal.Get("prototype")
	if proto.typ == TypeObject {
		obj.proto = proto.objVal
	}
	thisVal := ObjectVal(obj)

	// Call the constructor
	result := interp.callFunction(callee, thisVal, args)

	// If constructor returned an object, use it; otherwise return `this`
	if result.typ == TypeObject || result.typ == TypeFunction {
		return result
	}
	return thisVal
}

func (interp *Interpreter) evalTemplate(n *TemplateLiteral, env *Env) *Value {
	var sb strings.Builder
	for i, quasi := range n.Quasis {
		sb.WriteString(quasi)
		if i < len(n.Expressions) {
			v := interp.evalExpr(n.Expressions[i], env)
			sb.WriteString(v.ToString())
		}
	}
	return StringVal(sb.String())
}

// makeFunction creates a function value from declaration/expression components.
func (interp *Interpreter) makeFunction(name string, params []*Identifier, body *BlockStatement, env *Env) *Value {
	fn := &FunctionExpression{
		Params: params,
		Body:   body,
	}
	if name != "" {
		fn.ID = &Identifier{Name: name}
	}
	obj := &Object{
		props:    make(map[string]*Value),
		funcNode: fn,
		funcEnv:  env,
	}
	// Set up prototype for constructor use
	proto := NewObject()
	obj.Set("prototype", ObjectVal(proto))
	obj.Set("name", StringVal(name))
	return &Value{typ: TypeFunction, objVal: obj}
}

// ---- Comparison helpers ----

func looseEquals(a, b *Value) bool {
	// Same type - use strict equality
	if a.typ == b.typ {
		return strictEquals(a, b)
	}
	// null == undefined
	if a.IsNullish() && b.IsNullish() {
		return true
	}
	// number and string: convert string to number
	if a.typ == TypeNumber && b.typ == TypeString {
		return a.numVal == b.ToNumber()
	}
	if a.typ == TypeString && b.typ == TypeNumber {
		return a.ToNumber() == b.numVal
	}
	// boolean converts to number
	if a.typ == TypeBoolean {
		return looseEquals(NumberVal(a.ToNumber()), b)
	}
	if b.typ == TypeBoolean {
		return looseEquals(a, NumberVal(b.ToNumber()))
	}
	return false
}

func strictEquals(a, b *Value) bool {
	if a.typ != b.typ {
		return false
	}
	switch a.typ {
	case TypeUndefined, TypeNull:
		return true
	case TypeBoolean:
		return a.boolVal == b.boolVal
	case TypeNumber:
		if math.IsNaN(a.numVal) || math.IsNaN(b.numVal) {
			return false
		}
		return a.numVal == b.numVal
	case TypeString:
		return a.strVal == b.strVal
	case TypeObject, TypeFunction:
		return a.objVal == b.objVal
	}
	return false
}

func lessThan(a, b *Value) bool {
	if a.typ == TypeString && b.typ == TypeString {
		return a.strVal < b.strVal
	}
	an := a.ToNumber()
	bn := b.ToNumber()
	if math.IsNaN(an) || math.IsNaN(bn) {
		return false
	}
	return an < bn
}

// ---- Built-in methods ----

func (interp *Interpreter) stringMethod(s, name string) *Value {
	goFunc := func(fn func(*Value, []*Value) *Value) *Value {
		obj := &Object{props: make(map[string]*Value), goFunc: fn}
		return &Value{typ: TypeFunction, objVal: obj}
	}
	self := StringVal(s)
	switch name {
	case "toString", "valueOf":
		return goFunc(func(_ *Value, _ []*Value) *Value { return self })
	case "length":
		return NumberVal(float64(len([]rune(s))))
	case "charAt":
		return goFunc(func(_ *Value, args []*Value) *Value {
			idx := 0
			if len(args) > 0 {
				idx = int(args[0].ToNumber())
			}
			runes := []rune(s)
			if idx < 0 || idx >= len(runes) {
				return StringVal("")
			}
			return StringVal(string(runes[idx]))
		})
	case "charCodeAt":
		return goFunc(func(_ *Value, args []*Value) *Value {
			idx := 0
			if len(args) > 0 {
				idx = int(args[0].ToNumber())
			}
			runes := []rune(s)
			if idx < 0 || idx >= len(runes) {
				return NumberVal(math.NaN())
			}
			return NumberVal(float64(runes[idx]))
		})
	case "indexOf":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return NumberVal(-1)
			}
			sub := args[0].ToString()
			idx := strings.Index(s, sub)
			if idx < 0 {
				return NumberVal(-1)
			}
			// Count runes up to byte index
			return NumberVal(float64(len([]rune(s[:idx]))))
		})
	case "lastIndexOf":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return NumberVal(-1)
			}
			sub := args[0].ToString()
			idx := strings.LastIndex(s, sub)
			if idx < 0 {
				return NumberVal(-1)
			}
			return NumberVal(float64(len([]rune(s[:idx]))))
		})
	case "includes":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return BoolVal(false)
			}
			return BoolVal(strings.Contains(s, args[0].ToString()))
		})
	case "startsWith":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return BoolVal(false)
			}
			return BoolVal(strings.HasPrefix(s, args[0].ToString()))
		})
	case "endsWith":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return BoolVal(false)
			}
			return BoolVal(strings.HasSuffix(s, args[0].ToString()))
		})
	case "slice", "substring":
		return goFunc(func(_ *Value, args []*Value) *Value {
			runes := []rune(s)
			n := len(runes)
			start := 0
			end := n
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if name == "slice" && start < 0 {
					start = n + start
				}
				if start < 0 {
					start = 0
				}
				if start > n {
					start = n
				}
			}
			if len(args) > 1 && !args[1].IsUndefined() {
				end = int(args[1].ToNumber())
				if name == "slice" && end < 0 {
					end = n + end
				}
				if end < 0 {
					end = 0
				}
				if end > n {
					end = n
				}
			}
			if start > end {
				if name == "substring" {
					start, end = end, start
				} else {
					return StringVal("")
				}
			}
			return StringVal(string(runes[start:end]))
		})
	case "substr":
		return goFunc(func(_ *Value, args []*Value) *Value {
			runes := []rune(s)
			n := len(runes)
			start := 0
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if start < 0 {
					start = n + start
				}
				if start < 0 {
					start = 0
				}
			}
			length := n - start
			if len(args) > 1 && !args[1].IsUndefined() {
				length = int(args[1].ToNumber())
			}
			if length <= 0 || start >= n {
				return StringVal("")
			}
			end := start + length
			if end > n {
				end = n
			}
			return StringVal(string(runes[start:end]))
		})
	case "toUpperCase", "toLocaleUpperCase":
		return goFunc(func(_ *Value, _ []*Value) *Value { return StringVal(strings.ToUpper(s)) })
	case "toLowerCase", "toLocaleLowerCase":
		return goFunc(func(_ *Value, _ []*Value) *Value { return StringVal(strings.ToLower(s)) })
	case "trim":
		return goFunc(func(_ *Value, _ []*Value) *Value { return StringVal(strings.TrimSpace(s)) })
	case "trimStart", "trimLeft":
		return goFunc(func(_ *Value, _ []*Value) *Value {
			return StringVal(strings.TrimLeftFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }))
		})
	case "trimEnd", "trimRight":
		return goFunc(func(_ *Value, _ []*Value) *Value {
			return StringVal(strings.TrimRightFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }))
		})
	case "split":
		return goFunc(func(_ *Value, args []*Value) *Value {
			arr := NewArray()
			if len(args) == 0 || args[0].IsUndefined() {
				arr.Push(StringVal(s))
				return ObjectVal(arr)
			}
			sep := args[0].ToString()
			limit := -1
			if len(args) > 1 && !args[1].IsUndefined() {
				limit = int(args[1].ToNumber())
			}
			var parts []string
			if sep == "" {
				for _, ch := range []rune(s) {
					parts = append(parts, string(ch))
				}
			} else {
				parts = strings.Split(s, sep)
			}
			for i, part := range parts {
				if limit >= 0 && i >= limit {
					break
				}
				arr.Push(StringVal(part))
			}
			return ObjectVal(arr)
		})
	case "replace":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) < 2 {
				return StringVal(s)
			}
			search := args[0].ToString()
			replace := args[1].ToString()
			return StringVal(strings.Replace(s, search, replace, 1))
		})
	case "replaceAll":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) < 2 {
				return StringVal(s)
			}
			search := args[0].ToString()
			replace := args[1].ToString()
			return StringVal(strings.ReplaceAll(s, search, replace))
		})
	case "repeat":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return StringVal("")
			}
			n := int(args[0].ToNumber())
			if n <= 0 {
				return StringVal("")
			}
			return StringVal(strings.Repeat(s, n))
		})
	case "padStart":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return StringVal(s)
			}
			targetLen := int(args[0].ToNumber())
			pad := " "
			if len(args) > 1 && !args[1].IsUndefined() {
				pad = args[1].ToString()
			}
			runes := []rune(s)
			for len(runes) < targetLen {
				runes = append([]rune(pad), runes...)
			}
			return StringVal(string(runes[len(runes)-targetLen:]))
		})
	case "padEnd":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return StringVal(s)
			}
			targetLen := int(args[0].ToNumber())
			pad := " "
			if len(args) > 1 && !args[1].IsUndefined() {
				pad = args[1].ToString()
			}
			runes := []rune(s)
			for len(runes) < targetLen {
				runes = append(runes, []rune(pad)...)
			}
			return StringVal(string(runes[:targetLen]))
		})
	case "concat":
		return goFunc(func(_ *Value, args []*Value) *Value {
			result := s
			for _, a := range args {
				result += a.ToString()
			}
			return StringVal(result)
		})
	case "match":
		// Simplified: treat as includes check
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) == 0 {
				return Null
			}
			sub := args[0].ToString()
			if strings.Contains(s, sub) {
				arr := NewArray()
				arr.Push(StringVal(sub))
				return ObjectVal(arr)
			}
			return Null
		})
	case "at":
		return goFunc(func(_ *Value, args []*Value) *Value {
			runes := []rune(s)
			if len(args) == 0 {
				return Undefined
			}
			idx := int(args[0].ToNumber())
			if idx < 0 {
				idx = len(runes) + idx
			}
			if idx < 0 || idx >= len(runes) {
				return Undefined
			}
			return StringVal(string(runes[idx]))
		})
	}
	return nil
}

func (interp *Interpreter) numberMethod(n float64, name string) *Value {
	goFunc := func(fn func(*Value, []*Value) *Value) *Value {
		obj := &Object{props: make(map[string]*Value), goFunc: fn}
		return &Value{typ: TypeFunction, objVal: obj}
	}
	switch name {
	case "toString":
		return goFunc(func(_ *Value, args []*Value) *Value {
			if len(args) > 0 {
				base := int(args[0].ToNumber())
				if base == 16 {
					return StringVal(fmt.Sprintf("%x", int64(n)))
				}
				if base == 2 {
					return StringVal(fmt.Sprintf("%b", int64(n)))
				}
			}
			return StringVal(numberToString(n))
		})
	case "toFixed":
		return goFunc(func(_ *Value, args []*Value) *Value {
			digits := 0
			if len(args) > 0 {
				digits = int(args[0].ToNumber())
			}
			return StringVal(fmt.Sprintf("%.*f", digits, n))
		})
	case "toLocaleString":
		return goFunc(func(_ *Value, _ []*Value) *Value { return StringVal(numberToString(n)) })
	}
	return nil
}

func (interp *Interpreter) arrayMethod(arr *Object, name string) *Value {
	goFunc := func(fn func(*Value, []*Value) *Value) *Value {
		obj := &Object{props: make(map[string]*Value), goFunc: fn}
		return &Value{typ: TypeFunction, objVal: obj}
	}
	switch name {
	case "push":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return Undefined
			}
			for _, v := range args {
				a.Push(v)
			}
			return a.Get("length")
		})
	case "pop":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return Undefined
			}
			n := int(a.Get("length").ToNumber())
			if n == 0 {
				return Undefined
			}
			key := fmt.Sprintf("%d", n-1)
			v := a.Get(key)
			a.Delete(key)
			a.props["length"] = NumberVal(float64(n - 1))
			return v
		})
	case "shift":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return Undefined
			}
			n := int(a.Get("length").ToNumber())
			if n == 0 {
				return Undefined
			}
			first := a.Get("0")
			for i := 1; i < n; i++ {
				a.Set(fmt.Sprintf("%d", i-1), a.Get(fmt.Sprintf("%d", i)))
			}
			a.Delete(fmt.Sprintf("%d", n-1))
			a.props["length"] = NumberVal(float64(n - 1))
			return first
		})
	case "unshift":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return Undefined
			}
			n := int(a.Get("length").ToNumber())
			// Shift existing elements right
			for i := n - 1; i >= 0; i-- {
				a.Set(fmt.Sprintf("%d", i+len(args)), a.Get(fmt.Sprintf("%d", i)))
			}
			for i, v := range args {
				a.Set(fmt.Sprintf("%d", i), v)
			}
			a.props["length"] = NumberVal(float64(n + len(args)))
			return a.Get("length")
		})
	case "splice":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			n := int(a.Get("length").ToNumber())
			start := 0
			deleteCount := n
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if start < 0 {
					start = n + start
				}
				if start < 0 {
					start = 0
				}
				if start > n {
					start = n
				}
				deleteCount = n - start
			}
			if len(args) > 1 {
				deleteCount = int(args[1].ToNumber())
				if deleteCount < 0 {
					deleteCount = 0
				}
				if deleteCount > n-start {
					deleteCount = n - start
				}
			}
			insertItems := args[2:]
			// Remove items
			removed := NewArray()
			for i := 0; i < deleteCount; i++ {
				removed.Push(a.Get(fmt.Sprintf("%d", start+i)))
			}
			// Build new array
			newElems := make([]*Value, 0, n-deleteCount+len(insertItems))
			for i := 0; i < start; i++ {
				newElems = append(newElems, a.Get(fmt.Sprintf("%d", i)))
			}
			newElems = append(newElems, insertItems...)
			for i := start + deleteCount; i < n; i++ {
				newElems = append(newElems, a.Get(fmt.Sprintf("%d", i)))
			}
			// Clear and refill
			for i := 0; i < n; i++ {
				a.Delete(fmt.Sprintf("%d", i))
			}
			for i, v := range newElems {
				a.Set(fmt.Sprintf("%d", i), v)
			}
			a.props["length"] = NumberVal(float64(len(newElems)))
			return ObjectVal(removed)
		})
	case "slice":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			n := int(a.Get("length").ToNumber())
			start := 0
			end := n
			if len(args) > 0 {
				start = int(args[0].ToNumber())
				if start < 0 {
					start = n + start
				}
			}
			if len(args) > 1 && !args[1].IsUndefined() {
				end = int(args[1].ToNumber())
				if end < 0 {
					end = n + end
				}
			}
			if start < 0 {
				start = 0
			}
			if end > n {
				end = n
			}
			result := NewArray()
			for i := start; i < end; i++ {
				result.Push(a.Get(fmt.Sprintf("%d", i)))
			}
			return ObjectVal(result)
		})
	case "concat":
		return goFunc(func(this *Value, args []*Value) *Value {
			result := NewArray()
			a := this.ToObject()
			if a != nil {
				for _, v := range a.ArrayElements() {
					result.Push(v)
				}
			}
			for _, arg := range args {
				if arg.typ == TypeObject && arg.objVal.isArray {
					for _, v := range arg.objVal.ArrayElements() {
						result.Push(v)
					}
				} else {
					result.Push(arg)
				}
			}
			return ObjectVal(result)
		})
	case "join":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return StringVal("")
			}
			sep := ","
			if len(args) > 0 && !args[0].IsUndefined() {
				sep = args[0].ToString()
			}
			elems := a.ArrayElements()
			parts := make([]string, len(elems))
			for i, e := range elems {
				if !e.IsNullish() {
					parts[i] = e.ToString()
				}
			}
			return StringVal(strings.Join(parts, sep))
		})
	case "reverse":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return this
			}
			elems := a.ArrayElements()
			n := len(elems)
			for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
				elems[i], elems[j] = elems[j], elems[i]
			}
			for i, e := range elems {
				a.Set(fmt.Sprintf("%d", i), e)
			}
			return this
		})
	case "sort":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return this
			}
			elems := a.ArrayElements()
			var cmpFn *Value
			if len(args) > 0 && args[0].typ == TypeFunction {
				cmpFn = args[0]
			}
			sort.SliceStable(elems, func(i, j int) bool {
				if cmpFn != nil {
					r := interp.callFunction(cmpFn, Undefined, []*Value{elems[i], elems[j]})
					return r.ToNumber() < 0
				}
				return elems[i].ToString() < elems[j].ToString()
			})
			for i, e := range elems {
				a.Set(fmt.Sprintf("%d", i), e)
			}
			return this
		})
	case "indexOf":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return NumberVal(-1)
			}
			search := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				if strictEquals(e, search) {
					return NumberVal(float64(i))
				}
			}
			return NumberVal(-1)
		})
	case "includes":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return BoolVal(false)
			}
			search := args[0]
			elems := a.ArrayElements()
			for _, e := range elems {
				if strictEquals(e, search) {
					return BoolVal(true)
				}
			}
			return BoolVal(false)
		})
	case "find":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return Undefined
			}
			fn := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if r.ToBoolean() {
					return e
				}
			}
			return Undefined
		})
	case "findIndex":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return NumberVal(-1)
			}
			fn := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if r.ToBoolean() {
					return NumberVal(float64(i))
				}
			}
			return NumberVal(-1)
		})
	case "filter":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return ObjectVal(NewArray())
			}
			fn := args[0]
			result := NewArray()
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if r.ToBoolean() {
					result.Push(e)
				}
			}
			return ObjectVal(result)
		})
	case "map":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return ObjectVal(NewArray())
			}
			fn := args[0]
			result := NewArray()
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				result.Push(r)
			}
			return ObjectVal(result)
		})
	case "forEach":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return Undefined
			}
			fn := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
			}
			return Undefined
		})
	case "reduce":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return Undefined
			}
			fn := args[0]
			elems := a.ArrayElements()
			if len(elems) == 0 {
				if len(args) > 1 {
					return args[1]
				}
				return Undefined
			}
			acc := args[1]
			start := 0
			if len(args) < 2 {
				acc = elems[0]
				start = 1
			}
			for i := start; i < len(elems); i++ {
				acc = interp.callFunction(fn, Undefined, []*Value{acc, elems[i], NumberVal(float64(i)), this})
			}
			return acc
		})
	case "some":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return BoolVal(false)
			}
			fn := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if r.ToBoolean() {
					return BoolVal(true)
				}
			}
			return BoolVal(false)
		})
	case "every":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return BoolVal(true)
			}
			fn := args[0]
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if !r.ToBoolean() {
					return BoolVal(false)
				}
			}
			return BoolVal(true)
		})
	case "flat":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			depth := 1
			if len(args) > 0 && !args[0].IsUndefined() {
				depth = int(args[0].ToNumber())
			}
			result := NewArray()
			flatInto(result, a, depth)
			return ObjectVal(result)
		})
	case "flatMap":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return ObjectVal(NewArray())
			}
			fn := args[0]
			result := NewArray()
			elems := a.ArrayElements()
			for i, e := range elems {
				r := interp.callFunction(fn, Undefined, []*Value{e, NumberVal(float64(i)), this})
				if r.typ == TypeObject && r.objVal.isArray {
					for _, elem := range r.objVal.ArrayElements() {
						result.Push(elem)
					}
				} else {
					result.Push(r)
				}
			}
			return ObjectVal(result)
		})
	case "toString":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return StringVal("")
			}
			elems := a.ArrayElements()
			parts := make([]string, len(elems))
			for i, e := range elems {
				if !e.IsNullish() {
					parts[i] = e.ToString()
				}
			}
			return StringVal(strings.Join(parts, ","))
		})
	case "at":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil || len(args) == 0 {
				return Undefined
			}
			n := int(a.Get("length").ToNumber())
			idx := int(args[0].ToNumber())
			if idx < 0 {
				idx = n + idx
			}
			if idx < 0 || idx >= n {
				return Undefined
			}
			return a.Get(fmt.Sprintf("%d", idx))
		})
	case "fill":
		return goFunc(func(this *Value, args []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return this
			}
			n := int(a.Get("length").ToNumber())
			val := Undefined
			if len(args) > 0 {
				val = args[0]
			}
			start := 0
			end := n
			if len(args) > 1 {
				start = int(args[1].ToNumber())
			}
			if len(args) > 2 {
				end = int(args[2].ToNumber())
			}
			for i := start; i < end && i < n; i++ {
				a.Set(fmt.Sprintf("%d", i), val)
			}
			return this
		})
	case "keys":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			result := NewArray()
			n := int(a.Get("length").ToNumber())
			for i := 0; i < n; i++ {
				result.Push(NumberVal(float64(i)))
			}
			return ObjectVal(result)
		})
	case "values":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			result := NewArray()
			for _, e := range a.ArrayElements() {
				result.Push(e)
			}
			return ObjectVal(result)
		})
	case "entries":
		return goFunc(func(this *Value, _ []*Value) *Value {
			a := this.ToObject()
			if a == nil {
				return ObjectVal(NewArray())
			}
			result := NewArray()
			for i, e := range a.ArrayElements() {
				pair := NewArray()
				pair.Push(NumberVal(float64(i)))
				pair.Push(e)
				result.Push(ObjectVal(pair))
			}
			return ObjectVal(result)
		})
	}
	return nil
}

func flatInto(target *Object, src *Object, depth int) {
	elems := src.ArrayElements()
	for _, e := range elems {
		if depth > 0 && e.typ == TypeObject && e.objVal.isArray {
			flatInto(target, e.objVal, depth-1)
		} else {
			target.Push(e)
		}
	}
}
