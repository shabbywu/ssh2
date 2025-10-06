package plugins

import (
	"github.com/ActiveState/termtest/expect"
	"github.com/tidwall/gjson"
	"os"
	"os/exec"
	"ssh2/models"
)

type Plugin struct {
	Kind string `yaml:"kind" json:"kind"`
	Args interface{}
}

type Console struct {
	Children []*exec.Cmd
	*expect.Console
}

func (c *Console) Wait() error {
	for _, child := range c.Children {
		if err := child.Wait(); err != nil {
			return err
		}
	}
	return nil
}

type ExpectAble interface {
	ToExpectCommand(session *models.Session) (func(cp *Console) error, error)
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

func NewConsole() (*Console, error) {
	cp, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	return &Console{Console: cp}, err
}
