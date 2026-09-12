// Package main provides the entry point for the groceries application.
// It initializes logging, database connections, and the HTTP server with graceful shutdown handling.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/api"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/db"

	"github.com/taiidani/groceries/internal/server"
)

func main() {
	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	// Set up logging. Records are wrapped so that any log emitted with an
	// active span context is annotated with trace_id/span_id for correlation
	// with traces in the observability backend.
	initLogging()

	// Set up the Redis/Memory database
	rds := cache.NewClient(ctx)

	// Set up the relational database
	conn, err := db.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.ErrorContext(ctx, "could not connect to database", "err", err)
		os.Exit(2)
	}
	defer conn.Close()

	// Start the instances
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start the web UI
		if err := initServer(ctx, conn, rds); err != nil {
			slog.ErrorContext(ctx, "fatal server error", "err", err)
			os.Exit(1)
		}
	}()

	wg.Wait()

	slog.Info("Shutdown successful")
}

func initLogging() {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	// Dev mode emits human-readable text logs; otherwise structured JSON is
	// emitted for ingestion by a log backend.
	var handler slog.Handler
	if os.Getenv("DEV") == "true" {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))
}

func initServer(ctx context.Context, conn *sql.DB, rds *redis.Client) error {
	port := os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("required PORT environment variable not present")
	}

	oidcCfg := server.OIDCConfig{
		IssuerURL:    os.Getenv("OIDC_ISSUER_URL"),
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		BaseURL:      os.Getenv("URL"),
	}
	if oidcCfg.IssuerURL == "" || oidcCfg.ClientID == "" || oidcCfg.ClientSecret == "" || oidcCfg.BaseURL == "" {
		return fmt.Errorf("required OIDC_ISSUER_URL, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and URL environment variables must all be present")
	}

	// The web server owns the mux. The API server registers its routes onto
	// the same mux so both share a single listener and connection pool.
	mux := http.NewServeMux()
	if _, err := api.NewServer(ctx, conn, rds, mux, oidcCfg.IssuerURL); err != nil {
		return fmt.Errorf("could not initialize API server: %w", err)
	}
	srv, err := server.NewServer(ctx, conn, rds, port, mux, oidcCfg)
	if err != nil {
		return fmt.Errorf("could not initialize web server: %w", err)
	}

	go func() {
		slog.Info("Server starting", "port", port)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Unclean server shutdown encountered", "error", err)
		}
	}()

	<-ctx.Done()

	// Gracefully shut down over 60 seconds
	slog.Info("Server shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("Server shutdown successful")
	return nil
}
