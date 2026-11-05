package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freyliee/template/internal/config"
	api "github.com/freyliee/template/internal/generated"
	"github.com/freyliee/template/internal/handler"
	"github.com/freyliee/template/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	var level slog.Level

	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		log.Fatal("invalid LOG_LEVEL")
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	)))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)

	pool, err := storage.NewPostgresPool(ctx, cfg)
	cancel()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")

	trips := storage.NewTripRepository(pool)

	txManager := storage.NewTxManager(pool, cfg.DatabaseQueryTimeout)

	h := handler.New(pool, cfg.DatabaseQueryTimeout, trips, txManager)
	router := api.HandlerWithOptions(h, api.ChiServerOptions{
		ErrorHandlerFunc: handler.HandleAPIError,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	exitCode := 0

	select {
	case <-signalCtx.Done():
		log.Println("Shutdown signal received")

		stop()

	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server failed: %v", err)
			exitCode = 1
		}
	}

	deadline := time.Now().Add(cfg.ShutdownTimeout)

	shutdownCtx, cancelShutdown := context.WithDeadline(
		context.Background(),
		deadline,
	)
	defer cancelShutdown()

	shutdownDeadline := time.AfterFunc(
		time.Until(deadline),
		func() {
			log.Println("Shutdown timeout exceeded, forcing exit")
			os.Exit(1)
		},
	)

	log.Println("Stopping HTTP server")

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown failed: %v", err)
		_ = server.Close()
		exitCode = 1
	}

	log.Println("Closing PostgreSQL pool")
	pool.Close()

	shutdownDeadline.Stop()

	log.Println("Service stopped")

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
