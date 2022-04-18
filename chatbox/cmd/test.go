package cmd

import (
	"github.com/spf13/cobra"

	"fmt"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/connect"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func init() {
	rootCmd.AddCommand(testCmd)
}


var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run test scenario",
	Run: func(cmd *cobra.Command, args []string) {
		room := room.NewRoom()

		go func() {
			admin := user.NewUser("Admin", true)
			if err := connect.ConnectUser(room, admin); err != nil {
				panic(err)
			}

			john := user.NewUser("John", false)
			if err := connect.ConnectUser(room, john); err != nil {
				panic(err)
			}

			kyle := user.NewUser("Kyle", false)

			time.Sleep(time.Second)
			john.SendMessage("Hello empty room!")
			time.Sleep(time.Second)

			if err := connect.ConnectUser(room, kyle); err != nil {
				panic(err)
			}

			time.Sleep(time.Second)

			kyle.SendMessage("Hello John")
			john.SendMessage("Hi Kyle")

			for x := 3; x > 0; x-- {
				msg := fmt.Sprintf("Room will self destruct in %d...", x)
				admin.SendMessage(msg)
				time.Sleep(time.Second)
			}

			room.Stop()
		}()

		room.Wait()
	},
}
