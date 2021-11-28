package integrated

import (
	"encoding/json"
	"ssh2/models"
	"ssh2/plugins"
	"ssh2/utils/tempfile"
	"strings"
)

func ToExpectFile(s *models.Session) (string, error) {
	cmd, err := ToExpect(s)
	if err != nil {
		return "", err
	}

	file, err := tempfile.GetManager("").TempFile(s.GetName())
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.WriteString(cmd)

	return file.Name(), nil
}

func ToExpect(s *models.Session) (string, error) {
	cmds := []string{
		"#!/usr/bin/expect",
		"set timeout 20s",
		`
trap {
	set rows [stty rows]
	set cols [stty columns]
	stty rows $rows columns $cols < $spawn_out(slave,name)
} WINCH
`,
	}

	ps, err := getPlugins(s)
	if err != nil {
		return "", err
	}

	for _, p := range ps {
		cmd, err := p.ToExpectCommand(s)
		if err != nil {
			return "", err
		}
		cmds = append(cmds, cmd)
	}

	if cmds[len(cmds)-1] != "interact" {
		cmds = append(cmds, "interact")
	}

	return strings.Join(cmds, "\n"), nil
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
