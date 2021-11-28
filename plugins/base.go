package plugins

import (
	"ssh2/models"
)

type Plugin struct {
	Kind string `yaml:"kind" json:"kind"`
	Args interface{}
}

type ExpectAble interface {
	ToExpectCommand(session *models.Session) (string, error)
}

var handlers = map[string]func(args interface{}) ExpectAble{}

func Register(kind string, parser func(args interface{}) ExpectAble) {
	handlers[kind] = parser
}

func Parse(p Plugin) ExpectAble {
	if parser, ok := handlers[p.Kind]; ok {
		return parser(p.Args)
	}
	return nil
}
