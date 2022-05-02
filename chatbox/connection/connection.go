package connection

import (
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type Channels struct {
	ToUser chan message.Message
	ToRoom chan message.Message
}

type Connection struct {
	FromOther <-chan message.Message
	ToOther   chan<- message.Message
}

func NewChannels() *Channels {
	return &Channels{
		ToUser: make(chan message.Message),
		ToRoom: make(chan message.Message),
	}
}

func NewConnectionForUser(c *Channels) Connection {
	fromRoom := (<-chan message.Message)(c.ToUser)
	toRoom := (chan<- message.Message)(c.ToRoom)
	return Connection{
		FromOther: fromRoom,
		ToOther:   toRoom,
	}
}

func NewConnectionForRoom(c *Channels) Connection {
	fromUser := (<-chan message.Message)(c.ToRoom)
	toUser := (chan<- message.Message)(c.ToUser)
	return Connection{
		FromOther: fromUser,
		ToOther:   toUser,
	}
}
