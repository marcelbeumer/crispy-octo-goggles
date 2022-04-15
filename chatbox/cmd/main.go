package main

import (
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func main() {
	r := room.NewRoom()
	r.Start()

	p := user.NewUser(chatbox.UserState{Name: "Example user", Status: chatbox.StatusBusy})
	p.Start()

	r.AddUser(p)

	// go func() {
	// 	time.Sleep(time.Second)
	// 	fmt.Println("xx")
	// }()
	r.WaitDone()
}
