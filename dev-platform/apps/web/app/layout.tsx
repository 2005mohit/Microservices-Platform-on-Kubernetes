import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'DevPlatform - Internal Developer Platform',
  description: 'Self-hosted Vercel-like platform for internal teams',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen bg-background antialiased">{children}</body>
    </html>
  );
}
