package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func frame(s string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(s), s)
}

func TestResolveMode(t *testing.T) {
	t.Setenv("PONYTAIL_DEFAULT_MODE", "ultra")
	if got := ResolveMode("lite"); got != "lite" {
		t.Fatalf("valid mode = %q", got)
	}
	for _, input := range []string{"off", "review", "nonsense", ""} {
		if got := ResolveMode(input); got != "ultra" {
			t.Fatalf("ResolveMode(%q) = %q, want default ultra", input, got)
		}
	}

	t.Setenv("PONYTAIL_DEFAULT_MODE", "review")
	if got := ResolveMode(""); got != "full" {
		t.Fatalf("review default must fall back to full, got %q", got)
	}
}

func TestRunToolCall(t *testing.T) {
	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ponytail_instructions","arguments":{"mode":"ultra"}}}`
	var out bytes.Buffer
	if err := Run(strings.NewReader(frame(req)), &out); err != nil {
		t.Fatal(err)
	}
	responses, err := decodeFrames(out.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 1 {
		t.Fatalf("responses = %d, want 1", len(responses))
	}
	b, _ := json.Marshal(responses[0]["result"])
	var result struct {
		StructuredContent struct {
			Mode         string `json:"mode"`
			Instructions string `json:"instructions"`
		} `json:"structuredContent"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if result.StructuredContent.Mode != "ultra" {
		t.Fatalf("mode = %q", result.StructuredContent.Mode)
	}
	if !strings.Contains(result.StructuredContent.Instructions, "PONYTAIL MODE ACTIVE") {
		t.Fatal("instructions missing active-mode banner")
	}
}
