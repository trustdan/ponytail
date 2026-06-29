// Package statusline renders the Claude Code mode badge that used to live in
// the old hooks/ponytail-statusline scripts.
package statusline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DietrichGebert/ponytail/internal/mode"
)

const stateFile = ".ponytail-active"

// Output returns the ANSI badge for the active Claude mode, or "" when off.
func Output() string {
	b, err := os.ReadFile(filepath.Join(mode.ClaudeDir(), stateFile))
	if err != nil {
		return ""
	}
	m := strings.TrimSpace(string(b))
	if m == "" || m == "full" {
		return "\033[38;5;108m[PONYTAIL]\033[0m"
	}
	return "\033[38;5;108m[PONYTAIL:" + strings.ToUpper(m) + "]\033[0m"
}

// Print writes Output without a trailing newline, matching the old scripts.
func Print() { fmt.Print(Output()) }
