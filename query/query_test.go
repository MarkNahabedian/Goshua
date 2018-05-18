package query

import "reflect"
import "testing"

import "goshua/goshua"
import "goshua/unification"
import _ "goshua/variables"
import _ "goshua/bindings"
import _ "goshua/equality"

type testStruct struct {
	a int
	b string
	c *testStruct
}

func (ts *testStruct) A() interface{} { return ts.a }
func (ts *testStruct) B() interface{} { return ts.b }
func (ts *testStruct) C() interface{} { return ts.c }

var testObject = &testStruct{
	a: 12,
	b: "foobar",
}

// Compile-time check that the value returned by newQuery is a goshua.Query.
var _ goshua.Query = newQuery(reflect.TypeOf(testObject), nil, map[string]interface{}{
	"A": 12,
	"B": "foobar",
})

func TestUnifyQueryStruct(t *testing.T) {
	scope := goshua.NewScope()
	v := scope.Lookup("B")
	o := &testStruct{
		a: 16,
		b: "foo",
	}
	itself := scope.Lookup("itself")
	q := goshua.NewQuery(reflect.TypeOf(o), itself, map[string]interface{}{
		"A": 16,
		"B": v,
	})
	tc := unification.MakeTestContinuation(t)
	goshua.Unify(q, o, goshua.EmptyBindings(), tc.Continuation)
	if !tc.WasContinued() {
		t.Fatalf("Failed to unify Query and struct")
	}
	if val, ok := tc.Bindings().Get(itself); !ok {
		t.Errorf("%v should have been bound", itself)
		tc.Bindings().Dump()
	} else if eq, err := goshua.Equal(val, reflect.ValueOf(o).Interface()); err != nil {
		t.Fatalf("%s", err.Error())
	} else if !eq {
		t.Errorf("%v bound to wrong value, got %#v, want %#v", itself, val, o)
	}
	if val, ok := tc.Bindings().Get(v); !ok {
		t.Errorf("Variable %s should have been bound", v.Name())
		tc.Bindings().Dump()
	} else if eq, err := goshua.Equal(val, o.B()); err != nil {
		t.Fatalf("%s", err.Error())
	} else if !eq {
		t.Errorf("Variable bound to wrong value, got %#v, want %#v", val, o.b)
	}
}

func TestFailUnifyQueryStruct(t *testing.T) {
	o := &testStruct{
		a: 4,
		b: "foo",
	}
	q := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 0,
		"B": "foo",
	})
	tc := unification.MakeTestContinuation(t)
	goshua.Unify(q, o, goshua.EmptyBindings(), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("Should not have unified")
	}
}

func TestBindRecursiveStruct(t *testing.T) {
}

func TestUnifyQueryQuerySucceed(t *testing.T) {
	scope := goshua.NewScope()
	v := scope.Lookup("v")
	o := &testStruct{}
	foo := "foo"
	q1 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 8,
		"B": foo,
	})
	q2 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 8,
		"B": v,
	})
	tc := unification.MakeTestContinuation(t)
	goshua.Unify(q1, q2, goshua.EmptyBindings(), tc.Continuation)
	if !tc.WasContinued() {
		t.Errorf("Failed to unify two Queries")
	} else if val, ok := tc.Bindings().Get(v); !ok {
		t.Errorf("Variable %s should have been bound", v.Name())
	} else if eq, err := goshua.Equal(val, foo); err != nil || !eq {
		t.Errorf("Variable bound to wrong value, got %#v, want %#v.  %s", val, foo, err)
	}
}

func TestUnifyQueryQueryFail(t *testing.T) {
	scope := goshua.NewScope()
	v := scope.Lookup("v")
	o := &testStruct{}
	tc := unification.MakeTestContinuation(t)
	q1 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 2,
		"B": "foo",
	})
	q2 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 2,
		"B": v,
	})
	b := goshua.EmptyBindings()
	b, _ = b.Bind(v, "bar")
	goshua.Unify(q1, q2, b, tc.Continuation)
	if tc.WasContinued() {
		tc.Bindings().Dump()
		t.Errorf("queries should not have unified")
	}
}
