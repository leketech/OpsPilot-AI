package qwen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"github.com/leketech/OpsPilot-AI/backend/internal/config"
)

const maxAttempts = 3

// Client is the raw HTTP layer. It knows nothing about agents, incidents, or
// prompts — it only sends messages to Qwen and returns responses.
//
// Agents that need high-level methods (AnalyzeIncident, etc.) should use
// Service, which wraps Client and adds business logic.
type Client struct {
	http  *resty.Client
	model string
}

// New creates a Client from application config.
// All HTTP concerns — auth, timeout, retry — are configured here so that
// every agent gets identical, consistent behaviour.
func New(cfg *config.Config) *Client {
	r := resty.New().
		SetBaseURL(cfg.QwenBaseURL).
		SetTimeout(60 * time.Second).
		SetHeader("Authorization", "Bearer "+cfg.QwenAPIKey).
		SetHeader("Content-Type", "application/json").
		// Retry config: resty will retry up to (maxAttempts-1) additional times.
		// First attempt + 2 retries = 3 total, matching the original behaviour.
		SetRetryCount(maxAttempts - 1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(4 * time.Second). // gives exponential-ish backoff: 1s → 2s → 4s cap
		// Only retry on transient errors (rate-limit or server-side failures).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			return r.StatusCode() == http.StatusTooManyRequests || r.StatusCode() >= 500
		}).
		// Log each retry attempt so engineers can see the backoff in action.
		AddRetryHook(func(r *resty.Response, err error) {
			log.WithFields(log.Fields{
				"status": statusCode(r),
				"error":  err,
			}).Warn("qwen: retrying request")
		})

	return &Client{http: r, model: cfg.QwenModel}
}

// ── Public API ────────────────────────────────────────────────────────────────

// Chat is the generic building block for all higher-level methods.
// It sends a conversation and returns the plain-text reply.
func (c *Client) Chat(ctx context.Context, messages []Message) (*CompletionResult, error) {
	resp, err := c.call(ctx, messages, nil, false)
	if err != nil {
		return nil, err
	}
	return &CompletionResult{Content: resp.Content, Usage: resp.Usage}, nil
}

// CompleteJSON calls Qwen in JSON-object mode and unmarshals the reply into out.
// The prompt must instruct the model to respond with a valid JSON object.
func (c *Client) CompleteJSON(ctx context.Context, messages []Message, out any) (*Usage, error) {
	resp, err := c.call(ctx, messages, nil, true)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(resp.Content), out); err != nil {
		return nil, fmt.Errorf("parse json response: %w (raw: %.200s)", err, resp.Content)
	}
	return &resp.Usage, nil
}

// CompleteWithTools sends messages with tool definitions. The LLM may respond
// by requesting tool calls (HasToolCalls() == true) or produce a final answer.
func (c *Client) CompleteWithTools(ctx context.Context, messages []Message, tools []ToolDefinition) (*AgentResponse, error) {
	return c.call(ctx, messages, tools, false)
}

// ── Wire types (HTTP-layer only) ──────────────────────────────────────────────

type completionRequest struct {
	Model          string           `json:"model"`
	Messages       []Message        `json:"messages"`
	Tools          []ToolDefinition `json:"tools,omitempty"`
	ToolChoice     string           `json:"tool_choice,omitempty"`
	ResponseFormat *responseFormat  `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// ── Internal ──────────────────────────────────────────────────────────────────

// call is the single internal request path. resty handles retry/backoff.
func (c *Client) call(ctx context.Context, messages []Message, tools []ToolDefinition, jsonMode bool) (*AgentResponse, error) {
	req := completionRequest{Model: c.model, Messages: messages}
	if len(tools) > 0 {
		req.Tools = tools
		req.ToolChoice = "auto"
	}
	if jsonMode {
		req.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	start := time.Now()
	var resp completionResponse

	result, err := c.http.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/chat/completions")

	duration := time.Since(start)

	if err != nil {
		log.WithFields(log.Fields{
			"model":    c.model,
			"duration": duration,
			"error":    err,
		}).Error("qwen: request failed")
		return nil, fmt.Errorf("qwen http: %w", err)
	}

	if result.StatusCode() != http.StatusOK {
		msg := fmt.Sprintf("status %d", result.StatusCode())
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, &APIError{StatusCode: result.StatusCode(), Message: msg}
	}

	if len(resp.Choices) == 0 {
		return nil, &APIError{StatusCode: result.StatusCode(), Message: "no choices in response"}
	}

	choice := resp.Choices[0]
	out := &AgentResponse{
		Content:      choice.Message.Content,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
	}
	if resp.Usage != nil {
		out.Usage = Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	log.WithFields(log.Fields{
		"model":             c.model,
		"duration_ms":       duration.Milliseconds(),
		"prompt_tokens":     out.Usage.PromptTokens,
		"completion_tokens": out.Usage.CompletionTokens,
		"total_tokens":      out.Usage.TotalTokens,
		"finish_reason":     out.FinishReason,
	}).Debug("qwen: request succeeded")

	return out, nil
}

func statusCode(r *resty.Response) int {
	if r == nil {
		return 0
	}
	return r.StatusCode()
}
