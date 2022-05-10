package chatbox

import (
	"io"
)

type Connection interface {
	// SendEvent posts event. Non-blocking, shoot and forget
	SendEvent(e Event)
	// ReceiveEvent wait for next Event. Error when reading fails.
	// Returns error io.EOF when connection closed
	ReadEvent() (Event, error)
	// Closed return chan that is closed when connection is closed
	Closed() bool
	// Disconnect disconnects from the server, if connected.
	// Blocks until disconnected.
	Close() error
}

type TestConnection struct {
	EventOutCh chan Event
	EventInCh  chan Event
	closed     chan struct{}
}

func (c *TestConnection) SendEvent(e Event) {
	go func() {
		c.EventOutCh <- e
	}()
}

func (c *TestConnection) ReadEvent() (Event, error) {
	select {
	case <-c.closed:
		return nil, io.EOF
	case e := <-c.EventInCh:
		return e, nil
	}
}

func (c *TestConnection) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *TestConnection) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func NewTestConnection(eventInCh chan Event, eventOutCh chan Event) *TestConnection {
	return &TestConnection{
		EventInCh:  eventInCh,
		EventOutCh: eventOutCh,
		closed:     make(chan struct{}),
	}
}
