# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Ponytail is a "lazy senior dev" ruleset packaged for 16 different AI agent harnesses
(Claude Code, Codex, Copilot, Gemini, OpenCode, pi, Devin, Hermes, Kiro, Cursor,
Windsurf, Cline, etc.). There is almost no application logic — the "product" is:

1. **The rule text**, shipped in many host-specific copies that must stay in sync.
2. **Go lifecycle hooks** that inject the rules and track the active mode per session.
3. **Skills** (`/ponytail`, `/ponytail-review`, …) that host harnesses expose as commands.

Most changes are edits to text/manifests, not features. The CI guards below are the
load-bearing part: they exist because the same content lives in 6+ files at once.

## Commands

```bash
npm test                          # node --test tests/*.test.js, then pi-extension tests
node --test tests/hooks.test.js   # run a single test file (drives the committed binary)
go test ./...                     # Go hook + generation tests (golden-compared)
go run ./cmd/ponytail gen         # regenerate every duplicated host artifact from the embedded source
go run ./cmd/ponytail check       # drift + version guard (also run in CI)
go run ./cmd/ponytail doctor      # configure local Claude statusLine wiring
sh scripts/build-bin.sh           # rebuild the committed per-platform binaries in bin/
goreleaser release --snapshot --clean  # verify the release build matrix
```

`go` must be on PATH (the binary, the generation guard, and the gen/check tooling all live
in it). `node` is needed only to run the JS host shims and their tests, not the hooks
themselves. The correctness benchmark spawns Python (`python3` then `python`); CSV checks
need `pandas`.

## The two invariants that break CI

The duplicated artifacts are now **generated, not hand-synced.** One embedded source
(`AGENTS.md`, `skills/`, and the `Version` constant in `ponytail.go`) drives everything;
`ponytail gen` writes the copies and `ponytail check` is the CI guard. Edit the source,
run `go run ./cmd/ponytail gen`, commit the result. The two failure modes `check` catches:

**Rule text is duplicated and must match.** `AGENTS.md` holds the canonical compact
ruleset; its body is regenerated (modulo host frontmatter) into `.cursor/rules/ponytail.mdc`,
`.windsurf/rules/ponytail.md`, `.clinerules/ponytail.md`, `.agents/rules/ponytail.md`,
`.github/copilot-instructions.md`, `.kiro/steering/ponytail.md`, and the `.openclaw/skills/`.
`skills/ponytail/SKILL.md` is the longer *runtime* source of truth — not byte-compared,
but a set of load-bearing phrases (the safety carve-outs, the test reflex, etc.) must
appear verbatim in both it and `AGENTS.md` (the `ruleInvariants` list in `internal/gen`).

**Version is single-sourced from `ponytail.go`.** `gen` writes the `Version` constant into
the `version` field of all six manifests (`.claude-plugin/`, `.codex-plugin/`, `.devin-plugin/`,
`.github/plugin/` plugin.json, `gemini-extension.json`, `package.json`);
`check` fails if any drifts, and on a release-tag CI run if the tag ≠ `Version`. A release bumps
the one constant, not six files by hand.

## Hook runtime (the `ponytail` binary)

The lifecycle hooks are now subcommands of the Go binary (`internal/hooks`, `internal/mode`,
`internal/content`); there is no Node hook runtime. State is a flag file `.ponytail-active`
written under the host's config dir (`mode.ClaudeDir()`, or `PLUGIN_DATA`/`COPILOT_PLUGIN_DATA`
for Codex/Copilot). Absent flag = off.

- `ponytail activate` — SessionStart: writes the flag, emits the ruleset filtered to the
  active level as hidden context, and nudges statusline setup if unconfigured.
- `ponytail track` — UserPromptSubmit (reads the payload on stdin): detects `/ponytail <level>`
  and `stop ponytail` / `normal mode` (whole-message only) to change the live mode.
- `ponytail subagent` — SubagentStart: re-injects the mode into subagents.
- `ponytail instructions <mode>` / `default-mode` / `set-default <mode>` — the pieces the
  in-process JS shims (OpenCode `.mjs`, `pi-extension/`) call instead of
  re-implementing; they keep only host wiring + trivial inline mode parsing.
- `ponytail statusline` — prints the active-mode badge for Claude Code statusLine.
- `ponytail mcp` — serves the `ponytail` prompt and `ponytail_instructions` tool over MCP stdio.
- `ponytail doctor` — configures local Claude Code statusLine wiring when no user-owned statusLine is present.
- `ponytail uninstall` — removes the flag, config file, and the statusLine entry it added.

Host detection (Codex/Copilot/native) and the per-host stdout shape live in `internal/hooks`
(`render`), golden-tested in `output_test.go`. Each host wires the subcommands via its own
manifest: `hooks/claude-codex-hooks.json` (Claude + Codex) and `hooks/copilot-hooks.json` call
the binary through `bin/ponytail` (POSIX launcher) or `bin/ponytail-windows-amd64.exe`.

**Binaries are committed and generated.** `bin/` holds five per-platform builds produced by
`sh scripts/build-bin.sh` (reproducible: `-trimpath`, `CGO_ENABLED=0`, `-s -w`). `.goreleaser.yaml` defines the release archives for the same OS/arch matrix. Edit Go, rerun
the script, commit `bin/`. CI rebuilds and `git diff --exit-code bin/` fails on a stale binary —
the third drift guard, alongside generated artifacts and version.

## Conventions

- `ponytail:` code comments mark deliberate simplifications and their upgrade path — they are
  harvested by `/ponytail-debt`. Keep that convention when adding shortcuts.
- Adding a new host ecosystem: register its version file in `versionFiles` and its rule copy
  in `ruleCopies` (with the host frontmatter) in `internal/gen/gen.go`, or `gen`/`check` won't cover it.
- `AGENTS.md` applies to agents working on *this* repo too — the ladder governs changes here.
