package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wadeling/origin-check/internal/store"
)

const queueKey = "origin-check:jobs"

type JobPayload struct {
	JobID   uuid.UUID      `json:"job_id"`
	RelayID uuid.UUID      `json:"relay_id"`
	Type    store.JobType  `json:"type"`
	Model   string         `json:"model,omitempty"`
}

type Queue struct {
	client *redis.Client
}

func New(redisURL string) (*Queue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opt)
	return &Queue{client: client}, nil
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *Queue) Close() error {
	return q.client.Close()
}

func (q *Queue) Enqueue(ctx context.Context, payload JobPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, queueKey, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context, timeout time.Duration) (*JobPayload, error) {
	result, err := q.client.BRPop(ctx, timeout, queueKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(result) < 2 {
		return nil, nil
	}
	var payload JobPayload
	if err := json.Unmarshal([]byte(result[1]), &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
