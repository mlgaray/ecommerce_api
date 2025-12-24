package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type DeliveryMethodSQLRepository struct {
	db *sql.DB
}

func NewDeliveryMethodRepository(dataBaseConnection DataBaseConnection) ports.DeliveryMethodRepository {
	return &DeliveryMethodSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

func (r *DeliveryMethodSQLRepository) GetAll(ctx context.Context) ([]*models.DeliveryMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getAllWithTx(ctx, tx)
	}
	return r.getAllWithDB(ctx)
}

func (r *DeliveryMethodSQLRepository) getAllWithTx(ctx context.Context, tx *sql.Tx) ([]*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE is_active = true`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDeliveryMethods(rows)
}

func (r *DeliveryMethodSQLRepository) getAllWithDB(ctx context.Context) ([]*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE is_active = true`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDeliveryMethods(rows)
}

func (r *DeliveryMethodSQLRepository) scanDeliveryMethods(rows *sql.Rows) ([]*models.DeliveryMethod, error) {
	var methods []*models.DeliveryMethod
	for rows.Next() {
		var method models.DeliveryMethod
		var description sql.NullString
		err := rows.Scan(&method.ID, &method.Name, &method.Code, &description, &method.IsActive)
		if err != nil {
			return nil, err
		}
		if description.Valid {
			method.Description = description.String
		}
		methods = append(methods, &method)
	}
	return methods, rows.Err()
}

func (r *DeliveryMethodSQLRepository) GetByID(ctx context.Context, id int) (*models.DeliveryMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getByIDWithTx(ctx, tx, id)
	}
	return r.getByIDWithDB(ctx, id)
}

func (r *DeliveryMethodSQLRepository) getByIDWithTx(ctx context.Context, tx *sql.Tx, id int) (*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE id = $1`

	var method models.DeliveryMethod
	var description sql.NullString
	err := tx.QueryRowContext(ctx, query, id).Scan(&method.ID, &method.Name, &method.Code, &description, &method.IsActive)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		method.Description = description.String
	}
	return &method, nil
}

func (r *DeliveryMethodSQLRepository) getByIDWithDB(ctx context.Context, id int) (*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE id = $1`

	var method models.DeliveryMethod
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&method.ID, &method.Name, &method.Code, &description, &method.IsActive)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		method.Description = description.String
	}
	return &method, nil
}

func (r *DeliveryMethodSQLRepository) GetByCode(ctx context.Context, code models.DeliveryMethodCode) (*models.DeliveryMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getByCodeWithTx(ctx, tx, code)
	}
	return r.getByCodeWithDB(ctx, code)
}

func (r *DeliveryMethodSQLRepository) getByCodeWithTx(ctx context.Context, tx *sql.Tx, code models.DeliveryMethodCode) (*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE code = $1`

	var method models.DeliveryMethod
	var description sql.NullString
	err := tx.QueryRowContext(ctx, query, code).Scan(&method.ID, &method.Name, &method.Code, &description, &method.IsActive)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		method.Description = description.String
	}
	return &method, nil
}

func (r *DeliveryMethodSQLRepository) getByCodeWithDB(ctx context.Context, code models.DeliveryMethodCode) (*models.DeliveryMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM delivery_methods WHERE code = $1`

	var method models.DeliveryMethod
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, code).Scan(&method.ID, &method.Name, &method.Code, &description, &method.IsActive)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		method.Description = description.String
	}
	return &method, nil
}
