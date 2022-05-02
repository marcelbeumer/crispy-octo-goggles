package room

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/connection"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type userInfo struct {
	name   string
	conn   connection.Connection
	l      *sync.RWMutex
	stopCh chan struct{}
}

type Room struct {
	users []userInfo
	l     *sync.RWMutex
}

func (r *Room) ConnectUser(name string, conn connection.Connection) error {
	if _, err := r.getUser(name); err == nil {
		return fmt.Errorf("user \"%s\" already exists", name)
	}

	r.l.Lock()
	user := userInfo{
		name:   name,
		conn:   conn,
		l:      &sync.RWMutex{},
		stopCh: make(chan struct{}),
	}
	r.users = append(r.users, user)
	usernames := make([]string, len(r.users))
	for i, u := range r.users {
		usernames[i] = u.name
	}
	r.l.Unlock()

	go (func() {
		for {
			select {
			case m := <-user.conn.FromOther:
				if err := r.handleUserMessage(name, m); err != nil {
					log.Println(err)
				}
				break
			case <-user.stopCh:
				// disconnect, stop
				return
			}
		}
	})()

	msgUser, err := message.NewMessage(message.USER_LIST, usernames)
	if err != nil {
		return err
	}
	if err := r.sendMessageToUser(user.name, msgUser); err != nil {
		return err
	}

	msgAll, err := message.NewMessage(message.NEW_USER, user.name)
	if err != nil {
		return err
	}
	r.broadcastMessage(msgAll, user.name)

	return nil
}

func (r *Room) DisconnectUser(name string) error {
	user, err := r.getUser(name)
	if err != nil {
		return err
	}
	close(user.stopCh)
	return nil
}

func (r *Room) handleUserMessage(name string, m message.Message) error {
	switch m.Name {
	case message.SEND_MESSAGE:
		mData, err := message.GetMessageData[message.MessageData](m)
		if err != nil {
			return err
		}
		mAllData := message.NewMessageData{
			Sender:  name,
			Message: string(mData),
			Time:    time.Now(),
		}
		mAll, err := message.NewMessage(message.NEW_MESSAGE, mAllData)
		if err != nil {
			return err
		}
		r.broadcastMessage(mAll)
	}
	return nil
}

func (r *Room) broadcastMessage(msg message.Message, exceptNames ...string) {
	except := map[string]bool{}
	for _, n := range exceptNames {
		except[n] = true
	}
	for _, user := range r.users {
		if !except[user.name] {
			if err := r.sendMessageToUser(user.name, msg); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (r *Room) sendMessageToUser(name string, msg message.Message) error {
	user, err := r.getUser(name)
	if err != nil {
		return err
	}
	errChan := make(chan error)
	go (func() {
		select {
		case user.conn.ToOther <- msg:
			errChan <- nil
			return
			//
		case <-time.After(time.Second * 2):
			errChan <- fmt.Errorf("message to user \"%s\" timed out", name)
			return
			//
		}
	})()
	return <-errChan
}

func (r *Room) getUser(name string) (*userInfo, error) {
	for _, user := range r.users {
		if user.name == name {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user \"%s\" does not exists", name)
}

func NewRoom() *Room {
	return &Room{
		users: []userInfo{},
		l:     &sync.RWMutex{},
	}
}