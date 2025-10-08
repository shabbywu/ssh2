package plugins

import (
	"github.com/tidwall/gjson"
	"ssh2/models"
	"ssh2/utils/console"
)

type Plugin struct {
	Kind string `yaml:"kind" json:"kind"`
	Args interface{}
}

type ExpectAble interface {
	ToExpectCommand(session *models.Session) (func(cp *console.Console) error, error)
}

var handlers = map[string]func(args gjson.Result) ExpectAble{}

func Register(kind string, parser func(args gjson.Result) ExpectAble) {
	handlers[kind] = parser
}

func Parse(p gjson.Result) ExpectAble {
	if parser, ok := handlers[p.Get("kind").Str]; ok {
		return parser(p.Get("args"))
	}
	return nil
}
