package chat

import (
	"sync"
	"testing"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/now"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger log.NoopLoggerAdapter

type EventLogger struct {
	l     sync.RWMutex
	items map[string][]Event
}

func (l *EventLogger) Log(username string, event Event) {
	l.l.Lock()
	defer l.l.Unlock()
	if l.items == nil {
		l.items = make(map[string][]Event)
	}
	lg := l.items[username]
	lg = append(lg, event)
	l.items[username] = lg
}

func (l *EventLogger) Get(username string) []Event {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.items[username]
}

func TestHubConnectUserUntilClosed(t *testing.T) {
	hub := NewHub(&logger)
	conn := TestConnection{
		EventOutCh: make(chan<- Event),
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	g := test.ErrGroup{}
	g.Go(func() error {
		return hub.ConnectUser("user", &conn)
	})

	conn.Close()
	err := g.WaitTimeout(t)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConnectionClosed)
}

func TestHubConnectUserEvents(t *testing.T) {
	t.Skip("BROKEN")

	now.EnableStub()
	now.ResetStub()

	nowStub := now.CurrStub()
	// Manage time in this test: less realistic but less brittle
	nowStub.Frozen = true
	startTime := nowStub.Time

	t.Cleanup(func() {
		now.DisableStub()
	})

	hub := NewHub(&logger)

	toUser1 := make(chan Event)
	hubConnUser1 := TestConnection{
		EventOutCh: toUser1,
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	toUser2 := make(chan Event)
	hubConnUser2 := TestConnection{
		EventOutCh: toUser2,
		EventInCh:  make(<-chan Event),
		closed:     make(chan struct{}),
	}

	// channels to mark points-in-time
	user1Connected := make(chan struct{})
	user1Closed := make(chan struct{})
	user2Closed := make(chan struct{})

	events := EventLogger{}
	g := test.ErrGroup{}

	g.Go(func() error {
		hub.ConnectUser("user1", &hubConnUser1)
		return nil
	})

	g.Go(func() error {
		<-user1Connected
		hub.ConnectUser("user2", &hubConnUser2)
		return nil
	})

	g.Go(func() error {
		hubConnUser1.WaitClosed()
		close(user1Closed)
		return nil
	})

	g.Go(func() error {
		time.Sleep(time.Second / 2)
		return hubConnUser1.Close()
	})

	g.Go(func() error {
		hubConnUser2.WaitClosed()
		close(user2Closed)
		return nil
	})

	g.Go(func() error {
		for {
			select {
			case <-user1Closed:
				return nil
			case e := <-toUser1:
				// nowStub.Inc()
				events.Log("user1", e)
				switch e.(type) {
				case *EventConnected:
					close(user1Connected)
				}
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-user2Closed:
				return nil
			case e := <-toUser2:
				// nowStub.Inc()
				events.Log("user2", e)
				switch e.(type) {
				case *EventConnected:
					hubConnUser1.Close()
					hubConnUser2.Close()
				}
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
		},
		&EventUserListUpdate{
			EventMeta: EventMeta{Time: startTime},
			Users:     []string{"user1"},
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
		},
		&EventUserListUpdate{
			EventMeta: EventMeta{Time: startTime},
			Users:     []string{"user1", "user2"},
		},
	}

	assert.Equal(t, expectedUser1, events.Get("user1"))
	assert.Equal(t, expectedUser2, events.Get("user2"))
}
