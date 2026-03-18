# Observability & Metrics

This project implements production-grade observability following **Google SRE's 4 Golden Signals**: Latency, Traffic, Errors, and Saturation.

## Stack

| Component | Purpose | Port |
|-----------|---------|------|
| **Prometheus** | Metrics collection & storage | 9090 |
| **Grafana** | Dashboards & visualization | 3000 |
| **Loki** | Log aggregation | 3100 |
| **Promtail** | Log shipping (Docker → Loki) | 9080 (internal) |
| **Node Exporter** | Host-level metrics | 9100 |

## Metrics Architecture

All custom metrics use the `ecommerce_` prefix to avoid collisions with Go/process built-in metrics.

### HTTP Metrics (Middleware)

**File**: `internal/infraestructure/adapters/http/middleware/prometheus.go`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ecommerce_http_requests_total` | Counter | method, route, status_code | Total HTTP requests |
| `ecommerce_http_request_duration_seconds` | Histogram | method, route, status_code | Request latency |
| `ecommerce_http_request_size_bytes` | Histogram | method, route | Request payload size |
| `ecommerce_http_response_size_bytes` | Histogram | method, route, status_code | Response payload size |
| `ecommerce_http_requests_in_flight` | Gauge | — | Concurrent requests (saturation) |
| `ecommerce_http_requests_by_status_family_total` | Counter | status_family, route | Requests by 2xx/3xx/4xx/5xx |

**Key design decisions**:
- `route` label uses Gorilla Mux templates (`/shops/{shop_id}`) not literal URLs — prevents cardinality explosion
- Duration buckets optimized for API workloads: `{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}`
- `/metrics` endpoint excluded from in-flight counting to avoid scraping noise

### DB Pool Metrics (Custom Collector)

**File**: `internal/infraestructure/adapters/metrics/db_pool_collector.go`

| Metric | Type | Description |
|--------|------|-------------|
| `ecommerce_db_pool_open_connections` | Gauge | Open connections to the database |
| `ecommerce_db_pool_idle_connections` | Gauge | Idle connections in the pool |
| `ecommerce_db_pool_in_use_connections` | Gauge | Connections currently in use |
| `ecommerce_db_pool_max_open_connections` | Gauge | Maximum open connections allowed |
| `ecommerce_db_pool_wait_total` | Counter | Total connections waited for |
| `ecommerce_db_pool_wait_duration_seconds_total` | Counter | Total time blocked waiting |

**Implementation**: Uses `prometheus.Collector` interface — reads `sql.DBStats` lazily at scrape time (no background goroutine). Registered via `fx.Invoke` in `main.go`.

### DB Query Metrics (Instrumented Driver)

**File**: `internal/infraestructure/adapters/metrics/instrumented_driver.go`

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ecommerce_db_query_duration_seconds` | Histogram | operation | Query latency |
| `ecommerce_db_query_total` | Counter | operation | Total queries executed |
| `ecommerce_db_query_errors_total` | Counter | operation | Failed queries |

**Operation label values**: `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `OTHER` (BEGIN/COMMIT/ROLLBACK)

**Implementation**: Wraps `lib/pq` at the `database/sql/driver` level. All queries are instrumented transparently — no changes needed in repository code. The driver is registered once via `sync.Once` and used by the DB connection singleton.

### Go Runtime & Process Metrics

These are registered by default in Prometheus's `DefaultRegisterer`:
- `go_goroutines` — number of goroutines
- `go_gc_duration_seconds` — GC pause duration
- `go_memstats_*` — heap, alloc, sys memory stats
- `process_cpu_seconds_total` — CPU usage
- `process_resident_memory_bytes` — RSS memory

## DB Connection Singleton

**File**: `internal/infraestructure/adapters/repositories/postgresql/data_base_connection.go`

`Connect()` uses `sync.Once` to ensure a single `*sql.DB` pool is created regardless of how many repositories call it. The `*sql.DB` is also provided as an FX singleton for the health handler and DB pool collector.

**Why this matters**: Without it, each repository constructor creates a separate pool (~11 pools × 25 max connections = 275 potential connections instead of 25).

## Health Endpoint

**File**: `internal/infraestructure/adapters/http/health_handler.go`

`GET /health` returns:

```json
{
  "status": "healthy|degraded|unhealthy",
  "service": "ecommerce-api",
  "timestamp": "2026-03-15T...",
  "dependencies": {
    "database": { "status": "healthy", "latency_ms": 2 }
  }
}
```

| Condition | Status | HTTP Code |
|-----------|--------|-----------|
| DB ping OK, < 500ms | `healthy` | 200 |
| DB ping OK, > 500ms | `degraded` | 200 |
| DB ping fails | `unhealthy` | 503 |

## Alert Thresholds

Documented as comments in the metric definitions:

| Metric | Condition | Severity |
|--------|-----------|----------|
| HTTP error rate (5xx) | > 1% for 5min | Critical |
| P99 latency | > 2s for 3min | Warning |
| DB pool saturation | in_use/max_open > 80% | Warning |
| DB pool wait duration | > 100ms | Critical |
| Goroutines | > 10,000 | Critical (leak) |

## Grafana Dashboards

Dashboards are **auto-provisioned** via Grafana's provisioning system — no manual import needed.

### Directory Structure

```
grafana/
  provisioning/
    datasources/
      datasources.yml          # Prometheus (uid: "prometheus") + Loki (uid: "loki")
    dashboards/
      dashboards.yml           # Provider pointing to /var/lib/grafana/dashboards
  dashboards/
    sre-golden-signals.json    # HTTP: latency, traffic, errors, saturation
    database-health.json       # Pool stats, query duration/rate/errors
    go-runtime.json            # Goroutines, GC, memory, CPU
    logs-overview.json         # Log volume, by level, errors, HTTP requests
```

### How Provisioning Works

1. Grafana reads `provisioning/datasources/datasources.yml` → creates Prometheus + Loki datasources with fixed UIDs
2. Grafana reads `provisioning/dashboards/dashboards.yml` → scans `/var/lib/grafana/dashboards` for JSON files
3. Dashboard JSONs reference datasources by UID (`prometheus`, `loki`) → everything connects automatically
4. Changes to dashboard JSON files are picked up on Grafana restart

### Dashboard: SRE Golden Signals

**UID**: `ecommerce-sre-golden-signals`

Panels: Total Request Rate, Requests In Flight, Response Time Percentiles (P50/P95/P99), Request Rate by Route, Status Distribution (pie chart), Error Rate (with `or vector(0)` safe division), Request/Response Size P95, Top 10 Routes.

### Dashboard: Database Health

**UID**: `ecommerce-database-health`

Panels: Pool Open vs Max, Pool In-Use vs Idle, Pool Saturation %, Wait Count Rate, Wait Duration, Query Duration by Operation (P50/P95/P99), Query Rate, Query Error Rate.

### Dashboard: Go Runtime

**UID**: `ecommerce-go-runtime`

Panels: Goroutines (with 10k leak threshold line), GC Pause Duration, Heap Memory (alloc/sys/idle/inuse), Process RSS, Process CPU, GC Cycles/sec.

### Dashboard: Logs Overview

**UID**: `ecommerce-logs-overview`

Panels: Log Volume, Logs by Level (stacked, color-coded), Recent Logs, Error Logs Only, HTTP Request Logs, Log Count by Level (24h), Top Operations (24h).

**Important**: All Loki queries include a noise exclusion filter chain (12 patterns) to filter out Loki/Promtail internal noise.

## Structured Logging for Loki

### JSON Format

**File**: `internal/infraestructure/adapters/logs/logger.go`

| Environment | Format | Reason |
|-------------|--------|--------|
| `development` | Text (logrus default) | Human-readable in terminal |
| `dev`, `test`, `production` | JSON | Machine-parseable for Promtail/Loki |

JSON field mapping: `time` → `timestamp`, `msg` → `message`

Default fields injected via hook: `service: "ecommerce-api"`, `environment: <from env>`

### HTTP Request Logging

**File**: `internal/infraestructure/adapters/http/middleware/logging.go`

Each request gets:
- `request_id` — unique per request
- `trace_id` — same as request_id (distributed tracing compatibility)
- `event` — `http_request_started` or `http_request_completed`
- `method`, `path`, `remote_addr`, `user_agent`
- `status_code`, `duration_ms` (on completion)

### Promtail Pipeline

**File**: `promtail-config.yml`

Extracts `level` and `service` from JSON log body as Loki labels. Other fields (`trace_id`, `method`, `route`) stay in the log body — queryable via `| json` in Loki.

## Docker Compose

The observability stack is defined in `docker-compose.yml`:

```yaml
# Grafana provisioning volumes
- ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources
- ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards
- ./grafana/dashboards:/var/lib/grafana/dashboards
```

### Local Testing URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| App Health | http://localhost:8080/health | — |
| App Metrics | http://localhost:8080/metrics | — |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |

## Adding New Metrics

1. Define metrics with `promauto.New*` in the appropriate package under `internal/infraestructure/adapters/metrics/`
2. Always use `ecommerce_` prefix
3. For custom collectors (lazy evaluation), implement `prometheus.Collector` interface
4. Register collectors in `metrics/runtime_collectors.go` via `prometheus.MustRegister()`
5. Add panels to the relevant Grafana dashboard JSON
6. Document alert thresholds as comments on metric definitions
