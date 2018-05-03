// Package rete implements the Rete algorthm.
package rete

import "fmt"
import "reflect"

type node interface {
	// Label returns the node's label.
	Label() string

	// Inputs returns thenodes that can send data to this node.
	Inputs() []node

	// Outputs returns the nodes that this node can output to.
	Outputs() []node

	addInput(node)
	addOutput(node)

	// Emit outputs item to this node's Outputs.  It does so by calling
	// Receive on each Output.
	Emit(item interface{})

	// Receive causes the node to process an input item.
	Receive(item interface{})

	// IsValid check the node to make sure it's valid.
	IsValid() bool

	// InitializeNode should be called once the entire rete is constructed but
	// before any data is entered.
	InitializeNode()

	// Invoke a function when a node receives an item.  Only some nodes
	// support this.
	AddListener(func(interface{}))
}

// Initialize should be called on the root node of a rete after the rete is
// constructed but before it is used to make sure every node is ready to run.
func Initialize(n node) {
	initialized := make(map[node]bool)
	var walker func(node)
	walker = func(n node) {
		if initialized[n] {
			return
		}
		n.InitializeNode()
		initialized[n] = true
		for _, o := range n.Outputs() {
			walker(o)
		}
	}
	walker(n)
}

// Connect arranges for from to output to to.
func Connect(from node, to node) {
	from.addOutput(to)
	to.addInput(from)
}

// basicNode provides a common implementation of the node interface's
// Inputs, Outputs, and Emit methods.
// basicNode is abstract.  It should not be instantiated.
type basicNode struct {
	node
	label   string
	inputs  []node
	outputs []node
}

// Label is part of the node interface.
func (n *basicNode) Label() string {
	if n.label == "" {
		return fmt.Sprintf("%s-%x",
			reflect.TypeOf(n).Name(),
			reflect.ValueOf(n).Pointer())
	}
	return n.label
}

// Inputs is part of the node interface.
func (n *basicNode) Inputs() []node {
	return n.inputs
}

// Outputs is part of the node interface.
func (n *basicNode) Outputs() []node {
	return n.outputs
}

func (n1 *basicNode) addInput(n2 node) {
	n1.inputs = append(n1.inputs, n2)
}

func (n1 *basicNode) addOutput(n2 node) {
	n1.outputs = append(n1.outputs, n2)
}

// Emit is part of the node interface.
func (n *basicNode) Emit(item interface{}) {
	for _, o := range n.outputs {
		o.Receive(item)
	}
}

// Receive  is part of the node interface.
func (n *basicNode) Receive(interface{}) {
	panic("basicNode.Receive")
}

// InitializeNode is part of the node interface.
func (n *basicNode) InitializeNode() {
	// Defualt implementation is to do nothing.
}

// IsValid is part of the node interface.
func (n *basicNode) IsValid() bool {
	// Dummy method
	panic("basicNode is abstract.  It should not have been instantiated.")
}

func (n *basicNode) AddListener(func(interface{})) {
	panic("basicNode doesn't support AddListener")
}

// ActionNode is a node that can perform some action on its input item,
// like construct and assert a fact.
type ActionNode struct {
	// node
	basicNode
	actionFunction func(item interface{})
}

// Receive is part of the Node interface.
func (n *ActionNode) Receive(item interface{}) {
	n.actionFunction(item)
	// Pass item through to any outputs.
	n.Emit(item)
}

// IsValid is part of the node interface.
func (n *ActionNode) IsValid() bool {
	return len(n.Inputs()) == 1
}

// TestNode implements a rete node with a single input.  items Received
// by a TestNode are only Emited if they satisfy a test function.
type TestNode struct {
	// node
	basicNode
	testFunction func(interface{}) bool
}

// Receive is part of the node interface.
func (n *TestNode) Receive(item interface{}) {
	if n.testFunction(item) {
		n.Emit(item)
	}
}

// IsValid is part of the Node interface.
func (n *TestNode) IsValid() bool {
	return len(n.Inputs()) == 1
}

// BufferNode collects items into a buffer.  Listener functions can
// be registered to be called on each item as it is received.
// BufferNode also provides cursors for iterating over the collected
// items.  Only BufferNodes can be the inputs of a JoinNode.
type BufferNode struct {
	// node
	basicNode
	items     []interface{}
	listeners []func(interface{})
}

// IsValid is part of the Node interface.
func (n *BufferNode) IsValid() bool {
	return len(n.Inputs()) == 1
}

/*
func (n *BufferNode) Emit(item interface{}) {
	panic("Emit called on a BufferNode")
}
*/

// AddListener registers f as a function to be called on an item
// when it is Received by the BufferNode.
func (n *BufferNode) AddListener(f func(interface{})) {
	n.listeners = append(n.listeners, f)
}

// Receive is part of the Node interface.
// When a BufferNode receives an item each of its cursors
// calls its newItemFunction so that the JoinNodfe that
// created that cursor can attempt to join that item with
// each item in the other branch of the JoinNode's BufferNode.
func (n *BufferNode) Receive(item interface{}) {
	n.items = append(n.items, item)
	for _, l := range n.listeners {
		l(item)
	}
}

type cursor struct {
	done   bool
	buffer *BufferNode
	index  int
}

// GetCursor returns a new cursor into n.
func (n *BufferNode) GetCursor() *cursor {
	var c cursor
	c.buffer = n
	c.done = false
	c.index = 0
	return &c
}

// Next returns the item that the cursor is currently referring to and
// advances the cursor.  Next returns nil, false if there are no more items.
func (c *cursor) Next() (interface{}, bool) {
	if c.index >= len(c.buffer.items) {
		return nil, false
	}
	i := c.buffer.items[c.index]
	c.index += 1
	return i, true
}

// JoinNode combines the items in its two input BufferNodes pairwise,
// Emiting the cross-product.
type JoinNode struct {
	// node
	basicNode
}

// IsValid is part of the Node interface.
func (n *JoinNode) IsValid() bool {
	if len(n.Inputs()) != 2 {
		return false
	}
	// The inputs of a JoinNode must be BufferNodes.
	if _, ok := n.Inputs()[0].(*BufferNode); !ok {
		return false
	}
	if _, ok := n.Inputs()[1].(*BufferNode); !ok {
		return false
	}
	return true
}

// InitializeNode is part of the node interface
func (n *JoinNode) InitializeNode() {
	listener := func(inputIndex int) func(interface{}) {
		return func(item1 interface{}) {
			otherInput := n.Inputs()[(inputIndex+1)%2].(*BufferNode)
			c := otherInput.GetCursor()
			for item2, present := c.Next(); present; item2, present = c.Next() {
				if inputIndex == 0 {
					n.Emit([]interface{}{item1, item2})
				} else {
					n.Emit([]interface{}{item2, item1})
				}
			}
		}
	}
	n.Inputs()[0].AddListener(listener(0))
	n.Inputs()[1].AddListener(listener(1))
}
