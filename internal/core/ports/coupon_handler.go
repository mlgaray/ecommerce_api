package ports

import "net/http"

// CouponHandler handles HTTP requests for coupon operations.
type CouponHandler interface {
	// Create handles POST /shops/{shop_id}/coupons requests.
	Create(http.ResponseWriter, *http.Request)

	// Update handles PUT /shops/{shop_id}/coupons/{coupon_id} requests.
	Update(http.ResponseWriter, *http.Request)

	// Delete handles DELETE /shops/{shop_id}/coupons/{coupon_id} requests.
	Delete(http.ResponseWriter, *http.Request)

	// GetByID handles GET /shops/{shop_id}/coupons/{coupon_id} requests.
	GetByID(http.ResponseWriter, *http.Request)

	// GetAll handles GET /shops/{shop_id}/coupons requests.
	GetAll(http.ResponseWriter, *http.Request)
}
