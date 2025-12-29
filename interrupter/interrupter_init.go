package interrupt

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func InitInterrupter() Interrupter {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

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
