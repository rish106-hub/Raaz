package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/raaz/server/db"
	"github.com/raaz/server/store"
)

func main() {
	app := buildApp()
	app.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("raaz server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(app)))
}

// buildApp selects in-memory or production (PostgreSQL + Redis) store impls
// based on whether DATABASE_URL and REDIS_URL are set. Always falls back to
// in-memory so tests pass without external services.
func buildApp() *App {
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" || redisURL == "" {
		return NewApp()
	}

	pool, err := db.NewPool(context.Background())
	if err != nil {
		log.Printf("postgres unavailable, falling back to in-memory: %v", err)
		return NewApp()
	}

	rdb, err := db.NewClient()
	if err != nil {
		log.Printf("redis client error, falling back to in-memory: %v", err)
		pool.Close()
		return NewApp()
	}

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis unavailable, falling back to in-memory: %v", err)
		pool.Close()
		return NewApp()
	}

	log.Println("postgres ready")
	log.Println("redis ready")

	globalStrikes = store.NewPGStrikeTracker(pool, rdb)

	return &App{
		hub:      NewHub(),
		queue:    store.NewRedisQueueStore(rdb),
		sessions: store.NewRedisSessionStore(rdb),
		vault:    store.NewPGVaultStore(pool),
	}
}
