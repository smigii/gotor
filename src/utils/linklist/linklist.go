package linklist

type LinkedList[T any] struct {
	head *Node[T]
	size int64
}

type Node[T any] struct {
	Data T
	next *Node[T]
	prev *Node[T]
}

// ============================================================================
// ============================================================================

func (n *Node[T]) Next() *Node[T] {
	return n.next
}

func (n *Node[T]) Prev() *Node[T] {
	return n.prev
}

// ============================================================================
// ============================================================================

func Make[T any]() LinkedList[T] {
	return LinkedList[T]{
		head: nil,
		size: 0,
	}
}

func (ll *LinkedList[T]) AddDataFront(data T) *Node[T] {
	n := Node[T]{
		Data: data,
		next: nil,
		prev: nil,
	}

	ll.AddNodeFront(&n)
	return &n
}

func (ll *LinkedList[T]) AddNodeFront(n *Node[T]) {
	if ll.size == 0 {
		n.prev = nil
		n.next = nil
		ll.head = n
	} else {
		n.next = ll.head
		n.prev = nil
		ll.head.prev = n
		ll.head = n
	}
	ll.size++
}

func (ll *LinkedList[T]) Remove(n *Node[T]) {
	if ll.size == 0 {
		panic("wtf?")
	} else if ll.size == 1 {
		n.prev = nil
		n.next = nil
		ll.head = nil
		ll.size = 0
		return
	}

	if n == ll.head {
		ll.head = n.next
	}
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	ll.size--
}

func (ll *LinkedList[T]) Clear() {
	ll.head = nil
	ll.size = 0
}

func (ll *LinkedList[T]) Head() *Node[T] {
	return ll.head
}
