package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyServiceUnavailable = errors.New("key service unavailable")
	ErrRedisUnavailable      = errors.New("redis unavailable")
)

type KeyService struct {
	redisClient *redis.Client
	httpClient  *http.Client
	serviceURL  string
	queueName   string
}

func NewKeyService(redisClient *redis.Client, serviceURL, queueName string) *KeyService {
	return &KeyService{
		redisClient: redisClient,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		serviceURL: serviceURL,
		queueName:  queueName,
	}
}
func (s *KeyService) GetShortCode(ctx context.Context) (string, error) {
	// Try to get from Redis queue first
	shortCode, err := s.getFromRedisQueue(ctx)
	if err == nil && shortCode != "" {
		return shortCode, nil
	}
	
	// Generate locally instead of calling external service
	shortCode = s.generateShortCode()
	return shortCode, nil
}
func (s *KeyService) getFromRedisQueue(ctx context.Context) (string, error) {
	if s.redisClient == nil {
		return "", ErrRedisUnavailable
	}
	result, err := s.redisClient.LPop(ctx, s.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get short code from redis: %w", err)
	}
	return result, nil
}
func (s *KeyService) getFromKeyGenService(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.serviceURL+"/generate", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get short code from key generation service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get short code from key generation service: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	var response struct {
		ShortCode string `json:"short_code"`
		Error     string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if response.Error != "" {
		return "", fmt.Errorf("failed to get short code from key generation service: %s", response.Error)
	}
	return response.ShortCode, nil
}

// generateShortCode generates a random short code locally
// Uses base64 URL-safe encoding for shorter codes (6-8 characters)
func (s *KeyService) generateShortCode() string {
	// Generate 6 random bytes
	b := make([]byte, 6)
	rand.Read(b)
	
	// Encode to base64 URL-safe string and take first 8 characters
	encoded := base64.URLEncoding.EncodeToString(b)
	// Remove padding and take 8 chars for short code
	code := strings.TrimRight(encoded, "=")
	if len(code) > 8 {
		code = code[:8]
	}
	return code
}

// GenerateShortCode is a public method to generate a short code
// Used by handlers for the /generate endpoint
func (s *KeyService) GenerateShortCode() string {
	return s.generateShortCode()
}
