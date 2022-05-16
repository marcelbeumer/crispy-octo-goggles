package chat

import (
	"errors"
	"io"
)

// Connection is the interface that describes the connection between
// senders and receivers of events (e.g. client/server).
type Connection interface {
	// SendEvent sends event to the receiver.
	// Returns ErrConnectionClosed when connection closed.
	// Returns error when sending failed.
	SendEvent(e Event) error
	// ReceiveEvent wait for next Event. Error when reading fails.
	// Returns error io.EOF when connection closed.
	ReadEvent() (Event, error)
	// Closed return chan that is closed when connection is closed.
	Closed() bool
	// Disconnect disconnects from the server, if connected.
	// Blocks until disconnected.
	Close() error
}

// ErrConnectionClosed is the error returned when the connection is
// closed and still requesting I/O operations.
var ErrConnectionClosed = errors.New("connection closed")

// TestConnection is a simple example implementation of
// the Connection interface.
type TestConnection struct {
	// EventOutCh is the channel used for sending events.
	EventOutCh chan<- Event
	// EventInCh is the channel used for receiving events.
	EventInCh <-chan Event
	// closed is the channel used for marking the connection as
	// (permantently) closed.
	closed chan struct{}
}

func (c *TestConnection) SendEvent(e Event) error {
	select {
	case <-c.closed:
		return ErrConnectionClosed
	default:
		c.EventOutCh <- e
		return nil
	}
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

// NewTestConnection creates a TestConnection using the passed in/out channels.
func NewTestConnection(eventInCh <-chan Event, eventOutCh chan<- Event) *TestConnection {
	return &TestConnection{
		EventInCh:  eventInCh,
		EventOutCh: eventOutCh,
		closed:     make(chan struct{}),
	}
}
