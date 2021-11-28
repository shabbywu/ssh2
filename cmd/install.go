package cmd

import (
	_ "embed"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"path/filepath"
	"ssh2/utils"
)

//go:embed ssh2_wrapper.sh
var wrappersh []byte

var installCommand = &cli.Command{Name: "install-ssh2-auto-complete",
	Usage: "安装快捷命令",
	Action: func(context *cli.Context) (err error) {
		wrapperPath := filepath.Join(utils.SSH2_HOME, "ssh2_wrapper.sh")
		ioutil.WriteFile(wrapperPath, wrappersh, 0666)
		fmt.Println(wrapperPath)
		return nil
	},
}
