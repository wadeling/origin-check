'use client';

import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { formatDateCST, type MetricsResponse } from '@/lib/api';

function avg(values: number[]) {
  if (values.length === 0) return undefined;
  return values.reduce((a, b) => a + b, 0) / values.length;
}

/** Roll up per-model daily rows into one point per calendar day. */
function aggregateDaily(metrics: MetricsResponse) {
  const buckets = new Map<
    string,
    { ttft: number[]; latency: number[]; availability: number[] }
  >();

  for (const row of metrics.daily) {
    if (row.probe_count != null && row.probe_count === 0) continue;
    const dateKey = row.date.slice(0, 10);
    const bucket = buckets.get(dateKey) ?? { ttft: [], latency: [], availability: [] };
    if (row.avg_ttft_ms != null) bucket.ttft.push(row.avg_ttft_ms);
    if (row.avg_latency_ms != null) bucket.latency.push(row.avg_latency_ms);
    if (row.availability_pct != null) bucket.availability.push(row.availability_pct);
    buckets.set(dateKey, bucket);
  }

  return Array.from(buckets.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, vals]) => ({
      time: formatDateCST(date),
      ttft: avg(vals.ttft),
      latency: avg(vals.latency),
      availability: avg(vals.availability),
    }));
}

export function MetricsChart({ metrics }: { metrics: MetricsResponse }) {
  const data = aggregateDaily(metrics);

  if (data.length === 0) {
    return (
      <p className="text-sm text-muted">
        暂无性能数据，等待后台 performance 探测任务运行（默认每天一次）。
      </p>
    );
  }

  return (
    <div className="h-72 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
          <XAxis dataKey="time" tick={{ fontSize: 12 }} />
          <YAxis yAxisId="left" tick={{ fontSize: 12 }} />
          <YAxis yAxisId="right" orientation="right" domain={[0, 100]} tick={{ fontSize: 12 }} />
          <Tooltip />
          <Legend />
          <Line yAxisId="left" type="monotone" dataKey="ttft" name="TTFT (ms)" stroke="#2563eb" strokeWidth={2} dot={{ r: 3 }} connectNulls={false} />
          <Line yAxisId="left" type="monotone" dataKey="latency" name="延迟 (ms)" stroke="#7c3aed" strokeWidth={2} dot={{ r: 3 }} connectNulls={false} />
          <Line yAxisId="right" type="monotone" dataKey="availability" name="可用率 (%)" stroke="#16a34a" strokeWidth={2} dot={{ r: 3 }} connectNulls={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
