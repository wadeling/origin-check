package sched

import (
	"context"
	"log/slog"
	"time"

	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/store"
)

type Scheduler struct {
	store *store.Store
	queue *queue.Queue
	cfg   config.ScheduleConfig
}

func New(st *store.Store, q *queue.Queue, cfg config.ScheduleConfig) *Scheduler {
	return &Scheduler{
		store: st,
		queue: q,
		cfg:   cfg,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	slog.Info("scheduler started",
		"health_interval", s.cfg.HealthInterval,
		"performance_interval", s.cfg.PerformanceInterval,
		"authenticity_interval", s.cfg.AuthenticityInterval,
	)

	s.startLoop(ctx, "health", s.cfg.HealthInterval, s.cfg.HealthOnStartup, s.scheduleHealth)
	s.startLoop(ctx, "performance", s.cfg.PerformanceInterval, s.cfg.PerformanceOnStartup, s.schedulePerformance)
	s.startLoop(ctx, "authenticity", s.cfg.AuthenticityInterval, s.cfg.AuthenticityOnStartup, s.scheduleAuthenticity)

	<-ctx.Done()
	return ctx.Err()
}

func (s *Scheduler) startLoop(ctx context.Context, name string, interval time.Duration, onStartup bool, fn func(context.Context)) {
	run := func(reason string) {
		slog.Info("probe tick", "type", name, "reason", reason)
		fn(ctx)
	}

	if onStartup {
		go run("startup")
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run("interval")
			}
		}
	}()
}

func (s *Scheduler) scheduleHealth(ctx context.Context) {
	s.scheduleForAll(ctx, store.JobHealth, func(r store.Relay) string {
		return r.HealthModel
	})
}

func (s *Scheduler) scheduleForAll(ctx context.Context, jobType store.JobType, modelFn func(store.Relay) string) {
	relays, err := s.store.ListActiveRelays(ctx)
	if err != nil {
		slog.Error("list relays failed", "error", err)
		return
	}

	for _, relay := range relays {
		model := modelFn(relay)
		s.enqueueJob(ctx, relay, jobType, model)
	}
}

func (s *Scheduler) schedulePerformance(ctx context.Context) {
	s.scheduleForAllModels(ctx, store.JobPerformance)
}

func (s *Scheduler) scheduleAuthenticity(ctx context.Context) {
	s.scheduleForAllModels(ctx, store.JobAuthenticity)
}

func (s *Scheduler) scheduleForAllModels(ctx context.Context, jobType store.JobType) {
	relays, err := s.store.ListActiveRelays(ctx)
	if err != nil {
		slog.Error("list relays failed", "error", err)
		return
	}

	for _, relay := range relays {
		models := relay.ClaimedModels
		if len(models) == 0 {
			models = []string{relay.HealthModel}
		}
		for _, model := range models {
			s.enqueueJob(ctx, relay, jobType, model)
		}
	}
}

func (s *Scheduler) enqueueJob(ctx context.Context, relay store.Relay, jobType store.JobType, model string) {
	job := &store.ProbeJob{
		RelayID:     relay.ID,
		JobType:     jobType,
		Model:       &model,
		ScheduledAt: time.Now(),
		Status:      store.JobPending,
	}
	if err := s.store.CreateJob(ctx, job); err != nil {
		slog.Error("create job failed", "relay", relay.Name, "error", err)
		return
	}
	if err := s.queue.Enqueue(ctx, queue.JobPayload{
		JobID:   job.ID,
		RelayID: relay.ID,
		Type:    jobType,
		Model:   model,
	}); err != nil {
		slog.Error("enqueue failed", "relay", relay.Name, "error", err)
	} else {
		slog.Info("job scheduled", "relay", relay.Name, "type", jobType, "model", model)
	}
}
