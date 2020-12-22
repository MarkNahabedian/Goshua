package example

import "reflect"
import "testing"
import "goshua/rete"


func TestRete(t *testing.T) {
	root := rete.MakeRootNode()
	var thing3_rule rete.Rule
	for _, rule := range rete.AllRules {
		rule.Installer()(root)
		if rule.Name() == "thing3" {
			thing3_rule = rule
		}
	}
	for _, typ := range thing3_rule.EmitTypes() {
		rete.GetBuffered(rete.GetTypeTestNode(root, typ))
	}
	var thing3_buffer rete.AbstractBufferNode
	rete.Walk(root, func(n rete.Node) {
		if abn, ok := n.(rete.AbstractBufferNode); ok {
			if abn.Inputs()[0].(*rete.TypeTestNode).Type == reflect.TypeOf(func(*thing3){}).In(0) {
				thing3_buffer = abn
			}
		}
	})
	// Validate
	rete.Walk(root, func(n rete.Node) {
		for _, err := range n.Validate() {
			t.Errorf("%s", err)
		}
	})
	// Graph
	if graph, err := rete.MakeGraph(root); err != nil {
		t.Error(err)
	} else {
		rete.WriteGraphvizFile(graph, "example_rete.dot")
	}
	// Dump
	dump := func() {
		rete.Walk(root, func(n rete.Node) {
			if buf, ok := n.(rete.AbstractBufferNode); ok {
				t.Logf("Node %s, %d items:\n", buf.Label(), buf.Count())
				buf.DoItems(func (x interface{}) {
					t.Logf("    %s\n", x)
				})
			}
		})
	}
	// Assert stuff:
	assert := func(item interface{}) {
		t.Logf("\nasserting %s\n", item)
		root.Receive(item)
	}
	assert(&thing1{ Id: "a" })
	assert(&thing2{ Id: "b" })
	if want, got := 0, thing3_buffer.Count(); want != got {
		// rule_thing3 excludes cases where the two thing2 elements are identical.
		dump()
		t.Errorf("Wrong count for thing3 after first thing2 asserted: watd %d, got %d", want, got)
	}
	assert(&thing2{ Id: "c" })
	if want, got := 2, thing3_buffer.Count(); want != got {
		dump()
		t.Errorf("Wrong count for thing3 after second thing2 asserted: want %d, got %d", want, got)		
	}
	assert(&thing1{ Id: "d" })
	if want, got := 4, thing3_buffer.Count(); want != got {
		dump()
		t.Errorf("Wrong count for thing3 after second thing1 asserted: want %d, got %d", want, got)		
	}
}
