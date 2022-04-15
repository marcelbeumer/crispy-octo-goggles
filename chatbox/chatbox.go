package chatbox

import (
	"fmt"
	"reflect"
)

type Room interface{}

type Message struct {
	Sender string
	SenderName string
	Message string
}

type RoomState struct {
	Users map[string]UserState
	Messages []Message
}

type Event struct {
	Sender string
	Name   EventName
	Data   any
}

type EventName string

const (
	RequestInitialUserState = "RequestInitialUserState"
	InitialUserState        = "InitialUserState"
	RoomStateUpdate         = "RoomStateUpdate"
	SendMessage				= "SendMessage"
	NewMessage				= "NewMessage"
)

type Status int

const (
	StatusOffline = iota + 1
	StatusOnline
	StatusBusy
)

type UserState struct {
	Name   string
	Status Status
}

type User interface {
	Uuid() string
	Chan(in *<-chan Event, out *chan<- Event)
}

type EventError struct {
	Event   Event
	Message string
}

func (e EventError) Error() string {
	return fmt.Sprintf(
		"error for event \"%s\": %s",
		e.Event.Name,
		e.Message,
	)
}

func NewEventError(event Event, message string) EventError {
	return EventError{
		Event:   event,
		Message: message,
	}
}

type UhandledEventError struct {
	Receiver string
	Event   Event
}

func (e UhandledEventError) Error() string {
	return fmt.Sprintf(
		"%s could not handle event \"%s\"",
		e.Receiver,
		e.Event.Name,
	)
}

type DataTypeError struct {
	Name     EventName
	Expected reflect.Type
	Actual   reflect.Type
}

func (e DataTypeError) Error() string {
	return fmt.Sprintf(
		"'%s' has wrong data type: expected %s but got %s",
		e.Name,
		e.Expected,
		e.Actual,
	)
}

func NewDataTypeError(event Event, expected any) DataTypeError {
	return DataTypeError{
		Name:     event.Name,
		Expected: reflect.TypeOf(expected),
		Actual:   reflect.TypeOf(event.Data),
	}
}

func GetData[T any](e Event) (T, error) {
	if data, ok := e.Data.(T); ok {
		return data, nil
	} else {
		return data, NewDataTypeError(e, new(T))
	}
}

