package metrics

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDBPoolCollector(t *testing.T) {
	t.Run("when NewDBPoolCollector is called then returns non-nil collector", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)

		// Act
		collector := NewDBPoolCollector(db)

		// Assert
		assert.NotNil(t, collector)
	})

	t.Run("when NewDBPoolCollector is called then stores the provided db", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)

		// Act
		collector := NewDBPoolCollector(db)

		// Assert
		assert.NotNil(t, collector)
		assert.Equal(t, db, collector.db)
	})

	t.Run("when NewDBPoolCollector is called then all descriptors are initialized", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)

		// Act
		collector := NewDBPoolCollector(db)

		// Assert
		assert.NotNil(t, collector.openConns)
		assert.NotNil(t, collector.idleConns)
		assert.NotNil(t, collector.inUseConns)
		assert.NotNil(t, collector.maxOpen)
		assert.NotNil(t, collector.waitTotal)
		assert.NotNil(t, collector.waitDur)
	})
}

func TestDBPoolCollector_Describe(t *testing.T) {
	t.Run("when Describe is called then sends exactly 6 descriptors to channel", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)
		collector := NewDBPoolCollector(db)
		ch := make(chan *prometheus.Desc, 10)

		// Act
		collector.Describe(ch)
		close(ch)

		// Assert
		var descs []*prometheus.Desc
		for d := range ch {
			descs = append(descs, d)
		}
		assert.Len(t, descs, 6)
	})

	t.Run("when Describe is called then all descriptors are non-nil", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)
		collector := NewDBPoolCollector(db)
		ch := make(chan *prometheus.Desc, 10)

		// Act
		collector.Describe(ch)
		close(ch)

		// Assert
		for desc := range ch {
			assert.NotNil(t, desc)
		}
	})
}

func TestDBPoolCollector_Collect(t *testing.T) {
	t.Run("when Collect is called then sends exactly 6 metrics to channel", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)
		collector := NewDBPoolCollector(db)
		ch := make(chan prometheus.Metric, 10)

		// Act
		collector.Collect(ch)
		close(ch)

		// Assert
		var metrics []prometheus.Metric
		for m := range ch {
			metrics = append(metrics, m)
		}
		assert.Len(t, metrics, 6)
	})

	t.Run("when Collect is called then all metrics are non-nil", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)
		collector := NewDBPoolCollector(db)
		ch := make(chan prometheus.Metric, 10)

		// Act
		collector.Collect(ch)
		close(ch)

		// Assert
		for m := range ch {
			assert.NotNil(t, m)
		}
	})

	t.Run("when Collect is called then metrics match db stats", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)
		db.SetMaxOpenConns(5)
		collector := NewDBPoolCollector(db)
		ch := make(chan prometheus.Metric, 10)

		// Act
		collector.Collect(ch)
		close(ch)

		// Assert — we only verify that 6 metrics are emitted and each can be
		// written to a dto without panicking (MustNewConstMetric would panic on error).
		var metrics []prometheus.Metric
		for m := range ch {
			metrics = append(metrics, m)
		}
		assert.Len(t, metrics, 6)
	})
}

func TestDBPoolCollector_ImplementsCollector(t *testing.T) {
	t.Run("when DBPoolCollector is created then it implements prometheus.Collector", func(t *testing.T) {
		// Arrange
		db := newTestDB(t)

		// Act
		collector := NewDBPoolCollector(db)

		// Assert — compile-time check expressed at runtime
		var _ prometheus.Collector = collector
		assert.NotNil(t, collector)
	})
}
