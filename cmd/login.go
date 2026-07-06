package cmd

import (
	"fmt"
	"github.com/ActiveState/termtest/expect"
	"github.com/iyzyi/aiopty/term"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	osexec "os/exec"
	"os/signal"
	"ssh2/integrated"
	"ssh2/models"
	"ssh2/plugins"
	"ssh2/utils/console"
	"strings"
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
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"print"},
			Usage:   "print ssh command and expect steps without connecting",
		},
		&cli.BoolFlag{
			Name:  "direct",
			Usage: "use system ssh directly; skip ssh2 pty and EXPECT automation",
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

		if ctx.Bool("dry-run") {
			return printManualLogin(&session, ctx.Bool("direct"))
		}

		if ctx.Bool("direct") {
			return runDirectLogin(&session)
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

		stopSignals()
		stopInteractiveSignals := forwardInteractiveInterrupts(cp)
		defer stopInteractiveSignals()

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
			if isInteractiveInterruptExit(err) {
				return nil
			}
			return fmt.Errorf("login %q ssh process exited: %w", tag, err)
		}
		return nil
	},
}

func runDirectLogin(session *models.Session) error {
	step, err := plugins.SSHManualStep(session)
	if err != nil {
		return err
	}
	if step.Cleanup != nil {
		defer step.Cleanup()
	}
	if len(step.Command) == 0 {
		return fmt.Errorf("session %q has no ssh command", session.Tag)
	}
	cmd := osexec.Command(step.Command[0], step.Command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*osexec.ExitError); ok {
			return cli.Exit(fmt.Sprintf("direct ssh exited with status %d", exitErr.ExitCode()), exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func printManualLogin(session *models.Session, direct bool) error {
	var steps []plugins.ManualStep
	var err error
	if direct {
		step, err := plugins.SSHManualStep(session)
		if err != nil {
			return err
		}
		steps = []plugins.ManualStep{step}
	} else {
		steps, err = integrated.GetManualSteps(session)
	}
	if err != nil {
		return err
	}
	for i, step := range steps {
		fmt.Printf("# step %d: %s\n", i+1, step.Kind)
		if len(step.Command) > 0 {
			if step.CleanupPath != "" {
				fmt.Println("(")
				fmt.Printf("  ssh2_key_file=%s\n", shellQuote(step.CleanupPath))
				fmt.Println("  trap 'rm -f \"$ssh2_key_file\"' EXIT INT TERM HUP")
				fmt.Printf("  %s\n", shellQuoteCommandWithRaw(step.Command, map[string]string{
					step.CleanupPath: "\"$ssh2_key_file\"",
				}))
				fmt.Println(")")
				fmt.Println("# note: temporary key file is removed when the subshell exits")
			} else {
				fmt.Println(shellQuoteCommand(step.Command))
			}
		}
		if step.Expect != "" {
			fmt.Printf("expect: %q\n", step.Expect)
		}
		if step.Send != "" {
			fmt.Printf("send:   %q\n", step.Send)
		}
		if step.Note != "" {
			fmt.Printf("# note: %s\n", step.Note)
		}
		if i != len(steps)-1 {
			fmt.Println()
		}
	}
	return nil
}

func shellQuoteCommand(args []string) string {
	return shellQuoteCommandWithRaw(args, nil)
}

func shellQuoteCommandWithRaw(args []string, raw map[string]string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if value, ok := raw[arg]; ok {
			quoted = append(quoted, value)
		} else {
			quoted = append(quoted, shellQuote(arg))
		}
	}
	return strings.Join(quoted, " ")
}

func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}
	if strings.IndexFunc(arg, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || strings.ContainsRune("@%_+=:,./-", r))
	}) == -1 {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
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
			if cp != nil {
				cp.KillChildren()
				_ = cp.Close()
			}
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
	var stopOnce sync.Once
	return interrupted, func() {
		stopOnce.Do(func() {
			signal.Stop(signals)
			close(done)
		})
	}
}

func forwardInteractiveInterrupts(w io.Writer) func() {
	const interruptByte = byte(0x03)

	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, os.Interrupt)
	go func() {
		for {
			select {
			case <-signals:
				if w != nil {
					_, _ = w.Write([]byte{interruptByte})
				}
			case <-done:
				return
			}
		}
	}()
	var stopOnce sync.Once
	return func() {
		stopOnce.Do(func() {
			signal.Stop(signals)
			close(done)
		})
	}
}

func isInteractiveInterruptExit(err error) bool {
	exitErr, ok := err.(*osexec.ExitError)
	return ok && exitErr.ExitCode() == 130
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
