// Package testdata provides mock incidents for local development and demos.
// Use these to exercise the AI pipeline without a live Kubernetes cluster.
package testdata

import "github.com/leketech/OpsPilot-AI/backend/internal/models"

// HighCPUPayments simulates a CPU-saturation incident on the payments service.
var HighCPUPayments = models.Incident{
	ID:          "INC-001",
	Title:       "High CPU Usage — payments service",
	Severity:    "critical",
	Cluster:     "prod-us-east-1",
	Namespace:   "payments",
	Service:     "payments-api",
	Description: "CPU utilisation on the payments-api pods has been above 95% for the past 20 minutes, causing p99 latency to spike above 3 s and triggering PagerDuty alert PROD-4821.",
	Evidence: &models.Evidence{
		Logs: `2024-06-01T14:32:01Z ERROR payments-api: timeout waiting for DB connection pool (pool_size=10, wait_queue=142)
2024-06-01T14:32:05Z WARN  payments-api: retrying transaction id=txn_8821 attempt=3
2024-06-01T14:32:09Z ERROR payments-api: goroutine leak detected — 4,200 goroutines active (baseline 120)
2024-06-01T14:32:15Z ERROR payments-api: OOM risk — heap 1.8 GB / limit 2 GB`,
		Metrics: `cpu_utilisation:       97% (threshold 80%)
memory_utilisation:    90%
http_p99_latency_ms:   3200 (SLO 500 ms)
db_connection_wait_ms: 850 (normal <10 ms)
active_goroutines:     4200 (baseline 120)
error_rate:            12% (SLO <0.1%)`,
		Events: `WARN  Readiness probe failed: payments-api-7d9f8b (3 consecutive failures)
WARN  HPA scaled replicas from 4 → 8 at 14:28 UTC
INFO  Deployment payments-api rolled out v2.14.1 at 14:10 UTC
WARN  PodDisruptionBudget: only 3/4 required pods available`,
		SimilarIssues: []string{
			"INC-0087 (2024-05-12): Connection pool exhaustion after v2.12.0 deploy — resolved by rolling back",
			"INC-0063 (2024-04-03): Goroutine leak in payment retry logic — fixed in v2.13.2",
		},
	},
}

// OOMKillOrderService simulates repeated OOMKill restarts on the order service.
var OOMKillOrderService = models.Incident{
	ID:          "INC-002",
	Title:       "OOMKill loop — order-service",
	Severity:    "high",
	Cluster:     "prod-eu-west-1",
	Namespace:   "orders",
	Service:     "order-service",
	Description: "order-service pods are being OOMKilled every 4–6 minutes. Kubernetes has restarted 11 times in the last hour. Customer order submissions are failing with 503.",
	Evidence: &models.Evidence{
		Logs: `2024-06-01T09:10:42Z INFO  order-service: loading product catalogue into memory (items=2,400,000)
2024-06-01T09:10:55Z WARN  order-service: memory allocation 1.9 GB exceeds 75% of limit
2024-06-01T09:11:03Z FATAL order-service: killed by OOMKiller
2024-06-01T09:15:09Z INFO  order-service: restarting (restart #9)`,
		Metrics: `memory_utilisation:  99% → OOMKill
restart_count:       11 in last 60 min
http_5xx_rate:       34%
catalogue_load_ms:   13200 (expected <500 ms)
pod_ready_ratio:     0/3`,
		Events: `WARN  OOMKilled: order-service-6c7d9 at 09:11:03
WARN  OOMKilled: order-service-6c7d9 at 09:06:51
INFO  ConfigMap order-service-config updated at 08:55 UTC (catalogue_preload: true → added by team-orders)
INFO  Deployment order-service v1.8.0 rolled out at 08:52 UTC`,
		SimilarIssues: []string{
			"INC-0071 (2024-04-20): Catalogue full-load feature flag enabled by mistake — reverted within 30 min",
		},
	},
}

// KafkaConsumerLag simulates a Kafka consumer group falling behind.
var KafkaConsumerLag = models.Incident{
	ID:          "INC-003",
	Title:       "Kafka consumer lag — notification-worker",
	Severity:    "medium",
	Cluster:     "prod-us-east-1",
	Namespace:   "notifications",
	Service:     "notification-worker",
	Description: "notification-worker consumer group lag on topic user-events has grown from 200 to 1.4 million messages over 90 minutes. Email and push notifications are delayed by up to 45 minutes.",
	Evidence: &models.Evidence{
		Logs: `2024-06-01T11:00:01Z WARN  notification-worker: slow external call to SendGrid API (latency=4800ms, timeout=5000ms)
2024-06-01T11:00:06Z ERROR notification-worker: SendGrid rate limit hit (429) — backing off 60s
2024-06-01T11:00:07Z INFO  notification-worker: poll loop paused during back-off (partition 0-7 all paused)
2024-06-01T11:05:00Z WARN  notification-worker: consumer lag=180000 and growing`,
		Metrics: `consumer_lag:          1,400,000 (alert threshold 10,000)
sendgrid_p99_ms:       4800 (SLO 500 ms)
sendgrid_429_rate:     62 errors/min
messages_processed/s:  12 (normal 3,200)
notification_delay_s:  2700 (45 min)`,
		Events: `INFO  SendGrid status page: degraded performance reported at 10:45 UTC
WARN  notification-worker HPA at max replicas (8/8) — cannot scale further
INFO  Deployment notification-worker v3.2.0 at 07:30 UTC (no config changes)`,
		SimilarIssues: []string{
			"INC-0099 (2024-05-28): SendGrid outage caused 2-hour notification delay — mitigated with SQS dead-letter queue fallback (not yet in prod)",
		},
	},
}
