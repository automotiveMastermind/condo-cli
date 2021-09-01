//go:build !windows

package console

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func Start(c *exec.Cmd) (err error) {
	f, err := pty.Start(c)
	io.Copy(os.Stdout, f)

	return err
}
