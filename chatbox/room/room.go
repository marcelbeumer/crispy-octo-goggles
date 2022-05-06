package room

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/mutex"
)

type userInfo struct {
	name string
	conn channels.ChannelsOneDir
	l    *sync.RWMutex
	stop chan bool
	done chan struct{}
}

type Room struct {
	users *mutex.Map[*userInfo]
	l     *sync.RWMutex
}

func (r *Room) Connect(username string, conn channels.ChannelsOneDir) error {
	if _, ok := r.users.Get(username); ok {
		return fmt.Errorf("user \"%s\" already exists", username)
	}

	user := userInfo{
		name: username,
		conn: conn,
		l:    &sync.RWMutex{},
		stop: make(chan bool),
		done: make(chan struct{}),
	}
	go func() { // safe controller for close
		select {
		case <-user.done:
		case <-user.stop:
			close(user.done)
		}
	}()

	r.users.Set(user.name, &user)

	go func() {
		for {
			select {
			case m := <-user.conn.FromOther:
				if err := r.handleMessage(username, m); err != nil {
					log.Println(err)
				}
			case <-user.done:
				// disconnect, stop
				return
			}
		}
	}()

	msgUser, err := message.NewMessage(message.USER_LIST, r.users.Keys())
	if err != nil {
		return err
	}
	if err := r.sendMessage(user.name, msgUser); err != nil {
		return err
	}

	msgAll, err := message.NewMessage(message.NEW_USER, user.name)
	if err != nil {
		return err
	}
	r.broadcastMessage(msgAll, user.name)

	return nil
}

func (r *Room) Disconnect(username string) error {
	if user, ok := r.users.Get(username); ok {
		user.stop <- true
	}
	_ = r.users.Remove(username)
	return nil
}

func (r *Room) handleMessage(username string, m message.Message) error {
	switch m.Name {
	case message.SEND_MESSAGE:
		mData, err := message.GetData[string](m)
		if err != nil {
			return err
		}
		mAllData := message.NewMessageData{
			Sender:  username,
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
	for _, user := range r.users.Values() {
		if !except[user.name] {
			r.sendMessage(user.name, msg)
		}
	}
}

func (r *Room) sendMessage(username string, msg message.Message) error {
	user, ok := r.users.Get(username)
	if !ok {
		return fmt.Errorf("can not send message to unknown user %s", username)
	}
	go (func() {
		select {
		case user.conn.ToOther <- msg:
			return
			//
		case <-time.After(time.Second * 2):
			err := fmt.Errorf("message to user \"%s\" timed out", username)
			log.Println(err)
			//
		}
	})()
	return nil
}

func NewRoom() *Room {
	return &Room{
		users: mutex.NewMap[*userInfo](),
		l:     &sync.RWMutex{},
	}
}
