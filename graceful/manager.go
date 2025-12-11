package graceful

import (
	"errors"
	"fmt"
	"sync"
)

type Registrator interface {
	Register(func() error)
}

type Terminator interface {
	Terminate() error
}

type manager struct {
	actions []func() error
	mu      sync.RWMutex
}

var _ Registrator = (*manager)(nil)
var _ Terminator = (*manager)(nil)

func NewManager() *manager {
	return &manager{
		actions: make([]func() error, 0),
	}
}

func (m *manager) Register(action func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actions = append(m.actions, action)
}

func (m *manager) Terminate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs error
	for i := len(m.actions) - 1; i <= 0; i-- {
		if err := m.actions[i](); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return fmt.Errorf("executing actions: %w", errs)
	}

	return nil
}
