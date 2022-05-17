package safe

import (
	"errors"
	"sync"
)

var ErrQueueNoItem = errors.New("no last queue item")
var ErrQueueClosed = errors.New("queue closed")

// Queue is a thread-safe, slice-based queue
type Queue[T any] struct {
	items  []T
	l      sync.RWMutex
	wait   chan struct{}
	closed chan struct{}
}

// Add adds items to the queue.
// Returns error if queue is closed
func (q *Queue[T]) Add(e T) error {
	select {
	case <-q.closed:
		return ErrQueueClosed
	default:
	}

	q.l.Lock()
	q.items = append(q.items, e)

	// FIXME: waiting for the go routine to start to unlock seems
	// not right (slow)
	if len(q.wait) == 0 {
		go func() {
			q.l.Unlock()
			q.wait <- struct{}{}
		}()
		return nil
	}

	q.l.Unlock()
	return nil
}

// Close closes the queue and clears internal data structures.
// Pending reads will return error
func (q *Queue[T]) Close() {
	q.l.Lock()
	defer q.l.Unlock()

	select {
	case <-q.closed:
	default:
		q.items = nil
		close(q.closed)
	}
}

// Read returns and removes last item from the queue.
// Waits when there are no items.
// Returns ErrQueueClosed when queue is closed.
func (q *Queue[T]) Read() (T, error) {
	for {
		lastEvent, err := q.getLast()
		if err == nil {
			return lastEvent, nil
		}

		select {
		case <-q.wait:
		case <-q.closed:
			var zero T
			return zero, ErrQueueClosed
		}
	}
}

func (q *Queue[T]) getLast() (T, error) {
	q.l.Lock()
	defer q.l.Unlock()
	if len(q.items) > 0 {
		end := len(q.items) - 1
		event := q.items[end]
		q.items = q.items[:end]
		return event, nil
	}
	var empty T
	return empty, ErrQueueNoItem
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items:  []T{},
		l:      sync.RWMutex{},
		wait:   make(chan struct{}, 1),
		closed: make(chan struct{}),
	}
}
