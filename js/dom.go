package js

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/lukehoban/browser/dom"
	browserlog "github.com/lukehoban/browser/log"
)

// setupBuiltins registers all built-in globals in the interpreter's global environment.
func (interp *Interpreter) setupBuiltins() {
	g := interp.global

	// console
	console := NewObject()
	logFn := func(_ *Value, args []*Value) *Value {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.ToString()
		}
		browserlog.Infof("[JS] %s", strings.Join(parts, " "))
		return Undefined
	}
	console.Set("log", makeFn(logFn))
	console.Set("warn", makeFn(logFn))
	console.Set("error", makeFn(logFn))
	console.Set("info", makeFn(logFn))
	g.Define("console", ObjectVal(console))

	// Math
	g.Define("Math", ObjectVal(makeMathObject()))

	// JSON
	g.Define("JSON", ObjectVal(makeJSONObject()))

	// Array constructor
	arrayObj := NewObject()
	arrayObj.Set("isArray", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(args[0].typ == TypeObject && args[0].objVal != nil && args[0].objVal.isArray)
	}))
	arrayObj.Set("from", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		src := args[0]
		mapFn := Undefined
		if len(args) > 1 {
			mapFn = args[1]
		}
		result := NewArray()
		if src.typ == TypeString {
			for i, ch := range []rune(src.strVal) {
				v := StringVal(string(ch))
				if mapFn.typ == TypeFunction {
					v = interp.callFunction(mapFn, Undefined, []*Value{v, NumberVal(float64(i))})
				}
				result.Push(v)
			}
		} else if src.typ == TypeObject {
			if src.objVal.isArray {
				for i, e := range src.objVal.ArrayElements() {
					v := e
					if mapFn.typ == TypeFunction {
						v = interp.callFunction(mapFn, Undefined, []*Value{v, NumberVal(float64(i))})
					}
					result.Push(v)
				}
			} else {
				// Assume iterable with length
				n := int(src.objVal.Get("length").ToNumber())
				for i := 0; i < n; i++ {
					v := src.objVal.Get(fmt.Sprintf("%d", i))
					if mapFn.typ == TypeFunction {
						v = interp.callFunction(mapFn, Undefined, []*Value{v, NumberVal(float64(i))})
					}
					result.Push(v)
				}
			}
		}
		return ObjectVal(result)
	}))
	arrayObj.Set("of", makeFn(func(_ *Value, args []*Value) *Value {
		result := NewArray()
		for _, a := range args {
			result.Push(a)
		}
		return ObjectVal(result)
	}))
	// Make Array itself callable as a constructor
	arrayObj.goFunc = func(_ *Value, args []*Value) *Value {
		result := NewArray()
		if len(args) == 1 && args[0].typ == TypeNumber {
			n := int(args[0].ToNumber())
			result.props["length"] = NumberVal(float64(n))
		} else {
			for _, a := range args {
				result.Push(a)
			}
		}
		return ObjectVal(result)
	}
	g.Define("Array", &Value{typ: TypeFunction, objVal: arrayObj})

	// Object constructor
	objectCons := NewObject()
	objectCons.Set("keys", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		if args[0].typ != TypeObject {
			return ObjectVal(NewArray())
		}
		result := NewArray()
		for k := range args[0].objVal.props {
			result.Push(StringVal(k))
		}
		return ObjectVal(result)
	}))
	objectCons.Set("values", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		if args[0].typ != TypeObject {
			return ObjectVal(NewArray())
		}
		result := NewArray()
		for _, v := range args[0].objVal.props {
			result.Push(v)
		}
		return ObjectVal(result)
	}))
	objectCons.Set("entries", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		if args[0].typ != TypeObject {
			return ObjectVal(NewArray())
		}
		result := NewArray()
		for k, v := range args[0].objVal.props {
			pair := NewArray()
			pair.Push(StringVal(k))
			pair.Push(v)
			result.Push(ObjectVal(pair))
		}
		return ObjectVal(result)
	}))
	objectCons.Set("assign", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		target := args[0]
		if target.typ != TypeObject {
			return target
		}
		for _, src := range args[1:] {
			if src.typ == TypeObject {
				for k, v := range src.objVal.props {
					target.objVal.Set(k, v)
				}
			}
		}
		return target
	}))
	objectCons.Set("create", makeFn(func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		if len(args) > 0 && args[0].typ == TypeObject {
			obj.proto = args[0].objVal
		}
		return ObjectVal(obj)
	}))
	objectCons.Set("freeze", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 {
			return args[0]
		}
		return Undefined
	}))
	objectCons.Set("defineProperty", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 3 {
			return Undefined
		}
		obj := args[0]
		key := args[1].ToString()
		descriptor := args[2]
		if obj.typ == TypeObject && descriptor.typ == TypeObject {
			if v := descriptor.objVal.Get("value"); !v.IsUndefined() {
				obj.objVal.Set(key, v)
			}
		}
		return obj
	}))
	objectCons.goFunc = func(_ *Value, args []*Value) *Value {
		if len(args) == 0 || args[0].IsNullish() {
			return ObjectVal(NewObject())
		}
		return args[0]
	}
	g.Define("Object", &Value{typ: TypeFunction, objVal: objectCons})

	// String constructor
	strCons := NewObject()
	strCons.Set("fromCharCode", makeFn(func(_ *Value, args []*Value) *Value {
		var sb strings.Builder
		for _, a := range args {
			sb.WriteRune(rune(int(a.ToNumber())))
		}
		return StringVal(sb.String())
	}))
	strCons.goFunc = func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("")
		}
		return StringVal(args[0].ToString())
	}
	g.Define("String", &Value{typ: TypeFunction, objVal: strCons})

	// Number constructor
	numCons := NewObject()
	numCons.Set("isNaN", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		if args[0].typ != TypeNumber {
			return BoolVal(false)
		}
		return BoolVal(args[0].numVal != args[0].numVal) // NaN check
	}))
	numCons.Set("isFinite", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		if args[0].typ != TypeNumber {
			return BoolVal(false)
		}
		return BoolVal(!isInfOrNaN(args[0].numVal))
	}))
	numCons.Set("isInteger", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 || args[0].typ != TypeNumber {
			return BoolVal(false)
		}
		n := args[0].numVal
		return BoolVal(n == float64(int64(n)))
	}))
	numCons.Set("parseInt", makeFn(parseIntFn))
	numCons.Set("parseFloat", makeFn(parseFloatFn))
	numCons.Set("MAX_SAFE_INTEGER", NumberVal(9007199254740991))
	numCons.Set("MIN_SAFE_INTEGER", NumberVal(-9007199254740991))
	numCons.Set("MAX_VALUE", NumberVal(1.7976931348623157e+308))
	numCons.Set("POSITIVE_INFINITY", NumberVal(math.Inf(1)))
	numCons.Set("NEGATIVE_INFINITY", NumberVal(math.Inf(-1)))
	numCons.Set("NaN", NumberVal(nan()))
	numCons.goFunc = func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(0)
		}
		return NumberVal(args[0].ToNumber())
	}
	g.Define("Number", &Value{typ: TypeFunction, objVal: numCons})

	// Boolean constructor
	boolCons := NewObject()
	boolCons.goFunc = func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(args[0].ToBoolean())
	}
	g.Define("Boolean", &Value{typ: TypeFunction, objVal: boolCons})

	// Global functions
	g.Define("parseInt", makeFn(parseIntFn))
	g.Define("parseFloat", makeFn(parseFloatFn))
	g.Define("isNaN", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(true)
		}
		n := args[0].ToNumber()
		return BoolVal(n != n) // NaN check
	}))
	g.Define("isFinite", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(!isInfOrNaN(args[0].ToNumber()))
	}))
	g.Define("encodeURIComponent", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		s := args[0].ToString()
		return StringVal(percentEncode(s, false))
	}))
	g.Define("decodeURIComponent", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		return StringVal(args[0].ToString()) // simplified
	}))
	g.Define("encodeURI", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		s := args[0].ToString()
		return StringVal(percentEncode(s, true))
	}))
	g.Define("decodeURI", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		return StringVal(args[0].ToString())
	}))
	g.Define("escape", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		return StringVal(args[0].ToString())
	}))
	g.Define("unescape", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("undefined")
		}
		return StringVal(args[0].ToString())
	}))

	// setTimeout / setInterval stubs (execute immediately)
	g.Define("setTimeout", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 && args[0].typ == TypeFunction {
			interp.callFunction(args[0], Undefined, nil)
		}
		return NumberVal(0)
	}))
	g.Define("setInterval", makeFn(func(_ *Value, _ []*Value) *Value {
		return NumberVal(0)
	}))
	g.Define("clearTimeout", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	g.Define("clearInterval", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	g.Define("requestAnimationFrame", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 && args[0].typ == TypeFunction {
			interp.callFunction(args[0], Undefined, []*Value{NumberVal(0)})
		}
		return NumberVal(0)
	}))
	g.Define("cancelAnimationFrame", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))

	// alert/confirm/prompt stubs
	g.Define("alert", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 {
			browserlog.Infof("[JS alert] %s", args[0].ToString())
		}
		return Undefined
	}))
	g.Define("confirm", makeFn(func(_ *Value, _ []*Value) *Value { return BoolVal(false) }))
	g.Define("prompt", makeFn(func(_ *Value, _ []*Value) *Value { return Null }))

	// NaN, Infinity, undefined
	g.Define("NaN", NumberVal(nan()))
	g.Define("Infinity", NumberVal(math.Inf(1)))
	g.Define("undefined", Undefined)

	// Error constructor
	errorCons := NewObject()
	errorCons.goFunc = func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		msg := ""
		if len(args) > 0 {
			msg = args[0].ToString()
		}
		obj.Set("message", StringVal(msg))
		obj.Set("name", StringVal("Error"))
		return ObjectVal(obj)
	}
	g.Define("Error", &Value{typ: TypeFunction, objVal: errorCons})

	// Promise stub (just runs executor immediately, no async)
	promiseCons := NewObject()
	promiseCons.goFunc = func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		if len(args) > 0 && args[0].typ == TypeFunction {
			var resolvedVal *Value
			resolveFn := makeFn(func(_ *Value, a []*Value) *Value {
				if len(a) > 0 {
					resolvedVal = a[0]
				}
				return Undefined
			})
			rejectFn := makeFn(func(_ *Value, _ []*Value) *Value { return Undefined })
			interp.callFunction(args[0], Undefined, []*Value{resolveFn, rejectFn})
			obj.Set("__resolved__", func() *Value {
				if resolvedVal != nil {
					return resolvedVal
				}
				return Undefined
			}())
		}
		thenFn := makeFn(func(this *Value, a []*Value) *Value {
			if len(a) > 0 && a[0].typ == TypeFunction {
				if this.typ == TypeObject {
					resolvedVal := this.objVal.Get("__resolved__")
					interp.callFunction(a[0], Undefined, []*Value{resolvedVal})
				}
			}
			return this
		})
		catchFn := makeFn(func(this *Value, _ []*Value) *Value { return this })
		obj.Set("then", thenFn)
		obj.Set("catch", catchFn)
		return ObjectVal(obj)
	}
	promiseCons.Set("resolve", makeFn(func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		val := Undefined
		if len(args) > 0 {
			val = args[0]
		}
		obj.Set("__resolved__", val)
		obj.Set("then", makeFn(func(this *Value, a []*Value) *Value {
			if len(a) > 0 && a[0].typ == TypeFunction {
				interp.callFunction(a[0], Undefined, []*Value{val})
			}
			return this
		}))
		obj.Set("catch", makeFn(func(this *Value, _ []*Value) *Value { return this }))
		return ObjectVal(obj)
	}))
	promiseCons.Set("reject", makeFn(func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		obj.Set("then", makeFn(func(this *Value, _ []*Value) *Value { return this }))
		obj.Set("catch", makeFn(func(this *Value, a []*Value) *Value {
			if len(a) > 0 && a[0].typ == TypeFunction {
				val := Undefined
				if len(args) > 0 {
					val = args[0]
				}
				interp.callFunction(a[0], Undefined, []*Value{val})
			}
			return this
		}))
		return ObjectVal(obj)
	}))
	promiseCons.Set("all", makeFn(func(_ *Value, args []*Value) *Value {
		obj := NewObject()
		obj.Set("then", makeFn(func(this *Value, _ []*Value) *Value { return this }))
		obj.Set("catch", makeFn(func(this *Value, _ []*Value) *Value { return this }))
		return ObjectVal(obj)
	}))
	g.Define("Promise", &Value{typ: TypeFunction, objVal: promiseCons})

	// Map constructor
	mapCons := NewObject()
	mapCons.goFunc = func(_ *Value, _ []*Value) *Value {
		return ObjectVal(makeMapObject(interp))
	}
	g.Define("Map", &Value{typ: TypeFunction, objVal: mapCons})

	// Set constructor
	setCons := NewObject()
	setCons.goFunc = func(_ *Value, _ []*Value) *Value {
		return ObjectVal(makeSetObject(interp))
	}
	g.Define("Set", &Value{typ: TypeFunction, objVal: setCons})

	// window (alias for global)
	win := NewObject()
	win.Set("document", Undefined) // will be set when DOM is available
	win.Set("location", ObjectVal(makeLocationObject()))
	win.Set("navigator", ObjectVal(makeNavigatorObject()))
	win.Set("innerWidth", NumberVal(800))
	win.Set("innerHeight", NumberVal(600))
	win.goFunc = nil
	g.Define("window", ObjectVal(win))
	g.Define("globalThis", ObjectVal(win))
	g.Define("self", ObjectVal(win))
}

// SetupDOM binds the DOM tree to the interpreter's global environment.
func (interp *Interpreter) SetupDOM(doc *dom.Node) {
	docObj := interp.wrapDocument(doc)
	interp.global.Define("document", docObj)
	if win, ok := interp.global.Get("window"); ok && win.typ == TypeObject {
		win.objVal.Set("document", docObj)
	}
}

// ---- DOM wrapper ----

// wrapDocument wraps a DOM document node.
func (interp *Interpreter) wrapDocument(doc *dom.Node) *Value {
	obj := NewObject()

	obj.Set("getElementById", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		id := args[0].ToString()
		node := findByID(doc, id)
		if node == nil {
			return Null
		}
		return interp.wrapElement(node)
	}))

	obj.Set("querySelector", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		sel := args[0].ToString()
		node := querySelector(doc, sel)
		if node == nil {
			return Null
		}
		return interp.wrapElement(node)
	}))

	obj.Set("querySelectorAll", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		sel := args[0].ToString()
		nodes := querySelectorAll(doc, sel)
		arr := NewArray()
		for _, n := range nodes {
			arr.Push(interp.wrapElement(n))
		}
		return ObjectVal(arr)
	}))

	obj.Set("getElementsByTagName", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		tag := strings.ToLower(args[0].ToString())
		arr := NewArray()
		forEachElement(doc, func(n *dom.Node) {
			if n.Data == tag || tag == "*" {
				arr.Push(interp.wrapElement(n))
			}
		})
		return ObjectVal(arr)
	}))

	obj.Set("getElementsByClassName", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		cls := args[0].ToString()
		arr := NewArray()
		forEachElement(doc, func(n *dom.Node) {
			if hasClass(n, cls) {
				arr.Push(interp.wrapElement(n))
			}
		})
		return ObjectVal(arr)
	}))

	obj.Set("createElement", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		tag := strings.ToLower(args[0].ToString())
		node := dom.NewElement(tag)
		return interp.wrapElement(node)
	}))

	obj.Set("createTextNode", makeFn(func(_ *Value, args []*Value) *Value {
		text := ""
		if len(args) > 0 {
			text = args[0].ToString()
		}
		node := dom.NewText(text)
		return interp.wrapElement(node)
	}))

	obj.Set("createDocumentFragment", makeFn(func(_ *Value, _ []*Value) *Value {
		node := dom.NewElement("div") // simplified fragment
		return interp.wrapElement(node)
	}))

	// body / head accessors
	bodyNode := findByTag(doc, "body")
	if bodyNode != nil {
		obj.Set("body", interp.wrapElement(bodyNode))
	} else {
		obj.Set("body", Null)
	}
	headNode := findByTag(doc, "head")
	if headNode != nil {
		obj.Set("head", interp.wrapElement(headNode))
	} else {
		obj.Set("head", Null)
	}
	obj.Set("documentElement", interp.wrapElement(doc))

	// title
	titleNode := findByTag(doc, "title")
	titleVal := ""
	if titleNode != nil && len(titleNode.Children) > 0 && titleNode.Children[0].Type == dom.TextNode {
		titleVal = titleNode.Children[0].Data
	}
	obj.Set("title", StringVal(titleVal))

	// readyState
	obj.Set("readyState", StringVal("complete"))

	// write/writeln stubs
	obj.Set("write", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("writeln", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))

	// event stubs
	obj.Set("addEventListener", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("removeEventListener", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("dispatchEvent", makeFn(func(_ *Value, _ []*Value) *Value { return BoolVal(true) }))

	return ObjectVal(obj)
}

// elementCache caches wrapped element objects to preserve identity.
var elementCacheKey = "__elementCache__"

// wrapElement wraps a DOM element node.
func (interp *Interpreter) wrapElement(node *dom.Node) *Value {
	if node == nil {
		return Null
	}

	obj := NewObject()

	// tagName / nodeName
	tagName := strings.ToUpper(node.Data)
	obj.Set("tagName", StringVal(tagName))
	obj.Set("nodeName", StringVal(tagName))
	obj.Set("nodeType", NumberVal(float64(domNodeType(node))))

	// id
	obj.Set("id", StringVal(node.GetAttribute("id")))

	// className
	obj.Set("className", StringVal(node.GetAttribute("class")))

	// classList
	obj.Set("classList", interp.makeClassList(node))

	// style
	obj.Set("style", interp.makeStyleObject(node))

	// innerHTML
	// We use a getter/setter pattern via special "innerHTML" key
	obj.Set("innerHTML", StringVal(getInnerHTML(node)))

	// textContent
	obj.Set("textContent", StringVal(getTextContent(node)))

	// innerText (alias)
	obj.Set("innerText", StringVal(getTextContent(node)))

	// href, src, value for specific elements
	obj.Set("href", StringVal(node.GetAttribute("href")))
	obj.Set("src", StringVal(node.GetAttribute("src")))
	obj.Set("value", StringVal(node.GetAttribute("value")))
	obj.Set("type", StringVal(node.GetAttribute("type")))
	obj.Set("name", StringVal(node.GetAttribute("name")))
	obj.Set("checked", BoolVal(node.GetAttribute("checked") != ""))
	obj.Set("disabled", BoolVal(node.GetAttribute("disabled") != ""))
	obj.Set("hidden", BoolVal(node.GetAttribute("hidden") != ""))

	// children / childNodes
	obj.Set("children", interp.makeChildrenArray(node))
	obj.Set("childNodes", interp.makeChildrenArray(node))
	obj.Set("childElementCount", NumberVal(float64(countChildren(node))))
	obj.Set("firstChild", interp.firstChild(node))
	obj.Set("lastChild", interp.lastChild(node))
	obj.Set("firstElementChild", interp.firstElementChild(node))
	obj.Set("lastElementChild", interp.lastElementChild(node))
	obj.Set("nextSibling", Null)   // simplified
	obj.Set("previousSibling", Null)
	obj.Set("parentNode", Null)   // simplified
	obj.Set("parentElement", Null)

	// setAttribute / getAttribute / hasAttribute / removeAttribute
	obj.Set("setAttribute", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return Undefined
		}
		name := args[0].ToString()
		val := args[1].ToString()
		node.SetAttribute(name, val)
		// Sync back special props
		if name == "id" {
			obj.Set("id", StringVal(val))
		} else if name == "class" {
			obj.Set("className", StringVal(val))
			obj.Set("classList", interp.makeClassList(node))
		}
		return Undefined
	}))

	obj.Set("getAttribute", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		val := node.GetAttribute(args[0].ToString())
		if val == "" {
			return Null
		}
		return StringVal(val)
	}))

	obj.Set("hasAttribute", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		_, ok := node.Attributes[args[0].ToString()]
		return BoolVal(ok)
	}))

	obj.Set("removeAttribute", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 {
			delete(node.Attributes, args[0].ToString())
		}
		return Undefined
	}))

	// DOM mutation
	obj.Set("appendChild", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		child := args[0]
		if childNode := unwrapElement(child); childNode != nil {
			node.AppendChild(childNode)
			// Update children list
			obj.Set("children", interp.makeChildrenArray(node))
			obj.Set("childNodes", interp.makeChildrenArray(node))
		}
		return child
	}))

	obj.Set("insertBefore", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return Null
		}
		newChild := unwrapElement(args[0])
		refChild := unwrapElement(args[1])
		if newChild != nil {
			if refChild == nil {
				node.AppendChild(newChild)
			} else {
				// Insert before refChild
				insertBefore(node, newChild, refChild)
			}
			obj.Set("children", interp.makeChildrenArray(node))
		}
		return args[0]
	}))

	obj.Set("removeChild", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		childNode := unwrapElement(args[0])
		if childNode != nil {
			removeChild(node, childNode)
			obj.Set("children", interp.makeChildrenArray(node))
		}
		return args[0]
	}))

	obj.Set("replaceChild", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return Null
		}
		newChild := unwrapElement(args[0])
		oldChild := unwrapElement(args[1])
		if newChild != nil && oldChild != nil {
			for i, c := range node.Children {
				if c == oldChild {
					node.Children[i] = newChild
					newChild.Parent = node
					break
				}
			}
		}
		return args[1]
	}))

	obj.Set("cloneNode", makeFn(func(_ *Value, args []*Value) *Value {
		deep := false
		if len(args) > 0 {
			deep = args[0].ToBoolean()
		}
		cloned := cloneNode(node, deep)
		return interp.wrapElement(cloned)
	}))

	obj.Set("contains", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		other := unwrapElement(args[0])
		return BoolVal(nodeContains(node, other))
	}))

	// Setters for textContent, innerHTML, etc.
	// We intercept Set via a special setter function mechanism.
	// Since our object system is simple, we implement mutation via the JS-level
	// assignment path. We need to intercept property sets on DOM nodes.
	// Approach: use a special "setter" wrapper.
	obj.Set("__domNode__", interp.makeNodeRef(node))
	obj.Set("__setTextContent__", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		setTextContent(node, args[0].ToString())
		obj.Set("textContent", StringVal(args[0].ToString()))
		obj.Set("innerText", StringVal(args[0].ToString()))
		obj.Set("innerHTML", StringVal(getInnerHTML(node)))
		return Undefined
	}))
	obj.Set("__setInnerHTML__", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		setInnerHTML(node, args[0].ToString())
		obj.Set("innerHTML", StringVal(args[0].ToString()))
		obj.Set("textContent", StringVal(getTextContent(node)))
		obj.Set("children", interp.makeChildrenArray(node))
		return Undefined
	}))

	// querySelector / querySelectorAll on element
	obj.Set("querySelector", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		sel := args[0].ToString()
		found := querySelector(node, sel)
		if found == nil {
			return Null
		}
		return interp.wrapElement(found)
	}))
	obj.Set("querySelectorAll", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		sel := args[0].ToString()
		nodes := querySelectorAll(node, sel)
		arr := NewArray()
		for _, n := range nodes {
			arr.Push(interp.wrapElement(n))
		}
		return ObjectVal(arr)
	}))
	obj.Set("getElementsByTagName", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		tag := strings.ToLower(args[0].ToString())
		arr := NewArray()
		forEachElement(node, func(n *dom.Node) {
			if n.Data == tag || tag == "*" {
				arr.Push(interp.wrapElement(n))
			}
		})
		return ObjectVal(arr)
	}))
	obj.Set("getElementsByClassName", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(NewArray())
		}
		cls := args[0].ToString()
		arr := NewArray()
		forEachElement(node, func(n *dom.Node) {
			if hasClass(n, cls) {
				arr.Push(interp.wrapElement(n))
			}
		})
		return ObjectVal(arr)
	}))

	// focus / blur / click stubs
	obj.Set("focus", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("blur", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("click", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("scrollIntoView", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))

	// event listeners stubs
	obj.Set("addEventListener", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("removeEventListener", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("dispatchEvent", makeFn(func(_ *Value, _ []*Value) *Value { return BoolVal(true) }))

	// getBoundingClientRect
	obj.Set("getBoundingClientRect", makeFn(func(_ *Value, _ []*Value) *Value {
		rect := NewObject()
		rect.Set("top", NumberVal(0))
		rect.Set("left", NumberVal(0))
		rect.Set("right", NumberVal(0))
		rect.Set("bottom", NumberVal(0))
		rect.Set("width", NumberVal(0))
		rect.Set("height", NumberVal(0))
		return ObjectVal(rect)
	}))

	// offsetWidth / offsetHeight
	obj.Set("offsetWidth", NumberVal(0))
	obj.Set("offsetHeight", NumberVal(0))
	obj.Set("offsetTop", NumberVal(0))
	obj.Set("offsetLeft", NumberVal(0))
	obj.Set("scrollTop", NumberVal(0))
	obj.Set("scrollLeft", NumberVal(0))
	obj.Set("scrollWidth", NumberVal(0))
	obj.Set("scrollHeight", NumberVal(0))
	obj.Set("clientWidth", NumberVal(0))
	obj.Set("clientHeight", NumberVal(0))

	// matches
	obj.Set("matches", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(matchesSelector(node, args[0].ToString()))
	}))

	// closest
	obj.Set("closest", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Null
		}
		sel := args[0].ToString()
		current := node
		for current != nil {
			if current.Type == dom.ElementNode && matchesSelector(current, sel) {
				return interp.wrapElement(current)
			}
			current = current.Parent
		}
		return Null
	}))

	// insertAdjacentHTML
	obj.Set("insertAdjacentHTML", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return Undefined
		}
		position := strings.ToLower(args[0].ToString())
		html := args[1].ToString()
		insertAdjacentHTML(node, position, html)
		obj.Set("innerHTML", StringVal(getInnerHTML(node)))
		obj.Set("children", interp.makeChildrenArray(node))
		return Undefined
	}))

	// append / prepend
	obj.Set("append", makeFn(func(_ *Value, args []*Value) *Value {
		for _, arg := range args {
			if arg.typ == TypeString {
				node.AppendChild(dom.NewText(arg.strVal))
			} else if childNode := unwrapElement(arg); childNode != nil {
				node.AppendChild(childNode)
			}
		}
		obj.Set("children", interp.makeChildrenArray(node))
		return Undefined
	}))
	obj.Set("prepend", makeFn(func(_ *Value, args []*Value) *Value {
		var newChildren []*dom.Node
		for _, arg := range args {
			if arg.typ == TypeString {
				newChildren = append(newChildren, dom.NewText(arg.strVal))
			} else if childNode := unwrapElement(arg); childNode != nil {
				newChildren = append(newChildren, childNode)
			}
		}
		node.Children = append(newChildren, node.Children...)
		obj.Set("children", interp.makeChildrenArray(node))
		return Undefined
	}))
	obj.Set("remove", makeFn(func(_ *Value, _ []*Value) *Value {
		if node.Parent != nil {
			removeChild(node.Parent, node)
		}
		return Undefined
	}))
	obj.Set("replaceWith", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 && node.Parent != nil {
			if newNode := unwrapElement(args[0]); newNode != nil {
				for i, c := range node.Parent.Children {
					if c == node {
						node.Parent.Children[i] = newNode
						newNode.Parent = node.Parent
						break
					}
				}
			}
		}
		return Undefined
	}))

	// data-* attributes via dataset
	obj.Set("dataset", interp.makeDataset(node))

	// Install a propSetHook so that direct property assignments (e.g. elem.textContent = "x")
	// are reflected back to the underlying DOM node.
	obj.propSetHook = func(key string, val *Value) {
		interp.syncDOMProp(node, obj, key, val)
	}

	return ObjectVal(obj)
}

// syncDOMProp synchronises a JS property assignment back to the underlying DOM node.
// It is called by the propSetHook installed on every element wrapper object.
func (interp *Interpreter) syncDOMProp(node *dom.Node, obj *Object, key string, val *Value) {
	switch key {
	case "textContent", "innerText":
		text := val.ToString()
		setTextContent(node, text)
		// Keep both keys in sync (Set would re-trigger hook, so update props directly)
		obj.props["textContent"] = StringVal(text)
		obj.props["innerText"] = StringVal(text)
		obj.props["innerHTML"] = StringVal(getInnerHTML(node))
	case "innerHTML":
		html := val.ToString()
		setInnerHTML(node, html)
		obj.props["textContent"] = StringVal(getTextContent(node))
		obj.props["children"] = interp.makeChildrenArray(node)
	case "id":
		node.SetAttribute("id", val.ToString())
	case "className":
		node.SetAttribute("class", val.ToString())
		obj.props["classList"] = interp.makeClassList(node)
	case "hidden":
		if val.ToBoolean() {
			node.SetAttribute("hidden", "")
		} else {
			delete(node.Attributes, "hidden")
		}
	case "value":
		node.SetAttribute("value", val.ToString())
	case "href":
		node.SetAttribute("href", val.ToString())
	case "src":
		node.SetAttribute("src", val.ToString())
	case "disabled":
		if val.ToBoolean() {
			node.SetAttribute("disabled", "")
		} else {
			delete(node.Attributes, "disabled")
		}
	case "checked":
		if val.ToBoolean() {
			node.SetAttribute("checked", "")
		} else {
			delete(node.Attributes, "checked")
		}
	}
}

// makeNodeRef creates a sentinel value that holds a dom.Node reference.
func (interp *Interpreter) makeNodeRef(node *dom.Node) *Value {
	obj := NewObject()
	nodeID := interp.registerNode(node)
	obj.Set("__nodeID__", NumberVal(float64(nodeID)))
	return ObjectVal(obj)
}

var nodeRegistry = map[int]*dom.Node{}
var nodeRegistryCounter int

func (interp *Interpreter) registerNode(node *dom.Node) int {
	nodeRegistryCounter++
	nodeRegistry[nodeRegistryCounter] = node
	return nodeRegistryCounter
}

// unwrapElement retrieves the *dom.Node from a wrapped element value.
func unwrapElement(v *Value) *dom.Node {
	if v == nil || v.typ != TypeObject || v.objVal == nil {
		return nil
	}
	nodeRef := v.objVal.Get("__domNode__")
	if nodeRef.typ != TypeObject || nodeRef.objVal == nil {
		return nil
	}
	nodeIDVal := nodeRef.objVal.Get("__nodeID__")
	if nodeIDVal.typ != TypeNumber {
		return nil
	}
	nodeID := int(nodeIDVal.numVal)
	return nodeRegistry[nodeID]
}

// makeClassList creates a classList object for a DOM node.
func (interp *Interpreter) makeClassList(node *dom.Node) *Value {
	obj := NewObject()

	getClasses := func() []string {
		cls := node.GetAttribute("class")
		if cls == "" {
			return nil
		}
		return strings.Fields(cls)
	}
	setClasses := func(classes []string) {
		node.SetAttribute("class", strings.Join(classes, " "))
	}

	obj.Set("add", makeFn(func(_ *Value, args []*Value) *Value {
		classes := getClasses()
		for _, a := range args {
			name := a.ToString()
			found := false
			for _, c := range classes {
				if c == name {
					found = true
					break
				}
			}
			if !found {
				classes = append(classes, name)
			}
		}
		setClasses(classes)
		return Undefined
	}))

	obj.Set("remove", makeFn(func(_ *Value, args []*Value) *Value {
		classes := getClasses()
		toRemove := make(map[string]bool)
		for _, a := range args {
			toRemove[a.ToString()] = true
		}
		var newClasses []string
		for _, c := range classes {
			if !toRemove[c] {
				newClasses = append(newClasses, c)
			}
		}
		setClasses(newClasses)
		return Undefined
	}))

	obj.Set("contains", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		name := args[0].ToString()
		for _, c := range getClasses() {
			if c == name {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	}))

	obj.Set("toggle", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		name := args[0].ToString()
		classes := getClasses()
		found := false
		var newClasses []string
		for _, c := range classes {
			if c == name {
				found = true
			} else {
				newClasses = append(newClasses, c)
			}
		}
		if !found {
			newClasses = append(newClasses, name)
			setClasses(newClasses)
			return BoolVal(true)
		}
		setClasses(newClasses)
		return BoolVal(false)
	}))

	obj.Set("replace", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return BoolVal(false)
		}
		oldName := args[0].ToString()
		newName := args[1].ToString()
		classes := getClasses()
		found := false
		for i, c := range classes {
			if c == oldName {
				classes[i] = newName
				found = true
				break
			}
		}
		setClasses(classes)
		return BoolVal(found)
	}))

	// toString
	obj.Set("toString", makeFn(func(_ *Value, _ []*Value) *Value {
		return StringVal(node.GetAttribute("class"))
	}))

	return ObjectVal(obj)
}

// makeStyleObject creates a style object for a DOM node.
func (interp *Interpreter) makeStyleObject(node *dom.Node) *Value {
	obj := NewObject()

	// Parse existing inline style
	styleStr := node.GetAttribute("style")
	styles := parseInlineStyle(styleStr)
	for k, v := range styles {
		obj.Set(camelToCSS(k), StringVal(v))
		obj.Set(cssPropertyToCamel(k), StringVal(v))
	}

	// The style object needs to intercept property sets and update node's style attribute.
	// We implement this by wrapping the set operation.
	// When JS does elem.style.color = 'red', it calls obj.Set("color", ...)
	// We intercept via special setter properties.

	// cssText
	obj.Set("cssText", StringVal(styleStr))
	obj.Set("__styleNode__", interp.makeNodeRef(node))
	obj.Set("setProperty", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return Undefined
		}
		prop := strings.ToLower(args[0].ToString())
		val := args[1].ToString()
		updateNodeStyle(node, prop, val)
		obj.Set(cssPropertyToCamel(prop), StringVal(val))
		obj.Set(prop, StringVal(val))
		return Undefined
	}))
	obj.Set("getPropertyValue", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return StringVal("")
		}
		prop := strings.ToLower(args[0].ToString())
		styles := parseInlineStyle(node.GetAttribute("style"))
		return StringVal(styles[prop])
	}))
	obj.Set("removeProperty", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) > 0 {
			prop := strings.ToLower(args[0].ToString())
			updateNodeStyle(node, prop, "")
		}
		return StringVal("")
	}))

	// Install a hook so direct assignments like elem.style.color = 'red'
	// are reflected back to the DOM node's inline style.
	obj.propSetHook = func(key string, val *Value) {
		// Skip internal/method properties
		if strings.HasPrefix(key, "__") || key == "cssText" || key == "setProperty" ||
			key == "getPropertyValue" || key == "removeProperty" {
			if key == "cssText" {
				// Update entire style
				node.SetAttribute("style", val.ToString())
			}
			return
		}
		// Convert camelCase to kebab-case CSS property name
		cssProp := camelToCSS(key)
		cssVal := val.ToString()
		updateNodeStyle(node, cssProp, cssVal)
		// Keep the kebab-case key in sync too
		obj.props[cssProp] = StringVal(cssVal)
	}

	return ObjectVal(obj)
}

// makeChildrenArray wraps a node's children as a JS array.
func (interp *Interpreter) makeChildrenArray(node *dom.Node) *Value {
	arr := NewArray()
	for _, child := range node.Children {
		arr.Push(interp.wrapElement(child))
	}
	return ObjectVal(arr)
}

func (interp *Interpreter) firstChild(node *dom.Node) *Value {
	if len(node.Children) > 0 {
		return interp.wrapElement(node.Children[0])
	}
	return Null
}

func (interp *Interpreter) lastChild(node *dom.Node) *Value {
	if len(node.Children) > 0 {
		return interp.wrapElement(node.Children[len(node.Children)-1])
	}
	return Null
}

func (interp *Interpreter) firstElementChild(node *dom.Node) *Value {
	for _, c := range node.Children {
		if c.Type == dom.ElementNode {
			return interp.wrapElement(c)
		}
	}
	return Null
}

func (interp *Interpreter) lastElementChild(node *dom.Node) *Value {
	for i := len(node.Children) - 1; i >= 0; i-- {
		if node.Children[i].Type == dom.ElementNode {
			return interp.wrapElement(node.Children[i])
		}
	}
	return Null
}

// makeDataset creates a dataset object for data-* attributes.
func (interp *Interpreter) makeDataset(node *dom.Node) *Value {
	obj := NewObject()
	// Populate existing data-* attributes
	for k, v := range node.Attributes {
		if strings.HasPrefix(k, "data-") {
			camel := dataToCamel(k[5:])
			obj.Set(camel, StringVal(v))
		}
	}
	return ObjectVal(obj)
}

// ---- DOM helpers ----

func findByID(node *dom.Node, id string) *dom.Node {
	if node.Type == dom.ElementNode && node.GetAttribute("id") == id {
		return node
	}
	for _, child := range node.Children {
		if found := findByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func findByTag(node *dom.Node, tag string) *dom.Node {
	if node.Type == dom.ElementNode && node.Data == tag {
		return node
	}
	for _, child := range node.Children {
		if found := findByTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func forEachElement(node *dom.Node, fn func(*dom.Node)) {
	if node.Type == dom.ElementNode {
		fn(node)
	}
	for _, child := range node.Children {
		forEachElement(child, fn)
	}
}

// querySelector implements a basic CSS selector engine.
func querySelector(root *dom.Node, sel string) *dom.Node {
	nodes := querySelectorAll(root, sel)
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

func querySelectorAll(root *dom.Node, sel string) []*dom.Node {
	sel = strings.TrimSpace(sel)
	// Split by comma for multiple selectors
	parts := strings.Split(sel, ",")
	seen := map[*dom.Node]bool{}
	var result []*dom.Node
	for _, part := range parts {
		part = strings.TrimSpace(part)
		nodes := querySelectorAllSingle(root, part)
		for _, n := range nodes {
			if !seen[n] {
				seen[n] = true
				result = append(result, n)
			}
		}
	}
	return result
}

func querySelectorAllSingle(root *dom.Node, sel string) []*dom.Node {
	// Handle descendant combinator (space-separated parts)
	// This is a simplified implementation
	parts := splitSelectorParts(sel)
	if len(parts) == 1 {
		var result []*dom.Node
		forEachElement(root, func(n *dom.Node) {
			if matchesSelector(n, parts[0]) {
				result = append(result, n)
			}
		})
		return result
	}

	// Multiple parts: find all nodes matching the last part
	// that have an ancestor matching each preceding part
	var result []*dom.Node
	forEachElement(root, func(n *dom.Node) {
		if matchesComplex(n, parts) {
			result = append(result, n)
		}
	})
	return result
}

// splitSelectorParts splits a selector by combinators (space, >, +, ~).
func splitSelectorParts(sel string) []string {
	var parts []string
	var current strings.Builder
	for i := 0; i < len(sel); i++ {
		ch := sel[i]
		if ch == ' ' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else if ch == '>' || ch == '+' || ch == '~' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			// Skip combinator and following spaces
			for i+1 < len(sel) && sel[i+1] == ' ' {
				i++
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// matchesComplex checks if a node matches a complex selector (descendant chain).
func matchesComplex(node *dom.Node, parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	if !matchesSelector(node, parts[len(parts)-1]) {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	// Check ancestors
	ancestor := node.Parent
	for ancestor != nil {
		if matchesSelector(ancestor, parts[len(parts)-2]) {
			if len(parts) == 2 {
				return true
			}
			// Recursively check remaining parts
			subParts := parts[:len(parts)-1]
			if matchesComplex(ancestor, subParts) {
				return true
			}
		}
		ancestor = ancestor.Parent
	}
	return false
}

// matchesSelector checks if a DOM node matches a simple CSS selector.
func matchesSelector(node *dom.Node, sel string) bool {
	if node.Type != dom.ElementNode {
		return false
	}
	sel = strings.TrimSpace(sel)
	if sel == "" || sel == "*" {
		return true
	}

	// Parse selector into parts: tag, #id, .class, [attr], :pseudo
	return matchesSelectorParts(node, sel)
}

func matchesSelectorParts(node *dom.Node, sel string) bool {
	// Split selector into simple parts
	remaining := sel
	tagMatched := false

	for len(remaining) > 0 {
		if remaining[0] == '#' {
			// ID selector
			end := findSelectorTokenEnd(remaining[1:])
			id := remaining[1 : 1+end]
			if node.GetAttribute("id") != id {
				return false
			}
			remaining = remaining[1+end:]
		} else if remaining[0] == '.' {
			// Class selector
			end := findSelectorTokenEnd(remaining[1:])
			cls := remaining[1 : 1+end]
			if !hasClass(node, cls) {
				return false
			}
			remaining = remaining[1+end:]
		} else if remaining[0] == '[' {
			// Attribute selector
			end := strings.Index(remaining, "]")
			if end < 0 {
				break
			}
			attr := remaining[1:end]
			remaining = remaining[end+1:]
			if !matchesAttrSelector(node, attr) {
				return false
			}
		} else if remaining[0] == ':' {
			// Pseudo-class - simplified handling
			end := findSelectorTokenEnd(remaining[1:])
			pseudo := remaining[1 : 1+end]
			remaining = remaining[1+end:]
			if !matchesPseudo(node, pseudo) {
				return false
			}
		} else {
			// Tag name
			end := findSelectorTokenEnd(remaining)
			tag := remaining[:end]
			remaining = remaining[end:]
			if tag != "*" && strings.ToLower(node.Data) != strings.ToLower(tag) {
				if !tagMatched {
					return false
				}
			}
			tagMatched = true
		}
	}
	return true
}

func findSelectorTokenEnd(s string) int {
	for i, c := range s {
		if c == '.' || c == '#' || c == '[' || c == ':' || c == ' ' || c == '(' {
			return i
		}
	}
	return len(s)
}

func matchesAttrSelector(node *dom.Node, attr string) bool {
	if strings.Contains(attr, "=") {
		// [attr=val], [attr^=val], [attr$=val], [attr*=val], [attr~=val]
		op := "="
		for _, o := range []string{"^=", "$=", "*=", "~=", "|="} {
			if strings.Contains(attr, o) {
				op = o
				break
			}
		}
		idx := strings.Index(attr, op)
		attrName := strings.TrimSpace(attr[:idx])
		attrVal := strings.Trim(strings.TrimSpace(attr[idx+len(op):]), `"'`)
		nodeVal := node.GetAttribute(attrName)
		switch op {
		case "=":
			return nodeVal == attrVal
		case "^=":
			return strings.HasPrefix(nodeVal, attrVal)
		case "$=":
			return strings.HasSuffix(nodeVal, attrVal)
		case "*=":
			return strings.Contains(nodeVal, attrVal)
		case "~=":
			for _, c := range strings.Fields(nodeVal) {
				if c == attrVal {
					return true
				}
			}
			return false
		case "|=":
			return nodeVal == attrVal || strings.HasPrefix(nodeVal, attrVal+"-")
		}
	}
	// [attr] - just check existence
	attrName := strings.TrimSpace(attr)
	_, ok := node.Attributes[attrName]
	return ok
}

func matchesPseudo(node *dom.Node, pseudo string) bool {
	// Strip parentheses for functional pseudo-classes
	name := pseudo
	if idx := strings.Index(pseudo, "("); idx >= 0 {
		name = pseudo[:idx]
	}
	switch strings.ToLower(name) {
	case "first-child":
		return isFirstChild(node)
	case "last-child":
		return isLastChild(node)
	case "first-of-type":
		return isFirstOfType(node)
	case "last-of-type":
		return isLastOfType(node)
	case "not":
		inner := pseudo[strings.Index(pseudo, "(")+1 : strings.LastIndex(pseudo, ")")]
		return !matchesSelector(node, inner)
	case "nth-child", "nth-of-type":
		return true // simplified
	case "hover", "focus", "active", "visited", "checked", "disabled", "enabled":
		return false
	case "root":
		return node.Parent == nil || node.Parent.Type == dom.DocumentNode
	case "empty":
		return len(node.Children) == 0
	}
	return false
}

func isFirstChild(n *dom.Node) bool {
	if n.Parent == nil {
		return true
	}
	for _, c := range n.Parent.Children {
		if c.Type == dom.ElementNode {
			return c == n
		}
	}
	return false
}

func isLastChild(n *dom.Node) bool {
	if n.Parent == nil {
		return true
	}
	for i := len(n.Parent.Children) - 1; i >= 0; i-- {
		if n.Parent.Children[i].Type == dom.ElementNode {
			return n.Parent.Children[i] == n
		}
	}
	return false
}

func isFirstOfType(n *dom.Node) bool {
	if n.Parent == nil {
		return true
	}
	for _, c := range n.Parent.Children {
		if c.Type == dom.ElementNode && c.Data == n.Data {
			return c == n
		}
	}
	return false
}

func isLastOfType(n *dom.Node) bool {
	if n.Parent == nil {
		return true
	}
	for i := len(n.Parent.Children) - 1; i >= 0; i-- {
		c := n.Parent.Children[i]
		if c.Type == dom.ElementNode && c.Data == n.Data {
			return c == n
		}
	}
	return false
}

func hasClass(node *dom.Node, cls string) bool {
	classes := strings.Fields(node.GetAttribute("class"))
	for _, c := range classes {
		if c == cls {
			return true
		}
	}
	return false
}

func countChildren(node *dom.Node) int {
	count := 0
	for _, c := range node.Children {
		if c.Type == dom.ElementNode {
			count++
		}
	}
	return count
}

func domNodeType(node *dom.Node) int {
	switch node.Type {
	case dom.ElementNode:
		return 1
	case dom.TextNode:
		return 3
	case dom.DocumentNode:
		return 9
	}
	return 0
}

func getInnerHTML(node *dom.Node) string {
	var sb strings.Builder
	for _, child := range node.Children {
		writeHTML(&sb, child)
	}
	return sb.String()
}

func writeHTML(sb *strings.Builder, node *dom.Node) {
	switch node.Type {
	case dom.TextNode:
		sb.WriteString(node.Data)
	case dom.ElementNode:
		sb.WriteString("<")
		sb.WriteString(node.Data)
		for k, v := range node.Attributes {
			sb.WriteString(" ")
			sb.WriteString(k)
			sb.WriteString(`="`)
			sb.WriteString(v)
			sb.WriteString(`"`)
		}
		sb.WriteString(">")
		for _, child := range node.Children {
			writeHTML(sb, child)
		}
		sb.WriteString("</")
		sb.WriteString(node.Data)
		sb.WriteString(">")
	}
}

func getTextContent(node *dom.Node) string {
	var sb strings.Builder
	getTextContentInner(node, &sb)
	return sb.String()
}

func getTextContentInner(node *dom.Node, sb *strings.Builder) {
	if node.Type == dom.TextNode {
		sb.WriteString(node.Data)
		return
	}
	for _, child := range node.Children {
		getTextContentInner(child, sb)
	}
}

func setTextContent(node *dom.Node, text string) {
	node.Children = []*dom.Node{dom.NewText(text)}
}

func setInnerHTML(node *dom.Node, htmlStr string) {
	// Parse the HTML string and set as children
	// Import the html package to parse
	parsed := parseHTMLFragment(htmlStr)
	node.Children = parsed
	for _, child := range node.Children {
		child.Parent = node
	}
}

func removeChild(parent *dom.Node, child *dom.Node) {
	var newChildren []*dom.Node
	for _, c := range parent.Children {
		if c != child {
			newChildren = append(newChildren, c)
		}
	}
	parent.Children = newChildren
}

func insertBefore(parent *dom.Node, newChild, refChild *dom.Node) {
	var newChildren []*dom.Node
	for _, c := range parent.Children {
		if c == refChild {
			newChildren = append(newChildren, newChild)
		}
		newChildren = append(newChildren, c)
	}
	if len(newChildren) == len(parent.Children) {
		// refChild not found, append
		newChildren = append(newChildren, newChild)
	}
	newChild.Parent = parent
	parent.Children = newChildren
}

func cloneNode(node *dom.Node, deep bool) *dom.Node {
	cloned := &dom.Node{
		Type: node.Type,
		Data: node.Data,
		Attributes: make(map[string]string),
		Children: make([]*dom.Node, 0),
	}
	for k, v := range node.Attributes {
		cloned.Attributes[k] = v
	}
	if deep {
		for _, child := range node.Children {
			clonedChild := cloneNode(child, true)
			clonedChild.Parent = cloned
			cloned.Children = append(cloned.Children, clonedChild)
		}
	}
	return cloned
}

func nodeContains(parent, child *dom.Node) bool {
	if child == nil {
		return false
	}
	if parent == child {
		return true
	}
	for _, c := range parent.Children {
		if nodeContains(c, child) {
			return true
		}
	}
	return false
}

func insertAdjacentHTML(node *dom.Node, position, htmlStr string) {
	newNodes := parseHTMLFragment(htmlStr)
	switch position {
	case "beforebegin":
		if node.Parent != nil {
			var newChildren []*dom.Node
			for _, c := range node.Parent.Children {
				if c == node {
					newChildren = append(newChildren, newNodes...)
				}
				newChildren = append(newChildren, c)
			}
			node.Parent.Children = newChildren
		}
	case "afterbegin":
		node.Children = append(newNodes, node.Children...)
	case "beforeend":
		node.Children = append(node.Children, newNodes...)
	case "afterend":
		if node.Parent != nil {
			var newChildren []*dom.Node
			for _, c := range node.Parent.Children {
				newChildren = append(newChildren, c)
				if c == node {
					newChildren = append(newChildren, newNodes...)
				}
			}
			node.Parent.Children = newChildren
		}
	}
}

// parseHTMLFragment parses an HTML string and returns a list of nodes.
func parseHTMLFragment(htmlStr string) []*dom.Node {
	// Use a simple wrapper
	wrapped := "<div>" + htmlStr + "</div>"
	// We need to import html package - but to avoid circular deps we'll do simple parsing inline
	return parseSimpleHTML(wrapped)
}

// parseSimpleHTML parses HTML and returns children of the root element.
func parseSimpleHTML(s string) []*dom.Node {
	// Very simple tokenizer for innerHTML setting
	var nodes []*dom.Node
	i := 0
	for i < len(s) {
		if s[i] == '<' {
			// Find end of tag
			end := strings.IndexByte(s[i:], '>')
			if end < 0 {
				break
			}
			tag := s[i+1 : i+end]
			i += end + 1
			if strings.HasPrefix(tag, "/") {
				// closing tag
				_ = strings.TrimSpace(tag[1:])
				break
			}
			if strings.HasPrefix(tag, "!") {
				continue
			}
			// Parse tag name and attributes
			selfClose := strings.HasSuffix(tag, "/")
			if selfClose {
				tag = tag[:len(tag)-1]
			}
			parts := strings.Fields(tag)
			if len(parts) == 0 {
				continue
			}
			tagName := strings.ToLower(parts[0])
			node := dom.NewElement(tagName)
			// Parse attributes (simplified)
			for _, part := range parts[1:] {
				if strings.Contains(part, "=") {
					idx := strings.Index(part, "=")
					k := part[:idx]
					v := strings.Trim(part[idx+1:], `"'`)
					node.SetAttribute(k, v)
				} else {
					node.SetAttribute(part, "")
				}
			}
			if !selfClose && !isVoidElement(tagName) {
				// Parse children
				children := parseSimpleHTML(s[i:])
				for _, child := range children {
					node.AppendChild(child)
				}
			}
			nodes = append(nodes, node)
		} else {
			// Text node
			end := strings.IndexByte(s[i:], '<')
			var text string
			if end < 0 {
				text = s[i:]
				i = len(s)
			} else {
				text = s[i : i+end]
				i += end
			}
			if text != "" {
				nodes = append(nodes, dom.NewText(text))
			}
		}
	}
	return nodes
}

func isVoidElement(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

// parseInlineStyle parses a CSS inline style string into a map.
func parseInlineStyle(styleStr string) map[string]string {
	result := make(map[string]string)
	for _, decl := range strings.Split(styleStr, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		idx := strings.Index(decl, ":")
		if idx < 0 {
			continue
		}
		prop := strings.TrimSpace(decl[:idx])
		val := strings.TrimSpace(decl[idx+1:])
		result[prop] = val
	}
	return result
}

// updateNodeStyle updates a single CSS property in a node's inline style.
func updateNodeStyle(node *dom.Node, prop, val string) {
	styles := parseInlineStyle(node.GetAttribute("style"))
	if val == "" {
		delete(styles, prop)
	} else {
		styles[prop] = val
	}
	node.SetAttribute("style", buildInlineStyle(styles))
}

func buildInlineStyle(styles map[string]string) string {
	var parts []string
	for k, v := range styles {
		parts = append(parts, k+": "+v)
	}
	return strings.Join(parts, "; ")
}

// cssPropertyToCamel converts a CSS property name to camelCase.
// e.g. "background-color" -> "backgroundColor"
func cssPropertyToCamel(prop string) string {
	parts := strings.Split(prop, "-")
	if len(parts) == 1 {
		return prop
	}
	var sb strings.Builder
	sb.WriteString(parts[0])
	for _, part := range parts[1:] {
		if len(part) > 0 {
			sb.WriteString(strings.ToUpper(part[:1]))
			sb.WriteString(part[1:])
		}
	}
	return sb.String()
}

// camelToCSS converts camelCase to CSS property name.
// e.g. "backgroundColor" -> "background-color"
func camelToCSS(prop string) string {
	var sb strings.Builder
	for i, ch := range prop {
		if unicode.IsUpper(ch) && i > 0 {
			sb.WriteRune('-')
			sb.WriteRune(unicode.ToLower(ch))
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// dataToCamel converts a data attribute name (after "data-") to camelCase.
// e.g. "foo-bar" -> "fooBar"
func dataToCamel(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		return s
	}
	var sb strings.Builder
	sb.WriteString(parts[0])
	for _, part := range parts[1:] {
		if len(part) > 0 {
			sb.WriteString(strings.ToUpper(part[:1]))
			sb.WriteString(part[1:])
		}
	}
	return sb.String()
}

// ---- Built-in objects ----

func makeMathObject() *Object {
	m := NewObject()
	m.Set("PI", NumberVal(3.141592653589793))
	m.Set("E", NumberVal(2.718281828459045))
	m.Set("LN2", NumberVal(0.6931471805599453))
	m.Set("LN10", NumberVal(2.302585092994046))
	m.Set("LOG2E", NumberVal(1.4426950408889634))
	m.Set("LOG10E", NumberVal(0.4342944819032518))
	m.Set("SQRT2", NumberVal(1.4142135623730951))
	m.Set("abs", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Abs(args[0].ToNumber()))
	}))
	m.Set("ceil", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Ceil(args[0].ToNumber()))
	}))
	m.Set("floor", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Floor(args[0].ToNumber()))
	}))
	m.Set("round", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Round(args[0].ToNumber()))
	}))
	m.Set("trunc", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Trunc(args[0].ToNumber()))
	}))
	m.Set("sqrt", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Sqrt(args[0].ToNumber()))
	}))
	m.Set("cbrt", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Cbrt(args[0].ToNumber()))
	}))
	m.Set("pow", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return NumberVal(nan())
		}
		return NumberVal(math.Pow(args[0].ToNumber(), args[1].ToNumber()))
	}))
	m.Set("log", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Log(args[0].ToNumber()))
	}))
	m.Set("log2", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Log2(args[0].ToNumber()))
	}))
	m.Set("log10", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Log10(args[0].ToNumber()))
	}))
	m.Set("exp", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Exp(args[0].ToNumber()))
	}))
	m.Set("sin", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Sin(args[0].ToNumber()))
	}))
	m.Set("cos", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Cos(args[0].ToNumber()))
	}))
	m.Set("tan", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Tan(args[0].ToNumber()))
	}))
	m.Set("atan", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Atan(args[0].ToNumber()))
	}))
	m.Set("atan2", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return NumberVal(nan())
		}
		return NumberVal(math.Atan2(args[0].ToNumber(), args[1].ToNumber()))
	}))
	m.Set("asin", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Asin(args[0].ToNumber()))
	}))
	m.Set("acos", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(math.Acos(args[0].ToNumber()))
	}))
	m.Set("max", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(math.Inf(-1))
		}
		result := args[0].ToNumber()
		for _, a := range args[1:] {
			n := a.ToNumber()
			if n > result {
				result = n
			}
		}
		return NumberVal(result)
	}))
	m.Set("min", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(math.Inf(1))
		}
		result := args[0].ToNumber()
		for _, a := range args[1:] {
			n := a.ToNumber()
			if n < result {
				result = n
			}
		}
		return NumberVal(result)
	}))
	m.Set("random", makeFn(func(_ *Value, _ []*Value) *Value {
		// Deterministic pseudo-random for reproducibility
		pseudoRandState = (pseudoRandState*6364136223846793005 + 1442695040888963407) >> 33
		return NumberVal(float64(pseudoRandState%1000000) / 1000000.0)
	}))
	m.Set("sign", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		n := args[0].ToNumber()
		if n > 0 {
			return NumberVal(1)
		}
		if n < 0 {
			return NumberVal(-1)
		}
		return NumberVal(0)
	}))
	m.Set("hypot", makeFn(func(_ *Value, args []*Value) *Value {
		sum := 0.0
		for _, a := range args {
			n := a.ToNumber()
			sum += n * n
		}
		return NumberVal(math.Sqrt(sum))
	}))
	m.Set("clz32", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(32)
		}
		n := uint32(int32(args[0].ToNumber()))
		if n == 0 {
			return NumberVal(32)
		}
		count := 0
		for (n & 0x80000000) == 0 {
			count++
			n <<= 1
		}
		return NumberVal(float64(count))
	}))
	m.Set("imul", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return NumberVal(0)
		}
		a := int32(args[0].ToNumber())
		b := int32(args[1].ToNumber())
		return NumberVal(float64(a * b))
	}))
	m.Set("fround", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return NumberVal(nan())
		}
		return NumberVal(float64(float32(args[0].ToNumber())))
	}))
	return m
}

var pseudoRandState uint64 = 12345678

func makeJSONObject() *Object {
	j := NewObject()
	j.Set("stringify", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		indent := ""
		if len(args) > 2 && !args[2].IsUndefined() {
			if args[2].typ == TypeNumber {
				indent = strings.Repeat(" ", int(args[2].ToNumber()))
			} else {
				indent = args[2].ToString()
			}
		}
		result := jsonStringify(args[0], indent, "")
		if result == "" {
			return Undefined
		}
		return StringVal(result)
	}))
	j.Set("parse", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			panic(throwSignal{val: StringVal("SyntaxError: Unexpected end of JSON input")})
		}
		// Simplified JSON parser
		result, _ := jsonParse(strings.TrimSpace(args[0].ToString()))
		return result
	}))
	return j
}

func jsonStringify(v *Value, indent, currentIndent string) string {
	switch v.typ {
	case TypeUndefined:
		return ""
	case TypeNull:
		return "null"
	case TypeBoolean:
		if v.boolVal {
			return "true"
		}
		return "false"
	case TypeNumber:
		if math.IsNaN(v.numVal) || math.IsInf(v.numVal, 0) {
			return "null"
		}
		return numberToString(v.numVal)
	case TypeString:
		return jsonQuoteString(v.strVal)
	case TypeFunction:
		return ""
	case TypeObject:
		obj := v.objVal
		if obj.isArray {
			var sb strings.Builder
			sb.WriteString("[")
			elems := obj.ArrayElements()
			for i, e := range elems {
				if i > 0 {
					sb.WriteString(",")
				}
				if indent != "" {
					sb.WriteString("\n" + currentIndent + indent)
				}
				s := jsonStringify(e, indent, currentIndent+indent)
				if s == "" {
					s = "null"
				}
				sb.WriteString(s)
			}
			if indent != "" && len(elems) > 0 {
				sb.WriteString("\n" + currentIndent)
			}
			sb.WriteString("]")
			return sb.String()
		}
		var sb strings.Builder
		sb.WriteString("{")
		first := true
		for k, val := range obj.props {
			s := jsonStringify(val, indent, currentIndent+indent)
			if s == "" {
				continue
			}
			if !first {
				sb.WriteString(",")
			}
			first = false
			if indent != "" {
				sb.WriteString("\n" + currentIndent + indent)
			}
			sb.WriteString(jsonQuoteString(k))
			sb.WriteString(":")
			if indent != "" {
				sb.WriteString(" ")
			}
			sb.WriteString(s)
		}
		if indent != "" && !first {
			sb.WriteString("\n" + currentIndent)
		}
		sb.WriteString("}")
		return sb.String()
	}
	return "null"
}

func jsonQuoteString(s string) string {
	var sb strings.Builder
	sb.WriteRune('"')
	for _, ch := range s {
		switch ch {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(ch)
		}
	}
	sb.WriteRune('"')
	return sb.String()
}

func jsonParse(s string) (*Value, int) {
	if len(s) == 0 {
		return Undefined, 0
	}
	switch s[0] {
	case '"':
		end := 1
		var sb strings.Builder
		for end < len(s) {
			if s[end] == '"' {
				end++
				break
			}
			if s[end] == '\\' && end+1 < len(s) {
				end++
				switch s[end] {
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case '"':
					sb.WriteRune('"')
				case '\\':
					sb.WriteRune('\\')
				default:
					sb.WriteByte(s[end])
				}
			} else {
				sb.WriteByte(s[end])
			}
			end++
		}
		return StringVal(sb.String()), end
	case '{':
		obj := NewObject()
		i := 1
		for i < len(s) {
			for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r') {
				i++
			}
			if i >= len(s) || s[i] == '}' {
				i++
				break
			}
			if s[i] == ',' {
				i++
				continue
			}
			key, n := jsonParse(s[i:])
			i += n
			for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == ':') {
				i++
			}
			val, n2 := jsonParse(strings.TrimSpace(s[i:]))
			i += n2
			obj.Set(key.ToString(), val)
		}
		return ObjectVal(obj), i
	case '[':
		arr := NewArray()
		i := 1
		for i < len(s) {
			for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r') {
				i++
			}
			if i >= len(s) || s[i] == ']' {
				i++
				break
			}
			if s[i] == ',' {
				i++
				continue
			}
			val, n := jsonParse(s[i:])
			i += n
			arr.Push(val)
		}
		return ObjectVal(arr), i
	case 't':
		if strings.HasPrefix(s, "true") {
			return BoolVal(true), 4
		}
	case 'f':
		if strings.HasPrefix(s, "false") {
			return BoolVal(false), 5
		}
	case 'n':
		if strings.HasPrefix(s, "null") {
			return Null, 4
		}
	}
	// Number
	i := 0
	for i < len(s) && (s[i] == '-' || (s[i] >= '0' && s[i] <= '9') || s[i] == '.' || s[i] == 'e' || s[i] == 'E' || s[i] == '+') {
		i++
	}
	if i > 0 {
		var n float64
		fmt.Sscanf(s[:i], "%g", &n)
		return NumberVal(n), i
	}
	return Undefined, 1
}

func makeLocationObject() *Object {
	obj := NewObject()
	obj.Set("href", StringVal(""))
	obj.Set("protocol", StringVal(""))
	obj.Set("host", StringVal(""))
	obj.Set("pathname", StringVal("/"))
	obj.Set("search", StringVal(""))
	obj.Set("hash", StringVal(""))
	obj.Set("hostname", StringVal(""))
	obj.Set("port", StringVal(""))
	obj.Set("assign", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("replace", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("reload", makeFn(func(_ *Value, _ []*Value) *Value { return Undefined }))
	obj.Set("toString", makeFn(func(_ *Value, _ []*Value) *Value { return StringVal("") }))
	return obj
}

func makeNavigatorObject() *Object {
	obj := NewObject()
	obj.Set("userAgent", StringVal("GoLangBrowser/1.0"))
	obj.Set("language", StringVal("en-US"))
	obj.Set("languages", func() *Value {
		arr := NewArray()
		arr.Push(StringVal("en-US"))
		arr.Push(StringVal("en"))
		return ObjectVal(arr)
	}())
	obj.Set("platform", StringVal("Go"))
	obj.Set("cookieEnabled", BoolVal(false))
	obj.Set("onLine", BoolVal(false))
	return obj
}

func makeMapObject(interp *Interpreter) *Object {
	obj := NewObject()
	var keys []*Value
	var vals []*Value

	findKey := func(k *Value) int {
		for i, key := range keys {
			if strictEquals(key, k) {
				return i
			}
		}
		return -1
	}

	obj.Set("set", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) < 2 {
			return ObjectVal(obj)
		}
		idx := findKey(args[0])
		if idx < 0 {
			keys = append(keys, args[0])
			vals = append(vals, args[1])
		} else {
			vals[idx] = args[1]
		}
		obj.Set("size", NumberVal(float64(len(keys))))
		return ObjectVal(obj)
	}))
	obj.Set("get", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		idx := findKey(args[0])
		if idx < 0 {
			return Undefined
		}
		return vals[idx]
	}))
	obj.Set("has", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(findKey(args[0]) >= 0)
	}))
	obj.Set("delete", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		idx := findKey(args[0])
		if idx < 0 {
			return BoolVal(false)
		}
		keys = append(keys[:idx], keys[idx+1:]...)
		vals = append(vals[:idx], vals[idx+1:]...)
		obj.Set("size", NumberVal(float64(len(keys))))
		return BoolVal(true)
	}))
	obj.Set("clear", makeFn(func(_ *Value, _ []*Value) *Value {
		keys = nil
		vals = nil
		obj.Set("size", NumberVal(0))
		return Undefined
	}))
	obj.Set("forEach", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		fn := args[0]
		for i, k := range keys {
			interp.callFunction(fn, Undefined, []*Value{vals[i], k, ObjectVal(obj)})
		}
		return Undefined
	}))
	obj.Set("keys", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for _, k := range keys {
			arr.Push(k)
		}
		return ObjectVal(arr)
	}))
	obj.Set("values", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for _, v := range vals {
			arr.Push(v)
		}
		return ObjectVal(arr)
	}))
	obj.Set("entries", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for i, k := range keys {
			pair := NewArray()
			pair.Push(k)
			pair.Push(vals[i])
			arr.Push(ObjectVal(pair))
		}
		return ObjectVal(arr)
	}))
	obj.Set("size", NumberVal(0))
	return obj
}

func makeSetObject(interp *Interpreter) *Object {
	obj := NewObject()
	var items []*Value

	hasItem := func(v *Value) bool {
		for _, item := range items {
			if strictEquals(item, v) {
				return true
			}
		}
		return false
	}

	obj.Set("add", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return ObjectVal(obj)
		}
		if !hasItem(args[0]) {
			items = append(items, args[0])
			obj.Set("size", NumberVal(float64(len(items))))
		}
		return ObjectVal(obj)
	}))
	obj.Set("has", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		return BoolVal(hasItem(args[0]))
	}))
	obj.Set("delete", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return BoolVal(false)
		}
		for i, item := range items {
			if strictEquals(item, args[0]) {
				items = append(items[:i], items[i+1:]...)
				obj.Set("size", NumberVal(float64(len(items))))
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	}))
	obj.Set("clear", makeFn(func(_ *Value, _ []*Value) *Value {
		items = nil
		obj.Set("size", NumberVal(0))
		return Undefined
	}))
	obj.Set("forEach", makeFn(func(_ *Value, args []*Value) *Value {
		if len(args) == 0 {
			return Undefined
		}
		fn := args[0]
		for _, item := range items {
			interp.callFunction(fn, Undefined, []*Value{item, item, ObjectVal(obj)})
		}
		return Undefined
	}))
	obj.Set("values", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for _, item := range items {
			arr.Push(item)
		}
		return ObjectVal(arr)
	}))
	obj.Set("keys", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for _, item := range items {
			arr.Push(item)
		}
		return ObjectVal(arr)
	}))
	obj.Set("entries", makeFn(func(_ *Value, _ []*Value) *Value {
		arr := NewArray()
		for _, item := range items {
			pair := NewArray()
			pair.Push(item)
			pair.Push(item)
			arr.Push(ObjectVal(pair))
		}
		return ObjectVal(arr)
	}))
	obj.Set("size", NumberVal(0))
	return obj
}

// ---- Helper functions ----

func makeFn(fn func(*Value, []*Value) *Value) *Value {
	obj := &Object{props: make(map[string]*Value), goFunc: fn}
	return &Value{typ: TypeFunction, objVal: obj}
}

var parseIntFn = func(_ *Value, args []*Value) *Value {
	if len(args) == 0 {
		return NumberVal(nan())
	}
	s := strings.TrimSpace(args[0].ToString())
	base := 10
	if len(args) > 1 && !args[1].IsUndefined() {
		base = int(args[1].ToNumber())
	}
	if base == 0 {
		base = 10
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	// Strip non-digit chars for the given base
	end := 0
	for end < len(s) {
		c := s[end]
		valid := false
		if base <= 10 {
			valid = c >= '0' && c < '0'+byte(base)
		} else {
			valid = (c >= '0' && c <= '9') || (c >= 'a' && c < 'a'+byte(base-10)) || (c >= 'A' && c < 'A'+byte(base-10))
		}
		if c == '-' && end == 0 {
			end++
			continue
		}
		if !valid {
			break
		}
		end++
	}
	if end == 0 || (end == 1 && s[0] == '-') {
		return NumberVal(nan())
	}
	var n int64
	fmt.Sscanf(s[:end], fmt.Sprintf("%%d"), &n)
	if base != 10 {
		parsed, err := fmt.Sscanf(s[:end], fmt.Sprintf("%%%d", base)+"d", &n)
		_ = parsed
		if err != nil {
			n2, err2 := parseInt64Base(s[:end], base)
			if err2 != nil {
				return NumberVal(nan())
			}
			return NumberVal(float64(n2))
		}
	}
	_ = n
	n2, err := parseInt64Base(s[:end], base)
	if err != nil {
		return NumberVal(nan())
	}
	return NumberVal(float64(n2))
}

func parseInt64Base(s string, base int) (int64, error) {
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var n int64
	for _, c := range s {
		var digit int64
		if c >= '0' && c <= '9' {
			digit = int64(c - '0')
		} else if c >= 'a' && c <= 'z' {
			digit = int64(c-'a') + 10
		} else if c >= 'A' && c <= 'Z' {
			digit = int64(c-'A') + 10
		} else {
			break
		}
		if digit >= int64(base) {
			break
		}
		n = n*int64(base) + digit
	}
	if neg {
		n = -n
	}
	return n, nil
}

var parseFloatFn = func(_ *Value, args []*Value) *Value {
	if len(args) == 0 {
		return NumberVal(nan())
	}
	s := strings.TrimSpace(args[0].ToString())
	var n float64
	if _, err := fmt.Sscanf(s, "%g", &n); err != nil {
		return NumberVal(nan())
	}
	return NumberVal(n)
}

func isInfOrNaN(n float64) bool {
	return math.IsInf(n, 0) || math.IsNaN(n)
}

func nan() float64 {
	return math.NaN()
}

func percentEncode(s string, isURI bool) string {
	var sb strings.Builder
	unreserved := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.!~*'()"
	uriReserved := ";,/?:@&=+$#"
	for _, ch := range []byte(s) {
		c := rune(ch)
		if strings.ContainsRune(unreserved, c) {
			sb.WriteRune(c)
		} else if isURI && strings.ContainsRune(uriReserved, c) {
			sb.WriteRune(c)
		} else {
			sb.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}
	return sb.String()
}
