package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/base"
)

type User struct {
	name     string
	uuid     string
	in       <-chan base.Event
	out      chan<- base.Event
	doneCh   chan struct{}
	room     base.RoomState
	CanPrint bool
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

func (u *User) Name() string {
	return u.name
}

func (u *User) Connect(in <-chan base.Event, out chan<- base.Event) error {
	u.in = in
	u.out = out
	e, err := base.NewEvent(base.Connect, base.UserState{
		Name:   u.name,
		Status: base.StatusOnline,
	}, u.uuid)
	if err != nil {
		return err
	}
	u.sendEvent(e)
	return nil
}

func (u *User) SendMessage(m string) {
	e, err := base.NewEvent(base.SendMessage, m, u.uuid)
	if err != nil {
		panic(err)
	}
	go u.sendEvent(e)
}

func (u *User) handleEvent(e base.Event) error {
	switch e.Name {

	case base.RoomStateUpdate:
		data, err := base.GetData[base.RoomState](e)
		if err != nil {
			return err
		}
		u.room = data

	case base.NewUser:
		data, err := base.GetData[base.UserRef](e)
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

	case base.NewMessage:
		data, err := base.GetData[base.Message](e)
		if err != nil {
			return err
		}
		u.printf("[%s %s] %s\n", time.Now().Local(), data.SenderName, data.Message)

	default:
		return base.UhandledEventError{Event: e, Receiver: u.uuid}
	}

	return nil
}

func (u *User) sendEvent(e base.Event) {
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

func NewUser(name string, canPrint bool) *User {
	u := &User{
		name:     name,
		CanPrint: canPrint,
		uuid:     "user:" + uuid.NewString(),
	}
	u.Start()
	return u
}
