package reviewer

import (
	"context"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// Reviewer validates the ExecutionPlan for safety before it reaches the caller.
type Reviewer struct {
	llm *qwen.Client
}

func New(llm *qwen.Client) *Reviewer { return &Reviewer{llm: llm} }

// Review calls Qwen to validate the plan and returns a Review.
// If Review.Approved is false the orchestrator should surface the concerns
// to the caller rather than proceeding with execution.
func (r *Reviewer) Review(
	ctx context.Context,
	inc *models.Incident,
	rec *models.Recommendation,
	plan *models.ExecutionPlan,
) (*models.Review, error) {
	msgs := qwen.ReviewerMessages(inc, rec, plan)

	var review models.Review
	if _, err := r.llm.CompleteJSON(ctx, msgs, &review); err != nil {
		return nil, fmt.Errorf("reviewer: %w", err)
	}
	return &review, nil
}
