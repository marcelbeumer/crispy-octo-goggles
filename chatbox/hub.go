package chatbox

import (
	"fmt"
	"reflect"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type HubUser struct {
	name string
	conn Connection
}

type Hub struct {
	logger log.Logger
	users  *SafeMap[*HubUser]
}

func (h *Hub) ConnectUser(
	username string,
	conn Connection,
) error {
	if _, ok := h.users.Get(username); ok {
		return fmt.Errorf("user \"%s\" already exists", username)
	}

	user := HubUser{name: username, conn: conn}
	h.users.Set(username, &user)

	h.pumpUser(&user)

	h.sendEvent(username, EventUserListUpdate{
		EventMeta: *NewEventMetaNow(),
		Users:     h.users.Keys(),
	})

	h.broadcastEvent(EventNewUser{
		EventMeta: *NewEventMetaNow(),
		Name:      username,
	}, username)

	return nil
}

func (h *Hub) DisconnectUser(username string) error {
	if user, ok := h.users.Get(username); ok {
		select {
		case <-user.conn.Closed():
		default:
			user.conn.Close()
		}
	}
	_ = h.users.Remove(username)
	return nil
}

func (h *Hub) pumpUser(user *HubUser) {
	logger := h.logger
	go func() {
		for {
			select {
			case <-user.conn.Closed():
				h.DisconnectUser(user.name)
				return
			case e := <-user.conn.ReceiveEvent():
				if e == nil {

				}
				if err := h.handleEvent(user.name, e); err != nil {
					logger.Error(
						"could not handle user event",
						map[string]any{
							"user":  user.name,
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
	for _, user := range h.users.Values() {
		if !except[user.name] {
			go h.sendEvent(user.name, e)
		}
	}
}

func (h *Hub) sendEvent(username string, e Event) {
	logger := h.logger
	user, ok := h.users.Get(username)
	if !ok {
		logger.Error(
			"can not send message to unknown user",
			map[string]any{
				"user": username,
			})
		return
	}
	go func() {
		user.conn.SendEvent(e)
	}()
}

func NewHub(logger log.Logger) *Hub {
	return &Hub{
		logger: logger,
		users:  NewSafeMap[*HubUser](),
	}
}
