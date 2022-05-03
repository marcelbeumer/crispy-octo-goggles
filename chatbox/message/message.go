package message

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type MessageName string

const (
	USER_LIST    = "userList"
	NEW_USER     = "newUser"
	SEND_MESSAGE = "sendMessage"
	NEW_MESSAGE  = "newMessage"
)

type Message struct {
	Name MessageName `json:"name"`
	Data any         `json:"data"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type JsonMessage struct {
		Name MessageName
		Data json.RawMessage
	}
	result := JsonMessage{Data: json.RawMessage{}}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	m.Name = result.Name
	switch m.Name {
	case USER_LIST:
		m.Data = new(string)
	case NEW_USER:
		m.Data = new(string)
	case SEND_MESSAGE:
		m.Data = new(string)
	case NEW_MESSAGE:
		m.Data = NewMessageData{}
	default:
		return fmt.Errorf("unknown message name \"%s\"", m.Name)
	}
	if err := json.Unmarshal(result.Data, &m.Data); err != nil {
		return err
	}
	return nil
}

func NewMessage(name MessageName, data any) (Message, error) {
	msg := Message{Name: name, Data: data}
	return msg, ValidateMessage(&msg)
}

func ValidateMessage(m *Message) error {
	switch m.Name {
	case USER_LIST:
		_, err := GetDataType[[]string](m.Data)
		return err
	case NEW_USER:
		_, err := GetDataType[string](m.Data)
		return err
	case SEND_MESSAGE:
		_, err := GetDataType[string](m.Data)
		return err
	case NEW_MESSAGE:
		_, err := GetDataType[NewMessageData](m.Data)
		return err
	default:
		return fmt.Errorf("unknown message name \"%s\"", m.Name)
	}
}

type UserListData []string

type NewUserData string

type MessageData string

type NewMessageData struct {
	Sender  string    `json:"sender"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
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

func GetDataType[T any](v any) (*T, error) {
	// FIXME: this type conversion seems to copy the data instead of reference oly
	// maybe this is how Go works (only), but not sure
	data, ok := v.(*T)
	if !ok {
		empty := *new(T)
		return &empty, NewDataTypeError(v, empty)
	}
	return data, nil
}

func GetData[T any](m *Message) (*T, error) {
	empty := *new(T)
	err := ValidateMessage(&Message{Name: m.Name, Data: empty})
	if err != nil {
		return &empty, fmt.Errorf("invalid request for data: %w", err)
	}
	return GetDataType[T](&m.Data)
}
