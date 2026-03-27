package postgresql

import (
	"context"
	"database/sql"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignupSQLRepository struct {
	db        *sql.DB
	userRepo  ports.UserRepository
	shopRepo  ports.ShopRepository
	roleRepo  ports.RoleRepository
	staffRepo ports.StaffRepository
}

func NewSignupRepository(
	dataBaseConnection DataBaseConnection,
	userRepo ports.UserRepository,
	shopRepo ports.ShopRepository,
	roleRepo ports.RoleRepository,
	staffRepo ports.StaffRepository,
) ports.SignupRepository {
	return &SignupSQLRepository{
		db:        dataBaseConnection.Connect(),
		userRepo:  userRepo,
		shopRepo:  shopRepo,
		roleRepo:  roleRepo,
		staffRepo: staffRepo,
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

	// 2. Crear shop usando ShopRepository (sin userID)
	_, err = r.shopRepo.Create(txCtx, shop)
	if err != nil {
		return nil, err
	}

	// 3. Crear staff (user ↔ shop) via StaffRepository
	staff, err := r.staffRepo.Create(txCtx, createdUser.ID, shop.ID)
	if err != nil {
		return nil, err
	}

	// 4. Asignar rol owner via staff_roles
	ownerRole, err := r.roleRepo.GetByName(txCtx, "owner")
	if err != nil {
		return nil, err
	}

	err = r.staffRepo.AssignRole(txCtx, staff.ID, ownerRole.ID)
	if err != nil {
		return nil, err
	}

	// Agregar el rol al usuario en memoria
	createdUser.Roles = append(createdUser.Roles, ownerRole)

	// 5. Commit de la transacción
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
