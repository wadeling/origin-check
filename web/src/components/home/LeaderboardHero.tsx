import { StatsOverview } from '@/components/home/StatsOverview';

export function LeaderboardHero({ relayCount }: { relayCount: number }) {
  return (
    <section className="relative overflow-hidden pb-2 pt-4 sm:pt-8">
      <div className="pointer-events-none absolute inset-0 bg-hero-gradient" aria-hidden />
      <div className="pointer-events-none absolute inset-0 bg-grid-pattern bg-grid opacity-60" aria-hidden />

      <div className="relative animate-fade-up opacity-0" style={{ animationDelay: '0ms' }}>
        <p className="inline-flex items-center gap-2 rounded-full border border-line bg-surface px-3 py-1 text-xs font-medium text-foreground-muted">
          <span className="relative flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-pass opacity-60" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-pass" />
          </span>
          实时探测 · {relayCount} 个中转站
        </p>

        <h1 className="mt-5 max-w-3xl text-4xl font-bold leading-[1.1] tracking-tight sm:text-5xl lg:text-6xl">
          <span className="text-gradient">AI API 中转站</span>
          <br />
          <span className="text-foreground">真伪与性能评测</span>
        </h1>

        <p className="mt-5 max-w-2xl text-base leading-relaxed text-foreground-muted sm:text-lg">
          平台定时探测模型一致性、CDN 缓存与响应性能，用数据帮你识别「挂羊头卖狗肉」与不稳定服务。
        </p>
      </div>
    </section>
  );
}

export { StatsOverview };
