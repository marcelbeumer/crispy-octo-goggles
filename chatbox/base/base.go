package base

import (
	"fmt"
	"reflect"
)

type Message struct {
	Sender     string
	SenderName string
	Message    string
}

type RoomState struct {
	Users    map[string]UserState
	Messages []Message
}

type Event struct {
	Sender string
	Name   EventName
	Data   any
}

type EventFn func() (Event, error)

func NewEvent[T any](name EventName, data T, sender string) (Event, error) {
	e := Event{Sender: sender, Name: name, Data: data}
	switch name {

	case Connect:
		if _, err := GetData[UserState](e); err != nil {
			return e, err
		}

	case RoomStateUpdate:
		if _, err := GetData[RoomState](e); err != nil {
			return e, err
		}

	case NewUser:
		if _, err := GetData[UserRef](e); err != nil {
			return e, err
		}

	case SendMessage:
		if _, err := GetData[string](e); err != nil {
			return e, err
		}

	case NewMessage:
		if _, err := GetData[Message](e); err != nil {
			return e, err
		}

	default:
		return e, UnknownEventError{Event: e}
	}

	return e, nil
}

type EventName string

const (
	Connect         = "Connect"
	RoomStateUpdate = "RoomStateUpdate"
	NewUser         = "NewUser"
	SendMessage     = "SendMessage"
	NewMessage      = "NewMessage"
)

type Status int

const (
	StatusOffline = iota
	StatusOnline
	StatusBusy
)

type UserState struct {
	Name   string
	Status Status
}

type UserRef struct {
	Uuid  string
	State UserState
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

type UnknownEventError struct {
	Event Event
}

func (e UnknownEventError) Error() string {
	return fmt.Sprintf(
		"unknown event \"%s\"",
		e.Event.Name,
	)
}

type UhandledEventError struct {
	Receiver string
	Event    Event
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
