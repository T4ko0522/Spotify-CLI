package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeForTest(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := NewRootCommand("v9.8.7")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), err
}

func TestRootCommandVersion(t *testing.T) {
	out, err := executeForTest(t, "--version")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "v9.8.7") {
		t.Fatalf("version output = %q, want injected version", out)
	}
}

func TestHelpDoesNotRequireSpotifyConfig(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	for _, args := range [][]string{
		{"--help"},
		{"init", "--help"},
		{"play", "--help"},
		{"volume", "--help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, err := executeForTest(t, args...)
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if !strings.Contains(out, "Usage:") {
				t.Fatalf("help output = %q, want usage", out)
			}
		})
	}
}
