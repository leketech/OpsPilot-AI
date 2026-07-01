package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ── LLM-tolerant JSON primitives ─────────────────────────────────────────────
// LLMs routinely return "bullet-point" or "numbered step" fields as JSON arrays
// despite being told to return strings. These types accept either format so the
// pipeline never crashes on a formatting quirk.

// flexString accepts a JSON string or a JSON array and always yields a string.
// Arrays are joined with newlines so the content is still human-readable.
type flexString string

func (f *flexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var arr []any
	if err := json.Unmarshal(data, &arr); err == nil {
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = fmt.Sprint(v)
		}
		*f = flexString(strings.Join(parts, "\n"))
		return nil
	}
	return fmt.Errorf("flexString: cannot unmarshal %s", data)
}

// flexFloat64 accepts a JSON number or a numeric string ("87.5", "87.5%").
type flexFloat64 float64

func (f *flexFloat64) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err == nil {
		*f = flexFloat64(v)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		s = strings.TrimSuffix(strings.TrimSpace(s), "%")
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			*f = flexFloat64(v)
			return nil
		}
	}
	return fmt.Errorf("flexFloat64: cannot unmarshal %s", data)
}

// ── Domain types ──────────────────────────────────────────────────────────────

// Evidence groups all observability data attached to an incident.
type Evidence struct {
	Logs          string   `json:"logs,omitempty"`
	Metrics       string   `json:"metrics,omitempty"`
	Events        string   `json:"events,omitempty"`
	Deployment    string   `json:"deployment,omitempty"`
	SimilarIssues []string `json:"similar_issues,omitempty"`
}

// Incident is the unit of work passed into the agent pipeline.
type Incident struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	Namespace   string    `json:"namespace"`
	Cluster     string    `json:"cluster"`
	Service     string    `json:"service"`
	Description string    `json:"description"`
	Evidence    *Evidence `json:"evidence,omitempty"`
}

// InvestigationReport is produced by the Investigator agent.
type InvestigationReport struct {
	Summary     string `json:"summary"`
	KeyFindings string `json:"key_findings"`
	DataGaps    string `json:"data_gaps"`
}

func (r *InvestigationReport) UnmarshalJSON(data []byte) error {
	var aux struct {
		Summary     flexString `json:"summary"`
		KeyFindings flexString `json:"key_findings"`
		DataGaps    flexString `json:"data_gaps"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Summary = string(aux.Summary)
	r.KeyFindings = string(aux.KeyFindings)
	r.DataGaps = string(aux.DataGaps)
	return nil
}

// Recommendation is produced by the Planner agent.
type Recommendation struct {
	RootCause        string  `json:"root_cause"`
	Reasoning        string  `json:"reasoning"`
	Confidence       float64 `json:"confidence"`
	Action           string  `json:"action"`
	Risk             string  `json:"risk"`
	SimilarIncidents int     `json:"similar_incidents"`
}

func (r *Recommendation) UnmarshalJSON(data []byte) error {
	var aux struct {
		RootCause        flexString  `json:"root_cause"`
		Reasoning        flexString  `json:"reasoning"`
		Confidence       flexFloat64 `json:"confidence"`
		Action           flexString  `json:"action"`
		Risk             flexString  `json:"risk"`
		SimilarIncidents int         `json:"similar_incidents"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.RootCause = string(aux.RootCause)
	r.Reasoning = string(aux.Reasoning)
	r.Confidence = float64(aux.Confidence)
	r.Action = string(aux.Action)
	r.Risk = string(aux.Risk)
	r.SimilarIncidents = aux.SimilarIncidents
	return nil
}

// ExecutionStep is a single concrete remediation action.
type ExecutionStep struct {
	Order    int    `json:"order"`
	Type     string `json:"type"`
	Command  string `json:"command"`
	Purpose  string `json:"purpose"`
	Risk     string `json:"risk"`
	Rollback string `json:"rollback"`
}

func (s *ExecutionStep) UnmarshalJSON(data []byte) error {
	var aux struct {
		Order    int        `json:"order"`
		Type     flexString `json:"type"`
		Command  flexString `json:"command"`
		Purpose  flexString `json:"purpose"`
		Risk     flexString `json:"risk"`
		Rollback flexString `json:"rollback"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Order = aux.Order
	s.Type = string(aux.Type)
	s.Command = string(aux.Command)
	s.Purpose = string(aux.Purpose)
	s.Risk = string(aux.Risk)
	s.Rollback = string(aux.Rollback)
	return nil
}

// ExecutionPlan is produced by the Executor agent.
type ExecutionPlan struct {
	Steps    []ExecutionStep `json:"steps"`
	Requires string          `json:"requires"`
}

func (p *ExecutionPlan) UnmarshalJSON(data []byte) error {
	var aux struct {
		Steps    []ExecutionStep `json:"steps"`
		Requires flexString      `json:"requires"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Steps = aux.Steps
	p.Requires = string(aux.Requires)
	return nil
}

// Review is produced by the Reviewer agent.
type Review struct {
	Approved      bool   `json:"approved"`
	OverallRisk   string `json:"overall_risk"`
	Concerns      string `json:"concerns"`
	Modifications string `json:"modifications"`
	FinalVerdict  string `json:"final_verdict"`
}

func (r *Review) UnmarshalJSON(data []byte) error {
	var aux struct {
		Approved      bool       `json:"approved"`
		OverallRisk   flexString `json:"overall_risk"`
		Concerns      flexString `json:"concerns"`
		Modifications flexString `json:"modifications"`
		FinalVerdict  flexString `json:"final_verdict"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Approved = aux.Approved
	r.OverallRisk = string(aux.OverallRisk)
	r.Concerns = string(aux.Concerns)
	r.Modifications = string(aux.Modifications)
	r.FinalVerdict = string(aux.FinalVerdict)
	return nil
}

// Report is the final artefact produced by one full pipeline run.
type Report struct {
	Incident       *Incident
	Investigation  *InvestigationReport
	Recommendation *Recommendation
	ExecutionPlan  *ExecutionPlan
	Review         *Review
}

// IncidentRecord is the complete post-mortem stored in the Memory agent's
// historical index.
type IncidentRecord struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Service       string  `json:"service"`
	Namespace     string  `json:"namespace"`
	RootCause     string  `json:"root_cause"`
	Confidence    float64 `json:"confidence"`
	Evidence      string  `json:"evidence"`
	Fix           string  `json:"fix"`
	HumanDecision string  `json:"human_decision"`
	Outcome       string  `json:"outcome"`
	OccurredAt    string  `json:"occurred_at"`
}
