package player

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/racingsim/game"
)

type Player struct {
	uuid  string
	ctime time.Time // current
	ptime time.Time // previous
	in    *<-chan game.Message
	out   *chan<- game.Message
	done  chan struct{}
	world game.WorldState
}

func (p *Player) Start() error {
	if p.done != nil {
		return errors.New("already started")
	}
	tick := time.Tick(16 * time.Millisecond)
	p.done = make(chan struct{})
	go (func() {
		for {
			select {
			case <-tick:
				p.Update()
			case m := <-*p.in:
				p.message(m)
			case <-p.done:
				break
			}
		}
	})()
	return nil
}

func (p *Player) Stop() error {
	if p.done == nil {
		return errors.New("not started")
	}
	close(p.done)
	return nil
}

func (p *Player) WaitDone() {
	_ = <-p.done
}

func (p *Player) Update() {
	p.ptime = p.ctime
	p.ctime = time.Now()
	// for _, player := range w.players {
	// 	player.In <- PlayerUpdateMessage{}
	// }
}

func (p *Player) Uuid() string {
	return p.uuid
}

func (p *Player) Chan(in *<-chan game.Message, out *chan<- game.Message) {
	p.in = in
	p.out = out
}

func (p *Player) message(m game.Message) error {
	switch t := m.Data.(type) {
	case game.WorldState:
		fmt.Printf("player %s setting world state\n", p.uuid)
		p.world = t
	}

	return nil
}

func NewPlayer() *Player {
	return &Player{
		uuid: uuid.NewString(),
	}
}
