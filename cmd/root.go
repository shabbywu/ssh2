package cmd

import (
	"github.com/urfave/cli/v2"
)

var App = cli.NewApp()

func init() {
	App.EnableBashCompletion = true
	App.Commands = append(App.Commands, []*cli.Command{
		getCommand,
		applyCommand,
	}...)
}
