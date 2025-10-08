//go:build !windows
// +build !windows

package console

import "io"

func (c *Console) CopyStdout(dest io.Writer) error {
	_, _ = io.Copy(dest, c.Tty())
	return nil
}
