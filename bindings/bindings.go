// Package bindings provides an implementation of the goshua.Bindings interface.
package bindings

import "log"
import "goshua/goshua"
import "goshua/bindings/immutable"

// *bindings implements the goshua.Bindings interface.
type bindings struct {
	ply *immutable.Ply
}

func emptyBindings() goshua.Bindings {
	return &bindings{ply: immutable.EmptyPly()}
}

func init() {
	goshua.EmptyBindings = emptyBindings
}

func (b *bindings) Dump() {
	b.ply.Dump()
}

func (b *bindings) Get(v goshua.Variable) (interface{}, bool) {
	for p := b.ply; p != nil; p = p.Previous() {
		if p.Has(v) {
			if p.HasValue() {
				return p.Value(), true
			}
			return nil, false
		}
	}
	return nil, false
}

func (b *bindings) Bind(v goshua.Variable, other interface{}) (goshua.Bindings, bool) {
	log.Printf("Binding %s to %#v", v.Name(), other)
	variables := make(map[goshua.Variable]bool)
	variables[v] = true
	var hasValue bool
	var value interface{}
	if v1, ok := other.(goshua.Variable); ok {
		log.Printf("Linking variables %s and %s", v, v1)
		variables[v1] = true
		value = nil
		hasValue = false
	} else {
		value = other
		hasValue = true
	}

	// What other variables and values are equivalent to these?
	for p := b.ply; p != nil; p = p.Previous() {
		if !p.HasAny(variables) {
			continue
		}
		p.GetVariables(variables)
		if p.HasValue() {
			// This ply provides a value.  If we already have a value then it better
			// match.
			if hasValue {
				eq, err := goshua.Equal(p.Value(), value)
				if err != nil {
					log.Printf("%s", err.Error())
					return b, false
				}
				if !eq {
					// Inconsistent values.  Fail.
					return b, false
				}
			}
			// Use the value from this Ply
			value = p.Value()
			hasValue = true
		}
	}
	/* Shouldn't need this
	   if hasValue {
	      if eq, ok := goshua.Equal(value, other); !(ok && eq) {
	      	 // Inconsistent values.  Fail.
		 return b, false
	      }
	   }
	*/
	return &bindings{ply: immutable.NewPly(
		variables,
		value, hasValue,
		b.ply)}, true
}
