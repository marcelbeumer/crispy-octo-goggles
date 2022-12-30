package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var verbose bool
	var cmd = &cobra.Command{
		Use:   "demo-server",
		Short: "Rate limited demo server",
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
