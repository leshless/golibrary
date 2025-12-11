package set

func Union[K comparable](a, b T[K]) T[K] {
	result := New[K]()
	for item := range a {
		result.Add(item)
	}
	for item := range b {
		result.Add(item)
	}

	return result
}

func Intersection[K comparable](a, b T[K]) T[K] {
	result := New[K]()
	for item := range a {
		if b.Contains(item) {
			result.Add(item)
		}
	}

	return result
}

func Diff[K comparable](a, b T[K]) T[K] {
	result := New[K]()
	for item := range a {
		if !b.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

func SymDiff[K comparable](a, b T[K]) T[K] {
	result := New[K]()
	for item := range a {
		if !b.Contains(item) {
			result.Add(item)
		}
	}
	for item := range b {
		if !a.Contains(item) {
			a.Add(item)
		}
	}

	return result
}

func IsSubset[K comparable](subset, set T[K]) bool {
	for item := range subset {
		if !set.Contains(item) {
			return false
		}
	}

	return true
}
