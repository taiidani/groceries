// Package main provides the entry point for the groceries application.
// It initializes logging, database connections, and the HTTP server with graceful shutdown handling.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/cache"
	"github.com/taiidani/groceries/internal/models"
	"github.com/taiidani/groceries/internal/server"
)

func main() {
	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	// Set up Sentry
	err := sentry.Init(sentry.ClientOptions{
		SampleRate:       1.0,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		EnableLogs:       true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Set up the structured logger
	initLogging(ctx)

	// Set up the Redis/Memory database
	rds := cache.NewClient(ctx)

	// Set up the relational database
	err = models.InitDB(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalf("database init: %s", err)
	}

	// Start the instances
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start the web UI
		if err := initServer(ctx, rds); err != nil {
			sentry.CaptureException(err)
			log.Fatal(err)
		}
	}()

	wg.Wait()

	slog.Info("Shutdown successful")
}

func initLogging(ctx context.Context) {
	var logger *slog.Logger

	switch os.Getenv("SENTRY_ENVIRONMENT") {
	case "prod", "production":
		handler := sentryslog.Option{
			// Explicitly specify the levels that you want to be captured.
			EventLevel: []slog.Level{slog.LevelError},                                 // Captures only [slog.LevelError] as error events.
			LogLevel:   []slog.Level{slog.LevelWarn, slog.LevelInfo, slog.LevelDebug}, // Captures remaining items as log entries.
		}.NewSentryHandler(ctx)
		logger = slog.New(handler)
	default:
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

		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
		logger = slog.New(handler)
	}

	slog.SetDefault(logger)
}

func initServer(ctx context.Context, rds *redis.Client) error {
	port := os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("required PORT environment variable not present")
	}

	srv := server.NewServer(ctx, rds, port)

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
