package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"

	"clam-catalog/internal/auth"
	"clam-catalog/internal/logger"

	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for catalog operations
type Handler struct {
	store *Store
}

// NewHandler creates a new catalog handler
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// ListProducts handles GET /api/products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	category := r.URL.Query().Get("category")

	products, err := h.store.ListProducts(r.Context(), limit, offset, category)
	if err != nil {
		logger.Errorf("ListProducts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	total, err := h.store.CountProducts(r.Context(), category)
	if err != nil {
		logger.Errorf("CountProducts: %v", err)
		total = len(products) // Fallback to current count
	}

	resp := ProductListResponse{
		Products: products,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetProduct handles GET /api/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	product, err := h.store.GetProduct(r.Context(), id)
	if err != nil {
		logger.Errorf("GetProduct: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreateProduct handles POST /api/products (admin only)
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Price <= 0 || req.SKU == "" {
		http.Error(w, "name, price, and sku are required", http.StatusBadRequest)
		return
	}

	product, err := h.store.CreateProduct(r.Context(), req)
	if err != nil {
		logger.Errorf("CreateProduct: %v", err)
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// UpdateProduct handles PUT /api/products/{id} (admin only)
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.store.UpdateProduct(r.Context(), id, req)
	if err != nil {
		logger.Errorf("UpdateProduct: %v", err)
		http.Error(w, "failed to update product", http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// DeleteProduct handles DELETE /api/products/{id} (admin only)
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	err := h.store.DeleteProduct(r.Context(), id)
	if err != nil {
		logger.Errorf("DeleteProduct: %v", err)
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RequireAdmin is middleware that requires a valid JWT token with admin role
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := auth.GetBearerToken(r)
		if tokenStr == "" {
			logger.Debugf("RequireAdmin: no bearer token provided")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			logger.Debugf("RequireAdmin: JWT parse error: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !auth.HasRole(claims.Roles, "admin") {
			logger.Debugf("RequireAdmin: user lacks admin role")
			http.Error(w, "forbidden - admin role required", http.StatusForbidden)
			return
		}

		// Token is valid and has admin role, proceed
		next(w, r)
	}
}
