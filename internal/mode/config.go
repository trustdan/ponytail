// Package mode is the Go port of hooks/ponytail-config.js: default-mode
// resolution, the mode normalizers, config/Claude dir resolution, and the
// shell-safety and deactivation-command helpers. Kept byte-for-byte behavioral
// parity with the JS so the hooks behave identically host-by-host.
package mode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const Default = "full"

// VALID_MODES (config-level, includes the independent "review" mode) and
// RUNTIME_MODES (what the activation flag may hold) from ponytail-config.js.
var validModes = map[string]bool{"off": true, "lite": true, "full": true, "ultra": true, "review": true}
var runtimeModes = map[string]bool{"off": true, "lite": true, "full": true, "ultra": true}

// Normalize → a runtime mode (off/lite/full/ultra) or "" (the JS null).
func Normalize(m string) string {
	n := strings.ToLower(strings.TrimSpace(m))
	if runtimeModes[n] {
		return n
	}
	return ""
}

// NormalizeConfig also accepts "review".
func NormalizeConfig(m string) string {
	n := strings.ToLower(strings.TrimSpace(m))
	if validModes[n] {
		return n
	}
	return ""
}

// NormalizePersisted = Normalize || NormalizeConfig (matches the JS fallthrough).
func NormalizePersisted(m string) string {
	if n := Normalize(m); n != "" {
		return n
	}
	return NormalizeConfig(m)
}

var deactivationTrail = regexp.MustCompile(`[.!?\s]+$`)

// IsDeactivation: the whole message (case- and trailing-punctuation-insensitive)
// must be exactly "stop ponytail" or "normal mode". Matching it anywhere turned
// ponytail off mid-task for requests like "add a normal mode toggle".
func IsDeactivation(text string) bool {
	t := deactivationTrail.ReplaceAllString(strings.ToLower(strings.TrimSpace(text)), "")
	return t == "stop ponytail" || t == "normal mode"
}

// ponytail: allowlist of ordinary path characters — beats escaping every shell's
// metacharacters. A hostile install path falls back to manual statusline setup.
var shellSafe = regexp.MustCompile(`^[A-Za-z0-9 _.\-:/\\~]+$`)

func IsShellSafe(p string) bool { return shellSafe.MatchString(p) }

func home() string { h, _ := os.UserHomeDir(); return h }

func ConfigDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ponytail")
	}
	if runtime.GOOS == "windows" {
		app := os.Getenv("APPDATA")
		if app == "" {
			app = filepath.Join(home(), "AppData", "Roaming")
		}
		return filepath.Join(app, "ponytail")
	}
	return filepath.Join(home(), ".config", "ponytail")
}

func ConfigPath() string { return filepath.Join(ConfigDir(), "config.json") }

// ClaudeDir: CLAUDE_CONFIG_DIR overrides ~/.claude, matching Claude Code.
func ClaudeDir() string {
	if c := os.Getenv("CLAUDE_CONFIG_DIR"); c != "" {
		return c
	}
	return filepath.Join(home(), ".claude")
}

// DefaultMode resolves: PONYTAIL_DEFAULT_MODE → config.json defaultMode → "full".
func DefaultMode() string {
	if env := os.Getenv("PONYTAIL_DEFAULT_MODE"); env != "" {
		if v := strings.ToLower(env); validModes[v] {
			return v
		}
	}
	if b, err := os.ReadFile(ConfigPath()); err == nil {
		var c struct {
			DefaultMode string `json:"defaultMode"`
		}
		if json.Unmarshal(b, &c) == nil && c.DefaultMode != "" {
			if v := strings.ToLower(c.DefaultMode); validModes[v] {
				return v
			}
		}
	}
	return Default
}

// WriteDefaultMode persists the default level to config.json (used by the `mode`
// subcommand). Returns the normalized mode written, or "" if invalid.
func WriteDefaultMode(m string) string {
	n := NormalizeConfig(m)
	if n == "" {
		return ""
	}
	p := ConfigPath()
	if os.MkdirAll(filepath.Dir(p), 0o755) != nil {
		return ""
	}
	b, _ := json.MarshalIndent(struct {
		DefaultMode string `json:"defaultMode"`
	}{n}, "", "  ")
	if os.WriteFile(p, b, 0o644) != nil {
		return ""
	}
	return n
}
