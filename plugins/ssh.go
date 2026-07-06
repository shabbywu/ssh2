package plugins

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"os/exec"
	"ssh2/models"
	"ssh2/utils/console"
	"strconv"
	"time"
)

type SSHPlugin struct {
}

func (plugin *SSHPlugin) ToExpectCommand(session *models.Session) (func(cp *console.Console) error, error) {
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
		return func(cp *console.Console) error {
			password, err := auth.DecryptedContent()
			if err != nil {
				return err
			}
			loginCmd := exec.Command("ssh", "-p", strconv.Itoa(serverConfig.Port), userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			if _, err := cp.ExpectString(auth.ExpectForPassword); err != nil {
				return fmt.Errorf("failed when expecting password input: %s", err)
			}
			time.Sleep(1)
			if _, err := cp.Send(password + "\r"); err != nil {
				return fmt.Errorf("failed when send password, detail: %s", err)
			}
			return nil

		}, nil
	case models.AUthInteractivePassword:
		return func(cp *console.Console) error {
			loginCmd := exec.Command("ssh", "-p", strconv.Itoa(serverConfig.Port), userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			if auth.ExpectForPassword != "" {
				if _, err := cp.ExpectString(auth.ExpectForPassword); err != nil {
					return fmt.Errorf("failed when expecting interactive password input: %s", err)
				}
			}
			return nil
		}, nil
	case models.AuthPublishKey:
		fallthrough
	case models.AUthPublishKeyFile:
		publishKeyPath, cleanup, err := auth.PublishKeyPath()
		if err != nil {
			return nil, err
		}
		return func(cp *console.Console) error {
			if cleanup != nil {
				cp.Cleanups = append(cp.Cleanups, cleanup)
			}
			loginCmd := exec.Command("ssh", "-i", publishKeyPath, "-p", strconv.Itoa(serverConfig.Port), userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				if cleanup != nil {
					cleanup()
				}
				return err
			}
			return nil
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("不支持的 auth 类型 %s", auth.Type))
	}
}

func ParseSSHPlugin(args gjson.Result) (ExpectAble, error) {
	return &SSHPlugin{}, nil
}

func init() {
	Register("SSH_LOGIN", ParseSSHPlugin)
}
