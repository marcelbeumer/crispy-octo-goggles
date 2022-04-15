package world

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcelbeumer/crispy-octo-goggles/racingsim/game"
)

type World struct {
	uuid     string
	ctime    time.Time // current
	ptime    time.Time // previous
	players  []game.Player
	ch       chan game.Message
	done     chan struct{}
	playerCh map[string]chan game.Message
	state    game.WorldState
}

func (w *World) Start() error {
	if w.done != nil {
		return errors.New("already started")
	}
	tick := time.Tick(16 * time.Millisecond)
	w.done = make(chan struct{})
	go (func() {
		for {
			select {
			case <-tick:
				w.Update()
			case msg := <-w.ch:
				fmt.Print(msg)
			case <-w.done:
				break
			}
		}
	})()
	return nil
}

func (w *World) Stop() error {
	if w.done == nil {
		return errors.New("not started")
	}
	close(w.done)
	return nil
}

func (w *World) WaitDone() {
	_ = <-w.done
}

func (w *World) Update() {
	w.ptime = w.ctime
	w.ctime = time.Now()
	// for _, player := range w.players {
	// 	player.In <- WorldUpdateMessage{}
	// }
}

func (w *World) GetPlayer(uuid string) (game.Player, error) {
	for _, player := range w.players {
		if player.Uuid() == uuid {
			return player, nil
		}
	}
	return nil, errors.New("player not found")
}

func (w *World) AddPlayer(p game.Player) error {
	if found, _ := w.GetPlayer(p.Uuid()); found != nil {
		return errors.New("player already added")
	}

	uuid := p.Uuid()
	w.state.Players[uuid] = game.PlayerState{Speed: 0, Loc: 0}
	w.players = append(w.players, p)

	pch := make(chan game.Message)
	w.playerCh[uuid] = pch
	in := (<-chan game.Message)(pch)
	out := (chan<- game.Message)(w.ch)
	p.Chan(&in, &out)

	w.castState()
	return nil
}

func (w *World) castState() {
	msg := game.Message{Sender: w.uuid, Data: w.state}
	for _, ch := range w.playerCh {
		ch <- msg
	}
}

func NewWorld() *World {
	return &World{
		ch: make(chan game.Message),
		state: game.WorldState{
			Players: make(map[string]game.PlayerState),
		},
		playerCh: make(map[string]chan game.Message),
		uuid:     uuid.NewString(),
	}
}
