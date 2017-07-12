package variables

import "testing"
import "goshua/goshua"

func TestInterfaces(t *testing.T) {
	s := goshua.NewScope()
	if _, ok := s.(goshua.Scope); !ok {
		t.Errorf("not a goshua.Scope")
	}
	if _, ok := s.Lookup("a").(goshua.Variable); !ok {
		t.Errorf("not a goshua.Variable")
	}
}

func TestScopeIdentity(t *testing.T) {
	scope1 := goshua.NewScope()
	scope2 := goshua.NewScope()
	if scope1 == scope2 {
		t.Errorf("Scopes are not unique")
	}
}

func TestVariableIdentity(t *testing.T) {
	scope := goshua.NewScope()

	a := scope.Lookup("a")
	if a.Name() != "a" {
		t.Errorf("Variable has wrong name %s", a.Name())
	}
	if !a.SameAs(scope.Lookup("a")) {
		t.Errorf("Two variables for a")
	}

	b := scope.Lookup("b")
	if b.SameAs(a) {
		t.Errorf("a and b are the same")
	}
}

func TestVariableDistinctness(t *testing.T) {
	s1 := goshua.NewScope()
	s2 := goshua.NewScope()
	v1 := s1.Lookup("a")
	v2 := s2.Lookup("a")
	if v1 == v2 {
		t.Errorf("Variables with same name but in different scopes should be different")
	}
}
