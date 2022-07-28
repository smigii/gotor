package linklist

import (
	"testing"
)

func TestLinkedList_AddData(t *testing.T) {

	ll := Make[int]()

	nLoops := 20
	nodes := make([]*Node[int], 0, nLoops)

	for i := 0; i < nLoops; i++ {
		n := ll.AddDataFront(i)
		nodes = append(nodes, n)
	}

	cur := ll.Head()
	for i := 0; i < nLoops; i++ {
		if cur.Data != nodes[nLoops-1-i].Data {
			t.Errorf("idx %v, got %v, want %v", i, cur.Data, nodes[i].Data)
		}
		cur = cur.next
	}

}

func TestLinkedList_Remove(t *testing.T) {
	ll := Make[int]()

	// Single
	e1 := ll.AddDataFront(20)
	ll.Remove(e1)

	if ll.size != 0 || ll.Head() != nil {
		t.Error("list should be empty")
	}

	// 2, remove front
	ll.Clear()
	e1 = ll.AddDataFront(1)
	e2 := ll.AddDataFront(2)
	ll.Remove(e2)

	if ll.Head() != e1 {
		t.Errorf("wrong head")
	}

	// 2, remove back
	ll.Clear()
	e1 = ll.AddDataFront(1)
	e2 = ll.AddDataFront(2)
	ll.Remove(e1)

	if ll.Head() != e2 {
		t.Errorf("wrong head")
	}

	// 3, remove middle
	ll.Clear()
	e1 = ll.AddDataFront(1)
	e2 = ll.AddDataFront(2)
	e3 := ll.AddDataFront(3)
	ll.Remove(e2)

	if e1.prev != e3 || e3.next != e1 {
		t.Errorf("bad remove")
	}
}
