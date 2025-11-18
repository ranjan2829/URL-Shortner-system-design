package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ranjanshahajishitole/url-shortener/backend/internal/models"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/repository"
)

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
	ErrURLExpired  = errors.New("URL expired")
	ErrURLInactive = errors.New("URL is inactive")
)

type URLService struct {
	repo       *repository.MongoRepository
	keyService *KeyService
}

func NewURLService(repo *repository.MongoRepository, keyService *KeyService) *URLService {
	return &URLService{
		repo:       repo,
		keyService: keyService,
	}
}

func (s *URLService) ShortenURL(ctx context.Context, originalURL string, expiresIn *time.Duration) (*models.ShortURL, error) {
	if !isValidURL(originalURL) {
		return nil, ErrInvalidURL
	}
	existing, _ := s.repo.GetShortURLByOriginal(ctx, originalURL)
	if existing != nil {
		return existing, nil
	}
	shortCode, err := s.keyService.GetShortCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}
	shortURL := &models.ShortURL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
	}
	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		shortURL.ExpiresAt = &expiresAt
	}
	if err := s.repo.CreateShortURL(ctx, shortURL); err != nil {
		return nil, fmt.Errorf("failed to create short URL: %w", err)
	}
	return shortURL, nil
}
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	shortURL, err := s.repo.GetShortURLByCode(ctx, shortCode)
	if err != nil {
		return "", ErrURLNotFound
	}
	if !shortURL.IsActive {
		return "", ErrURLInactive
	}
	if shortURL.ExpiresAt != nil && time.Now().After(*shortURL.ExpiresAt) {
		return "", ErrURLExpired
	}
	if err := s.repo.UpdateClickCount(ctx, shortCode); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update click count: %v\n", err)
	}
	return shortURL.OriginalURL, nil
}

func (s *URLService) GetStats(ctx context.Context, shortCode string) (*models.ShortURL, error) {
	shortURL, err := s.repo.GetStats(ctx, shortCode)
	if err != nil {
		return nil, ErrURLNotFound
	}
	return shortURL, nil
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}
