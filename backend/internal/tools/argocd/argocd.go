// Package argocd provides the argocd_query tool.
// Returns realistic mock ArgoCD application status for development and demo use.
// Replace Execute() with a real ArgoCD API call for production.
package argocd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leketech/OpsPilot-AI/backend/internal/tools"
)

// Tool implements the argocd_query tool.
type Tool struct{}

func New() *Tool { return &Tool{} }

func (t *Tool) Name() string { return "argocd_query" }

func (t *Tool) Description() string {
	return "Check the ArgoCD sync and health status for a deployed application. Use this to determine whether a recent GitOps deployment is in sync with the cluster state."
}

func (t *Tool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"application": {
				"type":        "string",
				"description": "ArgoCD application name, e.g. 'payments-api-prod'"
			}
		},
		"required": ["application"]
	}`)
}

func (t *Tool) Execute(_ context.Context, input any) (any, error) {
	args, _ := input.(map[string]any)
	app := tools.StrArg(args, "application")
	if app == "" {
		return nil, fmt.Errorf("argocd_query: 'application' argument is required")
	}
	return mockArgoCDStatus(app), nil
}

func mockArgoCDStatus(app string) string {
	return fmt.Sprintf(`ArgoCD Application: %s
Cluster:         prod-us-east-1 (https://k8s.prod.example.com)
Namespace:       payments
Project:         production

Sync Status:     OutOfSync  ← 2 resources differ from Git
Health Status:   Degraded   ← pods failing readiness probes

Last Sync:       2024-06-01T14:10:22Z  (22 minutes ago)
Sync Trigger:    Automated (push to main, commit a3f9c81)
Sync Duration:   47 seconds

OutOfSync Resources:
  apps/Deployment/payments-api     (desired: v2.14.1, live: mixed v2.14.1/v2.13.9 during rollout)
  autoscaling/HPA/payments-api     (desired: maxReplicas=8, live: 8 — at ceiling)

Health Summary:
  Deployment payments-api: Progressing (6/8 pods healthy)
  Service     payments-api: Healthy
  HPA         payments-api: Degraded (at max replicas, cpu > target)

Rollback available: argocd app rollback %s --revision 13`, app, app)
}
