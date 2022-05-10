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
	// Closed return chan that is closed when connection is closed
	Closed() <-chan struct{}
	// Disconnect disconnects from the server, if connected.
	// Blocks until disconnected.
	Close() error
}

type TestConnection struct {
	hub       *Hub
	toUser    chan Event
	toHub     chan Event
	closed    chan struct{}
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

// Closed return chan that is closed when connection is closed
func (c *TestConnection) Closed() <-chan struct{} {
	return c.closed
}

// Disconnect disconnects from the server, if connected.
// Blocks until disconnected.
func (c *TestConnection) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})
	return nil
}

func NewTestConnection(username string, logger log.Logger) (*TestConnection, error) {
	hub := NewHub(logger)
	conn := TestConnection{
		hub:       hub,
		toUser:    make(chan Event),
		toHub:     make(chan Event),
		closed:    make(chan struct{}),
		closeOnce: &sync.Once{},
	}
	if err := hub.ConnectUser(username, &conn); err != nil {
		return nil, err
	}
	return &conn, nil
}
