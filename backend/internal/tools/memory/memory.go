// Package memory implements the Memory tool and session store.
// It serves two roles:
//  1. As a Tool — the LLM calls it during the ReAct loop to surface similar
//     past incidents from the historical index.
//  2. As a Store — the orchestrator calls it directly to load/append session
//     history and record completed incidents for future retrieval.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

const (
	sessionPrefix  = "agent:session:"
	historyKey     = "agent:history:all"
	historyPrefix  = "agent:history:rec:"
	sessionTTL     = 24 * time.Hour
	maxTurns       = 20
	maxSimilar     = 5
	historyScanMax = 30
)

// Store is both a Tool (for the LLM to call) and a session/history manager
// (for the orchestrator to use directly).
type Store struct {
	redis *redis.Client
}

func New(rdb *redis.Client) *Store {
	return &Store{redis: rdb}
}

// ── Tool interface ────────────────────────────────────────────────────────────

func (s *Store) Name() string { return "memory_query" }

func (s *Store) Description() string {
	return "Search the historical incident database for similar past incidents. Returns root causes and resolutions that may be relevant to the current incident."
}

func (s *Store) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"service":   {"type": "string", "description": "Service name to find similar incidents for"},
			"namespace": {"type": "string", "description": "Kubernetes namespace to scope the search"}
		},
		"required": ["service"]
	}`)
}

// Execute is called by the Registry during the ReAct loop. It returns a
// formatted string of similar past incidents for the LLM to reason over.
// Each record now surfaces evidence, the fix that worked, human decision, and outcome
// so the LLM can directly inform its recommendation with historical precedent.
func (s *Store) Execute(ctx context.Context, input any) (any, error) {
	args, _ := input.(map[string]any)
	service, _ := args["service"].(string)
	namespace, _ := args["namespace"].(string)

	inc := &models.Incident{Service: service, Namespace: namespace}
	similar, err := s.findSimilar(ctx, inc)
	if err != nil {
		return "", err
	}
	if len(similar) == 0 {
		return fmt.Sprintf("No similar historical incidents found for service %q.", service), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d similar past incident(s):\n\n", len(similar))
	for _, r := range similar {
		fmt.Fprintf(&b, "[%s] %s (%s/%s)\n", r.OccurredAt, r.Title, r.Namespace, r.Service)
		fmt.Fprintf(&b, "  Root cause:      %s (confidence: %.0f%%)\n", r.RootCause, r.Confidence)
		fmt.Fprintf(&b, "  Evidence:        %s\n", r.Evidence)
		fmt.Fprintf(&b, "  Fix applied:     %s\n", r.Fix)
		fmt.Fprintf(&b, "  Human decision:  %s\n", r.HumanDecision)
		fmt.Fprintf(&b, "  Outcome:         %s\n\n", r.Outcome)
	}
	return b.String(), nil
}

// ── Session history (orchestrator direct use) ─────────────────────────────────

// LoadSession returns the stored conversation messages for a session.
func (s *Store) LoadSession(ctx context.Context, sessionID string) ([]qwen.Message, error) {
	raw, err := s.redis.Get(ctx, sessionPrefix+sessionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session load: %w", err)
	}
	var msgs []qwen.Message
	if err := json.Unmarshal(raw, &msgs); err != nil {
		return nil, fmt.Errorf("session decode: %w", err)
	}
	return msgs, nil
}

// AppendSession adds messages to the session history, capping at maxTurns.
func (s *Store) AppendSession(ctx context.Context, sessionID string, msgs ...qwen.Message) error {
	history, err := s.LoadSession(ctx, sessionID)
	if err != nil {
		return err
	}
	history = append(history, msgs...)
	if len(history) > maxTurns {
		history = history[len(history)-maxTurns:]
	}
	raw, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("session encode: %w", err)
	}
	return s.redis.Set(ctx, sessionPrefix+sessionID, raw, sessionTTL).Err()
}

// Clear deletes the session history for a given session.
func (s *Store) Clear(ctx context.Context, sessionID string) error {
	return s.redis.Del(ctx, sessionPrefix+sessionID).Err()
}

// ── Historical incident index (orchestrator direct use) ───────────────────────

// Record saves a completed incident report to the historical index.
// It accepts the full pipeline Report so it can persist evidence, fix,
// human decision, and outcome — enabling richer future retrievals.
func (s *Store) Record(ctx context.Context, report *models.Report) error {
	inc := report.Incident
	if inc == nil || inc.ID == "" {
		return nil
	}

	outcome := "rejected"
	if report.Review != nil && report.Review.Approved {
		outcome = "approved"
	}
	humanDecision := ""
	if report.Review != nil {
		humanDecision = report.Review.FinalVerdict
	}
	evidence := ""
	if report.Investigation != nil {
		evidence = report.Investigation.KeyFindings
	}
	var fix string
	var confidence float64
	var rootCause string
	if report.Recommendation != nil {
		fix = report.Recommendation.Action
		confidence = report.Recommendation.Confidence
		rootCause = report.Recommendation.RootCause
	}

	entry := models.IncidentRecord{
		ID:            inc.ID,
		Title:         inc.Title,
		Service:       inc.Service,
		Namespace:     inc.Namespace,
		RootCause:     rootCause,
		Confidence:    confidence,
		Evidence:      evidence,
		Fix:           fix,
		HumanDecision: humanDecision,
		Outcome:       outcome,
		OccurredAt:    time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("record encode: %w", err)
	}
	if err := s.redis.Set(ctx, historyPrefix+inc.ID, raw, 0).Err(); err != nil {
		return fmt.Errorf("record store: %w", err)
	}
	return s.redis.ZAdd(ctx, historyKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: inc.ID,
	}).Err()
}

// ── Private ───────────────────────────────────────────────────────────────────

func (s *Store) findSimilar(ctx context.Context, inc *models.Incident) ([]models.IncidentRecord, error) {
	ids, err := s.redis.ZRevRange(ctx, historyKey, 0, int64(historyScanMax-1)).Result()
	if err == redis.Nil || len(ids) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("history scan: %w", err)
	}

	var matched, fallback []models.IncidentRecord
	for _, id := range ids {
		if id == inc.ID {
			continue
		}
		raw, err := s.redis.Get(ctx, historyPrefix+id).Bytes()
		if err != nil {
			continue
		}
		var rec models.IncidentRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			continue
		}
		if rec.Service == inc.Service || rec.Namespace == inc.Namespace {
			matched = append(matched, rec)
		} else {
			fallback = append(fallback, rec)
		}
		if len(matched) >= maxSimilar {
			break
		}
	}

	if len(matched) > 0 {
		return matched, nil
	}
	if len(fallback) > maxSimilar {
		fallback = fallback[:maxSimilar]
	}
	return fallback, nil
}
