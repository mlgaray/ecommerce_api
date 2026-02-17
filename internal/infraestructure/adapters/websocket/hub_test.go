package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

func TestMain(m *testing.M) {
	logs.Init()
	os.Exit(m.Run())
}

// =============================================================================
// Test Helpers
// =============================================================================

// createTestWSConnection creates a test WebSocket server and returns a client connection.
func createTestWSConnection(t *testing.T) (*ws.Conn, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := ws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("failed to upgrade: %v", err)
		}
		// Keep the connection alive until the server closes
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		server.Close()
	}
	return conn, cleanup
}

func newTestOrder() *models.Order {
	return &models.Order{
		ID:          1,
		OrderNumber: "ORD-2024-00001",
		Status:      models.OrderStatusPending,
		Store: &models.Store{
			ID:   1,
			Name: "Test Store",
			Slug: "test-store",
		},
		Customer: &models.Customer{
			Name: "John Doe",
		},
		Items:    []*models.OrderItem{},
		Subtotal: 100.00,
		Total:    110.00,
	}
}

// =============================================================================
// Register Tests
// =============================================================================

func TestHub_Register(t *testing.T) {
	t.Run("when registering connection then adds to correct shop", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn, cleanup := createTestWSConnection(t)
		defer cleanup()

		// Act
		hub.Register(1, conn)

		// Assert
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		assert.Len(t, hub.connections[1], 1)
		assert.Equal(t, conn, hub.connections[1][0])
	})

	t.Run("when registering multiple connections then all tracked", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn1, cleanup1 := createTestWSConnection(t)
		defer cleanup1()
		conn2, cleanup2 := createTestWSConnection(t)
		defer cleanup2()

		// Act
		hub.Register(1, conn1)
		hub.Register(1, conn2)

		// Assert
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		assert.Len(t, hub.connections[1], 2)
	})
}

// =============================================================================
// Unregister Tests
// =============================================================================

func TestHub_Unregister(t *testing.T) {
	t.Run("when unregistering connection then removes from shop", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn, cleanup := createTestWSConnection(t)
		defer cleanup()
		hub.Register(1, conn)

		// Act
		hub.Unregister(1, conn)

		// Assert
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		assert.Empty(t, hub.connections[1])
		_, exists := hub.connections[1]
		assert.False(t, exists)
	})

	t.Run("when unregistering from non-existent shop then no panic", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn, cleanup := createTestWSConnection(t)
		defer cleanup()

		// Act & Assert (no panic)
		hub.Unregister(999, conn)
	})

	t.Run("when unregistering one connection then keeps remaining connections", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn1, cleanup1 := createTestWSConnection(t)
		defer cleanup1()
		conn2, cleanup2 := createTestWSConnection(t)
		defer cleanup2()

		hub.Register(1, conn1)
		hub.Register(1, conn2)

		// Act
		hub.Unregister(1, conn1)

		// Assert
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		assert.Len(t, hub.connections[1], 1)
		assert.Equal(t, conn2, hub.connections[1][0])
	})
}

// =============================================================================
// NotifyNewOrder Tests
// =============================================================================

func TestHub_NotifyNewOrder(t *testing.T) {
	t.Run("when order is nil then does nothing", func(t *testing.T) {
		// Arrange
		hub := NewHub()

		// Act & Assert (no panic)
		hub.NotifyNewOrder(context.Background(), nil)
	})

	t.Run("when order store is nil then does nothing", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		order := &models.Order{ID: 1}

		// Act & Assert (no panic)
		hub.NotifyNewOrder(context.Background(), order)
	})

	t.Run("when no connections for shop then logs and returns", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		order := newTestOrder()

		// Act & Assert (no panic, no error)
		hub.NotifyNewOrder(context.Background(), order)
	})

	t.Run("when connection exists then sends notification", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		order := newTestOrder()

		// Create a test WS server that receives the hub's messages
		serverConn := make(chan *ws.Conn, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := ws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			serverConn <- conn
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		clientConn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer clientConn.Close()

		// Get the server-side connection (this is what the hub writes to)
		sConn := <-serverConn
		hub.Register(order.Store.ID, sConn)

		// Act
		hub.NotifyNewOrder(context.Background(), order)

		// Assert - read message from client side
		_, message, err := clientConn.ReadMessage()
		require.NoError(t, err)

		var notification NewOrderNotification
		err = json.Unmarshal(message, &notification)
		require.NoError(t, err)
		assert.Equal(t, "new_order", notification.Event)
		assert.Equal(t, order.ID, notification.Order.ID)
		assert.Equal(t, order.OrderNumber, notification.Order.OrderNumber)

		// Verify the response uses the same format as OrderResponseFromModel
		expected := responses.OrderResponseFromModel(order)
		assert.Equal(t, expected.Total, notification.Order.Total)
	})

	t.Run("when notification sent to shop 1 then shop 2 does not receive", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		order := newTestOrder() // store ID = 1

		// Connection for shop 2
		serverConn2 := make(chan *ws.Conn, 1)
		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := ws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			serverConn2 <- conn
		}))
		defer server2.Close()

		wsURL2 := "ws" + strings.TrimPrefix(server2.URL, "http")
		clientConn2, _, err := ws.DefaultDialer.Dial(wsURL2, nil)
		require.NoError(t, err)
		defer clientConn2.Close()

		sConn2 := <-serverConn2
		hub.Register(2, sConn2) // Register for shop 2

		// Act
		hub.NotifyNewOrder(context.Background(), order) // Notify shop 1

		// Assert - shop 2 should not receive anything
		hub.mu.RLock()
		connsShop1 := hub.connections[1]
		hub.mu.RUnlock()
		assert.Empty(t, connsShop1) // No connections for shop 1
	})

	t.Run("when connection is dead then removes it from hub", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		order := newTestOrder() // store ID = 1

		serverConn := make(chan *ws.Conn, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := ws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			serverConn <- conn
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		clientConn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer clientConn.Close()

		sConn := <-serverConn
		hub.Register(order.Store.ID, sConn)

		// Close the server-side connection so WriteMessage will fail
		sConn.Close()

		// Act
		hub.NotifyNewOrder(context.Background(), order)

		// Assert - dead connection should have been removed
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		_, exists := hub.connections[order.Store.ID]
		assert.False(t, exists, "dead connection should be removed from hub")
	})
}

// =============================================================================
// CloseAll Tests
// =============================================================================

func TestHub_CloseAll(t *testing.T) {
	t.Run("when closing all then connections map is empty", func(t *testing.T) {
		// Arrange
		hub := NewHub()
		conn1, cleanup1 := createTestWSConnection(t)
		defer cleanup1()
		conn2, cleanup2 := createTestWSConnection(t)
		defer cleanup2()

		hub.Register(1, conn1)
		hub.Register(2, conn2)

		// Act
		hub.CloseAll()

		// Assert
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		assert.Empty(t, hub.connections)
	})
}
