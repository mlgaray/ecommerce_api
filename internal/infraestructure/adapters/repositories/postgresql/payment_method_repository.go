package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type PaymentMethodSQLRepository struct {
	db *sql.DB
}

func NewPaymentMethodRepository(dataBaseConnection DataBaseConnection) ports.PaymentMethodRepository {
	return &PaymentMethodSQLRepository{
		db: dataBaseConnection.Connect(),
	}
}

func (r *PaymentMethodSQLRepository) GetAll(ctx context.Context) ([]*models.PaymentMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getAllWithTx(ctx, tx)
	}
	return r.getAllWithDB(ctx)
}

func (r *PaymentMethodSQLRepository) getAllWithTx(ctx context.Context, tx *sql.Tx) ([]*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE is_active = true`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPaymentMethods(rows)
}

func (r *PaymentMethodSQLRepository) getAllWithDB(ctx context.Context) ([]*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE is_active = true`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPaymentMethods(rows)
}

func (r *PaymentMethodSQLRepository) scanPaymentMethods(rows *sql.Rows) ([]*models.PaymentMethod, error) {
	var methods []*models.PaymentMethod
	for rows.Next() {
		var method models.PaymentMethod
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

func (r *PaymentMethodSQLRepository) GetByID(ctx context.Context, id int) (*models.PaymentMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getByIDWithTx(ctx, tx, id)
	}
	return r.getByIDWithDB(ctx, id)
}

func (r *PaymentMethodSQLRepository) getByIDWithTx(ctx context.Context, tx *sql.Tx, id int) (*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE id = $1`

	var method models.PaymentMethod
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

func (r *PaymentMethodSQLRepository) getByIDWithDB(ctx context.Context, id int) (*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE id = $1`

	var method models.PaymentMethod
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

func (r *PaymentMethodSQLRepository) GetByCode(ctx context.Context, code models.PaymentMethodCode) (*models.PaymentMethod, error) {
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.getByCodeWithTx(ctx, tx, code)
	}
	return r.getByCodeWithDB(ctx, code)
}

func (r *PaymentMethodSQLRepository) getByCodeWithTx(ctx context.Context, tx *sql.Tx, code models.PaymentMethodCode) (*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE code = $1`

	var method models.PaymentMethod
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

func (r *PaymentMethodSQLRepository) getByCodeWithDB(ctx context.Context, code models.PaymentMethodCode) (*models.PaymentMethod, error) {
	const query = `SELECT id, name, code, description, is_active FROM payment_methods WHERE code = $1`

	var method models.PaymentMethod
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
