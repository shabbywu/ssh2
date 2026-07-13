package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v2"

	"ssh2/utils"
)

//go:embed ssh2_wrapper.sh
var wrappersh []byte

//go:embed ssh2_wrapper.ps1
var wrapperps1 []byte

type wrapperAsset struct {
	filename string
	content  []byte
}

var (
	shellWrapper      = wrapperAsset{filename: "ssh2_wrapper.sh", content: wrappersh}
	powerShellWrapper = wrapperAsset{filename: "ssh2_wrapper.ps1", content: wrapperps1}
)

func wrapperForGOOS(goos string) wrapperAsset {
	if goos == "windows" {
		return powerShellWrapper
	}
	return shellWrapper
}

func wrapperPath(wrapper wrapperAsset) string {
	return filepath.Join(utils.SSH2_HOME, wrapper.filename)
}

func installWrapper(wrapper wrapperAsset) (string, error) {
	if err := os.MkdirAll(utils.SSH2_HOME, 0700); err != nil {
		return "", err
	}
	path := wrapperPath(wrapper)
	if err := os.WriteFile(path, wrapper.content, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func newWrapperPathCommand(name, usage string, wrapper wrapperAsset) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: usage,
		Action: func(context *cli.Context) error {
			path, err := installWrapper(wrapper)
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
}

var installCommand = &cli.Command{
	Name:  "install-ssh2-auto-complete",
	Usage: "安装当前平台的 go2s 快捷脚本",
	Action: func(context *cli.Context) error {
		path, err := installWrapper(wrapperForGOOS(runtime.GOOS))
		if err != nil {
			return err
		}
		fmt.Printf("install at %s\n", path)
		return nil
	},
}

var wrapperPathCommand = newWrapperPathCommand(
	"get-wrapper-dot-sh",
	"install the Bash/Zsh go2s wrapper and print its path",
	shellWrapper,
)

var powerShellWrapperPathCommand = newWrapperPathCommand(
	"get-wrapper-dot-ps1",
	"install the PowerShell go2s wrapper and print its path",
	powerShellWrapper,
)
