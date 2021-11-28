package utils

import (
	"log"
	"os"
	"path/filepath"
)

var SSH2_HOME string

func init() {
	var ok bool
	SSH2_HOME, ok = os.LookupEnv("SSH2_HOME")
	if !ok {
		HOME, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		SSH2_HOME = filepath.Join(HOME, ".ssh", "ssh2")
	}
}
