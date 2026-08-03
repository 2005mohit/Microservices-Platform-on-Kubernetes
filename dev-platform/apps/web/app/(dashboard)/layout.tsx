'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import { LayoutDashboard, FolderGit2, Plus, Terminal, Settings } from 'lucide-react';
import { useState } from 'react';

const queryClient = new QueryClient();

const navItems = [
  { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { href: '/dashboard/projects', label: 'Projects', icon: FolderGit2 },
  { href: '/dashboard/projects/new', label: 'New Project', icon: Plus },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen flex">
        <aside className="hidden md:flex w-60 flex-col border-r border-border bg-card">
          <div className="h-16 border-b border-border flex items-center gap-3 px-5">
            <div className="h-7 w-7 rounded-md bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center">
              <span className="text-white text-[10px] font-bold">D</span>
            </div>
            <span className="font-semibold">DevPlatform</span>
          </div>

          <nav className="flex-1 p-3 space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    'flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-lg transition-all',
                    isActive
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-secondary'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {item.label}
                </Link>
              );
            })}
          </nav>

          <div className="p-3 border-t border-border">
            <Link
              href="/dashboard/projects"
              className="flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-all"
            >
              <Terminal className="w-4 h-4" />
              Deployments
            </Link>
          </div>
        </aside>

        <div className="flex-1 flex flex-col min-w-0">
          <header className="h-16 border-b border-border flex items-center px-6 gap-4">
            <button
              className="md:hidden p-2 -ml-2 text-muted-foreground hover:text-foreground"
              onClick={() => setSidebarOpen(!sidebarOpen)}
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <div className="flex-1" />
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="h-2 w-2 rounded-full bg-emerald-500" />
              All Systems Operational
            </div>
          </header>

          {sidebarOpen && (
            <div className="md:hidden border-b border-border bg-card p-3 space-y-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={() => setSidebarOpen(false)}
                    className={cn(
                      'flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-lg transition-all',
                      isActive ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    <Icon className="w-4 h-4" />
                    {item.label}
                  </Link>
                );
              })}
            </div>
          )}

          <main className="flex-1 overflow-auto p-6">{children}</main>
          <footer className="border-t border-border px-6 py-3 text-center text-xs text-muted-foreground">
            &copy; 2026 mohit chandra fulara . All rights reserved.
          </footer>
        </div>
      </div>
    </QueryClientProvider>
  );
}
