package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
	"ssh2/models"
	"ssh2/utils"
)

func TestApplyCommandKeepsCreateAlias(t *testing.T) {
	if len(applyCommand.Aliases) != 1 || applyCommand.Aliases[0] != "create" {
		t.Fatalf("apply aliases = %#v", applyCommand.Aliases)
	}
}

func TestLoginCommandKeepsTagFlag(t *testing.T) {
	var foundTag, foundDryRun, foundDirect bool
	for _, flag := range execCommand.Flags {
		if stringFlag, ok := flag.(*cli.StringFlag); ok && stringFlag.Name == "tag" {
			foundTag = true
		}
		if boolFlag, ok := flag.(*cli.BoolFlag); ok && boolFlag.Name == "dry-run" {
			foundDryRun = true
		}
		if boolFlag, ok := flag.(*cli.BoolFlag); ok && boolFlag.Name == "direct" {
			foundDirect = true
		}
	}
	if !foundTag {
		t.Fatal("login command missing tag flag")
	}
	if !foundDryRun {
		t.Fatal("login command missing dry-run flag")
	}
	if !foundDirect {
		t.Fatal("login command missing direct flag")
	}
}

func TestShellQuoteCommand(t *testing.T) {
	got := shellQuoteCommand([]string{"ssh", "-i", "/tmp/key file", "root@example.com", "echo 'ok'"})
	want := `ssh -i '/tmp/key file' root@example.com 'echo '"'"'ok'"'"''`
	if got != want {
		t.Fatalf("quoted command = %q, want %q", got, want)
	}
}

func TestShellQuoteCommandWithRaw(t *testing.T) {
	got := shellQuoteCommandWithRaw([]string{"ssh", "-i", "/tmp/key file", "root@example.com"}, map[string]string{
		"/tmp/key file": `"$ssh2_key_file"`,
	})
	want := `ssh -i "$ssh2_key_file" root@example.com`
	if got != want {
		t.Fatalf("quoted command = %q, want %q", got, want)
	}
}

func TestWrapperKeepsGo2S(t *testing.T) {
	content := string(wrappersh)
	for _, expected := range []string{
		"function go2s",
		`ssh2 login ${direct} "${ssh_tag}"`,
		`--direct`,
		`ssh2 get --kind Session --template "{{ .Tag }}"`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("wrapper missing %q", expected)
		}
	}
}

func TestHandleLoginSignalsStopIsIdempotent(t *testing.T) {
	_, stopSignals := handleLoginSignals(nil)
	stopSignals()
	stopSignals()
}

func TestForwardInteractiveInterruptsStopIsIdempotent(t *testing.T) {
	stopSignals := forwardInteractiveInterrupts(nil)
	stopSignals()
	stopSignals()
}

func TestGetWrapperPathCommandInstallsWrapper(t *testing.T) {
	originalHome := utils.SSH2_HOME
	utils.SSH2_HOME = t.TempDir()
	t.Cleanup(func() {
		utils.SSH2_HOME = originalHome
	})

	app := cli.NewApp()
	app.Commands = []*cli.Command{wrapperPathCommand}
	if err := app.Run([]string{"ssh2", "get-wrapper-dot-sh"}); err != nil {
		t.Fatal(err)
	}

	installedPath := filepath.Join(utils.SSH2_HOME, "ssh2_wrapper.sh")
	content, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(wrappersh) {
		t.Fatal("installed wrapper content does not match embedded wrapper")
	}
}

func TestApplyCommandReturnsYamlErrors(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "bad-yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("kind: AuthMethod\nspec:\n  expect_for_password: password:\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{applyCommand}
	err = app.Run([]string{"ssh2", "apply", "-f", file.Name()})
	if err == nil {
		t.Fatal("expected invalid YAML to return an error")
	}
}

func TestApplyCommandResolvesSiblingYamlRef(t *testing.T) {
	dir := t.TempDir()
	suffix := filepath.Base(dir)
	clientName := "cmd-test-client-" + suffix
	authName := "cmd-test-auth-" + suffix
	serverName := "cmd-test-server-" + suffix
	sessionTag := "cmd-test-session-" + suffix

	clientFile := filepath.Join(dir, "client.yaml")
	if err := os.WriteFile(clientFile, []byte(fmt.Sprintf(`
kind: ClientConfig
spec:
  name: %s
  user: tester
  auth:
    spec:
      name: %s
      type: INTERACTIVE_PASSWORD
      expect_for_password: "password:"
`, clientName, authName)), 0600); err != nil {
		t.Fatal(err)
	}

	sessionFile := filepath.Join(dir, "session.yaml")
	if err := os.WriteFile(sessionFile, []byte(fmt.Sprintf(`
kind: Session
spec:
  tag: %s
  name: %s
  plugins:
    - kind: EXPECT
      args:
        expect: Password
        send: "secret"
  client:
    ref:
      field: name
      value: %s
  server:
    spec:
      name: %s
      host: 127.0.0.1
      port: 22
`, sessionTag, serverName, clientName, serverName)), 0600); err != nil {
		t.Fatal(err)
	}

	app := cli.NewApp()
	app.Commands = []*cli.Command{applyCommand}
	if err := app.Run([]string{"ssh2", "apply", "-f", sessionFile}); err != nil {
		t.Fatal(err)
	}

	session, err := models.GetByField[models.Session]("Session", "tag", sessionTag)
	if err != nil {
		t.Fatal(err)
	}
	if session.ClientConfigId == 0 {
		t.Fatal("session was not linked to sibling client config")
	}
}
