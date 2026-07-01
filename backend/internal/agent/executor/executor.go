package executor

import (
	"context"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// Executor translates a Recommendation into a concrete ExecutionPlan.
type Executor struct {
	llm *qwen.Client
}

func New(llm *qwen.Client) *Executor { return &Executor{llm: llm} }

// Prepare calls Qwen and returns a structured ExecutionPlan.
func (e *Executor) Prepare(
	ctx context.Context,
	inc *models.Incident,
	rec *models.Recommendation,
) (*models.ExecutionPlan, error) {
	msgs := qwen.ExecutorMessages(inc, rec)

	var plan models.ExecutionPlan
	if _, err := e.llm.CompleteJSON(ctx, msgs, &plan); err != nil {
		return nil, fmt.Errorf("executor: %w", err)
	}
	return &plan, nil
}
