// Package content is the Go port of hooks/ponytail-instructions.js: it filters
// the embedded ruleset down to the active intensity level and assembles the
// "PONYTAIL MODE ACTIVE" payload the hooks inject. Output must stay byte-for-byte
// identical to getPonytailInstructions() — the golden test enforces it.
package content

import (
	"regexp"
	"strings"

	ponytail "github.com/DietrichGebert/ponytail"
	"github.com/DietrichGebert/ponytail/internal/mode"
)

// review (and any future independent mode) is documented by its own skill, not
// the intensity-filtered ruleset.
var independentModes = map[string]bool{"review": true}

var (
	frontmatterRe  = regexp.MustCompile(`(?s)^---.*?---\s*`)
	tableLabelRe   = regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|`)
	exampleLabelRe = regexp.MustCompile(`^-\s*([^:]+):\s*`)
)

// filterForMode keeps every line except an intensity-table row or worked-example
// line keyed to a *different* level. A bullet whose label is not a mode name
// (e.g. "No unrequested abstractions: …") is a normal rule and stays verbatim.
func filterForMode(body, m string) string {
	eff := mode.Normalize(m)
	if eff == "" {
		eff = mode.Default
	}
	// JS splits on /\r?\n/ and joins with \n, so all output is LF. Normalize
	// first; the frontmatter strip and filter then match regardless of the
	// embedded file's line endings.
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = frontmatterRe.ReplaceAllString(body, "")

	lines := strings.Split(body, "\n")
	out := lines[:0]
	for _, line := range lines {
		if mm := tableLabelRe.FindStringSubmatch(line); mm != nil {
			if lm := mode.Normalize(strings.TrimSpace(mm[1])); lm != "" && lm != eff {
				continue
			}
		}
		if mm := exampleLabelRe.FindStringSubmatch(line); mm != nil {
			if lm := mode.Normalize(strings.TrimSpace(mm[1])); lm != "" && lm != eff {
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// Instructions ports getPonytailInstructions(mode).
func Instructions(m string) string {
	configured := mode.NormalizePersisted(m)
	if configured == "" {
		configured = mode.Default
	}
	if independentModes[configured] {
		return "PONYTAIL MODE ACTIVE — level: " + configured +
			". Behavior defined by /ponytail-" + configured + " skill."
	}
	eff := mode.Normalize(configured)
	if eff == "" {
		eff = mode.Default
	}
	return "PONYTAIL MODE ACTIVE — level: " + eff + "\n\n" +
		filterForMode(ponytail.SkillMarkdown(), eff)
}
