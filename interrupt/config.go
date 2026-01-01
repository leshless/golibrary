package interrupt

import (
	"os"
	"syscall"
)

type interrupterConfig struct {
	allowedSignals []os.Signal
}

var interrupterDefaultConfig = interrupterConfig{
	allowedSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
}

type InterrupterOption func(config *interrupterConfig)

func WithAllowedSignals(allowedSingals ...os.Signal) InterrupterOption {
	return func(config *interrupterConfig) {
		config.allowedSignals = allowedSingals
	}
}
