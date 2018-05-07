package bindings

import "fmt"
import "reflect"
import "testing"
import "goshua/goshua"
import _ "goshua/variables"
import _ "goshua/equality"
import _ "goshua/unification"

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

func TestUnify(t *testing.T) {
	s := goshua.NewScope()
	v0 := s.Lookup("v0")
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	v3 := s.Lookup("v3")
	v4 := s.Lookup("v4")

	b0 := goshua.EmptyBindings()
	b1 := goshua.EmptyBindings()
	b2 := goshua.EmptyBindings()

	bind := func(b goshua.Bindings, v goshua.Variable, val interface{}) goshua.Bindings {
		var ok bool
		if b, ok = b.Bind(v, val); !ok {
			t.Errorf("binding %s to %v failed", v.Name(), val)
		}
		return b
	}

	b0 = bind(b0, v0, 0)
	b1 = bind(b1, v1, v2)
	b1 = bind(b1, v0, v4)
	b2 = bind(b2, v3, 3)
	b2 = bind(b2, v2, v3)

	var unified goshua.Bindings = nil
	goshua.Unify(b1, b2, b0, func(b goshua.Bindings) {
		unified = b
	})

	if unified == nil {
		t.Errorf("Unifiy failed")
	}

	expectBound := func(v goshua.Variable, want interface{}) error {
		val, ok := unified.Get(v)
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

	expectBound(v0, 0)
	expectBound(v1, 3)
	expectBound(v2, 3)
	expectBound(v3, 3)
	expectBound(v4, 0)
}

func TestUnifyFail(t *testing.T) {
	s := goshua.NewScope()
	v0 := s.Lookup("v0")
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")

	b0 := goshua.EmptyBindings()
	b1 := goshua.EmptyBindings()
	b2 := goshua.EmptyBindings()

	bind := func(b goshua.Bindings, v goshua.Variable, val interface{}) goshua.Bindings {
		var ok bool
		if b, ok = b.Bind(v, val); !ok {
			t.Errorf("binding %s to %v failed", v.Name(), val)
		}
		return b
	}

	b0 = bind(b0, v0, 0)
	b1 = bind(b1, v1, v2)
	b2 = bind(b2, v2, 2)
	b2 = bind(b2, v0, v1)

	goshua.Unify(b1, b2, b0, func(b goshua.Bindings) {
		t.Errorf("Unifiy should have failed")
	})
}
