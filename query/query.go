// Package query provides a way to test some field values of a
// structure while extracting others to variables during unification.
package query

import "log"
import "fmt"
import "reflect"
import "goshua/goshua"

func underlyingType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}


// query implements the Unifier interface to test and extract fields
// of a struct.
type query struct {
	structType reflect.Type
	itself     goshua.Variable
	fields     map[string]interface{}
}

// newQuery makes a Query for unifying against a struct of type t with field.
// values as described in fieldValues.  The values in fieldValues can be
// Variables.  If itself is provided that variable will be bound to the object
// itself that the Query matched.
func newQuery(t reflect.Type, itself goshua.Variable, fieldValues map[string]interface{}) goshua.Query {
	t = underlyingType(t)
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Query only works against struct types. %v %v", t, t.Kind()))
	}
	q := query{
		structType: t,
		itself:     itself,
		fields:     fieldValues,
	}
	return &q
}

func init() {
	goshua.NewQuery = newQuery
}

func (q *query) IsQuery() bool { return true }


// Unify implements goshua.Unify for query.
// query can unify against a struct of its specified type, or with another
// query of the same specified struct type.  Keys in a query which do not
// match a filed of that struct type are ignored.
func (q *query) Unify(thing interface{}, b goshua.Bindings, continuation func(goshua.Bindings)) {
	t := underlyingType(q.structType)
	// query should also be able to unify against another query
	if thingQ, ok := thing.(query); ok {
		if t != underlyingType(thingQ.structType) {
			return
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			name := t.Field(i).Name
			i1, ok1 := q.fields[name]
			i2, ok2 := thingQ.fields[name]
			cont := false
			if ok1 && ok2 {
				goshua.Unify(i1, i2, b, func(b1 goshua.Bindings) {
					b = b1
					cont = true
				})
			} else {
				continue
			}
			if !cont {
				return
			}
		}
		continuation(b)
		return
	}
	// Unifying the Query against a struct:
	v := reflect.ValueOf(thing)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	thingType := v.Type()
	if thingType.Kind() != reflect.Struct {
		// query only unifies with struct
		return
	}
	if t != thingType {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if i1, found := q.fields[name]; !found {
			// The query is agnostic about this field
			continue
		} else {
			cont := false
			goshua.Unify(i1, v.FieldByName(name).Interface(), b,
				func(b1 goshua.Bindings) {
					b = b1
					cont = true
				})
			// Field value didn't unify, so unification fails.
			if !cont {
				return
			}
		}
	}
	if q.itself == nil {
		continuation(b)
	} else {
		if b1, ok := b.Bind(q.itself, thing); ok {
			continuation(b1)
		} else {
			log.Printf("Binding %v failed", q.itself)
		}
	}
}
