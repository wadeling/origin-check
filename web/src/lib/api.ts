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
    avg_latency_ms?: number;
    avg_ttft_ms?: number;
    availability_pct?: number;
  }>;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, { next: { revalidate: 60 } });
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
      return { text: '一致', className: 'text-pass bg-green-50 border-green-200' };
    case 'suspicious':
      return { text: '存疑', className: 'text-warn bg-yellow-50 border-yellow-200' };
    case 'fail':
      return { text: '不符', className: 'text-fail bg-red-50 border-red-200' };
    default:
      return { text: '待测', className: 'text-muted bg-slate-50 border-slate-200' };
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
