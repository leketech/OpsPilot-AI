package qwen

import (
	"fmt"
	"strings"

	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// sreSysPrompt is the shared persona injected into every agent conversation.
const sreSysPrompt = `You are a Senior Site Reliability Engineer with deep expertise in distributed systems, Kubernetes, cloud infrastructure, and observability.

Your objectives for every incident are:
1. Identify the most probable root cause from the available evidence only.
2. Explain your reasoning step by step, citing specific signals from the data.
3. Estimate your confidence (0–100) based on how complete the evidence is.
4. Recommend the safest, fastest remediation path with numbered steps and rollback instructions.
5. Clearly state the risk of each recommended action.
6. Never invent information that is not present in the incident data provided.

Always respond with valid JSON only — no markdown fences, no prose outside the JSON object.`

// dataGatheringSysPrompt directs the LLM to drive the tool-calling ReAct loop.
const dataGatheringSysPrompt = `You are an SRE Data Gathering Agent embedded in an automated incident-response pipeline.

Your ONLY job is to collect all relevant observability data using the available tools. Be systematic:

1. Check pod status and recent restarts   → kubernetes_query (resource: pods)
2. Retrieve logs from the failing service  → kubernetes_query (resource: logs)
3. Check Kubernetes events                 → kubernetes_query (resource: events)
4. Inspect the current deployment         → kubernetes_query (resource: deployment)
5. Query CPU, memory, and error metrics   → prometheus_query
6. Check recent git commits and PRs       → github_query
7. Verify ArgoCD sync status              → argocd_query
8. Search for similar historical incidents → memory_query

Call tools until you have sufficient data to understand the incident. When done, respond with a plain-text summary of your findings — do NOT use tools in your final response.`

// ── Data-Gathering Phase (ReAct loop) ───────────────────────────────────────

// DataGatheringMessages builds the initial conversation for the ReAct
// data-gathering loop. The orchestrator appends tool results and LLM responses
// as the loop progresses.
func DataGatheringMessages(inc *models.Incident) []Message {
	return []Message{
		{Role: "system", Content: dataGatheringSysPrompt},
		{Role: "user", Content: buildIncidentBlock(inc)},
	}
}

// ── Analysis Phase ───────────────────────────────────────────────────────────

// InvestigationMessages builds the prompt for the Investigator agent.
// history contains the full conversation from the data-gathering phase,
// giving the investigator access to all tool outputs.
func InvestigationMessages(inc *models.Incident, history []Message) []Message {
	msgs := []Message{{Role: "system", Content: sreSysPrompt}}
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: buildInvestigationContent(inc)})
	return msgs
}

// PlannerMessages builds the prompt for the Planner agent.
func PlannerMessages(report *models.InvestigationReport, history []Message) []Message {
	msgs := []Message{{Role: "system", Content: sreSysPrompt}}
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: buildPlannerContent(report)})
	return msgs
}

// PlannerRevisionMessages builds the prompt for a revision pass.
// It reuses the original planner messages and appends the reviewer's concerns
// so the planner can produce a safer recommendation.
func PlannerRevisionMessages(
	report *models.InvestigationReport,
	history []Message,
	prevRec *models.Recommendation,
	concerns string,
) []Message {
	msgs := PlannerMessages(report, history)
	msgs = append(msgs, Message{
		Role: "assistant",
		Content: fmt.Sprintf(`{"root_cause":%q,"reasoning":%q,"confidence":%g,"action":%q,"risk":%q,"similar_incidents":%d}`,
			prevRec.RootCause, prevRec.Reasoning, prevRec.Confidence,
			prevRec.Action, prevRec.Risk, prevRec.SimilarIncidents),
	})
	msgs = append(msgs, Message{
		Role: "user",
		Content: fmt.Sprintf(
			"REVISION REQUEST: The safety reviewer rejected the plan above.\n\nConcerns:\n%s\n\nRevise the recommendation to address every concern. Return the same JSON format.",
			concerns,
		),
	})
	return msgs
}

// ExecutorMessages builds the prompt for the Executor agent.
func ExecutorMessages(inc *models.Incident, rec *models.Recommendation) []Message {
	return []Message{
		{Role: "system", Content: sreSysPrompt},
		{Role: "user", Content: buildExecutorContent(inc, rec)},
	}
}

// ReviewerMessages builds the prompt for the Reviewer agent.
func ReviewerMessages(inc *models.Incident, rec *models.Recommendation, plan *models.ExecutionPlan) []Message {
	return []Message{
		{Role: "system", Content: sreSysPrompt},
		{Role: "user", Content: buildReviewerContent(inc, rec, plan)},
	}
}

// ── Content builders (unexported) ────────────────────────────────────────────

func buildIncidentBlock(inc *models.Incident) string {
	var b strings.Builder
	fmt.Fprintf(&b, "--- INCIDENT ---\n")
	fmt.Fprintf(&b, "ID:          %s\n", inc.ID)
	fmt.Fprintf(&b, "Title:       %s\n", inc.Title)
	fmt.Fprintf(&b, "Severity:    %s\n", inc.Severity)
	fmt.Fprintf(&b, "Cluster:     %s | Namespace: %s | Service: %s\n", inc.Cluster, inc.Namespace, inc.Service)
	fmt.Fprintf(&b, "Description:\n%s\n", inc.Description)
	if ev := inc.Evidence; ev != nil {
		if ev.Logs != "" {
			fmt.Fprintf(&b, "\n--- CALLER-PROVIDED LOGS ---\n%s\n", ev.Logs)
		}
		if ev.Metrics != "" {
			fmt.Fprintf(&b, "\n--- CALLER-PROVIDED METRICS ---\n%s\n", ev.Metrics)
		}
		if ev.Events != "" {
			fmt.Fprintf(&b, "\n--- CALLER-PROVIDED EVENTS ---\n%s\n", ev.Events)
		}
		if ev.Deployment != "" {
			fmt.Fprintf(&b, "\n--- CALLER-PROVIDED DEPLOYMENT ---\n%s\n", ev.Deployment)
		}
		if len(ev.SimilarIssues) > 0 {
			b.WriteString("\n--- CALLER-PROVIDED SIMILAR ISSUES ---\n")
			for i, s := range ev.SimilarIssues {
				fmt.Fprintf(&b, "%d. %s\n", i+1, s)
			}
		}
	}
	return b.String()
}

func buildInvestigationContent(inc *models.Incident) string {
	return fmt.Sprintf(`Using ALL the data collected above by the tools, analyse the incident and respond with a JSON object with EXACTLY these fields:
{
  "summary":      "one-paragraph synopsis of what is happening",
  "key_findings": "the most significant signals, each on a new line — MUST be a single JSON string, NOT an array",
  "data_gaps":    "missing data that would increase confidence — MUST be a single JSON string, NOT an array"
}

IMPORTANT: every value must be a plain JSON string. Do not use JSON arrays for any field.
Incident reference: %s — %s (%s)`, inc.ID, inc.Title, inc.Severity)
}

func buildPlannerContent(report *models.InvestigationReport) string {
	return fmt.Sprintf(`Based on the investigation report and all tool data above, respond with a JSON object with EXACTLY these fields:
{
  "root_cause":        "most probable root cause in one sentence",
  "reasoning":         "step-by-step explanation citing specific signals — a single string, NOT an array",
  "confidence":        75.0,
  "action":            "numbered remediation steps as a single string, e.g. '1. Do X\n2. Do Y' — NOT an array",
  "risk":              "risk level and specific risks — a single string",
  "similar_incidents": 0
}

IMPORTANT: string fields must be plain JSON strings. Numbers must be plain JSON numbers (not strings or percentages).
Investigation summary: %s

Key findings: %s`, report.Summary, report.KeyFindings)
}

func buildExecutorContent(inc *models.Incident, rec *models.Recommendation) string {
	return fmt.Sprintf(`Translate this recommendation into concrete, runnable steps.

Rules:
1. Generate specific commands (kubectl, helm, terraform, shell) — no vague instructions.
2. Classify each step's risk: "low", "medium", or "high".
3. Include a rollback command for every step.
4. Mark irreversible or destructive actions as type "manual".
5. Never reference services or namespaces not in the incident data.

Incident: %s / %s / %s
Root cause (%.0f%% confidence): %s
Action plan: %s

Respond with a JSON object with EXACTLY these fields:
{
  "steps": [
    {
      "order":    1,
      "type":     "kubectl",
      "command":  "kubectl rollout undo deployment/<name> -n %s",
      "purpose":  "Revert to the last stable deployment",
      "risk":     "low",
      "rollback": "kubectl rollout undo deployment/<name> -n %s --to-revision=N"
    }
  ],
  "requires": "Write access to %s cluster"
}`,
		inc.Cluster, inc.Namespace, inc.Service,
		rec.Confidence, rec.RootCause, rec.Action,
		inc.Namespace, inc.Namespace, inc.Cluster)
}

func buildReviewerContent(inc *models.Incident, rec *models.Recommendation, plan *models.ExecutionPlan) string {
	var steps strings.Builder
	for _, s := range plan.Steps {
		fmt.Fprintf(&steps, "Step %d [%s/%s]: %s\n  Purpose:  %s\n  Rollback: %s\n\n",
			s.Order, s.Type, s.Risk, s.Command, s.Purpose, s.Rollback)
	}
	return fmt.Sprintf(`Perform a safety review of this execution plan.

You must:
1. Verify each command is scoped to the correct namespace/service.
2. Flag commands referencing resources not in the incident data (hallucination check).
3. Flag any step that could widen the outage or is irreversible without a safe rollback.
4. Set "approved": false and explain in "concerns" if any step is critical-risk.
5. Suggest safe modifications in "modifications" for flagged steps.

Incident:   %s (%s)
Root cause: %s  [confidence: %.0f%%]
Requires:   %s

Execution plan:
%s
Respond with a JSON object with EXACTLY these fields:
{
  "approved":      true,
  "overall_risk":  "low",
  "concerns":      "None — a single string, NOT an array",
  "modifications": "None required — a single string, NOT an array",
  "final_verdict": "Safe to execute — a single string"
}

IMPORTANT: all string fields must be plain JSON strings, not arrays.`,
		inc.Title, inc.Severity,
		rec.RootCause, rec.Confidence,
		plan.Requires,
		steps.String())
}
