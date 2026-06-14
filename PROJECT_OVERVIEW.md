# Raaz - Project Setup Complete ✅

## What You Have

A fully set up **Go backend** for **Raaz**, an anonymous conversation platform built for Indian Gen Z.

### Tech Stack
- **Language:** Go 1.26.4
- **Runtime:** Docker 29.5.3 + Docker Compose
- **Databases:** PostgreSQL (persistent data) + Redis (session state)
- **Protocol:** WebSockets (real-time messaging)
- **Architecture:** Microservices (modular Go packages)

---

## The Raaz Platform

**Tagline:** *Some conversations are too real for Instagram.*

### What It Does
- Users receive one emotional prompt daily at 8 AM IST
- They enter a matching pool to be paired with another user anonymously
- They have a 20-minute conversation window (with 10-min extension option)
- Messages are ephemeral (auto-deleted after 48 hours)
- Zero profile data exposed - just real, vulnerable conversations

### Key Features
✅ **Anonymous** - No names, photos, or profiles during chat  
✅ **Smart Matching** - City + age bracket pairing  
✅ **Ephemeral** - Conversations auto-delete after 48 hours  
✅ **Moderation** - Real-time abuse detection & strike system  
✅ **Safe** - Crisis support resources for vulnerable users  
✅ **Freemium** - Free tier + premium features (₹99-₹199/month)  

---

## Files You Need to Know

### Documentation
| File | Purpose |
|------|---------|
| `QUICKSTART.txt` | **READ THIS FIRST** - Quick reference for running the project |
| `SETUP.md` | Comprehensive setup guide with troubleshooting |
| `README.md` | Product overview (what is Raaz?) |
| `PRD.md` | Detailed product requirements |
| `Makefile` | All build/test/deploy commands |

### Source Code
| File | Purpose |
|------|---------|
| `server/main.go` | Entry point - starts the HTTP server |
| `server/app.go` | Core application logic (150+ lines) |
| `server/router.go` | HTTP routes (WebSocket upgrade, health check) |
| `server/hub.go` | WebSocket connection manager |
| `server/matching.go` | Smart matching algorithm |
| `server/session.go` | Session state management |
| `server/message.go` | Message routing logic |
| `server/moderation.go` | Abuse detection & enforcement |
| `server/strikes.go` | Strike tracking system |
| `server/db/` | Database clients (PostgreSQL, Redis) |
| `server/store/` | Data layer (interfaces + implementations) |

### Configuration
| File | Purpose |
|------|---------|
| `.env` | Environment variables (created) |
| `docker-compose.yml` | Docker services (PostgreSQL, Redis, Server) |
| `server/Dockerfile` | Container image for the Go server |
| `go.mod` | Go module definition & dependencies |

---

## Getting Started (3 Steps)

### 1️⃣ Quick Start (In-Memory Mode)
No database needed - perfect for testing:
```bash
make run
```
Server runs at `http://localhost:8080` with in-memory storage.

### 2️⃣ Full Stack (With Database)
Production-like setup:
```bash
make docker-up
```
- PostgreSQL on localhost:5432
- Redis on localhost:6379
- Server on http://localhost:8080

### 3️⃣ Explore the Code
```bash
# Format & lint
make fmt
make lint

# Run tests
make test

# View logs
make docker-logs
```

---

## API Endpoints

### WebSocket
```
GET /ws?promptId=<id>&ageBucket=<age>&city=<city>&anonymousId=<id>
```
Upgrade to WebSocket for real-time messaging.

### HTTP
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Health check |
| POST | `/sessions` | Create session |
| GET | `/sessions/:id` | Get session details |

### Message Format (WebSocket)
```json
{
  "type": "MESSAGE",
  "content": "Hey, how are you?"
}
```

Types: `MESSAGE`, `CONNECTED`, `TYPING`, `DISCONNECT`, `SESSION_ENDED`, etc.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Raaz Backend (Go)                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ HTTP Router (router.go)                              │   │
│  │ • POST /sessions                                     │   │
│  │ • GET /ws (WebSocket upgrade)                        │   │
│  │ • GET /health                                        │   │
│  └──────────────────────────────────────────────────────┘   │
│           ↓ upgrades to                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ WebSocket Hub (hub.go)                               │   │
│  │ • Connection manager                                 │   │
│  │ • Broadcast messages                                 │   │
│  └──────────────────────────────────────────────────────┘   │
│           ↓ calls                                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ App Logic (app.go, matching.go, session.go)          │   │
│  │ • Smart matching algorithm                           │   │
│  │ • Session management                                 │   │
│  │ • Moderation & strikes                               │   │
│  └──────────────────────────────────────────────────────┘   │
│           ↓ uses                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Data Layer (store/ & db/)                            │   │
│  │ • PostgreSQL: persistent data                        │   │
│  │ • Redis: session state & cache                       │   │
│  │ • In-Memory: fallback                                │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Concepts

### Matching Algorithm
- Users enter a pool with their city, age bracket, and prompt
- After 15 seconds, if no local match, goes national
- Matched pairs are connected via WebSocket
- Session created with 20-min timeout + optional 10-min extension

### Message Flow
1. User connects via WebSocket with anonymousId
2. User enters pool by sending REGISTER message
3. Matching engine finds partner
4. CONNECTED message sent to both
5. Messages routed through hub
6. DISCONNECT/SESSION_ENDED on timeout or manual exit

### Moderation
- Real-time abuse detection (keywords, patterns)
- Strike system: 1 warn → 2: 7-day ban → 3: permanent ban
- Crisis resources triggered on self-harm indicators

---

## Development Workflow

```bash
# Make changes to server/
cd server
vim app.go

# Format & lint
make fmt
make lint

# Test
make test

# Rebuild
make build

# Run
./raaz-server
# or
make run
```

---

## Production Checklist

Before deploying:
- [ ] Use real PostgreSQL credentials (not `changeme`)
- [ ] Enable TLS/SSL (`sslmode=require`)
- [ ] Set up Firebase Cloud Messaging for push notifications
- [ ] Configure Redis persistence
- [ ] Add rate limiting middleware
- [ ] Implement JWT authentication
- [ ] Set up monitoring & alerting
- [ ] Enable encrypted logs
- [ ] GDPR/DPDP compliance review

---

## Common Tasks

### Add a new endpoint
1. Edit `server/router.go` - add route
2. Edit `server/app.go` - add handler
3. Update tests in `server/app_test.go`

### Add a new database query
1. Create interface in `server/store/interfaces.go`
2. Implement for PostgreSQL in `server/store/pg_*.go`
3. Implement in-memory version in `server/store/memory_*.go`
4. Use in `server/app.go`

### Debug WebSocket issue
```bash
# Check connection logs
make docker-logs | grep -i websocket

# Test locally
./raaz-server  # runs on :8080

# Use wscat tool
npm install -g wscat
wscat -c 'ws://localhost:8080/ws?promptId=p1&ageBucket=18-25&city=Bengaluru&anonymousId=test'
```

---

## What's Next?

1. **Read** `QUICKSTART.txt` for immediate next steps
2. **Explore** the codebase starting with `server/app.go`
3. **Run** `make docker-up` to see it in action
4. **Test** WebSocket connections with a simple client
5. **Review** the moderation & strike logic in `server/moderation.go`

---

## Support

- 📚 Documentation: See files in project root
- 🐛 Debug: `make docker-logs` for full output
- 🧪 Tests: `make test` to verify everything works
- 📖 Code: Well-commented in `server/*.go`

---

**You're all set! Happy coding! 🚀**
