// Package test implements some test helpers.
package test

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// TimeoutDefault is the default timeout duration.
var TimeoutDefault = time.Second

// ErrTimeout is the error that is returned when something timed out
var ErrTimeout = errors.New("timeout")

// ChTimeout either returns value from passed channel or zero value
// of the channel and ErrTimeout
func ChTimeout[T any](t *testing.T, ch <-chan T) (T, error) {
	select {
	case r := <-ch:
		return r, nil
	case <-time.After(TimeoutDefault):
		var empty T
		return empty, ErrTimeout
	}
}

// GoTimeout runs passed function in a go routine and returns
// ErrTimeout when the function timed out
func GoTimeout(t *testing.T, fn func() error) error {
	result := make(chan error)
	go func() {
		defer close(result)
		result <- fn()
	}()
	select {
	case err := <-result:
		return err
	case <-time.After(TimeoutDefault):
		return ErrTimeout
	}
}

// ErrGroup is golang.org/x/sync/errgroup:ErrGroup with extra method(s)
type ErrGroup struct {
	errgroup.Group
}

// WaitTimeout waits for the group to finish or error or timeout.
// Returns ErrTimeout on timeout.
func (g *ErrGroup) WaitTimeout(t *testing.T) error {
	result := make(chan error)
	go func() {
		defer close(result)
		result <- g.Wait()
	}()
	select {
	case err := <-result:
		return err
	case <-time.After(TimeoutDefault):
		return ErrTimeout
	}
}
