// Package immutable implements an immutable binding variable binding
// frame.
package immutable

import "log"
import "goshua/goshua"

type Ply struct {
	variables []goshua.Variable
	value     interface{}
	hasValue  bool
	previous  *Ply
}

func EmptyPly() *Ply {
	return &Ply{}
}

func NewPly(variables map[goshua.Variable]bool, value interface{}, hasValue bool, previous *Ply) *Ply {
	copy := []goshua.Variable{}
	for v, has := range variables {
		if has {
			copy = append(copy, v)
		}
	}
	return &Ply{
		variables: copy,
		value:     value,
		hasValue:  hasValue,
		previous:  previous,
	}
}

func (p *Ply) Has(variable goshua.Variable) bool {
	for _, v := range p.variables {
		if variable == v {
			return true
		}
	}
	return false
}

func (p *Ply) HasAny(collection map[goshua.Variable]bool) bool {
	for v, has := range collection {
		if has {
			if p.Has(v) {
				return true
			}
		}
	}
	return false
}

func (p *Ply) GetVariables(collection map[goshua.Variable]bool) {
	for _, v := range p.variables {
		collection[v] = true
	}
}

func (p *Ply) HasValue() bool {
	return p.hasValue
}

func (p *Ply) Value() interface{} {
	return p.value
}

func (p *Ply) Previous() *Ply {
	return p.previous
}

func (p *Ply) Dump() {
	log.Printf("ply.Dump")
	for i, p1 := 0, p; p1 != nil; i, p1 = i+1, p1.previous {
		vars := ""
		for _, v := range p1.variables {
			if vars != "" {
				vars += ", "
			}
			vars += v.Name()
		}
		log.Printf("  %2d %s: %v %#v", i, vars, p1.hasValue, p1.value)
	}
}
