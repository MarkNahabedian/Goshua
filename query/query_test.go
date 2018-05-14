package query

import "reflect"
import "testing"

import "goshua/goshua"
import "goshua/unification"
import _ "goshua/variables"
import _ "goshua/bindings"
import _ "goshua/equality"

type testStruct struct {
	A int
	B string
	C *testStruct
}

var testObject = testStruct{
	A: 12,
	B: "foobar",
}


func TestUnderlyingType(t *testing.T) {
	o := testStruct{
		A: 1,
		B: "foo",
	}
	ut := underlyingType(reflect.TypeOf(&o))
	if ut != reflect.TypeOf(o) {
		t.Errorf("underlyingType failed, %T, %v", o, ut)
	}
}


// Compile time check that *variable implements goshua.Variable.
var _ goshua.Query = newQuery(reflect.TypeOf(testObject), nil, map[string]interface{}{
	"A": 12,
	"B": "foobar",
})


func TestUnifyQueryStruct(t *testing.T) {
	scope := goshua.NewScope()
	v := scope.Lookup("B")
	o := &testStruct{
		A: 16,
		B: "foo",
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
	} else if eq, err := goshua.Equal(val, o.B); err != nil {
		t.Fatalf("%s", err.Error())
	} else if !eq {
		t.Errorf("Variable bound to wrong value, got %#v, want %#v", val, o.B)
	}
}

func TestFailUnifyQueryStruct(t *testing.T) {
	o := testStruct{
		A: 16,
		B: "foo",
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

/*  depends on unifying maps, which is not yet implemented
func TestUnifyQueryQuerySucceed(t *testing.T) {
     scope := goshua.NewScope()
     v := scope.Lookup("v")
     o := testStruct {}
     foo := "foo"
     q1 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string] interface{} {
       "A": 16,
       "B": foo,
     })
     q2 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string] interface{} {
       "A": 16,
       "B": v,
     })
     tc := unification.MakeTestContinuation(t)
     goshua.Unify(q1, q2, goshua.EmptyBindings(), tc.Continuation)
     if !tc.WasContinued() {
     	t.Errorf("Failed to unify two Queries")
     } else if val, ok := tc.Bindings().Get(v); !ok {
     	t.Errorf("Variable %s should have been bound", v.Name())
	tc.Bindings().Dump()
     } else if eq, ok := goshua.Equal(val, foo); !ok || !eq {
       	t.Errorf("Variable bound to wrong value, got %#v, want %#v", val, foo)
     }
}
*/

func TestUnifyQueryQueryFail(t *testing.T) {
	o := testStruct{}
	tc := unification.MakeTestContinuation(t)
	q1 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 16,
		"B": "foo",
	})
	q2 := goshua.NewQuery(reflect.TypeOf(o), nil, map[string]interface{}{
		"A": 16,
		"B": "foo",
	})
	goshua.Unify(q1, q2, goshua.EmptyBindings(), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("queries should not have unified")
	}
}
