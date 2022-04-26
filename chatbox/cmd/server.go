package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/server"
)

var serverHost string
var serverPort int

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&remotePort, "port", "p", 9998, "Server port")
	serverCmd.Flags().StringVarP(&remoteHost, "host", "H", "", "Server host")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		addr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
		fmt.Printf("Starting server on %s\n", addr)
		err := server.StartServer(addr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
