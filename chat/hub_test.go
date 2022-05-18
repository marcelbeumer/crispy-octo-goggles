package chat

import (
	"testing"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/util/now"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger log.NoopLoggerAdapter

func TestHubConnectUserEvents(t *testing.T) {
	now.EnableStub()
	now.ResetStub()
	t.Cleanup(func() {
		now.DisableStub()
	})

	nowStub := now.CurrStub()
	nowStub.Frozen = true // less brittle
	startTime := nowStub.Time

	hub := NewHub(&logger)

	user1Ch := make(chan Event)
	user2Ch := make(chan Event)
	user1Conn := NewTestConnection(make(chan Event), user1Ch)
	user2Conn := NewTestConnection(make(chan Event), user2Ch)
	user1Events := []Event{}
	user2Events := []Event{}

	var user1Id hubId
	var user2Id hubId
	var g test.ErrGroup

	didUser1Connect := make(chan struct{})
	canDisconnect := make(chan struct{})
	done := make(chan struct{})

	g.Go(func() error {
		id, err := hub.Connect("user1", user1Conn)
		user1Id = id
		<-didUser1Connect
		id, err = hub.Connect("user2", user2Conn)
		user2Id = id
		return err
	})

	g.Go(func() error {
		<-canDisconnect
		defer close(done)
		if err := hub.Disconnect(user1Id); err != nil {
			return err
		}
		if err := hub.Disconnect(user2Id); err != nil {
			return err
		}
		return nil
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
