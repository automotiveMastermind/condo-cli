package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

type VersionControl string
type CI string
type Language string

type structure struct {
	versioncontrol VersionControl
	ci             []CI
	languages      []Language
}

const (
	Git VersionControl = "Git"
)

const (
	Go     Language = "Go"
	Python          = "Python"
	Dotnet          = "Dotnet"
	Unity           = "Unity"
)

const (
	CircleCI    CI = "CircleCI"
	GitLab		   = "GitLab"
	Github		   = "Github"
	AzureDevops    = "AzureDevops"
)

// runCmd represents the run command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a condo build",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		prep()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func prep() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
}
