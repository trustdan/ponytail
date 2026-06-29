// Command ponytail is the single binary that replaces the JS lifecycle hooks
// (and, in later phases, the CI guards, MCP server, and statusline). Subcommand
// dispatch via stdlib only. ponytail: cobra only if `doctor` UX outgrows this.
package main

import (
	"fmt"
	"io"
	"os"

	ponytail "github.com/DietrichGebert/ponytail"
	"github.com/DietrichGebert/ponytail/internal/content"
	"github.com/DietrichGebert/ponytail/internal/gen"
	"github.com/DietrichGebert/ponytail/internal/hooks"
	"github.com/DietrichGebert/ponytail/internal/mode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ponytail <activate|track|subagent|instructions|default-mode|set-default|uninstall|gen|check|version>")
		os.Exit(0)
	}
	switch os.Args[1] {
	case "activate":
		hooks.Activate()
	case "track":
		in, _ := io.ReadAll(os.Stdin)
		hooks.Track(in)
	case "subagent":
		hooks.Subagent()
	case "instructions":
		// Raw ruleset for the active mode — what the in-process shims inject.
		fmt.Print(content.Instructions(arg(2)))
	case "default-mode":
		fmt.Println(mode.DefaultMode())
	case "set-default":
		// Persist the default level; print the normalized value, or nothing on invalid.
		if m := mode.WriteDefaultMode(arg(2)); m != "" {
			fmt.Println(m)
		}
	case "uninstall":
		hooks.Uninstall()
	case "gen":
		runGen()
	case "check":
		runCheck()
	case "version":
		fmt.Println(ponytail.Version)
	default:
		// Unknown subcommand must not fail a hook invocation.
	}
}

// arg returns os.Args[i] or "" if absent, so a missing mode argument falls back
// to the default inside content/mode rather than panicking.
func arg(i int) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return ""
}

func runGen() {
	changed, err := gen.Generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
	for _, p := range changed {
		fmt.Println("wrote", p)
	}
	if len(changed) == 0 {
		fmt.Println("all generated artifacts up to date")
	}
}

func runCheck() {
	problems, err := gen.Check()
	if err != nil {
		fmt.Fprintln(os.Stderr, "check:", err)
		os.Exit(1)
	}
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
	fmt.Println("generated artifacts, rule invariants, and version all in sync")
}
