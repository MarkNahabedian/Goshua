// Package rete implements the Rete algorthm.
package rete

type node interface {
	// Label returns the node's label.
	Label() string

	// Inputs returns thenodes that can send data to this node.
	Inputs() []node

	// Outputs returns the nodes that this node can output to.
	Outputs() []node

	// OutputsTo connects the receiving node with n such that n is an
	// Output of the receiver and the receiver is an Input of n.
	OutputsTo(n node)

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
}

// Initialize should be called on the root node of a rete after the rete is
// constructed but before it is used to make sure every node is ready to run.
func Initialize(n node) {
	initialized := make(map [node]bool)
	if initialized[n] { return }
	n.InitializeNode()
	initialized[n] = true
	for _, o := range n.Outputs() {
		Initialize(o)
	}
}


// basicNode provides a common implementation of the node interface's
// Inputs, Outputs, Emit and OutputsTo methods.
// basicNode is abstract.  It should not be instantiated.
type basicNode struct {
	node
	label string
	inputs []node
	outputs []node
}

// Label is part of the node interface.
func (n *basicNode) Label() string {
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

// OutputsTo connects n1 and n2 such that n1 is an Input of n2 and
// n2 is an Output of n1.
func (n1 *basicNode) OutputsTo(n2 node) {
	n1.addOutput(n2)
	n2.addInput(n1)
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


// BufferNode collects items into a buffer and provides cursors for
// iterating over the collected items.  Only BufferNodes can be the
// inputs of a Join node.
type BufferNode struct {
	// node
	basicNode
	items []interface{}
	cursors []*cursor
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

// Receive is part of the Node interface.
// When a BufferNode receives an item each of its cursors
// calls its newItemFunction so that the JoinNodfe that
// created that cursor can attempt to join that item with
// each item in the other branch of the JoinNode's BufferNode.
func (n *BufferNode) Receive(item interface{}) {
	n.items = append(n.items, item)
	for _, c := range n.cursors {
		c.newItemFunction(item)
	}
}

type cursor struct {
	node
	done bool
	buffer *BufferNode
	index int
	newItemFunction func(interface{})
}

// GetCursor returns a new cursor into n.
func (n *BufferNode) GetCursor(newItemFunction func(interface{})) *cursor {
	var c cursor
	c.buffer = n
	c.done = false
	c.index = 0
	c.newItemFunction = newItemFunction
	n.cursors = append(n.cursors, &c)
	return &c
}

// Done should be called on a cursor when it is no longer to be used.
func (c *cursor) Done() {
	c.done = true
	for i, c1 := range c.buffer.cursors {
		if c == c1 {
		   c.buffer.cursors = append(c.buffer.cursors[:i],
		                             c.buffer.cursors[i+1:]...)
		   break
		}
	}
}

// Next returns the item that the cursor is currently referring to and
//  advances the cursor.  Next returns nil if there are no more items.
func (c *cursor) Next() interface{} {
	if c.index >= len(c.buffer.items) {
	   return nil
	}
	i := c.buffer.items[c.index]
	c.index += 1
	return i
}


type JoinNode struct {
	// node
	basicNode
	cursors []*cursor
}

// IsValid is part of the Node interface.
func (n *JoinNode) IsValid() bool {
	if  len(n.Inputs()) != 2 {
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
	// Set up cursors
}

