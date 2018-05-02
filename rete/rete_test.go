package rete

import "fmt"
import "sort"
import "testing"

func TestActionNode(t *testing.T) {
	s := ""
	n := &ActionNode{
		actionFunction: func(item interface{}) {
			s = fmt.Sprintf("%v", item)
		},
	}
	n.Receive(5)
	if s != "5" {
		t.Errorf("Action not performed.")
	}
}

func TestTestNode(t *testing.T) {
	s := ""
	n1 := &TestNode{
		testFunction: func(item interface{}) bool {
			return item.(int) == 4
		},
	}
	n2 := &ActionNode{
		actionFunction: func(item interface{}) {
			s = fmt.Sprintf("%v", item)
		},
	}
	n1.OutputsTo(n2)
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
	n1 := &TestNode{
		testFunction: func(item interface{}) bool { return true },
	}
	n2 := &BufferNode{}
	n1.OutputsTo(n2)
	s1 := ""
	s2 := ""
	n2.AddListener(func(item interface{}) {
		s1 = fmt.Sprintf("%v", item)
	})
	n2.AddListener(func(item interface{}) {
		s2 = fmt.Sprintf("%v", item)
	})
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
}

func TestInitialization(t *testing.T) {
}
