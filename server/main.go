package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/raaz/server/db"
	"github.com/raaz/server/store"
)

func main() {
	initLogger()

	app := buildApp()
	app.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("raaz server starting", "port", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(app)); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

// initLogger configures slog. When LOG_FORMAT=json (e.g. inside Docker),
// emits structured JSON; otherwise uses the human-readable text handler.
func initLogger() {
	if os.Getenv("LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	// else: default text handler — clear for local dev
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
		slog.Warn("postgres unavailable, falling back to in-memory", "err", err)
		return NewApp()
	}

	rdb, err := db.NewClient()
	if err != nil {
		slog.Warn("redis client error, falling back to in-memory", "err", err)
		pool.Close()
		return NewApp()
	}

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Warn("redis unavailable, falling back to in-memory", "err", err)
		pool.Close()
		return NewApp()
	}

	slog.Info("postgres ready")
	slog.Info("redis ready")

	globalStrikes = store.NewPGStrikeTracker(pool, rdb)

	return &App{
		hub:      NewHub(),
		queue:    store.NewRedisQueueStore(rdb),
		sessions: store.NewRedisSessionStore(rdb),
		vault:    store.NewPGVaultStore(pool),
	}
}
