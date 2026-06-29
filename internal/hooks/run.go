package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DietrichGebert/ponytail/internal/content"
	"github.com/DietrichGebert/ponytail/internal/mode"
)

// UTF-8 BOM some shells/editors prepend; strip before parsing.
var bom = []byte{0xEF, 0xBB, 0xBF}

// Activate ports ponytail-activate.js (SessionStart).
func Activate() {
	m := mode.DefaultMode()
	if m == "off" {
		ClearMode()
		ctx := "OK"
		if isCodex() || isCopilot() {
			ctx = ""
		}
		WriteOutput("SessionStart", "off", ctx)
		return
	}
	_ = SetMode(m) // best-effort; flag is advisory
	out := content.Instructions(m)
	if !isCodex() && !isCopilot() {
		out += statuslineNudge(mode.ClaudeDir())
	}
	WriteOutput("SessionStart", m, out)
}

// Uninstall ports scripts/uninstall.js: removes the state ponytail wrote outside
// the plugin's own files — the mode flag, the config file, and the statusLine
// entry it added to settings.json. Host uninstall commands remove plugin files;
// this cleans up what they can't see.
func Uninstall() {
	removeIfExists(filepath.Join(mode.ClaudeDir(), stateFile), "mode flag")
	removeIfExists(mode.ConfigPath(), "config file")

	settings := filepath.Join(mode.ClaudeDir(), "settings.json")
	b, err := os.ReadFile(settings)
	if err != nil {
		return
	}
	b = bytes.TrimPrefix(b, bom)
	// Decode into an ordered-agnostic map: we only need to drop one key and the
	// JS wrote it with json.Marshal indent-2 anyway, so key order isn't load-bearing.
	var s map[string]json.RawMessage
	if json.Unmarshal(b, &s) != nil {
		return
	}
	var sl struct {
		Command string `json:"command"`
	}
	// ponytail: substring-match the script name, then drop the whole statusLine
	// key — a combined statusline (caveman+ponytail) is removed wholesale. Parse
	// out only ponytail's part if combined statuslines become common.
	if raw, ok := s["statusLine"]; ok && json.Unmarshal(raw, &sl) == nil &&
		strings.Contains(sl.Command, "ponytail-statusline") {
		delete(s, "statusLine")
		out, _ := json.MarshalIndent(s, "", "  ")
		if os.WriteFile(settings, out, 0o644) == nil {
			fmt.Println("Removed ponytail statusLine entry from " + settings)
		}
	}
}

func removeIfExists(path, label string) {
	if err := os.Remove(path); err == nil {
		fmt.Println("Removed " + label + ": " + path)
	}
}

var cmdRe = regexp.MustCompile(`^[/@$]ponytail`)

// Track ports ponytail-mode-tracker.js (UserPromptSubmit). stdin is the raw
// hook payload ({"prompt": "..."}).
func Track(stdin []byte) {
	stdin = bytes.TrimPrefix(stdin, bom)
	var data struct {
		Prompt string `json:"prompt"`
	}
	if json.Unmarshal(stdin, &data) != nil {
		return
	}
	prompt := strings.ToLower(strings.TrimSpace(data.Prompt))

	if cmdRe.MatchString(prompt) {
		parts := strings.Fields(prompt)
		cmd := parts[0]
		if strings.HasPrefix(cmd, "@") || strings.HasPrefix(cmd, "$") {
			cmd = "/" + cmd[1:]
		}
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		m := ""
		switch cmd {
		case "/ponytail-review", "/ponytail:ponytail-review":
			m = "review"
		case "/ponytail", "/ponytail:ponytail":
			switch arg {
			case "lite", "full", "ultra", "off":
				m = arg
			default:
				m = mode.DefaultMode()
			}
		}

		if m != "" && m != "off" {
			_ = SetMode(m)
			WriteOutput("UserPromptSubmit", m, "PONYTAIL MODE CHANGED — level: "+m)
		} else if m == "off" {
			ClearMode()
			WriteOutput("UserPromptSubmit", "off", "PONYTAIL MODE OFF")
		}
	}

	if mode.IsDeactivation(prompt) {
		ClearMode()
		WriteOutput("UserPromptSubmit", "off", "PONYTAIL MODE OFF")
	}
}

// Subagent ports ponytail-subagent.js (SubagentStart): re-inject the ruleset
// into Task-spawned agents, which never see SessionStart context.
func Subagent() {
	m := ReadMode()
	if m == "" || m == "off" {
		return
	}
	WriteOutput("SubagentStart", m, content.Instructions(m))
}

// statuslineNudge returns "" if a statusLine is already configured, else the
// "\n\n…SETUP NEEDED…" suffix activate appends. ponytail: the command points at
// this binary's `statusline` subcommand (phase 4 wires it); the wording mirrors
// ponytail-activate.js so the agent's first-interaction offer is unchanged.
func statuslineNudge(claudeDir string) string {
	settings := filepath.Join(claudeDir, "settings.json")
	if b, err := os.ReadFile(settings); err == nil {
		b = bytes.TrimPrefix(b, bom)
		var s struct {
			StatusLine json.RawMessage `json:"statusLine"`
		}
		if json.Unmarshal(b, &s) == nil && len(s.StatusLine) > 0 {
			return ""
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if !mode.IsShellSafe(exe) {
		return "\n\n" +
			"STATUSLINE SETUP NEEDED: The ponytail plugin includes a statusline badge showing active mode. " +
			"Its install path contains characters unsafe to embed in a shell command, so configure it manually: " +
			"add a statusLine command of type \"command\" that runs the ponytail binary's `statusline` subcommand " +
			"to ~/.claude/settings.json, quoting/escaping the path for your shell. " +
			"Proactively offer to set this up for the user on first interaction."
	}

	command := "\"" + exe + "\" statusline"
	cmdJSON, _ := json.Marshal(command)
	snippet := "\"statusLine\": { \"type\": \"command\", \"command\": " + string(cmdJSON) + " }"
	return "\n\n" +
		"STATUSLINE SETUP NEEDED: The ponytail plugin includes a statusline badge showing active mode " +
		"(e.g. [PONYTAIL], [PONYTAIL:ULTRA]). It is not configured yet. " +
		"To enable, add this to ~/.claude/settings.json: " + snippet + " " +
		"Proactively offer to set this up for the user on first interaction."
}
