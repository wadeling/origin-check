'use client';

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import type { RelaySummary } from '@/lib/api';

const BAR_COLORS = ['#3b82f6', '#6366f1', '#8b5cf6', '#a855f7', '#06b6d4', '#10b981'];

function CustomTooltip({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number }>; label?: string }) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-xl border border-line bg-surface-elevated px-3 py-2 text-xs shadow-card">
      <p className="font-medium text-foreground">{label}</p>
      <p className="mt-0.5 font-mono text-accent">{payload[0].value.toFixed(0)} 分</p>
    </div>
  );
}

export function PerformanceOverview({ relays }: { relays: RelaySummary[] }) {
  const data = [...relays]
    .filter((r) => r.perf_score != null)
    .sort((a, b) => (b.perf_score ?? 0) - (a.perf_score ?? 0))
    .map((r) => ({
      name: r.name.length > 10 ? `${r.name.slice(0, 9)}…` : r.name,
      fullName: r.name,
      score: r.perf_score ?? 0,
      ttft: r.latest_ttft_ms ?? 0,
    }));

  if (data.length === 0) {
    return (
      <div className="glass flex h-64 items-center justify-center rounded-2xl text-sm text-foreground-muted">
        暂无性能数据，等待 daily performance 探测
      </div>
    );
  }

  return (
    <div className="glass animate-fade-up rounded-2xl p-5 shadow-card opacity-0 sm:p-6" style={{ animationDelay: '400ms' }}>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-sm font-semibold uppercase tracking-wider text-foreground-muted">Performance</h2>
          <p className="mt-1 text-lg font-medium text-foreground">综合性能对比</p>
        </div>
        <p className="text-xs text-foreground-muted">TTFT + 24h 可用率加权</p>
      </div>
      <div className="h-56 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} margin={{ top: 4, right: 4, left: -20, bottom: 0 }} barCategoryGap="28%">
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
            <XAxis
              dataKey="name"
              tick={{ fill: 'var(--foreground-muted)', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              domain={[0, 100]}
              tick={{ fill: 'var(--foreground-muted)', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip content={<CustomTooltip />} cursor={{ fill: 'var(--accent-soft)' }} />
            <Bar dataKey="score" radius={[6, 6, 0, 0]} maxBarSize={48}>
              {data.map((_, i) => (
                <Cell key={i} fill={BAR_COLORS[i % BAR_COLORS.length]} fillOpacity={0.9} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
