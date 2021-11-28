package cmd

import (
	"github.com/urfave/cli/v2"
	"html/template"
	"os"
	"ssh2/models"
)

var getCommand = &cli.Command{
	Name:      "get",
	Usage:     "get resource",
	ArgsUsage: "[资源类型]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "template",
			Required: false,
			Value:    "{{ .Tag }}",
		},
		&cli.StringFlag{
			Name:     "kind",
			Required: false,
			Value:    "Session",
		},
	},
	Action: func(ctx *cli.Context) (err error) {
		kind := ctx.Value("kind").(string)
		objs := models.List(kind)

		if objs == nil {
			return cli.Exit("not found.", 0)
		}

		templator, _ := template.New("template").Parse(ctx.Value("template").(string))

		for _, obj := range objs {
			_ = templator.Execute(os.Stdout, obj)
			os.Stdout.Write([]byte("\n"))
		}
		return nil
	},
}
