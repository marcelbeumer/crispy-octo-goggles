package channel

import "sync"

type Multiplexer[T any] struct {
	source  chan T
	targets []*chan T
	l       sync.RWMutex
	doneCh  chan struct{}
}

func (e *Multiplexer[T]) Start() chan struct{} {
	e.doneCh = make(chan struct{})
	go (func() {
		for {
			select {
			case <-e.doneCh:
				e.doneCh = nil
				return
			case v, ok := <-e.source:
				if e.doneCh == nil {
					return
				}
				targets := e.targets
				if !ok {
					close(e.doneCh)
					e.targets = []*chan T{}
					for _, target := range targets {
						target := target
						go (func() {
							close(*target)
						})()
					}
					return
				}
				for _, target := range targets {
					target := target
					go (func() {
						*target <- v
					})()
				}
			}
		}
	})()
	return e.doneCh
}

func (e *Multiplexer[T]) Stop() {
	if e.doneCh != nil {
		close(e.doneCh)
	}
}

func (e *Multiplexer[T]) Add() *<-chan T {
	e.l.Lock()
	c := make(chan T)
	e.targets = append(e.targets, &c)
	e.l.Unlock()
	rc := (<-chan T)(c)
	return &rc
}

func (e *Multiplexer[T]) Remove(ch *<-chan T) {
	e.l.Lock()
	newTargets := []*chan T{}
	for _, target := range e.targets {
		t := (<-chan T)(*target)
		if ch != &t {
			newTargets = append(newTargets, target)
		}
	}
	e.targets = newTargets
	e.l.Unlock()
}

func NewMultiplexer[T any](source chan T) *Multiplexer[T] {
	e := Multiplexer[T]{
		source:  source,
		targets: []*chan T{},
		l:       sync.RWMutex{},
	}
	e.Start()
	return &e
}
