package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignupSQLRepository struct {
	db       *sql.DB
	userRepo ports.UserRepository
	shopRepo ports.ShopRepository
	roleRepo ports.RoleRepository
}

func NewSignupRepository(dataBaseConnection DataBaseConnection, userRepo ports.UserRepository, shopRepo ports.ShopRepository, roleRepo ports.RoleRepository) ports.SignupRepository {
	return &SignupSQLRepository{
		db:       dataBaseConnection.Connect(),
		userRepo: userRepo,
		shopRepo: shopRepo,
		roleRepo: roleRepo,
	}
}

func (r *SignupSQLRepository) CreateUserWithShop(ctx context.Context, user *models.User, shop *models.Shop) (*models.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// TODO: Log rollback error but don't override original error
				_ = rollbackErr
			}
		}
	}()

	// Crear contexto con transacción
	txCtx := context.WithValue(ctx, TxContextKey, tx)

	// 1. Crear usuario usando UserRepository
	createdUser, err := r.userRepo.Create(txCtx, user)
	if err != nil {
		return nil, err
	}

	// 2. Asignar rol admin por defecto
	adminRole, err := r.roleRepo.GetByName(txCtx, "admin")
	if err != nil {
		return nil, err
	}

	err = r.userRepo.AssignRole(txCtx, createdUser.ID, adminRole.ID)
	if err != nil {
		return nil, err
	}

	// Agregar el rol al usuario en memoria
	createdUser.Roles = append(createdUser.Roles, adminRole)

	// 3. Asignar UserID al shop y crearlo usando ShopRepository
	shop.UserID = createdUser.ID
	_, err = r.shopRepo.Create(txCtx, shop)
	if err != nil {
		return nil, err
	}

	// 4. Commit de la transacción
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
