
.PHONY: run build test migrate-up migrate-down seed sqlc

# Run the API locally
run:
	go run ./cmd/api

# Build the API
build:
	go build -o bin/api ./cmd/api

# Run all Go tests
test:
	go test ./...

# Apply database migrations
migrate-up:
	migrate -path db/migrations -database "%DATABASE_URL%" up

# Roll back the latest migration
migrate-down:
	migrate -path db/migrations -database "%DATABASE_URL%" down 1
# Seed the database
seed:
	go run ./cmd/seeder

# Generate sqlc code
sqlc:
	sqlc generate
