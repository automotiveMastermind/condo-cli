package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCommand represents the stop command
var stopCommand = &cobra.Command{
	Use:   "stop",
	Short: "stop <resource>",
	Long:  `stop a resource owned by condo. ex: condo stop cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Additional arguments required \n Common usage: condo stop cluster")
	},
}

func init() {
	rootCmd.AddCommand(stopCommand)
}
