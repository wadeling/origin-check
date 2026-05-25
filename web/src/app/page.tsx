import { AuthenticityOverview } from '@/components/home/AuthenticityOverview';
import { LeaderboardHero } from '@/components/home/LeaderboardHero';
import { PerformanceOverview } from '@/components/home/PerformanceOverview';
import { RelayLeaderboard } from '@/components/home/RelayLeaderboard';
import { StatsOverview } from '@/components/home/StatsOverview';
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
    <div className="space-y-10 sm:space-y-12">
      <LeaderboardHero relayCount={relays.length} />

      {error ? (
        <div className="animate-fade-in rounded-2xl border border-fail/30 bg-fail/10 p-5 text-sm text-fail">
          无法连接后端 API：{error}。请确认 API 服务已启动。
        </div>
      ) : relays.length === 0 ? (
        <div className="glass rounded-2xl p-12 text-center text-foreground-muted shadow-card">
          暂无收录的中转站，请运行 seed 命令导入数据。
        </div>
      ) : (
        <>
          <StatsOverview relays={relays} />

          <div className="grid gap-6 xl:grid-cols-5">
            <div className="xl:col-span-3">
              <PerformanceOverview relays={relays} />
            </div>
            <div className="xl:col-span-2">
              <AuthenticityOverview relays={relays} />
            </div>
          </div>

          <RelayLeaderboard relays={relays} />
        </>
      )}
    </div>
  );
}
