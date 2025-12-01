# Complete Code Walkthrough - Line by Line Explanation

This document explains every file in the codebase with detailed explanations.

---

## 📁 File: `cmd/server/main.go` - Application Entry Point

### Purpose
This is the **entry point** of the application. It initializes all components and starts the HTTP server.

### Code Breakdown

```go
package main  // 'main' package creates an executable
```

**Why `package main`?**
- Go programs start from `main` package
- Must have a `main()` function
- Creates a binary executable

```go
import (
    "context"      // For cancellation and timeouts
    "fmt"           // String formatting
    "log"           // Logging
    "net/http"      // HTTP server
    "os"            // OS operations (signals)
    "os/signal"     // Signal handling
    "syscall"       // System calls
    "time"          // Time operations
    
    // External dependencies
    "github.com/gin-gonic/gin"           // Web framework
    "github.com/joho/godotenv"           // .env file loader
    "github.com/redis/go-redis/v9"       // Redis client
    "go.mongodb.org/mongo-driver/mongo"  // MongoDB driver
)
```

**Import Groups:**
1. **Standard library**: Built-in Go packages
2. **External packages**: Third-party dependencies

```go
func main() {
    // 1. Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")  // Not fatal, use defaults
    }
```

**What's happening:**
- Tries to load `.env` file
- If not found, continues (uses default values)
- `log.Println`: Logs to standard output

```go
    // 2. Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load Config: %v", err)  // Exit if fails
    }
```

**Key Concepts:**
- `config.LoadConfig()`: Reads environment variables
- `log.Fatalf()`: Logs error and exits program
- `%v`: Verbose format (prints any type)

```go
    // 3. Connect to MongoDB
    mongoClient, err := connectMongoDB(cfg.MongoDB.URI)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer mongoClient.Disconnect(context.Background())
```

**Key Concepts:**
- `defer`: Executes when function returns (cleanup)
- Always disconnect database connections
- `context.Background()`: Root context (no timeout)

```go
    // 4. Connect to Redis
    redisClient := connectRedis(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
    if redisClient == nil {
        log.Fatalf("Failed to connect to Redis")
    }
```

**Why check for `nil`?**
- `connectRedis` returns `nil` on failure
- Go doesn't have exceptions, so functions return `nil` for errors

```go
    // 5. Initialize Repository Layer
    mongoRepo, err := repository.NewMongoRepository(
        mongoClient, 
        cfg.MongoDB.Database, 
        "short_urls",
    )
```

**Dependency Injection:**
- Pass dependencies to constructors
- Makes code testable (can inject mocks)

```go
    // 6. Initialize Services
    keyService := services.NewKeyService(redisClient, cfg.KeyGenServiceURL, "short_code_queue")
    urlService := services.NewURLService(mongoRepo, keyService)
```

**Service Composition:**
- `urlService` uses `keyService`
- Services depend on repositories
- Clear dependency chain

```go
    // 7. Setup HTTP Router
    router := setupRouter(urlService, keyService)
    
    // 8. Create HTTP Server
    server := &http.Server{
        Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
        Handler: router,
    }
```

**HTTP Server:**
- `&http.Server{}`: Creates server struct
- `Addr`: Address to listen on (`:8080`)
- `Handler`: Router that handles requests

```go
    // 9. Start server in goroutine
    go func() {
        log.Printf("Server starting on port %s", cfg.Server.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
```

**Goroutine Explanation:**
- `go func() {...}()`: Starts async function
- `ListenAndServe()`: Blocks until server stops
- Runs in background so main can handle shutdown

```go
    // 10. Graceful Shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit  // Block until signal received
    
    log.Println("Shutting down the server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("server forced to shutdown: %v", err)
    }
    log.Println("Server shutdown gracefully")
}
```

**Graceful Shutdown:**
1. Create channel for signals
2. Notify on SIGINT (Ctrl+C) or SIGTERM
3. Block until signal received
4. Give server 5 seconds to finish requests
5. Shutdown cleanly

---

## 📁 File: `internal/config/config.go` - Configuration Management

### Purpose
Loads configuration from environment variables with sensible defaults.

### Code Breakdown

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
    KeyGenServiceURL string
}
```

**Struct Organization:**
- Nested structs group related config
- `struct {}`: Anonymous struct (no name needed)
- Clear, hierarchical structure

```go
func LoadConfig() (*Config, error) {
    cfg := &Config{}  // Create pointer to Config
    
    // Set defaults or read from environment
    cfg.Server.Port = getEnv("PORT", "8080")
    cfg.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")
    // ...
    
    return cfg, nil
}
```

**Why return `(*Config, error)`?**
- Pointer: Avoids copying large struct
- Error: Can return error if config invalid

```go
func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value  // Found in environment
    }
    return fallback  // Use default
}
```

**Key Concepts:**
- `os.LookupEnv()`: Returns `(value, found)`
- `ok`: Boolean indicating if key exists
- Fallback pattern: Always have defaults

---

## 📁 File: `internal/models/models.go` - Data Models

### Purpose
Defines data structures used throughout the application.

### Code Breakdown

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

**Struct Tags Explained:**
- `bson:"_id"`: MongoDB field name
- `json:"id"`: JSON field name
- `omitempty`: Omit if zero value
- Different names for DB vs API

**Why `*time.Time` for ExpiresAt?**
- Pointer allows `nil` (optional field)
- `time.Time` zero value is not `nil`, it's `0001-01-01`
- Pointer: `nil` = not set, `*time` = set

**Field Types:**
- `primitive.ObjectID`: MongoDB's ID type
- `int64`: Large integer (for click counts)
- `bool`: Boolean flag

---

## 📁 File: `internal/repository/mongo_repository.go` - Data Access Layer

### Purpose
Handles all MongoDB operations. Abstracts database details from services.

### Code Breakdown

```go
type MongoRepository struct {
    collection *mongo.Collection  // MongoDB collection reference
}
```

**Why store `collection`?**
- Collection is where documents are stored
- Reuse connection for all operations
- Efficient (no lookup each time)

```go
func NewMongoRepository(client *mongo.Client, dbName, collectionName string) (*MongoRepository, error) {
    db := client.Database(dbName)
    collection := db.Collection(collectionName)
```

**MongoDB Hierarchy:**
- Client → Database → Collection → Document
- `client.Database()`: Get database
- `db.Collection()`: Get collection

```go
    // Create index on short_code for faster lookups
    indexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "short_code", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err := collection.Indexes().CreateOne(context.Background(), indexModel)
```

**Indexes Explained:**
- **Index**: Speeds up queries (like book index)
- `short_code`: Field to index
- `Value: 1`: Ascending order
- `Unique: true`: No duplicates allowed
- **Why?**: Fast lookups by short_code

```go
func (r *MongoRepository) CreateShortURL(ctx context.Context, shortURL *models.ShortURL) error {
    // Set defaults if not set
    if shortURL.CreatedAt.IsZero() {
        shortURL.CreatedAt = time.Now()
    }
    if !shortURL.IsActive {
        shortURL.IsActive = true
    }
    
    _, err := r.collection.InsertOne(ctx, shortURL)
    return err
}
```

**Method Receiver:**
- `(r *MongoRepository)`: Method belongs to MongoRepository
- `r`: Receiver name (like `this` in other languages)
- `*`: Pointer receiver (modifies struct)

**Why set defaults in repository?**
- Ensures data consistency
- Service doesn't need to remember defaults
- Single source of truth

```go
func (r *MongoRepository) GetShortURLByCode(ctx context.Context, shortCode string) (*models.ShortURL, error) {
    var shortURL models.ShortURL
    err := r.collection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&shortURL)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, err  // Not found
        }
        return nil, err  // Other error
    }
    return &shortURL, nil
}
```

**Query Explanation:**
- `bson.M{"short_code": shortCode}`: MongoDB query filter
- `FindOne()`: Get single document
- `Decode(&shortURL)`: Convert BSON to Go struct
- `&shortURL`: Pass pointer to decode into

**Error Handling:**
- `mongo.ErrNoDocuments`: Specific error for "not found"
- Can handle differently if needed

```go
func (r *MongoRepository) UpdateClickCount(ctx context.Context, shortCode string) error {
    filter := bson.M{"short_code": shortCode}
    update := bson.M{"$inc": bson.M{"click_count": 1}}
    _, err := r.collection.UpdateOne(ctx, filter, update)
    return err
}
```

**MongoDB Update Operators:**
- `$inc`: Increment field by value
- Atomic operation (thread-safe)
- Efficient (no read-modify-write)

---

## 📁 File: `internal/services/url_service.go` - Business Logic

### Purpose
Contains business logic for URL shortening. Orchestrates repository and key service.

### Code Breakdown

```go
type URLService struct {
    repo       *repository.MongoRepository
    keyService *KeyService
}
```

**Composition:**
- Service contains dependencies
- Not inheritance (Go doesn't have it)
- "Has-a" relationship

```go
func NewURLService(repo *repository.MongoRepository, keyService *KeyService) *URLService {
    return &URLService{
        repo:       repo,
        keyService: keyService,
    }
}
```

**Constructor Pattern:**
- `NewXxx()`: Convention for constructors
- Returns pointer (efficient)
- Sets up dependencies

```go
func (s *URLService) ShortenURL(ctx context.Context, originalURL string, expiresIn *time.Duration) (*models.ShortURL, error) {
    // 1. Validate URL
    if !isValidURL(originalURL) {
        return nil, ErrInvalidURL
    }
```

**Validation First:**
- Fail fast if invalid
- Don't waste resources
- Clear error messages

```go
    // 2. Check if URL already shortened
    existing, _ := s.repo.GetShortURLByOriginal(ctx, originalURL)
    if existing != nil {
        return existing, nil  // Return existing short URL
    }
```

**Idempotency:**
- Same input → same output
- Don't create duplicates
- `_`: Ignore error (we check `existing != nil`)

```go
    // 3. Generate short code
    shortCode, err := s.keyService.GetShortCode(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate short code: %w", err)
    }
```

**Error Wrapping:**
- `fmt.Errorf()`: Creates new error
- `%w`: Wraps original error (preserves stack)
- Adds context: "failed to generate short code"

```go
    // 4. Create ShortURL object
    shortURL := &models.ShortURL{
        OriginalURL: originalURL,
        ShortCode:   shortCode,
        CreatedAt:   time.Now(),
        IsActive:    true,
        ClickCount:  0,
    }
    
    // 5. Set expiration if provided
    if expiresIn != nil {
        expiresAt := time.Now().Add(*expiresIn)
        shortURL.ExpiresAt = &expiresAt
    }
```

**Struct Literal:**
- `&models.ShortURL{...}`: Create and get pointer
- Field names: Clear what each value is
- `*expiresIn`: Dereference pointer to get value

```go
    // 6. Save to database
    if err := s.repo.CreateShortURL(ctx, shortURL); err != nil {
        return nil, fmt.Errorf("failed to create short URL: %w", err)
    }
    
    return shortURL, nil
}
```

**Error Propagation:**
- Wrap errors with context
- Return to caller
- Caller decides how to handle

```go
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
    // 1. Get from database
    shortURL, err := s.repo.GetShortURLByCode(ctx, shortCode)
    if err != nil {
        return "", ErrURLNotFound
    }
    
    // 2. Check if active
    if !shortURL.IsActive {
        return "", ErrURLInactive
    }
    
    // 3. Check if expired
    if shortURL.ExpiresAt != nil && time.Now().After(*shortURL.ExpiresAt) {
        return "", ErrURLExpired
    }
    
    // 4. Update click count (fire and forget)
    if err := s.repo.UpdateClickCount(ctx, shortCode); err != nil {
        fmt.Printf("Failed to update click count: %v\n", err)
        // Don't fail the request
    }
    
    return shortURL.OriginalURL, nil
}
```

**Business Rules:**
- Multiple validation checks
- Order matters (check active before expired)
- Non-critical operations don't fail request

**Why fire-and-forget for click count?**
- Analytics is not critical
- Don't slow down redirect
- Can retry later if needed

```go
func isValidURL(rawURL string) bool {
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return false
    }
    return parsedURL.Scheme != "" && parsedURL.Host != ""
}
```

**URL Validation:**
- Use standard library `url.Parse()`
- Check for scheme (`http://`, `https://`)
- Check for host (domain name)
- Simple but effective

---

## 📁 File: `internal/services/key_service.go` - Key Generation

### Purpose
Generates unique short codes for URLs.

### Code Breakdown

```go
type KeyService struct {
    redisClient *redis.Client
    httpClient  *http.Client
    serviceURL  string
    queueName   string
}
```

**Multiple Dependencies:**
- Redis: For caching/pre-generated codes
- HTTP: For external service (if needed)
- Config: Service URL and queue name

```go
func (s *KeyService) GetShortCode(ctx context.Context) (string, error) {
    // 1. Try Redis queue first (fastest)
    shortCode, err := s.getFromRedisQueue(ctx)
    if err == nil && shortCode != "" {
        return shortCode, nil
    }
    
    // 2. Generate locally (fallback)
    shortCode = s.generateShortCode()
    return shortCode, nil
}
```

**Fallback Strategy:**
1. Try Redis (pre-generated codes)
2. Generate locally if Redis empty
3. Always succeeds (no external dependency)

```go
func (s *KeyService) generateShortCode() string {
    // Generate 6 random bytes
    b := make([]byte, 6)
    rand.Read(b)
    
    // Encode to base64 URL-safe
    encoded := base64.URLEncoding.EncodeToString(b)
    
    // Remove padding and take 8 chars
    code := strings.TrimRight(encoded, "=")
    if len(code) > 8 {
        code = code[:8]
    }
    return code
}
```

**Code Generation Algorithm:**
1. **Random bytes**: `rand.Read()` fills slice with random data
2. **Base64 encode**: Converts bytes to string (URL-safe)
3. **Trim padding**: Remove `=` characters
4. **Take 8 chars**: Short, readable code

**Why base64?**
- URL-safe characters (no special chars)
- More characters than hex (64 vs 16)
- Shorter codes for same randomness

---

## 📁 File: `internal/handlers/url_handler.go` - HTTP Handlers

### Purpose
Handles HTTP requests. Parses input, calls services, returns responses.

### Code Breakdown

```go
type URLHandler struct {
    urlService *services.URLService
}
```

**Handler is Thin:**
- Only HTTP concerns
- Business logic in service
- Easy to test

```go
func (h *URLHandler) ShortenURL(c *gin.Context) {
    // 1. Parse request body
    var req ShortenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
```

**Gin Framework:**
- `c *gin.Context`: Request/response context
- `ShouldBindJSON()`: Parse JSON body into struct
- `c.JSON()`: Send JSON response
- `return`: Stop processing

```go
    // 2. Convert expires_in to Duration
    var expiresIn *time.Duration
    if req.ExpiresIn != nil {
        duration := time.Duration(*req.ExpiresIn) * time.Hour
        expiresIn = &duration
    }
```

**Type Conversion:**
- `*req.ExpiresIn`: Dereference pointer
- `time.Duration(...)`: Convert int to Duration
- `* time.Hour`: Multiply by hour constant
- `&duration`: Get pointer to result

```go
    // 3. Call service
    shortURL, err := h.urlService.ShortenURL(c.Request.Context(), req.URL, expiresIn)
    if err != nil {
        if err == services.ErrInvalidURL {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
        return
    }
```

**Error Handling:**
- Check specific errors
- Different HTTP status codes
- Don't expose internal errors to client

```go
    // 4. Format response
    var expiresAtStr *string
    if shortURL.ExpiresAt != nil {
        formatted := shortURL.ExpiresAt.Format(time.RFC3339)
        expiresAtStr = &formatted
    }
    
    response := ShortenResponse{
        ShortURL:    fmt.Sprintf("http://localhost:8080/%s", shortURL.ShortCode),
        ShortCode:   shortURL.ShortCode,
        OriginalURL: shortURL.OriginalURL,
        ExpiresAt:   expiresAtStr,
    }
    
    c.JSON(http.StatusOK, response)
}
```

**Response Formatting:**
- Convert time to string (RFC3339 format)
- Build full URL
- Return structured response

```go
func (h *URLHandler) RedirectURL(c *gin.Context) {
    shortCode := c.Param("code")
    if shortCode == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "short code is needed"})
        return
    }
    
    originalURL, err := h.urlService.GetOriginalURL(c.Request.Context(), shortCode)
    // ... error handling ...
    
    c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
```

**URL Parameters:**
- `c.Param("code")`: Get route parameter
- `c.Redirect()`: Send HTTP redirect
- `StatusTemporaryRedirect`: 307 status code

---

## 📁 File: `internal/middleware/middleware.go` - HTTP Middleware

### Purpose
Intercepts requests for logging, authentication, etc.

### Code Breakdown

```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        t := time.Now()  // Start time
        
        c.Next()  // Continue to next handler
        
        // After request completes
        latency := time.Since(t)
        status := c.Writer.Status()
        
        log.Printf("Path: %s | Status: %d | Latency: %v", 
            c.Request.URL.Path, status, latency)
    }
}
```

**Middleware Pattern:**
- Returns function that takes `*gin.Context`
- `c.Next()`: Calls next handler
- Code after `c.Next()` runs after handler

**Why Middleware?**
- Cross-cutting concerns (logging, auth)
- Reusable across routes
- Don't repeat code in handlers

---

## 🔄 Complete Request Flow Example

### Request: `POST /api/v1/shorten`

```
1. HTTP Request arrives
   ↓
2. Gin Router matches route
   POST /api/v1/shorten → urlHandler.ShortenURL
   ↓
3. Middleware executes
   Logger() → logs request start
   ↓
4. Handler processes
   Parse JSON body → ShortenRequest{URL: "https://example.com"}
   ↓
5. Service business logic
   Validate URL → Check existing → Generate code → Create ShortURL
   ↓
6. Repository saves
   Insert into MongoDB collection "short_urls"
   ↓
7. Response sent
   JSON: {short_url: "...", short_code: "ABC123", ...}
   ↓
8. Middleware logs
   Logs: Path, Status, Latency
```

### Data Flow Diagram

```
Client Request
    │
    ├─> [Gin Router] ──> Route: POST /api/v1/shorten
    │
    ├─> [Middleware] ──> Logger() ──> Start timer
    │
    ├─> [Handler] ──> Parse JSON ──> ShortenRequest
    │
    ├─> [Service] ──> URLService.ShortenURL()
    │   │
    │   ├─> Validate URL ──> isValidURL()
    │   │
    │   ├─> Check existing ──> Repository.GetShortURLByOriginal()
    │   │   └─> MongoDB Query: {original_url: "..."}
    │   │
    │   ├─> Generate code ──> KeyService.GetShortCode()
    │   │   ├─> Try Redis ──> LPop from queue
    │   │   └─> Generate locally ──> generateShortCode()
    │   │
    │   └─> Create & Save ──> Repository.CreateShortURL()
    │       └─> MongoDB Insert: {original_url, short_code, ...}
    │
    ├─> [Handler] ──> Format response ──> ShortenResponse
    │
    ├─> [Middleware] ──> Logger() ──> Log completion
    │
    └─> HTTP Response ──> JSON to client
```

---

## 🎯 Key Design Decisions Explained

### 1. Why Repository Pattern?

**Problem:** Service directly using MongoDB
```go
// ❌ Bad: Tight coupling
func (s *Service) Create(url string) {
    collection.InsertOne(...)  // Can't test, can't swap DB
}
```

**Solution:** Repository abstraction
```go
// ✅ Good: Loose coupling
type Repository interface {
    Create(url *ShortURL) error
}

func (s *Service) Create(url string) {
    s.repo.Create(...)  // Can inject mock for testing
}
```

### 2. Why Service Layer?

**Separation of Concerns:**
- **Handler**: HTTP (parsing, status codes)
- **Service**: Business logic (validation, rules)
- **Repository**: Data access (queries, saves)

**Benefits:**
- Test business logic without HTTP
- Reuse logic in different handlers
- Clear responsibilities

### 3. Why Context Everywhere?

**Context provides:**
- **Cancellation**: Cancel long operations
- **Timeouts**: Auto-cancel after time
- **Request ID**: Track requests
- **Deadlines**: Enforce SLAs

**Example:**
```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// All operations use ctx
repo.Get(ctx, id)  // Will cancel after 5 seconds
```

### 4. Why Pointers?

**Performance:**
```go
// Value: Copies entire struct (slow for large structs)
func process(url ShortURL) { ... }

// Pointer: Shares reference (fast)
func process(url *ShortURL) { ... }
```

**Modification:**
```go
// Value: Can't modify original
func update(url ShortURL) {
    url.ClickCount++  // Only modifies copy
}

// Pointer: Modifies original
func update(url *ShortURL) {
    url.ClickCount++  // Modifies original
}
```

---

## 📚 Go Language Concepts Summary

### 1. **Packages**
- Group related code
- `package main`: Executable
- Other packages: Libraries

### 2. **Imports**
- Import what you use
- Path-based: `github.com/user/repo`
- Aliases: `import redis "github.com/..."`

### 3. **Types**
- `type Name struct { ... }`: Custom type
- `type Name interface { ... }`: Interface
- `type Name = Other`: Type alias

### 4. **Functions**
- `func Name() returnType`: Function
- `func (r Receiver) Name()`: Method
- Multiple returns: `func() (int, error)`

### 5. **Error Handling**
- Always return errors
- Check explicitly: `if err != nil`
- Wrap with context: `fmt.Errorf("...: %w", err)`

### 6. **Concurrency**
- `go func()`: Goroutine
- `chan Type`: Channel
- `select`: Choose from channels

### 7. **Interfaces**
- Implicit implementation
- Duck typing: "If it quacks like a duck..."
- Enables polymorphism

---

## 🚀 Next Steps for Learning

1. **Read the code**: Start with `main.go`, follow the flow
2. **Run it**: See it in action
3. **Modify it**: Add features, break things, fix them
4. **Test it**: Write unit tests
5. **Profile it**: Measure performance
6. **Deploy it**: Put it in production

This codebase demonstrates:
- ✅ Clean Architecture
- ✅ Dependency Injection
- ✅ Error Handling
- ✅ Context Usage
- ✅ Concurrency
- ✅ Testing Patterns
- ✅ Production Practices

Happy coding! 🎉

