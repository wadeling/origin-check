package sched

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/store"
)

type Scheduler struct {
	store *store.Store
	queue *queue.Queue
	cron  *cron.Cron
}

func New(st *store.Store, q *queue.Queue) *Scheduler {
	return &Scheduler{
		store: st,
		queue: q,
		cron:  cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	_, _ = s.cron.AddFunc("0 */15 * * * *", func() { s.scheduleHealth(ctx) })
	_, _ = s.cron.AddFunc("0 0 */6 * * *", func() { s.schedulePerformance(ctx) })
	_, _ = s.cron.AddFunc("0 0 3 * * *", func() { s.scheduleAuthenticity(ctx) })

	s.cron.Start()
	slog.Info("scheduler started")

	// Run initial jobs on startup
	go s.scheduleHealth(ctx)
	go s.schedulePerformance(ctx)

	<-ctx.Done()
	s.cron.Stop()
	return ctx.Err()
}

func (s *Scheduler) scheduleHealth(ctx context.Context) {
	s.scheduleForAll(ctx, store.JobHealth, func(r store.Relay) string {
		return r.HealthModel
	})
}

func (s *Scheduler) schedulePerformance(ctx context.Context) {
	s.scheduleForAll(ctx, store.JobPerformance, func(r store.Relay) string {
		if len(r.ClaimedModels) > 0 {
			return r.ClaimedModels[0]
		}
		return r.HealthModel
	})
}

func (s *Scheduler) scheduleAuthenticity(ctx context.Context) {
	s.scheduleForAll(ctx, store.JobAuthenticity, func(r store.Relay) string {
		if len(r.ClaimedModels) > 0 {
			return r.ClaimedModels[0]
		}
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
		job := &store.ProbeJob{
			RelayID:     relay.ID,
			JobType:     jobType,
			Model:       &model,
			ScheduledAt: time.Now(),
			Status:      store.JobPending,
		}
		if err := s.store.CreateJob(ctx, job); err != nil {
			slog.Error("create job failed", "relay", relay.Name, "error", err)
			continue
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
}
