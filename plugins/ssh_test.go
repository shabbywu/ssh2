package plugins

import (
	"reflect"
	"testing"
)

func TestSSHCommandIncludesTimeoutOptions(t *testing.T) {
	cmd := sshCommand(32200, "user@example.com", "-i", "/tmp/key")
	expected := []string{
		"ssh",
		"-tt",
		"-o", "ConnectTimeout=10",
		"-o", "ConnectionAttempts=1",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ServerAliveInterval=15",
		"-o", "ServerAliveCountMax=2",
		"-i", "/tmp/key",
		"-p", "32200",
		"user@example.com",
	}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Fatalf("ssh args = %#v, want %#v", cmd.Args, expected)
	}
}

func TestSSHCommandIncludesKeyAuthOptions(t *testing.T) {
	keyArgs := append([]string{}, keyAuthSSHOptions...)
	keyArgs = append(keyArgs, "-i", "/tmp/key")
	cmd := sshCommand(22, "root@example.com", keyArgs...)

	for _, expected := range []string{
		"BatchMode=yes",
		"PreferredAuthentications=publickey",
		"PasswordAuthentication=no",
		"KbdInteractiveAuthentication=no",
		"IdentitiesOnly=yes",
		"IdentityAgent=none",
	} {
		if !containsArgValue(cmd.Args, expected) {
			t.Fatalf("ssh args %#v missing %q", cmd.Args, expected)
		}
	}
}

func TestSSHCommandIncludesPasswordAuthOptions(t *testing.T) {
	passwordArgs := append([]string{}, passwordAuthSSHOptions...)
	cmd := sshCommand(22, "root@example.com", passwordArgs...)

	for _, expected := range []string{
		"PreferredAuthentications=password,keyboard-interactive",
		"PubkeyAuthentication=no",
		"PasswordAuthentication=yes",
		"KbdInteractiveAuthentication=yes",
		"IdentitiesOnly=yes",
		"IdentityAgent=none",
		"ControlPath=none",
		"NumberOfPasswordPrompts=1",
	} {
		if !containsArgValue(cmd.Args, expected) {
			t.Fatalf("ssh args %#v missing %q", cmd.Args, expected)
		}
	}
}

func containsArgValue(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
