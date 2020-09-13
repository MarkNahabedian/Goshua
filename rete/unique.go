package rete

import "fmt"


// UniqueBufferNode remembers those items it receives that are unique
// to one another based on equivalence_function.
type UniqueBufferNode struct {
	BasicNode
	AbstractBufferNode
	items []interface{}
	// equivalence_function implements the uniqueness test for
	// this node.
	equivalence_function func(interface{}, interface{}) bool
}

// IsValid is part of the Node interface.
func (n *UniqueBufferNode) IsValid() bool {
	return true
}

func (n *UniqueBufferNode) Count() int {
	return len(n.items)
}

func (n *UniqueBufferNode) DoItems(f func(interface{})) {
	for _, item := range n.items {
		f(item)
	}
}

func (n *UniqueBufferNode) Clear() {
	n.items = nil
}

func (n *UniqueBufferNode) Receive(item interface{}) {
	for _, i := range n.items {
		if n.equivalence_function(i, item) {
			return
		}
	}
	n.items = append(n.items, item)
	n.Emit(item)
}

// GetUniqueBuffered finds or creates a UniqueBufferNode which buffers
// the output of n.
func GetUniqueBuffered(n Node, equivalence_function func(interface{}, interface{}) bool) *UniqueBufferNode {
	if b, ok := n.(*UniqueBufferNode); ok {
		return b
	}
	for _, o := range n.Outputs() {
		if b, ok := o.(*UniqueBufferNode); ok {
			return b
		}
	}
	bn := &UniqueBufferNode {
		equivalence_function: equivalence_function}
	bn.label = fmt.Sprintf("%s - unique", n.Label())
	Connect(n, bn)
	return bn
}
