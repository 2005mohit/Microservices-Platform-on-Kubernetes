package main

import (
	"log"
	"os"

	"github.com/devplatform/api/internal/api"
	"github.com/devplatform/api/internal/db"
	"github.com/devplatform/api/internal/k8s"
	"github.com/devplatform/api/internal/queue"
)

func main() {
	pgDSN := getEnv("DATABASE_URL", "postgres://devplatform:devplatform@localhost:5432/devplatform?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	port := getEnv("PORT", "8080")
	domain := getEnv("BASE_DOMAIN", "localhost")

	if err := db.Init(pgDSN); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if err := queue.Init(redisAddr); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
	defer queue.Close()

	if err := k8s.Init(); err != nil {
		log.Printf("warn: kubernetes not available: %v", err)
	}

	router := api.NewRouter(domain)
	log.Printf("server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
