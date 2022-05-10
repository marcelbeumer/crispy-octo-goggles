package chatbox

import "time"

type Event interface{}

type EventMeta struct {
	Time time.Time
}

type EventUserListUpdate struct {
	EventMeta
	Users []string
}

type EventNewUser struct {
	EventMeta
	Name string
}

type EventSendMessage struct {
	EventMeta
	Message string
}

type EventNewMessage struct {
	EventMeta
	Sender  string
	Message string
}

func NewEventMetaNow() *EventMeta {
	return &EventMeta{
		Time: time.Now(),
	}
}
