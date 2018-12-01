// --------------------------------------------------------------------------------------------------------------------
// <copyright file="root.go" company="automotiveMastermind and contributors">
// © automotiveMastermind and contributors. Licensed under MIT. See LICENSE and CREDITS for details.
// </copyright>
// --------------------------------------------------------------------------------------------------------------------

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "condo",
	Short: "Condo is a build system for <any> project",
	Long: `Condo is a cross-platform command line interface (CLI) build system for projects using NodeJS, CoreCLR, .NET Framework, or... well, anything. 
It is capable of automatically detecting and executing all of the steps necessary to make project function correctly. 
Some of the most-used features of the build system include:

	* Automatic semantic versioning
	* Restoring package manager dependencies (NuGet, NPM, Bower)
	* Executing default task runner commands
	* Compiling projects and test projects (package.json and msbuild)
	* Executing unit tests (xunit, mocha, jasmine, karma, protractor)
	* Packing NuGet packages
	* Pushing (Publishing) NuGet packages
	`,
}

// Execute command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
