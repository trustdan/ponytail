package gen

import (
	"os"
	"strings"
	"testing"

	ponytail "github.com/DietrichGebert/ponytail"
)

// Tests run from the package dir; gen uses paths relative to the repo root.
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// The committed tree must already be what gen would write — this is the drift
// guard the three deleted JS scripts used to be, collapsed into one assertion.
func TestCommittedArtifactsInSync(t *testing.T) {
	problems, err := Check()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}
}

// Swapping the version field rewrites only that field and preserves the rest of
// the manifest verbatim — including CRLF line endings (package.json is CRLF).
func TestVersionFieldRewrite(t *testing.T) {
	raw, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatal(err)
	}
	stale := strings.Replace(string(raw), `"version": "`+ponytail.Version+`"`, `"version": "0.0.0"`, 1)
	if stale == string(raw) {
		t.Fatal("setup: version field not found")
	}
	got := versionField.ReplaceAllString(stale, `"version": "`+ponytail.Version+`"`)
	if got != string(raw) {
		t.Error("rewrite did not restore the manifest byte-for-byte")
	}
}

// The YAML version swap (Hermes' plugin.yaml) must rewrite only the value and
// leave the rest of the file — including its line endings — byte-for-byte.
func TestYAMLVersionRewrite(t *testing.T) {
	raw, err := os.ReadFile("plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	stale := yamlVersionField.ReplaceAllString(string(raw), "version: 0.0.0")
	if stale == string(raw) {
		t.Fatal("setup: version line not found")
	}
	got := yamlVersionField.ReplaceAllString(stale, "version: "+ponytail.Version)
	if got != string(raw) {
		t.Error("rewrite did not restore plugin.yaml byte-for-byte")
	}
}

// A reworded rule that drops a safety carve-out must trip Check, not slip through.
func TestRuleInvariantGuards(t *testing.T) {
	if !strings.Contains(norm(ponytail.AgentsMarkdown()), ruleInvariants[0]) {
		t.Fatalf("invariant %q not actually in AGENTS.md — guard is vacuous", ruleInvariants[0])
	}
}
