// Package rete implements the Rete algorthm.
package rete

import "fmt"
import "reflect"

type Node interface {
	// Label returns the node's label.
	Label() string

	// Inputs returns thenodes that can send data to this node.
	Inputs() []Node

	// Outputs returns the nodes that this node can output to.
	Outputs() []Node

	addInput(Node)
	addOutput(Node)

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
func Initialize(n Node) {
	initialized := make(map[Node]bool)
	var walker func(Node)
	walker = func(n Node) {
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
func Connect(from Node, to Node) {
	from.addOutput(to)
	to.addInput(from)
}

// BasicNode provides a common implementation of the node interface's
// Inputs, Outputs, and Emit methods.
// BasicNode is abstract.  It should not be instantiated.
type BasicNode struct {
	Node
	label   string
	inputs  []Node
	outputs []Node
}

// Label is part of the node interface.
func (n *BasicNode) Label() string {
	if n.label == "" {
		return fmt.Sprintf("%s-%x",
			reflect.TypeOf(n).Name(),
			reflect.ValueOf(n).Pointer())
	}
	return n.label
}

// Inputs is part of the node interface.
func (n *BasicNode) Inputs() []Node {
	return n.inputs
}

// Outputs is part of the node interface.
func (n *BasicNode) Outputs() []Node {
	return n.outputs
}

func (n1 *BasicNode) addInput(n2 Node) {
	n1.inputs = append(n1.inputs, n2)
}

func (n1 *BasicNode) addOutput(n2 Node) {
	n1.outputs = append(n1.outputs, n2)
}

// Emit is part of the node interface.
func (n *BasicNode) Emit(item interface{}) {
	for _, o := range n.outputs {
		o.Receive(item)
	}
}

// Receive  is part of the node interface.
func (n *BasicNode) Receive(interface{}) {
	panic("BasicNode.Receive")
}

// InitializeNode is part of the node interface.
func (n *BasicNode) InitializeNode() {
	// Defualt implementation is to do nothing.
}

// IsValid is part of the node interface.
func (n *BasicNode) IsValid() bool {
	// Dummy method
	panic("BasicNode is abstract.  It should not have been instantiated.")
}

func (n *BasicNode) AddListener(func(interface{})) {
	panic("BasicNode doesn't support AddListener")
}

// ActionNode is a node that can perform some action on its input item,
// like construct and assert a fact.
type ActionNode struct {
	// node
	BasicNode
	actionFunction func(item interface{})
}

func MakeActionNode(actionFunction func(item interface{})) *ActionNode {
	return &ActionNode{
		actionFunction: actionFunction,
	}
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
	BasicNode
	testFunction func(interface{}) bool
}

func MakeTestNode(testFunction func(interface{}) bool) *TestNode {
	return &TestNode{testFunction: testFunction}
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

// FunctionNode calls function on the incoming item.  It can
// conditionally Emit that item or something else.
type FunctionNode struct {
	// node
	BasicNode
	function func(Node, interface{})
}

func MakeFunctionNode(label string, function func(Node, interface{})) *FunctionNode {
	n := &FunctionNode{
		function: function,
	}
	n.label = label
	return n
}

// Receive is part of the node interface.
func (n *FunctionNode) Receive(item interface{}) {
	n.function(n, item)
}

// IsValid is part of the Node interface.
func (n *FunctionNode) IsValid() bool {
	return len(n.Inputs()) == 1
}

// BufferNode collects items into a buffer.  Listener functions can
// be registered to be called on each item as it is received.
// BufferNode also provides cursors for iterating over the collected
// items.  Only BufferNodes can be the inputs of a JoinNode.
type BufferNode struct {
	// node
	BasicNode
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
	BasicNode
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
