package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"

	"ssh2/utils"
)

func TestWrapperForGOOS(t *testing.T) {
	tests := []struct {
		goos string
		want wrapperAsset
	}{
		{goos: "windows", want: powerShellWrapper},
		{goos: "linux", want: shellWrapper},
		{goos: "darwin", want: shellWrapper},
		{goos: "freebsd", want: shellWrapper},
	}

	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			got := wrapperForGOOS(test.goos)
			if got.filename != test.want.filename {
				t.Fatalf("wrapper filename = %q, want %q", got.filename, test.want.filename)
			}
			if !bytes.Equal(got.content, test.want.content) {
				t.Fatal("wrapper content does not match selected asset")
			}
		})
	}
}

func TestPowerShellWrapperKeepsGo2S(t *testing.T) {
	content := string(wrapperps1)
	for _, expected := range []string{
		"function global:go2s",
		`[Alias("d")]`,
		`"--direct"`,
		"& ssh2 @ssh2Arguments",
		`& ssh2 get --kind Session --template "{{ .Tag }}"`,
		"Register-ArgumentCompleter -CommandName go2s -ParameterName Arguments",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("PowerShell wrapper missing %q", expected)
		}
	}
}

func TestGetPowerShellWrapperPathCommandInstallsWrapper(t *testing.T) {
	originalHome := utils.SSH2_HOME
	utils.SSH2_HOME = t.TempDir()
	t.Cleanup(func() {
		utils.SSH2_HOME = originalHome
	})

	app := cli.NewApp()
	app.Commands = []*cli.Command{powerShellWrapperPathCommand}
	if err := app.Run([]string{"ssh2", "get-wrapper-dot-ps1"}); err != nil {
		t.Fatal(err)
	}

	installedPath := filepath.Join(utils.SSH2_HOME, "ssh2_wrapper.ps1")
	content, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, wrapperps1) {
		t.Fatal("installed PowerShell wrapper does not match embedded wrapper")
	}
}

func TestInstallCommandUsesPlatformWrapper(t *testing.T) {
	originalHome := utils.SSH2_HOME
	utils.SSH2_HOME = t.TempDir()
	t.Cleanup(func() {
		utils.SSH2_HOME = originalHome
	})

	app := cli.NewApp()
	app.Commands = []*cli.Command{installCommand}
	if err := app.Run([]string{"ssh2", "install-ssh2-auto-complete"}); err != nil {
		t.Fatal(err)
	}

	wrapper := wrapperForGOOS(runtime.GOOS)
	content, err := os.ReadFile(filepath.Join(utils.SSH2_HOME, wrapper.filename))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, wrapper.content) {
		t.Fatal("installed platform wrapper does not match embedded wrapper")
	}
}

func TestPowerShellWrapperBehavior(t *testing.T) {
	powerShellPaths := findPowerShellExecutables()
	if len(powerShellPaths) == 0 {
		t.Skip("PowerShell is not installed")
	}

	tempDir := t.TempDir()
	wrapperPath := filepath.Join(tempDir, "ssh2_wrapper.ps1")
	if err := os.WriteFile(wrapperPath, wrapperps1, 0600); err != nil {
		t.Fatal(err)
	}

	harness := fmt.Sprintf(`
$script:calls = @()
function global:ssh2 {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    $script:calls += ,($Arguments -join '|')
    if ($Arguments.Count -gt 0 -and $Arguments[0] -eq 'get') {
        'session-1'
        'session-2'
    }
}

. '%s'

foreach ($completionCommand in @('go2s ses', 'go2s -d ses', 'go2s --direct ses')) {
    $completionMatches = (TabExpansion2 $completionCommand $completionCommand.Length).CompletionMatches.CompletionText
    if (($completionMatches -join '|') -ne 'session-1|session-2') {
        throw "$completionCommand completion returned: $($completionMatches -join '|')"
    }
}

$listed = @(go2s)
if (($listed -join '|') -ne 'session-1|session-2') {
    throw "go2s list returned: $($listed -join '|')"
}

$script:calls = @()
go2s -d session-1
if (($script:calls -join ',') -ne 'get|--kind|Session|--template|{{ .Tag }},login|--direct|session-1') {
    throw "go2s -d called: $($script:calls -join ',')"
}

$script:calls = @()
go2s --direct session-2
if (($script:calls -join ',') -ne 'get|--kind|Session|--template|{{ .Tag }},login|--direct|session-2') {
    throw "go2s --direct called: $($script:calls -join ',')"
}

$script:calls = @()
$previousErrorActionPreference = $ErrorActionPreference
$ErrorActionPreference = 'SilentlyContinue'
go2s missing
$ErrorActionPreference = $previousErrorActionPreference
if (($script:calls -join ',') -ne 'get|--kind|Session|--template|{{ .Tag }}') {
    throw "go2s missing called: $($script:calls -join ',')"
}
`, strings.ReplaceAll(wrapperPath, "'", "''"))
	harnessPath := filepath.Join(tempDir, "wrapper_test.ps1")
	if err := os.WriteFile(harnessPath, []byte(harness), 0600); err != nil {
		t.Fatal(err)
	}

	for _, powerShellPath := range powerShellPaths {
		powerShellPath := powerShellPath
		t.Run(filepath.Base(powerShellPath), func(t *testing.T) {
			arguments := []string{
				"-NoLogo",
				"-NoProfile",
				"-NonInteractive",
			}
			if runtime.GOOS == "windows" {
				arguments = append(arguments, "-ExecutionPolicy", "Bypass")
			}
			arguments = append(arguments, "-File", harnessPath)
			command := exec.Command(powerShellPath, arguments...)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("PowerShell wrapper smoke test failed: %v\n%s", err, output)
			}
		})
	}
}

func findPowerShellExecutables() []string {
	names := []string{"pwsh"}
	if runtime.GOOS == "windows" {
		names = []string{"powershell.exe", "pwsh"}
	}

	var paths []string
	seen := map[string]bool{}
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err != nil || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	return paths
}
