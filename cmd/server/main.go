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
	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/db"
	"github.com/taiidani/groceries/internal/server"
)

func main() {
	// Handle signal interrupts.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	teardown, err := initSentry()
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// Flush buffered Sentry events before the program terminates.
	defer teardown()

	// Set up the Redis/Memory database
	db := db.NewDB()

	// Start the instances
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start the web UI
		if err := initServer(ctx, db); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	fmt.Println("Shutdown successful")
}

func initSentry() (func(), error) {
	// Set up Sentry
	err := sentry.Init(sentry.ClientOptions{
		SampleRate:       1.0,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return func() {}, err
	}

	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

func initServer(ctx context.Context, db data.DB) error {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Required PORT environment variable not present")
	}

	srv := server.NewServer(db, port)

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
