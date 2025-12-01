# Complete System Architecture & Go Code Deep Dive

## 📚 Table of Contents
1. [System Architecture Overview](#system-architecture-overview)
2. [Go Language Fundamentals Used](#go-language-fundamentals-used)
3. [Project Structure & Code Organization](#project-structure--code-organization)
4. [Low-Level Design (LLD)](#low-level-design-lld)
5. [Data Flow & Request Processing](#data-flow--request-processing)
6. [Component-by-Component Breakdown](#component-by-component-breakdown)
7. [Design Patterns Used](#design-patterns-used)
8. [Best Practices & Conventions](#best-practices--conventions)

---

## System Architecture Overview

### High-Level Architecture

```
┌─────────────┐
│   Client    │ (Browser/Frontend)
└──────┬──────┘
       │ HTTP Request
       ▼
┌─────────────────────────────────────┐
│      API Gateway (Gin Router)       │
│      Port: 8080                      │
└──────┬───────────────────────────────┘
       │
       ├─────────────────┬─────────────────┐
       ▼                 ▼                 ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  Handlers   │  │ Middleware  │  │  Services  │
│  (HTTP)     │  │ (Logging)   │  │ (Business) │
└──────┬──────┘  └─────────────┘  └──────┬─────┘
       │                                  │
       ▼                                  ▼
┌──────────────────────────────────────────────┐
│         Repository Layer (Data Access)        │
└──────┬───────────────────────────────────────┘
       │
       ├──────────────────┬──────────────────┐
       ▼                  ▼                  ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  MongoDB    │  │   Redis    │  │   Kafka     │
│ (Persistence)│  │  (Cache)   │  │ (Events)    │
└─────────────┘  └─────────────┘  └─────────────┘
```

### Component Responsibilities

1. **API Gateway (Gin Router)**: Routes HTTP requests to appropriate handlers
2. **Handlers**: Parse HTTP requests, validate input, call services
3. **Services**: Business logic (URL shortening, key generation)
4. **Repository**: Data access layer (MongoDB operations)
5. **MongoDB**: Persistent storage for URLs
6. **Redis**: Caching and queue for short codes
7. **Kafka**: Event streaming for analytics and async processing

---

## Go Language Fundamentals Used

### 1. **Packages & Imports**

```go
package main  // Entry point package

import (
    "context"           // Context for cancellation/timeouts
    "fmt"               // Formatting and printing
    "log"               // Logging
    "net/http"           // HTTP server
    "time"              // Time operations
    
    // External packages
    "github.com/gin-gonic/gin"  // Web framework
    "go.mongodb.org/mongo-driver/mongo"  // MongoDB driver
)
```

**Key Concepts:**
- `package main`: Special package that creates an executable
- Import paths: Go modules use URLs (e.g., `github.com/user/repo`)
- Public vs Private: Capitalized = exported (public), lowercase = private

### 2. **Structs (Data Structures)**

```go
type URLService struct {
    repo       *repository.MongoRepository  // Dependency injection
    keyService *KeyService                  // Composition
}
```

**Key Concepts:**
- Structs group related data
- Fields can be pointers (`*Type`) or values
- Methods are attached to structs

### 3. **Interfaces (Polymorphism)**

```go
// Implicit interface - if a type has these methods, it implements the interface
type Repository interface {
    CreateShortURL(ctx context.Context, url *ShortURL) error
    GetShortURLByCode(ctx context.Context, code string) (*ShortURL, error)
}
```

**Key Concepts:**
- Go uses implicit interfaces (duck typing)
- No `implements` keyword needed
- Enables dependency injection and testing

### 4. **Pointers & References**

```go
func NewURLService(repo *repository.MongoRepository) *URLService {
    return &URLService{  // & returns address (pointer)
        repo: repo,
    }
}

// Usage
service := NewURLService(repo)  // service is *URLService
service.ShortenURL(...)          // Go automatically dereferences
```

**Key Concepts:**
- `*Type`: Pointer type
- `&value`: Get address of value
- `*pointer`: Dereference pointer
- Pointers allow sharing data without copying

### 5. **Error Handling**

```go
func (s *URLService) ShortenURL(ctx context.Context, url string) (*ShortURL, error) {
    if !isValidURL(url) {
        return nil, ErrInvalidURL  // Return error, nil result
    }
    
    shortURL, err := s.repo.CreateShortURL(ctx, url)
    if err != nil {
        return nil, fmt.Errorf("failed to create: %w", err)  // Wrap error
    }
    
    return shortURL, nil  // Success: return result, nil error
}
```

**Key Concepts:**
- Go doesn't have exceptions
- Functions return `(result, error)` tuple
- `%w` verb wraps errors with context
- Always check errors explicitly

### 6. **Context (Cancellation & Timeouts)**

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()  // Always cancel to free resources

// Pass context to all operations
client, err := mongo.Connect(ctx, options)
```

**Key Concepts:**
- Context carries cancellation signals
- `context.Background()`: Root context
- `context.WithTimeout`: Creates timeout context
- `defer`: Executes when function returns

### 7. **Goroutines & Concurrency**

```go
go func() {  // Start goroutine (lightweight thread)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}()
```

**Key Concepts:**
- `go` keyword starts goroutine
- Goroutines are lightweight (not OS threads)
- Used for async operations

### 8. **Channels (Communication)**

```go
quit := make(chan os.Signal, 1)  // Buffered channel
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // Block until signal received
```

**Key Concepts:**
- Channels enable goroutine communication
- `make(chan Type, buffer)`: Create channel
- `<-channel`: Receive from channel
- `channel <- value`: Send to channel

---

## Project Structure & Code Organization

### Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models
│   ├── repository/          # Data access layer
│   └── services/            # Business logic
└── go.mod                   # Go module definition
```

### Why This Structure?

1. **`cmd/`**: Contains main applications (can have multiple)
2. **`internal/`**: Code that can't be imported by other projects
3. **Separation of Concerns**: Each package has single responsibility

---

## Low-Level Design (LLD)

### 1. Request Flow (Detailed)

```
HTTP Request
    │
    ▼
[Gin Router] ──> Route Matching
    │
    ▼
[Middleware] ──> Logger, CORS, Auth (if any)
    │
    ▼
[Handler] ──> Parse Request Body/Params
    │
    ▼
[Service] ──> Business Logic
    │         ├─> Validate Input
    │         ├─> Check Cache (Redis)
    │         └─> Generate Short Code
    │
    ▼
[Repository] ──> Database Operations
    │           ├─> Check if exists
    │           ├─> Create new record
    │           └─> Update counters
    │
    ▼
[MongoDB] ──> Persist Data
    │
    ▼
[Response] ──> JSON Response to Client
```

### 2. Data Models

```go
type ShortURL struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    OriginalURL string             `bson:"original_url" json:"original_url"`
    ShortCode   string             `bson:"short_code" json:"short_code"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
    ClickCount  int64              `bson:"click_count" json:"click_count"`
    IsActive    bool               `bson:"is_active" json:"is_active"`
}
```

**Key Concepts:**
- **Struct Tags**: `bson:"field"` for MongoDB, `json:"field"` for HTTP
- **Pointers for Optional**: `*time.Time` allows `nil` (optional field)
- **Primitive Types**: MongoDB uses `primitive.ObjectID` for IDs

### 3. Repository Pattern

```go
type MongoRepository struct {
    collection *mongo.Collection  // MongoDB collection
}

func (r *MongoRepository) CreateShortURL(ctx context.Context, url *ShortURL) error {
    _, err := r.collection.InsertOne(ctx, url)
    return err
}
```

**Why Repository Pattern?**
- **Abstraction**: Service doesn't know about MongoDB
- **Testability**: Can mock repository for testing
- **Flexibility**: Can swap MongoDB for another DB

### 4. Service Layer Pattern

```go
type URLService struct {
    repo       *repository.MongoRepository
    keyService *KeyService
}

func (s *URLService) ShortenURL(ctx context.Context, originalURL string) (*ShortURL, error) {
    // 1. Validate
    if !isValidURL(originalURL) {
        return nil, ErrInvalidURL
    }
    
    // 2. Check if exists
    existing, _ := s.repo.GetShortURLByOriginal(ctx, originalURL)
    if existing != nil {
        return existing, nil  // Return existing
    }
    
    // 3. Generate short code
    shortCode, err := s.keyService.GetShortCode(ctx)
    if err != nil {
        return nil, err
    }
    
    // 4. Create and save
    shortURL := &models.ShortURL{
        OriginalURL: originalURL,
        ShortCode:   shortCode,
        CreatedAt:   time.Now(),
        IsActive:    true,
    }
    
    return shortURL, s.repo.CreateShortURL(ctx, shortURL)
}
```

**Service Layer Responsibilities:**
- Business logic
- Validation
- Orchestration (calls multiple repositories/services)
- Error handling

---

## Data Flow & Request Processing

### Example: Shorten URL Request

#### Step 1: HTTP Request Arrives
```go
// Client sends:
POST /api/v1/shorten
Content-Type: application/json
{"url": "https://example.com"}
```

#### Step 2: Router Matches Route
```go
// main.go
api.POST("/shorten", urlHandler.ShortenURL)
```

#### Step 3: Handler Processes Request
```go
// handlers/url_handler.go
func (h *URLHandler) ShortenURL(c *gin.Context) {
    // 1. Parse JSON body
    var req ShortenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 2. Call service
    shortURL, err := h.urlService.ShortenURL(c.Request.Context(), req.URL, nil)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed"})
        return
    }
    
    // 3. Return response
    c.JSON(200, shortURL)
}
```

#### Step 4: Service Business Logic
```go
// services/url_service.go
func (s *URLService) ShortenURL(ctx context.Context, url string) (*ShortURL, error) {
    // Validate
    if !isValidURL(url) {
        return nil, ErrInvalidURL
    }
    
    // Check existing
    existing, _ := s.repo.GetShortURLByOriginal(ctx, url)
    if existing != nil {
        return existing, nil
    }
    
    // Generate code
    code, err := s.keyService.GetShortCode(ctx)
    // ... create and save
}
```

#### Step 5: Repository Saves to Database
```go
// repository/mongo_repository.go
func (r *MongoRepository) CreateShortURL(ctx context.Context, url *ShortURL) error {
    _, err := r.collection.InsertOne(ctx, url)
    return err
}
```

#### Step 6: Response Sent Back
```json
{
  "short_url": "http://localhost:8080/ABC123",
  "short_code": "ABC123",
  "original_url": "https://example.com"
}
```

---

## Component-by-Component Breakdown

### 1. Main Entry Point (`cmd/server/main.go`)

```go
func main() {
    // 1. Load configuration
    cfg, err := config.LoadConfig()
    
    // 2. Connect to databases
    mongoClient := connectMongoDB(cfg.MongoDB.URI)
    redisClient := connectRedis(cfg.Redis.Address, ...)
    
    // 3. Initialize layers
    mongoRepo := repository.NewMongoRepository(...)
    keyService := services.NewKeyService(...)
    urlService := services.NewURLService(mongoRepo, keyService)
    
    // 4. Setup routes
    router := setupRouter(urlService, keyService)
    
    // 5. Start server
    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    
    // 6. Graceful shutdown
    go server.ListenAndServe()
    <-quit  // Wait for interrupt
    server.Shutdown(ctx)
}
```

**Key Concepts:**
- **Dependency Injection**: Services receive dependencies via constructors
- **Graceful Shutdown**: Server stops cleanly on SIGINT/SIGTERM
- **Initialization Order**: Config → DB → Repos → Services → Handlers → Server

### 2. Configuration (`internal/config/config.go`)

```go
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
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    cfg.Server.Port = getEnv("PORT", "8080")  // Default: 8080
    cfg.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")
    // ...
    return cfg, nil
}
```

**Key Concepts:**
- **Environment Variables**: Configuration from env or defaults
- **Struct Embedding**: Nested structs for organization
- **Zero Values**: Go initializes structs to zero values

### 3. Handlers (`internal/handlers/url_handler.go`)

```go
type URLHandler struct {
    urlService *services.URLService  // Dependency
}

func NewURLHandler(urlService *services.URLService) *URLHandler {
    return &URLHandler{urlService: urlService}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
    // Handler is thin - just HTTP concerns
    // Business logic is in service
}
```

**Key Concepts:**
- **Handler Pattern**: Handlers are HTTP-specific
- **Constructor Functions**: `NewXxx()` creates instances
- **Method Receivers**: `(h *URLHandler)` attaches method to type

### 4. Services (`internal/services/url_service.go`)

```go
type URLService struct {
    repo       *repository.MongoRepository
    keyService *KeyService
}

func (s *URLService) ShortenURL(ctx context.Context, url string) (*ShortURL, error) {
    // Business logic here
    // 1. Validate
    // 2. Check existing
    // 3. Generate code
    // 4. Save
}
```

**Key Concepts:**
- **Service Layer**: Contains business logic
- **Composition**: Service uses other services
- **Context Propagation**: Pass context through all layers

### 5. Repository (`internal/repository/mongo_repository.go`)

```go
type MongoRepository struct {
    collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, dbName, collName string) (*MongoRepository, error) {
    db := client.Database(dbName)
    collection := db.Collection(collName)
    
    // Create indexes
    indexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "short_code", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    collection.Indexes().CreateOne(context.Background(), indexModel)
    
    return &MongoRepository{collection: collection}, nil
}
```

**Key Concepts:**
- **Repository Pattern**: Abstracts database operations
- **Indexes**: Improve query performance
- **BSON**: Binary JSON used by MongoDB

---

## Design Patterns Used

### 1. **Dependency Injection**

```go
// Instead of creating dependencies inside:
func NewURLService() *URLService {
    repo := repository.NewMongoRepository(...)  // ❌ Tight coupling
    
    return &URLService{repo: repo}
}

// Inject dependencies:
func NewURLService(repo *repository.MongoRepository) *URLService {
    return &URLService{repo: repo}  // ✅ Loose coupling
}
```

**Benefits:**
- Testable (can inject mocks)
- Flexible (can swap implementations)
- Clear dependencies

### 2. **Repository Pattern**

```go
// Service doesn't know about MongoDB
type Repository interface {
    CreateShortURL(ctx context.Context, url *ShortURL) error
}

// Can swap implementations:
// - MongoRepository
// - PostgresRepository
// - InMemoryRepository (for testing)
```

### 3. **Service Layer Pattern**

```go
// Handlers → Services → Repositories
// Each layer has single responsibility:
// - Handler: HTTP concerns
// - Service: Business logic
// - Repository: Data access
```

### 4. **Factory Pattern**

```go
// Constructor functions create instances
func NewURLService(...) *URLService
func NewMongoRepository(...) *MongoRepository
func NewKeyService(...) *KeyService
```

---

## Best Practices & Conventions

### 1. **Error Handling**

```go
// ✅ Good: Check and wrap errors
if err != nil {
    return nil, fmt.Errorf("failed to create URL: %w", err)
}

// ❌ Bad: Ignore errors
_ = repo.CreateShortURL(ctx, url)
```

### 2. **Context Usage**

```go
// ✅ Always pass context
func (r *Repository) Create(ctx context.Context, url *ShortURL) error

// ❌ Don't use background context in functions
func (r *Repository) Create(url *ShortURL) error {
    ctx := context.Background()  // ❌
}
```

### 3. **Naming Conventions**

```go
// Exported (public): Capitalized
func NewURLService() *URLService  // Can be imported

// Unexported (private): lowercase
func generateShortCode() string   // Internal only

// Interfaces: Often end with "er"
type Reader interface { ... }
type Writer interface { ... }
```

### 4. **Package Organization**

```go
// ✅ Good: Clear package boundaries
package handlers  // Only HTTP handlers
package services  // Only business logic
package repository // Only data access

// ❌ Bad: Mixed concerns
package utils  // Everything mixed together
```

### 5. **Defer for Cleanup**

```go
// ✅ Always defer cleanup
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()  // Executes when function returns

defer mongoClient.Disconnect(context.Background())
```

---

## Advanced Concepts

### 1. **Goroutines for Async Operations**

```go
// Start server in goroutine
go func() {
    server.ListenAndServe()
}()

// Main goroutine waits for shutdown signal
<-quit
```

### 2. **Channel Communication**

```go
// Signal handling
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // Block until signal
```

### 3. **Struct Embedding**

```go
// Can embed structs for composition
type BaseService struct {
    logger *Logger
}

type URLService struct {
    BaseService  // Embeds logger
    repo *Repository
}

// Can access: urlService.logger.Log(...)
```

### 4. **Method Sets**

```go
// Value receiver: works with both T and *T
func (s URLService) Method() {}

// Pointer receiver: only works with *T
func (s *URLService) Method() {}
```

---

## Summary: Key Takeaways

1. **Go is Simple**: No classes, inheritance, or exceptions
2. **Composition over Inheritance**: Use structs and interfaces
3. **Explicit Error Handling**: Always check errors
4. **Context for Cancellation**: Pass context everywhere
5. **Packages for Organization**: Clear boundaries
6. **Pointers for Sharing**: Avoid copying large structs
7. **Interfaces for Flexibility**: Enable testing and swapping implementations

This architecture follows **Clean Architecture** principles:
- **Independence**: Business logic doesn't depend on frameworks
- **Testability**: Easy to test each layer
- **Flexibility**: Can swap databases, frameworks, etc.
- **Maintainability**: Clear separation of concerns

