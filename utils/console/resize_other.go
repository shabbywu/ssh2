//go:build !windows
// +build !windows

package console

import (
	"golang.org/x/term"
	"os"
	"os/signal"
	"syscall"
)

func (c *Console) MonitorTerminalSize() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	// 触发初始大小
	ch <- syscall.SIGWINCH

	for {
		select {
		case <-ch:
			width, height, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				continue
			}
			c.Pty.Resize(uint16(width), uint16(height))
		case <-c.closed:
			return
		}
	}
}
