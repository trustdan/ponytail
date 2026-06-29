package hooks

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// Golden test guarding the per-host stdout shapes (Phase 3 removed the JS
// reference these were diffed against). Regenerate with
// `go test ./internal/hooks -update` after an intentional shape change.
var update = flag.Bool("update", false, "rewrite golden files")

type host struct{ copilot, codex string } // env values; "" = unset

func setHost(t *testing.T, h host) {
	t.Helper()
	t.Setenv("COPILOT_PLUGIN_DATA", h.copilot)
	t.Setenv("PLUGIN_DATA", h.codex)
}

func TestRenderGolden(t *testing.T) {
	// Context with the bytes that expose marshaling bugs: < > & (must stay raw,
	// not HTML-escaped), an em dash and a newline (UTF-8 + control escape).
	tricky := "a <b> & c — d\ne"

	cases := []struct {
		name             string
		h                host
		event, mode, ctx string
	}{
		{"native_SessionStart", host{}, "SessionStart", "full", tricky},
		{"native_SubagentStart", host{}, "SubagentStart", "full", tricky},
		{"native_SessionStart-empty", host{}, "SessionStart", "full", ""},
		{"codex_SessionStart", host{codex: "x"}, "SessionStart", "full", tricky},
		{"codex_UserPromptSubmit-empty", host{codex: "x"}, "UserPromptSubmit", "ultra", ""},
		{"copilot_SessionStart", host{copilot: "x"}, "SessionStart", "lite", tricky},
		{"copilot_SubagentStart", host{copilot: "x"}, "SubagentStart", "full", "x"},
		{"copilot_SessionStart-empty", host{copilot: "x"}, "SessionStart", "full", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setHost(t, c.h)
			got := render(c.event, c.mode, c.ctx)
			golden := filepath.Join("testdata", "render_"+c.name+".golden")
			if *update {
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run with -update): %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("render(%s) drifted from golden.\n got: %q\nwant: %q", c.name, got, want)
			}
		})
	}
}
