package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

// RunOptions contains the input for the run command
type RunOptions struct {
	ImageTag string
	Args     []string
	GoOS     string
	GoArch   string
}

// NewRunOptions creates a default RunOptions with ImageTag set to beta-golang
func NewRunOptions() *RunOptions {
	return &RunOptions{
		ImageTag: "latest",
		GoOS:     runtime.GOOS,
		GoArch:   runtime.GOARCH,
	}
}

var options = NewRunOptions()

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run condo against current path",
	Long:  `Will pull automotiveMastermind/condo:beta-golang and run it against the current path.`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	runCmd.Flags().StringVar(&options.ImageTag, "image-tag", options.ImageTag, "Sets the condo image tag to use when running")
	runCmd.Flags().StringVar(&options.GoOS, "os", options.GoOS, "Sets the default OS to build")
	runCmd.Flags().StringVar(&options.GoArch, "arch", options.GoArch, "Sets the default Arch to build")
	runCmd.Flags().StringSliceVar(&options.Args, "args", options.Args, "Sets condo arguments")

	rootCmd.AddCommand(runCmd)
}

func run() {
	imageName := fmt.Sprintf("sjk07/condo:%s", options.ImageTag)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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

	// current path
	pwd, _ := filepath.Abs("./")

	// create condo container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		Cmd:          append([]string{"condo"}, options.Args...),
		WorkingDir:   "/target",
		AttachStderr: true,
		AttachStdout: true,
		Env:          []string{"GOOS=" + options.GoOS, "GOARCH=" + options.GoArch},
	},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: pwd,
					Target: "/target",
				},
			},
		}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// output container logs
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
