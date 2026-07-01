// Package qwen provides the types, client, and service used to communicate
// with Alibaba's Qwen LLM via its OpenAI-compatible chat completions API.
package qwen

import (
	"encoding/json"
	"fmt"
)

// ── Conversation ──────────────────────────────────────────────────────────────

// Message is a single turn in a conversation.
// ToolCalls, ToolCallID, and Name are populated only in tool-calling flows.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolResultMessage builds the message that carries a tool's output back to the LLM.
func ToolResultMessage(toolCallID, result string) Message {
	return Message{Role: "tool", Content: result, ToolCallID: toolCallID}
}

// ── Tool calling ──────────────────────────────────────────────────────────────

// ToolDefinition describes a callable tool to the LLM (OpenAI function format).
type ToolDefinition struct {
	Type     string      `json:"type"` // always "function"
	Function FunctionDef `json:"function"`
}

// FunctionDef is the function schema embedded in a ToolDefinition.
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

// ToolCall is what the LLM returns when it wants to invoke a tool.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function ToolCallFunc `json:"function"`
}

// ToolCallFunc holds the function name and JSON-encoded argument string.
type ToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded string
}

// ── Usage tracking ────────────────────────────────────────────────────────────

// Usage reports token consumption for a single API call.
// Accumulate these across agents to track pipeline cost.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ── Response types ────────────────────────────────────────────────────────────

// CompletionResult is returned by Client.Chat and Client.CompleteJSON.
type CompletionResult struct {
	Content string
	Usage   Usage
}

// AgentResponse is returned by Client.CompleteWithTools.
// Call HasToolCalls to decide whether to execute tools or treat Content as the
// final response.
type AgentResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string // "stop" | "tool_calls"
	Usage        Usage
}

// HasToolCalls returns true when the LLM wants to invoke one or more tools.
func (r *AgentResponse) HasToolCalls() bool {
	return r.FinishReason == "tool_calls" && len(r.ToolCalls) > 0
}

// AssistantMessage converts this response into a Message suitable for
// appending to the conversation history before tool results are added.
func (r *AgentResponse) AssistantMessage() Message {
	return Message{
		Role:      "assistant",
		Content:   r.Content,
		ToolCalls: r.ToolCalls,
	}
}

// ── Error type ────────────────────────────────────────────────────────────────

// APIError is returned when Qwen responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("qwen [%d]: %s", e.StatusCode, e.Message)
}

func (e *APIError) retryable() bool {
	return e.StatusCode == 429 || e.StatusCode >= 500
}
