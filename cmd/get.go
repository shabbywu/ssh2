package cmd

import (
	"fmt"
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
		kind := ctx.String("kind")
		if ctx.NArg() > 0 {
			kind = ctx.Args().First()
		}
		objs, err := listObjects(kind)
		if err != nil {
			return err
		}

		if objs == nil {
			return cli.Exit("not found.", 0)
		}

		template, err := template.New("template").Parse(ctx.String("template"))
		if err != nil {
			return err
		}
		for _, obj := range objs {
			if err := template.Execute(os.Stdout, obj); err != nil {
				return err
			}
			os.Stdout.Write([]byte("\n"))
		}
		return nil
	},
}

func listObjects(kind string) ([]interface{}, error) {
	var result []interface{}
	switch kind {
	case "AuthMethod", "auth":
		for _, obj := range models.List[models.AuthMethod]("AuthMethod") {
			result = append(result, obj)
		}
	case "ClientConfig", "client":
		for _, obj := range models.List[models.ClientConfig]("ClientConfig") {
			result = append(result, obj)
		}
	case "ServerConfig", "server":
		for _, obj := range models.List[models.ServerConfig]("ServerConfig") {
			result = append(result, obj)
		}
	case "Session", "session":
		for _, obj := range models.List[models.Session]("Session") {
			result = append(result, obj)
		}
	default:
		return nil, fmt.Errorf("unsupported kind %s", kind)
	}
	return result, nil
}
