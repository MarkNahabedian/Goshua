package rete

import "fmt"
import "sort"
import "testing"

func TestActionNode(t *testing.T) {
	s := ""
	n := MakeActionNode(
		func(item interface{}) {
			s = fmt.Sprintf("%v", item)
		})
	n.Receive(5)
	if s != "5" {
		t.Errorf("Action not performed.")
	}
}

func TestTestNode(t *testing.T) {
	s := ""
	n1 := MakeTestNode(func(item interface{}) bool {
		return item.(int) == 4
	})
	n2 := MakeActionNode(
		func(item interface{}) {
			s = fmt.Sprintf("%v", item)
		})
	Connect(n1, n2)
	n1.Receive(5)
	if s != "" {
		t.Errorf("Action should not have been executed: %s", s)
	}
	s = ""
	n1.Receive(4)
	if s != "4" {
		t.Errorf("Action should have been executed")
	}
}

func TestBuffer(t *testing.T) {
	n1 := MakeTestNode(func(item interface{}) bool { return true })
	n2 := &BufferNode{}
	Connect(n1, n2)
	s1 := ""
	s2 := ""
	a1 := MakeActionNode(func(item interface{}) {
		s1 = fmt.Sprintf("%v", item)
	})
	a2 := MakeActionNode(func(item interface{}) {
		s2 = fmt.Sprintf("%v", item)
	})
	Connect(n2, a1)
	Connect(n2, a2)

	n1.Receive(2)
	if s1 != "2" {
		t.Errorf("First listener not invoked: %s", s1)
	}
	if s2 != "2" {
		t.Errorf("Second listener not invoked: %s", s2)
	}
	n1.Receive(3)
	n1.Receive(4)
	expect := []int{2, 3, 4, 5}
	c3 := n2.GetCursor()
	n1.Receive(5)
	got := []int{}
	for item, present := c3.Next(); present; item, present = c3.Next() {
		got = append(got, item.(int))
	}
	if len(expect) != len(got) {
		t.Errorf("BufferNode iteration wrong count: want %v, got %v", expect, got)
	}
	sort.Ints(got)
	for i, exp := range expect {
		if exp != got[i] {
			t.Errorf("Valus differ: want %v, got %v", exp, got[i])
		}
	}
}

func TestJoinNode(t *testing.T) {
	root_node := MakeTestNode(func(item interface{}) bool { return true })
	root_node.label = "root"
	n1 := MakeTestNode(
		func(item interface{}) bool {
			_, is := item.(string)
			return is
		})
	n1.label = "letters"
	Connect(root_node, n1)
	log1 := MakeActionNode(
		func(item interface{}) {
			t.Logf("log1: %#v", item)
		})
	log1.label = "log1"
	Connect(n1, log1)

	n2 := MakeTestNode(
		func(item interface{}) bool {
			_, is := item.(int)
			return is
		})
	n2.label = "digits"
	Connect(root_node, n2)
	log2 := MakeActionNode(
		func(item interface{}) {
			t.Logf("log2: %#v", item)
		})
	log2.label = "log2"
	Connect(n2, log2)

	jn := Join("join", log1, log2)
	outputs := []string{}
	output_node := MakeActionNode(
		func(item interface{}) {
			t.Logf("joined %#v", item)
			pair := item.([2]interface{})
			s := pair[0].(string)
			i := pair[1].(int)
			outputs = append(outputs, fmt.Sprintf("%s%d", s, i))
		})
	Connect(jn, output_node)
	Initialize(root_node)

	Walk(root_node, func(n Node) {
		t.Logf(`node "%s" %T`, n.Label(), n)
	})

	root_node.Receive(1)
	root_node.Receive(2)
	root_node.Receive("a")
	root_node.Receive("b")
	root_node.Receive(3)
	expect := []string{
		"a1", "a2", "b1", "b2", "a3", "b3",
	}
	if len(expect) != len(outputs) {
		t.Errorf("wrong count: want %v, got %v", expect, outputs)
	} else {
		for i, exp := range expect {
			if exp != outputs[i] {
				t.Errorf("Values differ: want %v, got %v", exp, outputs[i])
			}
		}
	}
//	t.Logf("bn1 items: %#v", bn1.items)
//	t.Logf("bn2 items: %#v", bn2.items)
}
