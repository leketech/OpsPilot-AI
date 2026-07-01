package qwen

import (
	"context"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/config"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// Service wraps the Client and exposes high-level, intent-named methods.
// Callers never deal with HTTP, authentication, retries, or prompt assembly —
// those concerns live in Client and prompts.go respectively.
//
// Relationship:
//
//	Agent/Handler
//	     │
//	     ▼
//	  Service          ← intent-named methods, prompt assembly
//	     │
//	     ▼
//	  Client           ← HTTP, auth, retry
//	     │
//	     ▼
//	 Qwen Cloud API
type Service struct {
	client *Client
}

// NewService creates a Service backed by a freshly configured Client.
func NewService(cfg *config.Config) *Service {
	return &Service{client: New(cfg)}
}

// NewServiceFromClient creates a Service from a pre-built Client.
// Useful in tests where the Client is configured with a mock base URL.
func NewServiceFromClient(c *Client) *Service {
	return &Service{client: c}
}

// Client exposes the underlying Client for callers that need direct access
// (e.g. the orchestrator's ReAct loop, which calls CompleteWithTools).
func (s *Service) Client() *Client { return s.client }

// ── Generic method ────────────────────────────────────────────────────────────

// Chat is the generic building block. Every higher-level method calls this.
// It sends a conversation turn and returns the plain-text reply.
func (s *Service) Chat(ctx context.Context, messages []Message) (string, error) {
	result, err := s.client.Chat(ctx, messages)
	if err != nil {
		return "", err
	}
	return result.Content, nil
}

// ── Incident analysis ─────────────────────────────────────────────────────────

// AnalyzeIncident performs a one-shot SRE analysis of an incident.
// It assembles the SRE system prompt and incident context, calls Qwen once,
// and returns a structured Recommendation.
//
// Use this for quick analysis. For the full 5-agent pipeline (with tool
// calling, revision loops, and memory) use the orchestrator.
func (s *Service) AnalyzeIncident(ctx context.Context, inc *models.Incident) (*models.Recommendation, error) {
	msgs := oneShot(inc)
	var rec models.Recommendation
	if _, err := s.client.CompleteJSON(ctx, msgs, &rec); err != nil {
		return nil, fmt.Errorf("service.AnalyzeIncident: %w", err)
	}
	return &rec, nil
}

// ── Prompt assembly (private, centralised here) ───────────────────────────────

// oneShot builds a single-turn prompt for quick incident analysis.
// The SRE system prompt sets the persona; the user turn provides incident
// context and asks for a structured JSON recommendation.
func oneShot(inc *models.Incident) []Message {
	return []Message{
		{Role: "system", Content: sreSysPrompt},
		{Role: "user", Content: buildOneShotContent(inc)},
	}
}

func buildOneShotContent(inc *models.Incident) string {
	return fmt.Sprintf(`%s

Using the incident details above, respond with a JSON object with EXACTLY these fields:
{
  "root_cause":        "most probable root cause in one sentence",
  "reasoning":         "step-by-step explanation citing specific signals from the data",
  "confidence":        75.0,
  "action":            "numbered remediation steps, each with a rollback instruction",
  "risk":              "risk level and specific risks of the recommended action",
  "similar_incidents": 0
}`,
		buildIncidentBlock(inc))
}
