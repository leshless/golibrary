package stupid

func Reflect[T any](t T) func() T {
	return func() T {
		return t
	}
}
