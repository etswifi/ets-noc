package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etswifi/ets-noc/internal/api"
	"github.com/etswifi/ets-noc/internal/gcs"
	"github.com/etswifi/ets-noc/internal/storage"
)

func main() {
	log.Println("Starting ETS Properties API server...")

	// Get environment variables
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is required")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	gcsBucket := os.Getenv("GCS_BUCKET")
	if gcsBucket == "" {
		log.Fatal("GCS_BUCKET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize storage
	postgres, err := storage.NewPostgresStore(postgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgres.Close()
	log.Println("Connected to PostgreSQL")

	redis, err := storage.NewRedisStore(redisAddr, redisPassword, 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("Connected to Redis")

	// Initialize GCS client
	ctx := context.Background()
	gcsClient, err := gcs.NewClient(ctx, gcsBucket)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}
	defer gcsClient.Close()
	log.Println("Connected to GCS")

	// Create server and setup routes
	server := api.NewServer(postgres, redis, gcsClient)
	router := server.SetupRouter()

	// Start HTTP server
	go func() {
		log.Printf("API server listening on port %s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	time.Sleep(2 * time.Second)
	log.Println("Server stopped")
}
