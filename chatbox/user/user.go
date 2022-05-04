package user

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type User struct {
	conn   channels.ChannelsOneDir
	stopCh chan struct{}
}

func (u *User) ConnectRoom(conn channels.ChannelsOneDir) error {
	u.DisconnectRoom()
	u.stopCh = make(chan struct{})
	u.conn = conn
	go (func() {
		stopCh := u.stopCh
		fromOther := conn.FromOther
		for {
			select {
			case m := <-fromOther:
				if err := u.handleRoomMessage(m); err != nil {
					log.Println(err)
				}
				break
			case <-stopCh:
				// disconnect, stop
				return
			}
		}
	})()
	return nil
}

func (u *User) DisconnectRoom() {
	if u.stopCh != nil {
		close(u.stopCh)
		u.stopCh = nil
	}
}

func (u *User) SendMessage(s string) error {
	if u.stopCh == nil {
		return errors.New("No connection with room")
	}
	msg, err := message.NewMessage(message.NEW_MESSAGE, s)
	if err != nil {
		return err
	}
	errCh := make(chan error)
	go (func() {
		select {
		case u.conn.ToOther <- msg:
			errCh <- nil
			return
		case <-time.After(time.Second * 2):
			errCh <- errors.New("message to room timed out")
			return
		}
	})()
	return <-errCh
}

func (u *User) handleRoomMessage(m message.Message) error {
	switch m.Name {

	case message.NEW_USER:
		data, err := message.GetData[string](m)
		if err != nil {
			return err
		}
		fmt.Printf(
			"[%s] <<user \"%s\" entered the room>>\n",
			time.Now().Local(),
			data,
		)

	case message.NEW_MESSAGE:
		data, err := message.GetData[message.NewMessageData](m)
		if err != nil {
			return err
		}
		fmt.Printf(
			"[%s %s] >> %s\n",
			data.Time.Local(),
			data.Sender,
			data.Message,
		)
	}
	return nil
}

func NewUser() *User {
	return &User{}
}
