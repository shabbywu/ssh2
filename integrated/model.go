package integrated

import (
	"github.com/tidwall/gjson"
	"ssh2/models"
	"ssh2/plugins"
	"ssh2/utils/console"
)

func GetLoginCommands(s *models.Session) (cmds []func(cp *console.Console) error, err error) {
	ps, err := getPlugins(s)
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		cmd, err := p.ToExpectCommand(s)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	return
}

func getPlugins(s *models.Session) (result []plugins.ExpectAble, err error) {
	data := []byte(s.Plugins)
	if err != nil {
		return nil, err
	}
	for _, p := range gjson.ParseBytes(data).Array() {
		result = append(result, plugins.Parse(p))
	}
	return result, nil
}
