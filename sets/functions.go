package sets

import "github.com/leshless/golibrary/set"

func Union[K comparable](a, b set.T[K]) set.T[K] {
	result := set.New[K]()
	for item := range a {
		result.Add(item)
	}
	for item := range b {
		result.Add(item)
	}

	return result
}

func Intersection[K comparable](a, b set.T[K]) set.T[K] {
	result := set.New[K]()
	for item := range a {
		if b.Contains(item) {
			result.Add(item)
		}
	}

	return result
}

func Diff[K comparable](a, b set.T[K]) set.T[K] {
	result := set.New[K]()
	for item := range a {
		if !b.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

func SymDiff[K comparable](a, b set.T[K]) set.T[K] {
	result := set.New[K]()
	for item := range a {
		if !b.Contains(item) {
			result.Add(item)
		}
	}
	for item := range b {
		if !a.Contains(item) {
			result.Add(item)
		}
	}

	return result
}

func IsSubset[K comparable](subset, set set.T[K]) bool {
	for item := range subset {
		if !set.Contains(item) {
			return false
		}
	}

	return true
}
