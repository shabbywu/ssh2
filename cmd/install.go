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

var installCommand = &cli.Command{Name: "install-ssh2-auto-complete",
	Usage: "安装快捷命令",
	Action: func(context *cli.Context) (err error) {
		wrapperPath := filepath.Join(utils.SSH2_HOME, "ssh2_wrapper.sh")
		if err := os.WriteFile(wrapperPath, wrappersh, 0666); err != nil {
			return err
		}
		fmt.Printf("install at %s\n", wrapperPath)
		return nil
	},
}
