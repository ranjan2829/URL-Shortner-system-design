# URL Shortener System Design

A full-stack URL shortener application with Go backend and Next.js frontend.

## 🚀 Features

- **URL Shortening**: Convert long URLs into short, shareable links
- **Key Generation**: Integrated key generation service (no external dependencies)
- **Analytics**: Track click counts and view statistics
- **Expiration**: Optional URL expiration time
- **Modern UI**: Beautiful, responsive frontend built with Next.js and Tailwind CSS
- **Scalable Backend**: Go-based REST API with MongoDB and Redis

## 📁 Project Structure

```
URL-Shortner-system-design/
├── backend/                 # Go backend service
│   ├── cmd/
│   │   └── server/         # Main server application
│   ├── internal/
│   │   ├── config/        # Configuration management
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── models/         # Data models
│   │   ├── repository/     # Database repository layer
│   │   └── services/      # Business logic services
│   └── go.mod
├── frontend/               # Next.js frontend
│   ├── app/               # Next.js app directory
│   └── package.json
└── README.md
```

## 🛠️ Tech Stack

### Backend
- **Go** - Programming language
- **Gin** - Web framework
- **MongoDB** - Database for storing URLs
- **Redis** - Caching and queue management
- **Docker** - MongoDB containerization

### Frontend
- **Next.js 16** - React framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **React 19** - UI library

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (for MongoDB)
- Redis

### Backend Setup

1. **Start MongoDB** (using Docker):
   ```bash
   docker run -d --name mongodb -p 27017:27017 -e MONGO_INITDB_DATABASE=url_shortener mongo:latest
   ```

2. **Start Redis**:
   ```bash
   brew services start redis
   # or
   redis-server
   ```

3. **Navigate to backend directory**:
   ```bash
   cd backend
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Run the server**:
   ```bash
   go run ./cmd/server
   ```

   Server will start on `http://localhost:8080`

### Frontend Setup

1. **Navigate to frontend directory**:
   ```bash
   cd frontend
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Create `.env.local` file** (optional):
   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```

4. **Run the development server**:
   ```bash
   npm run dev
   ```

   Frontend will start on `http://localhost:3000`

## 📡 API Endpoints

### POST `/api/v1/shorten`
Shorten a URL.

**Request:**
```json
{
  "url": "https://example.com/very/long/url",
  "expires_in": 24
}
```

**Response:**
```json
{
  "short_url": "http://localhost:8080/ABC123",
  "short_code": "ABC123",
  "original_url": "https://example.com/very/long/url",
  "expires_at": "2025-12-02T09:00:00Z"
}
```

### GET `/api/v1/generate`
Generate a new short code.

**Response:**
```json
{
  "short_code": "XYZ789"
}
```

### GET `/:code`
Redirect to the original URL.

### GET `/api/v1/:code/stats`
Get statistics for a short URL.

**Response:**
```json
{
  "id": "...",
  "original_url": "https://example.com",
  "short_code": "ABC123",
  "created_at": "2025-12-01T09:00:00Z",
  "click_count": 42,
  "is_active": true
}
```

## 🗄️ Database

### MongoDB Collections

- **short_urls**: Stores all shortened URLs
  - `_id`: ObjectId
  - `original_url`: string
  - `short_code`: string (unique, indexed)
  - `created_at`: timestamp
  - `expires_at`: timestamp (optional)
  - `click_count`: int64
  - `is_active`: boolean

### Viewing Data

Connect to MongoDB:
```bash
docker exec -it mongodb mongosh url_shortener
```

Query data:
```javascript
db.short_urls.find().pretty()
db.short_urls.countDocuments()
```

## 🧪 Testing

### Test URL Shortening
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

### Test Key Generation
```bash
curl http://localhost:8080/api/v1/generate
```

## 📝 Environment Variables

### Backend
- `PORT` - Server port (default: 8080)
- `MONGODB_URI` - MongoDB connection string (default: mongodb://localhost:27017)
- `MONGODB_DB` - Database name (default: url_shortener)
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `REDIS_PASSWORD` - Redis password (optional)
- `KEY_GEN_SERVICE_URL` - Key generation service URL (not needed - integrated)

### Frontend
- `NEXT_PUBLIC_API_URL` - Backend API URL (default: http://localhost:8080)

## 🎯 Features in Detail

### Key Generation
- Integrated into main server (no separate service)
- Falls back to local generation if Redis queue is empty
- Generates 8-character base64 URL-safe codes

### URL Shortening
- Validates URL format
- Checks for existing URLs (returns existing if found)
- Supports optional expiration time
- Tracks click counts

### Analytics
- View click statistics
- Check creation date
- Monitor active status

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

This project is open source and available under the MIT License.

## 👨‍💻 Author

Built as a system design project demonstrating:
- Microservices architecture
- Database design
- API design
- Full-stack development
