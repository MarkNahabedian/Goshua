package example

import "testing"
import "goshua/rete"
import "goshua/rete/rule_compiler/runtime"


func TestRete(t *testing.T) {
	t.Log("FOO")
	root := rete.MakeRootNode()
	var thing3_rule runtime.Rule
	for _, rule := range runtime.AllRules {
		rule.Installer()(root)
		if rule.Name() == "thing3" {
			thing3_rule = rule
		}
	}
	for _, typ := range thing3_rule.EmitTypes() {
		rete.GetBuffered(rete.GetTypeTestNode(root, typ))
	}
	// Assert stuff:
	root.Receive(&thing1{})
	root.Receive(&thing2{})
	root.Receive(&thing2{})	
	// Dump:
	rete.Walk(root, func(n runtime.Node) {
		if buf, ok := n.(rete.AbstractBufferNode); ok {
			t.Log(buf.Label())
			buf.DoItems(func (x interface{}) {
				t.Log("    ", x)
			})
			
		}
	})
}
