package channel

import (
	"testing"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skip_TestFnChRunsPassedFnInGoroutine(t *testing.T) {
	var isCalled bool
	fn := func() any {
		isCalled = true
		return nil
	}
	FnChan(fn)
	assert.Equal(t, false, isCalled)
}

func TestFnChReturnsFnResultInChannel(t *testing.T) {
	expected := 1234
	fn := func() int { return expected }
	ch := FnChan(fn)

	result, err := test.ChTimeout(t, ch)
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestFnChClosesReturnCh(t *testing.T) {
	fn := func() int { return 1 }
	ch := FnChan(fn)

	v1, err := test.ChTimeout(t, ch)
	require.NoError(t, err)
	assert.Equal(t, 1, v1)

	v2, err := test.ChTimeout(t, ch)
	require.NoError(t, err)
	assert.Equal(t, 0, v2) // closed zero value
}
