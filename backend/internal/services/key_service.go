package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	shortCode, err := s.getFromRedisQueue(ctx)
	if err == nil && shortCode != "" {
		return shortCode, nil
	}
	shortCode, err = s.getFromKeyGenService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get short code from key generation service: %w", err)
	}
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
