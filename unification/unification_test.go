package unification

import "testing"
import "goshua/goshua"
import _ "goshua/variables"
import _ "goshua/bindings"
import _ "goshua/equality"

// Unify equal numbers of different types
func TestUnifyNumbers(t *testing.T) {
	/* int float equality not yet implemented
	   tc := MakeTestContinuation(t)
	   goshua.Unify(1, 1.0, goshua.EmptyBindings(), tc.Continuation)
	   if !tc.WasContinued() {
	   	t.Errorf("Integer 1 and float 1 should unify.")
	   }
	*/
	/* int complex erquality not yet implemented
	   tc = MakeTestContinuation(t)
	   goshua.Unify(1, 1.0 + 0.0i, goshua.EmptyBindings(), tc.Continuation)
	   if !tc.WasContinued() {
	   	t.Errorf("Integer 1 and complex 1 should unify.")
	   }
	*/
}

// Fail to unify different numbers
func TestUnifyUnequalNumbers(t *testing.T) {
	tc := MakeTestContinuation(t)
	goshua.Unify(1, 2, goshua.EmptyBindings(), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("1 and 2 unified.")
	}
}

// unify equal strings
func TestUnifyEqualStrings(t *testing.T) {
	tc := MakeTestContinuation(t)
	goshua.Unify("foo", "foo", goshua.EmptyBindings(), tc.Continuation)
	if !tc.WasContinued() {
		t.Errorf("Equal strings should unify.")
	}
}

// fail to unify different strings
func TestUnifyUnequalStrings(t *testing.T) {
	tc := MakeTestContinuation(t)
	goshua.Unify("foo", "bar", goshua.EmptyBindings(), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("Unequal strings shopuld not unify.")
	}
}

// Bind a variable
func TestVariableBinding(t *testing.T) {
	s := goshua.NewScope()
	v := s.Lookup("v")
	tc := MakeTestContinuation(t)
	goshua.Unify(5, v, goshua.EmptyBindings(), tc.Continuation)
	if !tc.WasContinued() {
		t.Errorf("Unification with unbound variable failed.")
		return
	}
	value, bound := tc.Bindings().Get(v)
	t.Log(value, bound)
	if !bound {
		t.Errorf("Variable should have been bound.")
		return
	}
	eq, err := goshua.Equal(value, 5)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if !eq {
		t.Errorf("Variable should have been bound to 5")
	}
}

// Unify against a variable with an equal value
func TestVariableBoundToEqualValue(t *testing.T) {
	s := goshua.NewScope()
	v := s.Lookup("v")
	tc := MakeTestContinuation(t)
	bind := func(b goshua.Bindings, v goshua.Variable, value interface{}) goshua.Bindings {
		b1, ok := b.Bind(v, value)
		if !ok {
			t.Fatal("Bind failed")
		}
		return b1
	}
	goshua.Unify(5, v, bind(goshua.EmptyBindings(), v, 5), tc.Continuation)
	if !tc.WasContinued() {
		t.Errorf("Unification with variable of equal value failed.")
		return
	}
	value, bound := tc.Bindings().Get(v)
	if !bound {
		t.Errorf("Variable should still be bound")
	}
	eq, err := goshua.Equal(value, 5)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if !eq {
		t.Errorf("Variable should still be bound to 5")
	}
}

// Fail to unify against a variable with a different value
func TestVariableBoundToDifferentValue(t *testing.T) {
	s := goshua.NewScope()
	v := s.Lookup("v")
	tc := MakeTestContinuation(t)
	bind := func(b goshua.Bindings, v goshua.Variable, value interface{}) goshua.Bindings {
		b1, ok := b.Bind(v, value)
		if !ok {
			t.Fatal("Bind failed")
		}
		return b1
	}
	goshua.Unify(5, v, bind(goshua.EmptyBindings(), v, 0), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("Unification with variable of unequal value succeeded.")
	}
}

// Unify two variables with equal values
func TestTwoEqualVariables(t *testing.T) {
	s := goshua.NewScope()
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	tc := MakeTestContinuation(t)
	bind := func(b goshua.Bindings, v goshua.Variable, value interface{}) goshua.Bindings {
		b1, ok := b.Bind(v, value)
		if !ok {
			t.Fatal("Bind failed")
		}
		return b1
	}
	goshua.Unify(v1, v2, bind(bind(goshua.EmptyBindings(), v1, 4), v2, 4), tc.Continuation)
	if !tc.WasContinued() {
		t.Errorf("Unification of variables with equal values failed.")
	}
}

// Fail to unify two variables with different values
func TestTwoUnequalVariables(t *testing.T) {
	s := goshua.NewScope()
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	tc := MakeTestContinuation(t)
	bind := func(b goshua.Bindings, v goshua.Variable, value interface{}) goshua.Bindings {
		b1, ok := b.Bind(v, value)
		if !ok {
			t.Fatal("Bind failed")
		}
		return b1
	}
	goshua.Unify(v1, v2, bind(bind(goshua.EmptyBindings(), v1, 4), v2, 5), tc.Continuation)
	if tc.WasContinued() {
		t.Errorf("Unification of variables with different values succeeded.")
		tc.Bindings().Dump()
	}
}

// unify similar array and slice

// fail to unify different array and slice

// unify similar structures

// fail to unify different structures

// Confirm variable getting bound to struct.

// unify similar maps
//   func (v Value) MapIndex(key Value) Value
//   func (v Value) MapKeys() []Value
// are map keys subject to unification, or just the associated values?
