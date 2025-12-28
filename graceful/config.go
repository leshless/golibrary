package graceful

import "time"

type managerConfig struct {
	terminateTotalTimeout   time.Duration
	terminateActionTimeout  time.Duration
	enableTerminateParallel bool
}

var managerDefaultConfig = managerConfig{
	terminateTotalTimeout:   time.Second * 20,
	terminateActionTimeout:  time.Second * 5,
	enableTerminateParallel: false,
}

type ManagerOption func(config *managerConfig)

func WithTerminateTotalTimeout(timeout time.Duration) ManagerOption {
	return func(config *managerConfig) {
		config.terminateTotalTimeout = timeout
	}
}

func WithTerminateActionTimeout(timeout time.Duration) ManagerOption {
	return func(config *managerConfig) {
		config.terminateActionTimeout = timeout
	}
}

func WithEnableTerminateParallel() ManagerOption {
	return func(config *managerConfig) {
		config.enableTerminateParallel = true
	}
}
