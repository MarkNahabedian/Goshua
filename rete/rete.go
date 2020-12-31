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

	AddInput(Node)
	AddOutput(Node)

	// Emit outputs item to this node's Outputs.  It does so by calling
	// Receive on each Output.
	Emit(item interface{})

	// Receive causes the node to process an input item.
	Receive(item interface{})

	// Validate returns an error if the Node doesn't pass
	// validity checks.
	Validate() []error

	// Clear causes a Node to forget any stored items.
	Clear()
}

// Connect arranges for from to output to to.
func Connect(from Node, to Node) {
	from.AddOutput(to)
	to.AddInput(from)
}

func ValidateConnectivity(n Node) []error {
	errors := []error{}
	find := func (node Node, nodes []Node) bool {
		for _, candidate := range nodes {
			if candidate == node {
				return true
			}
		}
		return false
	}
	for _, input := range n.Inputs() {
		if !find(n, input.Outputs()) {
			errors = append(errors,
				fmt.Errorf("Node %s is not an output of input %s",
					n.Label(), input.Label()))
		}
	}
	for _, output := range n.Outputs() {
		if !find(n, output.Inputs()) {
			errors = append(errors,
				fmt.Errorf("Node %s is not an input of output %s",
					n.Label(), output.Label()))
		}
	}
	return errors
}


// BasicNode provides a common implementation of the node interface's
// Inputs, Outputs, and Emit methods.
// BasicNode is abstract.  It should not be instantiated.
type BasicNode struct {
	label   string
	inputs  []Node
	outputs []Node
}

// *BasicNode must implement the Node interface:
// var _ Node = (*BasicNode)(nil)

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

func (n1 *BasicNode) AddInput(n2 Node) {
	n1.inputs = append(n1.inputs, n2)
}

func (n1 *BasicNode) AddOutput(n2 Node) {
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

func (n *BasicNode) Clear() {
	// Default method.
}


// MakeRootNode creates a note to serve as the root node of a rete.
func MakeRootNode() Node {
	return MakeFunctionNode("root", func(n Node, item interface{}) {
		n.Emit(item)
	})
}


// ActionNode is a node that can perform some action on its input item,
// like construct and assert a fact.
type ActionNode struct {
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

// Validate is part is part of the Node interface.
func (n *ActionNode) Validate() []error {
	return ValidateConnectivity(n)
}


// TestNode implements a rete node with a single input.  items Received
// by a TestNode are only Emited if they satisfy a test function.
type TestNode struct {
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

// Validate is part of the Node interface.
func (n *TestNode) Validate() []error {
	return ValidateConnectivity(n)
}


// TypeFilterNode only passes items that satisfy the specified type.
type TypeFilterNode struct {
	// Node
	BasicNode
	testType reflect.Type
}

func MakeTypeFilterNode(t reflect.Type) *TypeFilterNode {
	n := &TypeFilterNode{testType: t}
    n.label = fmt.Sprintf("%v", t)
	return n
}

// Receive is part of the node interface.
func (n *TypeFilterNode) Receive(item interface{}) {
	if reflect.TypeOf(item) == n.testType {
		n.Emit(item)
	}
}

// Validate is part of the Node interface.
func (n *TypeFilterNode) Validate() []error {
	return ValidateConnectivity(n)
}

// GetTypeFilterNode find or create a Node that filters by the specified type t.
// n should be the root node of a rete.
func GetTypeFilterNode(n Node, t reflect.Type) *TypeFilterNode {
	for _, output := range n.Outputs() {
		if output, ok := output.(*TypeFilterNode); ok {
			if output.testType == t {
				return output
			}
		}
	}
	o := MakeTypeFilterNode(t)
	Connect(n, o)
	return o
}


// FunctionNode calls function on the incoming item.  It can
// conditionally Emit that item or something else.
type FunctionNode struct {
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

// Validate is part of the Node interface.
func (n *FunctionNode) Validate() []error {
	return ValidateConnectivity(n)
}


type AbstractBufferNode interface {
	Node
	Clear()
	Count() int
	DoItems(func(interface{}))
}


type TypedBufferNode interface {
	AbstractBufferNode
	Type() reflect.Type
}


// BufferNode collects items into a buffer.
type BufferNode struct {
	BasicNode
	items []interface{}
}

// *BufferNode must implement the AbstractBufferNode interface:
var _ AbstractBufferNode = (*BufferNode)(nil)

// Validate is part of the Node interface.
func (n *BufferNode) Validate() []error {
	return ValidateConnectivity(n)
}

func (n *BufferNode) Count() int {
	return len(n.items)
}

// Receive is part of the Node interface.
func (n *BufferNode) Receive(item interface{}) {
	n.items = append(n.items, item)
	n.Emit(item)
}

func (n *BufferNode) Clear() {
	n.items = nil
}

func (n *BufferNode) DoItems(f func(interface{})) {
	for _, item := range n.items {
		f(item)
	}
}


// GetBuffered finds or creates a BufferNode which buffers the output of n.
func GetBuffered(n Node) *BufferNode {
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


func Walk(root Node, f func(n Node)) {
	visited := make(map[Node]bool)
	var visit func(Node)
	visit = func(n Node) {
		if visited[n] {
			return
		}
		f(n)
		visited[n] = true
		for _, o := range n.Outputs() {
			visit(o)
		}
	}
	visit(root)
}
