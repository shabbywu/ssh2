package cmd

import (
	"bytes"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"ssh2/parser"
)

var applyCommand = &cli.Command{
	Name:  "apply",
	Usage: "apply resource definition",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) (err error) {
		data, err := ioutil.ReadFile(ctx.Value("file").(string))
		if err != nil {
			return err
		}

		var record parser.DocumentRecord
		decoder := yaml.NewDecoder(bytes.NewReader(data))

		for decoder.Decode(&record) == nil {
			_, err := parser.YamlParser{}.ParseRecord(record)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
