package http

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Product handler log field constants
const (
	ProductHandlerField           = "product_handler"
	GetAllByShopIDFunctionField   = "get_all_by_shop_id"
	GetByIDFunctionField          = "get_by_id"
	CreateProductFunctionField    = "create"
	UpdateProductFunctionField    = "update"
	ParseShopIDSubFuncField       = "parse_shop_id"
	ParseProductIDSubFuncField    = "parse_product_id"
	ParsePaginationSubFuncField   = "parse_pagination_params"
	BuildRequestSubFuncField      = "build_request"
	ConvertImagesToBuffersSubFunc = "convert_images_to_buffers"
)

type ProductHandler struct {
	createProduct  ports.CreateProductUseCase
	getAllByShopID ports.GetAllByShopIDUseCase
	getByID        ports.GetByIDUseCase
	updateProduct  ports.UpdateProductUseCase
}

func (p *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Parse multipart form (13MB limit - allows 4 images of 3MB each + product data)
	stepStart := time.Now()
	err := r.ParseMultipartForm(13 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "parse_multipart_form",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 1: Multipart form parsed")

	// Create ProductCreateRequest
	stepStart = time.Now()
	request, err := p.buildProductCreateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": BuildRequestSubFuncField,
			"error":    err.Error(),
		}).Error("Error building product create request")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "build_request",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 2: Request built")

	// Validate request (includes product data and images)
	stepStart = time.Now()
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     CreateProductFunctionField,
			"sub_func":     "request.Validate",
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Product creation validation failed")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "validate_request",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 3: Request validated")

	// Convert images to buffers for upload service
	stepStart = time.Now()
	imageBuffers, err := request.ToImageBuffers()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": ConvertImagesToBuffersSubFunc,
			"error":    err.Error(),
		}).Error("Error converting images to buffers")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "convert_images_to_buffers",
		"image_count": len(imageBuffers),
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 4: Images converted to buffers")

	// Create product via use case
	stepStart = time.Now()
	createdProduct, err := p.createProduct.Execute(ctx, &request.Product, imageBuffers, request.ShopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     CreateProductFunctionField,
			"product_name": request.Product.Name,
			"shop_id":      request.ShopID,
			"error":        err.Error(),
		}).Error("Error creating product")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "execute_use_case",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 5: Use case executed")

	logs.WithFields(map[string]interface{}{
		"operation":         "create_product_total",
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Product creation completed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdProduct); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) buildProductCreateRequest(r *http.Request) (*contracts.ProductCreateRequest, error) {
	// Extract product JSON from form data
	productJSON := r.FormValue("product")
	if strings.TrimSpace(productJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "product_json_required"}
	}

	// Parse product JSON
	var product models.Product
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_product_json_format"}
	}

	// Get shop ID from form
	shopIDStr := r.FormValue("shop_id")
	if strings.TrimSpace(shopIDStr) == "" {
		return nil, &httpErrors.BadRequestError{Message: "shop_id_required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
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

func NewProductHandler(createProductUseCase ports.CreateProductUseCase, getAllUseCase ports.GetAllByShopIDUseCase, getByIDUseCase ports.GetByIDUseCase, updateProductUseCase ports.UpdateProductUseCase) *ProductHandler {
	return &ProductHandler{
		createProduct:  createProductUseCase,
		getAllByShopID: getAllUseCase,
		getByID:        getByIDUseCase,
		updateProduct:  updateProductUseCase,
	}
}

func (p *ProductHandler) GetAllByShopID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate shop_id
	shopID, err := p.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse and validate pagination parameters
	limit, cursor, err := p.parsePaginationParams(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	products, nextCursor, hasMore, err := p.getAllByShopID.Execute(ctx, shopID, limit, cursor)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": GetAllByShopIDFunctionField,
			"shop_id":  shopID,
			"limit":    limit,
			"cursor":   cursor,
			"error":    err.Error(),
		}).Error("Error retrieving products")
		httpErrors.HandleError(w, err)
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
			"file":     ProductHandlerField,
			"function": GetAllByShopIDFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]
	if strings.TrimSpace(shopIDStr) == "" {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": ParseShopIDSubFuncField,
			"error":    "shop_id_parameter_required",
		}).Error("Missing shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "shop_id_parameter_required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": ParseShopIDSubFuncField,
			"sub_func": "strconv.Atoi",
			"shop_id":  shopIDStr,
			"error":    err.Error(),
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
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
				"file":     ProductHandlerField,
				"function": ParsePaginationSubFuncField,
				"sub_func": "strconv.Atoi",
				"limit":    limitStr,
				"error":    err,
			}).Error("Invalid limit parameter")
			return 0, 0, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		limit = parsedLimit
	}

	cursor := 0 // default (first page)
	if cursorStr != "" {
		parsedCursor, err := strconv.Atoi(cursorStr)
		if err != nil || parsedCursor < 0 {
			logs.WithFields(map[string]interface{}{
				"file":     ProductHandlerField,
				"function": ParsePaginationSubFuncField,
				"sub_func": "strconv.Atoi",
				"cursor":   cursorStr,
				"error":    err,
			}).Error("Invalid cursor parameter")
			return 0, 0, &httpErrors.BadRequestError{Message: "invalid_cursor_format"}
		}
		cursor = parsedCursor
	}

	return limit, cursor, nil
}

func (p *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate product_id
	productID, err := p.parseProductID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	product, err := p.getByID.Execute(ctx, productID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   GetByIDFunctionField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Error retrieving product")
		httpErrors.HandleError(w, err)
		return
	}

	// Return product directly (no DTO wrapper needed for single product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(product); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": GetByIDFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) parseProductID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	if strings.TrimSpace(productIDStr) == "" {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": ParseProductIDSubFuncField,
			"error":    "product_id_parameter_required",
		}).Error("Missing product_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "product_id_parameter_required"}
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   ParseProductIDSubFuncField,
			"sub_func":   "strconv.Atoi",
			"product_id": productIDStr,
			"error":      err,
		}).Error("Invalid product_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_product_id_format"}
	}

	return productID, nil
}

func (p *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate product_id
	productID, err := p.parseProductID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(13 << 20) // 13MB limit
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	// Build product update request (different from create)
	request, err := p.buildProductUpdateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": BuildRequestSubFuncField,
			"error":    err.Error(),
		}).Error("Error building product update request")
		httpErrors.HandleError(w, err)
		return
	}

	// Set product ID from path param (override any ID in JSON)
	request.Product.ID = productID

	// Validate request
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     UpdateProductFunctionField,
			"sub_func":     "request.Validate",
			"product_id":   productID,
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Product update validation failed")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert new images to buffers for upload service
	imageBuffers, err := request.ToImageBuffers()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": ConvertImagesToBuffersSubFunc,
			"error":    err.Error(),
		}).Error("Error converting images to buffers")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}

	// Update product via use case
	err = p.updateProduct.Execute(ctx, productID, &request.Product, imageBuffers)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     UpdateProductFunctionField,
			"product_id":   productID,
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Error updating product")
		httpErrors.HandleError(w, err)
		return
	}

	// Return success message (no product returned - frontend navigates to list)
	type UpdateSuccessResponse struct {
		Message string `json:"message"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(UpdateSuccessResponse{Message: "product_updated_successfully"}); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) buildProductUpdateRequest(r *http.Request) (*contracts.ProductUpdateRequest, error) {
	// Extract product JSON from form data
	productJSON := r.FormValue("product")
	if strings.TrimSpace(productJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "product_json_required"}
	}

	// Parse product JSON (includes existing images with IDs)
	var product models.Product
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_product_json_format"}
	}

	// Get shop ID from form
	shopIDStr := r.FormValue("shop_id")
	if strings.TrimSpace(shopIDStr) == "" {
		return nil, &httpErrors.BadRequestError{Message: "shop_id_required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	// Get new images from multipart form (optional for update)
	var newImages []*multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for i := 0; ; i++ {
			key := "images[" + strconv.Itoa(i) + "]"
			if files, exists := r.MultipartForm.File[key]; exists && len(files) > 0 {
				newImages = append(newImages, files[0])
			} else {
				break
			}
		}
	}

	return &contracts.ProductUpdateRequest{
		Product:   product,
		ShopID:    shopID,
		NewImages: newImages,
	}, nil
}
