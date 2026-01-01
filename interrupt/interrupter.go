package interrupt

import (
	"context"
	"os"
	"os/signal"
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

// NewInterrupter returns interrupter ready to provide it's own context that would be cancelled once one of the coresponding signal handlers is invoked
// Note that this constructor is not clean, i.e. new goroutine will be launched in the background
func NewInterrupter(options ...InterrupterOption) *interrupter {
	config := interrupterDefaultConfig
	for _, option := range options {
		option(&config)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, config.allowedSignals...)

	interrupter := &interrupter{}
	ctx, cancel := context.WithCancel(context.Background())
	interrupter.ctx = ctx

	go func() {
		<-signalCh
		close(signalCh)

		interrupter.mu.Lock()
		defer interrupter.mu.Unlock()

		cancel()
	}()

	return interrupter
}

func (i *interrupter) Context() context.Context {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.ctx
}
