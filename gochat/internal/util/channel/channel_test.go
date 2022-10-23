package channel

import (
	"testing"

	"github.com/marcelbeumer/go-playground/gochat/internal/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//lint:ignore U1000 will do later
func skip_TestFnChRunsPassedFnInGoroutine(t *testing.T) {
	var isCalled bool
	fn := func() any {
		isCalled = true
		return nil
	}
	FnToChan(fn)
	assert.Equal(t, false, isCalled)
}

func TestFnChReturnsFnResultInChannel(t *testing.T) {
	expected := 1234
	fn := func() int { return expected }
	ch := FnToChan(fn)

	result, err := test.ChTimeout(t, ch)
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestFnChClosesReturnCh(t *testing.T) {
	fn := func() int { return 1 }
	ch := FnToChan(fn)

	v1, err := test.ChTimeout(t, ch)
	require.NoError(t, err)
	assert.Equal(t, 1, v1)

	v2, err := test.ChTimeout(t, ch)
	require.NoError(t, err)
	assert.Equal(t, 0, v2) // closed zero value
}
