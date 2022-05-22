package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/marcelbeumer/crispy-octo-goggles/chat"
)

type Message struct {
	Name string     `json:"name"`
	Data chat.Event `json:"data"`
}

type MessageRaw struct {
	Name string           `json:"name"`
	Data *json.RawMessage `json:"data"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	raw := MessageRaw{Data: &json.RawMessage{}}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	handler := handlers[raw.Name]
	if handler == nil {
		return fmt.Errorf(`unknown event name "%s"`, raw.Name)
	}

	m.Name = raw.Name
	m.Data = handler()

	if err := json.Unmarshal(*raw.Data, &m.Data); err != nil {
		return err
	}

	return nil
}
