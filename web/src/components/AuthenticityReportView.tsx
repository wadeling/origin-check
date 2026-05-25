import type { AuthenticityReport } from '@/lib/api';
import { verdictLabel } from '@/lib/api';

function metadataAlertLabel(alert?: string) {
  switch (alert) {
    case 'metadata_missing':
      return { text: 'API metadata 缺失', className: 'text-fail font-medium' };
    case 'metadata_partial':
      return { text: 'API metadata 部分缺失', className: 'text-warn font-medium' };
    default:
      return null;
  }
}

function signalBorderClass(signal: AuthenticityReport['signals'][0]) {
  if (signal.signal !== 'metadata') return 'border-slate-200';
  if (signal.alert === 'metadata_missing') return 'border-red-300 bg-red-50/50';
  if (signal.alert === 'metadata_partial') return 'border-yellow-300 bg-yellow-50/50';
  return 'border-slate-200';
}

export function AuthenticityReportView({ report }: { report: AuthenticityReport }) {
  const badge = verdictLabel(report.verdict);
  const metadataSignal = report.signals.find((s) => s.signal === 'metadata');
  const metadataAlert = metadataAlertLabel(metadataSignal?.alert);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <span className={`rounded-full border px-3 py-1 text-sm font-medium ${badge.className}`}>{badge.text}</span>
        <span className="text-sm text-muted">模型: {report.claimed_model}</span>
        <span className="text-sm text-muted">得分: {report.score.toFixed(1)}</span>
        <span className="text-sm text-muted">置信度: {report.confidence.toFixed(1)}%</span>
        {metadataAlert && (
          <span className={`rounded-full border border-current px-2 py-0.5 text-xs ${metadataAlert.className}`}>
            {metadataAlert.text}
          </span>
        )}
      </div>

      <div className="space-y-3">
        {report.signals.map((signal) => (
          <details
            key={signal.signal}
            className={`rounded-lg border bg-white p-4 ${signalBorderClass(signal)}`}
          >
            <summary className="cursor-pointer font-medium text-ink">
              {signal.signal} — {signal.score.toFixed(0)} 分
              {signal.signal === 'metadata' && signal.alert === 'metadata_missing' && (
                <span className="ml-2 text-sm text-fail">（metadata 缺失）</span>
              )}
              {signal.signal === 'metadata' && signal.alert === 'metadata_partial' && (
                <span className="ml-2 text-sm text-warn">（部分缺失）</span>
              )}
            </summary>
            <p className={`mt-2 text-sm ${signal.alert === 'metadata_missing' ? 'text-fail' : 'text-muted'}`}>
              {signal.detail}
            </p>
            {signal.response && (signal.response.includes('\n') || signal.signal === 'metadata' || signal.signal === 'cache') ? (
              <pre className="mt-2 overflow-x-auto whitespace-pre-wrap text-xs text-muted">
                {signal.signal === 'metadata' && (
                  <strong className="text-ink">各请求 model 字段：</strong>
                )}
                {signal.signal === 'metadata' && '\n'}
                {signal.response}
              </pre>
            ) : signal.response ? (
              <p className="mt-1 text-xs text-muted">
                <strong>Response:</strong> {signal.response}
              </p>
            ) : null}
            {signal.prompt && (
              <p className="mt-2 text-xs text-muted">
                <strong>Prompt:</strong> {signal.prompt}
              </p>
            )}
          </details>
        ))}
      </div>
    </div>
  );
}
