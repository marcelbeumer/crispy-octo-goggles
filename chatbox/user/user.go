package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
)

type User struct {
	uuid         string
	initialState chatbox.UserState
	ctime        time.Time // current
	ptime        time.Time // previous
	in           *<-chan chatbox.Event
	out          *chan<- chatbox.Event
	done         chan struct{}
	room         chatbox.RoomState
}

func (u *User) Start() error {
	if u.done != nil {
		return errors.New("already started")
	}
	tick := time.Tick(16 * time.Millisecond)
	u.done = make(chan struct{})
	go (func() {
		for {
			select {
			case <-tick:
				u.Update()
			case e := <-*u.in:
				if err := u.handleEvent(e); err != nil {
					fmt.Println(err)
				}
			case <-u.done:
				break
			}
		}
	})()
	return nil
}

func (u *User) Stop() error {
	if u.done == nil {
		return errors.New("not started")
	}
	close(u.done)
	return nil
}

func (u *User) WaitDone() {
	_ = <-u.done
}

func (u *User) Update() {
	u.ptime = u.ctime
	u.ctime = time.Now()
	// for _, user := range w.users {
	// 	user.In <- UserUpdateMessage{}
	// }
}

func (u *User) Uuid() string {
	return u.uuid
}

func (u *User) Chan(in *<-chan chatbox.Event, out *chan<- chatbox.Event) {
	u.in = in
	u.out = out
}

func (u *User) handleEvent(e chatbox.Event) error {
	switch e.Name {

	case chatbox.RequestInitialUserState:
		*u.out <- chatbox.Event{Sender: u.uuid, Name: chatbox.InitialUserState, Data: u.initialState}

	case chatbox.RoomStateUpdate:
		data, err := chatbox.GetData[chatbox.RoomState](e)
		if err != nil {
			return err
		}
		u.room = data
	}

	return nil
}

func NewUser(initialState chatbox.UserState) *User {
	return &User{
		initialState: initialState,
		uuid:         uuid.NewString(),
	}
}
