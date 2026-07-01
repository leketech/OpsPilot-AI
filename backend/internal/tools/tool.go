// Package tools defines the Tool interface and Registry used by the agent
// orchestrator to dispatch tool calls decided by the LLM.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
)

// Tool is the interface every OpsPilot integration must implement.
//
// Name and Description are surfaced to the LLM verbatim so it can decide
// when and how to call this tool. Parameters is a JSON Schema that tells the
// LLM what arguments it may pass.
//
// Execute receives whatever the caller provides as input:
//   - When called by the Registry (from an LLM tool_call): input is map[string]any
//     decoded from the JSON arguments string.
//   - When called directly in tests: input may be a typed struct specific to
//     the tool, making unit tests readable without JSON round-trips.
//
// The return value (any) is converted to a string by the Registry before it is
// sent back to the LLM as a tool result message. Tools can return a plain
// string for human-readable output, or any JSON-serialisable struct for typed
// output.
type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Execute(ctx context.Context, input any) (any, error)
}

// Registry maps tool names to implementations and converts them to the
// ToolDefinition format expected by the Qwen client.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry builds a Registry pre-populated with the supplied tools.
func NewRegistry(ts ...Tool) *Registry {
	r := &Registry{tools: make(map[string]Tool, len(ts))}
	for _, t := range ts {
		r.tools[t.Name()] = t
	}
	return r
}

// Register adds or replaces a tool in the registry.
func (r *Registry) Register(t Tool) { r.tools[t.Name()] = t }

// Definitions returns every registered tool in the format the LLM expects.
func (r *Registry) Definitions() []qwen.ToolDefinition {
	defs := make([]qwen.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, qwen.ToolDefinition{
			Type: "function",
			Function: qwen.FunctionDef{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}
	return defs
}

// Execute dispatches a tool call from the LLM. rawArgs is the JSON-encoded
// argument string the LLM placed in its tool_call message. The result is
// converted to a string and returned as the tool message content.
func (r *Registry) Execute(ctx context.Context, name, rawArgs string) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %q", name)
	}

	// Decode the JSON arguments. The resulting map[string]any is passed as
	// input any, which tools receive and type-assert internally.
	var input any
	if err := json.Unmarshal([]byte(rawArgs), &input); err != nil {
		return "", fmt.Errorf("parse args for %q: %w", name, err)
	}

	result, err := t.Execute(ctx, input)
	if err != nil {
		return "", err
	}
	return stringify(result), nil
}

// stringify converts any tool result into a string the LLM can read.
// Strings and byte slices are returned as-is; everything else is JSON-encoded.
func stringify(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	}
}

// ── Helpers for tool implementations ─────────────────────────────────────────

// StrArg safely extracts a string value from an args map.
func StrArg(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}

// IntArg safely extracts an int value from an args map (float64 from JSON).
func IntArg(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok && v > 0 {
		return int(v)
	}
	return defaultVal
}
