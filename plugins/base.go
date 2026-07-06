package plugins

import (
	"fmt"
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

type ManualStep struct {
	Kind        string
	Command     []string
	Expect      string
	Send        string
	Note        string
	CleanupPath string
	Cleanup     func()
}

type ManualAble interface {
	ToManualSteps(session *models.Session) ([]ManualStep, error)
}

var handlers = map[string]func(args gjson.Result) (ExpectAble, error){}

func Register(kind string, parser func(args gjson.Result) (ExpectAble, error)) {
	handlers[kind] = parser
}

func Parse(p gjson.Result) (ExpectAble, error) {
	kind := p.Get("kind").Str
	if parser, ok := handlers[kind]; ok {
		return parser(p.Get("args"))
	}
	return nil, fmt.Errorf("unsupported plugin kind %q", kind)
}
