package integrated

import (
	"encoding/json"
	"ssh2/models"
	"ssh2/plugins"
)

func GetLoginCommands(s *models.Session) (cmds []func(cp *plugins.Console) error, err error) {
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
	ps := []plugins.Plugin{}
	err = json.Unmarshal([]byte(s.Plugins), &ps)
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		result = append(result, plugins.Parse(p))
	}
	return result, nil
}
