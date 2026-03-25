package services

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type PermissionServiceImpl struct {
	data atomic.Value // stores map[string]map[string]struct{}
	repo ports.PermissionRepository
}

func NewPermissionService(repo ports.PermissionRepository) *PermissionServiceImpl {
	svc := &PermissionServiceImpl{repo: repo}
	// Initialize with empty map (never nil)
	svc.data.Store(make(map[string]map[string]struct{}))
	// Initial load — if it fails, starts with empty map (fail-safe, denies all)
	svc.loadPermissions()
	return svc
}

// StartRefreshLoop starts the periodic refresh goroutine.
// Stops cleanly when ctx is cancelled. Intended to be called via FX lifecycle.
func (s *PermissionServiceImpl) StartRefreshLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.loadPermissions()
		case <-ctx.Done():
			return
		}
	}
}

func (s *PermissionServiceImpl) loadPermissions() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := s.repo.GetAllRolePermissions(ctx)
	if err != nil {
		// Use stdlib log as fallback — structured logger may not be initialized yet at startup
		log.Printf("[permission_service] WARNING: Failed to load permissions: %v (keeping previous state)", err)
		return
	}

	// Build new map and do atomic swap
	newMap := make(map[string]map[string]struct{}, len(data))
	for role, perms := range data {
		newMap[role] = make(map[string]struct{}, len(perms))
		for _, p := range perms {
			newMap[role][p] = struct{}{}
		}
	}
	s.data.Store(newMap)

	log.Printf("[permission_service] Permissions loaded: %d roles", len(newMap))
}

// HasPermission checks if a role has a specific permission.
// Lock-free read via atomic.Value — safe for concurrent access.
func (s *PermissionServiceImpl) HasPermission(role, permission string) bool {
	rolePerms, ok := s.data.Load().(map[string]map[string]struct{})
	if !ok {
		return false
	}
	perms, exists := rolePerms[role]
	if !exists {
		return false
	}
	_, has := perms[permission]
	return has
}

// GetPermissions returns all permissions for a role.
// Used during sign-in to embed permissions in the JWT token.
func (s *PermissionServiceImpl) GetPermissions(role string) []string {
	rolePerms, ok := s.data.Load().(map[string]map[string]struct{})
	if !ok {
		return nil
	}
	perms, ok := rolePerms[role]
	if !ok {
		return nil
	}
	result := make([]string, 0, len(perms))
	for p := range perms {
		result = append(result, p)
	}
	return result
}
