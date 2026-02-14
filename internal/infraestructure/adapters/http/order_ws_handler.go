package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	ws "github.com/gorilla/websocket"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	websocketHub "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/websocket"
)

// OrderWSHandler log field constants
const (
	OrderWSHandlerField    = "order_ws_handler"
	WSConnectFunctionField = "connect"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to allowed origins in production
		return true
	},
}

// OrderWSHandler handles WebSocket upgrade requests for order notifications.
type OrderWSHandler struct {
	hub          *websocketHub.Hub
	tokenService ports.TokenService
}

// NewOrderWSHandler creates a new OrderWSHandler.
func NewOrderWSHandler(
	hub *websocketHub.Hub,
	tokenService ports.TokenService,
) *OrderWSHandler {
	return &OrderWSHandler{
		hub:          hub,
		tokenService: tokenService,
	}
}

// Connect handles GET /shops/{shop_id}/orders/ws?token=<jwt>.
// Authenticates via JWT token in query parameter, then upgrades to WebSocket.
func (h *OrderWSHandler) Connect(w http.ResponseWriter, r *http.Request) {
	// 1. Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		logs.WithFields(map[string]interface{}{
			"file":     OrderWSHandlerField,
			"function": WSConnectFunctionField,
			"path":     r.URL.Path,
		}).Warn("Missing token query parameter")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "token_required"})
		return
	}

	// 2. Validate and parse JWT
	claims, err := h.tokenService.ValidateAndParseClaims(token)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderWSHandlerField,
			"function": WSConnectFunctionField,
			"error":    err.Error(),
		}).Warn("Invalid token for WebSocket connection")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "invalid_token"})
		return
	}

	// 3. Parse shop_id from URL path
	shopID, err := parseWSShopID(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderWSHandlerField,
			"function": WSConnectFunctionField,
			"error":    err.Error(),
		}).Warn("Invalid shop_id parameter")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"})
		return
	}

	// 4. Verify the user owns this shop
	if !containsShopID(claims.ShopIDs, shopID) {
		logs.WithFields(map[string]interface{}{
			"file":     OrderWSHandlerField,
			"function": WSConnectFunctionField,
			"user_id":  claims.UserID,
			"shop_id":  shopID,
		}).Warn("User does not own shop")
		httpErrors.HandleError(w, &httpErrors.ForbiddenError{Message: "shop_access_denied"})
		return
	}

	// 5. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderWSHandlerField,
			"function": WSConnectFunctionField,
			"error":    err.Error(),
		}).Error("WebSocket upgrade failed")
		return
	}

	// 6. Register connection
	h.hub.Register(shopID, conn)

	logs.WithFields(map[string]interface{}{
		"file":     OrderWSHandlerField,
		"function": WSConnectFunctionField,
		"user_id":  claims.UserID,
		"shop_id":  shopID,
	}).Info("WebSocket connection established")

	// 7. Start write pump (ping heartbeat) and read pump (detect disconnection)
	go h.writePump(conn)
	h.readPump(shopID, conn)
}

// readPump reads from the WebSocket to detect disconnection.
// Sets pong handler to reset read deadline on heartbeat response.
func (h *OrderWSHandler) readPump(shopID int, conn *ws.Conn) {
	defer func() {
		h.hub.Unregister(shopID, conn)
		conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(websocketHub.PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(websocketHub.PongWait))
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseNormalClosure) {
				logs.WithFields(map[string]interface{}{
					"file":     OrderWSHandlerField,
					"function": "read_pump",
					"shop_id":  shopID,
					"error":    err.Error(),
				}).Warn("WebSocket unexpected close")
			}
			break
		}
	}
}

// writePump sends periodic pings to keep the connection alive.
func (h *OrderWSHandler) writePump(conn *ws.Conn) {
	ticker := time.NewTicker(websocketHub.PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		_ = conn.SetWriteDeadline(time.Now().Add(websocketHub.WriteWait))
		if err := conn.WriteMessage(ws.PingMessage, nil); err != nil {
			return
		}
	}
}

// parseWSShopID extracts and validates shop_id from Gorilla Mux URL vars.
func parseWSShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		return 0, fmt.Errorf("invalid shop_id: %s", shopIDStr)
	}

	return shopID, nil
}

// containsShopID checks if a shopID exists in the user's shop IDs.
func containsShopID(shopIDs []int, shopID int) bool {
	for _, id := range shopIDs {
		if id == shopID {
			return true
		}
	}
	return false
}
