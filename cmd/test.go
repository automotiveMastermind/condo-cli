/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	// "os"
	// "os/exec"

	// log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//var HELM_DEFAULT_GITHUB = "https://automotivemastermind@dev.azure.com/automotivemastermind/aM/_git/am.devops.helm"

// createCmd represents the create command
var testCommand = &cobra.Command{
	Use:   "test",
	Short: "short discription",
	Long:  `long discription`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test command called")

		// home, err := os.UserHomeDir()
		// if err != nil {
		// 	log.Fatal("can't find users home directory\n")
		// }

		// clusterRootPath = fmt.Sprintf(home + "/.am/clusters/local/")

		// errCreateDir := os.MkdirAll(clusterRootPath, 0755)
		// check(errCreateDir)

		// commandExec := exec.Command("git", "clone", HELM_DEFAULT_GITHUB, "helm")
		// commandExec.Dir = clusterRootPath
		// out, err := commandExec.CombinedOutput()
		// if err != nil {
		// 	log.Fatalf("Failed to the clone auxiliary configurations for ", err)
		// }

		// log.Infof("%s", out)

	},
}

func init() {
	//rootCmd.AddCommand(testCommand)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
