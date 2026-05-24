-- Origin Check initial schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE relay_status AS ENUM ('active', 'disabled', 'error');
CREATE TYPE job_type AS ENUM ('health', 'performance', 'authenticity');
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE verdict_type AS ENUM ('pass', 'suspicious', 'fail', 'unknown');

CREATE TABLE relays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    website_url TEXT NOT NULL,
    api_base_url TEXT NOT NULL,
    backup_api_base_urls TEXT[] NOT NULL DEFAULT '{}',
    api_key_encrypted BYTEA,
    auth_type TEXT NOT NULL DEFAULT 'bearer_token',
    claimed_models TEXT[] NOT NULL DEFAULT '{}',
    health_model TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    status relay_status NOT NULL DEFAULT 'active',
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE probe_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    job_type job_type NOT NULL,
    model TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status job_status NOT NULL DEFAULT 'pending',
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_probe_jobs_relay ON probe_jobs(relay_id, created_at DESC);
CREATE INDEX idx_probe_jobs_status ON probe_jobs(status) WHERE status IN ('pending', 'running');

CREATE TABLE probe_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES probe_jobs(id) ON DELETE CASCADE,
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    model TEXT NOT NULL,
    probe_type job_type NOT NULL,
    stream BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INT,
    ttft_ms INT,
    tpot_ms NUMERIC(10,2),
    input_tokens INT,
    output_tokens INT,
    http_status INT,
    response_model TEXT,
    error TEXT,
    raw_response_hash TEXT,
    response_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_probe_results_relay_time ON probe_results(relay_id, created_at DESC);
CREATE INDEX idx_probe_results_model ON probe_results(relay_id, model, created_at DESC);

CREATE TABLE prompt_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    model_target TEXT NOT NULL,
    prompt TEXT NOT NULL,
    expected_traits JSONB NOT NULL DEFAULT '{}',
    weight NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE authenticity_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    job_id UUID REFERENCES probe_jobs(id) ON DELETE SET NULL,
    claimed_model TEXT NOT NULL,
    score NUMERIC(5,2) NOT NULL,
    confidence NUMERIC(5,2) NOT NULL,
    verdict verdict_type NOT NULL DEFAULT 'unknown',
    signals JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_reports_relay ON authenticity_reports(relay_id, created_at DESC);

CREATE TABLE daily_aggregates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    model TEXT NOT NULL,
    date DATE NOT NULL,
    avg_latency_ms NUMERIC(10,2),
    avg_ttft_ms NUMERIC(10,2),
    avg_tpot_ms NUMERIC(10,2),
    availability_pct NUMERIC(5,2),
    probe_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    perf_score NUMERIC(5,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (relay_id, model, date)
);

CREATE INDEX idx_daily_agg_relay_date ON daily_aggregates(relay_id, date DESC);

-- Seed default fingerprint prompt cases
INSERT INTO prompt_cases (name, model_target, prompt, expected_traits, weight) VALUES
(
    'json_format_strict',
    'general',
    'Reply with ONLY valid JSON, no markdown: {"answer": 42, "reason": "because"}',
    '{"must_contain": ["42"], "must_be_json": true, "max_length": 200}',
    1.0
),
(
    'logic_puzzle',
    'general',
    'If all bloops are razzies and all razzies are lazzies, are all bloops definitely lazzies? Answer yes or no in one word.',
    '{"must_match_one": ["yes", "Yes", "YES"], "max_length": 20}',
    1.0
),
(
    'math_simple',
    'general',
    'What is 17 * 23? Reply with only the number.',
    '{"must_contain": ["391"], "max_length": 10}',
    1.0
),
(
    'word_count',
    'general',
    'Write exactly three words about the sky.',
    '{"word_count_min": 3, "word_count_max": 5, "max_length": 50}',
    1.0
),
(
    'no_markdown',
    'general',
    'List two colors. Plain text only, no bullets, no markdown, no numbering.',
    '{"must_not_contain": ["*", "-", "1.", "```"], "max_length": 100}',
    1.0
);
