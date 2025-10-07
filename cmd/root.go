package cmd

import (
	"github.com/urfave/cli/v2"
	"ssh2/utils/tempfile"
)

var App = cli.NewApp()

func init() {
	App.Usage = "ssh连接管理工具"
	App.EnableBashCompletion = true
	App.After = func(c *cli.Context) error {
		tempfile.GetManager("").Clean()
		return nil
	}
	App.Commands = append(App.Commands, []*cli.Command{
		getCommand,
		applyCommand,
		installCommand,
	}...)
}
