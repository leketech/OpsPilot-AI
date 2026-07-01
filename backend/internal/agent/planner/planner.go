package planner

import (
	"context"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// Planner generates a prioritised Recommendation from an InvestigationReport.
type Planner struct {
	llm *qwen.Client
}

func New(llm *qwen.Client) *Planner { return &Planner{llm: llm} }

// Plan calls Qwen with the investigation report and the gathering context,
// returning a structured Recommendation.
func (p *Planner) Plan(
	ctx context.Context,
	report *models.InvestigationReport,
	gatheringCtx []qwen.Message,
) (*models.Recommendation, error) {
	msgs := qwen.PlannerMessages(report, gatheringCtx)

	var rec models.Recommendation
	if _, err := p.llm.CompleteJSON(ctx, msgs, &rec); err != nil {
		return nil, fmt.Errorf("planner: %w", err)
	}
	return &rec, nil
}

// Revise re-runs the planner after a reviewer rejection. It gives Qwen the
// original recommendation plus the reviewer's concerns so the revision
// targets the specific safety issues that were flagged.
func (p *Planner) Revise(
	ctx context.Context,
	report *models.InvestigationReport,
	prev *models.Recommendation,
	concerns string,
	gatheringCtx []qwen.Message,
) (*models.Recommendation, error) {
	msgs := qwen.PlannerRevisionMessages(report, gatheringCtx, prev, concerns)

	var rec models.Recommendation
	if _, err := p.llm.CompleteJSON(ctx, msgs, &rec); err != nil {
		return nil, fmt.Errorf("planner.revise: %w", err)
	}
	return &rec, nil
}
