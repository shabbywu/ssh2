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

func GetManualSteps(s *models.Session) (steps []plugins.ManualStep, err error) {
	ps, err := getPlugins(s)
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		manualAble, ok := p.(plugins.ManualAble)
		if !ok {
			continue
		}
		pluginSteps, err := manualAble.ToManualSteps(s)
		if err != nil {
			for _, step := range steps {
				if step.Cleanup != nil {
					step.Cleanup()
				}
			}
			return nil, err
		}
		steps = append(steps, pluginSteps...)
	}
	return steps, nil
}

func getPlugins(s *models.Session) (result []plugins.ExpectAble, err error) {
	data := []byte(s.Plugins)
	for _, p := range gjson.ParseBytes(data).Array() {
		plugin, err := plugins.Parse(p)
		if err != nil {
			return nil, err
		}
		result = append(result, plugin)
	}
	return result, nil
}
