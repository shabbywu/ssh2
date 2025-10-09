//go:build !windows
// +build !windows

package console

import (
	"io"
	"os"
)

func (c *Console) CopyStdout(dest io.Writer) error {
	file := os.NewFile(c.Pty.TerminalOutFd(), "stdout")
	_, _ = io.Copy(dest, file)
	return nil
}
