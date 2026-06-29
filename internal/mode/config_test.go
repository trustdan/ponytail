package mode

import "testing"

// The deactivation matcher is whole-message only: "stop ponytail" as a command
// turns ponytail off, but "add a normal mode toggle" must not.
func TestIsDeactivation(t *testing.T) {
	on := []string{"stop ponytail", "Stop Ponytail", "normal mode", " normal mode! ", "stop ponytail."}
	off := []string{"add a normal mode toggle", "please stop ponytail now", "", "ponytail", "stop"}
	for _, s := range on {
		if !IsDeactivation(s) {
			t.Errorf("IsDeactivation(%q) = false, want true", s)
		}
	}
	for _, s := range off {
		if IsDeactivation(s) {
			t.Errorf("IsDeactivation(%q) = true, want false", s)
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		" FULL ": "full", "Lite": "lite", "ultra": "ultra", "off": "off",
		"review": "", "bogus": "", "": "",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
	if NormalizeConfig("review") != "review" {
		t.Error("NormalizeConfig should accept review")
	}
}

// IsShellSafe gates the statusline setup snippet (issue #200): ordinary install
// paths pass, paths carrying shell metacharacters are rejected so they never get
// embedded in a shell command.
func TestIsShellSafe(t *testing.T) {
	safe := []string{
		`C:\Users\x\.claude\plugins\ponytail\bin\ponytail-windows-amd64.exe`,
		`/home/u/.claude/plugins/ponytail/bin/ponytail`,
	}
	unsafe := []string{`/tmp/a"&calc.exe&"/x.sh`, `/tmp/$(calc)/x.sh`, `/tmp/a;rm -rf/x.sh`}
	for _, p := range safe {
		if !IsShellSafe(p) {
			t.Errorf("IsShellSafe(%q) = false, want true", p)
		}
	}
	for _, p := range unsafe {
		if IsShellSafe(p) {
			t.Errorf("IsShellSafe(%q) = true, want false", p)
		}
	}
}
