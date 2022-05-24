package queue

import (
	"errors"
	"sync"
)

// Error returned when adding items when the queue is closed
var ErrClosed = errors.New("queue closed")

// Error returned when the queue is closed and there are no items left.
var ErrEmpty = errors.New("queue closed and empty")

// Queue is a thread-safe, slice-based queue
type Queue[T any] struct {
	items  []T
	mu     sync.RWMutex
	wait   chan struct{}
	closed chan struct{}
	empty  chan struct{}
}

// Add adds items to the queue.
// Returns ErrClosed if queue is closed.
func (q *Queue[T]) Add(e T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.closed:
		return ErrClosed
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

// Close closes the queue.
// Pending reads will still continue until the queue is empty.
func (q *Queue[T]) Close() error {
	select {
	case <-q.closed:
		return ErrClosed
	default:
		close(q.closed)
		q.mu.Lock()
		if len(q.items) == 0 {
			select {
			case <-q.empty:
			default:
				close(q.empty)
			}
		}
		q.mu.Unlock()
		return nil
	}
}

// Read returns and removes last item from the queue.
// Waits when there are no items yet.
// Returns ErrEmpty when queue is closed and no items are left.
func (q *Queue[T]) Read() (T, error) {
	for {
		q.mu.Lock()
		if q.wait == nil {
			q.wait = make(chan struct{})
		}

		if len(q.items) > 0 {
			item := q.items[0]
			q.items = q.items[1:]
			if len(q.items) == 0 {
				select {
				case <-q.empty:
				case <-q.closed:
					close(q.empty)
				default:
					//
				}
			}
			q.mu.Unlock()
			return item, nil
		}

		q.mu.Unlock()

		select {
		case <-q.wait:
		case <-q.empty:
			var zero T
			return zero, ErrEmpty
		}
	}
}

// Returns closed channel when queue is closed and empty.
func (q *Queue[T]) Empty() <-chan struct{} {
	return q.empty
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items:  []T{},
		mu:     sync.RWMutex{},
		wait:   make(chan struct{}),
		empty:  make(chan struct{}),
		closed: make(chan struct{}),
	}
}
