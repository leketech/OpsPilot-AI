// Package github provides the github_query tool.
// Returns realistic mock git history for development and demo use.
// Replace Execute() with a real GitHub API call for production.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leketech/OpsPilot-AI/backend/internal/tools"
)

// Tool implements the github_query tool.
type Tool struct{}

func New() *Tool { return &Tool{} }

func (t *Tool) Name() string { return "github_query" }

func (t *Tool) Description() string {
	return "Query GitHub for recent commits, pull requests, or releases for a given repository. Use this to identify code changes that may have caused or contributed to the incident."
}

func (t *Tool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"repo": {
				"type":        "string",
				"description": "Repository in owner/name format, e.g. 'acme/payments-api'"
			},
			"resource": {
				"type":        "string",
				"enum":        ["commits","pulls","releases"],
				"description": "Type of GitHub resource to query"
			},
			"limit": {
				"type":        "integer",
				"description": "Maximum number of results to return (default 10)"
			}
		},
		"required": ["repo", "resource"]
	}`)
}

func (t *Tool) Execute(_ context.Context, input any) (any, error) {
	args, _ := input.(map[string]any)
	repo := tools.StrArg(args, "repo")
	resource := tools.StrArg(args, "resource")
	limit := tools.IntArg(args, "limit", 10)

	switch resource {
	case "commits":
		return mockCommits(repo, limit), nil
	case "pulls":
		return mockPulls(repo, limit), nil
	case "releases":
		return mockReleases(repo, limit), nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %q", resource)
	}
}

func mockCommits(repo string, limit int) string {
	svc := lastSegment(repo)
	return fmt.Sprintf(`Recent commits to %s (showing %d):

a3f9c81  2024-06-01 14:05  ci-bot      feat: optimise payment retry logic with goroutine pool
                            ↑ DEPLOYED as v2.14.1 at 14:10 UTC — 22min before incident
b7d2e45  2024-06-01 09:12  alice       fix: increase DB pool size default from 5 to 10
c1a8f93  2024-05-31 17:44  bob         chore: upgrade go-redis to v9.21.0
d4e7b22  2024-05-31 14:30  alice       feat: add payment retry with exponential backoff
e9f3c61  2024-05-30 11:15  ci-bot      chore: bump dependencies

Notable: a3f9c81 introduced changes to the %s goroutine pool — review for leaks.`, repo, limit, svc)
}

func mockPulls(repo string, limit int) string {
	return fmt.Sprintf(`Recent pull requests for %s (showing %d):

#482  MERGED   2024-06-01 13:58  "feat: optimise payment retry logic with goroutine pool"
                                   Author: alice  Reviewer: bob  ← merged 12min before incident
      Files: internal/retry/pool.go (+142 -38), internal/payment/processor.go (+21 -9)

#481  MERGED   2024-06-01 08:45  "fix: increase DB pool size default"
      Files: config/defaults.go (+3 -3)

#479  OPEN     2024-05-31 16:00  "feat: async notification delivery"
      Status: awaiting review`, repo, limit)
}

func mockReleases(repo string, limit int) string {
	return fmt.Sprintf(`Releases for %s (showing %d):

v2.14.1  2024-06-01 14:10  ← CURRENT (deployed 22min ago)
          Changelog: Optimised retry goroutine pool; performance improvements

v2.13.9  2024-06-01 06:30  ← PREVIOUS (stable, 7.5h uptime before this release)
          Changelog: Bugfix: DB connection timeout handling

v2.13.2  2024-05-28 11:00
          Changelog: Fix goroutine leak in payment retry logic

Rollback candidate: v2.13.9 (last stable)`, repo, limit)
}

func lastSegment(s string) string {
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}
