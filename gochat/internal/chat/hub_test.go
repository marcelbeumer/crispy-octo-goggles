package chat

import (
	"testing"

	"github.com/marcelbeumer/go-playground/gochat/internal/util/now"
	"github.com/marcelbeumer/go-playground/gochat/internal/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubConnectUserEvents(t *testing.T) {
	nowStub := now.SetupStub()
	t.Cleanup(func() {
		now.ClearStub()
	})

	nowStub.Frozen = true // less brittle
	startTime := nowStub.Time

	hub := NewHub(test.NewTestLogger(true))

	user1Ch := make(chan Event)
	user2Ch := make(chan Event)
	user1Conn := NewTestConnection(make(chan Event), user1Ch)
	user2Conn := NewTestConnection(make(chan Event), user2Ch)
	user1Events := []Event{}
	user2Events := []Event{}

	didUser1Connect := make(chan struct{})
	canDisconnect := make(chan struct{})
	done := make(chan struct{})

	var g test.ErrGroup

	g.Go(func() error {
		_, err := hub.Connect("user1", user1Conn)
		if err != nil {
			return err
		}
		<-didUser1Connect
		_, err = hub.Connect("user2", user2Conn)
		return err
	})

	g.Go(func() error {
		<-canDisconnect
		defer close(done)
		return hub.Close()
	})

	g.Go(func() error {
		for {
			select {
			case <-done:
				return nil
			case e := <-user1Ch:
				user1Events = append(user1Events, e)
				switch e.(type) {
				case *EventConnected:
					close(didUser1Connect)
				case *EventUserListUpdate:
					close(canDisconnect)
				}
			case e := <-user2Ch:
				user2Events = append(user2Events, e)
			}
		}
	})

	err := g.WaitTimeout(t)
	require.NoError(t, err)

	expectedUser1 := []Event{
		&EventConnected{
			EventMeta: EventMeta{
				Time: startTime,
			},
			Users: []string{"user1"},
		},
		&EventUserEnter{
			EventMeta: EventMeta{Time: startTime},
			Name:      "user2",
		},
		&EventUserListUpdate{
			EventMeta: EventMeta{Time: startTime},
			Users:     []string{"user1", "user2"},
		},
	}

	expectedUser2 := []Event{
		&EventConnected{
			EventMeta: EventMeta{
				Time: startTime,
			},
			Users: []string{"user1", "user2"},
		},
	}

	assert.Equal(t, expectedUser1, user1Events)
	assert.Equal(t, expectedUser2, user2Events)
}
