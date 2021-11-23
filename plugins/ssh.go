package plugins

import "ssh2/models"

type SSHPlugin struct {
}

func (plugin *SSHPlugin) ToExpectCommand(session models.Session) (string, error) {
	return "", nil
}
