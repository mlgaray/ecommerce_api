package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	StoreVisitRepositoryField = "store_visit_repository"
	StoreVisitRecordFuncField = "record_visit"
)

type StoreVisitSQLRepository struct {
	db *sql.DB
}

func NewStoreVisitRepository(dataBaseConnection DataBaseConnection) ports.StoreVisitRepository {
	return &StoreVisitSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

// RecordVisit inserts a new visit record for the given shop.
func (r *StoreVisitSQLRepository) RecordVisit(ctx context.Context, shopID int) error {
	query := `INSERT INTO store_visits (shop_id) VALUES ($1)`

	_, err := r.db.ExecContext(ctx, query, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreVisitRepositoryField,
			"function": StoreVisitRecordFuncField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error recording store visit")
		return fmt.Errorf("error recording store visit: %w", err)
	}

	return nil
}
