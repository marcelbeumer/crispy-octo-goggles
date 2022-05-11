package chatbox

import "time"

type Event interface {
	// When returns time of the event.
	// Important we define *something* more than interface{}
	// for static analysis to work on *Struct{} vs Struct{}
	When() time.Time
}

type EventMeta struct {
	time time.Time
}

func (e *EventMeta) When() time.Time {
	return e.time
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
		time: time.Now(),
	}
}
