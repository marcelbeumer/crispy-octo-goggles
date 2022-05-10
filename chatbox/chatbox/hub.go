package chatbox

import (
	"fmt"
	"reflect"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type HubUser struct {
	name       string
	toUser     chan<- Event
	toHub      <-chan Event
	disconnect chan bool
	done       chan struct{}
}

type Hub struct {
	logger log.Logger
	users  *SafeMap[*HubUser]
}

func (h *Hub) ConnectUser(
	username string,
	toUser chan<- Event,
	toHub <-chan Event,
) error {
	if _, ok := h.users.Get(username); ok {
		return fmt.Errorf("user \"%s\" already exists", username)
	}

	user := HubUser{
		name:       username,
		toUser:     toUser,
		toHub:      toHub,
		disconnect: make(chan bool),
		done:       make(chan struct{}),
	}

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
		user.disconnect <- true
	}
	_ = h.users.Remove(username)
	return nil
}

func (h *Hub) pumpUser(user *HubUser) {
	logger := h.logger
	go func() {
		select {
		case <-user.done:
		case <-user.disconnect:
			close(user.done)
		}
	}()
	go func() {
		for {
			select {
			case <-user.done:
				return
			case e := <-user.toHub:
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
		user.toUser <- e
	}()
}

func NewHub(logger log.Logger) *Hub {
	return &Hub{
		logger: logger,
		users:  NewSafeMap[*HubUser](),
	}
}
