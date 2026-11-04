# TETER Product Inventory API

A RESTful Product Inventory API built with Go, PostgreSQL, Chi, sqlc, and Docker. It supports product creation, retrieval, updating, deletion, and pagination.

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
- Input validation
- Unique SKU validation
- Standard JSON error responses
- Database migrations
- Idempotent product seeder (10 sample products)
- Dockerized API and PostgreSQL database

## Project Structure

```text
.
├── cmd/api/              # Application entry point
├── db/
│   ├── migrations/       # Database migrations
│   ├── queries/          # SQL queries for sqlc
│   └── seed/             # Sample product data
├── internal/
│   ├── database/         # Database connection and generated sqlc code
│   └── handlers/         # Product API handlers
├── Dockerfile
├── docker-compose.yml
├── sqlc.yaml
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.21+
- Docker Desktop
- Docker Compose
- golang-migrate CLI

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

The API will be available at `http://localhost:8080`.

PostgreSQL is exposed on port `5433` on the host.

### 4. Apply database migration

migrate -path db/migrations -database "postgres://postgres:<URL_ENCODED_PASSWORD>@localhost:5433/teter_inventory?sslmode=disable" up

### 5. Seed sample products

```powershell
Get-Content .\db\seed\products.sql -Raw | docker compose exec -T db psql -U postgres -d teter_inventory
```

The seeder inserts 10 sample products and avoids duplicates using the SKU constraint.

## API Endpoints

Base URL: `http://localhost:8080`

| Method | Endpoint                | Description      |
| ------ | ----------------------- | ---------------- |
| POST   | `/api/v1/products`      | Create a product |
| GET    | `/api/v1/products`      | List products    |
| GET    | `/api/v1/products/{id}` | Get a product    |
| PUT    | `/api/v1/products/{id}` | Update a product |
| DELETE | `/api/v1/products/{id}` | Delete a product |

### Pagination

List products with optional query parameters:

```http
GET /api/v1/products?page=1&limit=5
```

Default page: `1`

Default limit: `10`

Maximum limit: `100`

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

The API uses JSON error responses and appropriate HTTP status codes:

- `400 Bad Request` – Invalid input
- `404 Not Found` – Product does not exist
- `409 Conflict` – Duplicate SKU
- `500 Internal Server Error` – Unexpected server error

## Stopping the Application

```bash
docker compose down
```

To remove containers and the database volume:

```bash
docker compose down -v
```

**Warning:** The `-v` option permanently deletes the Docker database data.

## Author

**Tnebeb Zewdu**

GitHub: [TnebebZewdu](https://github.com/TnebebZewdu)
