//go:build !windows
// +build !windows

package console

import (
	"io"
)

func (c *Console) GetStdout() io.Reader {
	reader, writer := io.Pipe()
	go func() {
		_, err := c.Pty.WriteTo(writer)
		_ = writer.CloseWithError(err)
	}()
	return reader
}
