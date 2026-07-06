package plugins

import (
	"errors"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestParseExpectSteps(t *testing.T) {
	plugin, err := Parse(gjson.Parse(`{
		"kind": "EXPECT",
		"args": {
			"steps": [
				{"expect": "jump$", "send": "ssh target\r"},
				{"expect": "password:", "send": "secret\r"}
			]
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}

	expectPlugin, ok := plugin.(*ExpectPlugin)
	if !ok {
		t.Fatalf("plugin type = %T", plugin)
	}
	if len(expectPlugin.Steps) != 2 {
		t.Fatalf("steps length = %d", len(expectPlugin.Steps))
	}
	if expectPlugin.Steps[0].Expect != "jump$" || expectPlugin.Steps[1].Send != "secret\r" {
		t.Fatalf("unexpected steps: %#v", expectPlugin.Steps)
	}
}

func TestParseExpectRawReturnsError(t *testing.T) {
	_, err := Parse(gjson.Parse(`{"kind":"EXPECT","args":{"raw":["expect \"x\""]}}`))
	if err == nil {
		t.Fatal("expected raw to return an error")
	}
	if !strings.Contains(err.Error(), "EXPECT.raw") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseUnknownPluginReturnsError(t *testing.T) {
	_, err := Parse(gjson.Parse(`{"kind":"SSH_WETERM","args":{}}`))
	if err == nil {
		t.Fatal("expected unknown plugin to return an error")
	}
	if !strings.Contains(err.Error(), "SSH_WETERM") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpectErrorIncludesOutput(t *testing.T) {
	err := expectError("Last login", "ssh: connect to host 10.202.0.79 port 32200: Network is unreachable\r\n", errors.New("EOF"))
	for _, expected := range []string{"Last login", "EOF", "Network is unreachable"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("error %q missing %q", err.Error(), expected)
		}
	}
}
