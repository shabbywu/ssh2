//go:build windows

package tempfile

import "ssh2/utils"

func defaultDir() string {
	return utils.SSH2_HOME
}
