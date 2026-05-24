import Link from 'next/link';
import { notFound } from 'next/navigation';
import { AuthenticityReportView } from '@/components/AuthenticityReportView';
import { getRelay, getRelayReports } from '@/lib/api';

export default async function RelayReportPage({ params }: { params: { id: string } }) {
  let relay;
  let reports;

  try {
    relay = await getRelay(params.id);
    reports = await getRelayReports(params.id);
  } catch {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div>
        <Link href={`/relays/${params.id}`} className="text-sm text-accent hover:underline">
          ← 返回 {relay.relay.name}
        </Link>
        <h1 className="mt-4 text-3xl font-bold">真伪鉴定历史</h1>
        <p className="mt-2 text-muted">{relay.relay.name} 的全部鉴定报告</p>
      </div>

      {reports.length === 0 ? (
        <p className="text-muted">暂无历史报告。</p>
      ) : (
        <div className="space-y-8">
          {reports.map((report) => (
            <section key={report.id} className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
              <h2 className="text-sm font-medium text-muted">
                {new Date(report.created_at).toLocaleString('zh-CN')}
              </h2>
              <div className="mt-4">
                <AuthenticityReportView report={report} />
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
