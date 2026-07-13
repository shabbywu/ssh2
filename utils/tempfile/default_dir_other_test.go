//go:build !windows

package tempfile

import (
	"os"
	"testing"
)

func TestNewManagerUsesSystemTempDirByDefault(t *testing.T) {
	if got := NewManager("").dir; got != os.TempDir() {
		t.Fatalf("default temp dir = %q, want %q", got, os.TempDir())
	}
}
