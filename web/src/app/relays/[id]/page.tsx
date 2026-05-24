import Link from 'next/link';
import { notFound } from 'next/navigation';
import { AuthenticityReportView } from '@/components/AuthenticityReportView';
import { MetricsChart } from '@/components/MetricsChart';
import { VerdictBadge } from '@/components/RelayTable';
import { formatMs, formatPct, formatDateTimeCST, getRelay, getRelayMetrics } from '@/lib/api';

export default async function RelayDetailPage({ params }: { params: { id: string } }) {
  let data;
  let metrics;

  try {
    data = await getRelay(params.id);
    metrics = await getRelayMetrics(params.id);
  } catch {
    notFound();
  }

  const { relay, recent_results, availability_24h, latest_latency_ms, latest_ttft_ms, authenticity_report } = data;

  return (
    <div className="space-y-8">
      <div>
        <Link href="/" className="text-sm text-accent hover:underline">← 返回排行榜</Link>
        <div className="mt-4 flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">{relay.name}</h1>
            <p className="mt-2 text-sm text-muted">
              <a href={relay.website_url} target="_blank" rel="noreferrer" className="text-accent hover:underline">
                {relay.website_url}
              </a>
            </p>
            <p className="mt-1 text-xs text-muted">API: {relay.api_base_url}</p>
          </div>
          <VerdictBadge verdict={relay.authenticity_verdict ?? authenticity_report?.verdict} />
        </div>
      </div>

      <section className="grid gap-4 sm:grid-cols-3">
        <StatCard label="TTFT" value={formatMs(latest_ttft_ms)} />
        <StatCard label="端到端延迟" value={formatMs(latest_latency_ms)} />
        <StatCard label="24h 可用率" value={formatPct(availability_24h)} />
      </section>

      <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-lg font-semibold">24 小时性能趋势</h2>
        <div className="mt-4">
          <MetricsChart metrics={metrics} />
        </div>
      </section>

      <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">真伪鉴定报告</h2>
          <Link href={`/relays/${params.id}/report`} className="text-sm text-accent hover:underline">
            查看历史报告 →
          </Link>
        </div>
        <div className="mt-4">
          {authenticity_report ? (
            <AuthenticityReportView report={authenticity_report} />
          ) : (
            <p className="text-sm text-muted">尚无鉴定报告，等待后台 authenticity 任务运行。</p>
          )}
        </div>
      </section>

      <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-lg font-semibold">最近探测记录</h2>
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted">
                <th className="py-2 pr-4">时间</th>
                <th className="py-2 pr-4">类型</th>
                <th className="py-2 pr-4">模型</th>
                <th className="py-2 pr-4">TTFT</th>
                <th className="py-2 pr-4">延迟</th>
                <th className="py-2">状态</th>
              </tr>
            </thead>
            <tbody>
              {recent_results.map((r) => (
                <tr key={r.id} className="border-b border-slate-100">
                  <td className="py-2 pr-4">{formatDateTimeCST(r.created_at)}</td>
                  <td className="py-2 pr-4">{r.probe_type}</td>
                  <td className="py-2 pr-4">{r.model}</td>
                  <td className="py-2 pr-4">{formatMs(r.ttft_ms)}</td>
                  <td className="py-2 pr-4">{formatMs(r.latency_ms)}</td>
                  <td className="py-2">{r.error ? <span className="text-fail">失败</span> : <span className="text-pass">成功</span>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="text-xs uppercase tracking-wide text-muted">{label}</div>
      <div className="mt-2 text-2xl font-semibold">{value}</div>
    </div>
  );
}
