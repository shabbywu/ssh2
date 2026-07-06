package cmd

import (
	"fmt"
	"github.com/ActiveState/termtest/expect"
	"github.com/iyzyi/aiopty/term"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"os/signal"
	"ssh2/integrated"
	"ssh2/models"
	"ssh2/utils/console"
	"sync"
	"syscall"
)

var execCommand = &cli.Command{
	Name:      "login",
	Usage:     "登录服务器",
	ArgsUsage: "[Session Tag]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "tag",
			Usage: "session tag",
		},
	},
	Before: func(ctx *cli.Context) error {
		if ctx.NArg() != 1 && ctx.String("tag") == "" {
			cli.ShowCommandHelp(ctx, "login")
			objs := models.List[models.Session]("Session")
			fmt.Println("avaialbe sessions:")
			for _, session := range objs {
				fmt.Printf("- %s\n", session.Tag)
			}
			os.Exit(1)
		}
		return nil
	},
	Action: func(ctx *cli.Context) error {
		tag := ctx.String("tag")
		if tag == "" {
			tag = ctx.Args().First()
		}
		session, err := models.GetByField[models.Session]("Session", "tag", tag)
		if err != nil {
			return fmt.Errorf("session %q not found: %w", tag, err)
		}

		cmds, err := integrated.GetLoginCommands(&session)

		if err != nil {
			return err
		}
		cp, err := console.NewConsole(expect.WithStdout(os.Stdout))
		if err != nil {
			return err
		}
		defer cp.Close()
		interrupted, stopSignals := handleLoginSignals(cp)
		defer stopSignals()

		for i, cmd := range cmds {
			if err := cmd(cp); err != nil {
				if interruptErr := loginInterruptError(tag, interrupted); interruptErr != nil {
					return interruptErr
				}
				return fmt.Errorf("login %q failed at step %d/%d: %w", tag, i+1, len(cmds), err)
			}
		}

		// When the terminal window size changes, synchronize the size of the pty
		onSizeChange := func(cols, rows uint16) {
			cp.Pty.Resize(cols, rows)
		}

		// enable terminal
		t, err := term.Open(os.Stdin, os.Stdout, onSizeChange)
		if err != nil {
			return err
		}
		defer t.Close()

		// start data exchange between terminal and pty
		go func() { io.Copy(os.Stdout, cp.GetStdout()) }()
		go func() { io.Copy(cp, t) }()
		if err := cp.Wait(); err != nil {
			if interruptErr := loginInterruptError(tag, interrupted); interruptErr != nil {
				return interruptErr
			}
			return fmt.Errorf("login %q ssh process exited: %w", tag, err)
		}
		return nil
	},
}

func handleLoginSignals(cp *console.Console) (<-chan os.Signal, func()) {
	signals := make(chan os.Signal, 1)
	interrupted := make(chan os.Signal, 1)
	done := make(chan struct{})
	var interruptOnce sync.Once
	interrupt := func(sig os.Signal) {
		interruptOnce.Do(func() {
			select {
			case interrupted <- sig:
			default:
			}
			cp.KillChildren()
			_ = cp.Close()
		})
	}
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-signals:
			interrupt(sig)
		case <-done:
		}
	}()
	return interrupted, func() {
		signal.Stop(signals)
		close(done)
	}
}

func loginInterruptError(tag string, interrupted <-chan os.Signal) error {
	select {
	case sig := <-interrupted:
		return fmt.Errorf("login %q interrupted by %s", tag, sig)
	default:
		return nil
	}
}

func init() {
	App.Commands = append(App.Commands, execCommand)
}
