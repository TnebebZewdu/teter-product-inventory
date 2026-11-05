
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
)

func main() {
	// Read the database connection URL.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Read the SQL seed file.
	seedSQL, err := os.ReadFile("db/seed/products.sql")
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	// Connect to PostgreSQL.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	// Execute the seed SQL.
	_, err = pool.Exec(ctx, string(seedSQL))
	if err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Println("Product seeding completed successfully!")
}
