package room

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
)

type Room struct {
	uuid   string
	outCh  chan chatbox.Event
	userCh map[string]chan chatbox.Event
	doneCh chan struct{}
	state  chatbox.RoomState
	l      *sync.RWMutex
}

func (r *Room) Start() {
	if r.doneCh != nil {
		return
	}
	r.doneCh = make(chan struct{})
	go (func() {
		for {
			select {
			case e := <-r.outCh:
				if err := r.handleEvent(e); err != nil {
					fmt.Println("Error handling event", err)
				}
			case <-r.doneCh:
				r.doneCh = nil
				break
			}
		}

	})()
}

func (r *Room) Stop() {
	if r.doneCh != nil {
		close(r.doneCh)
	}
}

func (r *Room) Wait() {
	if r.doneCh != nil {
		<-r.doneCh
	}
}

func (r *Room) handleEvent(e chatbox.Event) error {
	switch e.Name {

	case chatbox.Connect:
		uuid, data, err := unpackUserEvent[chatbox.UserState](*r, e)
		if err != nil {
			return err
		}

		r.l.Lock()
		r.state.Users[uuid] = data
		r.l.Unlock()

		m := chatbox.UserRef{Uuid: uuid, State: data}
		ne, err := chatbox.NewEvent(chatbox.NewUser, m, r.uuid)
		if err != nil {
			return err
		}
		r.sendEventToAll(ne)

	case chatbox.SendMessage:
		uuid, data, err := unpackUserEvent[string](*r, e)
		if err != nil {
			return err
		}

		r.l.Lock()
		name := r.state.Users[uuid].Name
		msg := chatbox.Message{Sender: e.Sender, SenderName: name, Message: data}
		r.state.Messages = append(r.state.Messages, msg)
		r.l.Unlock()

		n1, err := chatbox.NewEvent(chatbox.RoomStateUpdate, r.state, r.uuid)
		if err != nil {
			return err
		}
		n2, err := chatbox.NewEvent(chatbox.NewMessage, msg, r.uuid)
		if err != nil {
			return err
		}
		r.sendEventToAll(n1, n2)

	default:
		return chatbox.UhandledEventError{Event: e, Receiver: r.uuid}
	}
	return nil
}

func (r *Room) HasUuid(uuid string) bool {
	for key := range r.userCh {
		if key == uuid {
			return true
		}
	}
	return false
}

func (r *Room) Connect(uuid string) (in <-chan chatbox.Event, out chan<- chatbox.Event, err error) {
	if found := r.HasUuid(uuid); found {
		return nil, nil, errors.New("uuid already added")
	}

	userCh := make(chan chatbox.Event)

	r.l.Lock()
	r.state.Users[uuid] = chatbox.UserState{}
	r.userCh[uuid] = userCh
	r.l.Unlock()

	in = (<-chan chatbox.Event)(userCh)
	out = (chan<- chatbox.Event)(r.outCh)

	return in, out, nil
}

func (r *Room) sendEventToAll(events ...chatbox.Event) {
	for _, e := range events {
		for _, ch := range r.userCh {
			r.sendEvent(ch, e)
		}
	}
}

func (r *Room) sendEvent(ch chan chatbox.Event, e chatbox.Event) {
	go (func() {
		ch <- e
	})()
}

func NewRoom() *Room {
	return &Room{
		outCh: make(chan chatbox.Event),
		state: chatbox.RoomState{
			Users:    make(map[string]chatbox.UserState),
			Messages: make([]chatbox.Message, 0),
		},
		userCh: make(map[string]chan chatbox.Event),
		uuid:   "room:" + uuid.NewString(),
		l:      &sync.RWMutex{},
	}
}

func unpackUserEvent[T any](r Room, e chatbox.Event) (string, T, error) {
	uuid := e.Sender
	data, err := chatbox.GetData[T](e)
	if err != nil {
		return uuid, data, err
	}
	if found := r.HasUuid(uuid); !found {
		err := chatbox.NewEventError(e, fmt.Sprintf("user %s not found", uuid))
		return uuid, data, err
	}
	return uuid, data, nil
}
