package main

import (
	"github.com/marcelbeumer/crispy-octo-goggles/racingsim/player"
	"github.com/marcelbeumer/crispy-octo-goggles/racingsim/world"
)

func main() {
	w := world.NewWorld()
	w.Start()
	p := player.NewPlayer()
	p.Start()
	w.AddPlayer(p)
	w.WaitDone()
}
