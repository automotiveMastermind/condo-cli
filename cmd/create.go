package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create <resource>",
	Long:  `create a resource mangaed by condo. ex: condo create cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Additional arguments required \n Common usage: condo create cluster")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
