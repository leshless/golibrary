package chans

func ReadAll[A any](ch <-chan A) []A {
	as := make([]A, 0, len(ch))
	for {
		if a, ok := <-ch; ok {
			as = append(as, a)
			continue
		}

		break
	}

	return as
}
