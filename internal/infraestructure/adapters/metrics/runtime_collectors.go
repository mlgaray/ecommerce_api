package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

// RegisterRuntimeCollectors registers the DB pool collector with Prometheus.
// Go runtime (goroutines, GC, memory) and process (CPU, RSS, FDs) collectors
// are already registered by default in prometheus.DefaultRegisterer.
// Called once at startup via fx.Invoke.
func RegisterRuntimeCollectors(db *sql.DB) {
	// DB pool stats (lazy — reads sql.DBStats at scrape time)
	prometheus.MustRegister(NewDBPoolCollector(db))
}
