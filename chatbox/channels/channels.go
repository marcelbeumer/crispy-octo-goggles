package channels

import (
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type Channels struct {
	ToUser chan message.Message
	ToRoom chan message.Message
}

type ChannelsOneDir struct {
	FromOther <-chan message.Message
	ToOther   chan<- message.Message
}

func NewChannels() *Channels {
	return &Channels{
		ToUser: make(chan message.Message),
		ToRoom: make(chan message.Message),
	}
}

func NewChannelsForUser(c *Channels) ChannelsOneDir {
	fromRoom := (<-chan message.Message)(c.ToUser)
	toRoom := (chan<- message.Message)(c.ToRoom)
	return ChannelsOneDir{
		FromOther: fromRoom,
		ToOther:   toRoom,
	}
}

func NewChannelsForRoom(c *Channels) ChannelsOneDir {
	fromUser := (<-chan message.Message)(c.ToRoom)
	toUser := (chan<- message.Message)(c.ToUser)
	return ChannelsOneDir{
		FromOther: fromUser,
		ToOther:   toUser,
	}
}
