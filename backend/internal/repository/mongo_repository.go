package repository

import (
	"context"
	"time"

	"github.com/ranjanshahajishitole/url-shortener/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository handles MongoDB operations for short URLs
// This is the data access layer - it only deals with database operations
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository instance
// Parameters:
//   - client: MongoDB client connection
//   - dbName: Database name
//   - collectionName: Collection name for short URLs
func NewMongoRepository(client *mongo.Client, dbName, collectionName string) (*MongoRepository, error) {
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	// Create index on short_code for faster lookups
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "short_code", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, err
	}

	// Create index on original_url for faster lookups
	indexModel2 := mongo.IndexModel{
		Keys: bson.D{{Key: "original_url", Value: 1}},
	}
	collection.Indexes().CreateOne(context.Background(), indexModel2)

	return &MongoRepository{
		collection: collection,
	}, nil
}

// CreateShortURL saves a new short URL to the database
// It sets CreatedAt and IsActive fields automatically
func (r *MongoRepository) CreateShortURL(ctx context.Context, shortURL *models.ShortURL) error {
	// Set default values if not already set
	if shortURL.CreatedAt.IsZero() {
		shortURL.CreatedAt = time.Now()
	}
	if !shortURL.IsActive {
		shortURL.IsActive = true
	}
	if shortURL.ClickCount == 0 {
		shortURL.ClickCount = 0
	}

	_, err := r.collection.InsertOne(ctx, shortURL)
	return err
}

// GetShortURLByCode retrieves a short URL by its short code
func (r *MongoRepository) GetShortURLByCode(ctx context.Context, shortCode string) (*models.ShortURL, error) {
	var shortURL models.ShortURL
	err := r.collection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&shortURL)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, err
	}
	return &shortURL, nil
}

// GetShortURLByOriginal retrieves a short URL by its original URL
func (r *MongoRepository) GetShortURLByOriginal(ctx context.Context, originalURL string) (*models.ShortURL, error) {
	var shortURL models.ShortURL
	err := r.collection.FindOne(ctx, bson.M{"original_url": originalURL}).Decode(&shortURL)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil, nil if not found (not an error)
		}
		return nil, err
	}
	return &shortURL, nil
}

// UpdateClickCount increments the click count for a short URL
func (r *MongoRepository) UpdateClickCount(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$inc": bson.M{"click_count": 1}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetStats retrieves statistics for a short URL (same as GetShortURLByCode)
// This method exists for semantic clarity - you might want to add more stats later
func (r *MongoRepository) GetStats(ctx context.Context, shortCode string) (*models.ShortURL, error) {
	return r.GetShortURLByCode(ctx, shortCode)
}

