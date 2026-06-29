// Package doctor wires safe local config that host installers cannot own.
package doctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DietrichGebert/ponytail/internal/mode"
)

var bom = []byte{0xEF, 0xBB, 0xBF}

// Run detects the local Claude config and installs/refreshes Ponytail's
// statusLine command. Lifecycle hook wiring stays in the plugin manifests.
func Run(w io.Writer) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	_, msg, err := ConfigureClaudeStatusline(mode.ClaudeDir(), exe)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "claude:", msg)
	return nil
}

// ConfigureClaudeStatusline writes ~/.claude/settings.json only when no
// unrelated statusLine exists, or when the existing statusLine already belongs
// to ponytail and should be refreshed to the current binary path.
func ConfigureClaudeStatusline(claudeDir, exe string) (bool, string, error) {
	if !mode.IsShellSafe(exe) {
		return false, "statusline skipped; binary path needs manual shell quoting", nil
	}

	settings := filepath.Join(claudeDir, "settings.json")
	raw, err := os.ReadFile(settings)
	if err != nil && !os.IsNotExist(err) {
		return false, "", err
	}
	raw = bytes.TrimPrefix(raw, bom)

	data := map[string]json.RawMessage{}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &data); err != nil {
			return false, "", fmt.Errorf("%s is not valid JSON", settings)
		}
	}

	command := `"` + exe + `" statusline`
	next := struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}{"command", command}
	nextRaw, _ := json.Marshal(next)

	if oldRaw, ok := data["statusLine"]; ok {
		var old struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		}
		if json.Unmarshal(oldRaw, &old) != nil || !isPonytailStatusline(old.Command) {
			return false, "statusline left unchanged; non-ponytail statusLine already exists", nil
		}
		if old.Type == "command" && old.Command == command {
			return false, "statusline already configured", nil
		}
	}

	data["statusLine"] = nextRaw
	out, _ := json.MarshalIndent(data, "", "  ")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		return false, "", err
	}
	if err := os.WriteFile(settings, append(out, '\n'), 0o644); err != nil {
		return false, "", err
	}
	return true, "statusline configured in " + settings, nil
}

func isPonytailStatusline(command string) bool {
	return strings.Contains(command, "ponytail-statusline") ||
		(strings.Contains(command, "ponytail") && strings.Contains(command, "statusline"))
}
