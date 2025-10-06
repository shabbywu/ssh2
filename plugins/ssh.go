package plugins

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"os/exec"
	"ssh2/models"
	"strconv"
	"time"
)

type SSHPlugin struct {
}

func (plugin *SSHPlugin) ToExpectCommand(session *models.Session) (func(cp *Console) error, error) {
	clientConfig, err := session.GetClientConfig()
	if err != nil {
		return nil, err
	}
	auth, err := clientConfig.GetAuthMethod()
	if err != nil {
		return nil, err
	}
	serverConfig, err := session.GetServerConfig()
	if err != nil {
		return nil, err
	}
	userHost := fmt.Sprintf("%s@%s", clientConfig.User, serverConfig.Host)

	switch auth.Type {
	case models.AuthPassword:
		return func(cp *Console) error {
			loginCmd := exec.Command("ssh", "-p", strconv.Itoa(serverConfig.Port), userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			if _, err := cp.ExpectString(auth.ExpectForPassword); err != nil {
				return fmt.Errorf("failed when expect passowrd input, detail: %s", err)
			}
			time.Sleep(1)
			if _, err := cp.Send(auth.GetDecryptedContent() + "\r"); err != nil {
				return fmt.Errorf("failed when send password, detail: %s", err)
			}
			return nil

		}, nil
	case models.AuthPublishKey:
		fallthrough
	case models.AUthPublishKeyFile:
		publishKeyPath := auth.GetPublishKeyPath()
		return func(cp *Console) error {
			loginCmd := exec.Command("ssh", "-p", strconv.Itoa(serverConfig.Port), userHost, "-i", publishKeyPath)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			return nil
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("不支持的 auth 类型 %s", auth.Type))
	}
}

func ParseSSHPlugin(args gjson.Result) ExpectAble {
	return &SSHPlugin{}
}

func init() {
	Register("SSH_LOGIN", ParseSSHPlugin)
}
