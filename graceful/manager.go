package graceful

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Registrator interface {
	Register(func(ctx context.Context) error)
}

type Terminator interface {
	Terminate(ctx context.Context) error
}

type manager struct {
	config  managerConfig
	actions []func(ctx context.Context) error

	mu           sync.RWMutex
	isTerminated atomic.Bool
}

var _ Registrator = (*manager)(nil)
var _ Terminator = (*manager)(nil)

func NewManager(options ...ManagerOption) *manager {
	config := managerDefaultConfig
	for _, option := range options {
		option(&config)
	}

	return &manager{
		config:  config,
		actions: make([]func(ctx context.Context) error, 0),
	}
}

func (m *manager) Register(action func(ctx context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actions = append(m.actions, action)
}

func (m *manager) Terminate(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isTerminated.Load() {
		return errors.New("already terminated")
	}

	ctx, cancel := context.WithTimeout(ctx, m.config.terminateTotalTimeout)
	errCh := make(chan error)

	go func() {
		errCh <- m.terminateSequentially(ctx)
	}()

	m.isTerminated.Store(true)

	select {
	case err := <-errCh:
		cancel()
		return err
	case <-ctx.Done():
		cancel()
		return ctx.Err()
	}
}

func (m *manager) terminateSequentially(ctx context.Context) error {
	for _, action := range m.actions {
		errCh := make(chan error, 1)
		ctx, cancel := context.WithTimeout(ctx, m.config.terminateActionTimeout)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic: %v", r)
				}
			}()

			errCh <- action(ctx)
		}()

		select {
		case err := <-errCh:
			cancel()
			if err != nil {
				return fmt.Errorf("executing action: %w", err)
			}
		case <-ctx.Done():
			cancel()
			return ctx.Err()
		}
	}

	return nil
}
