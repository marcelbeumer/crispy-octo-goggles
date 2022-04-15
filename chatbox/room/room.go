package room

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
)

type Room struct {
	uuid   string
	ch     chan chatbox.Event
	done   chan struct{}
	userCh map[string]chan chatbox.Event
	state  chatbox.RoomState
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
		r.state.Users[uuid] = data

	case chatbox.SendMessage:
		uuid, data, err := unpackUserEvent[string](*r, e)
		if err != nil {
			return err
		}
		name := r.state.Users[uuid].Name
		msg := chatbox.Message{Sender: e.Sender, SenderName: name, Message: data}
		r.state.Messages = append(r.state.Messages, msg)
		r.emitEvent(chatbox.Event{Sender: r.uuid, Name: chatbox.RoomStateUpdate, Data: r.state})
		r.emitEvent(chatbox.Event{Sender: r.uuid, Name: chatbox.NewMessage, Data: msg})

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

	uuid := p.Uuid()
	r.state.Users[uuid] = chatbox.UserState{}

	userCh := make(chan chatbox.Event)
	r.userCh[uuid] = userCh
	in := (<-chan chatbox.Event)(userCh)
	out := (chan<- chatbox.Event)(r.ch)

	p.Chan(&in, &out)
	userCh <- chatbox.Event{Sender: r.uuid, Name: chatbox.RequestInitialUserState}

	return nil
}

func (r *Room) emitEvent(e chatbox.Event) {
	for _, ch := range r.userCh {
		ch <- e
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
