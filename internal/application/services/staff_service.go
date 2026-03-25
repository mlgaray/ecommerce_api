package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type StaffServiceImpl struct {
	staffRepository ports.StaffRepository
	roleRepository  ports.RoleRepository
	assetService    ports.AssetService
	authService     ports.AuthService
}

func NewStaffService(
	staffRepository ports.StaffRepository,
	roleRepository ports.RoleRepository,
	assetService ports.AssetService,
	authService ports.AuthService,
) *StaffServiceImpl {
	return &StaffServiceImpl{
		staffRepository: staffRepository,
		roleRepository:  roleRepository,
		assetService:    assetService,
		authService:     authService,
	}
}

func (s *StaffServiceImpl) GetShopRolesByUserID(ctx context.Context, userID int) ([]claims.ShopRole, error) {
	return s.staffRepository.GetShopRolesByUserID(ctx, userID)
}

func (s *StaffServiceImpl) Create(ctx context.Context, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error) {
	if err := s.validateRole(ctx, roleID); err != nil {
		return nil, err
	}

	if err := s.validateUserFields(user, true); err != nil {
		return nil, err
	}

	user.IsActive = true

	if err := s.hashPasswordIfProvided(ctx, user); err != nil {
		return nil, err
	}

	profileImage, err := s.uploadImage(ctx, imageBuffer, shopID)
	if err != nil {
		return nil, err
	}

	staff, err := s.staffRepository.CreateWithUser(ctx, shopID, user, roleID, profileImage)
	if err != nil {
		s.cleanupAsset(ctx, profileImage, "create")
		return nil, err
	}

	return staff, nil
}

func (s *StaffServiceImpl) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) ([]*models.Staff, error) {
	return s.staffRepository.GetAllByShopIDWithFilters(ctx, shopID, filters)
}

func (s *StaffServiceImpl) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) (int, error) {
	return s.staffRepository.CountByShopIDWithFilters(ctx, shopID, filters)
}

func (s *StaffServiceImpl) GetByID(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	return s.staffRepository.GetByID(ctx, staffID, shopID)
}

func (s *StaffServiceImpl) Update(ctx context.Context, staffID, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error) {
	if err := s.validateRole(ctx, roleID); err != nil {
		return nil, err
	}

	if err := s.validateUserFields(user, false); err != nil {
		return nil, err
	}

	// Protect owner from being modified via staff management
	existingStaff, err := s.staffRepository.GetByID(ctx, staffID, shopID)
	if err != nil {
		return nil, err
	}
	if s.hasOwnerRole(existingStaff) {
		return nil, &errors.ValidationError{Message: errors.OwnerCannotBeModified}
	}

	if err := s.hashPasswordIfProvided(ctx, user); err != nil {
		return nil, err
	}

	// Upload new image if provided
	profileImage, err := s.uploadImage(ctx, imageBuffer, shopID)
	if err != nil {
		return nil, err
	}

	staff, err := s.staffRepository.Update(ctx, staffID, shopID, user, roleID, profileImage)
	if err != nil {
		s.cleanupAsset(ctx, profileImage, "update")
		return nil, err
	}

	// Clean up old image from cloud storage after successful update
	s.cleanupOldImage(ctx, existingStaff, imageBuffer)

	return staff, nil
}

func (s *StaffServiceImpl) Delete(ctx context.Context, staffID, shopID int) error {
	staff, err := s.staffRepository.GetByID(ctx, staffID, shopID)
	if err != nil {
		return err
	}
	if s.hasOwnerRole(staff) {
		return &errors.ValidationError{Message: errors.OwnerCannotBeDeleted}
	}
	return s.staffRepository.Delete(ctx, staffID, shopID)
}

func (s *StaffServiceImpl) ToggleStatus(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	staff, err := s.staffRepository.GetByID(ctx, staffID, shopID)
	if err != nil {
		return nil, err
	}
	if s.hasOwnerRole(staff) {
		return nil, &errors.ValidationError{Message: errors.OwnerStatusCannotChange}
	}
	return s.staffRepository.ToggleStatus(ctx, staffID, shopID)
}

func (s *StaffServiceImpl) hashPasswordIfProvided(ctx context.Context, user *models.User) error {
	if user.Password == "" {
		return nil
	}
	hashedPassword, err := s.authService.HashPassword(ctx, user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return nil
}

func (s *StaffServiceImpl) uploadImage(ctx context.Context, imageBuffer []byte, shopID int) (*models.Image, error) {
	if len(imageBuffer) == 0 {
		return nil, nil
	}
	folder := fmt.Sprintf("shop_%d/staff", shopID)
	return s.assetService.Upload(ctx, imageBuffer, folder)
}

func (s *StaffServiceImpl) cleanupAsset(ctx context.Context, image *models.Image, operation string) {
	if image == nil || image.StorageRef == "" {
		return
	}
	if err := s.assetService.Delete(ctx, image.StorageRef); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        "staff_service",
			"function":    operation,
			"storage_ref": image.StorageRef,
			"error":       err.Error(),
		}).Warn("Failed to clean up uploaded image")
	}
}

func (s *StaffServiceImpl) cleanupOldImage(ctx context.Context, existingStaff *models.Staff, imageBuffer []byte) {
	if len(imageBuffer) == 0 || existingStaff.User == nil || existingStaff.User.ProfileImage == nil {
		return
	}
	oldRef := existingStaff.User.ProfileImage.StorageRef
	if oldRef == "" {
		return
	}
	if err := s.assetService.Delete(ctx, oldRef); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        "staff_service",
			"function":    "update",
			"storage_ref": oldRef,
			"error":       err.Error(),
		}).Warn("Failed to delete old profile image from cloud storage")
	}
}

// hasOwnerRole checks if a staff member has the owner role.
func (s *StaffServiceImpl) hasOwnerRole(staff *models.Staff) bool {
	for _, role := range staff.Roles {
		if role.Name == ownerRoleName {
			return true
		}
	}
	return false
}

func (s *StaffServiceImpl) validateRole(ctx context.Context, roleID int) error {
	if roleID <= 0 {
		return &errors.ValidationError{Message: errors.RoleNotFound}
	}

	role, err := s.roleRepository.GetByID(ctx, roleID)
	if err != nil {
		return &errors.ValidationError{Message: errors.RoleNotFound}
	}

	if role.Name == ownerRoleName {
		return &errors.ValidationError{Message: errors.OwnerRoleNotAllowed}
	}

	return nil
}

const (
	ownerRoleName     = "owner"
	minPasswordLength = 6
)

func (s *StaffServiceImpl) validateUserFields(user *models.User, requirePassword bool) error {
	if strings.TrimSpace(user.Name) == "" {
		return &errors.ValidationError{Message: "user_name_is_required"}
	}
	if strings.TrimSpace(user.LastName) == "" {
		return &errors.ValidationError{Message: "user_last_name_is_required"}
	}
	if strings.TrimSpace(user.Email) == "" {
		return &errors.ValidationError{Message: "user_email_is_required"}
	}
	if requirePassword {
		if len(user.Password) < minPasswordLength {
			return &errors.ValidationError{Message: "password_must_be_at_least_6_characters"}
		}
	} else if user.Password != "" && len(user.Password) < minPasswordLength {
		return &errors.ValidationError{Message: "password_must_be_at_least_6_characters"}
	}
	return nil
}
