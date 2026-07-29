package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/cache"
	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Connect to PostgreSQL
	log.Println("Connecting to PostgreSQL...")
	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	log.Println("Connecting to Redis...")
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()
	log.Println("Connected to Redis")

	// Create the service layer
	svc := service.NewService(repo, redisCache, cfg.BaseURL)

	// Create the HTTP handler
	h := handler.NewHandler(svc)

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Setup Gin router
	router := gin.Default()

	// Apply rate limiting to all routes
	router.Use(rateLimiter.Limit())

	// Register routes
	router.POST("/api/shorten", h.ShortenURL)
	router.GET("/:shortCode", h.RedirectURL)
	router.GET("/health", h.HealthCheck)

	// Start the server
	serverAddr := ":" + cfg.Port
	log.Printf("Server starting on http://localhost%s", serverAddr)
	log.Printf("Health check: http://localhost%s/health", serverAddr)

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
