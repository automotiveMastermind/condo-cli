package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

//Unit test entry point. Use "go run main.go test" to run. TO BE REMOVED FOR MAJOR RELEASE
var testCommand = &cobra.Command{
	Use:   "test",
	Short: "short discription",
	Long:  `long discription`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test command called")

	},
}

func init() {
	//comment below line to detach "test" command from application
	rootCmd.AddCommand(testCommand)

}
