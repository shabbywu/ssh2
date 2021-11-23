package cmd

import (
	"encoding/json"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"ssh2/models"
)

var getCommand = &cli.Command{
	Name:      "get",
	Usage:     "get resource",
	ArgsUsage: "[资源类型]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "format",
			Required: false,
			Value:    "yaml",
		},
	},

	Before: func(ctx *cli.Context) (err error) {
		if ctx.NArg() != 1 {
			return cli.Exit("缺失资源类型参数", 1)
		}
		return nil
	},
	Action: func(ctx *cli.Context) (err error) {
		kind := ctx.Args().First()
		objs := models.List(kind)

		if objs == nil {
			return cli.Exit("not found.", 0)
		}

		var data []byte
		if ctx.Value("format") == "yaml" {
			data, err = yaml.Marshal(objs)
		} else {
			data, err = json.Marshal(objs)
		}
		if err != nil {
			return err
		}
		println(string(data))
		return nil
	},
}
