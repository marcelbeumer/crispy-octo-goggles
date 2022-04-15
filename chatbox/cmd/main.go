package main

import (
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func main() {
	room := room.NewRoom()
	room.Start()

	go func() {
		voyeur := user.NewUser(chatbox.UserState{Name: "Voyeur", Status: chatbox.StatusOffline})
		voyeur.CanPrint = true
		if err := voyeur.Start(); err != nil {
			panic(err)
		}
		if err := room.AddUser(voyeur); err != nil {
			panic(err)
		}

		john := user.NewUser(chatbox.UserState{Name: "John", Status: chatbox.StatusOnline})
		if err := john.Start(); err != nil {
			panic(err)
		}
		if err := room.AddUser(john); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)

		john.SendMessage("Hello empty room!")

		time.Sleep(time.Second)

		kyle := user.NewUser(chatbox.UserState{Name: "Kyle", Status: chatbox.StatusOnline})
		if err := kyle.Start(); err != nil {
			panic(err)
		}
		if err := room.AddUser(kyle); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)

		kyle.SendMessage("Hello John")
		john.SendMessage("Hi Kyle")
	}()

	room.WaitDone()
}
