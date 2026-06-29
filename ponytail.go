// Package ponytail embeds the canonical content shipped inside the binary, so
// the hooks run with no repo files on disk. This package is the single source
// the host artifacts are generated from: `ponytail gen` writes them, `ponytail
// check` guards their drift (see internal/gen).
package ponytail

import "embed"

// Version is the shipped release and the one source the version manifests are
// generated from. `ponytail gen` writes it into all of them; `ponytail check`
// fails if any drift, or if a release tag doesn't match.
const Version = "4.8.4"

// AGENTS.md is the canonical compact ruleset; skills/ holds the runtime SKILL.md
// source of truth. The compact host copies and OpenClaw skills derive from these.
//
//go:embed AGENTS.md skills
var assetsFS embed.FS

func asset(p string) string {
	b, _ := assetsFS.ReadFile(p)
	return string(b)
}

// SkillMarkdown returns the raw ponytail SKILL.md (frontmatter + body).
func SkillMarkdown() string { return asset("skills/ponytail/SKILL.md") }

// SkillMarkdownNamed returns the raw SKILL.md for skills/<name>/.
func SkillMarkdownNamed(name string) string { return asset("skills/" + name + "/SKILL.md") }

// AgentsMarkdown returns the canonical AGENTS.md (the compact ruleset source).
func AgentsMarkdown() string { return asset("AGENTS.md") }
