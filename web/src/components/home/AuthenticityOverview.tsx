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

function AuthTooltip({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number }>; label?: string }) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-xl border border-line bg-surface-elevated px-3 py-2 text-xs shadow-card">
      <p className="font-medium text-foreground">{label}</p>
      <p className="mt-0.5 font-mono text-foreground">{payload[0].value.toFixed(0)} 分</p>
    </div>
  );
}

const VERDICT_COLORS: Record<string, string> = {
  pass: '#22c55e',
  suspicious: '#eab308',
  fail: '#ef4444',
  unknown: '#94a3b8',
};

export function AuthenticityOverview({ relays }: { relays: RelaySummary[] }) {
  const data = [...relays]
    .filter((r) => r.authenticity_score != null)
    .sort((a, b) => (b.authenticity_score ?? 0) - (a.authenticity_score ?? 0))
    .map((r) => ({
      name: r.name.length > 8 ? `${r.name.slice(0, 7)}…` : r.name,
      score: r.authenticity_score ?? 0,
      verdict: r.authenticity_verdict ?? 'unknown',
    }));

  if (data.length === 0) return null;

  return (
    <div className="glass animate-fade-up rounded-2xl p-5 shadow-card opacity-0 sm:p-6" style={{ animationDelay: '480ms' }}>
      <div className="mb-5">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-foreground-muted">Authenticity</h2>
        <p className="mt-1 text-lg font-medium text-foreground">真伪综合得分</p>
      </div>
      <div className="h-48 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} layout="vertical" margin={{ top: 0, right: 8, left: 0, bottom: 0 }} barCategoryGap="20%">
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" horizontal={false} />
            <XAxis type="number" domain={[0, 100]} hide />
            <YAxis
              type="category"
              dataKey="name"
              width={72}
              tick={{ fill: 'var(--foreground-muted)', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip content={<AuthTooltip />} cursor={{ fill: 'var(--accent-soft)' }} />
            <Bar dataKey="score" radius={[0, 6, 6, 0]} maxBarSize={20}>
              {data.map((entry, i) => (
                <Cell key={i} fill={VERDICT_COLORS[entry.verdict] ?? VERDICT_COLORS.unknown} fillOpacity={0.85} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
