package example

import "goshua/rete/rule_compiler/runtime"

//go:generate rule_compiler


type thing1 struct {
}

type thing2 struct {
}

type thing3 struct {
	t1 *thing1
	t2a *thing2
	t2b *thing2
}

func foo() *thing3 {
	return &(thing3 {
		t1: &thing1{},
		t2a: &thing2{},
		t2b: &thing2{},
	})
}


func rule_thing3(node runtime.Node, t1 *thing1, t2a, t2b *thing2) {
	if t2a == t2b {
		return
	}
	node.Emit(&thing3 {
		t1: t1,
		t2a: t2a,
		t2b: t2b,
	})
}

