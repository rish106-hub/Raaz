package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raaz/server/db"
	"github.com/raaz/server/store"
)

func main() {
	initLogger()

	app, cleanupStores := buildApp()
	app.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(app),
	}

	// L-8: graceful shutdown on SIGTERM/SIGINT
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		<-ctx.Done()
		slog.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown error", "err", err)
		}
	}()

	slog.Info("raaz server starting", "port", port)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server exited", "err", err)
		cleanupStores()
		os.Exit(1)
	}

	// L-6: close DB pool and Redis client after server drains
	cleanupStores()
	slog.Info("server stopped")
}

// initLogger configures slog. LOG_FORMAT=json emits structured JSON (Docker/prod);
// the default text handler is used in local dev.
func initLogger() {
	if os.Getenv("LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
}

// buildApp selects in-memory or production (PostgreSQL + Redis) store impls based on env vars.
// Returns the App and a cleanup function that must be called after the server stops (L-6).
func buildApp() (*App, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" || redisURL == "" {
		return NewApp(), func() {}
	}

	pool, err := db.NewPool(context.Background())
	if err != nil {
		slog.Warn("postgres unavailable, falling back to in-memory", "err", err)
		return NewApp(), func() {}
	}

	rdb, err := db.NewClient()
	if err != nil {
		slog.Warn("redis client error, falling back to in-memory", "err", err)
		pool.Close()
		return NewApp(), func() {}
	}

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Warn("redis unavailable, falling back to in-memory", "err", err)
		pool.Close()
		rdb.Close()
		return NewApp(), func() {}
	}

	slog.Info("postgres ready")
	slog.Info("redis ready")

	globalStrikes = store.NewPGStrikeTracker(pool, rdb)

	app := &App{
		hub:          NewHub(),
		queue:        store.NewRedisQueueStore(rdb),
		sessions:     store.NewRedisSessionStore(rdb),
		vault:        store.NewPGVaultStore(pool),
		matchHistory: store.NewPGMatchHistoryStore(pool),
		prompts:      store.NewPGPromptStore(pool),
		freemium:     store.NewPGFreemiumStore(pool),
	}

	cleanup := func() {
		if err := rdb.Close(); err != nil {
			slog.Warn("redis close error", "err", err)
		}
		pool.Close()
		slog.Info("database connections closed")
	}

	return app, cleanup
}
