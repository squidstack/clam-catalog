package catalog

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Store handles database operations for products
type Store struct {
	db *sql.DB
}

// NewStore creates a new product store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ListProducts retrieves products with pagination and optional category filter
func (s *Store) ListProducts(ctx context.Context, limit, offset int, category string) ([]Product, error) {
	query := `
		SELECT id, name, description, price, primary_image_url,
		       COALESCE(images, '{}'::text[]) as images,
		       category, sku, stock_count,
		       COALESCE(tags, '{}'::text[]) as tags,
		       rating, review_count, created_at, updated_at
		FROM catalog.products
	`
	args := []interface{}{}
	argCount := 0

	if category != "" {
		argCount++
		query += fmt.Sprintf(" WHERE category = $%d", argCount)
		args = append(args, category)
	}

	query += " ORDER BY created_at DESC"
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListProducts query: %w", err)
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var images, tags pq.StringArray
		var rating sql.NullFloat64

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.PrimaryImageURL,
			&images, &p.Category, &p.SKU, &p.StockCount, &tags,
			&rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ListProducts scan: %w", err)
		}

		p.Images = []string(images)
		p.Tags = []string(tags)
		if rating.Valid {
			p.Rating = &rating.Float64
		}

		products = append(products, p)
	}

	return products, rows.Err()
}

// GetProduct retrieves a single product by ID
func (s *Store) GetProduct(ctx context.Context, id string) (*Product, error) {
	query := `
		SELECT id, name, description, price, primary_image_url,
		       COALESCE(images, '{}'::text[]) as images,
		       category, sku, stock_count,
		       COALESCE(tags, '{}'::text[]) as tags,
		       rating, review_count, created_at, updated_at
		FROM catalog.products
		WHERE id = $1
	`

	var p Product
	var images, tags pq.StringArray
	var rating sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.PrimaryImageURL,
		&images, &p.Category, &p.SKU, &p.StockCount, &tags,
		&rating, &p.ReviewCount, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetProduct query: %w", err)
	}

	p.Images = []string(images)
	p.Tags = []string(tags)
	if rating.Valid {
		p.Rating = &rating.Float64
	}

	return &p, nil
}

// CreateProduct creates a new product
func (s *Store) CreateProduct(ctx context.Context, req CreateProductRequest) (*Product, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO catalog.products (
			id, name, description, price, primary_image_url, images,
			category, sku, stock_count, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, name, description, price, primary_image_url, images,
		          category, sku, stock_count, tags, rating, review_count,
		          created_at, updated_at
	`

	var p Product
	var images, tags pq.StringArray
	var rating sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query,
		id, req.Name, req.Description, req.Price, req.PrimaryImageURL,
		pq.Array(req.Images), req.Category, req.SKU, req.StockCount, pq.Array(req.Tags),
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.PrimaryImageURL, &images,
		&p.Category, &p.SKU, &p.StockCount, &tags, &rating, &p.ReviewCount,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("CreateProduct: %w", err)
	}

	p.Images = []string(images)
	p.Tags = []string(tags)
	if rating.Valid {
		p.Rating = &rating.Float64
	}

	return &p, nil
}

// UpdateProduct updates an existing product
func (s *Store) UpdateProduct(ctx context.Context, id string, req UpdateProductRequest) (*Product, error) {
	// Build dynamic UPDATE query based on provided fields
	query := "UPDATE catalog.products SET updated_at = now()"
	args := []interface{}{}
	argCount := 0

	if req.Name != nil {
		argCount++
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		argCount++
		query += fmt.Sprintf(", description = $%d", argCount)
		args = append(args, *req.Description)
	}
	if req.Price != nil {
		argCount++
		query += fmt.Sprintf(", price = $%d", argCount)
		args = append(args, *req.Price)
	}
	if req.PrimaryImageURL != nil {
		argCount++
		query += fmt.Sprintf(", primary_image_url = $%d", argCount)
		args = append(args, *req.PrimaryImageURL)
	}
	if req.Images != nil {
		argCount++
		query += fmt.Sprintf(", images = $%d", argCount)
		args = append(args, pq.Array(*req.Images))
	}
	if req.Category != nil {
		argCount++
		query += fmt.Sprintf(", category = $%d", argCount)
		args = append(args, *req.Category)
	}
	if req.SKU != nil {
		argCount++
		query += fmt.Sprintf(", sku = $%d", argCount)
		args = append(args, *req.SKU)
	}
	if req.StockCount != nil {
		argCount++
		query += fmt.Sprintf(", stock_count = $%d", argCount)
		args = append(args, *req.StockCount)
	}
	if req.Tags != nil {
		argCount++
		query += fmt.Sprintf(", tags = $%d", argCount)
		args = append(args, pq.Array(*req.Tags))
	}

	argCount++
	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	query += ` RETURNING id, name, description, price, primary_image_url, images,
	                     category, sku, stock_count, tags, rating, review_count,
	                     created_at, updated_at`

	var p Product
	var images, tags pq.StringArray
	var rating sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.PrimaryImageURL, &images,
		&p.Category, &p.SKU, &p.StockCount, &tags, &rating, &p.ReviewCount,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateProduct: %w", err)
	}

	p.Images = []string(images)
	p.Tags = []string(tags)
	if rating.Valid {
		p.Rating = &rating.Float64
	}

	return &p, nil
}

// DeleteProduct deletes a product by ID
func (s *Store) DeleteProduct(ctx context.Context, id string) error {
	query := "DELETE FROM catalog.products WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteProduct: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteProduct rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// CountProducts returns the total count of products (with optional category filter)
func (s *Store) CountProducts(ctx context.Context, category string) (int, error) {
	query := "SELECT COUNT(*) FROM catalog.products"
	args := []interface{}{}

	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountProducts: %w", err)
	}

	return count, nil
}
