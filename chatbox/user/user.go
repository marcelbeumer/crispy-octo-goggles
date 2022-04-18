package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
)

type User struct {
	uuid         string
	initialState chatbox.UserState
	in           <-chan chatbox.Event
	out          chan<- chatbox.Event
	doneCh       chan struct{}
	room         chatbox.RoomState
	CanPrint     bool
}

func (u *User) Start() {
	if u.doneCh != nil {
		return
	}
	u.doneCh = make(chan struct{})
	go (func() {
		for {
			select {
			case e := <-u.in:
				if err := u.handleEvent(e); err != nil {
					fmt.Println(err)
				}
			case <-u.doneCh:
				u.doneCh = nil
				break
			}
		}
	})()
}

func (u *User) Stop() {
	if u.doneCh != nil {
		close(u.doneCh)
	}
}

func (u *User) Wait() {
	if u.doneCh != nil {
		<-u.doneCh
	}
}

func (u *User) Uuid() string {
	return u.uuid
}

func (u *User) Chan(in <-chan chatbox.Event, out chan<- chatbox.Event) {
	u.in = in
	u.out = out
}

func (u *User) SendMessage(m string) {
	e, err := chatbox.NewEvent(chatbox.SendMessage, m, u.uuid)
	if err != nil {
		panic(err)
	}
	go u.sendEvent(e)
}

func (u *User) handleEvent(e chatbox.Event) error {
	switch e.Name {

	case chatbox.RequestInitialUserState:
		e, err := chatbox.NewEvent(chatbox.InitialUserState, u.initialState, u.uuid)
		if err != nil {
			return err
		}
		go u.sendEvent(e)

	case chatbox.RoomStateUpdate:
		data, err := chatbox.GetData[chatbox.RoomState](e)
		if err != nil {
			return err
		}
		u.room = data

	case chatbox.NewUser:
		data, err := chatbox.GetData[chatbox.UserRef](e)
		if err != nil {
			return err
		}
		name := data.State.Name
		if data.Uuid == u.uuid {
			name = fmt.Sprintf("%s (you)", name)
		}
		u.printf(
			"[%s] << %s entered the room >>\n",
			time.Now().Local(),
			name,
		)

	case chatbox.NewMessage:
		data, err := chatbox.GetData[chatbox.Message](e)
		if err != nil {
			return err
		}
		u.printf("[%s %s] %s\n", time.Now().Local(), data.SenderName, data.Message)

	default:
		return chatbox.UhandledEventError{Event: e, Receiver: u.uuid}
	}

	return nil
}

func (u *User) sendEvent(e chatbox.Event) {
	go (func() {
		u.out <- e
	})()
}

func (u *User) printf(format string, a ...any) {
	if u.CanPrint {
		s := fmt.Sprintf(format, a...)
		fmt.Print(s)
	}
}

func NewUser(initialState chatbox.UserState) *User {
	return &User{
		initialState: initialState,
		uuid:         "user:" + uuid.NewString(),
	}
}
