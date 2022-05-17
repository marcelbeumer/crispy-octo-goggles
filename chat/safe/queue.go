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
	q.l.Lock()
	defer q.l.Unlock()

	select {
	case <-q.closed:
		return ErrQueueClosed
	default:
	}

	q.items = append(q.items, e)

	if q.wait != nil {
		wait := q.wait
		q.wait = nil

		select {
		case <-wait:
		default:
			close(wait)
		}
	}

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
		q.l.Lock()
		if len(q.items) > 0 {
			end := len(q.items) - 1
			event := q.items[end]
			q.items = q.items[:end]
			q.l.Unlock()
			return event, nil
		}

		if q.wait == nil {
			q.wait = make(chan struct{})
		}

		wait := q.wait
		q.l.Unlock()

		select {
		case <-wait:
		case <-q.closed:
			var zero T
			return zero, ErrQueueClosed
		}
	}
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items:  []T{},
		l:      sync.RWMutex{},
		wait:   make(chan struct{}),
		closed: make(chan struct{}),
	}
}
