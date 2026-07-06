package cmd

import (
	_ "embed"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"ssh2/utils"
)

//go:embed ssh2_wrapper.sh
var wrappersh []byte

func wrapperPath() string {
	return filepath.Join(utils.SSH2_HOME, "ssh2_wrapper.sh")
}

func installWrapper() (string, error) {
	if err := os.MkdirAll(utils.SSH2_HOME, 0700); err != nil {
		return "", err
	}
	path := wrapperPath()
	if err := os.WriteFile(path, wrappersh, 0644); err != nil {
		return "", err
	}
	return path, nil
}

var installCommand = &cli.Command{Name: "install-ssh2-auto-complete",
	Usage: "安装快捷命令",
	Action: func(context *cli.Context) (err error) {
		wrapperPath, err := installWrapper()
		if err != nil {
			return err
		}
		fmt.Printf("install at %s\n", wrapperPath)
		return nil
	},
}

var wrapperPathCommand = &cli.Command{Name: "get-wrapper-dot-sh",
	Usage: "print ssh2 wrapper path",
	Action: func(context *cli.Context) (err error) {
		wrapperPath, err := installWrapper()
		if err != nil {
			return err
		}
		fmt.Println(wrapperPath)
		return nil
	},
}
