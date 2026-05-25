'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ThemeToggle } from './ThemeToggle';

const links = [
  { href: '/', label: '排行榜' },
  { href: '/about', label: '方法论' },
];

export function SiteHeader() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-50 border-b border-line glass">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
        <Link href="/" className="group flex items-center gap-2.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-soft text-sm font-bold text-accent transition-transform duration-300 group-hover:scale-105">
            OC
          </span>
          <span className="font-semibold tracking-tight text-foreground">Origin Check</span>
        </Link>

        <nav className="flex items-center gap-1 sm:gap-2">
          {links.map((link) => {
            const active = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`rounded-full px-3 py-1.5 text-sm transition-all duration-200 sm:px-4 ${
                  active
                    ? 'bg-accent-soft font-medium text-accent'
                    : 'text-foreground-muted hover:bg-surface-muted hover:text-foreground'
                }`}
              >
                {link.label}
              </Link>
            );
          })}
          <div className="ml-1 sm:ml-2">
            <ThemeToggle />
          </div>
        </nav>
      </div>
    </header>
  );
}
