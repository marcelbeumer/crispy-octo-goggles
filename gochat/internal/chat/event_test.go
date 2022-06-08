package chat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/marcelbeumer/go-playground/gochat/internal/util/now"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventMetaWhen tests When() method
func TestEventMetaWhen(t *testing.T) {
	time := time.Now()
	m := EventMeta{Time: time}
	assert.Equal(t, time, m.Time)
}

func TestNewEventMetaNow(t *testing.T) {
	nowStub := now.SetupStub()
	t.Cleanup(func() {
		now.ClearStub()
	})
	m := NewEventMetaNow()
	assert.Equal(t, nowStub.Time, m.Time)
}

func TestEventUserListUpdateJSON(t *testing.T) {
	e := EventUserListUpdate{
		EventMeta: EventMeta{Time: time.UnixMilli(1000)},
		Users:     []string{"u1", "u2"},
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"1970-01-01T01:00:01+01:00","users":["u1","u2"]}`
	assert.Equal(t, expected, string(json))
}

func TestEventUserEnterJSON(t *testing.T) {
	e := EventUserEnter{
		EventMeta: EventMeta{Time: time.UnixMilli(1000)},
		Name:      "u1",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"1970-01-01T01:00:01+01:00","name":"u1"}`
	assert.Equal(t, expected, string(json))
}

func TestEventUserLeaveJSON(t *testing.T) {
	e := EventUserLeave{
		EventMeta: EventMeta{Time: time.UnixMilli(1000)},
		Name:      "u1",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"1970-01-01T01:00:01+01:00","name":"u1"}`
	assert.Equal(t, expected, string(json))
}

func TestEventSendMessageJSON(t *testing.T) {
	e := EventSendMessage{
		EventMeta: EventMeta{Time: time.UnixMilli(1000)},
		Message:   "Hello.",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"1970-01-01T01:00:01+01:00","message":"Hello."}`
	assert.Equal(t, expected, string(json))
}
