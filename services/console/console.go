//go:build !windows

package console

import (
	"github.com/creack/pty"
	"io"
	"os"
	"os/exec"
)

func Start(c *exec.Cmd) (err error) {
	f, err := pty.Start(cmd)
	io.Copy(os.Stdout, f)
}
