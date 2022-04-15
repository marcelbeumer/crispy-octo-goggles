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
	ch     chan chatbox.Event
	done   chan struct{}
	userCh map[string]chan chatbox.Event
	state  chatbox.RoomState
	l      *sync.RWMutex
}

func (r *Room) Start() error {
	if r.done != nil {
		return errors.New("already started")
	}
	r.done = make(chan struct{})
	go (func() {
		for {
			select {
			case e := <-r.ch:
				if err := r.handleEvent(e); err != nil {
					fmt.Println(err)
				}
			case <-r.done:
				break
			}
		}
	})()
	return nil
}

func (r *Room) handleEvent(e chatbox.Event) error {
	switch e.Name {

	case chatbox.InitialUserState:
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
		r.emitEvents(ne)

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
		r.emitEvents(n1, n2)

	default:
		return chatbox.UhandledEventError{Event: e, Receiver: r.uuid}
	}
	return nil
}

func (r *Room) Stop() error {
	if r.done == nil {
		return errors.New("not started")
	}
	close(r.done)
	return nil
}

func (r *Room) WaitDone() {
	_ = <-r.done
}

func (r *Room) HasUser(uuid string) bool {
	for key := range r.userCh {
		if key == uuid {
			return true
		}
	}
	return false
}

func (r *Room) AddUser(p chatbox.User) error {
	if found := r.HasUser(p.Uuid()); found {
		return errors.New("user already added")
	}

	r.l.Lock()
	uuid := p.Uuid()
	r.state.Users[uuid] = chatbox.UserState{}
	userCh := make(chan chatbox.Event)
	r.userCh[uuid] = userCh
	r.l.Unlock()

	in := (<-chan chatbox.Event)(userCh)
	out := (chan<- chatbox.Event)(r.ch)

	p.Chan(&in, &out)
	userCh <- chatbox.Event{Sender: r.uuid, Name: chatbox.RequestInitialUserState}

	return nil
}

func (r *Room) emitEvents(events ...chatbox.Event) {
	for _, e := range events {
		for _, ch := range r.userCh {
			ch <- e
		}
	}
}

func NewRoom() *Room {
	return &Room{
		ch: make(chan chatbox.Event),
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
	if found := r.HasUser(uuid); !found {
		err := chatbox.NewEventError(e, fmt.Sprintf("user %s not found", uuid))
		return uuid, data, err
	}
	return uuid, data, nil
}
