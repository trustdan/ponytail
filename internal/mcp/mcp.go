// Package mcp serves Ponytail's ruleset over MCP stdio. The old Node server
// exposed one prompt and one read-only tool; this keeps that surface in stdlib
// JSON-RPC so the binary remains the single runtime.
package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	ponytail "github.com/DietrichGebert/ponytail"
	"github.com/DietrichGebert/ponytail/internal/content"
	"github.com/DietrichGebert/ponytail/internal/mode"
)

var servedModes = map[string]bool{"lite": true, "full": true, "ultra": true}

type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type callParams struct {
	Name      string `json:"name"`
	Arguments struct {
		Mode string `json:"mode"`
	} `json:"arguments"`
}

type promptParams struct {
	Name      string `json:"name"`
	Arguments struct {
		Mode string `json:"mode"`
	} `json:"arguments"`
}

// ResolveMode maps MCP's public mode argument to one of the served intensities.
// "off", "review", junk, and empty fall back to the configured default, then full.
func ResolveMode(requested string) string {
	if m := mode.Normalize(requested); servedModes[m] {
		return m
	}
	if m := mode.Normalize(mode.DefaultMode()); servedModes[m] {
		return m
	}
	return mode.Default
}

func buildInstructions(requested string) (string, string) {
	m := ResolveMode(requested)
	return m, content.Instructions(m)
}

// Run handles MCP JSON-RPC messages from r and writes framed responses to w.
func Run(r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	for {
		b, err := readMessage(br)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var msg message
		if err := json.Unmarshal(b, &msg); err != nil {
			writeResponse(w, nil, nil, &rpcError{-32700, "Parse error"})
			continue
		}
		if msg.ID == nil {
			continue
		}
		result, e := handle(msg)
		writeResponse(w, msg.ID, result, e)
	}
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handle(msg message) (any, *rpcError) {
	switch msg.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"prompts": map[string]any{},
				"tools":   map[string]any{},
			},
			"serverInfo": map[string]any{"name": "ponytail", "version": ponytail.Version},
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "prompts/list":
		return map[string]any{"prompts": []any{map[string]any{
			"name":        "ponytail",
			"title":       "Ponytail mode",
			"description": "Lazy senior dev instructions: YAGNI, stdlib first, the smallest correct change.",
			"arguments": []any{map[string]any{
				"name":        "mode",
				"description": "Ponytail intensity: lite, full, or ultra. Omit for the configured default.",
				"required":    false,
			}},
		}}}, nil
	case "prompts/get":
		var p promptParams
		_ = json.Unmarshal(msg.Params, &p)
		if p.Name != "ponytail" {
			return nil, &rpcError{-32602, "Unknown prompt"}
		}
		_, instructions := buildInstructions(p.Arguments.Mode)
		return map[string]any{"messages": []any{map[string]any{
			"role": "user",
			"content": map[string]any{
				"type": "text",
				"text": instructions,
			},
		}}}, nil
	case "tools/list":
		return map[string]any{"tools": []any{map[string]any{
			"name":         "ponytail_instructions",
			"title":        "Ponytail instructions",
			"description":  "Return the Ponytail ruleset for the given intensity (lite, full, or ultra).",
			"inputSchema":  modeArgSchema(),
			"outputSchema": outputSchema(),
			"annotations":  map[string]any{"readOnlyHint": true, "openWorldHint": false},
		}}}, nil
	case "tools/call":
		var p callParams
		_ = json.Unmarshal(msg.Params, &p)
		if p.Name != "ponytail_instructions" {
			return nil, &rpcError{-32602, "Unknown tool"}
		}
		m, instructions := buildInstructions(p.Arguments.Mode)
		structured := map[string]any{"mode": m, "instructions": instructions}
		return map[string]any{
			"content":           []any{map[string]any{"type": "text", "text": instructions}},
			"structuredContent": structured,
		}, nil
	default:
		return nil, &rpcError{-32601, "Method not found"}
	}
}

func modeArgSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{"mode": map[string]any{
			"type":        "string",
			"enum":        []string{"lite", "full", "ultra"},
			"description": "Ponytail intensity: lite, full, or ultra. Omit for the configured default.",
		}},
	}
}

func outputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"mode":         map[string]any{"type": "string"},
		"instructions": map[string]any{"type": "string"},
	}}
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if !strings.HasPrefix(strings.ToLower(line), "content-length:") {
		return []byte(line), nil
	}
	parts := strings.SplitN(line, ":", 2)
	n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}
	for {
		h, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimRight(h, "\r\n") == "" {
			break
		}
	}
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	return b, err
}

func writeResponse(w io.Writer, id any, result any, e *rpcError) {
	resp := map[string]any{"jsonrpc": "2.0", "id": id}
	if e != nil {
		resp["error"] = e
	} else {
		resp["result"] = result
	}
	b, _ := json.Marshal(resp)
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(b), b)
}

func decodeFrames(b []byte) ([]map[string]any, error) {
	var out []map[string]any
	r := bufio.NewReader(bytes.NewReader(b))
	for {
		msg, err := readMessage(r)
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(msg, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
}
