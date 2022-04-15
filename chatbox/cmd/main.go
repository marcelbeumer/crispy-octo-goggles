package main

import (
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func main() {
	r := room.NewRoom()
	r.Start()

	u := user.NewUser(chatbox.UserState{Name: "Example user", Status: chatbox.StatusBusy})
	u.Start()

	r.AddUser(u)

	go func() {
		time.Sleep(time.Second)
		u.SendMessage("Hello!")
	}()
	r.WaitDone()
}
