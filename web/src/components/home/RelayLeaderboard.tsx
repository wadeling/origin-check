'use client';

import Link from 'next/link';
import { VerdictBadge } from '@/components/RelayTable';
import type { RelaySummary } from '@/lib/api';

function rankClass(rank: number) {
  if (rank === 1) return 'rank-1';
  if (rank === 2) return 'rank-2';
  if (rank === 3) return 'rank-3';
  return 'rank-default';
}

function MetricBar({ value, max = 100, color }: { value: number; max?: number; color: string }) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100));
  return (
    <div className="metric-bar">
      <div className="metric-bar-fill" style={{ width: `${pct}%`, backgroundColor: color }} />
    </div>
  );
}

function verdictColor(verdict?: RelaySummary['authenticity_verdict']) {
  switch (verdict) {
    case 'pass':
      return '#22c55e';
    case 'suspicious':
      return '#eab308';
    case 'fail':
      return '#ef4444';
    default:
      return '#94a3b8';
  }
}

export function RelayLeaderboard({ relays }: { relays: RelaySummary[] }) {
  const sorted = [...relays].sort((a, b) => {
    const perfA = a.perf_score ?? 0;
    const perfB = b.perf_score ?? 0;
    if (perfB !== perfA) return perfB - perfA;
    return (b.authenticity_score ?? 0) - (a.authenticity_score ?? 0);
  });

  return (
    <section className="animate-fade-up opacity-0" style={{ animationDelay: '560ms' }}>
      <div className="mb-5 flex items-end justify-between gap-4">
        <div>
          <h2 className="text-sm font-semibold uppercase tracking-wider text-foreground-muted">Leaderboard</h2>
          <p className="mt-1 text-xl font-semibold text-foreground">中转站排行榜</p>
        </div>
        <p className="hidden text-xs text-foreground-muted sm:block">按综合性能排序 · 点击查看详情</p>
      </div>

      <div className="space-y-3 lg:hidden">
        {sorted.map((relay, i) => (
          <Link
            key={relay.id}
            href={`/relays/${relay.id}`}
            className="glass card-hover block rounded-2xl p-4 shadow-card"
          >
            <div className="flex items-start gap-3">
              <span className={`rank-badge ${rankClass(i + 1)}`}>{i + 1}</span>
              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-2">
                  <h3 className="truncate font-semibold text-foreground">{relay.name}</h3>
                  <VerdictBadge verdict={relay.authenticity_verdict} />
                </div>
                <p className="mt-1 truncate text-xs text-foreground-muted">{relay.claimed_models.join(' · ')}</p>
                <div className="mt-3 grid grid-cols-3 gap-2 text-center">
                  <div>
                    <p className="font-mono text-sm font-medium">{relay.latest_ttft_ms != null ? `${Math.round(relay.latest_ttft_ms)}` : '—'}</p>
                    <p className="text-[10px] uppercase text-foreground-muted">TTFT</p>
                  </div>
                  <div>
                    <p className="font-mono text-sm font-medium">{relay.availability_24h != null ? `${relay.availability_24h.toFixed(0)}%` : '—'}</p>
                    <p className="text-[10px] uppercase text-foreground-muted">可用</p>
                  </div>
                  <div>
                    <p className="font-mono text-sm font-medium">{relay.perf_score != null ? relay.perf_score.toFixed(0) : '—'}</p>
                    <p className="text-[10px] uppercase text-foreground-muted">性能</p>
                  </div>
                </div>
              </div>
            </div>
          </Link>
        ))}
      </div>

      <div className="glass hidden overflow-hidden rounded-2xl shadow-card lg:block">
        <table className="min-w-full">
          <thead>
            <tr className="border-b border-line text-left text-[11px] font-semibold uppercase tracking-wider text-foreground-muted">
              <th className="px-5 py-4 w-16">#</th>
              <th className="px-5 py-4">中转站</th>
              <th className="px-5 py-4">真伪</th>
              <th className="px-5 py-4">TTFT</th>
              <th className="px-5 py-4">可用率</th>
              <th className="px-5 py-4 w-40">性能</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((relay, i) => (
              <tr
                key={relay.id}
                className="group border-b border-line/60 transition-colors duration-200 last:border-0 hover:bg-accent-soft/30"
              >
                <td className="px-5 py-4">
                  <span className={`rank-badge ${rankClass(i + 1)}`}>{i + 1}</span>
                </td>
                <td className="px-5 py-4">
                  <Link
                    href={`/relays/${relay.id}`}
                    className="font-medium text-foreground transition-colors group-hover:text-accent"
                  >
                    {relay.name}
                  </Link>
                  <p className="mt-1 max-w-xs truncate text-xs text-foreground-muted">{relay.claimed_models.join(', ')}</p>
                </td>
                <td className="px-5 py-4">
                  <VerdictBadge verdict={relay.authenticity_verdict} />
                  {relay.authenticity_score != null && (
                    <div className="mt-2 w-28">
                      <MetricBar value={relay.authenticity_score} color={verdictColor(relay.authenticity_verdict)} />
                      <p className="mt-1 font-mono text-[10px] text-foreground-muted">近6次 {relay.authenticity_score.toFixed(0)}</p>
                    </div>
                  )}
                </td>
                <td className="px-5 py-4">
                  <span className="font-mono text-sm tabular-nums">
                    {relay.latest_ttft_ms != null ? `${Math.round(relay.latest_ttft_ms)} ms` : '—'}
                  </span>
                </td>
                <td className="px-5 py-4">
                  <span className="font-mono text-sm tabular-nums text-pass">
                    {relay.availability_24h != null ? `${relay.availability_24h.toFixed(1)}%` : '—'}
                  </span>
                </td>
                <td className="px-5 py-4">
                  <div className="flex items-center gap-3">
                    <span className="w-8 font-mono text-sm tabular-nums">
                      {relay.perf_score != null ? relay.perf_score.toFixed(0) : '—'}
                    </span>
                    {relay.perf_score != null && (
                      <div className="flex-1">
                        <MetricBar value={relay.perf_score} color="var(--accent)" />
                      </div>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
