import { RelayTable } from '@/components/RelayTable';
import { getLeaderboard, type RelaySummary } from '@/lib/api';

export default async function HomePage() {
  let relays: RelaySummary[] = [];
  let error: string | null = null;

  try {
    relays = await getLeaderboard();
  } catch (e) {
    error = e instanceof Error ? e.message : '加载失败';
  }

  return (
    <div className="space-y-6">
      <section>
        <h1 className="text-3xl font-bold tracking-tight">AI API 中转站评测</h1>
        <p className="mt-2 max-w-2xl text-muted">
          平台定时探测各中转站的模型一致性与性能指标，帮你识别「挂羊头卖狗肉」与不稳定服务。
        </p>
      </section>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-fail">
          无法连接后端 API：{error}。请确认 API 服务已启动。
        </div>
      ) : relays.length === 0 ? (
        <div className="rounded-lg border border-slate-200 bg-white p-8 text-center text-muted">
          暂无收录的中转站，请运行 seed 命令导入数据。
        </div>
      ) : (
        <RelayTable relays={relays} />
      )}
    </div>
  );
}
