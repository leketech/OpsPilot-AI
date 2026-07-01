// Package investigator implements the Investigator agent.
package investigator

import (
	"context"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// Runner is the Investigator agent.
// It analyses all the data gathered during the ReAct phase and returns a
// structured InvestigationReport.
type Runner struct {
	llm *qwen.Client
}

func New(llm *qwen.Client) *Runner { return &Runner{llm: llm} }

// Run receives the gathering context (full conversation history including all
// tool outputs) and produces a structured InvestigationReport.
func (r *Runner) Run(
	ctx context.Context,
	inc *models.Incident,
	gatheringCtx []qwen.Message,
) (*models.InvestigationReport, error) {
	msgs := qwen.InvestigationMessages(inc, gatheringCtx)

	var report models.InvestigationReport
	if _, err := r.llm.CompleteJSON(ctx, msgs, &report); err != nil {
		return nil, fmt.Errorf("investigator: %w", err)
	}
	return &report, nil
}
