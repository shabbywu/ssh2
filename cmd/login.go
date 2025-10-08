package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
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

		// 将标准输入设置为原始模式
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatal(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
		go cp.CopyStdout(os.Stdout)
		// Copy stdin to the pty and the pty to stdout.
		// NOTE: The goroutine will keep reading until the next keystroke before returning.
		if err = cp.Wait(); err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, execCommand)
}
