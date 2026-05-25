import type { Metadata } from 'next';
import { DM_Sans, JetBrains_Mono } from 'next/font/google';
import { SiteHeader } from '@/components/SiteHeader';
import { ThemeProvider } from '@/components/ThemeProvider';
import './globals.css';

const dmSans = DM_Sans({
  subsets: ['latin'],
  variable: '--font-dm-sans',
  display: 'swap',
});

const jetbrains = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-jetbrains',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Origin Check — AI API 中转站评测',
  description: '评测 AI API 中转站的模型真伪与性能表现',
};

const themeScript = `(function(){try{var t=localStorage.getItem('origin-check-theme');var d=t==='dark'||(t!=='light'&&window.matchMedia('(prefers-color-scheme: dark)').matches);document.documentElement.classList.toggle('dark',d);}catch(e){}})();`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning className={`${dmSans.variable} ${jetbrains.variable}`}>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body className="font-sans">
        <ThemeProvider>
          <div className="relative flex min-h-screen flex-col">
            <SiteHeader />
            <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6 sm:py-10">{children}</main>
            <footer className="border-t border-line py-8 text-center text-xs text-foreground-muted">
              <p>评测结果仅供参考，不构成商业指控。</p>
            </footer>
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
