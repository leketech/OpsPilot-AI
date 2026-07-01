package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/leketech/OpsPilot-AI/backend/internal/app"
	"github.com/leketech/OpsPilot-AI/backend/internal/models"
)

// analyzeResponse is the JSON contract returned to API callers.
type analyzeResponse struct {
	RootCause         string             `json:"rootCause"`
	Confidence        float64            `json:"confidence"`
	RecommendedAction string             `json:"recommendedAction"`
	Risk              string             `json:"risk"`
	SimilarIncidents  int                `json:"similarIncidents"`
	Investigation     investigationBlock `json:"investigation"`
	ExecutionPlan     executionPlanBlock `json:"executionPlan"`
	Review            reviewBlock        `json:"review"`
}

type investigationBlock struct {
	Summary     string `json:"summary"`
	KeyFindings string `json:"keyFindings"`
	DataGaps    string `json:"dataGaps"`
}

type executionPlanBlock struct {
	Steps    []models.ExecutionStep `json:"steps"`
	Requires string                 `json:"requires"`
}

type reviewBlock struct {
	Approved      bool   `json:"approved"`
	OverallRisk   string `json:"overallRisk"`
	Concerns      string `json:"concerns"`
	Modifications string `json:"modifications"`
	FinalVerdict  string `json:"finalVerdict"`
}

// AnalyzeIncident handles POST /api/v1/incidents/analyze.
//
// The request body must be a JSON-encoded models.Incident.
// Session identity is resolved from X-Session-ID → incident ID → caller IP.
func AnalyzeIncident(a *app.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var inc models.Incident
		if err := c.BodyParser(&inc); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body: " + err.Error(),
			})
		}
		if inc.Title == "" || inc.Description == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "title and description are required",
			})
		}

		sessionID := c.Get("X-Session-ID")
		if sessionID == "" {
			sessionID = inc.ID
		}
		if sessionID == "" {
			sessionID = c.IP()
		}

		report, err := a.Agent.Handle(c.Context(), sessionID, &inc)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(analyzeResponse{
			RootCause:         report.Recommendation.RootCause,
			Confidence:        report.Recommendation.Confidence,
			RecommendedAction: report.Recommendation.Action,
			Risk:              report.Recommendation.Risk,
			SimilarIncidents:  report.Recommendation.SimilarIncidents,
			Investigation: investigationBlock{
				Summary:     report.Investigation.Summary,
				KeyFindings: report.Investigation.KeyFindings,
				DataGaps:    report.Investigation.DataGaps,
			},
			ExecutionPlan: executionPlanBlock{
				Steps:    report.ExecutionPlan.Steps,
				Requires: report.ExecutionPlan.Requires,
			},
			Review: reviewBlock{
				Approved:      report.Review.Approved,
				OverallRisk:   report.Review.OverallRisk,
				Concerns:      report.Review.Concerns,
				Modifications: report.Review.Modifications,
				FinalVerdict:  report.Review.FinalVerdict,
			},
		})
	}
}
