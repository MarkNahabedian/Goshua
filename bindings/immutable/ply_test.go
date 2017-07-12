package immutable

import "testing"
import "goshua/goshua"
import _ "goshua/variables"

func variableMap(vars ...goshua.Variable) map[goshua.Variable]bool {
	m := make(map[goshua.Variable]bool)
	for _, v := range vars {
		m[v] = true
	}
	return m
}

func TestPly(t *testing.T) {
	s := goshua.NewScope()
	v1 := s.Lookup("v1")
	v2 := s.Lookup("v2")
	v3 := s.Lookup("v3")
	v4 := s.Lookup("v4")

	p := NewPly(variableMap(v1, v2), nil, false, EmptyPly())
	if !p.Has(v1) {
		t.Errorf("Has(v1) returned false")
	}

	if p.Has(v3) {
		t.Errorf("Has(v3) returned true")
	}

	p.Dump()

	p = NewPly(variableMap(v3), 3, true, p)
	p = NewPly(variableMap(v4), 4, true, p)
	p.Dump()
}
