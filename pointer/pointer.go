package pointer

func T[V any](v V) *V {
	return &v
}
