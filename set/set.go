package set

type T[K comparable] map[K]struct{}

func New[K comparable]() T[K] {
	return make(T[K])
}

func FromSlice[K comparable](items []K) T[K] {
	s := make(T[K])
	for _, item := range items {
		s.Add(item)
	}

	return s
}

func (s T[K]) Add(item K) {
	s[item] = struct{}{}
}

func (s T[K]) Remove(item K) {
	delete(s, item)
}

func (s T[K]) Contains(item K) bool {
	_, exists := s[item]
	return exists
}
