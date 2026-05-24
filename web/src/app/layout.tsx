import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Origin Check — AI API 中转站评测',
  description: '评测 AI API 中转站的模型真伪与性能表现',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <header className="border-b border-slate-200 bg-white">
          <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-4">
            <a href="/" className="text-lg font-bold text-ink">Origin Check</a>
            <nav className="flex gap-4 text-sm text-muted">
              <a href="/" className="hover:text-accent">排行榜</a>
              <a href="/about" className="hover:text-accent">方法论</a>
            </nav>
          </div>
        </header>
        <main className="mx-auto min-h-screen max-w-6xl px-4 py-8">{children}</main>
        <footer className="border-t border-slate-200 py-6 text-center text-xs text-muted">
          评测结果仅供参考，不构成商业指控。
        </footer>
      </body>
    </html>
  );
}
