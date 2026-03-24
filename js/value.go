package js

import (
	"fmt"
	"math"
	"strings"
)

// Type represents a JavaScript value type.
type Type int

const (
	TypeUndefined Type = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeFunction
)

// Value represents a JavaScript value.
type Value struct {
	typ      Type
	boolVal  bool
	numVal   float64
	strVal   string
	objVal   *Object
}

// Undefined is the JavaScript undefined value.
var Undefined = &Value{typ: TypeUndefined}

// Null is the JavaScript null value.
var Null = &Value{typ: TypeNull}

// BoolVal creates a boolean value.
func BoolVal(b bool) *Value { return &Value{typ: TypeBoolean, boolVal: b} }

// NumberVal creates a number value.
func NumberVal(n float64) *Value { return &Value{typ: TypeNumber, numVal: n} }

// StringVal creates a string value.
func StringVal(s string) *Value { return &Value{typ: TypeString, strVal: s} }

// ObjectVal creates an object value from an Object.
func ObjectVal(o *Object) *Value {
	if o == nil {
		return Null
	}
	if o.isFunction() {
		return &Value{typ: TypeFunction, objVal: o}
	}
	return &Value{typ: TypeObject, objVal: o}
}

// Type returns the value's type.
func (v *Value) Type() Type { return v.typ }

// IsUndefined returns true if the value is undefined.
func (v *Value) IsUndefined() bool { return v.typ == TypeUndefined }

// IsNull returns true if the value is null.
func (v *Value) IsNull() bool { return v.typ == TypeNull }

// IsNullish returns true if the value is null or undefined.
func (v *Value) IsNullish() bool { return v.typ == TypeUndefined || v.typ == TypeNull }

// ToBoolean converts the value to a boolean per JS semantics.
func (v *Value) ToBoolean() bool {
	switch v.typ {
	case TypeUndefined, TypeNull:
		return false
	case TypeBoolean:
		return v.boolVal
	case TypeNumber:
		return v.numVal != 0 && !math.IsNaN(v.numVal)
	case TypeString:
		return v.strVal != ""
	case TypeObject, TypeFunction:
		return true
	}
	return false
}

// ToNumber converts the value to a number per JS semantics.
func (v *Value) ToNumber() float64 {
	switch v.typ {
	case TypeUndefined:
		return math.NaN()
	case TypeNull:
		return 0
	case TypeBoolean:
		if v.boolVal {
			return 1
		}
		return 0
	case TypeNumber:
		return v.numVal
	case TypeString:
		s := strings.TrimSpace(v.strVal)
		if s == "" {
			return 0
		}
		var n float64
		if _, err := fmt.Sscanf(s, "%g", &n); err == nil {
			return n
		}
		return math.NaN()
	case TypeObject, TypeFunction:
		prim := v.objVal.toPrimitive("number")
		return prim.ToNumber()
	}
	return math.NaN()
}

// ToString converts the value to a string per JS semantics.
func (v *Value) ToString() string {
	switch v.typ {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		if v.boolVal {
			return "true"
		}
		return "false"
	case TypeNumber:
		return numberToString(v.numVal)
	case TypeString:
		return v.strVal
	case TypeObject, TypeFunction:
		prim := v.objVal.toPrimitive("string")
		return prim.ToString()
	}
	return "undefined"
}

func numberToString(n float64) string {
	if math.IsNaN(n) {
		return "NaN"
	}
	if math.IsInf(n, 1) {
		return "Infinity"
	}
	if math.IsInf(n, -1) {
		return "-Infinity"
	}
	if n == math.Trunc(n) && math.Abs(n) < 1e15 {
		return fmt.Sprintf("%d", int64(n))
	}
	return fmt.Sprintf("%g", n)
}

// ToObject returns the underlying Object (or nil for primitives).
func (v *Value) ToObject() *Object {
	if v.typ == TypeObject || v.typ == TypeFunction {
		return v.objVal
	}
	return nil
}

// Repr returns a debug representation of the value.
func (v *Value) Repr() string {
	switch v.typ {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		if v.boolVal {
			return "true"
		}
		return "false"
	case TypeNumber:
		return numberToString(v.numVal)
	case TypeString:
		return fmt.Sprintf("%q", v.strVal)
	case TypeFunction:
		return "[Function]"
	case TypeObject:
		if v.objVal != nil && v.objVal.isArray {
			return "[Array]"
		}
		return "[Object]"
	}
	return "undefined"
}

// Object represents a JavaScript object.
type Object struct {
	props     map[string]*Value
	proto     *Object
	isArray   bool
	// For function objects
	funcNode  Node   // *FunctionExpression, *FunctionDeclaration, *ArrowFunctionExpression
	funcEnv   *Env
	goFunc    func(this *Value, args []*Value) *Value
	// propSetHook is called after a property is set (used by DOM wrappers to sync
	// property writes back to the underlying DOM node).
	propSetHook func(key string, val *Value)
}

// NewObject creates a new empty object.
func NewObject() *Object {
	return &Object{props: make(map[string]*Value)}
}

// NewArray creates a new array object.
func NewArray() *Object {
	arr := &Object{props: make(map[string]*Value), isArray: true}
	arr.props["length"] = NumberVal(0)
	return arr
}

func (o *Object) isFunction() bool {
	return o.funcNode != nil || o.goFunc != nil
}

// Get retrieves a property value, checking the prototype chain.
func (o *Object) Get(key string) *Value {
	if v, ok := o.props[key]; ok {
		return v
	}
	if o.proto != nil {
		return o.proto.Get(key)
	}
	return Undefined
}

// Set sets a property value.
func (o *Object) Set(key string, v *Value) {
	if o.props == nil {
		o.props = make(map[string]*Value)
	}
	o.props[key] = v
	// Update length for arrays
	if o.isArray {
		if key != "length" {
			// If setting numeric index, update length
			var idx int
			if _, err := fmt.Sscanf(key, "%d", &idx); err == nil {
				curLen := int(o.Get("length").ToNumber())
				if idx >= curLen {
					o.props["length"] = NumberVal(float64(idx + 1))
				}
			}
		}
	}
	// Notify DOM wrapper hook
	if o.propSetHook != nil {
		o.propSetHook(key, v)
	}
}

// Delete removes a property.
func (o *Object) Delete(key string) {
	delete(o.props, key)
}

// Keys returns all own property keys (excluding prototype chain).
func (o *Object) Keys() []string {
	keys := make([]string, 0, len(o.props))
	for k := range o.props {
		if k != "length" || !o.isArray {
			keys = append(keys, k)
		}
	}
	return keys
}

// ArrayElements returns elements of an array in order.
func (o *Object) ArrayElements() []*Value {
	if !o.isArray {
		return nil
	}
	n := int(o.Get("length").ToNumber())
	elems := make([]*Value, n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("%d", i)
		if v, ok := o.props[key]; ok {
			elems[i] = v
		} else {
			elems[i] = Undefined
		}
	}
	return elems
}

// Push appends a value to an array.
func (o *Object) Push(v *Value) {
	n := int(o.Get("length").ToNumber())
	o.Set(fmt.Sprintf("%d", n), v)
	o.props["length"] = NumberVal(float64(n + 1))
}

// toPrimitive converts to primitive (for object->string/number coercion).
func (o *Object) toPrimitive(hint string) *Value {
	if hint == "string" {
		if toString := o.Get("toString"); toString.typ == TypeFunction {
			// call toString
			result := toString.objVal.call(ObjectVal(o), nil)
			if result.typ != TypeObject && result.typ != TypeFunction {
				return result
			}
		}
		return StringVal("[object Object]")
	}
	// number hint
	if valueOf := o.Get("valueOf"); valueOf.typ == TypeFunction {
		result := valueOf.objVal.call(ObjectVal(o), nil)
		if result.typ != TypeObject && result.typ != TypeFunction {
			return result
		}
	}
	return NumberVal(math.NaN())
}

// call invokes the object as a function.
func (o *Object) call(this *Value, args []*Value) *Value {
	if o.goFunc != nil {
		return o.goFunc(this, args)
	}
	return Undefined
}
