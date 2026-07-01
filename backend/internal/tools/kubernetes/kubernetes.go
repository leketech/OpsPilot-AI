// Package kubernetes provides the kubernetes_query tool.
//
// Current implementation: realistic mock data that tells a complete investigative
// story without needing a live cluster. Every mock function is designed so the
// LLM can reason from the data — each one reveals a different piece of evidence.
//
// To connect to a real cluster, replace Execute() with client-go calls.
// Nothing else in the codebase needs to change.
package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
)

// Input is the typed parameter struct for the kubernetes_query tool.
// When the LLM calls the tool the Registry passes map[string]any; tests may
// pass Input directly to avoid JSON round-trips.
type Input struct {
	Resource  string `json:"resource"`  // pods | logs | events | deployment | hpa | configmap
	Namespace string `json:"namespace"`
	Name      string `json:"name,omitempty"`
	Lines     int    `json:"lines,omitempty"` // for logs; default 50
}

// Tool implements the kubernetes_query tool.
type Tool struct{}

func New() *Tool { return &Tool{} }

func (t *Tool) Name() string { return "kubernetes_query" }

func (t *Tool) Description() string {
	return "Query Kubernetes cluster resources: pods (status + health), logs (application output), events (cluster events), deployments (rollout history), HPAs (autoscaling), and ConfigMaps (configuration). Use this to gather evidence about what is happening inside the cluster."
}

func (t *Tool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"resource": {
				"type":        "string",
				"enum":        ["pods","logs","events","deployment","hpa","configmap"],
				"description": "Kubernetes resource type to query"
			},
			"namespace": {"type": "string",  "description": "Target Kubernetes namespace"},
			"name":      {"type": "string",  "description": "Specific resource name (optional — omit to list/describe all)"},
			"lines":     {"type": "integer", "description": "Number of log lines to return (for logs resource only, default 50)"}
		},
		"required": ["resource", "namespace"]
	}`)
}

// Execute handles calls from both the LLM (input is map[string]any decoded from JSON)
// and from tests (input may be a typed Input struct).
func (t *Tool) Execute(_ context.Context, input any) (any, error) {
	params := decodeInput(input)

	switch params.Resource {
	case "pods":
		return podStatus(params.Namespace), nil
	case "logs":
		return podLogs(params.Namespace, params.Name, params.Lines), nil
	case "events":
		return clusterEvents(params.Namespace), nil
	case "deployment":
		return deploymentDetail(params.Namespace, params.Name), nil
	case "hpa":
		return hpaStatus(params.Namespace), nil
	case "configmap":
		return configMapDetail(params.Namespace, params.Name), nil
	default:
		return nil, fmt.Errorf("kubernetes_query: unsupported resource %q", params.Resource)
	}
}

// decodeInput converts any input type into a typed Input struct.
func decodeInput(input any) Input {
	switch v := input.(type) {
	case Input:
		return v
	case map[string]any:
		lines, _ := v["lines"].(float64)
		p := Input{
			Resource:  strVal(v, "resource"),
			Namespace: strVal(v, "namespace"),
			Name:      strVal(v, "name"),
			Lines:     int(lines),
		}
		if p.Lines <= 0 {
			p.Lines = 50
		}
		return p
	default:
		return Input{Lines: 50}
	}
}

func strVal(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

// ── Mock data functions ───────────────────────────────────────────────────────
// Each function reveals a different layer of evidence about the same incident.
// Together they let the LLM reconstruct the full root-cause chain:
//
//   v2.14.1 deployed → goroutine pool changed from bounded to unbounded
//   → goroutines grew exponentially → heap exhausted → OOMKill loop
//   → rollback to v2.13.9 is the safe remediation

func podStatus(ns string) string {
	return fmt.Sprintf(`$ kubectl get pods -n %s -o wide
NAMESPACE   NAME                               READY   STATUS        RESTARTS   AGE   IP           NODE
%s          %s-api-5f9d6c7b9-xk2p9     0/1     OOMKilled     11         67m   10.0.4.22    node-3
%s          %s-api-5f9d6c7b9-m3rq7     1/1     Running       3          49m   10.0.4.23    node-1
%s          %s-api-5f9d6c7b9-n8wt4     1/1     Running       2          49m   10.0.4.31    node-2
%s          %s-api-5f9d6c7b9-p7ht5     0/1     Pending       0          3m    <none>       <none>
%s          %s-api-5f9d6c7b9-q9rx2     1/1     Running       1          49m   10.0.4.12    node-1
%s          %s-api-5f9d6c7b9-r2wy8     1/1     Running       0          49m   10.0.4.19    node-4
%s          %s-api-5f9d6c7b9-s6kv3     0/1     OOMKilled     7          49m   10.0.4.44    node-2
%s          %s-api-5f9d6c7b9-t1mz6     1/1     Running       4          49m   10.0.4.55    node-3

Pod Conditions (%s-api-5f9d6c7b9-xk2p9):
  PodScheduled:   True
  Initialized:    True
  ContainersReady: False  (OOMKilled, back-off restart)
  Ready:          False

Container State (%s-api):
  Last State: Terminated — Reason: OOMKilled  Exit Code: 137
  Started:    2024-06-01T14:38:02Z
  Finished:   2024-06-01T14:41:19Z
  Restart Count: 11`,
		ns,
		ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns,
		ns, ns)
}

func podLogs(ns, name string, lines int) string {
	pod := name
	if pod == "" {
		pod = ns + "-api-5f9d6c7b9-m3rq7"
	}
	return fmt.Sprintf(`$ kubectl logs %s -n %s --tail=%d --previous=false

2024-06-01T14:09:55Z  INFO  [%s] Starting v2.14.1 — goroutine pool initialised (mode: unbounded)
2024-06-01T14:10:01Z  INFO  [%s] PostgreSQL pool connected (size=10, idle=10)
2024-06-01T14:10:03Z  INFO  [%s] HTTP server listening :8080
2024-06-01T14:11:47Z  INFO  [%s] Processing payment batch: 1,200 transactions/s
2024-06-01T14:12:31Z  WARN  [%s] Goroutine queue depth: 847 (warn threshold: 500)
2024-06-01T14:13:08Z  WARN  [%s] Goroutine queue depth: 1,429
2024-06-01T14:14:18Z  WARN  [%s] Goroutine queue depth: 2,103 — growth is exponential
2024-06-01T14:15:52Z  ERROR [%s] DB pool exhausted: 10/10 active, 142 requests waiting
2024-06-01T14:16:07Z  ERROR [%s] Request timeout txn_8821 — waited 30s for DB connection
2024-06-01T14:16:09Z  ERROR [%s] Request timeout txn_8822 — waited 30s for DB connection
2024-06-01T14:17:34Z  WARN  [%s] Heap usage: 1.40 GB / 2.00 GB (70%%)
2024-06-01T14:18:58Z  ERROR [%s] Goroutines: 3,841 (baseline: 120) — goroutine leak confirmed
2024-06-01T14:20:14Z  WARN  [%s] Heap usage: 1.72 GB / 2.00 GB (86%%)
2024-06-01T14:21:33Z  ERROR [%s] Heap usage: 1.88 GB / 2.00 GB (94%%) — OOM risk HIGH
2024-06-01T14:23:08Z  ERROR [%s] Goroutines: 4,200 — pool goroutines are never recycled
2024-06-01T14:23:47Z  ERROR [%s] Heap usage: 1.96 GB / 2.00 GB (98%%)
2024-06-01T14:24:03Z  FATAL [%s] Process killed by Linux OOMKiller (signal 9)
--- (pod restarted) ---
2024-06-01T14:24:47Z  INFO  [%s] Starting v2.14.1 — goroutine pool initialised (mode: unbounded)
2024-06-01T14:26:09Z  WARN  [%s] Goroutine queue depth: 612 — growth repeating`,
		pod, ns, lines,
		ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns)
}

func clusterEvents(ns string) string {
	return fmt.Sprintf(`$ kubectl get events -n %s --sort-by='.lastTimestamp'

LAST SEEN   TYPE      REASON               OBJECT                               MESSAGE
2m          Warning   OOMKilling           pod/%s-api-5f9d6c7b9-xk2p9    OOMKiller killed container %s-api: memory limit 2Gi exceeded (restart #11)
3m          Normal    BackOff              pod/%s-api-5f9d6c7b9-xk2p9    Back-off restarting failed container %s-api
3m          Warning   FailedScheduling     pod/%s-api-5f9d6c7b9-p7ht5    0/4 nodes available: insufficient memory — cluster at capacity
7m          Normal    SuccessfulRescale    hpa/%s-api                     HPA scaled deployment to 8 replicas (CPU 97%% > target 80%%)
8m          Warning   OOMKilling           pod/%s-api-5f9d6c7b9-s6kv3    OOMKiller killed container %s-api: memory limit 2Gi exceeded (restart #7)
12m         Normal    ScalingReplicaSet    deployment/%s-api              Scaled up replica set %s-api-5f9d6c7b9 from 4 to 8
19m         Normal    ScalingReplicaSet    deployment/%s-api              Scaled up replica set %s-api-5f9d6c7b9 from 2 to 4
25m         Normal    RollingUpdate        deployment/%s-api              Progressing: updated 8/8 pods to v2.14.1
25m         Normal    Pulled               pod/%s-api-5f9d6c7b9-xk2p9    Pulled image registry.prod.example.com/%s-api:v2.14.1
25m         Normal    Scheduled            pod/%s-api-5f9d6c7b9-xk2p9    Successfully assigned %s/%s-api-5f9d6c7b9-xk2p9 to node-3
26m         Normal    RollingUpdate        deployment/%s-api              Started rollout of v2.14.1 (triggered by image change)`,
		ns,
		ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns, ns)
}

func deploymentDetail(ns, name string) string {
	dep := name
	if dep == "" {
		dep = ns + "-api"
	}
	return fmt.Sprintf(`$ kubectl describe deployment %s -n %s

Name:               %s
Namespace:          %s
Labels:             app=%s, team=platform, env=production
Annotations:
  deployment.kubernetes.io/revision: "14"

Selector:           app=%s
Replicas:           8 desired | 8 updated | 8 total | 6 available | 2 unavailable
Strategy:           RollingUpdate  MaxSurge=1  MaxUnavailable=1
Min Ready Seconds:  30

Pod Template:
  Image:    registry.prod.example.com/%s:v2.14.1
  CPU:      request=500m  limit=2000m
  Memory:   request=512Mi limit=2Gi
  Env:
    DB_POOL_SIZE=10        (from ConfigMap %s-config)
    GOROUTINE_POOL_MODE=unbounded  ← NEW in v2.14.1 (was: bounded)
    GOROUTINE_POOL_SIZE=0          ← 0 means no cap (was: 50)

Rollout History:
  REVISION   IMAGE TAG   CHANGE CAUSE                                       STATUS
  11         v2.13.5     fix: payment processor retry ordering              Superseded
  12         v2.13.7     feat: DB pool size increase                        Superseded
  13         v2.13.9     fix: retry timeout handling — 7.5h stable          Superseded ← LAST STABLE
  14         v2.14.1     feat: optimise payment retry logic                 Running    ← CURRENT

Conditions:
  Available:   True    (6/8 pods healthy — below expected 8)
  Progressing: True    (rollout not complete; pods restarting)
  ReplicaFailure: True (pod %s-api-5f9d6c7b9-p7ht5 cannot be scheduled)

Events:
  26m   RollingUpdate   Started: v2.13.9 → v2.14.1  (git commit a3f9c81)
  25m   RollingUpdate   Updated 8/8 pods
  12m   ScalingReplicaSet  Scaled 4 → 8 (CPU pressure)

Key change in v2.14.1 (commit a3f9c81, merged by ci-bot, author: alice):
  internal/retry/pool.go     — goroutine pool changed from bounded (cap=50) to unbounded
  internal/payment/processor.go — async retry added without goroutine recycling`,
		dep, ns,
		dep, ns, dep,
		dep,
		dep, dep,
		dep)
}

func hpaStatus(ns string) string {
	return fmt.Sprintf(`$ kubectl get hpa -n %s

NAME        REFERENCE              TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
%s-api  Deployment/%s-api  cpu: 97%%/80%%  2         8         8          26m

HPA Status:
  Current Replicas: 8  (at maximum — cannot scale further)
  Desired Replicas: 8
  Last Scale Time:  2024-06-01T14:21:07Z (12 minutes ago)

  Scaling Events:
    2→4 replicas at 14:14:02  (cpu crossed 80%% threshold)
    4→8 replicas at 14:21:07  (cpu sustained above 80%%)

Warning: HPA at maxReplicas. Additional CPU pressure cannot be relieved by scaling.
         Root cause is likely a memory/goroutine leak, not insufficient capacity.`,
		ns, ns, ns)
}

func configMapDetail(ns, name string) string {
	cm := name
	if cm == "" {
		cm = ns + "-config"
	}
	return fmt.Sprintf(`$ kubectl describe configmap %s -n %s

Name:       %s
Namespace:  %s
Labels:     app=%s, managed-by=argocd

Data:
  DB_HOST:                 "postgres.%s.svc.cluster.local"
  DB_PORT:                 "5432"
  DB_POOL_SIZE:            "10"
  DB_MAX_IDLE_CONNS:       "5"
  DB_CONNECT_TIMEOUT_S:    "30"
  RETRY_MAX_ATTEMPTS:      "3"
  GOROUTINE_POOL_MODE:     "unbounded"    ← CHANGED at 14:09 UTC (was: "bounded")
  GOROUTINE_POOL_SIZE:     "0"            ← CHANGED at 14:09 UTC (was: "50")
  GOROUTINE_POOL_RECYCLE:  "false"        ← CHANGED at 14:09 UTC (was: "true")
  HTTP_READ_TIMEOUT_S:     "60"
  HTTP_WRITE_TIMEOUT_S:    "60"
  LOG_LEVEL:               "info"

BinaryData: <none>

Events:
  26m   Normal   Updated   configmap/%s   Updated by argocd-server (sync of commit a3f9c81)

Note: GOROUTINE_POOL_MODE, GOROUTINE_POOL_SIZE, and GOROUTINE_POOL_RECYCLE
      were all changed in the same deployment that started the incident.`,
		cm, ns,
		cm, ns, ns, ns,
		cm)
}
