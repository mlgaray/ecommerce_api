package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Heartbeat constants for WebSocket connections.
const (
	WriteWait  = 10 * time.Second // Max time to write a message to the peer.
	PongWait   = 60 * time.Second // Max time to wait for a pong from the peer.
	PingPeriod = 54 * time.Second // Ping period (must be < PongWait).
)

// Hub log field constants
const (
	HubFileField = "websocket_hub"
)

// NewOrderNotification is the WebSocket message payload for new orders.
type NewOrderNotification struct {
	Event string                   `json:"event"`
	Order *contracts.OrderResponse `json:"order"`
}

// Hub manages WebSocket connections grouped by shop ID.
// Implements ports.OrderEventNotifier.
// Safe for concurrent use.
type Hub struct {
	mu          sync.RWMutex
	connections map[int][]*ws.Conn
}

// NewHub creates a new WebSocket connection hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int][]*ws.Conn),
	}
}

// Register adds a WebSocket connection for a specific shop.
func (h *Hub) Register(shopID int, conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[shopID] = append(h.connections[shopID], conn)

	logs.WithFields(map[string]interface{}{
		"file":     HubFileField,
		"function": "register",
		"shop_id":  shopID,
		"total":    len(h.connections[shopID]),
	}).Debug("WebSocket client registered")
}

// Unregister removes a WebSocket connection for a specific shop.
func (h *Hub) Unregister(shopID int, conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[shopID]
	for i, c := range conns {
		if c == conn {
			h.connections[shopID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(h.connections[shopID]) == 0 {
		delete(h.connections, shopID)
	}

	logs.WithFields(map[string]interface{}{
		"file":     HubFileField,
		"function": "unregister",
		"shop_id":  shopID,
	}).Debug("WebSocket client unregistered")
}

// NotifyNewOrder implements ports.OrderEventNotifier.
// Sends the order to all connections subscribed to order.Store.ID.
// Fire-and-forget: errors are logged, dead connections are removed.
func (h *Hub) NotifyNewOrder(_ context.Context, order *models.Order) {
	if order == nil || order.Store == nil {
		return
	}

	shopID := order.Store.ID

	payload := NewOrderNotification{
		Event: "new_order",
		Order: contracts.OrderResponseFromModel(order),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     HubFileField,
			"function": "notify_new_order",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Failed to marshal order notification")
		return
	}

	h.mu.RLock()
	conns := make([]*ws.Conn, len(h.connections[shopID]))
	copy(conns, h.connections[shopID])
	h.mu.RUnlock()

	var deadConns []*ws.Conn
	for _, conn := range conns {
		_ = conn.SetWriteDeadline(time.Now().Add(WriteWait))
		if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     HubFileField,
				"function": "notify_new_order",
				"shop_id":  shopID,
				"error":    err.Error(),
			}).Warn("Failed to write to WebSocket, marking for removal")
			deadConns = append(deadConns, conn)
		}
	}

	for _, conn := range deadConns {
		h.Unregister(shopID, conn)
		conn.Close()
	}

	logs.WithFields(map[string]interface{}{
		"file":           HubFileField,
		"function":       "notify_new_order",
		"shop_id":        shopID,
		"order_id":       order.ID,
		"order_number":   order.OrderNumber,
		"notified_count": len(conns) - len(deadConns),
	}).Info("Order notification sent")
}

// CloseAll closes all WebSocket connections. Called on server shutdown.
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for shopID, conns := range h.connections {
		for _, conn := range conns {
			_ = conn.WriteMessage(ws.CloseMessage,
				ws.FormatCloseMessage(ws.CloseGoingAway, "server_shutdown"))
			conn.Close()
		}
		delete(h.connections, shopID)
	}

	logs.WithFields(map[string]interface{}{
		"file":     HubFileField,
		"function": "close_all",
	}).Info("All WebSocket connections closed")
}
