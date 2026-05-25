package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) UpsertRelay(ctx context.Context, r *Relay) error {
	query := `
		INSERT INTO relays (name, website_url, api_base_url, backup_api_base_urls, api_key_encrypted, auth_type, claimed_models, health_model, status, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (name) DO UPDATE SET
			website_url = EXCLUDED.website_url,
			api_base_url = EXCLUDED.api_base_url,
			backup_api_base_urls = EXCLUDED.backup_api_base_urls,
			api_key_encrypted = COALESCE(EXCLUDED.api_key_encrypted, relays.api_key_encrypted),
			auth_type = EXCLUDED.auth_type,
			claimed_models = EXCLUDED.claimed_models,
			health_model = EXCLUDED.health_model,
			status = EXCLUDED.status,
			tags = EXCLUDED.tags,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`
	return s.pool.QueryRow(ctx, query,
		r.Name, r.WebsiteURL, r.APIBaseURL, r.BackupAPIBaseURLs, r.APIKeyEncrypted,
		r.AuthType, r.ClaimedModels, r.HealthModel, r.Status, r.Tags,
	).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
}

func (s *Store) ListRelays(ctx context.Context) ([]Relay, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, website_url, api_base_url, backup_api_base_urls, auth_type, claimed_models, health_model, status, tags, created_at, updated_at
		FROM relays ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relays []Relay
	for rows.Next() {
		var r Relay
		if err := rows.Scan(&r.ID, &r.Name, &r.WebsiteURL, &r.APIBaseURL, &r.BackupAPIBaseURLs,
			&r.AuthType, &r.ClaimedModels, &r.HealthModel, &r.Status, &r.Tags, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		relays = append(relays, r)
	}
	return relays, rows.Err()
}

func (s *Store) GetRelay(ctx context.Context, id uuid.UUID) (*Relay, error) {
	var r Relay
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, website_url, api_base_url, backup_api_base_urls, api_key_encrypted, auth_type, claimed_models, health_model, status, tags, created_at, updated_at
		FROM relays WHERE id = $1`, id).Scan(
		&r.ID, &r.Name, &r.WebsiteURL, &r.APIBaseURL, &r.BackupAPIBaseURLs, &r.APIKeyEncrypted,
		&r.AuthType, &r.ClaimedModels, &r.HealthModel, &r.Status, &r.Tags, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetRelayByName(ctx context.Context, name string) (*Relay, error) {
	var r Relay
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, website_url, api_base_url, backup_api_base_urls, api_key_encrypted, auth_type, claimed_models, health_model, status, tags, created_at, updated_at
		FROM relays WHERE name = $1`, name).Scan(
		&r.ID, &r.Name, &r.WebsiteURL, &r.APIBaseURL, &r.BackupAPIBaseURLs, &r.APIKeyEncrypted,
		&r.AuthType, &r.ClaimedModels, &r.HealthModel, &r.Status, &r.Tags, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) ListActiveRelays(ctx context.Context) ([]Relay, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, website_url, api_base_url, backup_api_base_urls, api_key_encrypted, auth_type, claimed_models, health_model, status, tags, created_at, updated_at
		FROM relays WHERE status = 'active' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relays []Relay
	for rows.Next() {
		var r Relay
		if err := rows.Scan(&r.ID, &r.Name, &r.WebsiteURL, &r.APIBaseURL, &r.BackupAPIBaseURLs, &r.APIKeyEncrypted,
			&r.AuthType, &r.ClaimedModels, &r.HealthModel, &r.Status, &r.Tags, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		relays = append(relays, r)
	}
	return relays, rows.Err()
}

func (s *Store) CreateJob(ctx context.Context, job *ProbeJob) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO probe_jobs (relay_id, job_type, model, scheduled_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		job.RelayID, job.JobType, job.Model, job.ScheduledAt, job.Status,
	).Scan(&job.ID, &job.CreatedAt)
}

func (s *Store) GetJob(ctx context.Context, id uuid.UUID) (*ProbeJob, error) {
	var job ProbeJob
	err := s.pool.QueryRow(ctx, `
		SELECT id, relay_id, job_type, model, scheduled_at, started_at, completed_at, status, error, created_at
		FROM probe_jobs WHERE id = $1`, id).Scan(
		&job.ID, &job.RelayID, &job.JobType, &job.Model, &job.ScheduledAt,
		&job.StartedAt, &job.CompletedAt, &job.Status, &job.Error, &job.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Store) UpdateJobStatus(ctx context.Context, id uuid.UUID, status JobStatus, errMsg *string) error {
	now := time.Now()
	var query string
	switch status {
	case JobRunning:
		query = `UPDATE probe_jobs SET status = $2, started_at = $3 WHERE id = $1`
		_, err := s.pool.Exec(ctx, query, id, status, now)
		return err
	case JobCompleted, JobFailed:
		query = `UPDATE probe_jobs SET status = $2, completed_at = $3, error = $4 WHERE id = $1`
		_, err := s.pool.Exec(ctx, query, id, status, now, errMsg)
		return err
	default:
		query = `UPDATE probe_jobs SET status = $2 WHERE id = $1`
		_, err := s.pool.Exec(ctx, query, id, status)
		return err
	}
}

func (s *Store) InsertProbeResult(ctx context.Context, r *ProbeResult) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO probe_results (job_id, relay_id, model, probe_type, stream, latency_ms, ttft_ms, tpot_ms, input_tokens, output_tokens, http_status, response_model, error, raw_response_hash, response_summary)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at`,
		r.JobID, r.RelayID, r.Model, r.ProbeType, r.Stream, r.LatencyMS, r.TTFTMS, r.TPOTMS,
		r.InputTokens, r.OutputTokens, r.HTTPStatus, r.ResponseModel, r.Error, r.RawResponseHash, r.ResponseSummary,
	).Scan(&r.ID, &r.CreatedAt)
}

func (s *Store) ListPromptCases(ctx context.Context) ([]PromptCase, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, model_target, prompt, expected_traits, weight, active, created_at
		FROM prompt_cases WHERE active = TRUE ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []PromptCase
	for rows.Next() {
		var c PromptCase
		if err := rows.Scan(&c.ID, &c.Name, &c.ModelTarget, &c.Prompt, &c.ExpectedTraits, &c.Weight, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

func (s *Store) InsertAuthenticityReport(ctx context.Context, report *AuthenticityReport) error {
	signalsJSON, err := json.Marshal(report.Signals)
	if err != nil {
		return err
	}
	return s.pool.QueryRow(ctx, `
		INSERT INTO authenticity_reports (relay_id, job_id, claimed_model, score, confidence, verdict, signals)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`,
		report.RelayID, report.JobID, report.ClaimedModel, report.Score, report.Confidence, report.Verdict, signalsJSON,
	).Scan(&report.ID, &report.CreatedAt)
}

func (s *Store) GetLatestAuthenticityReport(ctx context.Context, relayID uuid.UUID, model string) (*AuthenticityReport, error) {
	var report AuthenticityReport
	var signalsJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, relay_id, job_id, claimed_model, score, confidence, verdict, signals, created_at
		FROM authenticity_reports
		WHERE relay_id = $1 AND claimed_model = $2
		ORDER BY created_at DESC LIMIT 1`, relayID, model).Scan(
		&report.ID, &report.RelayID, &report.JobID, &report.ClaimedModel, &report.Score, &report.Confidence,
		&report.Verdict, &signalsJSON, &report.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(signalsJSON, &report.Signals); err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *Store) ListAuthenticityReports(ctx context.Context, relayID uuid.UUID, models []string, limit int) ([]AuthenticityReport, error) {
	var rows pgx.Rows
	var err error
	if len(models) > 0 {
		rows, err = s.pool.Query(ctx, `
			SELECT id, relay_id, job_id, claimed_model, score, confidence, verdict, signals, created_at
			FROM authenticity_reports
			WHERE relay_id = $1 AND claimed_model = ANY($2)
			ORDER BY created_at DESC LIMIT $3`, relayID, models, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, relay_id, job_id, claimed_model, score, confidence, verdict, signals, created_at
			FROM authenticity_reports WHERE relay_id = $1
			ORDER BY created_at DESC LIMIT $2`, relayID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []AuthenticityReport
	for rows.Next() {
		var report AuthenticityReport
		var signalsJSON []byte
		if err := rows.Scan(&report.ID, &report.RelayID, &report.JobID, &report.ClaimedModel, &report.Score, &report.Confidence,
			&report.Verdict, &signalsJSON, &report.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(signalsJSON, &report.Signals); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

const authenticitySummaryFetchLimit = 9

func (s *Store) GetAuthenticitySummary(ctx context.Context, relayID uuid.UUID, models []string) (*AuthenticitySummary, error) {
	reports, err := s.ListAuthenticityReports(ctx, relayID, models, authenticitySummaryFetchLimit)
	if err != nil {
		return nil, err
	}
	return SummarizeAuthenticityReports(reports), nil
}

func (s *Store) DeleteAuthenticityReportsForModels(ctx context.Context, models []string) error {
	if len(models) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM authenticity_reports WHERE claimed_model = ANY($1)`, models)
	return err
}

func (s *Store) DeleteProbeResultsForModels(ctx context.Context, models []string) error {
	if len(models) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM probe_results WHERE model = ANY($1)`, models)
	return err
}

func (s *Store) GetProbeResults(ctx context.Context, relayID uuid.UUID, since time.Time, limit int) ([]ProbeResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, job_id, relay_id, model, probe_type, stream, latency_ms, ttft_ms, tpot_ms, input_tokens, output_tokens, http_status, response_model, error, raw_response_hash, response_summary, created_at
		FROM probe_results WHERE relay_id = $1 AND created_at >= $2
		ORDER BY created_at DESC LIMIT $3`, relayID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProbeResults(rows)
}

func scanProbeResults(rows pgx.Rows) ([]ProbeResult, error) {
	var results []ProbeResult
	for rows.Next() {
		var r ProbeResult
		if err := rows.Scan(&r.ID, &r.JobID, &r.RelayID, &r.Model, &r.ProbeType, &r.Stream, &r.LatencyMS, &r.TTFTMS, &r.TPOTMS,
			&r.InputTokens, &r.OutputTokens, &r.HTTPStatus, &r.ResponseModel, &r.Error, &r.RawResponseHash, &r.ResponseSummary, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) ComputeAvailability24h(ctx context.Context, relayID uuid.UUID) (*float64, error) {
	since := time.Now().Add(-24 * time.Hour)
	var total, errors int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE error IS NOT NULL OR http_status >= 400 OR http_status IS NULL)
		FROM probe_results WHERE relay_id = $1 AND created_at >= $2 AND probe_type = 'health'`, relayID, since).Scan(&total, &errors)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}
	pct := float64(total-errors) / float64(total) * 100
	return &pct, nil
}

func (s *Store) GetLatestPerfMetrics(ctx context.Context, relayID uuid.UUID) (*int, *float64, error) {
	var latency *int
	var ttftInt *int
	err := s.pool.QueryRow(ctx, `
		SELECT latency_ms, ttft_ms FROM probe_results
		WHERE relay_id = $1 AND probe_type IN ('health', 'performance') AND error IS NULL
		ORDER BY created_at DESC LIMIT 1`, relayID).Scan(&latency, &ttftInt)
	if err == pgx.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var ttftF *float64
	if ttftInt != nil {
		v := float64(*ttftInt)
		ttftF = &v
	}
	return latency, ttftF, nil
}

func (s *Store) ListRelaySummaries(ctx context.Context) ([]RelaySummary, error) {
	relays, err := s.ListRelays(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]RelaySummary, 0, len(relays))
	for _, r := range relays {
		sum := RelaySummary{Relay: r}
		latency, ttft, err := s.GetLatestPerfMetrics(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		sum.LatestLatencyMS = latency
		sum.LatestTTFTMS = ttft

		avail, err := s.ComputeAvailability24h(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		sum.Availability24h = avail

		if avail != nil && ttft != nil {
			// Lower is better for TTFT; normalize with simple heuristic
			ttftScore := max(0, 100-(*ttft/50))
			score := 0.4*ttftScore + 0.6*(*avail)
			sum.PerfScore = &score
		}

		if len(r.ClaimedModels) > 0 {
			summary, err := s.GetAuthenticitySummary(ctx, r.ID, r.ClaimedModels)
			if err != nil {
				return nil, err
			}
			if summary != nil {
				sum.AuthenticityScore = &summary.Score
				sum.AuthenticityVerdict = &summary.Verdict
			}
		}

		summaries = append(summaries, sum)
	}
	return summaries, nil
}

func (s *Store) UpsertDailyAggregate(ctx context.Context, relayID uuid.UUID, model string, date time.Time) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO daily_aggregates (relay_id, model, date, avg_latency_ms, avg_ttft_ms, avg_tpot_ms, availability_pct, probe_count, error_count, perf_score)
		SELECT
			$1, $2, $3::date,
			AVG(latency_ms), AVG(ttft_ms), AVG(tpot_ms),
			(COUNT(*) FILTER (WHERE error IS NULL AND (http_status IS NULL OR http_status < 400))::numeric / NULLIF(COUNT(*), 0) * 100),
			COUNT(*),
			COUNT(*) FILTER (WHERE error IS NOT NULL OR http_status >= 400),
			NULL
		FROM probe_results
		WHERE relay_id = $1 AND model = $2 AND created_at >= $4 AND created_at < $5
		ON CONFLICT (relay_id, model, date) DO UPDATE SET
			avg_latency_ms = EXCLUDED.avg_latency_ms,
			avg_ttft_ms = EXCLUDED.avg_ttft_ms,
			avg_tpot_ms = EXCLUDED.avg_tpot_ms,
			availability_pct = EXCLUDED.availability_pct,
			probe_count = EXCLUDED.probe_count,
			error_count = EXCLUDED.error_count`,
		relayID, model, dayStart, dayStart, dayEnd,
	)
	return err
}

func (s *Store) GetDailyAggregates(ctx context.Context, relayID uuid.UUID, days int) ([]DailyAggregate, error) {
	since := time.Now().AddDate(0, 0, -days)
	rows, err := s.pool.Query(ctx, `
		SELECT id, relay_id, model, date, avg_latency_ms, avg_ttft_ms, avg_tpot_ms, availability_pct, probe_count, error_count, perf_score
		FROM daily_aggregates WHERE relay_id = $1 AND date >= $2::date
		ORDER BY date ASC`, relayID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggs []DailyAggregate
	for rows.Next() {
		var a DailyAggregate
		if err := rows.Scan(&a.ID, &a.RelayID, &a.Model, &a.Date, &a.AvgLatencyMS, &a.AvgTTFTMS, &a.AvgTPOTMS,
			&a.AvailabilityPct, &a.ProbeCount, &a.ErrorCount, &a.PerfScore); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	return aggs, rows.Err()
}

func (s *Store) GetHourlyMetrics(ctx context.Context, relayID uuid.UUID, hours int) ([]map[string]interface{}, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.pool.Query(ctx, `
		SELECT date_trunc('hour', created_at) AS bucket,
			AVG(latency_ms) AS avg_latency,
			AVG(ttft_ms) AS avg_ttft,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE error IS NULL AND (http_status IS NULL OR http_status < 400)) AS success
		FROM probe_results
		WHERE relay_id = $1 AND created_at >= $2
		GROUP BY bucket ORDER BY bucket ASC`, relayID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []map[string]interface{}
	for rows.Next() {
		var bucket time.Time
		var avgLatency, avgTTFT *float64
		var total, success int
		if err := rows.Scan(&bucket, &avgLatency, &avgTTFT, &total, &success); err != nil {
			return nil, err
		}
		avail := 0.0
		if total > 0 {
			avail = float64(success) / float64(total) * 100
		}
		points = append(points, map[string]interface{}{
			"time":         bucket.Format(time.RFC3339),
			"avg_latency":  avgLatency,
			"avg_ttft":     avgTTFT,
			"availability": avail,
			"probe_count":  total,
		})
	}
	return points, rows.Err()
}
