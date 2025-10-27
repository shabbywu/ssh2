package cmd

import (
	"fmt"
	"github.com/iyzyi/aiopty/term"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"ssh2/integrated"
	"ssh2/models"
	"ssh2/utils/console"
)

var execCommand = &cli.Command{
	Name:      "login",
	Usage:     "登录服务器",
	ArgsUsage: "[Session Tag]",
	Before: func(ctx *cli.Context) error {
		if ctx.NArg() != 1 {
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
		tag := ctx.Args().First()
		session, err := models.GetByField[models.Session]("Session", "tag", tag)
		if err != nil {
			log.Fatal(err)
		}

		cmds, err := integrated.GetLoginCommands(&session)

		if err != nil {
			return err
		}
		cp, err := console.NewConsole()
		defer cp.Close()

		for _, cmd := range cmds {
			if err := cmd(cp); err != nil {
				log.Fatal(err)
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
		go func() { io.Copy(t, cp.GetStdout()) }()
		go func() { io.Copy(cp, t) }()
		return cp.Wait()
	},
}

func init() {
	App.Commands = append(App.Commands, execCommand)
}
