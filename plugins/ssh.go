package plugins

import (
	"errors"
	"fmt"
	"ssh2/models"
)

type SSHPlugin struct {
}

func (plugin *SSHPlugin) ToExpectCommand(session *models.Session) (string, error) {
	clientConfig, err := session.GetClientConfig()
	if err != nil {
		return "", err
	}
	auth, err := clientConfig.GetAuthMethod()
	if err != nil {
		return "", err
	}
	serverConfig, err := session.GetServerConfig()
	if err != nil {
		return "", err
	}
	userHost := fmt.Sprintf("%s@%s", clientConfig.User, serverConfig.Host)

	switch auth.Type {
	case models.AuthPassword:
		return fmt.Sprintf(
			`spawn ssh -p %d %s
expect "%s"
send "%s\r"
`, serverConfig.Port, userHost, auth.ExpectForPassword, auth.GetDecryptedContent()), nil
	case models.AuthPublishKey:
		fallthrough
	case models.AUthPublishKeyFile:
		publishKeyPath := auth.GetPublishKeyPath()
		return fmt.Sprintf(`spawn ssh -i %s -p %d %s`, publishKeyPath, serverConfig.Port, userHost), nil
	default:
		return "", errors.New(fmt.Sprintf("不支持的 auth 类型 %s", auth.Type))
	}
}

func ParseSSHPlugin(args interface{}) ExpectAble {
	return &SSHPlugin{}
}

func init() {
	Register("SSH_LOGIN", ParseSSHPlugin)
}
