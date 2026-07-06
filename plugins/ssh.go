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

var defaultSSHOptions = []string{
	"-tt",
	"-o", "ConnectTimeout=10",
	"-o", "ConnectionAttempts=1",
	"-o", "StrictHostKeyChecking=accept-new",
	"-o", "ServerAliveInterval=15",
	"-o", "ServerAliveCountMax=2",
}

var keyAuthSSHOptions = []string{
	"-o", "BatchMode=yes",
	"-o", "PreferredAuthentications=publickey",
	"-o", "PasswordAuthentication=no",
	"-o", "KbdInteractiveAuthentication=no",
	"-o", "IdentitiesOnly=yes",
	"-o", "IdentityAgent=none",
}

func sshCommand(port int, userHost string, extraArgs ...string) *exec.Cmd {
	args := append([]string{}, defaultSSHOptions...)
	args = append(args, extraArgs...)
	args = append(args, "-p", strconv.Itoa(port), userHost)
	return exec.Command("ssh", args...)
}

func sshCommandArgs(port int, userHost string, extraArgs ...string) []string {
	return sshCommand(port, userHost, extraArgs...).Args
}

func sessionSSHTarget(session *models.Session) (*models.ClientConfig, *models.AuthMethod, *models.ServerConfig, string, error) {
	clientConfig, err := session.GetClientConfig()
	if err != nil {
		return nil, nil, nil, "", err
	}
	auth, err := clientConfig.GetAuthMethod()
	if err != nil {
		return nil, nil, nil, "", err
	}
	serverConfig, err := session.GetServerConfig()
	if err != nil {
		return nil, nil, nil, "", err
	}
	userHost := fmt.Sprintf("%s@%s", clientConfig.User, serverConfig.Host)
	return clientConfig, auth, serverConfig, userHost, nil
}

func (plugin *SSHPlugin) ToExpectCommand(session *models.Session) (func(cp *console.Console) error, error) {
	_, auth, serverConfig, userHost, err := sessionSSHTarget(session)
	if err != nil {
		return nil, err
	}

	switch auth.Type {
	case models.AuthPassword:
		return func(cp *console.Console) error {
			password, err := auth.DecryptedContent()
			if err != nil {
				return err
			}
			loginCmd := sshCommand(serverConfig.Port, userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			output, err := cp.ExpectString(auth.ExpectForPassword)
			if err != nil {
				return expectError(auth.ExpectForPassword, output, err)
			}
			time.Sleep(1)
			if _, err := cp.Send(password + "\r"); err != nil {
				return fmt.Errorf("failed when send password, detail: %s", err)
			}
			return nil

		}, nil
	case models.AUthInteractivePassword:
		return func(cp *console.Console) error {
			loginCmd := sshCommand(serverConfig.Port, userHost)
			cp.Children = append(cp.Children, loginCmd)
			if err := cp.Pty.StartProcessInTerminal(loginCmd); err != nil {
				return err
			}
			if auth.ExpectForPassword != "" {
				output, err := cp.ExpectString(auth.ExpectForPassword)
				if err != nil {
					return expectError(auth.ExpectForPassword, output, err)
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
			keyArgs := append([]string{}, keyAuthSSHOptions...)
			keyArgs = append(keyArgs, "-i", publishKeyPath)
			loginCmd := sshCommand(serverConfig.Port, userHost, keyArgs...)
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

func (plugin *SSHPlugin) ToManualSteps(session *models.Session) ([]ManualStep, error) {
	step, err := SSHManualStep(session)
	if err != nil {
		return nil, err
	}
	return []ManualStep{step}, nil
}

func SSHManualStep(session *models.Session) (ManualStep, error) {
	_, auth, serverConfig, userHost, err := sessionSSHTarget(session)
	if err != nil {
		return ManualStep{}, err
	}

	switch auth.Type {
	case models.AuthPassword, models.AUthInteractivePassword:
		return ManualStep{
			Kind:    "SSH_LOGIN",
			Command: sshCommandArgs(serverConfig.Port, userHost),
		}, nil
	case models.AuthPublishKey:
		publishKeyPath, cleanup, err := auth.PublishKeyPath()
		if err != nil {
			return ManualStep{}, err
		}
		keyArgs := append([]string{}, keyAuthSSHOptions...)
		keyArgs = append(keyArgs, "-i", publishKeyPath)
		return ManualStep{
			Kind:        "SSH_LOGIN",
			Command:     sshCommandArgs(serverConfig.Port, userHost, keyArgs...),
			CleanupPath: publishKeyPath,
			Cleanup:     cleanup,
		}, nil
	case models.AUthPublishKeyFile:
		publishKeyPath, _, err := auth.PublishKeyPath()
		if err != nil {
			return ManualStep{}, err
		}
		keyArgs := append([]string{}, keyAuthSSHOptions...)
		keyArgs = append(keyArgs, "-i", publishKeyPath)
		return ManualStep{
			Kind:    "SSH_LOGIN",
			Command: sshCommandArgs(serverConfig.Port, userHost, keyArgs...),
		}, nil
	default:
		return ManualStep{}, errors.New(fmt.Sprintf("不支持的 auth 类型 %s", auth.Type))
	}
}

func ParseSSHPlugin(args gjson.Result) (ExpectAble, error) {
	return &SSHPlugin{}, nil
}

func init() {
	Register("SSH_LOGIN", ParseSSHPlugin)
}
