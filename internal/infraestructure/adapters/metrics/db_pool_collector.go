package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

// DBPoolCollector implements prometheus.Collector and reads sql.DBStats lazily at scrape time.
// No background goroutine needed — Prometheus calls Collect() on each scrape.
type DBPoolCollector struct {
	db *sql.DB

	openConns  *prometheus.Desc
	idleConns  *prometheus.Desc
	inUseConns *prometheus.Desc
	maxOpen    *prometheus.Desc
	waitTotal  *prometheus.Desc
	waitDur    *prometheus.Desc
}

// NewDBPoolCollector creates a collector for the given *sql.DB.
func NewDBPoolCollector(db *sql.DB) *DBPoolCollector {
	return &DBPoolCollector{
		db: db,
		openConns: prometheus.NewDesc(
			"ecommerce_db_pool_open_connections",
			"Number of open connections to the database",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			"ecommerce_db_pool_idle_connections",
			"Number of idle connections in the pool",
			nil, nil,
		),
		inUseConns: prometheus.NewDesc(
			// Alert: in_use / max_open > 0.8 → Warning (pool saturation)
			"ecommerce_db_pool_in_use_connections",
			"Number of connections currently in use",
			nil, nil,
		),
		maxOpen: prometheus.NewDesc(
			"ecommerce_db_pool_max_open_connections",
			"Maximum number of open connections allowed",
			nil, nil,
		),
		waitTotal: prometheus.NewDesc(
			"ecommerce_db_pool_wait_total",
			"Total number of connections waited for",
			nil, nil,
		),
		waitDur: prometheus.NewDesc(
			// Alert: wait_duration > 100ms → Critical
			"ecommerce_db_pool_wait_duration_seconds_total",
			"Total time blocked waiting for a connection",
			nil, nil,
		),
	}
}

func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.openConns
	ch <- c.idleConns
	ch <- c.inUseConns
	ch <- c.maxOpen
	ch <- c.waitTotal
	ch <- c.waitDur
}

func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(c.openConns, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(c.inUseConns, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(c.maxOpen, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(c.waitTotal, prometheus.CounterValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.waitDur, prometheus.CounterValue, stats.WaitDuration.Seconds())
}
