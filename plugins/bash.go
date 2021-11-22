package plugins

import "ssh2/models"

type ExpectAble interface {
	ToExpectCommand(session models.Session) (string, error)
}
