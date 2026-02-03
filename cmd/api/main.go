package main

import (
	"context"
	"log"
	"os"

	"trading-stock/internal/app"
)

func main() {
	// Create application context
	ctx := context.Background()

	// Initialize application with dependency injection
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Run application (blocks until shutdown signal)
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
		os.Exit(1)
	}
}
