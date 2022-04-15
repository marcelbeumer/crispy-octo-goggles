package room

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
)

type Room struct {
	uuid   string
	ctime  time.Time // current
	ptime  time.Time // previous
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
		uuid := e.Sender
		if found := r.HasUser(uuid); !found {
			return chatbox.NewEventError(e, fmt.Sprintf("user %s not found", uuid))
		}
		data, err := chatbox.GetData[chatbox.UserState](e)
		if err != nil {
			return err
		}
		r.state.Users[uuid] = data
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

func (r *Room) castState() {
	msg := chatbox.Event{Sender: r.uuid, Data: r.state}
	for _, ch := range r.userCh {
		ch <- msg
	}
}

func NewRoom() *Room {
	return &Room{
		ch: make(chan chatbox.Event),
		state: chatbox.RoomState{
			Users: make(map[string]chatbox.UserState),
		},
		userCh: make(map[string]chan chatbox.Event),
		uuid:   uuid.NewString(),
	}
}
