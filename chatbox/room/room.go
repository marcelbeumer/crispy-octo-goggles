package room

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type userInfo struct {
	name   string
	conn   channels.ChannelsOneDir
	l      *sync.RWMutex
	stopCh chan struct{}
}

type userStore struct {
	users map[string]*userInfo
	l     *sync.RWMutex
}

func (u *userStore) Keys() []string {
	keys := make([]string, 0, len(u.users))
	for k := range u.users {
		keys = append(keys, k)
	}
	return keys
}

func (u *userStore) Values() []*userInfo {
	values := make([]*userInfo, 0, len(u.users))
	for _, v := range u.users {
		values = append(values, v)
	}
	return values
}

func (u *userStore) Get(username string) (*userInfo, bool) {
	u.l.RLock()
	i, ok := u.users[username]
	u.l.RUnlock()
	return i, ok
}

func (u *userStore) Set(username string, info *userInfo) {
	u.l.Lock()
	u.users[username] = info
	u.l.Unlock()
}

func (u *userStore) Remove(username string) bool {
	u.l.Lock()
	if u.users[username] == nil {
		return false
	}
	delete(u.users, username)
	u.l.Unlock()
	return true
}

func newUserStore() *userStore {
	return &userStore{
		users: map[string]*userInfo{},
		l:     &sync.RWMutex{},
	}
}

type Room struct {
	users *userStore
	l     *sync.RWMutex
}

func (r *Room) ConnectUser(name string, conn channels.ChannelsOneDir) error {
	if _, ok := r.users.Get(name); ok {
		return fmt.Errorf("user \"%s\" already exists", name)
	}

	user := userInfo{
		name:   name,
		conn:   conn,
		l:      &sync.RWMutex{},
		stopCh: make(chan struct{}),
	}
	r.users.Set(user.name, &user)

	go (func() {
		for {
			select {
			case m := <-user.conn.FromOther:
				if err := r.handleUserMessage(name, m); err != nil {
					log.Println(err)
				}
			case <-user.stopCh:
				// disconnect, stop
				return
			}
		}
	})()

	msgUser, err := message.NewMessage(message.USER_LIST, r.users.Keys())
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
	if user, ok := r.users.Get(name); ok {
		close(user.stopCh)
	}
	_ = r.users.Remove(name)
	return nil
}

func (r *Room) handleUserMessage(name string, m message.Message) error {
	switch m.Name {
	case message.SEND_MESSAGE:
		mData, err := message.GetData[string](m)
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
	for _, user := range r.users.Values() {
		if !except[user.name] {
			r.sendMessageToUser(user.name, msg)
		}
	}
}

func (r *Room) sendMessageToUser(name string, msg message.Message) error {
	user, ok := r.users.Get(name)
	if !ok {
		return fmt.Errorf("can not send message to unknown user %s", name)
	}
	go (func() {
		select {
		case user.conn.ToOther <- msg:
			return
			//
		case <-time.After(time.Second * 2):
			err := fmt.Errorf("message to user \"%s\" timed out", name)
			log.Println(err)
			//
		}
	})()
	return nil
}

func NewRoom() *Room {
	return &Room{
		users: newUserStore(),
		l:     &sync.RWMutex{},
	}
}
