import Link from 'next/link';
import { verdictLabel, type RelaySummary } from '@/lib/api';

export function VerdictBadge({ verdict }: { verdict?: RelaySummary['authenticity_verdict'] }) {
  const { text, className } = verdictLabel(verdict);
  return (
    <span className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium backdrop-blur-sm ${className}`}>
      {text}
    </span>
  );
}

export function RelayTable({ relays }: { relays: RelaySummary[] }) {
  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <table className="min-w-full divide-y divide-slate-200">
        <thead className="bg-slate-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">中转站</th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">真伪</th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">TTFT</th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">24h 可用率</th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">性能分</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {relays.map((relay) => (
            <tr key={relay.id} className="hover:bg-slate-50">
              <td className="px-4 py-4">
                <Link href={`/relays/${relay.id}`} className="font-medium text-accent hover:underline">
                  {relay.name}
                </Link>
                <div className="mt-1 text-xs text-muted">{relay.claimed_models.join(', ')}</div>
              </td>
              <td className="px-4 py-4">
                <VerdictBadge verdict={relay.authenticity_verdict} />
                {relay.authenticity_score != null && (
                  <div className="mt-1 text-xs text-muted">近6次均 {relay.authenticity_score.toFixed(0)} 分</div>
                )}
              </td>
              <td className="px-4 py-4 text-sm">{relay.latest_ttft_ms != null ? `${Math.round(relay.latest_ttft_ms)} ms` : '—'}</td>
              <td className="px-4 py-4 text-sm">{relay.availability_24h != null ? `${relay.availability_24h.toFixed(1)}%` : '—'}</td>
              <td className="px-4 py-4 text-sm">{relay.perf_score != null ? relay.perf_score.toFixed(0) : '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
