package chatbox

import (
	"sync"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type Connection interface {
	// SendEvent posts event. Non-blocking, shoot and forget
	SendEvent(e Event)
	// ReceiveEvent returns chan for Event
	ReceiveEvent() <-chan Event
	// Disconnect disconnects from the server, if connected.
	// Blocks until disconnected.
	Close() error
}

type TestConnection struct {
	hub       *Hub
	toUser    chan Event
	toHub     chan Event
	stop      chan struct{}
	closeOnce *sync.Once
}

// SendEvent posts event. Non-blocking, shoot and forget
func (c *TestConnection) SendEvent(e Event) {
	go func() {
		c.toHub <- e
	}()
}

// ReceiveEvent returns chan for Event
func (c *TestConnection) ReceiveEvent() <-chan Event {
	return c.toUser
}

// Disconnect disconnects from the server, if connected.
// Blocks until disconnected.
func (c *TestConnection) Close() error {
	c.closeOnce.Do(func() {
		close(c.stop)
	})
	return nil
}

func NewTestConnection(username string, logger log.Logger) (*TestConnection, error) {
	hub := NewHub(logger)
	toUser := make(chan Event)
	toHub := make(chan Event)
	if err := hub.ConnectUser(username, toUser, toHub); err != nil {
		return nil, err
	}
	return &TestConnection{
		hub:       hub,
		toUser:    toUser,
		toHub:     toHub,
		stop:      make(chan struct{}),
		closeOnce: &sync.Once{},
	}, nil
}
