package main

import (
	"fmt"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func main() {
	room := room.NewRoom()
	room.Start()

	go func() {
		admin := user.NewUser("Admin")
		admin.CanPrint = true
		admin.Start()

		john := user.NewUser("John")
		john.Start()

		kyle := user.NewUser("Kyle")
		kyle.Start()

		if err := room.AddUser(admin); err != nil {
			panic(err)
		}
		if err := room.AddUser(john); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
		john.SendMessage("Hello empty room!")
		time.Sleep(time.Second)

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

		room.Stop()
	}()

	room.Wait()
}
