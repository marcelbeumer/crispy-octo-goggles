package chat

import (
	"testing"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEvent struct{}

func (testevent *TestEvent) When() time.Time {
	return time.Now()
}

func TestNewTestConnectionUsesPassedInCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := NewTestConnection(chIn, chOut)
	testEvent := &TestEvent{}
	go func() { chIn <- testEvent }()

	var err error
	var event Event
	test.GoTimeout(t, func() {
		event, err = c.ReadEvent()
	})

	require.Equal(t, err, nil)
	assert.Equal(t, event, testEvent)
}

func TestNewTestConnectionUsesPassedOutCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := NewTestConnection(chIn, chOut)
	testEvent := &TestEvent{}
	go c.SendEvent(testEvent)

	event := test.ChTimeout(t, chOut)
	assert.Equal(t, event, testEvent)
}

func TestTestConnectionUsesInCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := TestConnection{
		EventOutCh: chOut,
		EventInCh:  chIn,
		closed:     make(chan struct{}),
	}
	testEvent := &TestEvent{}
	go func() { chIn <- testEvent }()

	var err error
	var event Event
	test.GoTimeout(t, func() {
		event, err = c.ReadEvent()
	})

	require.Equal(t, err, nil)
	assert.Equal(t, event, testEvent)
}

func TestTestConnectionUsesOutCh(t *testing.T) {
	chIn := make(chan Event)
	chOut := make(chan Event)
	c := TestConnection{
		EventOutCh: chOut,
		EventInCh:  chIn,
		closed:     make(chan struct{}),
	}
	testEvent := &TestEvent{}
	go c.SendEvent(testEvent)

	event := test.ChTimeout(t, chOut)
	assert.Equal(t, event, testEvent)
}
