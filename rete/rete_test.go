package rete

import "fmt"
import "reflect"
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

func TestTypeFilterNode(t *testing.T) {
	root := MakeRootNode()
    n1 := GetTypeFilterNode(root, reflect.TypeOf(""))
    Connect(root, n1)
	s0 := "foo"
    s1 := ""
	a1 := MakeActionNode(func(item interface{}) {
		s1 = fmt.Sprintf("%v", item)
	})
    Connect(n1, a1)
	root.Receive(1)
    root.Receive(s0)
    if s1 != s0 {
		t.Errorf("TypeFilterNode failed: want %s, got %s", s0, s1)
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
	n1.Receive(5)
	got := []int{}
	n2.DoItems(func(item interface{}) {
		got = append(got, item.(int))		
	})
	if len(expect) != len(got) {
		t.Errorf("BufferNode iteration wrong count: want %v, got %v", expect, got)
	}
	sort.Ints(got)
	for i, exp := range expect {
		if exp != got[i] {
			t.Errorf("Values differ: want %v, got %v", exp, got[i])
		}
	}
}

func TestBufferClear (t *testing.T) {
	n1 := MakeTestNode(func(item interface{}) bool { return true })
	n2 := &BufferNode{}
	Connect(n1, n2)
	n1.Receive(1)
	n1.Receive(2)
	Walk(n1, Node.Clear)
	n1.Receive(3)
	n1.Receive(4)
	got := []int{}
	expect := []int{ 3, 4 }
	n2.DoItems(func(item interface{}) {
		got = append(got, item.(int))
	})
	if len(expect) != len(got) {
		t.Errorf("BufferNode iteration wrong count: want %v, got %v", expect, got)
	}
	sort.Ints(got)
	for i, exp := range expect {
		if exp != got[i] {
			t.Errorf("Values differ: want %v, got %v", exp, got[i])
		}
	}
}

