// Package test implements some test helpers.
package test

import (
	"testing"
	"time"
)

// TimeoutDefault is the default timeout duration.
var TimeoutDefault = time.Second

// FatalTimeout calls t.Fatal with standard log message.
// Can not be run in a go routine.
func FatalTimeout(t *testing.T) {
	t.Fatal("timeout")
}

// ChTimeout wraps passed channel and either returns the next value of the
// channel or does a t.Fatal returning the empty value of the channel when
// the channel takes too long. Can not be run in a go routine.
func ChTimeout[T any](t *testing.T, ch <-chan T) T {
	select {
	case r := <-ch:
		return r
	case <-time.After(TimeoutDefault):
		FatalTimeout(t)
		var empty T
		return empty
	}
}

// GoTimeout runs passed function in a go routine and does a t.Fatal when
// the go routine takes too long. Can not be run in a go routine.
func GoTimeout(t *testing.T, fn func()) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(TimeoutDefault):
		FatalTimeout(t)
	}
}
