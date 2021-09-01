package console

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func Start(c *exec.Cmd) (err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	c.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	err = c.Run()
	return err
}
