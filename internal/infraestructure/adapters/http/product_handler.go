package http

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type ProductHandler struct {
	createProduct  ports.CreateProductUseCase
	getAllByShopID ports.GetAllByShopIDUseCase
}

func (p *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form (13MB limit - allows 4 images of 3MB each + product data)
	err := r.ParseMultipartForm(13 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "parse_multipart_form",
			"error":     err.Error(),
		}).Error("Error parsing multipart form")
		errors.HandleError(w, &errors.BadRequestError{Message: "Error parsing multipart form"})
		return
	}

	// Create ProductCreateRequest
	request, err := p.buildProductCreateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "build_request",
			"error":     err.Error(),
		}).Error("Error building product create request")
		errors.HandleError(w, err)
		return
	}

	// Validate request (includes product data and images)
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"operation":    "validate_request",
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Product creation validation failed")
		errors.HandleError(w, err)
		return
	}

	// Convert images to buffers for upload service
	imageBuffers, err := request.ToImageBuffers()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "convert_images_to_buffers",
			"error":     err.Error(),
		}).Error("Error converting images to buffers")
		errors.HandleError(w, &errors.InternalServiceError{Message: err.Error()})
		return
	}

	// Create product via use case
	createdProduct, err := p.createProduct.Execute(ctx, &request.Product, imageBuffers, request.ShopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation":    "create_product",
			"product_name": request.Product.Name,
			"shop_id":      request.ShopID,
			"error":        err.Error(),
		}).Error("Error creating product")
		errors.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdProduct); err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "encode_response",
			"error":     err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) buildProductCreateRequest(r *http.Request) (*contracts.ProductCreateRequest, error) {
	// Extract product JSON from form data
	productJSON := r.FormValue("product")
	if strings.TrimSpace(productJSON) == "" {
		return nil, &errors.BadRequestError{Message: "product JSON is required"}
	}

	// Parse product JSON
	var product models.Product
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return nil, &errors.BadRequestError{Message: "invalid product JSON format: " + err.Error()}
	}

	// Get shop ID from form
	shopIDStr := r.FormValue("shop_id")
	if strings.TrimSpace(shopIDStr) == "" {
		return nil, &errors.BadRequestError{Message: "shop_id is required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return nil, &errors.BadRequestError{Message: "invalid shop_id format"}
	}

	// Get images from form
	var images []*multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for i := 0; ; i++ {
			key := "images[" + strconv.Itoa(i) + "]"
			if files, exists := r.MultipartForm.File[key]; exists && len(files) > 0 {
				images = append(images, files[0])
			} else {
				break
			}
		}
	}

	return &contracts.ProductCreateRequest{
		Product: product,
		ShopID:  shopID,
		Images:  images,
	}, nil
}

func NewProductHandler(createProductUseCase ports.CreateProductUseCase, getAllUseCase ports.GetAllByShopIDUseCase) *ProductHandler {
	return &ProductHandler{
		createProduct:  createProductUseCase,
		getAllByShopID: getAllUseCase,
	}
}

func (p *ProductHandler) GetAllByShopID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate shop_id
	shopID, err := p.parseShopID(r)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	// Parse and validate pagination parameters
	limit, cursor, err := p.parsePaginationParams(r)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	// Execute use case
	products, nextCursor, hasMore, err := p.getAllByShopID.Execute(ctx, shopID, limit, cursor)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "get_all_products_by_shop",
			"shop_id":   shopID,
			"limit":     limit,
			"cursor":    cursor,
			"error":     err.Error(),
		}).Error("Error retrieving products")
		errors.HandleError(w, err)
		return
	}

	// Build HTTP response
	response := contracts.PaginatedProductsResponse{
		Products:   products,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "encode_response",
			"error":     err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]
	if strings.TrimSpace(shopIDStr) == "" {
		logs.WithFields(map[string]interface{}{
			"operation": "get_all_products_by_shop",
			"error":     "shop_id parameter is required",
		}).Error("Missing shop_id parameter")
		return 0, &errors.BadRequestError{Message: "shop_id parameter is required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "get_all_products_by_shop",
			"shop_id":   shopIDStr,
			"error":     err.Error(),
		}).Error("Invalid shop_id parameter")
		return 0, &errors.BadRequestError{Message: "invalid shop_id format"}
	}

	return shopID, nil
}

func (p *ProductHandler) parsePaginationParams(r *http.Request) (int, int, error) {
	limitStr := r.URL.Query().Get("limit")
	cursorStr := r.URL.Query().Get("cursor")

	limit := 20 // default
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			logs.WithFields(map[string]interface{}{
				"operation": "parse_limit",
				"limit":     limitStr,
				"error":     err,
			}).Error("Invalid limit parameter")
			return 0, 0, &errors.BadRequestError{Message: "invalid limit format"}
		}
		limit = parsedLimit
	}

	cursor := 0 // default (first page)
	if cursorStr != "" {
		parsedCursor, err := strconv.Atoi(cursorStr)
		if err != nil || parsedCursor < 0 {
			logs.WithFields(map[string]interface{}{
				"operation": "parse_cursor",
				"cursor":    cursorStr,
				"error":     err,
			}).Error("Invalid cursor parameter")
			return 0, 0, &errors.BadRequestError{Message: "invalid cursor format"}
		}
		cursor = parsedCursor
	}

	return limit, cursor, nil
}
