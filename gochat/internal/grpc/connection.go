package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/marcelbeumer/go-playground/gochat/internal/chat"
	"github.com/marcelbeumer/go-playground/gochat/internal/log"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcConnection interface {
	Send(*EventEnvelope) error
	Recv() (*EventEnvelope, error)
}

type Connection struct {
	eventOutCh chan chat.Event
	closed     chan struct{}
	error      error
	grpcConn   GrpcConnection
}

// SendEvent sends event to the receiver.
// Returns ErrConnectionClosed when connection closed.
// Returns error when sending failed.
func (c *Connection) SendEvent(e chat.Event) error {
	select {
	case <-c.closed:
		return chat.ErrConnectionClosed
	default:
	}

	typeStr := reflect.TypeOf(e).String()
	envelope := EventEnvelope{}
	time := timestamppb.New(e.When())

	switch t := e.(type) {
	case *chat.EventConnected:
		envelope.Event = &EventEnvelope_Connected{
			Connected: &Connected{
				Time:  time,
				Users: t.Users,
			},
		}

	case *chat.EventUserListUpdate:
		envelope.Event = &EventEnvelope_UserListUpdate{
			UserListUpdate: &UserListUpdate{
				Time:  time,
				Users: t.Users,
			},
		}

	case *chat.EventUserEnter:
		envelope.Event = &EventEnvelope_UserEnter{
			UserEnter: &UserEnter{
				Time: time,
				Name: t.Name,
			},
		}

	case *chat.EventUserLeave:
		envelope.Event = &EventEnvelope_UserLeave{
			UserLeave: &UserLeave{
				Time: time,
				Name: t.Name,
			},
		}

	case *chat.EventSendMessage:
		envelope.Event = &EventEnvelope_SendMessage{
			SendMessage: &SendMessage{
				Time:    time,
				Message: t.Message,
			},
		}

	case *chat.EventNewMessage:
		envelope.Event = &EventEnvelope_NewMessage{
			NewMessage: &NewMessage{
				Time:    time,
				Message: t.Message,
				Sender:  t.Sender,
			},
		}

	default:
		return fmt.Errorf("unknown event type <%s>", typeStr)
	}

	err := c.grpcConn.Send(&envelope)
	if err != nil {
		return fmt.Errorf(
			"error writing event to grpc with type <%s>: %w",
			typeStr, err)
	}
	return err
}

// ReadEvent wiats for next Event. Error when reading fails.
// Returns error ErrConnectionClosed when connection closed.
func (c *Connection) ReadEvent() (chat.Event, error) {
	select {
	case <-c.closed:
		return nil, chat.ErrConnectionClosed
	case e := <-c.eventOutCh:
		return e, nil
	}
}

// Wait waits until connection is closed.
// Returns error with which the connection was closed (or nil)
func (c *Connection) Wait() error {
	<-c.closed
	return c.error
}

// WaitContext waits until connection is closed.
// Returns error with which the connection was closed (or nil)
func (c *Connection) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closed:
	}
	return c.error
}

// CLose closes connection, if connected.
// Blocks until disconnected.
func (c *Connection) Close(err error) error {
	select {
	case <-c.closed:
		return chat.ErrConnectionClosed
	default:
		c.error = err
		close(c.closed)
	}
	return nil
}

// Closed return chan that is closed when connection is closed.
func (c *Connection) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// Err returns the error with which the connection closed, or nil.
func (c *Connection) Err() error {
	return c.error
}

func (h *Connection) grpcReadPump() error {
	for {
		envelope, err := h.grpcConn.Recv()
		if err != nil {
			return err
		}

		var e chat.Event

		switch t := envelope.Event.(type) {
		case *EventEnvelope_Connected:
			meta := chat.EventMeta{Time: t.Connected.Time.AsTime()}
			e = &chat.EventConnected{
				EventMeta: meta,
				Users:     t.Connected.Users,
			}

		case *EventEnvelope_UserListUpdate:
			meta := chat.EventMeta{Time: t.UserListUpdate.Time.AsTime()}
			e = &chat.EventUserListUpdate{
				EventMeta: meta,
				Users:     t.UserListUpdate.Users,
			}

		case *EventEnvelope_UserEnter:
			meta := chat.EventMeta{Time: t.UserEnter.Time.AsTime()}
			e = &chat.EventUserEnter{
				EventMeta: meta,
				Name:      t.UserEnter.Name,
			}

		case *EventEnvelope_UserLeave:
			meta := chat.EventMeta{Time: t.UserLeave.Time.AsTime()}
			e = &chat.EventUserEnter{
				EventMeta: meta,
				Name:      t.UserLeave.Name,
			}

		case *EventEnvelope_SendMessage:
			meta := chat.EventMeta{Time: t.SendMessage.Time.AsTime()}
			e = &chat.EventSendMessage{
				EventMeta: meta,
				Message:   t.SendMessage.Message,
			}

		case *EventEnvelope_NewMessage:
			meta := chat.EventMeta{Time: t.NewMessage.Time.AsTime()}
			e = &chat.EventNewMessage{
				EventMeta: meta,
				Sender:    t.NewMessage.Sender,
				Message:   t.NewMessage.Message,
			}

		default:
			return fmt.Errorf(
				"unknown grpc payload type <%s>",
				reflect.TypeOf(envelope.Event).String(),
			)

		}

		h.eventOutCh <- e
	}
}

func NewConnection(
	grpcConn GrpcConnection,
	logger log.Logger,
) *Connection {
	conn := Connection{
		eventOutCh: make(chan chat.Event),
		closed:     make(chan struct{}),
		error:      nil,
		grpcConn:   grpcConn,
	}
	go func() {
		defer conn.Close(nil)
		err := conn.grpcReadPump()
		if errors.Is(err, chat.ErrConnectionClosed) {
			logger.Infow("grpc pump closed")
		} else {
			logger.Errorw("grpc pump error", log.Error(err))
		}
	}()
	return &conn
}
