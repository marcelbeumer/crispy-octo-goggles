package main

import (
	"fmt"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func main() {
	room := room.NewRoom()
	room.Start()

	go func() {
		admin := user.NewUser(chatbox.UserState{Name: "Admin", Status: chatbox.StatusBusy})
		admin.CanPrint = true
		if err := admin.Start(); err != nil {
			panic(err)
		}
		if err := room.AddUser(admin); err != nil {
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

		for x := 3; x > 0; x-- {
			msg := fmt.Sprintf("I've heard the room will self destruct in %d...", x)
			admin.SendMessage(msg)
			time.Sleep(time.Second)
		}

		// admin.WaitDone()
		// kyle.WaitDone()
		// john.WaitDone()
		room.Stop()
	}()

	room.WaitDone()
}
