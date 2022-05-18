package chat

import (
	"context"
	"errors"
)

// Connection is the interface that describes the connection between
// senders and receivers of events (e.g. client/server).
type Connection interface {
	// SendEvent sends event to the receiver.
	// Returns ErrConnectionClosed when connection closed.
	// Returns error when sending failed.
	SendEvent(e Event) error
	// ReadEvent wiats for next Event. Error when reading fails.
	// Returns error ErrConnectionClosed when connection closed.
	ReadEvent() (Event, error)
	// Wait waits until connection is closed.
	// Returns error with which the connection was closed (or nil)
	Wait() error
	// WaitContext waits until connection is closed.
	// Returns error with which the connection was closed (or nil)
	WaitContext(ctx context.Context) error
	// CLose closes connection, if connected.
	// Blocks until disconnected.
	Close(error) error
	// Closed return chan that is closed when connection is closed.
	Closed() bool
	// Err returns the error with which the connection closed, or nil.
	Err() error
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
	// error is the error the connection closed with
	error error
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
		return nil, ErrConnectionClosed
	case e := <-c.EventInCh:
		return e, nil
	}
}

func (c *TestConnection) Wait() error {
	<-c.closed
	return c.error
}

func (c *TestConnection) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closed:
	}
	return c.error
}

func (c *TestConnection) Close(err error) error {
	select {
	case <-c.closed:
		return ErrConnectionClosed
	default:
		c.error = err
		close(c.closed)
	}
	return nil
}

func (c *TestConnection) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *TestConnection) Err() error {
	return c.error
}

// NewTestConnection creates a TestConnection using the passed in/out channels.
func NewTestConnection(
	eventInCh <-chan Event,
	eventOutCh chan<- Event,
) *TestConnection {
	return &TestConnection{
		EventInCh:  eventInCh,
		EventOutCh: eventOutCh,
		closed:     make(chan struct{}),
	}
}
