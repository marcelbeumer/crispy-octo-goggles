package chat

import (
	"fmt"
	"reflect"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"golang.org/x/sync/errgroup"
)

type Hub struct {
	logger      log.Logger
	connections *SafeMap[Connection]
}

func (h *Hub) ConnectUser(
	username string,
	conn Connection,
) error {
	if _, ok := h.connections.Get(username); ok {
		return fmt.Errorf("user \"%s\" already exists", username)
	}

	h.connections.Set(username, conn)
	done := make(chan struct{})
	defer close(done)

	// Shoot and forget
	go h.broadcastEvent(&EventUserEnter{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	}, username)

	// Shoot and forget
	go h.broadcastEvent(&EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.connections.Keys(),
	})

	if err := h.pumpUser(username); err != nil {
		h.CloseUser(username)
		return err
	}
	return nil
}

func (h *Hub) CloseUser(username string) error {
	conn, _ := h.connections.Get(username)
	if conn != nil && !conn.Closed() {
		conn.Close()
	}
	_ = h.connections.Remove(username)
	// Shoot and forget
	go h.broadcastEvent(&EventUserLeave{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	})
	// Shoot and forget
	go h.broadcastEvent(&EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.connections.Keys(),
	})
	return nil
}

func (h *Hub) pumpUser(username string) error {
	conn, _ := h.connections.Get(username)
	if conn == nil {
		return fmt.Errorf(`user conn "%s" not found`, username)
	}
	for {
		e, err := conn.ReadEvent()
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
	case *EventUserListUpdate:
	case *EventUserEnter:
	case *EventUserLeave:
	case *EventSendMessage:
		// Shoot and forget
		go h.broadcastEvent(&EventNewMessage{
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

func (h *Hub) broadcastEvent(e Event, exceptUsers ...string) error {
	except := map[string]bool{}
	for _, n := range exceptUsers {
		except[n] = true
	}
	eg := errgroup.Group{}
	for _, username := range h.connections.Keys() {
		if !except[username] {
			username := username
			eg.Go(func() error {
				return h.sendEvent(username, e)
			})
		}
	}
	return eg.Wait()
}

func (h *Hub) sendEvent(username string, e Event) error {
	conn, ok := h.connections.Get(username)
	if !ok {
		return fmt.Errorf(`unknown user conn "%s"`, username)
	}
	if err := conn.SendEvent(e); err != nil {
		h.CloseUser(username) // unforgiving
		return err
	}
	return nil
}

func NewHub(logger log.Logger) *Hub {
	return &Hub{
		logger:      logger,
		connections: NewSafeMap[Connection](),
	}
}
