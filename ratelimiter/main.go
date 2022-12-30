package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var (
		hostKey = "host"
		portKey = "port"
		host    string
		port    int
	)
	var cmd = &cobra.Command{
		Use:   "ratelimiter",
		Short: "Start rate limited demo server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := StartServer(fmt.Sprintf("%s:%d", host, port)); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	viper.AutomaticEnv()
	viper.SetDefault(hostKey, "")
	viper.SetDefault(portKey, 8000)
	viper.BindPFlag(hostKey, cmd.Flags().Lookup(hostKey))
	viper.BindPFlag(portKey, cmd.Flags().Lookup(portKey))

	cmd.PersistentFlags().StringVarP(&host, hostKey, "H", viper.GetString(hostKey), "Host to listen on")
	cmd.Flags().IntVarP(&port, portKey, "P", viper.GetInt(portKey), "Port to listen on")

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
