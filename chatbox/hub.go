package chatbox

import (
	"fmt"
	"reflect"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
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
	h.pumpUser(username)

	h.sendEvent(username, EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.connections.Keys(),
	})

	h.broadcastEvent(EventNewUser{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	}, username)

	return nil
}

func (h *Hub) DisconnectUser(username string) error {
	if conn, ok := h.connections.Get(username); ok {
		select {
		case <-conn.Closed():
		default:
			conn.Close()
		}
	}
	_ = h.connections.Remove(username)
	return nil
}

func (h *Hub) pumpUser(username string) {
	conn, _ := h.connections.Get(username)
	if conn == nil {
		return
	}
	logger := h.logger
	go func() {
		for {
			select {
			case <-conn.Closed():
				h.DisconnectUser(username)
				return
			case e := <-conn.ReceiveEvent():
				if e == nil {

				}
				if err := h.handleEvent(username, e); err != nil {
					logger.Error(
						"could not handle user event",
						map[string]any{
							"user":  username,
							"error": err.Error(),
						})
				}
			}
		}
	}()
}

func (h *Hub) handleEvent(username string, e Event) error {
	logger := h.logger
	switch t := e.(type) {
	case EventUserListUpdate:
	case EventNewUser:
	case EventSendMessage:
		h.broadcastEvent(EventNewMessage{
			EventMeta: *NewEventMetaNow(),
			Sender:    username,
			Message:   t.Message,
		})
	case EventNewMessage:
	default:
		logger.Warn("unhandled event type",
			map[string]any{
				"type": reflect.TypeOf(e).String(),
			})
	}
	return nil
}

func (h *Hub) broadcastEvent(e Event, exceptUsers ...string) {
	except := map[string]bool{}
	for _, n := range exceptUsers {
		except[n] = true
	}
	for _, username := range h.connections.Keys() {
		if !except[username] {
			go h.sendEvent(username, e)
		}
	}
}

func (h *Hub) sendEvent(username string, e Event) {
	logger := h.logger
	conn, ok := h.connections.Get(username)
	if !ok {
		logger.Error(
			"can not send message to unknown user",
			map[string]any{
				"user": username,
			})
		return
	}
	go func() {
		conn.SendEvent(e)
	}()
}

func NewHub(logger log.Logger) *Hub {
	return &Hub{
		logger:      logger,
		connections: NewSafeMap[Connection](),
	}
}
