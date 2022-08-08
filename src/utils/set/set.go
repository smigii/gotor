package set

type Key interface {
	Key() string
}

type Set[T Key] struct {
	set map[string]T
}

func MakeSet[T Key]() Set[T] {
	return Set[T]{
		set: make(map[string]T),
	}
}

func (s *Set[T]) Add(hashable T) bool {
	if !s.Has(hashable) {
		s.set[hashable.Key()] = hashable
		return true
	} else {
		return false
	}
}

func (s *Set[T]) Has(hashable T) bool {
	_, has := s.set[hashable.Key()]
	return has
}

func (s *Set[T]) Remove(hashable T) bool {
	if s.Has(hashable) {
		delete(s.set, hashable.Key())
		return true
	} else {
		return false
	}
}

func (s *Set[T]) Size() int {
	return len(s.set)
}

func (s *Set[T]) Items() []T {
	items := make([]T, len(s.set))
	for _, v := range s.set {
		items = append(items, v)
	}
	return items
}
