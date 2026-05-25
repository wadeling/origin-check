package trigger

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/store"
)

type Options struct {
	Relay     store.Relay
	JobType   store.JobType
	Models    []string // empty = defaults for job type
}

type EnqueuedJob struct {
	JobID uuid.UUID
	Model string
}

func ModelsForJob(relay store.Relay, jobType store.JobType, modelFilter string) []string {
	if modelFilter != "" {
		return []string{modelFilter}
	}

	switch jobType {
	case store.JobHealth:
		if relay.HealthModel != "" {
			return []string{relay.HealthModel}
		}
		return []string{}
	case store.JobPerformance, store.JobAuthenticity:
		if len(relay.ClaimedModels) > 0 {
			return append([]string(nil), relay.ClaimedModels...)
		}
		if relay.HealthModel != "" {
			return []string{relay.HealthModel}
		}
		return []string{}
	default:
		return nil
	}
}

func Enqueue(ctx context.Context, st *store.Store, q *queue.Queue, opts Options) ([]EnqueuedJob, error) {
	models := opts.Models
	if len(models) == 0 {
		models = ModelsForJob(opts.Relay, opts.JobType, "")
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models to probe for relay %q", opts.Relay.Name)
	}

	out := make([]EnqueuedJob, 0, len(models))
	for _, model := range models {
		m := model
		job := &store.ProbeJob{
			RelayID:     opts.Relay.ID,
			JobType:     opts.JobType,
			Model:       &m,
			ScheduledAt: time.Now(),
			Status:      store.JobPending,
		}
		if err := st.CreateJob(ctx, job); err != nil {
			return out, err
		}
		if err := q.Enqueue(ctx, queue.JobPayload{
			JobID:   job.ID,
			RelayID: opts.Relay.ID,
			Type:    opts.JobType,
			Model:   model,
		}); err != nil {
			return out, err
		}
		out = append(out, EnqueuedJob{JobID: job.ID, Model: model})
	}
	return out, nil
}

func Wait(ctx context.Context, st *store.Store, jobIDs []uuid.UUID, timeout time.Duration) ([]store.ProbeJob, error) {
	if len(jobIDs) == 0 {
		return nil, nil
	}
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	pending := make(map[uuid.UUID]struct{}, len(jobIDs))
	for _, id := range jobIDs {
		pending[id] = struct{}{}
	}

	var finished []store.ProbeJob
	for {
		for id := range pending {
			job, err := st.GetJob(ctx, id)
			if err != nil {
				return finished, err
			}
			if job == nil {
				delete(pending, id)
				continue
			}
			switch job.Status {
			case store.JobCompleted, store.JobFailed:
				finished = append(finished, *job)
				delete(pending, id)
			}
		}
		if len(pending) == 0 {
			return finished, nil
		}
		if time.Now().After(deadline) {
			ids := make([]string, 0, len(pending))
			for id := range pending {
				ids = append(ids, id.String())
			}
			return finished, fmt.Errorf("timeout waiting for jobs: %s", strings.Join(ids, ", "))
		}

		select {
		case <-ctx.Done():
			return finished, ctx.Err()
		case <-ticker.C:
		}
	}
}
