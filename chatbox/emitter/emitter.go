package emitter

import "sync"

type Emitter struct {
	source  InChannel
	targets []*OutChannel
	l       sync.RWMutex
}

func (e *Emitter) Start() chan struct{} {
	done := make(chan struct{})
	go (func() {
		for {
			select {
			case v, ok := <-e.source.In():
				targets := e.targets
				if !ok {
					close(done)
					e.targets = []*OutChannel{}
					for _, target := range targets {
						(*target).Close()
					}
					return
				}
				for _, target := range targets {
					(*target).Out() <- v
				}
			}
		}
	})()
	return done
}

func (e *Emitter) Add(out *OutChannel) {
	e.l.Lock()
	e.targets = append(e.targets, out)
	e.l.Unlock()
}

func (e *Emitter) Remove(out *OutChannel) {
	e.l.Lock()
	newTargets := []*OutChannel{}
	for _, target := range e.targets {
		if out != target {
			newTargets = append(newTargets, target)
		}
	}
	e.targets = newTargets
	e.l.Unlock()
}

func New(source InChannel) *Emitter {
	e := Emitter{
		source:  source,
		targets: []*OutChannel{},
		l:       sync.RWMutex{},
	}
	return &e
}

type OutChannel interface {
	Out() chan any
	Close()
}

type InChannel interface {
	In() chan any
}
