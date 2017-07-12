package bindings

import "fmt"
import "reflect"
import "testing"
import "goshua/goshua"
import _ "goshua/variables"
import _ "goshua/equality"

func TestSimpleBinding(t *testing.T) {
	s := goshua.NewScope()
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	v3 := s.Lookup("v3")
	v4 := s.Lookup("v4")

	b := goshua.EmptyBindings()

	bind := func(v goshua.Variable, val interface{}) {
		var ok bool
		if b, ok = b.Bind(v, val); !ok {
			t.Errorf("binding %s to %v failed", v.Name(), val)
		}
		b.Dump()
	}

	expectBound := func(v goshua.Variable, want interface{}) error {
		val, ok := b.Get(v)
		if !ok {
			return fmt.Errorf("%s isn't bound", v)
		}
		eq, err := goshua.Equal(val, want)
		if err != nil {
			return err
		}
		if !eq {
			return fmt.Errorf("%s should have value %#v, not %#v", v, want, val)
		}
		return nil
	}

	bind(v1, 1)
	bind(v2, "two")
	bind(v3, 3)
	bind(v4, "IV")

	if err := expectBound(v1, 1); err != nil {
		t.Errorf("%s", err)
	}
	if err := expectBound(v2, "two"); err != nil {
		t.Errorf("%s", err)
	}
	if err := expectBound(v3, 3); err != nil {
		t.Errorf("%s", err)
	}
	if err := expectBound(v4, "IV"); err != nil {
		t.Errorf("%s", err)
	}
}

func TestLogicVariables(t *testing.T) {
	s := goshua.NewScope()
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	v3 := s.Lookup("v3")
	v4 := s.Lookup("v4")

	b := goshua.EmptyBindings()

	bind := func(v goshua.Variable, val interface{}) {
		var ok bool
		if b, ok = b.Bind(v, val); !ok {
			t.Errorf("binding %s to %v failed", v.Name(), val)
		}
		b.Dump()
	}

	expectBound := func(v goshua.Variable, want interface{}) error {
		val, ok := b.Get(v)
		if !ok {
			return fmt.Errorf("%s isn't bound", v)
		}
		eq, err := goshua.Equal(val, want)
		if err != nil {
			return err
		}
		if !eq {
			return fmt.Errorf("%s should have value %#v, not %#v", v, want, val)
		}
		return nil
	}

	// v1 shouldn't have a value yet
	if val, ok := b.Get(v1); ok {
		t.Errorf("v1 shouldn't have a value in empty Bindings: %v", val)
	}

	// make v1 and v2 equal
	bind(v1, v2)
	bind(v3, v2)

	if val, ok := b.Get(v1); ok {
		t.Errorf("v1 was bound to another variable.  It shouldn't have a value: %v", val)
	}

	// Bind v4 to 4
	bind(v4, 4)

	if val, ok := b.Get(v4); !(ok && reflect.ValueOf(val).Int() == 4) {
		t.Errorf("v4 should have value 4, %v %v", val, ok)
	}

	// Bind v1 to "foo"
	want := "foo"
	bind(v1, want)

	if err := expectBound(v1, want); err != nil {
		t.Errorf("%s", err)
	}
	// Was v2 set as well?
	if err := expectBound(v2, want); err != nil {
		t.Errorf("%s", err)
	}
	// How about v3?
	if err := expectBound(v3, want); err != nil {
		t.Errorf("%s", err)
	}
}
