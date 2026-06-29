// Package hooks is the Go port of the lifecycle hooks (activate/track/subagent)
// plus the per-host runtime from ponytail-runtime.js: flag-file state and the
// host-specific stdout shape each harness expects. The JSON shapes must match
// JS byte-for-byte — ordered structs preserve key order and SetEscapeHTML(false)
// stops Go escaping < > & that JS leaves raw.
package hooks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/DietrichGebert/ponytail/internal/mode"
)

const stateFile = ".ponytail-active"

func isCopilot() bool { return os.Getenv("COPILOT_PLUGIN_DATA") != "" }
func isCodex() bool   { return !isCopilot() && os.Getenv("PLUGIN_DATA") != "" }

func stateDir() string {
	if isCopilot() {
		return os.Getenv("COPILOT_PLUGIN_DATA")
	}
	if isCodex() {
		return os.Getenv("PLUGIN_DATA")
	}
	return mode.ClaudeDir()
}

func statePath() string { return filepath.Join(stateDir(), stateFile) }

// SetMode writes the activation flag (best-effort; callers ignore the error).
func SetMode(m string) error {
	p := statePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(m), 0o644)
}

func ClearMode() { _ = os.Remove(statePath()) }

// ReadMode returns the live mode, or "" (the JS null) when absent/empty.
func ReadMode() string {
	b, err := os.ReadFile(statePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

type hookSpecific struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type codexOut struct {
	SystemMessage string        `json:"systemMessage"`
	HookSpecific  *hookSpecific `json:"hookSpecificOutput,omitempty"`
}

type copilotOut struct {
	AdditionalContext string `json:"additionalContext,omitempty"`
}

type nativeSubOut struct {
	HookSpecific hookSpecific `json:"hookSpecificOutput"`
}

// render returns the exact bytes WriteOutput sends to stdout for the detected
// host. Pure function (host via env) so tests can diff it against the JS.
func render(event, m, context string) []byte {
	switch {
	case isCopilot():
		// Copilot reads additionalContext on SessionStart only; ignores it elsewhere.
		if event == "SessionStart" && context != "" {
			return marshal(copilotOut{AdditionalContext: context})
		}
		return marshal(copilotOut{})
	case isCodex():
		out := codexOut{SystemMessage: "PONYTAIL:" + strings.ToUpper(m)}
		if context != "" {
			out.HookSpecific = &hookSpecific{HookEventName: event, AdditionalContext: context}
		}
		return marshal(out)
	default:
		// Native Claude: SessionStart accepts raw stdout, SubagentStart needs the
		// hookSpecificOutput JSON form or the context is dropped.
		if event == "SubagentStart" {
			return marshal(nativeSubOut{HookSpecific: hookSpecific{HookEventName: event, AdditionalContext: context}})
		}
		return []byte(context)
	}
}

// WriteOutput emits the host-shaped hook payload to stdout.
func WriteOutput(event, m, context string) { os.Stdout.Write(render(event, m, context)) }

// marshal matches JSON.stringify: no trailing newline, and < > & left raw.
func marshal(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return bytes.TrimRight(b.Bytes(), "\n")
}
