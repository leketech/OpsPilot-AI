// Package orchestrator coordinates the five-agent, tool-using pipeline
// as an explicit state machine. Each state is a named constant; the loop
// drives transitions until stateDone or an unrecoverable error.
//
// State diagram:
//
//	stateGather → stateAnalyse → statePlan → stateExecute → stateReview
//	                                ↑                            │
//	                           stateRevise ←──── (rejected) ────┘
//	                                                             │
//	                                                     stateStore → stateDone
package orchestrator

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/leketech/OpsPilot-AI/backend/internal/agent/executor"
	"github.com/leketech/OpsPilot-AI/backend/internal/agent/investigator"
	"github.com/leketech/OpsPilot-AI/backend/internal/agent/planner"
	"github.com/leketech/OpsPilot-AI/backend/internal/agent/reviewer"
	"github.com/leketech/OpsPilot-AI/backend/internal/config"
	"github.com/leketech/OpsPilot-AI/backend/internal/llm/qwen"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
	"github.com/leketech/OpsPilot-AI/backend/internal/tools"
	argocdtool "github.com/leketech/OpsPilot-AI/backend/internal/tools/argocd"
	githubtool "github.com/leketech/OpsPilot-AI/backend/internal/tools/github"
	kubetool "github.com/leketech/OpsPilot-AI/backend/internal/tools/kubernetes"
	memtool "github.com/leketech/OpsPilot-AI/backend/internal/tools/memory"
	promtool "github.com/leketech/OpsPilot-AI/backend/internal/tools/prometheus"
)

// state is the set of named pipeline states.
type state int

const (
	stateGather   state = iota // ReAct loop: LLM calls tools until evidence is complete
	stateAnalyse               // Investigator: synthesise evidence into a report
	statePlan                  // Planner: generate a Recommendation
	stateExecute               // Executor: translate Recommendation into runnable steps
	stateReview                // Reviewer: safety-check the ExecutionPlan (Human Approval gate)
	stateRevise                // Re-plan addressing reviewer concerns (rejection branch)
	stateStore                 // Persist session + incident record to Redis
	stateDone
)

const (
	maxGatheringIterations = 8
	maxRevisions           = 2 // cap on plan–review–revise cycles before accepting best effort
)

// pipelineCtx carries the evolving state across all pipeline stages.
// It is the single source of truth for one Handle() invocation.
type pipelineCtx struct {
	sessionID     string
	incident      *models.Incident
	gatherCtx     []qwen.Message          // full ReAct conversation (tool calls + results)
	investigation *models.InvestigationReport
	recommendation *models.Recommendation
	plan          *models.ExecutionPlan
	review        *models.Review
	revisions     int // number of plan revisions performed so far
}

// Orchestrator coordinates the full pipeline. It does not know about any
// specific tool or cloud provider — it only calls agents and manages state.
type Orchestrator struct {
	llm      *qwen.Client
	memory   *memtool.Store
	registry *tools.Registry

	investigator *investigator.Runner
	planner      *planner.Planner
	executor     *executor.Executor
	reviewer     *reviewer.Reviewer
}

// New builds an Orchestrator wired to Qwen and Redis.
func New(cfg *config.Config, rdb *redis.Client) *Orchestrator {
	llm := qwen.New(cfg)
	mem := memtool.New(rdb)

	registry := tools.NewRegistry(
		kubetool.New(),
		promtool.New(),
		githubtool.New(),
		argocdtool.New(),
		mem, // memory.Store implements tools.Tool
	)

	return &Orchestrator{
		llm:          llm,
		memory:       mem,
		registry:     registry,
		investigator: investigator.New(llm),
		planner:      planner.New(llm),
		executor:     executor.New(llm),
		reviewer:     reviewer.New(llm),
	}
}

// Handle runs the full state machine and returns an assembled Report.
//
// Flow:
//
//	stateGather → stateAnalyse → statePlan → stateExecute → stateReview
//	                                  ↑ ←── stateRevise ←── (rejected, revisions < max)
//	                                                    ↓
//	                                               stateStore → stateDone
func (o *Orchestrator) Handle(ctx context.Context, sessionID string, inc *models.Incident) (*models.Report, error) {
	pc := &pipelineCtx{sessionID: sessionID, incident: inc}

	for s := state(stateGather); s != stateDone; {
		next, err := o.step(ctx, pc, s)
		if err != nil {
			return nil, err
		}
		s = next
	}

	return &models.Report{
		Incident:       pc.incident,
		Investigation:  pc.investigation,
		Recommendation: pc.recommendation,
		ExecutionPlan:  pc.plan,
		Review:         pc.review,
	}, nil
}

// step executes one state and returns the next state.
func (o *Orchestrator) step(ctx context.Context, pc *pipelineCtx, s state) (state, error) {
	switch s {

	// ── Gather ──────────────────────────────────────────────────────────────────
	case stateGather:
		gc, err := o.gatherData(ctx, pc.incident)
		if err != nil {
			return 0, fmt.Errorf("gather: %w", err)
		}
		pc.gatherCtx = gc
		return stateAnalyse, nil

	// ── Analyse ─────────────────────────────────────────────────────────────────
	case stateAnalyse:
		inv, err := o.investigator.Run(ctx, pc.incident, pc.gatherCtx)
		if err != nil {
			return 0, err
		}
		pc.investigation = inv
		return statePlan, nil

	// ── Plan ────────────────────────────────────────────────────────────────────
	case statePlan:
		rec, err := o.planner.Plan(ctx, pc.investigation, pc.gatherCtx)
		if err != nil {
			return 0, err
		}
		pc.recommendation = rec
		return stateExecute, nil

	// ── Execute (generate runbook) ───────────────────────────────────────────────
	case stateExecute:
		plan, err := o.executor.Prepare(ctx, pc.incident, pc.recommendation)
		if err != nil {
			return 0, err
		}
		pc.plan = plan
		return stateReview, nil

	// ── Review (Human Approval gate) ────────────────────────────────────────────
	// If the reviewer rejects and we have revisions remaining, loop back.
	// After maxRevisions the best plan is accepted as-is (non-fatal).
	case stateReview:
		rev, err := o.reviewer.Review(ctx, pc.incident, pc.recommendation, pc.plan)
		if err != nil {
			return 0, err
		}
		pc.review = rev

		if !rev.Approved && pc.revisions < maxRevisions {
			log.WithFields(log.Fields{
				"revision": pc.revisions + 1,
				"concerns": rev.Concerns,
			}).Info("orchestrator: reviewer rejected plan — requesting revision")
			return stateRevise, nil
		}

		if !rev.Approved {
			log.WithField("revisions", pc.revisions).
				Warn("orchestrator: max revisions reached — proceeding with best-effort plan")
		}
		return stateStore, nil

	// ── Revise (rejection branch) ────────────────────────────────────────────────
	case stateRevise:
		pc.revisions++
		rec, err := o.planner.Revise(
			ctx,
			pc.investigation,
			pc.recommendation,
			pc.review.Concerns,
			pc.gatherCtx,
		)
		if err != nil {
			return 0, err
		}
		pc.recommendation = rec
		return stateExecute, nil // re-generate runbook for revised plan, then re-review

	// ── Store (Learn & Store Memory) ─────────────────────────────────────────────
	case stateStore:
		report := &models.Report{
			Incident:       pc.incident,
			Investigation:  pc.investigation,
			Recommendation: pc.recommendation,
			ExecutionPlan:  pc.plan,
			Review:         pc.review,
		}

		memSummary := fmt.Sprintf(
			"Incident: %s (%s)\nRoot cause: %s\nAction: %s\nReviewer approved: %v\nRevisions: %d",
			pc.incident.Title, pc.incident.Severity,
			pc.recommendation.RootCause, pc.recommendation.Action,
			pc.review.Approved, pc.revisions,
		)
		if err := o.memory.AppendSession(ctx, pc.sessionID,
			qwen.Message{Role: "user", Content: pc.incident.Description},
			qwen.Message{Role: "assistant", Content: memSummary},
		); err != nil {
			return 0, fmt.Errorf("persist session: %w", err)
		}

		if err := o.memory.Record(ctx, report); err != nil {
			log.WithError(err).Warn("orchestrator: failed to record incident to history index")
		}

		return stateDone, nil
	}

	return 0, fmt.Errorf("orchestrator: unhandled state %d", s)
}

// gatherData runs the ReAct loop. It returns the full conversation including
// all tool calls and results — this becomes the shared context for every
// subsequent analysis agent.
func (o *Orchestrator) gatherData(ctx context.Context, inc *models.Incident) ([]qwen.Message, error) {
	messages := qwen.DataGatheringMessages(inc)
	toolDefs := o.registry.Definitions()

	for i := 0; i < maxGatheringIterations; i++ {
		resp, err := o.llm.CompleteWithTools(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("iteration %d: %w", i+1, err)
		}

		messages = append(messages, resp.AssistantMessage())

		if !resp.HasToolCalls() {
			log.WithField("iterations", i+1).Debug("orchestrator: data gathering complete")
			break
		}

		for _, tc := range resp.ToolCalls {
			log.WithFields(log.Fields{
				"tool": tc.Function.Name,
				"args": tc.Function.Arguments,
			}).Debug("orchestrator: executing tool")

			result, err := o.registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = fmt.Sprintf("[tool error: %s]", err)
			}
			messages = append(messages, qwen.ToolResultMessage(tc.ID, result))
		}
	}

	return messages, nil
}
