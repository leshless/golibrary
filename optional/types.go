package optional

type String T[string]

func SomeString(value string) T[string] {
	return T[string]{
		value:     value,
		isNotNull: true,
	}
}

func NoneString() T[string] {
	return T[string]{}
}

func StringFromPointer(pointer *string) T[string] {
	if pointer == nil {
		return T[string]{}
	}

	return T[string]{
		value:     *pointer,
		isNotNull: true,
	}
}

type Int T[int]

func SomeInt(value int) T[int] {
	return T[int]{
		value:     value,
		isNotNull: true,
	}
}

func NoneInt() T[int] {
	return T[int]{}
}

func IntFromPointer(pointer *int) T[int] {
	if pointer == nil {
		return T[int]{}
	}

	return T[int]{
		value:     *pointer,
		isNotNull: true,
	}
}
