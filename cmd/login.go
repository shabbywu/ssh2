//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"ssh2/integrated"
	"ssh2/models"
	"ssh2/utils/tempfile"
	"syscall"
)

var execCommand = &cli.Command{
	Name:  "login",
	Usage: "登录服务器",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "tag",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		model, err := models.GetByField("Session", "tag", ctx.Value("tag"))

		if err != nil {
			log.Fatal(err)
		}

		m := tempfile.GetManager("")
		defer m.Clean()

		session := model.(*models.Session)
		file, err := integrated.ToExpectFile(session)
		if err != nil {
			return err
		}
		fmt.Printf("file: %s\n", file)

		// Create arbitrary command.
		c := exec.Command("expect", "-f", file)

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
