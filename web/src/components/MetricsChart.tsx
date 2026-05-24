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
import { formatTimeCST, type MetricsResponse } from '@/lib/api';

export function MetricsChart({ metrics }: { metrics: MetricsResponse }) {
  const data = metrics.hourly
    .filter((p) => p.probe_count > 0)
    .map((p) => ({
      time: formatTimeCST(p.time),
      ttft: p.avg_ttft,
      latency: p.avg_latency,
      availability: p.availability,
    }));

  if (data.length === 0) {
    return <p className="text-sm text-muted">暂无性能数据，等待后台探测任务运行。</p>;
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
          <Line yAxisId="left" type="monotone" dataKey="ttft" name="TTFT (ms)" stroke="#2563eb" strokeWidth={2} dot={false} connectNulls={false} />
          <Line yAxisId="left" type="monotone" dataKey="latency" name="延迟 (ms)" stroke="#7c3aed" strokeWidth={2} dot={false} connectNulls={false} />
          <Line yAxisId="right" type="monotone" dataKey="availability" name="可用率 (%)" stroke="#16a34a" strokeWidth={2} dot={false} connectNulls={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
