package optional

import (
	"encoding/json"
	"fmt"
)

const (
	nullPlaceholder = "NULL"
)

type T[V any] struct {
	value     V
	isNotNull bool
}

var _ fmt.Stringer = (*T[struct{}])(nil)
var _ json.Unmarshaler = (*T[struct{}])(nil)
var _ json.Marshaler = (*T[struct{}])(nil)

func Some[V any](value V) T[V] {
	return T[V]{
		value:     value,
		isNotNull: true,
	}
}

func None[V any]() T[V] {
	return T[V]{}
}

func FromPointer[V any](pointer *V) T[V] {
	if pointer == nil {
		return T[V]{}
	}

	return T[V]{
		value:     *pointer,
		isNotNull: true,
	}
}

func (t *T[V]) IsNull() bool {
	return !t.isNotNull
}

func (t *T[V]) Value() (V, bool) {
	return t.value, t.isNotNull
}

func (t *T[V]) Pointer() *V {
	value := t.value

	if t.isNotNull {
		return &value
	}

	return nil
}

func (t *T[V]) String() string {
	if t.isNotNull {
		return fmt.Sprintf("%v", t.value)
	}

	return nullPlaceholder
}

func (t *T[V]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = T[V]{}
		return nil
	}

	return json.Unmarshal(data, &t.value)
}

func (t *T[V]) MarshalJSON() ([]byte, error) {
	if !t.isNotNull {
		return []byte("null"), nil
	}

	return json.Marshal(t.value)
}
