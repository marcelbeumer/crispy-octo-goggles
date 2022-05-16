package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func skip_TestFnChRunsPassedFnInGoroutine(t *testing.T) {
	var isCalled bool
	fn := func() any {
		isCalled = true
		return nil
	}
	fnCh(fn)
	assert.Equal(t, false, isCalled)
}

func TestFnChReturnsFnResultInChannel(t *testing.T) {
	expected := 1234
	fn := func() int { return expected }
	ch := fnCh(fn)

	select {
	case result := <-ch:
		assert.Equal(t, expected, result)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestFnChClosesReturnCh(t *testing.T) {
	fn := func() int { return 1 }
	ch := fnCh(fn)

	select {
	case r := <-ch:
		assert.Equal(t, 1, r)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	select {
	case r := <-ch:
		assert.Equal(t, 0, r) // closed zero value
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
