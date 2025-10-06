package cmd

import (
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"ssh2/integrated"
	"ssh2/models"
	"ssh2/plugins"
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

		session := model.(*models.Session)
		cmds, err := integrated.GetLoginCommands(session)

		if err != nil {
			return err
		}
		cp, err := plugins.NewConsole()
		defer cp.Close()

		for _, cmd := range cmds {
			if err := cmd(cp); err != nil {
				log.Fatal(err)
			}
		}

		// Copy stdin to the pty and the pty to stdout.
		// NOTE: The goroutine will keep reading until the next keystroke before returning.
		go func() { _, _ = io.Copy(cp.Pty.InPipe(), os.Stdin) }()
		go func() { _, _ = io.Copy(os.Stdout, cp.Pty.OutPipe()) }()
		if err = cp.Wait(); err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, execCommand)
}
