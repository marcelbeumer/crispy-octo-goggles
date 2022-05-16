package chat

import (
	"testing"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/test"
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

	result := test.ChTimeout(t, ch)
	assert.Equal(t, expected, result)
}

func TestFnChClosesReturnCh(t *testing.T) {
	fn := func() int { return 1 }
	ch := fnCh(fn)

	assert.Equal(t, 1, test.ChTimeout(t, ch))
	assert.Equal(t, 0, test.ChTimeout(t, ch)) // closed zero value
}
