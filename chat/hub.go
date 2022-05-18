package chat

import (
	"fmt"
	"reflect"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/safe"
)

// hubUser encapsulates a user in the hub.
type hubUser struct {
	conn Connection
	// Events are put in a queue for sending. This allows back pressure control
	// but this is not yet implemented.
	events *safe.Queue[Event]
}

// Hub is the chat hub/room where users can connect to.
type Hub struct {
	logger log.Logger
	users  *safe.Map[*hubUser]
}

func (h *Hub) ConnectUser(
	username string,
	conn Connection,
) error {
	if _, ok := h.users.Get(username); ok {
		return fmt.Errorf("user \"%s\" already exists", username)
	}

	user := hubUser{conn: conn, events: safe.NewQueue[Event]()}
	h.users.Set(username, &user)

	h.scheduleEvent(username, &EventConnected{})
	h.scheduleBroadcast(&EventUserEnter{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	}, username)
	h.scheduleBroadcast(&EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.users.Keys(),
	})

	var err error
	select {
	case err = <-fnCh(func() error { return h.pumpFromUser(username) }):
	case err = <-fnCh(func() error { return h.pumpToUser(username) }):
	}

	if err != nil {
		h.CloseUser(username)
	}

	return nil
}

func (h *Hub) CloseUser(username string) error {
	user, _ := h.users.Get(username)
	if user != nil && !user.conn.Closed() {
		user.conn.Close()
		user.events.Close()
	}
	_ = h.users.Delete(username)
	h.scheduleBroadcast(&EventUserLeave{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	})
	h.scheduleBroadcast(&EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.users.Keys(),
	})
	return nil
}

func (h *Hub) pumpToUser(username string) error {
	user, _ := h.users.Get(username)
	if user == nil {
		return fmt.Errorf(`user conn "%s" not found`, username)
	}
	for {
		e, err := user.events.Read()
		if err != nil {
			return err
		}
		err = user.conn.SendEvent(e)
		if err != nil {
			return err
		}
	}
}

func (h *Hub) pumpFromUser(username string) error {
	user, _ := h.users.Get(username)
	if user == nil {
		return fmt.Errorf(`user conn "%s" not found`, username)
	}
	for {
		e, err := user.conn.ReadEvent()
		if err != nil {
			return err
		}
		if err := h.handleEvent(username, e); err != nil {
			return nil
		}
	}
}

func (h *Hub) handleEvent(username string, e Event) error {
	logger := h.logger
	switch t := e.(type) {
	case *EventConnected:
	case *EventUserListUpdate:
	case *EventUserEnter:
	case *EventUserLeave:
	case *EventSendMessage:
		// Shoot and forget
		go h.scheduleBroadcast(&EventNewMessage{
			EventMeta: *NewEventMetaNow(),
			Sender:    username,
			Message:   t.Message,
		})
	case *EventNewMessage:
	default:
		logger.Warnw(
			"unhandled event type",
			"user", username,
			"type", reflect.TypeOf(e).String())
	}
	return nil
}

func (h *Hub) scheduleBroadcast(e Event, exceptUsers ...string) error {
	except := map[string]bool{}
	for _, n := range exceptUsers {
		except[n] = true
	}
	for _, username := range h.users.Keys() {
		if !except[username] {
			err := h.scheduleEvent(username, e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hub) scheduleEvent(username string, e Event) error {
	user, ok := h.users.Get(username)
	if !ok {
		return fmt.Errorf(`unknown user conn "%s"`, username)
	}
	if err := user.events.Add(e); err != nil {
		h.CloseUser(username) // unforgiving
		return err
	}
	return nil
}

func NewHub(logger log.Logger) *Hub {
	return &Hub{
		logger: logger,
		users:  safe.NewMap[*hubUser](),
	}
}
