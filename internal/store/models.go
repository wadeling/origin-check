package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelayStatus string

const (
	RelayActive   RelayStatus = "active"
	RelayDisabled RelayStatus = "disabled"
	RelayError    RelayStatus = "error"
)

type JobType string

const (
	JobHealth        JobType = "health"
	JobPerformance   JobType = "performance"
	JobAuthenticity  JobType = "authenticity"
)

type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

type Verdict string

const (
	VerdictPass       Verdict = "pass"
	VerdictSuspicious Verdict = "suspicious"
	VerdictFail       Verdict = "fail"
	VerdictUnknown    Verdict = "unknown"
)

type Relay struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	WebsiteURL         string      `json:"website_url"`
	APIBaseURL         string      `json:"api_base_url"`
	BackupAPIBaseURLs  []string    `json:"backup_api_base_urls"`
	APIKeyEncrypted    []byte      `json:"-"`
	AuthType           string      `json:"auth_type"`
	ClaimedModels      []string    `json:"claimed_models"`
	HealthModel        string      `json:"health_model"`
	Status             RelayStatus `json:"status"`
	Tags               []string    `json:"tags"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

type ProbeJob struct {
	ID          uuid.UUID  `json:"id"`
	RelayID     uuid.UUID  `json:"relay_id"`
	JobType     JobType    `json:"job_type"`
	Model       *string    `json:"model,omitempty"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      JobStatus  `json:"status"`
	Error       *string    `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ProbeResult struct {
	ID               uuid.UUID `json:"id"`
	JobID            uuid.UUID `json:"job_id"`
	RelayID          uuid.UUID `json:"relay_id"`
	Model            string    `json:"model"`
	ProbeType        JobType   `json:"probe_type"`
	Stream           bool      `json:"stream"`
	LatencyMS        *int      `json:"latency_ms,omitempty"`
	TTFTMS           *int      `json:"ttft_ms,omitempty"`
	TPOTMS           *float64  `json:"tpot_ms,omitempty"`
	InputTokens      *int      `json:"input_tokens,omitempty"`
	OutputTokens     *int      `json:"output_tokens,omitempty"`
	HTTPStatus       *int      `json:"http_status,omitempty"`
	ResponseModel    *string   `json:"response_model,omitempty"`
	Error            *string   `json:"error,omitempty"`
	RawResponseHash  *string   `json:"raw_response_hash,omitempty"`
	ResponseSummary  *string   `json:"response_summary,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type PromptCase struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	ModelTarget     string          `json:"model_target"`
	Prompt          string          `json:"prompt"`
	ExpectedTraits  json.RawMessage `json:"expected_traits"`
	Weight          float64         `json:"weight"`
	Active          bool            `json:"active"`
	CreatedAt       time.Time       `json:"created_at"`
}

type SignalEvidence struct {
	Signal   string  `json:"signal"`
	Score    float64 `json:"score"`
	Weight   float64 `json:"weight"`
	Detail   string  `json:"detail"`
	Prompt   string  `json:"prompt,omitempty"`
	Response string  `json:"response,omitempty"`
}

type AuthenticityReport struct {
	ID           uuid.UUID        `json:"id"`
	RelayID      uuid.UUID        `json:"relay_id"`
	JobID        *uuid.UUID       `json:"job_id,omitempty"`
	ClaimedModel string           `json:"claimed_model"`
	Score        float64          `json:"score"`
	Confidence   float64          `json:"confidence"`
	Verdict      Verdict          `json:"verdict"`
	Signals      []SignalEvidence `json:"signals"`
	CreatedAt    time.Time        `json:"created_at"`
}

type DailyAggregate struct {
	ID             uuid.UUID `json:"id"`
	RelayID        uuid.UUID `json:"relay_id"`
	Model          string    `json:"model"`
	Date           time.Time `json:"date"`
	AvgLatencyMS   *float64  `json:"avg_latency_ms,omitempty"`
	AvgTTFTMS      *float64  `json:"avg_ttft_ms,omitempty"`
	AvgTPOTMS      *float64  `json:"avg_tpot_ms,omitempty"`
	AvailabilityPct *float64 `json:"availability_pct,omitempty"`
	ProbeCount     int       `json:"probe_count"`
	ErrorCount     int       `json:"error_count"`
	PerfScore      *float64  `json:"perf_score,omitempty"`
}

type RelaySummary struct {
	Relay
	LatestTTFTMS       *float64 `json:"latest_ttft_ms,omitempty"`
	LatestLatencyMS    *int     `json:"latest_latency_ms,omitempty"`
	Availability24h    *float64 `json:"availability_24h,omitempty"`
	PerfScore          *float64 `json:"perf_score,omitempty"`
	AuthenticityScore  *float64 `json:"authenticity_score,omitempty"`
	AuthenticityVerdict *Verdict `json:"authenticity_verdict,omitempty"`
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
