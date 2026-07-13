//go:build !windows

package tempfile

import "os"

func defaultDir() string {
	return os.TempDir()
}
