package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wadeling/origin-check/internal/api"
	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/crypto"
	"github.com/wadeling/origin-check/internal/relay"
	"github.com/wadeling/origin-check/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect store", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	enc, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		slog.Error("init encryptor", "error", err)
		os.Exit(1)
	}

	if err := relay.Seed(ctx, st, enc, relay.DefaultSeeds()); err != nil {
		slog.Warn("seed relays", "error", err)
	}

	srv := api.NewServer(st)
	handler := srv.Router(cfg.CORSOrigin)

	httpServer := &http.Server{
		Addr:    cfg.APIAddr,
		Handler: handler,
	}

	go func() {
		slog.Info("api listening", "addr", cfg.APIAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api server", "error", err)
			os.Exit(1)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
