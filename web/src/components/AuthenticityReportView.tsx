import type { AuthenticityReport } from '@/lib/api';
import { verdictLabel } from '@/lib/api';

export function AuthenticityReportView({ report }: { report: AuthenticityReport }) {
  const badge = verdictLabel(report.verdict);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <span className={`rounded-full border px-3 py-1 text-sm font-medium ${badge.className}`}>{badge.text}</span>
        <span className="text-sm text-muted">模型: {report.claimed_model}</span>
        <span className="text-sm text-muted">得分: {report.score.toFixed(1)}</span>
        <span className="text-sm text-muted">置信度: {report.confidence.toFixed(1)}%</span>
      </div>

      <div className="space-y-3">
        {report.signals.map((signal) => (
          <details key={signal.signal} className="rounded-lg border border-slate-200 bg-white p-4">
            <summary className="cursor-pointer font-medium text-ink">
              {signal.signal} — {signal.score.toFixed(0)} 分
            </summary>
            <p className="mt-2 text-sm text-muted">{signal.detail}</p>
            {signal.prompt && <p className="mt-2 text-xs text-muted"><strong>Prompt:</strong> {signal.prompt}</p>}
            {signal.response && <p className="mt-1 text-xs text-muted"><strong>Response:</strong> {signal.response}</p>}
          </details>
        ))}
      </div>
    </div>
  );
}
