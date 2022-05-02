package message

import (
	"fmt"
	"reflect"
	"time"
)

type MessageName string

const (
	USER_LIST    = "userList"
	NEW_USER     = "newUser"
	SEND_MESSAGE = "sendMesage"
	NEW_MESSAGE  = "newMessage"
)

type Message struct {
	Name MessageName
	Data any
}

func NewMessage(name MessageName, data any) (Message, error) {
	msg := Message{Name: name, Data: data}
	return msg, ValidateMessage(msg)
}

func ValidateMessage(m Message) error {
	switch m.Name {
	case USER_LIST:
		_, err := getDataType[[]string](m.Data)
		return err
	case NEW_USER:
		_, err := getDataType[string](m.Data)
		return err
	case SEND_MESSAGE:
		_, err := getDataType[string](m.Data)
		return err
	case NEW_MESSAGE:
		_, err := getDataType[NewMessageData](m.Data)
		return err
	default:
		return fmt.Errorf("unknown message name \"%s\"", m.Name)
	}
}

type UserListData []string

type NewUserData string

type MessageData string

type NewMessageData struct {
	Sender  string
	Message string
	Time    time.Time
}

type DataTypeError struct {
	Expected reflect.Type
	Actual   reflect.Type
}

func (e DataTypeError) Error() string {
	return fmt.Sprintf(
		"expected data type %s but got %s",
		e.Expected,
		e.Actual,
	)
}

func NewDataTypeError(v any, expected any) DataTypeError {
	return DataTypeError{
		Expected: reflect.TypeOf(expected),
		Actual:   reflect.TypeOf(v),
	}
}

func getDataType[T any](v any) (T, error) {
	empty := *new(T)
	data, ok := v.(T)
	if !ok {
		return empty, NewDataTypeError(v, empty)
	}
	return data, nil
}

func GetData[T any](m Message) (T, error) {
	empty := *new(T)
	err := ValidateMessage(Message{Name: m.Name, Data: empty})
	if err != nil {
		return empty, fmt.Errorf("invalid request for data: %w", err)
	}
	return getDataType[T](m.Data)
}
