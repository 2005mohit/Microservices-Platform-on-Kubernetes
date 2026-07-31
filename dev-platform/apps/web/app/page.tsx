import Link from 'next/link';

export default function LandingPage() {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-border">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-violet-500 to-indigo-600 flex items-center justify-center">
              <span className="text-white text-xs font-bold">D</span>
            </div>
            <span className="font-semibold text-lg">DevPlatform</span>
          </div>
          <Link
            href="/dashboard"
            className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground px-5 py-2 text-sm font-medium hover:opacity-90 transition-opacity"
          >
            Launch Dashboard
          </Link>
        </div>
      </header>

      <main className="flex-1">
        <section className="max-w-7xl mx-auto px-6 pt-24 pb-16 text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-border bg-secondary px-4 py-1.5 text-sm text-muted-foreground mb-8">
            Internal Developer Platform
          </div>
          <h1 className="text-5xl sm:text-6xl md:text-7xl font-bold tracking-tight leading-[0.95] mb-6">
            Deploy Apps{' '}
            <span className="bg-gradient-to-r from-violet-400 to-indigo-400 bg-clip-text text-transparent">
              Instantly
            </span>
          </h1>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto mb-10 leading-relaxed">
            Self-hosted platform for internal teams. Connect your Git repository,
            and we build, deploy, and host your application on Kubernetes.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/dashboard"
              className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground px-8 py-3 text-base font-medium hover:opacity-90 transition-opacity"
            >
              Go to Dashboard
            </Link>
            <Link
              href="/dashboard/projects/new"
              className="inline-flex items-center justify-center rounded-lg border border-border bg-secondary text-foreground px-8 py-3 text-base font-medium hover:bg-secondary/80 transition-colors"
            >
              Create Project
            </Link>
          </div>
        </section>

        <section className="max-w-7xl mx-auto px-6 pb-24">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {[
              { title: 'Git-Backed Deployments', desc: 'Connect any Git repository and deploy with one click. Automatic builds on every push.' },
              { title: 'Real-Time Logs', desc: 'Watch build and deployment logs stream live via WebSocket. Debug with full terminal output.' },
              { title: 'Kubernetes Native', desc: 'Built on Kubernetes with automatic namespace isolation, resource quotas, and TLS via cert-manager.' },
            ].map((f) => (
              <div key={f.title} className="rounded-xl border border-border bg-card p-6">
                <div className="h-2 w-12 rounded-full bg-primary mb-4" />
                <h3 className="font-semibold text-lg mb-2">{f.title}</h3>
                <p className="text-sm text-muted-foreground">{f.desc}</p>
              </div>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
