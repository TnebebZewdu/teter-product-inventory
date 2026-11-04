package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
	"github.com/TnebebZewdu/teter-product-inventory/internal/handlers"
)

func main() {
	// Read the database connection URL.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to PostgreSQL.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	// Initialize sqlc queries.
	queries := database.New(pool)
	productHandler := handlers.NewProductHandler(queries)
	_ = queries // We'll use these queries when we build product handlers.

	// Configure the HTTP router.
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to TETER Product Inventory API!")
	})
	r.Post("/api/v1/products", productHandler.CreateProduct)
	r.Get("/api/v1/products", productHandler.ListProducts)
	r.Get("/api/v1/products/{id}", productHandler.GetProductByID)
	r.Put("/api/v1/products/{id}", productHandler.UpdateProduct)
	r.Delete("/api/v1/products/{id}", productHandler.DeleteProduct)

	fmt.Println("Server running at http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", r))
}