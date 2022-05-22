package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chat"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
)

type Connection struct {
	logger     log.Logger
	wsConn     *ws.Conn
	eventOutCh chan chat.Event
	l          sync.RWMutex
	closed     chan struct{}
	error      error
}

func (c *Connection) SendEvent(e chat.Event) error {
	c.l.Lock()
	defer c.l.Unlock()

	select {
	case <-c.closed:
		return chat.ErrConnectionClosed
	default:
	}

	m := Message{Data: e}
	eType := reflect.TypeOf(e)
	eTypeStr := eType.String()

	for name, handler := range handlers {
		if reflect.TypeOf(handler()) == eType {
			m.Name = name
			break
		}
	}

	if m.Name == "" {
		return fmt.Errorf("unknown event type <%s>", eTypeStr)
	}

	jsonText, err := json.Marshal(&m)
	if err != nil {
		return fmt.Errorf(
			"could not marshal event with type <%s>: %w",
			eTypeStr, err)
	}

	err = c.wsConn.WriteMessage(ws.TextMessage, jsonText)

	if err != nil {
		return fmt.Errorf(
			"error writing event to ws with type <%s>: %w",
			eTypeStr, err)
	}

	return nil
}

func (c *Connection) ReadEvent() (chat.Event, error) {
	select {
	case <-c.closed:
		return nil, chat.ErrConnectionClosed
	case e := <-c.eventOutCh:
		return e, nil
	}
}

func (c *Connection) Wait() error {
	<-c.closed
	return c.error
}

func (c *Connection) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closed:
	}
	return c.error
}

func (c *Connection) Close(err error) error {
	select {
	case <-c.closed:
		return chat.ErrConnectionClosed
	default:
		c.error = err
		close(c.closed)
		return c.wsConn.Close()
	}
}

func (c *Connection) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *Connection) Err() error {
	return c.error
}

func (c *Connection) wsReadPump() error {
	for {
		messageType, p, err := c.wsConn.ReadMessage()
		if err != nil {
			return err
		}

		switch messageType {
		case ws.TextMessage:
			var m Message
			if err := json.Unmarshal(p, &m); err != nil {
				return fmt.Errorf(`could not unmarshal message: %w`, err)
			}
			if m.Data == nil {
				return fmt.Errorf("data was nil after parsing message")
			}
			select {
			case <-c.closed:
				return chat.ErrConnectionClosed
			case c.eventOutCh <- m.Data:
				//
			}
		}
	}
}

func NewConnection(
	wsConn *ws.Conn,
	logger log.Logger,
) *Connection {
	conn := Connection{
		logger:     logger,
		wsConn:     wsConn,
		eventOutCh: make(chan chat.Event),
		closed:     make(chan struct{}),
	}
	go func() {
		defer conn.Close(nil)
		err := conn.wsReadPump()
		if errors.Is(err, chat.ErrConnectionClosed) {
			logger.Infow("websocket pump closed")
		} else {
			logger.Errorw("websocket pump error", log.Error(err))
		}
	}()
	return &conn
}
