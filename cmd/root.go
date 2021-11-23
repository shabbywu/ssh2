package cmd

import (
	"github.com/urfave/cli/v2"
)

var App = &cli.App{}

func init() {
	App.Commands = []*cli.Command{
		getCommand,
		applyCommand,
	}
}
