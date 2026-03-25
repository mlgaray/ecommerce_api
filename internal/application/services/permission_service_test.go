package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/mocks"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func samplePermissions() map[string][]string {
	return map[string][]string{
		"admin":  {"create_product", "delete_product", "view_orders"},
		"seller": {"create_product", "view_orders"},
		"viewer": {"view_orders"},
	}
}

// newServiceWithData builds a PermissionServiceImpl whose constructor load
// succeeds with the given data.
func newServiceWithData(t *testing.T, data map[string][]string) *PermissionServiceImpl {
	repo := mocks.NewPermissionRepository(t)
	repo.EXPECT().
		GetAllRolePermissions(mock.Anything).
		Return(data, nil).
		Once()
	return NewPermissionService(repo)
}

// ---------------------------------------------------------------------------
// NewPermissionService
// ---------------------------------------------------------------------------

func TestPermissionService_NewPermissionService(t *testing.T) {
	t.Run("constructor loads permissions on init", func(t *testing.T) {
		// Arrange & Act
		svc := newServiceWithData(t, samplePermissions())

		// Assert — data must be available immediately after construction
		assert.True(t, svc.HasPermission("admin", "create_product"))
		assert.True(t, svc.HasPermission("seller", "view_orders"))
	})

	t.Run("constructor handles load failure gracefully and starts with empty map", func(t *testing.T) {
		// Arrange
		repo := mocks.NewPermissionRepository(t)
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(nil, errors.New("db connection refused")).
			Once()

		// Act
		svc := NewPermissionService(repo)

		// Assert — service should exist and deny everything (empty map)
		assert.False(t, svc.HasPermission("admin", "create_product"))
		assert.Nil(t, svc.GetPermissions("admin"))
	})
}

// ---------------------------------------------------------------------------
// HasPermission
// ---------------------------------------------------------------------------

func TestPermissionService_HasPermission(t *testing.T) {
	svc := newServiceWithData(t, samplePermissions())

	t.Run("returns true when role has the permission", func(t *testing.T) {
		assert.True(t, svc.HasPermission("admin", "delete_product"))
		assert.True(t, svc.HasPermission("viewer", "view_orders"))
	})

	t.Run("returns false when role exists but permission does not", func(t *testing.T) {
		assert.False(t, svc.HasPermission("viewer", "delete_product"))
		assert.False(t, svc.HasPermission("seller", "delete_product"))
	})

	t.Run("returns false when role does not exist", func(t *testing.T) {
		assert.False(t, svc.HasPermission("unknown_role", "view_orders"))
		assert.False(t, svc.HasPermission("", "view_orders"))
	})
}

// ---------------------------------------------------------------------------
// GetPermissions
// ---------------------------------------------------------------------------

func TestPermissionService_GetPermissions(t *testing.T) {
	svc := newServiceWithData(t, samplePermissions())

	t.Run("returns all permissions for a known role", func(t *testing.T) {
		perms := svc.GetPermissions("admin")

		assert.NotNil(t, perms)
		assert.Len(t, perms, 3)
		assert.ElementsMatch(t, []string{"create_product", "delete_product", "view_orders"}, perms)
	})

	t.Run("returns nil for unknown role", func(t *testing.T) {
		perms := svc.GetPermissions("nonexistent")
		assert.Nil(t, perms)
	})
}

// ---------------------------------------------------------------------------
// loadPermissions
// ---------------------------------------------------------------------------

func TestPermissionService_loadPermissions(t *testing.T) {
	t.Run("loads data and stores in atomic value", func(t *testing.T) {
		// Arrange — start with empty data, then load real data
		repo := mocks.NewPermissionRepository(t)

		// First call during constructor — return empty
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(map[string][]string{}, nil).
			Once()

		svc := NewPermissionService(repo)
		assert.Nil(t, svc.GetPermissions("admin")) // empty initially

		// Second call — explicit loadPermissions with real data
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(samplePermissions(), nil).
			Once()

		svc.loadPermissions()

		// Assert — new data is available
		assert.True(t, svc.HasPermission("admin", "delete_product"))
		assert.ElementsMatch(t, []string{"view_orders"}, svc.GetPermissions("viewer"))
	})

	t.Run("on error keeps previous state", func(t *testing.T) {
		// Arrange — load real data first
		repo := mocks.NewPermissionRepository(t)

		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(samplePermissions(), nil).
			Once()

		svc := NewPermissionService(repo)
		assert.True(t, svc.HasPermission("admin", "create_product"))

		// Second call fails
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(nil, errors.New("temporary db error")).
			Once()

		svc.loadPermissions()

		// Assert — previous data is still intact
		assert.True(t, svc.HasPermission("admin", "create_product"))
		assert.True(t, svc.HasPermission("seller", "view_orders"))
	})
}

// ---------------------------------------------------------------------------
// StartRefreshLoop
// ---------------------------------------------------------------------------

func TestPermissionService_StartRefreshLoop(t *testing.T) {
	t.Run("context cancellation stops the loop", func(t *testing.T) {
		// Arrange
		repo := mocks.NewPermissionRepository(t)

		// Constructor call
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(samplePermissions(), nil).
			Once()

		svc := NewPermissionService(repo)

		// Refresh calls inside the loop — allow any number
		repo.EXPECT().
			GetAllRolePermissions(mock.Anything).
			Return(samplePermissions(), nil).
			Maybe()

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			svc.StartRefreshLoop(ctx, 10*time.Millisecond)
			close(done)
		}()

		// Let at least one tick happen
		time.Sleep(30 * time.Millisecond)

		// Act — cancel context
		cancel()

		// Assert — goroutine should exit promptly
		select {
		case <-done:
			// success
		case <-time.After(1 * time.Second):
			t.Fatal("StartRefreshLoop did not stop after context cancellation")
		}
	})
}

// ---------------------------------------------------------------------------
// Concurrent access safety
// ---------------------------------------------------------------------------

func TestPermissionService_ConcurrentAccess(t *testing.T) {
	t.Run("HasPermission is safe for concurrent reads", func(t *testing.T) {
		svc := newServiceWithData(t, samplePermissions())

		const goroutines = 50
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				// Mix of hits and misses
				svc.HasPermission("admin", "create_product")
				svc.HasPermission("viewer", "delete_product")
				svc.HasPermission("unknown", "view_orders")
				svc.GetPermissions("seller")
				svc.GetPermissions("nonexistent")
			}()
		}

		wg.Wait() // no race / panic = pass
	})
}
