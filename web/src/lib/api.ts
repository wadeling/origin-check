export type Verdict = 'pass' | 'suspicious' | 'fail' | 'unknown';

export interface RelaySummary {
  id: string;
  name: string;
  website_url: string;
  api_base_url: string;
  claimed_models: string[];
  health_model: string;
  status: string;
  tags: string[];
  latest_ttft_ms?: number;
  latest_latency_ms?: number;
  availability_24h?: number;
  perf_score?: number;
  authenticity_score?: number;
  authenticity_verdict?: Verdict;
}

export interface SignalEvidence {
  signal: string;
  score: number;
  weight: number;
  detail: string;
  alert?: string;
  prompt?: string;
  response?: string;
}

export interface AuthenticityReport {
  id: string;
  relay_id: string;
  claimed_model: string;
  score: number;
  confidence: number;
  verdict: Verdict;
  signals: SignalEvidence[];
  created_at: string;
}

export interface ProbeResult {
  id: string;
  model: string;
  probe_type: string;
  stream: boolean;
  latency_ms?: number;
  ttft_ms?: number;
  tpot_ms?: number;
  http_status?: number;
  response_model?: string;
  error?: string;
  response_summary?: string;
  created_at: string;
}

export interface RelayDetailResponse {
  relay: RelaySummary;
  recent_results: ProbeResult[];
  availability_24h?: number;
  latest_latency_ms?: number;
  latest_ttft_ms?: number;
  authenticity_score?: number;
  authenticity_verdict?: Verdict;
  authenticity_report?: AuthenticityReport;
}

export interface MetricsResponse {
  hourly: Array<{
    time: string;
    avg_latency?: number;
    avg_ttft?: number;
    availability: number;
    probe_count: number;
  }>;
  daily: Array<{
    date: string;
    model?: string;
    avg_latency_ms?: number;
    avg_ttft_ms?: number;
    availability_pct?: number;
    probe_count?: number;
  }>;
}

function getApiUrl() {
  if (typeof window === 'undefined') {
    return process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
}

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${getApiUrl()}${path}`, { next: { revalidate: 60 } });
  if (!res.ok) {
    throw new Error(`API ${path}: ${res.status}`);
  }
  return res.json();
}

export function getLeaderboard() {
  return fetchJSON<RelaySummary[]>('/api/v1/leaderboard');
}

export function getRelay(id: string) {
  return fetchJSON<RelayDetailResponse>(`/api/v1/relays/${id}`);
}

export function getRelayMetrics(id: string) {
  return fetchJSON<MetricsResponse>(`/api/v1/relays/${id}/metrics`);
}

export function getRelayReports(id: string) {
  return fetchJSON<AuthenticityReport[]>(`/api/v1/relays/${id}/reports`);
}

export function verdictLabel(v: Verdict | undefined) {
  switch (v) {
    case 'pass':
      return { text: '一致', className: 'text-pass bg-green-500/10 border-green-500/30 dark:bg-green-500/15' };
    case 'suspicious':
      return { text: '存疑', className: 'text-warn bg-yellow-500/10 border-yellow-500/30 dark:bg-yellow-500/15' };
    case 'fail':
      return { text: '不符', className: 'text-fail bg-red-500/10 border-red-500/30 dark:bg-red-500/15' };
    default:
      return { text: '待测', className: 'text-foreground-muted bg-surface-muted border-line' };
  }
}

export function formatMs(v?: number | null) {
  if (v == null) return '—';
  return `${Math.round(v)} ms`;
}

export function formatPct(v?: number | null) {
  if (v == null) return '—';
  return `${v.toFixed(1)}%`;
}

const cstOptions: Intl.DateTimeFormatOptions = {
  timeZone: 'Asia/Shanghai',
  hour12: false,
};

/** Format ISO timestamp for display in China Standard Time (UTC+8). */
export function formatDateTimeCST(iso: string) {
  return new Date(iso).toLocaleString('zh-CN', {
    ...cstOptions,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/** Format ISO date for chart axis in CST (e.g. 5/24). */
export function formatDateCST(iso: string) {
  return new Date(iso).toLocaleDateString('zh-CN', {
    timeZone: 'Asia/Shanghai',
    month: 'numeric',
    day: 'numeric',
  });
}
