// Package gen is the Go port of the three CI guard scripts it replaces:
// check-rule-copies.js, build-openclaw-skills.js, and check-versions.js. The
// duplicated host artifacts are *generated* from the embedded source (AGENTS.md,
// skills/, and ponytail.Version) instead of hand-synced: `ponytail gen` writes
// them, `ponytail check` fails CI if a committed copy drifts.
package gen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ponytail "github.com/DietrichGebert/ponytail"
)

const homepage = "https://github.com/DietrichGebert/ponytail"

// Compact rule copies: same body as AGENTS.md (minus its trailing self-reference
// paragraph), each with its host-specific frontmatter. Empty frontmatter = body
// only. Edit a rule in AGENTS.md → `ponytail gen` rewrites every copy.
var ruleCopies = []struct{ path, frontmatter string }{
	{".cursor/rules/ponytail.mdc", "---\ndescription: Ponytail, lazy senior dev mode. Always pick the simplest solution that works.\nglobs:\nalwaysApply: true\n---"},
	{".kiro/steering/ponytail.md", "---\ntitle: Ponytail, lazy senior dev mode\ninclusion: always\n---"},
	{".windsurf/rules/ponytail.md", ""},
	{".clinerules/ponytail.md", ""},
	{".agents/rules/ponytail.md", ""},
	{".github/copilot-instructions.md", ""},
}

// OpenClaw skills: SKILL.md body copied verbatim from skills/<name>/, only the
// frontmatter rewritten. OpenClaw requires a single-line description under 160
// chars, so each ships a short one (the canonical descriptions are longer).
var openclawSkills = []struct{ name, desc string }{
	{"ponytail", "Lazy senior dev mode for any coding task (write, refactor, fix, review): YAGNI, stdlib first, no unrequested abstractions. Not for non-coding requests."},
	{"ponytail-review", "Review a diff for over-engineering. Finds what to delete: reinvented stdlib, needless deps, speculative abstractions. One line per finding."},
	{"ponytail-audit", "Audit the whole repo for over-engineering. A ranked list of what to delete, simplify, or replace with stdlib or native features."},
	{"ponytail-debt", "Harvest every ponytail: shortcut comment into one debt ledger, so deferrals get tracked instead of forgotten. One-shot report."},
	{"ponytail-gain", "Show ponytail measured impact as a scoreboard: less code, less cost, more speed, from the benchmark medians. One-shot display."},
	{"ponytail-help", "Quick reference for ponytail's modes, skills, and commands. One-shot display."},
}

// Files that declare the project version; gen rewrites each to ponytail.Version.
var versionFiles = []string{
	".claude-plugin/plugin.json",
	".codex-plugin/plugin.json",
	".devin-plugin/plugin.json",
	".github/plugin/plugin.json",
	"gemini-extension.json",
	"package.json",
}

// Same single-sourcing for YAML manifests (Hermes' plugin.yaml), which use an
// unquoted `version:` line the JSON regex above doesn't match.
var yamlVersionFiles = []string{"plugin.yaml"}

// Load-bearing phrases that must survive verbatim in both SKILL.md and AGENTS.md.
// ponytail: canary, not full equality — the bodies differ in length, so we pin
// the safety carve-outs and reflexes a reword could silently drop.
var ruleInvariants = []string{
	"naive heuristic",
	"ONE runnable check",
	"flimsier algorithm",
	"input validation at trust boundaries",
	"prevents data loss",
	"security",
	"accessibility",
	"Lazy code without its check is unfinished",
}

var (
	trailingSelfRef  = regexp.MustCompile(`(?s)\n\n\(Yes, this file also applies.*?\)$`)
	leadingFrontmat  = regexp.MustCompile(`(?s)^---\n.*?\n---\n?`)
	versionField     = regexp.MustCompile(`"version":\s*"([^"]*)"`)
	yamlVersionField = regexp.MustCompile(`(?m)^version:[ \t]*([^\r\n]*)`)
)

func norm(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

// canonicalBody is AGENTS.md minus its trailing self-reference paragraph, trimmed.
func canonicalBody() string {
	a := strings.TrimSpace(norm(ponytail.AgentsMarkdown()))
	return strings.TrimSpace(trailingSelfRef.ReplaceAllString(a, ""))
}

func ruleCopyContent(frontmatter string) string {
	body := canonicalBody()
	if frontmatter == "" {
		return body + "\n"
	}
	return frontmatter + "\n\n" + body + "\n"
}

func openclawContent(name, desc string) string {
	src := norm(ponytail.SkillMarkdownNamed(name))
	body := leadingFrontmat.ReplaceAllString(src, "")
	return fmt.Sprintf("---\nname: %s\ndescription: \"%s\"\nhomepage: %s\nlicense: MIT\n---\n", name, desc, homepage) + body
}

// artifact is one fully-generated file: its path and the LF content it should
// hold. Version manifests aren't artifacts — gen swaps only their version field
// in place (see version funcs below) to preserve their bytes and line endings.
type artifact struct {
	path    string
	content string
}

func artifacts() []artifact {
	var out []artifact
	for _, c := range ruleCopies {
		out = append(out, artifact{c.path, ruleCopyContent(c.frontmatter)})
	}
	for _, s := range openclawSkills {
		out = append(out, artifact{".openclaw/skills/" + s.name + "/SKILL.md", openclawContent(s.name, s.desc)})
	}
	return out
}

// versionValue extracts the first "version": "..." field, "" if absent.
func versionValue(raw string) string {
	m := versionField.FindStringSubmatch(raw)
	if m == nil {
		return ""
	}
	return m[1]
}

// Generate writes every artifact and rewrites version fields. Returns changed paths.
func Generate() ([]string, error) {
	var changed []string
	for _, a := range artifacts() {
		old, _ := os.ReadFile(a.path)
		// Compare normalized so a CRLF working copy isn't needlessly rewritten,
		// but write LF — that's how these files are committed.
		if norm(string(old)) == a.content {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(a.path), 0o755); err != nil {
			return changed, err
		}
		if err := os.WriteFile(a.path, []byte(a.content), 0o644); err != nil {
			return changed, err
		}
		changed = append(changed, a.path)
	}
	for _, p := range versionFiles {
		raw, err := os.ReadFile(p)
		if err != nil {
			return changed, err
		}
		// Swap only the field so the manifest's bytes and line endings survive.
		next := versionField.ReplaceAllString(string(raw), `"version": "`+ponytail.Version+`"`)
		if next == string(raw) {
			continue
		}
		if err := os.WriteFile(p, []byte(next), 0o644); err != nil {
			return changed, err
		}
		changed = append(changed, p)
	}
	for _, p := range yamlVersionFiles {
		raw, err := os.ReadFile(p)
		if err != nil {
			return changed, err
		}
		next := yamlVersionField.ReplaceAllString(string(raw), "version: "+ponytail.Version)
		if next == string(raw) {
			continue
		}
		if err := os.WriteFile(p, []byte(next), 0o644); err != nil {
			return changed, err
		}
		changed = append(changed, p)
	}
	return changed, nil
}

// Check reports every drift without writing: stale artifacts, version mismatches,
// missing rule invariants, and a release-tag mismatch. Empty slice = clean.
func Check() ([]string, error) {
	var problems []string
	for _, a := range artifacts() {
		old, err := os.ReadFile(a.path)
		if err != nil {
			problems = append(problems, a.path+" missing — run: ponytail gen")
			continue
		}
		if norm(string(old)) != a.content {
			problems = append(problems, a.path+" stale — run: ponytail gen")
		}
	}

	for _, p := range versionFiles {
		raw, err := os.ReadFile(p)
		if err != nil {
			problems = append(problems, p+" missing")
			continue
		}
		if v := versionValue(string(raw)); v != ponytail.Version {
			problems = append(problems, fmt.Sprintf("%s version %q != %s — run: ponytail gen", p, v, ponytail.Version))
		}
	}

	for _, p := range yamlVersionFiles {
		raw, err := os.ReadFile(p)
		if err != nil {
			problems = append(problems, p+" missing")
			continue
		}
		v := ""
		if m := yamlVersionField.FindStringSubmatch(string(raw)); m != nil {
			v = m[1]
		}
		if v != ponytail.Version {
			problems = append(problems, fmt.Sprintf("%s version %q != %s — run: ponytail gen", p, v, ponytail.Version))
		}
	}

	skill := norm(ponytail.SkillMarkdown())
	agents := norm(ponytail.AgentsMarkdown())
	for _, phrase := range ruleInvariants {
		if !strings.Contains(skill, phrase) {
			problems = append(problems, "skills/ponytail/SKILL.md missing rule invariant: "+phrase)
		}
		if !strings.Contains(agents, phrase) {
			problems = append(problems, "AGENTS.md missing rule invariant: "+phrase)
		}
	}

	// On a release-tag CI run the embedded version must equal the tag, else a
	// release was tagged without bumping ponytail.Version (and so the manifests).
	if os.Getenv("GITHUB_REF_TYPE") == "tag" {
		tag := strings.TrimPrefix(os.Getenv("GITHUB_REF_NAME"), "v")
		if tag != "" && tag != ponytail.Version {
			problems = append(problems, fmt.Sprintf("release tag %s does not match version %s; bump ponytail.Version before tagging", tag, ponytail.Version))
		}
	}
	return problems, nil
}
