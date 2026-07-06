package plugins

import (
	"fmt"
	"github.com/ActiveState/termtest/expect"
	"github.com/tidwall/gjson"
	"ssh2/models"
	"ssh2/utils/console"
)

type ExpectStep struct {
	Expect string `yaml:"expect" json:"expect"`
	Send   string `yaml:"send" json:"send,omitempty"`
}

type ExpectPlugin struct {
	Expect string       `yaml:"expect" json:"expect"`
	Send   string       `yaml:"send" json:"send"`
	Steps  []ExpectStep `yaml:"steps" json:"steps,omitempty"`
}

func (plugin *ExpectPlugin) ToExpectCommand(session *models.Session) (func(cp *console.Console) error, error) {
	return func(cp *console.Console) error {
		steps := plugin.Steps
		if len(steps) == 0 {
			steps = []ExpectStep{{Expect: plugin.Expect, Send: plugin.Send}}
		}
		for _, step := range steps {
			if step.Expect == "" {
				return fmt.Errorf("EXPECT step requires an expect value")
			}
			if _, err := cp.Expect(expect.LongString(step.Expect)); err != nil {
				return err
			}
			if step.Send != "" {
				if _, err := cp.Send(step.Send); err != nil {
					return err
				}
			}
		}
		return nil
	}, nil
}

func ParseExpectPlugin(args gjson.Result) (ExpectAble, error) {
	if args.Get("raw").Exists() {
		return nil, fmt.Errorf("EXPECT.raw is not supported in the Go implementation; use EXPECT.args.steps instead")
	}
	plugin := &ExpectPlugin{
		Expect: args.Get("expect").Str,
		Send:   args.Get("send").Str,
	}
	for _, item := range args.Get("steps").Array() {
		step := ExpectStep{
			Expect: item.Get("expect").Str,
			Send:   item.Get("send").Str,
		}
		if step.Expect == "" {
			return nil, fmt.Errorf("EXPECT.steps[] requires expect")
		}
		plugin.Steps = append(plugin.Steps, step)
	}
	return plugin, nil
}

func init() {
	Register("EXPECT", ParseExpectPlugin)
}
