package plugins

import (
	"github.com/ActiveState/termtest/expect"
	"github.com/tidwall/gjson"
	"ssh2/models"
	"ssh2/utils/console"
)

type ExpectPlugin struct {
	Expect string `yaml:"expect" json:"expect"`
	Send   string `yaml:"send" json:"send"`
}

func (plugin *ExpectPlugin) ToExpectCommand(session *models.Session) (func(cp *console.Console) error, error) {
	return func(cp *console.Console) error {
		if _, err := cp.Expect(expect.LongString(plugin.Expect)); err != nil {
			return err
		}
		if _, err := cp.Send(plugin.Send); err != nil {
			return err
		}
		return nil
	}, nil
}

func ParseExpectPlugin(args gjson.Result) ExpectAble {
	return &ExpectPlugin{
		Expect: args.Get("expect").Str,
		Send:   args.Get("send").Str,
	}
}

func init() {
	Register("EXPECT", ParseExpectPlugin)
}
