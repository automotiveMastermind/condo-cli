// --------------------------------------------------------------------------------------------------------------------
// <copyright file="root.go" company="automotiveMastermind and contributors">
// © automotiveMastermind and contributors. Licensed under MIT. See LICENSE and CREDITS for details.
// </copyright>
// --------------------------------------------------------------------------------------------------------------------

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
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
	Run: func(cmd *cobra.Command, args []string) {
		run(args)
	},
}

// Execute command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(args []string) {
	imageName := "automotivemastermind/condo"
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.39"))
	if err != nil {
		panic(err)
	}

	// pull condo image
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// format and print docker output
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(reader, os.Stderr, termFd, isTerm, nil)

	// create condo container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      imageName,
		Cmd:        []string{"condo"},
		WorkingDir: "/target",
		Tty:        true,
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// wait for container to start
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	// output container logs
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}
