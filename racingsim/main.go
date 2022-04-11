package main

import (
	"time"
)

type WorldTime struct {
}

type World struct {
	ctime time.Time // current
	ptime time.Time // previous
}

func (w World) Start() {
	tick := time.Tick(16 * time.Millisecond)
	for {
		select {
		case <-tick:
			w.Update()
		}
	}
}

func (w World) Update() {
	w.ptime = w.ctime
	w.ctime = time.Now()
}

func main() {
	World{}.Start()
}
