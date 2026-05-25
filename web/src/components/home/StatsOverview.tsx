import type { RelaySummary } from '@/lib/api';

function avg(values: number[]) {
  if (values.length === 0) return null;
  return values.reduce((a, b) => a + b, 0) / values.length;
}

export function StatsOverview({ relays }: { relays: RelaySummary[] }) {
  const authScores = relays.map((r) => r.authenticity_score).filter((v): v is number => v != null);
  const avail = relays.map((r) => r.availability_24h).filter((v): v is number => v != null);
  const ttfts = relays.map((r) => r.latest_ttft_ms).filter((v): v is number => v != null);
  const perf = relays.map((r) => r.perf_score).filter((v): v is number => v != null);

  const passCount = relays.filter((r) => r.authenticity_verdict === 'pass').length;
  const suspCount = relays.filter((r) => r.authenticity_verdict === 'suspicious').length;

  const stats = [
    {
      label: '收录中转站',
      value: String(relays.length),
      sub: '持续探测中',
      accent: 'text-accent',
    },
    {
      label: '真伪均分',
      value: authScores.length ? avg(authScores)!.toFixed(0) : '—',
      sub: `一致 ${passCount} · 存疑 ${suspCount}`,
      accent: 'text-foreground',
    },
    {
      label: '24h 可用率',
      value: avail.length ? `${avg(avail)!.toFixed(1)}%` : '—',
      sub: 'health 探测均值',
      accent: 'text-pass',
    },
    {
      label: '最佳 TTFT',
      value: ttfts.length ? `${Math.round(Math.min(...ttfts))} ms` : '—',
      sub: perf.length ? `性能均分 ${avg(perf)!.toFixed(0)}` : '性能探测中',
      accent: 'text-foreground',
    },
  ];

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {stats.map((stat, i) => (
        <div
          key={stat.label}
          className="glass card-hover animate-fade-up rounded-2xl p-5 shadow-card opacity-0"
          style={{ animationDelay: `${120 + i * 70}ms` }}
        >
          <p className="text-xs font-medium uppercase tracking-wider text-foreground-muted">{stat.label}</p>
          <p className={`mt-2 font-mono text-3xl font-semibold tracking-tight ${stat.accent}`}>{stat.value}</p>
          <p className="mt-1 text-xs text-foreground-muted">{stat.sub}</p>
        </div>
      ))}
    </div>
  );
}
