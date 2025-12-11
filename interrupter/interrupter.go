package interrupter

import (
	"sync"
	"sync/atomic"
)

type Interrupter interface {
	Ch() <-chan struct{}
}

type interrupter struct {
	chs           []chan struct{}
	mu            sync.RWMutex
	isInterrupted atomic.Bool
}

var _ Interrupter = (*interrupter)(nil)

func NewInterrupter() *interrupter {
	return &interrupter{
		chs: make([]chan struct{}, 0),
	}
}

func (i *interrupter) Ch() <-chan struct{} {
	ch := make(chan struct{}, 1)

	i.mu.Lock()
	defer i.mu.Unlock()
	i.chs = append(i.chs, ch)

	if i.isInterrupted.Load() {
		ch <- struct{}{}
		close(ch)
	}

	return ch
}
