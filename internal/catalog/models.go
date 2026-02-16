package catalog

import "time"

// Product represents a product in the catalog
type Product struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	PrimaryImageURL string    `json:"primary_image_url"`
	Images          []string  `json:"images"`
	Category        string    `json:"category"`
	SKU             string    `json:"sku"`
	StockCount      int       `json:"stock_count"`
	Tags            []string  `json:"tags"`
	Rating          *float64  `json:"rating,omitempty"`
	ReviewCount     int       `json:"review_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProductListResponse wraps a list of products with pagination info
type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// CreateProductRequest represents the payload for creating a product
type CreateProductRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           float64  `json:"price"`
	PrimaryImageURL string   `json:"primary_image_url"`
	Images          []string `json:"images"`
	Category        string   `json:"category"`
	SKU             string   `json:"sku"`
	StockCount      int      `json:"stock_count"`
	Tags            []string `json:"tags"`
}

// UpdateProductRequest represents the payload for updating a product
type UpdateProductRequest struct {
	Name            *string   `json:"name,omitempty"`
	Description     *string   `json:"description,omitempty"`
	Price           *float64  `json:"price,omitempty"`
	PrimaryImageURL *string   `json:"primary_image_url,omitempty"`
	Images          *[]string `json:"images,omitempty"`
	Category        *string   `json:"category,omitempty"`
	SKU             *string   `json:"sku,omitempty"`
	StockCount      *int      `json:"stock_count,omitempty"`
	Tags            *[]string `json:"tags,omitempty"`
}
