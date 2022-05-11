package chat

import "time"

type Event interface {
	// When returns time of the event.
	// Important we define *something* more than interface{}
	// for static analysis to work on *Struct{} vs Struct{}
	When() time.Time
}

type EventMeta struct {
	Time time.Time `json:"time"`
}

func (e *EventMeta) When() time.Time {
	return e.Time
}

type EventUserListUpdate struct {
	EventMeta
	Users []string `json:"users"`
}

type EventNewUser struct {
	EventMeta
	Name string `json:"name"`
}

type EventSendMessage struct {
	EventMeta
	Message string `json:"message"`
}

type EventNewMessage struct {
	EventMeta
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

func NewEventMetaNow() *EventMeta {
	return &EventMeta{
		Time: time.Now(),
	}
}
