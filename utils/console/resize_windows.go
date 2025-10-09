//go:build windows
// +build windows

package console

import (
	"golang.org/x/term"
	"os"
	"time"
)

func (c *Console) MonitorTerminalSize() {
	var lastWidth, lastHeight int

	// 首次获取
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err == nil {
		lastWidth, lastHeight = width, height
		c.Pty.Resize(width, height)
	}

	// 轮询检测大小变化
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.closed:
			return
		case <-ticker.C:
			width, height, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				continue
			}

			// 只在大小改变时回调
			if width != lastWidth || height != lastHeight {
				lastWidth, lastHeight = width, height
				c.Pty.Resize(width, height)
			}
		}
	}
}
