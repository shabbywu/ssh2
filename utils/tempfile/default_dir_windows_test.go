//go:build windows

package tempfile

import (
	"os"
	"path/filepath"
	"testing"

	"ssh2/utils"
)

func TestNewManagerUsesSSH2HomeByDefault(t *testing.T) {
	originalHome := utils.SSH2_HOME
	utils.SSH2_HOME = filepath.Join(t.TempDir(), ".ssh", "ssh2")
	t.Cleanup(func() { utils.SSH2_HOME = originalHome })

	if err := os.MkdirAll(utils.SSH2_HOME, 0700); err != nil {
		t.Fatal(err)
	}
	manager := NewManager("")
	file, err := manager.TempFile("private-key-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = file.Close()
		manager.Clean()
	})

	if got := filepath.Dir(file.Name()); got != utils.SSH2_HOME {
		t.Fatalf("temp key dir = %q, want %q", got, utils.SSH2_HOME)
	}
}
