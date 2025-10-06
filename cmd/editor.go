package cmd

import (
	_ "embed"
	"github.com/urfave/cli/v2"
	"ssh2/db"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   string `json:"ttl,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var editorCommand = &cli.Command{Name: "editor",
	Usage: "終端編輯器",
	Action: func(context *cli.Context) (err error) {
		editor, err := db.NewEditor()
		if err != nil {
			return err
		}
		if err := editor.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, editorCommand)
}
