// Package query provides a way to test some field values of a
// structure while extracting others to variables during unification.
package query

import "log"
import "fmt"
import "reflect"
import "goshua/goshua"

// query implements the Unifier interface to test and extract fields
// of a struct.
type query struct {
	structType reflect.Type
	itself     goshua.Variable
	// Map from a reader method name to an object to Unify the read value against.
	matchers map[string]interface{}
}

func matchersKey(method reflect.Method) string {
	// We don't need the package because name is confined to the context
	// of structType.
	return method.Name
}

// newQuery makes a Query for unifying against an object of a specified type.
// readerValues is a map from the names pf reader methods on that obkect (as
// strings) to values or variables to be unified against.
// If itself is provided that variable will be bound to the object itself
// that the Query matched.
func newQuery(t reflect.Type, itself goshua.Variable, readerValues map[string]interface{}) goshua.Query {
	q := query{
		structType: t,
		itself:     itself,
		matchers:   make(map[string]interface{}),
	}
	for name, val := range readerValues {
		if method, ok := t.MethodByName(name); ok {
			q.matchers[matchersKey(method)] = val
		} else {
			/* Show available methods
			for i := 0; i < t.NumMethod(); i++ {
				m := t.Method(i)
				log.Printf("%d: method %s.%s", i, m.PkgPath, m.Name)
			}
			*/
			panic(fmt.Sprintf("No method %s for type %v", name, t))
		}
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
	t := q.structType
	// query should also be able to unify against another query
	if thingQ, ok := thing.(*query); ok {
		if t != thingQ.structType {
			// log.Printf("query types don't match %v %v", t, thingQ.structType)
			return
		}
		for i := 0; i < t.NumMethod(); i++ {
			method := t.Method(i)
			i1, ok1 := q.matchers[matchersKey(method)]
			i2, ok2 := thingQ.matchers[matchersKey(method)]
			cont := false
			if ok1 && ok2 {
				goshua.Unify(i1, i2, b, func(b1 goshua.Bindings) {
					b = b1
					cont = true
				})
			} else if ok1 || ok2 {
				log.Printf("%s matcher missing", method.Name)
				return
			} else {
				// Neither query cares about this value.
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
	thingType := v.Type()
	if t != thingType {
		// log.Printf("Types don't match: %v, %v", t, thingType)
		return
	}
	for name, val1 := range q.matchers {
		method, _ := t.MethodByName(name)
		val2 := method.Func.Call([]reflect.Value{v})[0].Interface()
		cont := false
		goshua.Unify(val1, val2, b,
			func(b1 goshua.Bindings) {
				b = b1
				cont = true
			})
		// Field value didn't unify, so unification fails.
		if !cont {
			// log.Printf("no match %s %v %v", method.Name, val1, val2)
			return
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
