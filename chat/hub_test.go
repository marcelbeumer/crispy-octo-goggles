package chat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/test"
	"github.com/stretchr/testify/assert"
)

var logger log.NoopLoggerAdapter

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

	assert.ErrorIs(t, err, ErrConnectionClosed)
}

func TestHubConnectUserEvents(t *testing.T) {
	hub := NewHub(&logger)

	fromUser1 := make(chan Event)
	toUser1 := make(chan Event)

	fromUser2 := make(chan Event)
	toUser2 := make(chan Event)

	connUser1 := TestConnection{
		EventOutCh: toUser1,
		EventInCh:  fromUser1,
		closed:     make(chan struct{}),
	}

	connUser2 := TestConnection{
		EventOutCh: toUser2,
		EventInCh:  fromUser2,
		closed:     make(chan struct{}),
	}

	user1Connected := make(chan struct{})
	user1Closed := make(chan struct{})
	user2Closed := make(chan struct{})
	var log test.Log
	g := test.ErrGroup{}

	g.Go(func() error {
		connUser1.WaitClosed()
		close(user1Closed)
		return nil
	})

	g.Go(func() error {
		connUser2.WaitClosed()
		close(user2Closed)
		return nil
	})

	g.Go(func() error {
		return hub.ConnectUser("user1", &connUser1)
	})

	g.Go(func() error {
		<-user1Connected
		return hub.ConnectUser("user2", &connUser2)
	})

	g.Go(func() error {
		listUpdateCount := 0
		for {
			select {
			case <-user1Closed:
				return nil
			case event := <-toUser1:
				switch e := event.(type) {
				case *EventUserListUpdate:
					listUpdateCount++
					msg := fmt.Sprintf(
						"user1 <- listupdate #%d = %s",
						listUpdateCount,
						strings.Join(e.Users, ","),
					)
					log.Add(msg)
				case *EventConnected:
					log.Add(fmt.Sprintf("user1 <- connected"))
					close(user1Connected)
				}
			}
		}
	})

	g.Go(func() error {
		userEnterCount := 0
		listUpdateCount := 0
		for {
			select {
			case <-user2Closed:
				return nil
			case event := <-toUser2:
				switch e := event.(type) {
				case *EventUserEnter:
					listUpdateCount++
					msg := fmt.Sprintf(
						"user2 <- userenter #%d = %s",
						userEnterCount,
						e.Name,
					)
					log.Add(msg)

					connUser1.Close()
					connUser2.Close()
				case *EventUserListUpdate:
					listUpdateCount++
					msg := fmt.Sprintf(
						"user2 <- listupdate #%d = %s",
						listUpdateCount,
						strings.Join(e.Users, ","),
					)
					log.Add(msg)
				case *EventConnected:
					log.Add(fmt.Sprintf("user2 <- connected"))
				}
			}
		}
	})

	err := g.WaitTimeout(t)
	assert.ErrorIs(t, err, ErrConnectionClosed)
	expected := []string{
		"user1 <- listupdate #1 = user1",
		"user1 <- connected",
		"user1 <- listupdate #2 = user2,user1",
		"user2 <- connected",
		"user2 <- listupdate #1 = user2,user1",
		"user2 <- userenter #1 = user1",
	}
	actual := log.Items()
	// sort.Strings(expected)
	// sort.Strings(actual)
	assert.Equal(t, expected, actual)

}
