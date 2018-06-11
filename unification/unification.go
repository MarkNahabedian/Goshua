// Package unification implements expert system unification.
package unification

import "log"
import "reflect"
import "goshua/goshua"

func unify(thing1, thing2 interface{}, b goshua.Bindings,
	continuation func(goshua.Bindings)) {
	// Variable implements Unifier
	if thing1, ok := thing1.(goshua.Unifier); ok {
		thing1.Unify(thing2, b, continuation)
		return
	}
	if thing2, ok := thing2.(goshua.Unifier); ok {
		thing2.Unify(thing1, b, continuation)
		return
	}

	for _, u := range typeUnifiers {
		if u.test(thing1) && u.test(thing2) {
			// log.Printf("typeUnifier %T for %v, %v\n", u, thing1, thing2)
			u.unify(thing1, thing2, b, continuation)
			return
		}
	}
}

func init() {
	goshua.Unify = unify
}

// typeUnifier tells how to unify two things that both satisfy test.
type typeUnifier interface {
	// Is the argument appropriate for this unifier
	test(interface{}) bool

	// If they are then unify them
	unify(thing1, thing2 interface{},
		bindings goshua.Bindings,
		continuation func(goshua.Bindings))
}

// typeUnifiers is a dispatch table that identifies how to unify
// objects based on their type.  Both arguments of Unify must satisfy the
// test method of the same typeUnifier for the type unifier to be used.
// If some object can be more promiscuous in its unification it should
// implemengt the Unifier interface.
var typeUnifiers []typeUnifier = []typeUnifier{
	&numberUnifier{},
	&stringUnifier{},
	&sequenceUnifier{},
	&structUnifier{},
}

// equalOrFail provides a unify method for objects of types which can
// only unify if the objects are equal.
type equalOrFail struct{}

func (u *equalOrFail) unify(thing1, thing2 interface{},
	b goshua.Bindings,
	continuation func(goshua.Bindings)) {
	eq, err := goshua.Equal(thing1, thing2)
	if err != nil {
		log.Printf("%s", err.Error())
		return
	}
	if eq {
		continuation(b)
	}
}

type numberUnifier struct {
	equalOrFail
}

func (u *numberUnifier) test(thing interface{}) bool {
	switch reflect.ValueOf(thing).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return true

	default:
		return false
	}
}

type stringUnifier struct {
	equalOrFail
}

func (u *stringUnifier) test(thing interface{}) bool {
	return reflect.ValueOf(thing).Kind() == reflect.String
}

// For aggregations whose reflection interfaces provide length and index methods.
func collectionUnifier(lengthFunction func(reflect.Value) int,
	indexFunction func(reflect.Value, int) reflect.Value) func(interface{}, interface{}, goshua.Bindings, func(goshua.Bindings)) {
	return func(thing1, thing2 interface{}, b goshua.Bindings, continuation func(goshua.Bindings)) {
		/*
			     // It's easy if they're identical.
			     // I don't know how to test identity in go though.
			     if thing1 == thing2 {
			     	continuation(b)
				return
			     }
		*/
		v1 := reflect.ValueOf(thing1)
		v2 := reflect.ValueOf(thing2)
		if lengthFunction(v1) != lengthFunction(v2) {
			return
		}

		for i := 0; i < lengthFunction(v1); i++ {
			fail := true
			goshua.Unify(indexFunction(v1, i).Interface(),
				indexFunction(v2, i).Interface(), b,
				func(b1 goshua.Bindings) {
					b = b1
					fail = false
				})
			if fail {
				return
			}
		}
		continuation(b)
	}
}

type sequenceUnifier struct{}

func (u sequenceUnifier) test(thing interface{}) bool {
	switch reflect.ValueOf(thing).Kind() {
	case reflect.Slice, reflect.Array:
		return true

	default:
		return false
	}
}

func (u sequenceUnifier) unify(thing1, thing2 interface{},
	b goshua.Bindings,
	continuation func(goshua.Bindings)) {
	collectionUnifier(reflect.Value.Len, reflect.Value.Index)(
		thing1, thing2, b, continuation)
}

type structUnifier struct{}

func (u *structUnifier) test(thing interface{}) bool {
	return reflect.ValueOf(thing).Kind() == reflect.Struct
}

func (u *structUnifier) unify(thing1, thing2 interface{},
	b goshua.Bindings,
	continuation func(goshua.Bindings)) {
	collectionUnifier(reflect.Value.NumField, reflect.Value.Index)(
		thing1, thing2, b, continuation)
}
