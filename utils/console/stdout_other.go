//go:build !windows
// +build !windows

package console

import (
	"io"
)

func (c *Console) GetStdout() io.Reader {
	return c.Tty()
}
