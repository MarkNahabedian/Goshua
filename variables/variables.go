// Package variables implements logic variables.
// A logic variable has a name and a scope.
// Two logic variables with the same name but in different scopes are
// different logic variables.
package variables

import "fmt"
import "log"
import "goshua/goshua"

// *scope implements the goshua.Scope interface.
type scope struct {
	id    uint64
	index map[string]goshua.Variable
}

func newScope() goshua.Scope {
	s := &scope{index: make(map[string]goshua.Variable)}
	// TODO: set id to a unique value
	return s
}

// Compile time check that we're implementing goshua.Scope.
var _ goshua.Scope = newScope()

func init() {
	goshua.NewScope = newScope
}

func (s *scope) Lookup(name string) goshua.Variable {
	if v, ok := s.index[name]; ok {
		return v
	}
	v := &variable{
		name:  name,
		scope: s,
	}
	s.index[name] = v
	return v
}

// *variable implements the goshua.Variable interface.
type variable struct {
	name  string
	scope *scope
}

func (v *variable) String() string {
	return fmt.Sprintf("?%s", v.Name())
}

// ^variable satisfies the goshua.Variable interface.
func (v *variable) IsLogicVariable() {}

func (v *variable) Name() string {
	return v.name
}

func (v *variable) SameAs(other goshua.Variable) bool {
	return v == other
}

func (v *variable) HasVariables() bool {
	return true
}

func (v *variable) Unify(other interface{},
	bindings goshua.Bindings,
	continuation func(goshua.Bindings)) {
	if b, ok := bindings.Bind(v, other); ok {
		// val, has := b.Get(v)
		// log.Printf("bound %s to %#v: %v %#v", v.Name(), other, has, val)
		// b.Dump()
		continuation(b)
	} else {
		existing, _ := b.Get(v)
		log.Printf("variable.Unify: Bind failed, %s value is %v", v, existing)
	}
}

// Compile time check that *variable implements goshua.Variable.
var _ goshua.Variable = newScope().Lookup("a")
