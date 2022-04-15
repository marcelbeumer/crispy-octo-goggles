package chatbox

import (
	"fmt"
	"reflect"
)

type Room interface{}

type RoomState struct {
	Users map[string]UserState
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
	event   Event
	message string
}

func (e EventError) Error() string {
	return fmt.Sprintf(
		"error for event '%s': %s",
		e.event.Name,
		e.message,
	)
}

func NewEventError(event Event, message string) EventError {
	return EventError{
		event:   event,
		message: message,
	}
}

type DataTypeError struct {
	name     EventName
	expected reflect.Type
	actual   reflect.Type
}

func (e DataTypeError) Error() string {
	return fmt.Sprintf(
		"'%s' has wrong data type: expected %s but got %s",
		e.name,
		e.expected,
		e.actual,
	)
}

func NewDataTypeError(event Event, expected any) DataTypeError {
	return DataTypeError{
		name:     event.Name,
		expected: reflect.TypeOf(expected),
		actual:   reflect.TypeOf(event.Data),
	}
}

func GetData[T any](e Event) (T, error) {
	if data, ok := e.Data.(T); ok {
		return data, nil
	} else {
		return data, NewDataTypeError(e, new(T))
	}
}

