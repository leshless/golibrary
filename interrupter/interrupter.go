package interrupt

import (
	"context"
	"sync"
)

type Interrupter interface {
	Context() context.Context
}

type interrupter struct {
	ctx context.Context
	mu  sync.RWMutex
}

var _ Interrupter = (*interrupter)(nil)

func (i *interrupter) Context() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.ctx
}
