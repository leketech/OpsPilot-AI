// Package prometheus provides the prometheus_query tool.
// Returns realistic mock PromQL results for development and demo use.
// Replace Execute() with a real Prometheus HTTP API call for production.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/leketech/OpsPilot-AI/backend/internal/tools"
)

// Tool implements the prometheus_query tool.
type Tool struct{}

func New() *Tool { return &Tool{} }

func (t *Tool) Name() string { return "prometheus_query" }

func (t *Tool) Description() string {
	return "Execute a PromQL query against the Prometheus metrics store. Use this to check CPU, memory, error rates, latency histograms, and saturation metrics for the affected service."
}

func (t *Tool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query":    {"type": "string",  "description": "PromQL expression to evaluate"},
			"duration": {"type": "string",  "description": "Look-back window, e.g. '5m', '1h', '24h' (default: '15m')"}
		},
		"required": ["query"]
	}`)
}

func (t *Tool) Execute(_ context.Context, input any) (any, error) {
	args, _ := input.(map[string]any)
	query := tools.StrArg(args, "query")
	duration := tools.StrArg(args, "duration")
	if duration == "" {
		duration = "15m"
	}
	return mockPrometheusResult(query, duration), nil
}

func mockPrometheusResult(query, duration string) string {
	now := time.Now().UTC()
	q := strings.ToLower(query)

	switch {
	case strings.Contains(q, "cpu"):
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT (instant vector):
  {namespace="payments", pod="payments-api-7d9f8b-xk2p9"}   0.97   ← 97%% CPU (threshold 80%%)
  {namespace="payments", pod="payments-api-7d9f8b-m3rq7"}   0.94
  {namespace="payments", pod="payments-api-7d9f8b-n8wt4"}   0.91
Alert: CPUThrottling FIRING for 22m`, query, duration, now.Format(time.RFC3339))

	case strings.Contains(q, "memory") || strings.Contains(q, "mem"):
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT (instant vector):
  {namespace="payments", pod="payments-api-7d9f8b-xk2p9"}   1.93e9   ← 1.93 GB / 2 GB limit (96.5%%)
  {namespace="payments", pod="payments-api-7d9f8b-m3rq7"}   1.78e9
Alert: MemoryPressure FIRING — OOMKill risk HIGH`, query, duration, now.Format(time.RFC3339))

	case strings.Contains(q, "error") || strings.Contains(q, "5xx"):
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT (range vector — rate over %s):
  {namespace="payments", service="payments-api"}   0.12   ← 12%% error rate (SLO <0.1%%)
Alert: ErrorRateSLOBreach FIRING for 18m`, query, duration, now.Format(time.RFC3339), duration)

	case strings.Contains(q, "latency") || strings.Contains(q, "duration") || strings.Contains(q, "p99"):
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT (p99 latency):
  {namespace="payments", service="payments-api"}   3.24   ← 3240ms (SLO 500ms)
  trend: increasing since 14:12 UTC`, query, duration, now.Format(time.RFC3339))

	case strings.Contains(q, "goroutine"):
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT:
  {namespace="payments", pod="payments-api-7d9f8b-m3rq7"}   4200   ← 4200 goroutines (baseline 120)
  {namespace="payments", pod="payments-api-7d9f8b-n8wt4"}   3980
Trend: exponential growth starting 14:10 UTC (correlates with v2.14.1 rollout)`, query, duration, now.Format(time.RFC3339))

	default:
		return fmt.Sprintf(`Query: %s [%s]  evaluated at %s
RESULT: No data returned — metric may not exist or namespace is incorrect.
Suggestion: verify the metric name with 'prometheus_query' using a simpler label selector.`, query, duration, now.Format(time.RFC3339))
	}
}
