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
	for _, o := range n.Outputs() {
		o.Receive(item)
	}
}

// Receive  is part of the node interface.
func (n *BasicNode) Receive(interface{}) {
	panic(fmt.Sprintf("BasicNode.Receive on %T", n))
}

// IsValid is part of the node interface.
func (n *BasicNode) IsValid() bool {
	// Dummy method
	panic("BasicNode is abstract.  It should not have been instantiated.")
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

// BufferNode collects items into a buffer.
// BufferNode provides cursors for iterating over the collected
// items.  Only BufferNodes can be the inputs of a JoinNode.
type BufferNode struct {
	// node
	BasicNode
	items     []interface{}
}

// IsValid is part of the Node interface.
func (n *BufferNode) IsValid() bool {
	return len(n.Inputs()) == 1
}

func (n *BufferNode) Count() int {
	return len(n.items)
}

// Receive is part of the Node interface.
// When a BufferNode receives an item each of its cursors
// calls its newItemFunction so that the JoinNode that
// created that cursor can attempt to join that item with
// each item in the other branch of the JoinNode's BufferNode.
func (n *BufferNode) Receive(item interface{}) {
	n.items = append(n.items, item)
	n.Emit(item)
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
// Emiting the cross-product as successive [2]interface{} arrays..
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
	if _, ok := n.Inputs()[0].(*JoinSide); !ok {
		return false
	}
	if _, ok := n.Inputs()[1].(*JoinSide); !ok {
		return false
	}
	return true
}


type JoinSide struct {
	joinNode *JoinNode
	input *BufferNode
	other *JoinSide
	swap bool
}

func (n *JoinSide) Label() string {
	var side string
	if n.swap {
		side = "B"
	} else {
		side = "A"
	}
	return fmt.Sprintf("%s - %s", n.joinNode.Label(), side)
}

func (n *JoinSide) Inputs() []Node {
	return []Node{ n.input }
}

func (n *JoinSide) Outputs() []Node {
	return []Node{ n.joinNode }
}

func (n *JoinSide) addInput(n2 Node) {
	panic("JoinSide.addInput called")
}

func (n *JoinSide) addOutput(n2 Node) {
	panic("JoinSide.addOutput called")
}

func (n *JoinSide) Emit(item interface{}) {
	n.joinNode.Emit(item)
}

func (n *JoinSide) IsValid() bool {
	return true
}


func (n *JoinSide) Receive(item1 interface{}) {
	c := n.other.input.GetCursor()
	for item2, present := c.Next(); present; item2, present = c.Next() {
		if n.swap {
			n.Emit([2]interface{}{item2, item1})
		} else {
			n.Emit([2]interface{}{item1, item2})
		}
	}
}


func getBuffered(n Node) *BufferNode {
	if b, ok := n.(*BufferNode); ok {
		return b
	}
	for _, o := range n.Outputs() {
		if b, ok := o.(*BufferNode); ok {
			return b
		}
	}
	bn := &BufferNode{}
	bn.label = fmt.Sprintf("%s - buffered", n.Label())
	Connect(n, bn)
	return bn
}

func Join(label string, a, b Node) *JoinNode {
	jn := &JoinNode{}
	jn.label = label
	aSide := &JoinSide{
		joinNode: jn,
		input: getBuffered(a),
		swap: false,
	}
	aSide.input.addOutput(aSide)
	bSide := &JoinSide{
		joinNode: jn,
		input: getBuffered(b),
		swap: true,
	}
	bSide.input.addOutput(bSide)
	aSide.other = bSide
	bSide.other = aSide
	jn.addInput(aSide)
	jn.addInput(bSide)
	return jn	
}


func Walk(root Node, f func(n Node)) {
	visited := make(map[Node]bool)
	var visit func(Node)
	visit = func(n Node) {
		if visited[n] { return }
		f(n)
		visited[n] = true
		for _, o := range n.Outputs() {
			visit(o)
		}
	}
	visit(root)
}
