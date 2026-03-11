package js

// Env is a lexical scope (environment) for variable bindings.
type Env struct {
	vars   map[string]*Value
	parent *Env
}

// NewEnv creates a new top-level environment.
func NewEnv() *Env {
	return &Env{vars: make(map[string]*Value)}
}

// NewChildEnv creates a new child environment with the given parent.
func NewChildEnv(parent *Env) *Env {
	return &Env{vars: make(map[string]*Value), parent: parent}
}

// Define creates a binding in the current scope.
func (e *Env) Define(name string, val *Value) {
	e.vars[name] = val
}

// Get looks up a binding by walking the scope chain.
func (e *Env) Get(name string) (*Value, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set updates an existing binding (walking the scope chain).
// Returns false if the binding doesn't exist (caller should create it).
func (e *Env) Set(name string, val *Value) bool {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return true
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return false
}

// SetOrDefine sets an existing binding or defines it in the global scope.
func (e *Env) SetOrDefine(name string, val *Value) {
	if !e.Set(name, val) {
		// Walk to global scope and define there (mimics var hoisting to global)
		global := e
		for global.parent != nil {
			global = global.parent
		}
		global.Define(name, val)
	}
}
