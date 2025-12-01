package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/config"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/handlers"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/middleware"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/repository"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/services"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load Config: %v", err)
	}
	mongoClient, err := connectMongoDB(cfg.MongoDB.URI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Println("Connected to MongoDB")

	redisClient := connectRedis(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if redisClient == nil {
		log.Fatalf("Failed to connect to Redis")
	}
	mongoRepo, err := repository.NewMongoRepository(mongoClient, cfg.MongoDB.Database, "short_urls")
	if err != nil {
		log.Fatalf("Failed to create MongoDB repository: %v", err)
	}
	_ = repository.NewHealthCheckRepository(mongoClient, cfg.MongoDB.Database, "health_checks") // Reserved for future health check endpoints
	keyService := services.NewKeyService(redisClient, cfg.KeyGenServiceURL, "short_code_queue")
	urlService := services.NewURLService(mongoRepo, keyService)

	router := setupRouter(urlService, keyService)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server shutdown gracefully")
}

// setupRouter configures all the routes for the application
func setupRouter(urlService *services.URLService, keyService *services.KeyService) *gin.Engine {
	router := gin.Default()

	// Add middleware (logging, CORS, etc.)
	router.Use(middleware.Logger())

	// Create handlers
	urlHandler := handlers.NewURLHandler(urlService)
	keyHandler := handlers.NewKeyHandler(keyService)

	// API routes
	api := router.Group("/api/v1")
	api.POST("/shorten", urlHandler.ShortenURL)
	api.GET("/generate", keyHandler.GenerateKey) // Key generation endpoint
	api.GET("/:code/stats", urlHandler.GetStats)

	// Redirect route (should be last to avoid conflicts)
	router.GET("/:code", urlHandler.RedirectURL)

	return router
}
func connectMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func connectRedis(address, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil
	}
	return client
}
