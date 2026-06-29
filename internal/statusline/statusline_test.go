package statusline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)

	if got := Output(); got != "" {
		t.Fatalf("absent flag = %q, want empty", got)
	}

	flag := filepath.Join(dir, stateFile)
	if err := os.WriteFile(flag, []byte("full\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Output(); got != "\033[38;5;108m[PONYTAIL]\033[0m" {
		t.Fatalf("full badge = %q", got)
	}

	if err := os.WriteFile(flag, []byte("ultra\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Output(); got != "\033[38;5;108m[PONYTAIL:ULTRA]\033[0m" {
		t.Fatalf("ultra badge = %q", got)
	}
}
