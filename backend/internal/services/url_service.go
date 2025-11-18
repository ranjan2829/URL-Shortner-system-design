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
	ErrInvalidURL  = error.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
	ErrURLExpired  = errors.New("URL expired")
	ErrURLInactive = errors.New("URL is inactive")
)

type URLService struct {
	repo       *repository.MongoRepository
	keyService *keyService
}

func NewURLService(repo *repository.MongoRepository, keyService *keyService) *URLService {
	return &URLService{
		repo:       repo,
		keyService: keyService,
	}
}
