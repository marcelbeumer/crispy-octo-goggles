package chat

import (
	"errors"
	"testing"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEvent struct{}

func (testevent *TestEvent) When() time.Time {
	return time.Now()
}

// TestNewTestConnectionUsesPassedInCh tests if passed input channel
// is used correctly
func TestNewTestConnectionUsesPassedInCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := NewTestConnection(chIn, chOut)

	testEvent := &TestEvent{}
	go func() { chIn <- testEvent }()

	err := test.GoTimeout(t, func() error {
		event, err := c.ReadEvent()
		assert.Equal(t, event, testEvent)
		return err
	})

	require.NoError(t, err)
}

// TestNewTestConnectionUsesPassedOutCh tests if passed output channel
// is used correctly
func TestNewTestConnectionUsesPassedOutCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := NewTestConnection(chIn, chOut)
	testEvent := &TestEvent{}
	go c.SendEvent(testEvent)

	event, err := test.ChTimeout(t, chOut)
	require.NoError(t, err)

	assert.Equal(t, testEvent, event)
}

// TestTestConnectionSendEvent tests if sendEvents uses the outCh to
// send the event.
func TestTestConnectionSendEvent(t *testing.T) {
	chOut := make(chan Event)
	c := TestConnection{
		EventOutCh: chOut,
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	testEvent := &TestEvent{}
	g := test.ErrGroup{}

	g.Go(func() error {
		return c.SendEvent(testEvent)
	})

	var event Event
	g.Go(func() error {
		event = <-chOut
		return nil
	})

	err := g.WaitTimeout(t)
	require.NoError(t, err)
	require.Equal(t, testEvent, event)
}

// TestTestConnectionSendEventClosed tests sending when the
// connection is closed.
func TestTestConnectionSendEventClosed(t *testing.T) {
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}
	err := c.Close(nil)
	require.NoError(t, err)

	err = test.GoTimeout(t, func() error {
		return c.SendEvent(&TestEvent{})
	})

	assert.ErrorIs(t, err, ErrConnectionClosed)
}

// TestTestConnectionReadEvent tests reading an event
func TestTestConnectionReadEvent(t *testing.T) {
	chIn := make(chan Event)
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  chIn,
		closed:     make(chan struct{}),
	}

	g := test.ErrGroup{}
	testEvent := &TestEvent{}

	g.Go(func() error {
		chIn <- testEvent
		return nil
	})

	var event Event
	g.Go(func() error {
		e, err := c.ReadEvent()
		event = e
		return err
	})

	err := g.WaitTimeout(t)
	require.NoError(t, err)
	assert.Equal(t, testEvent, event)
}

// TestTestConnectionReadEventClosed tests behavior reading events
// when connection is closed
func TestTestConnectionReadEventClosed(t *testing.T) {
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}
	err := c.Close(nil)
	require.NoError(t, err)
	err = test.GoTimeout(t, func() error {
		_, err := c.ReadEvent()
		return err
	})

	assert.ErrorIs(t, err, ErrConnectionClosed)
}

// TestTestConnectionClosed tests Closed method
func TestTestConnectionClosed(t *testing.T) {
	closed := make(chan struct{})
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     closed,
	}

	before := c.Closed()
	close(closed)
	after := c.Closed()

	assert.Equal(t, false, before)
	assert.Equal(t, true, after)
}

// TestTestConnectionClose tests closing the connection.
func TestTestConnectionClose(t *testing.T) {
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	err := c.Close(nil)
	assert.NoError(t, err)

	err = c.Close(nil)
	assert.ErrorIs(t, err, ErrConnectionClosed)

	err = c.Close(nil)
	assert.ErrorIs(t, err, ErrConnectionClosed)
}

// TestTestConnectionErr tests getting the error from a closed connection.
func TestTestConnectionErr(t *testing.T) {
	c := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	testErr := errors.New("test error")

	err := c.Close(testErr)
	assert.NoError(t, err)
	assert.ErrorIs(t, c.Err(), testErr)
}
