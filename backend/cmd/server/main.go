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
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/services"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/repository"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/handlers"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/middleware"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/routes"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/utils"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/validators"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/models"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/database"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main(){
	if err:=godotenv.Load();err!=nil{
		log.Println("No .env file found")

	}
	cfg,err:=config.LoadConfig()
	if err!=nil{
		log.Fatalf("Failed to load Config: %v",err)

	}
	mongoClient,err:=connectMongoDB(cfg.MongoDB.URI)
	if err!=nil{
		log.Fatalf("Failed to connect to MongoDB: %v",err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.println("Connected to MongoDB")

	redisClient:=connectRedis(cfg.Redis.Address,cfg.Redis.Password,cfg.Redis.DB)
	if redisClient==nil{
		log.Fatalf("Failed to connect to Redis: %v",err)
	}
	mongoRepo,err:=repository.NewMongoRepository(mongoClient,cfg.MongoDB.Database,"short_urls")
	if err!=nil{
		log.Fatalf("Failed to create MongoDB repository: %v",err)
	}
	healthRepo:=repository.newHealthCheckRepository(mongoClient,cfg.MongoDB.Database,"health_checks")
	if err!=nil{
		log.Fatalf("Failed to create Health Check repository: %v",err)
	}
	keyService:=services.NewKeyService(redisClient,cfg.KeyGenServiceURL,"short_code_queue")
	urlService:=services.NewKeyService(mongorepo,keyService)

	router:=setupRouter(urlService)
	server:=&http.Server{
		Addr:fmt.Sprintf(":%s",cfg.Server.Port),
		Handler:router,
	}
	go func(){
		log.PrintF("Server starting on port %s",cfg.Server.Port)
		if err:=server.ListenAndServe();err!=nil && err!=http.ErrServerClosed{
			log.Fatalf("Failed to start server: %v",err)
		}
	}()
	quit:=make(chan os.Signal,1)
	signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)
	<-quit
	log.Println("Shutting down the server...")
	ctx,cancel:=context.withTimeout(context.Background(),5*time.Second)
	defer cancel()
	if err:=server shutdown(ctx);err!=nil{
		log.Fatalf("server forced to shutdown:%v",err)

	}
	log.Println("Server shutdown gracefully")
	
}






