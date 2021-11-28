//go:build !windows
// +build !windows

package cmd

import (
	"github.com/creack/pty"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var execCommand = &cli.Command{
	Name:      "exec",
	Usage:     "执行指令",
	ArgsUsage: "[资源类型]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "format",
			Required:    false,
			DefaultText: "yaml",
		},
	},
	Action: func(ctx *cli.Context) error {
		// Create arbitrary command.
		c := exec.Command("ssh", "-i", "/Users/shabbywu/.ssh/TX-WIN10", "-p", "22", "root@81.71.45.54")

		// Start the command with a pty.
		ptmx, err := pty.Start(c)
		if err != nil {
			return err
		}
		// Make sure to close the pty at the end.
		defer func() { _ = ptmx.Close() }() // Best effort.

		// Handle pty size.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
					log.Printf("error resizing pty: %s", err)
				}
			}
		}()
		ch <- syscall.SIGWINCH // Initial resize.

		// Set stdin in raw mode.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

		// Copy stdin to the pty and the pty to stdout.
		// NOTE: The goroutine will keep reading until the next keystroke before returning.
		go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
		_, _ = io.Copy(os.Stdout, ptmx)

		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, execCommand)
}
