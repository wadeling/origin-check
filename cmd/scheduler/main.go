package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/crypto"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/relay"
	"github.com/wadeling/origin-check/internal/sched"
	"github.com/wadeling/origin-check/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect store", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	q, err := queue.New(cfg.RedisURL)
	if err != nil {
		slog.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer q.Close()

	enc, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		slog.Error("init encryptor", "error", err)
		os.Exit(1)
	}

	if err := relay.Seed(ctx, st, enc, relay.DefaultSeeds()); err != nil {
		slog.Warn("seed relays", "error", err)
	}

	scheduleCfg, err := config.LoadScheduleConfig()
	if err != nil {
		slog.Error("load schedule config", "error", err)
		os.Exit(1)
	}

	scheduler := sched.New(st, q, scheduleCfg)
	if err := scheduler.Start(ctx); err != nil && err != context.Canceled {
		slog.Error("scheduler stopped", "error", err)
		os.Exit(1)
	}
}
