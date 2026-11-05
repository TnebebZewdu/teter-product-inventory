# TETER Product Inventory API

A RESTful Product Inventory API built with Go, PostgreSQL, Chi, sqlc, and Docker. It supports product creation, retrieval, updating, deletion, filtering, and pagination.

## Technologies

- Go
- PostgreSQL
- Chi Router
- sqlc
- pgx
- Docker & Docker Compose
- golang-migrate

## Features

- Complete product CRUD operations
- Pagination for product listings
- Product filtering by status
- Product search
- Low-stock filtering
- Input validation
- Unique SKU validation
- Standardized JSON error responses
- Database migrations with up/down support
- Idempotent product seeder (10 sample products)
- Database health-check endpoint
- Dockerized API and PostgreSQL database
- Non-root Docker application user
- Automated handler tests

## Project Structure

```text
.
├── cmd/
│   ├── api/                # API application entry point
│   └── seeder/             # Database seeder
├── db/
│   ├── migrations/         # Database migrations
│   ├── queries/             # SQL queries for sqlc
│   └── seed/                # Sample product data
├── internal/
│   ├── database/            # Database connection and generated sqlc code
│   └── handlers/            # Product API handlers and tests
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── sqlc.yaml
└── go.mod
```

## Getting Started

### Prerequisites

- Go
- Docker Desktop
- Docker Compose
- golang-migrate CLI
- GNU Make (optional)

### 1. Clone the repository

```bash
git clone https://github.com/TnebebZewdu/teter-product-inventory.git
cd teter-product-inventory
```

### 2. Configure environment variables

Create a `.env` file in the project root:

```env
POSTGRES_PASSWORD=your_secure_password
POSTGRES_PASSWORD_URLENCODED=your_url_encoded_password
```

Use the same password for both variables. URL-encode special characters in `POSTGRES_PASSWORD_URLENCODED`.

### 3. Start the containers

```bash
docker compose up --build -d
```

The API will be available at:

`http://localhost:8080`

PostgreSQL is exposed on port `5433` on the host.

### 4. Apply database migrations

Run the following command, replacing the password placeholder with your URL-encoded password:

```bash
migrate -path db/migrations -database "postgres://postgres:<URL_ENCODED_PASSWORD>@localhost:5433/teter_inventory?sslmode=disable" up
```

Two migrations are included:

- Create the products table with validation constraints.
- Add an index on the product status column.

### 5. Seed sample products

Run in PowerShell:

```powershell
Get-Content .\db\seed\products.sql -Raw | docker compose exec -T db psql -U postgres -d teter_inventory
```

The seeder inserts 10 sample products and updates existing records using the unique SKU constraint to avoid duplicates.

Alternatively, use `make seed` after configuring `DATABASE_URL`.

## API Endpoints

Base URL: `http://localhost:8080`

| Method | Endpoint                | Description                   |
| ------ | ----------------------- | ----------------------------- |
| GET    | `/healthz`              | Check API and database health |
| POST   | `/api/v1/products`      | Create a product              |
| GET    | `/api/v1/products`      | List products                 |
| GET    | `/api/v1/products/{id}` | Get a product                 |
| PUT    | `/api/v1/products/{id}` | Update a product              |
| DELETE | `/api/v1/products/{id}` | Delete a product              |

## Filtering and Pagination

List products using optional query parameters.

**Pagination:**

```http
GET /api/v1/products?page=1&limit=5
```

- Default page: `1`
- Default limit: `10`
- Maximum limit: `100`

**Filter by status:**

```http
GET /api/v1/products?status=active
```

Valid statuses: `draft`, `active`, and `archived`.

**Search products:**

```http
GET /api/v1/products?search=mouse
```

**Filter low-stock products:**

```http
GET /api/v1/products?low_stock=true
```

Filters can be combined:

```http
GET /api/v1/products?status=active&low_stock=true&page=1&limit=5
```

## Product Request Example

```json
{
  "name": "Wireless Mouse",
  "sku": "MOUSE-001",
  "description": "Ergonomic wireless computer mouse",
  "price": "25.50",
  "stock_quantity": 30,
  "status": "active"
}
```

Valid product statuses:

- `draft`
- `active`
- `archived`

## Error Handling

The API returns standardized JSON error responses with appropriate HTTP status codes.

| Status                    | Description             |
| ------------------------- | ----------------------- |
| 400 Bad Request           | Invalid input           |
| 404 Not Found             | Product does not exist  |
| 409 Conflict              | Duplicate SKU           |
| 500 Internal Server Error | Unexpected server error |

Example:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Product not found"
  }
}
```

## Development Commands

The Makefile provides convenient commands:

| Command             | Description                    |
| ------------------- | ------------------------------ |
| `make run`          | Run the API locally            |
| `make build`        | Build the API                  |
| `make test`         | Run all Go tests               |
| `make migrate-up`   | Apply database migrations      |
| `make migrate-down` | Roll back the latest migration |
| `make seed`         | Seed sample products           |
| `make sqlc`         | Generate sqlc code             |

## Running Tests

Run all tests:

```bash
go test ./...
```

## Health Check

Check API and database availability:

```http
GET /healthz
```

Returns `200 OK` when the database is reachable, or `503 Service Unavailable` when it is unavailable.

## Stopping the Application

Stop containers while preserving database data:

```bash
docker compose down
```

To remove containers and the database volume:

```bash
docker compose down -v
```

**Warning:** The `-v` option permanently deletes the Docker database volume and its stored data.

## Author

**Tnebeb Zewdu**

GitHub: [TnebebZewdu](https://github.com/TnebebZewdu)
