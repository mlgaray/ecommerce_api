package metrics

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ecommerce_db_query_duration_seconds — Histogram of query latency by operation type
	// Alert: p99 > 500ms for 3min → Warning
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ecommerce_db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5},
		},
		[]string{"operation"},
	)

	// ecommerce_db_query_total — Counter of total queries by operation type
	dbQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ecommerce_db_query_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"},
	)

	// ecommerce_db_query_errors_total — Counter of failed queries by operation type
	dbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ecommerce_db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"operation"},
	)

	registerOnce sync.Once
	driverName   string
)

// RegisterInstrumentedDriver registers the instrumented driver once and returns its name.
// Safe to call multiple times — subsequent calls return the cached name.
func RegisterInstrumentedDriver() string {
	registerOnce.Do(func() {
		driverName = "instrumented-postgres"
		sql.Register(driverName, &instrumentedDriver{wrapped: &pq.Driver{}})
	})
	return driverName
}

// SQL operation constants for metrics labeling.
const (
	opSelect  = "SELECT"
	opInsert  = "INSERT"
	opUpdate  = "UPDATE"
	opDelete  = "DELETE"
	opOther   = "OTHER"
	opUnknown = "UNKNOWN"
)

// detectOperation extracts the SQL operation from the first keyword of the query.
func detectOperation(query string) string {
	q := strings.TrimSpace(query)
	if len(q) == 0 {
		return opUnknown
	}
	// Find first space or end of string
	idx := strings.IndexByte(q, ' ')
	if idx == -1 {
		idx = len(q)
	}
	switch strings.ToUpper(q[:idx]) {
	case opSelect:
		return opSelect
	case opInsert:
		return opInsert
	case opUpdate:
		return opUpdate
	case opDelete:
		return opDelete
	default:
		return opOther
	}
}

// --- Driver-level wrapper ---

type instrumentedDriver struct {
	wrapped driver.Driver
}

func (d *instrumentedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.wrapped.Open(name)
	if err != nil {
		return nil, err
	}
	return &instrumentedConn{wrapped: conn}, nil
}

// --- Connection-level wrapper ---

type instrumentedConn struct {
	wrapped driver.Conn
}

func (c *instrumentedConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.wrapped.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &instrumentedStmt{wrapped: stmt, query: query}, nil
}

func (c *instrumentedConn) Close() error {
	return c.wrapped.Close()
}

func (c *instrumentedConn) Begin() (driver.Tx, error) {
	return c.wrapped.Begin() //nolint:staticcheck // required by driver.Conn interface
}

// QueryerContext instruments direct query execution (used by database/sql for simple queries).
func (c *instrumentedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := c.wrapped.(driver.QueryerContext); ok {
		op := detectOperation(query)
		start := time.Now()
		rows, err := qc.QueryContext(ctx, query, args)
		duration := time.Since(start).Seconds()

		dbQueryTotal.WithLabelValues(op).Inc()
		dbQueryDuration.WithLabelValues(op).Observe(duration)
		if err != nil {
			dbQueryErrors.WithLabelValues(op).Inc()
		}
		return rows, err
	}
	// Fallback: prepare + execute (instrumented via stmt)
	stmt, err := c.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	sqc, ok := stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, fmt.Errorf("statement does not implement StmtQueryContext")
	}
	return sqc.QueryContext(ctx, args)
}

// ExecerContext instruments direct exec calls.
func (c *instrumentedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := c.wrapped.(driver.ExecerContext); ok {
		op := detectOperation(query)
		start := time.Now()
		result, err := ec.ExecContext(ctx, query, args)
		duration := time.Since(start).Seconds()

		dbQueryTotal.WithLabelValues(op).Inc()
		dbQueryDuration.WithLabelValues(op).Observe(duration)
		if err != nil {
			dbQueryErrors.WithLabelValues(op).Inc()
		}
		return result, err
	}
	stmt, err := c.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	sec, ok := stmt.(driver.StmtExecContext)
	if !ok {
		return nil, fmt.Errorf("statement does not implement StmtExecContext")
	}
	return sec.ExecContext(ctx, args)
}

// --- Statement-level wrapper ---

type instrumentedStmt struct {
	wrapped driver.Stmt
	query   string
}

func (s *instrumentedStmt) Close() error {
	return s.wrapped.Close()
}

func (s *instrumentedStmt) NumInput() int {
	return s.wrapped.NumInput()
}

func (s *instrumentedStmt) Exec(args []driver.Value) (driver.Result, error) {
	op := detectOperation(s.query)
	start := time.Now()
	result, err := s.wrapped.Exec(args) //nolint:staticcheck // required by driver.Stmt
	duration := time.Since(start).Seconds()

	dbQueryTotal.WithLabelValues(op).Inc()
	dbQueryDuration.WithLabelValues(op).Observe(duration)
	if err != nil {
		dbQueryErrors.WithLabelValues(op).Inc()
	}
	return result, err
}

func (s *instrumentedStmt) Query(args []driver.Value) (driver.Rows, error) {
	op := detectOperation(s.query)
	start := time.Now()
	rows, err := s.wrapped.Query(args) //nolint:staticcheck // required by driver.Stmt
	duration := time.Since(start).Seconds()

	dbQueryTotal.WithLabelValues(op).Inc()
	dbQueryDuration.WithLabelValues(op).Observe(duration)
	if err != nil {
		dbQueryErrors.WithLabelValues(op).Inc()
	}
	return rows, err
}
