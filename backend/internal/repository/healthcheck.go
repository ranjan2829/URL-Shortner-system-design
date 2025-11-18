package repository

import (
	"context"

	"github.com/ranjanshahajishitole/url-shortener/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// HealthCheckRepository handles MongoDB operations for health checks
type HealthCheckRepository struct {
	collection *mongo.Collection
}

// NewHealthCheckRepository creates a new health check repository instance
func NewHealthCheckRepository(client *mongo.Client, dbName, collectionName string) *HealthCheckRepository {
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	return &HealthCheckRepository{
		collection: collection,
	}
}

// SaveHealthCheck saves a health check record to the database
func (r *HealthCheckRepository) SaveHealthCheck(ctx context.Context, healthCheck *models.HealthCheck) error {
	_, err := r.collection.InsertOne(ctx, healthCheck)
	return err
}

// GetLatestHealthCheck retrieves the most recent health check record
func (r *HealthCheckRepository) GetLatestHealthCheck(ctx context.Context) (*models.HealthCheck, error) {
	var healthCheck models.HealthCheck
	opts := options.FindOne().SetSort(bson.D{{Key: "checked_at", Value: -1}})

	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&healthCheck)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &healthCheck, nil
}
