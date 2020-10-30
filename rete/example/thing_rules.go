package example

import "goshua/rete/rule_compiler/runtime"

//go:generate rule_compiler


type thing1 struct {
	Id string
}

func (thing *thing1) String() string {
	return "thing1-" + thing.Id
}

type thing2 struct {
	Id string
}

func (thing *thing2) String() string {
	return "thing2-" + thing.Id
}

type thing3 struct {
	t1 *thing1
	t2a *thing2
	t2b *thing2
}

func (thing *thing3) String() string {
	return "thing3" + thing.t1.Id + thing.t2a.Id + thing.t2b.Id
}


/*
func foo() *thing3 {
	return &(thing3 {
		t1: &thing1{},
		t2a: &thing2{},
		t2b: &thing2{},
	})
}
*/


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

