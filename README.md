# 🐚 clam-catalog

Product catalog service for SquidStack.
**Status**: Implemented with full CRUD operations and JWT-based authorization.
**Images**: Product images served via codlocker-assets with SVG support.
**Database**: PostgreSQL with Liquibase migrations and CSV seed data.

---

## ✨ Purpose

- Provide product listings and metadata with pagination and filtering
- Support category browsing and product search
- Serve public product data to **squid-ui** without authentication
- Protect admin operations (create/update/delete) with JWT validation
- Integrate with **codlocker-assets** for product images
- **Manage stock levels** directly in the products table (`stock_count` field)
- Store product ratings and review counts (future integration with **barnacle-reviews** for detailed reviews)

**Note**: Stock management is consolidated in clam-catalog. The original plan included a separate `nautilus-inventory` service, but the architecture decision was made to keep stock levels directly in the catalog for simplicity. The `nautilus-inventory` service remains a stub and is not currently used.

---

## 📡 API Endpoints

### Public Endpoints (No Authentication Required)

| Method | Path                  | Description                          | Query Parameters                               |
|--------|-----------------------|--------------------------------------|------------------------------------------------|
| GET    | `/health`             | Liveness probe                       | None                                           |
| GET    | `/ready`              | Readiness probe (checks DB)          | None                                           |
| GET    | `/_flags`             | Current feature flag values          | None                                           |
| GET    | `/api/products`       | List products with pagination        | `limit` (1-100, default 20)<br>`offset` (default 0)<br>`category` (optional filter) |
| GET    | `/api/products/{id}`  | Get single product by ID             | None                                           |

### Protected Endpoints (Require JWT with admin role)

| Method | Path                  | Description                          | Auth Header                   |
|--------|-----------------------|--------------------------------------|-------------------------------|
| POST   | `/api/products`       | Create new product                   | `Authorization: Bearer <jwt>` |
| PUT    | `/api/products/{id}`  | Update existing product              | `Authorization: Bearer <jwt>` |
| DELETE | `/api/products/{id}`  | Delete product                       | `Authorization: Bearer <jwt>` |

---

## 🔐 Authentication

clam-catalog validates JWT tokens but **does not generate them**. Token generation is handled by **kraken-auth**.

### JWT Requirements
- **Algorithm**: HMAC (HS256/HS384/HS512)
- **Secret**: Shared `JWT_SECRET` environment variable (must match kraken-auth)
- **Claims**: Must include `roles` array with `"admin"` for protected endpoints

### Authorization Flow
1. User authenticates with **kraken-auth** and receives JWT
2. User sends JWT in `Authorization: Bearer <token>` header to clam-catalog
3. clam-catalog validates token signature and checks for `admin` role
4. Public read endpoints (GET /api/products*) require no authentication

---

## 📋 API Examples

### List Products
```bash
# Get first 20 products
curl http://localhost:8080/api/products

# Get 10 products with offset
curl http://localhost:8080/api/products?limit=10&offset=20

# Filter by category
curl http://localhost:8080/api/products?category=Electronics
```

**Response:**
```json
{
  "products": [
    {
      "id": "uuid",
      "name": "Premium Wireless Headphones",
      "description": "High-quality over-ear headphones...",
      "price": 199.99,
      "primary_image_url": "http://codlocker-assets:8080/assets/products/electronics/product-001.jpg",
      "images": [
        "http://codlocker-assets:8080/assets/products/electronics/product-001.jpg",
        "http://codlocker-assets:8080/assets/products/electronics/product-001-alt1.jpg"
      ],
      "category": "Electronics",
      "sku": "ELEC-HDPHN-001",
      "stock_count": 50,
      "tags": ["electronics", "audio", "wireless", "premium"],
      "rating": 4.5,
      "review_count": 128,
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-15T10:00:00Z"
    }
  ],
  "total": 55,
  "limit": 20,
  "offset": 0
}
```

### Get Single Product
```bash
curl http://localhost:8080/api/products/{id}
```

### Create Product (Admin Only)
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Product",
    "description": "Product description",
    "price": 49.99,
    "primary_image_url": "http://codlocker-assets:8080/assets/products/new-product.jpg",
    "images": ["http://codlocker-assets:8080/assets/products/new-product.jpg"],
    "category": "Electronics",
    "sku": "ELEC-NEW-001",
    "stock_count": 100,
    "tags": ["electronics", "new"]
  }'
```

### Update Product (Admin Only)
```bash
curl -X PUT http://localhost:8080/api/products/{id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "price": 39.99,
    "stock_count": 75
  }'
```

### Delete Product (Admin Only)
```bash
curl -X DELETE http://localhost:8080/api/products/{id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🗄 Database Schema

### Schema: `catalog`

### Table: `catalog.products`

| Column              | Type         | Constraints       | Description                                    |
|---------------------|--------------|-------------------|------------------------------------------------|
| id                  | uuid         | PRIMARY KEY       | Unique product identifier                      |
| name                | varchar(255) | NOT NULL          | Product name                                   |
| description         | text         |                   | Product description                            |
| price               | numeric(10,2)| NOT NULL          | Product price (USD)                            |
| primary_image_url   | varchar(500) |                   | Main product image URL                         |
| images              | text[]       |                   | Array of image URLs (primary + alternates)     |
| category            | varchar(100) |                   | Product category (Electronics, Clothing, etc.) |
| sku                 | varchar(100) | UNIQUE            | Stock Keeping Unit                             |
| stock_count         | integer      | DEFAULT 0         | Current stock level                            |
| tags                | text[]       |                   | Array of tags for search/filtering             |
| rating              | numeric(3,2) |                   | Average rating (0-5.00)                        |
| review_count        | integer      | DEFAULT 0         | Number of reviews                              |
| created_at          | timestamptz  | DEFAULT now()     | Creation timestamp                             |
| updated_at          | timestamptz  | DEFAULT now()     | Last update timestamp                          |

**Indexes:**
- `idx_catalog_products_category` on `category`
- `idx_catalog_products_tags` (GIN index) on `tags` for efficient array searches

**Seed Data:**
- 55 products seeded via Liquibase from CSV
- Categories: Electronics (15), Clothing (15), Home & Kitchen (10), Sports & Outdoors (10), Books & Media (5)

---

## 🚀 Database Migrations

All database changes are managed via **Liquibase** in `chart/clam-catalog/liquibase/`:

| File                             | Description                                    |
|----------------------------------|------------------------------------------------|
| `changelog-root.xml`             | Root changelog orchestrating all migrations    |
| `0001-create-catalog-schema.xml` | Creates `catalog` schema                       |
| `0002-create-products-table.xml` | Creates `products` table with indexes          |
| `0003-seed-products-from-csv.xml`| Loads 55 products from CSV seed data           |
| `data/products.csv`              | Product seed data (55 products)                |

**Running Migrations:**
Liquibase runs automatically during deployment. For local testing:
```bash
liquibase --changeLogFile=chart/clam-catalog/liquibase/changelog-root.xml update
```

---

## 🔧 Environment Variables

| Variable          | Required | Default | Description                                     |
|-------------------|----------|---------|-------------------------------------------------|
| `DB_HOST`         | Yes      | —       | PostgreSQL host                                 |
| `DB_PORT`         | Yes      | 5432    | PostgreSQL port                                 |
| `DB_NAME`         | Yes      | —       | Database name                                   |
| `DB_USER`         | Yes      | —       | Database user                                   |
| `DB_PASSWORD`     | Yes      | —       | Database password                               |
| `JWT_SECRET`      | Yes      | —       | JWT signing secret (must match kraken-auth)     |
| `FEATURE_FLAGS_KEY`| No      | —       | LaunchDarkly SDK key (optional)                 |

---

## 🔗 Dependencies

**Current dependencies:**
- **PostgreSQL**: Database for product catalog and stock management
- **kraken-auth**: JWT token generation (clam-catalog only validates tokens)
- **codlocker-assets**: Product image serving (all product images are served from codlocker-assets)

**Future planned integrations:**
- **barnacle-reviews**: Will provide detailed review data to supplement the rating/review_count fields
- **cuttlefish-orders**: Will update stock_count when orders are placed

**Not used:**
- **nautilus-inventory**: Originally planned for stock management, but stock is now managed directly in clam-catalog

---

## 🧪 Local Testing

### Prerequisites
- PostgreSQL running on `localhost:5432`
- Database created: `clam_catalog_db`
- Liquibase migrations applied
- `JWT_SECRET` environment variable set (same as kraken-auth)

### Run Service
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=clam_catalog_db
export DB_USER=postgres
export DB_PASSWORD=yourpassword
export JWT_SECRET=your-shared-secret

go run main.go
```

### Test Endpoints
```bash
# Health check
curl http://localhost:8080/health

# List products
curl http://localhost:8080/api/products

# Filter by category
curl http://localhost:8080/api/products?category=Electronics

# Get single product
curl http://localhost:8080/api/products/{uuid}

# Create product (requires JWT from kraken-auth)
curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","price":29.99,"sku":"TEST-001"}'
```

### Verify Database
```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d clam_catalog_db

# Check products seeded
SELECT category, COUNT(*) FROM catalog.products GROUP BY category;

# Test array queries
SELECT name FROM catalog.products WHERE tags @> ARRAY['electronics'];

# Test category filter
SELECT name, category FROM catalog.products WHERE category = 'Electronics';
```

---

## 📦 Docker Build

```bash
docker build -t clam-catalog:latest .
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_NAME=clam_catalog_db \
  -e DB_USER=postgres \
  -e DB_PASSWORD=yourpassword \
  -e JWT_SECRET=your-shared-secret \
  clam-catalog:latest
```

<!-- Build trigger: 2025-12-11 -->


# Testing Smart Tests integration

<!-- Test trigger: Thu 22 Jan 2026 10:46:21 GMT -->

Last updated: 2026-02-11 21:51 UTC
