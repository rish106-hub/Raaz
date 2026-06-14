# Raaz Project Setup Guide

## ✅ What's Installed

- **Go 1.26.4** (darwin/arm64)
- **Docker 29.5.3**
- **Go Dependencies** (`go mod tidy` completed)
- **Built Server Binary** (`./raaz-server`)

## 🚀 Quick Start

### Option 1: Local Development (In-Memory Mode)
```bash
# Start the server without external dependencies
./raaz-server

# Server runs at http://localhost:8080
```

### Option 2: Full Stack with Docker Compose (PostgreSQL + Redis)
```bash
# Start Docker daemon first (if not running)
# From Spotlight: search "Docker" and launch the app

# Then run:
docker-compose up

# Server runs at http://localhost:8080
# PostgreSQL: localhost:5432
# Redis: localhost:6379
```

## 📁 Project Structure

```
raaz/
├── server/                 # Go backend
│   ├── main.go            # Entry point
│   ├── app.go             # Core application logic
│   ├── router.go          # HTTP routes
│   ├── hub.go             # WebSocket connection manager
│   ├── matching.go        # Matching algorithm
│   ├── session.go         # Session management
│   ├── message.go         # Message handling
│   ├── moderation.go      # Content moderation
│   ├── db/                # Database clients
│   │   ├── postgres.go    # PostgreSQL connection pool
│   │   └── redis.go       # Redis client
│   └── store/             # Data storage implementations
│       ├── pg_*.go        # PostgreSQL stores
│       ├── redis_*.go     # Redis stores
│       └── memory_*.go    # In-memory stores
├── app/                   # Android app (Kotlin)
├── docker-compose.yml     # Services: PostgreSQL, Redis, Server
├── .env                   # Environment variables (created)
└── Dockerfile             # Server container image
```

## 🔧 Environment Variables

File: `.env`

| Variable | Purpose | Default |
|----------|---------|---------|
| `DATABASE_URL` | PostgreSQL connection | `postgres://raaz:changeme@postgres:5432/raaz?sslmode=disable` |
| `REDIS_URL` | Redis connection | `redis://redis:6379` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `changeme` |
| `PORT` | Server port | `8080` |
| `LOG_FORMAT` | Log output (`json` or default text) | `json` |

## 🏗️ Architecture

### Frontend
- Android (primary)
- iOS

### Backend (Go Microservices)
- **Core API** (`app.go`, `router.go`)
- **Matching Engine** (`matching.go`)
- **WebSocket Hub** (`hub.go`)
- **Moderation** (`moderation.go`)

### Data Layer
- **PostgreSQL** - Persistent data (users, sessions, history)
- **Redis** - Session state, queue, cache
- **In-Memory** - Fallback if DB unavailable

### Key Features
✅ Anonymous conversations  
✅ Smart matching (city + age bucket)  
✅ Real-time messaging (WebSocket)  
✅ Moderation & abuse detection  
✅ Strike system (3 strikes = ban)  
✅ Session timeouts (20 min + optional extension)  
✅ Ephemeral content (48-hour deletion)  

## 🧪 Running Tests

```bash
cd server
go test ./... -v                    # Run all tests
go test ./... -run TestName         # Run specific test
go test -race ./...                # Run with race detector
```

## 🐛 Debugging

### View Go server logs
```bash
# Local mode
./raaz-server

# Docker mode
docker-compose logs -f server
```

### Check database
```bash
# Connect to PostgreSQL (if Docker running)
docker exec -it raaz-postgres-1 psql -U raaz -d raaz

# Connect to Redis
docker exec -it raaz-redis-1 redis-cli
```

## 📦 Building & Deployment

### Rebuild binary
```bash
cd server
go build -o ../raaz-server .
```

### Build Docker image
```bash
docker build -t raaz-server:latest ./server
```

### Development mode with hot reload
```bash
# Install entr for file watching
brew install entr

# Watch and rebuild on changes
ls server/*.go | entr -r go run ./server
```

## 🔐 Security Notes

- **Production**: Use real passwords, enable SSL/TLS
- **Data**: Conversations auto-delete after 48 hours
- **Secrets**: Never commit `.env` (already in `.gitignore`)
- **Authentication**: Implement JWT/OAuth before launch

## 📊 API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/ws` | WebSocket upgrade |
| `GET` | `/health` | Health check |
| `POST` | `/sessions` | Create session |
| `GET` | `/sessions/:id` | Get session |

## 🚧 Common Issues

| Issue | Solution |
|-------|----------|
| Server starts but can't connect via WebSocket | Check WebSocket URL format: `ws://localhost:8080/ws?...` |
| Docker containers crash | Run `docker-compose logs` to see error |
| PostgreSQL won't connect | Ensure `DATABASE_URL` is correct in `.env` |
| Port 8080 already in use | Change `PORT` in `.env` or kill existing process |

## 📚 Next Steps

1. **Install Docker Desktop** if you want full stack (PostgreSQL + Redis)
2. **Review architecture** in `server/app.go`
3. **Run tests** to understand behavior
4. **Explore database models** in `server/store/`
5. **Set up IDE** (VSCode with Go extension recommended)

## 💡 Development Tips

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Get dependencies graph
go mod graph

# Update dependencies
go get -u ./...
```

---

**Happy coding! 🚀**
