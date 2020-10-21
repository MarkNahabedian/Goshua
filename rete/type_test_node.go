package rete

import "fmt"
import "reflect"


// TypeTestNode filters its incoming items by the specified type.
type TypeTestNode struct {
	// Node
	BasicNode
	Type reflect.Type
}

func MakeTypeTestNode(t reflect.Type) *TypeTestNode {
	n := &TypeTestNode{Type: t}
	n.label = fmt.Sprintf("type test %s", t)
	return n
}

func (n *TypeTestNode) TypeName() string { return n.Type.String() }

// Receive is part of the node interface.
func (n *TypeTestNode) Receive(item interface{}) {
	if reflect.TypeOf(item).ConvertibleTo(n.Type) {
		n.Emit(item)
	}
}

// IsValid is part of the Node interface.
func (n *TypeTestNode) IsValid() bool {
	return true
}

// GetTypeTestNode finds or creates a Node that filters by the specified type t.
// n should be the root node of a rete.
func GetTypeTestNode(n Node, typ reflect.Type) *TypeTestNode {
	for _, output := range n.Outputs() {
		if output, ok := output.(*TypeTestNode); ok {
			if output.Type == typ {
				return output
			}
		}
	}
	o := MakeTypeTestNode(typ)
	Connect(n, o)
	return o
}
