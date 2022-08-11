package ds

type Set[T Key] struct {
	set map[string]T
}

func MakeSet[T Key]() Set[T] {
	return Set[T]{
		set: make(map[string]T),
	}
}

func (s *Set[T]) Add(keyable T) bool {
	if !s.Has(keyable) {
		s.set[keyable.Key()] = keyable
		return true
	} else {
		return false
	}
}

func (s *Set[T]) Has(keyable T) bool {
	_, has := s.set[keyable.Key()]
	return has
}

func (s *Set[T]) Remove(keyable T) bool {
	if s.Has(keyable) {
		delete(s.set, keyable.Key())
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
