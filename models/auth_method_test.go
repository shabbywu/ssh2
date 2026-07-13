package models

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ssh2/utils"
	"ssh2/utils/crypto"
)

func TestAuthMethodEncryptedContent(t *testing.T) {
	auth := &AuthMethod{Type: AuthPassword, Content: "secret"}

	if err := auth.EncryptContent(); err != nil {
		t.Fatal(err)
	}
	if auth.Content == "secret" {
		t.Fatal("content was not encrypted")
	}
	if got, err := auth.DecryptedContent(); err != nil || got != "secret" {
		t.Fatalf("DecryptedContent() = %q, %v", got, err)
	}
}

func TestAuthMethodPlainTextContentStillReads(t *testing.T) {
	auth := &AuthMethod{Type: AuthPassword, Content: "legacy-go-plain-text"}

	got, err := auth.DecryptedContent()
	if err != nil {
		t.Fatal(err)
	}
	if got != "legacy-go-plain-text" {
		t.Fatalf("DecryptedContent() = %q", got)
	}
}

func TestAuthMethodPublishKeyContentTempFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		originalHome := utils.SSH2_HOME
		utils.SSH2_HOME = filepath.Join(t.TempDir(), ".ssh", "ssh2")
		if err := os.MkdirAll(utils.SSH2_HOME, 0700); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { utils.SSH2_HOME = originalHome })
	}

	auth := &AuthMethod{Name: "test-key", Type: AuthPublishKey, Content: "PRIVATE KEY"}
	if err := auth.EncryptContent(); err != nil {
		t.Fatal(err)
	}

	path, cleanup, err := auth.PublishKeyPath()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "PRIVATE KEY" {
		t.Fatalf("temp key content = %q", data)
	}
	if runtime.GOOS == "windows" && filepath.Dir(path) != utils.SSH2_HOME {
		t.Fatalf("temp key dir = %q, want %q", filepath.Dir(path), utils.SSH2_HOME)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if mode := info.Mode().Perm(); mode != 0600 {
			t.Fatalf("temp key mode = %o", mode)
		}
	}
}

func TestPythonCryptCredentialReturnsClearError(t *testing.T) {
	if _, err := crypto.Decrypt("crypt$payload"); err == nil {
		t.Fatal("expected crypt$ credential to return an error")
	}
}
