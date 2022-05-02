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

func NewConnectionForUser(c Channels) Connection {
	return Connection{
		FromOther: make(<-chan message.Message),
		ToOther:   make(chan<- message.Message),
	}
}

func NewConnectionForRoom(c Channels) Connection {
	return Connection{
		FromOther: make(<-chan message.Message),
		ToOther:   make(chan<- message.Message),
	}
}
