package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/etswifi/ets-noc/internal/monitor"
	"github.com/etswifi/ets-noc/internal/storage"
)

func main() {
	log.Println("Starting ETS Properties Worker...")

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

	maxConcurrentPings := 150 // Default from plan

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

	// Get settings from database
	ctx := context.Background()
	settings, err := postgres.GetSettings(ctx)
	if err == nil && settings.MaxConcurrentPings > 0 {
		maxConcurrentPings = settings.MaxConcurrentPings
	}

	// Create and start pinger
	pinger := monitor.NewPinger(postgres, redis, maxConcurrentPings)

	// Start pinger in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := pinger.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Received shutdown signal")
		pinger.Stop()
	case err := <-errChan:
		log.Printf("Pinger error: %v", err)
	}

	log.Println("Worker stopped")
}
