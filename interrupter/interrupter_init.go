package interrupter

import (
	"os"
	"os/signal"
	"syscall"
)

func InitInterrupter() Interrupter {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	interrupter := &interrupter{}

	go func() {
		<-signalCh

		interrupter.isInterrupted.Store(true)

		interrupter.mu.RLock()
		defer interrupter.mu.RUnlock()

		for _, ch := range interrupter.chs {
			ch <- struct{}{}
			close(ch)
		}
	}()

	return interrupter
}
