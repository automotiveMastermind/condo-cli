package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCommand represents the stop command
var getCommand = &cobra.Command{
	Use:   "get",
	Short: "get <resource>",
	Long:  `get a resource owned by condo. ex: condo get cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Additional arguments required \n Common usage: condo get cluster")
	},
}

func init() {
	rootCmd.AddCommand(getCommand)
}
