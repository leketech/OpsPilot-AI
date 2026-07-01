import type { AnalysisResult, IncidentInput } from "./types";

export async function analyzeIncident(
  incident: IncidentInput
): Promise<AnalysisResult> {
  const res = await fetch("/api/v1/incidents/analyze", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(incident),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Request failed (HTTP ${res.status})`);
  }

  return res.json();
}

export const SAMPLE_INCIDENT: IncidentInput = {
  id: "INC-001",
  title: "High CPU Usage — payments service",
  service: "payments-api",
  severity: "critical",
  namespace: "payments",
  cluster: "prod-us-east-1",
  description:
    "CPU utilisation on the payments-api pods has been above 95% for the past 20 minutes, causing p99 latency to spike above 3s and triggering PagerDuty alert PROD-4821.",
  evidence: {
    logs: `2024-06-01T14:32:01Z ERROR payments-api: timeout waiting for DB connection pool (pool_size=10, wait_queue=142)
2024-06-01T14:32:05Z WARN  payments-api: retrying transaction id=txn_8821 attempt=3
2024-06-01T14:32:09Z ERROR payments-api: goroutine leak detected — 4,200 goroutines active (baseline 120)
2024-06-01T14:32:15Z ERROR payments-api: OOM risk — heap 1.8 GB / limit 2 GB`,
    metrics: `cpu_utilisation:       97% (threshold 80%)
memory_utilisation:    90%
http_p99_latency_ms:   3200 (SLO 500 ms)
active_goroutines:     4200 (baseline 120)
error_rate:            12% (SLO <0.1%)`,
    events: `WARN  HPA scaled replicas from 4 → 8 at 14:28 UTC
INFO  Deployment payments-api rolled out v2.14.1 at 14:10 UTC
WARN  Readiness probe failed: payments-api-7d9f8b (3 consecutive failures)`,
  },
};
