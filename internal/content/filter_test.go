package content

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// The JS reference these were diffed against is gone (Phase 3); this is now a
// golden test guarding the canonical Go output against drift. Regenerate with
// `go test ./internal/content -update` after an intentional ruleset change.
var update = flag.Bool("update", false, "rewrite golden files")

func TestInstructionsGolden(t *testing.T) {
	// Every level the activate/subagent hooks can pass, plus the empty/garbage
	// inputs that must fall back to "full".
	for _, mode := range []string{"lite", "full", "ultra", "review", "empty", "bogus", "off"} {
		t.Run("mode="+mode, func(t *testing.T) {
			in := mode
			if in == "empty" {
				in = ""
			}
			got := Instructions(in)
			golden := filepath.Join("testdata", "instructions_"+mode+".golden")
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run with -update): %v", err)
			}
			if got != string(want) {
				t.Errorf("Instructions(%q) drifted from golden.\n--- got (%d bytes) ---\n%q", in, len(got), got)
			}
		})
	}
}
