package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigureClaudeStatusline(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "bin", "ponytail")

	changed, msg, err := ConfigureClaudeStatusline(filepath.Join(dir, ".claude"), exe)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatalf("changed = false, msg = %q", msg)
	}

	settings := filepath.Join(dir, ".claude", "settings.json")
	var data struct {
		StatusLine struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"statusLine"`
	}
	b, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if data.StatusLine.Type != "command" || data.StatusLine.Command != `"`+exe+`" statusline` {
		t.Fatalf("statusLine = %#v", data.StatusLine)
	}

	changed, _, err = ConfigureClaudeStatusline(filepath.Join(dir, ".claude"), exe)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second configure should be a no-op")
	}
}

func TestConfigureClaudeStatuslineLeavesUserStatusline(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claude, 0o755); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(claude, "settings.json")
	original := `{"statusLine":{"type":"command","command":"custom-status"}}`
	if err := os.WriteFile(settings, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, _, err := ConfigureClaudeStatusline(claude, filepath.Join(dir, "bin", "ponytail"))
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("custom statusline should be left alone")
	}
	if got := string(mustRead(t, settings)); got != original {
		t.Fatalf("settings changed:\n%s", got)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
