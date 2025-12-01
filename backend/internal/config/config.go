package config

import (
	"os"
)

type Config struct {
	Server struct {
		Port string
	}
	MongoDB struct {
		URI      string
		Database string
	}
	Redis struct {
		Address  string
		Password string
		DB       int
	}
	KeyGenServiceURL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")
	cfg.MongoDB.Database = getEnv("MONGODB_DB", "url_shortener")
	cfg.Redis.Address = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.KeyGenServiceURL = getEnv("KEY_GEN_SERVICE_URL", "http://localhost:8081")
	
	// Redis DB is an int, handling it simply here for now, default 0
	cfg.Redis.DB = 0

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
