package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectOperation(t *testing.T) {
	t.Run("when query starts with SELECT then returns SELECT", func(t *testing.T) {
		// Arrange
		query := "SELECT * FROM users"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opSelect, result)
	})

	t.Run("when query starts with lowercase select then returns SELECT", func(t *testing.T) {
		// Arrange
		query := "select * from users"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opSelect, result)
	})

	t.Run("when query starts with mixed case Select then returns SELECT", func(t *testing.T) {
		// Arrange
		query := "Select id FROM users WHERE id = $1"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opSelect, result)
	})

	t.Run("when query starts with INSERT then returns INSERT", func(t *testing.T) {
		// Arrange
		query := "INSERT INTO users (name, email) VALUES ($1, $2)"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opInsert, result)
	})

	t.Run("when query starts with lowercase insert then returns INSERT", func(t *testing.T) {
		// Arrange
		query := "insert into products (name) values ($1)"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opInsert, result)
	})

	t.Run("when query starts with UPDATE then returns UPDATE", func(t *testing.T) {
		// Arrange
		query := "UPDATE users SET name = $1 WHERE id = $2"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opUpdate, result)
	})

	t.Run("when query starts with lowercase update then returns UPDATE", func(t *testing.T) {
		// Arrange
		query := "update orders set status = $1 where id = $2"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opUpdate, result)
	})

	t.Run("when query starts with DELETE then returns DELETE", func(t *testing.T) {
		// Arrange
		query := "DELETE FROM users WHERE id = $1"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opDelete, result)
	})

	t.Run("when query starts with lowercase delete then returns DELETE", func(t *testing.T) {
		// Arrange
		query := "delete from sessions where token = $1"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opDelete, result)
	})

	t.Run("when query starts with CREATE then returns OTHER", func(t *testing.T) {
		// Arrange
		query := "CREATE TABLE users (id serial PRIMARY KEY)"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opOther, result)
	})

	t.Run("when query starts with DROP then returns OTHER", func(t *testing.T) {
		// Arrange
		query := "DROP TABLE users"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opOther, result)
	})

	t.Run("when query starts with EXPLAIN then returns OTHER", func(t *testing.T) {
		// Arrange
		query := "EXPLAIN SELECT * FROM users"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opOther, result)
	})

	t.Run("when query has leading whitespace then returns correct operation", func(t *testing.T) {
		// Arrange
		query := "  SELECT id, name FROM products"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opSelect, result)
	})

	t.Run("when query has leading tabs then returns correct operation", func(t *testing.T) {
		// Arrange
		query := "\t\tINSERT INTO logs (message) VALUES ($1)"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opInsert, result)
	})

	t.Run("when query is empty then returns UNKNOWN", func(t *testing.T) {
		// Arrange
		query := ""

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opUnknown, result)
	})

	t.Run("when query is only whitespace then returns UNKNOWN", func(t *testing.T) {
		// Arrange
		query := "   "

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opUnknown, result)
	})

	t.Run("when query is a single keyword with no spaces then returns correct operation", func(t *testing.T) {
		// Arrange
		query := "SELECT"

		// Act
		result := detectOperation(query)

		// Assert
		assert.Equal(t, opSelect, result)
	})
}

func TestRegisterInstrumentedDriver(t *testing.T) {
	t.Run("when RegisterInstrumentedDriver is called then returns instrumented-postgres", func(t *testing.T) {
		// Act
		name := RegisterInstrumentedDriver()

		// Assert
		assert.Equal(t, "instrumented-postgres", name)
	})

	t.Run("when RegisterInstrumentedDriver is called multiple times then returns same driver name", func(t *testing.T) {
		// Act
		first := RegisterInstrumentedDriver()
		second := RegisterInstrumentedDriver()
		third := RegisterInstrumentedDriver()

		// Assert
		assert.Equal(t, first, second)
		assert.Equal(t, second, third)
		assert.Equal(t, "instrumented-postgres", third)
	})
}
