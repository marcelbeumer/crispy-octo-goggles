package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/client"
)

var remoteHost string
var remotePort int

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().IntVarP(&remotePort, "port", "p", 9998, "Server port")
	clientCmd.Flags().StringVarP(&remoteHost, "host", "H", "", "Server host")
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start client",
	Run: func(cmd *cobra.Command, args []string) {
		addr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
		fmt.Printf("Connecting to %s\n", addr)
		c := client.NewClient()
		<- c.Start()
	},
}
