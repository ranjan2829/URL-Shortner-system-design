package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShortURL represents a shortened URL in the database
type ShortURL struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OriginalURL string             `bson:"original_url" json:"original_url"`
	ShortCode   string             `bson:"short_code" json:"short_code"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	ClickCount  int64              `bson:"click_count" json:"click_count"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
}

// HealthCheck represents a health check record in the database
type HealthCheck struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Status    string             `bson:"status" json:"status"`
	CheckedAt time.Time          `bson:"checked_at" json:"checked_at"`
	Message   string             `bson:"message" json:"message"`
}
