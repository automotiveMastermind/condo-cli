package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// destroyCommand represents the destroy command
var destroyCommand = &cobra.Command{
	Use:   "destroy",
	Short: "short discription",
	Long:  `long discription`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Additional arguments required \n Common usage: condo destroy cluster")
	},
}

func init() {
	rootCmd.AddCommand(destroyCommand)

}
