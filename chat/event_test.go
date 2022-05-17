package chat

import (
	"encoding/json"
	"testing"
	"time"

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
	oldNow := now
	nowTime := time.Now()
	now = func() time.Time {
		return nowTime
	}
	t.Cleanup(func() {
		now = oldNow
	})
	m := NewEventMetaNow()
	assert.Equal(t, nowTime, m.Time)
}

func TestEventUserListUpdateJSON(t *testing.T) {
	e := EventUserListUpdate{
		EventMeta: EventMeta{Time: time.Time{}},
		Users:     []string{"u1", "u2"},
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"0001-01-01T00:00:00Z","users":["u1","u2"]}`
	assert.Equal(t, expected, string(json))
}

func TestEventUserEnterJSON(t *testing.T) {
	e := EventUserEnter{
		EventMeta: EventMeta{Time: time.Time{}},
		Name:      "u1",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"0001-01-01T00:00:00Z","name":"u1"}`
	assert.Equal(t, expected, string(json))
}

func TestEventUserLeaveJSON(t *testing.T) {
	e := EventUserLeave{
		EventMeta: EventMeta{Time: time.Time{}},
		Name:      "u1",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"0001-01-01T00:00:00Z","name":"u1"}`
	assert.Equal(t, expected, string(json))
}

func TestEventSendMessageJSON(t *testing.T) {
	e := EventSendMessage{
		EventMeta: EventMeta{Time: time.Time{}},
		Message:   "Hello.",
	}
	json, err := json.Marshal(&e)
	require.NoError(t, err)
	expected := `{"time":"0001-01-01T00:00:00Z","message":"Hello."}`
	assert.Equal(t, expected, string(json))
}
