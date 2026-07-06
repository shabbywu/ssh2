package cmd

import (
	"bytes"
	"errors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"ssh2/parser"
)

var applyCommand = &cli.Command{
	Name:    "apply",
	Aliases: []string{"create"},
	Usage:   "apply resource definition",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) (err error) {
		data, err := ioutil.ReadFile(ctx.String("file"))
		if err != nil {
			return err
		}

		var record parser.DocumentRecord
		decoder := yaml.NewDecoder(bytes.NewReader(data))

		for {
			err := decoder.Decode(&record)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			_, err = parser.YamlParser{}.ParseRecord(record)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
