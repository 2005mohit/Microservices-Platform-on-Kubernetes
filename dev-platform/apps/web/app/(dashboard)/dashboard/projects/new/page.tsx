'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { createProject } from '@/lib/api';
import { ArrowLeft, Loader2 } from 'lucide-react';
import Link from 'next/link';

export default function NewProjectPage() {
  const router = useRouter();
  const [name, setName] = useState('');
  const [gitRepo, setGitRepo] = useState('');
  const [gitBranch, setGitBranch] = useState('main');
  const [envKey, setEnvKey] = useState('');
  const [envVal, setEnvVal] = useState('');
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const addEnvVar = () => {
    if (!envKey.trim()) return;
    setEnvVars((prev) => ({ ...prev, [envKey.trim()]: envVal }));
    setEnvKey('');
    setEnvVal('');
  };

  const removeEnvVar = (key: string) => {
    setEnvVars((prev) => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !gitRepo.trim()) return;
    setLoading(true);
    setError('');

    try {
      const project = await createProject({ name: name.trim(), git_repo: gitRepo.trim(), git_branch: gitBranch, env_vars: envVars });
      router.push('/dashboard/projects');
      router.refresh();
    } catch (err: any) {
      setError(err.message || 'Failed to create project');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <Link href="/dashboard/projects" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-4">
          <ArrowLeft className="w-4 h-4" />
          Back to Projects
        </Link>
        <h1 className="text-2xl font-bold">Create Project</h1>
        <p className="text-muted-foreground mt-1">Deploy a new application from a Git repository</p>
      </div>

      <form onSubmit={handleSubmit} className="rounded-xl border border-border bg-card p-6 space-y-5">
        <div className="space-y-2">
          <label className="text-sm font-medium">Project Name</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="my-awesome-app"
            className="w-full rounded-lg border border-input bg-background px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
          <p className="text-xs text-muted-foreground">Used as the subdomain (my-awesome-app.devplatform.local)</p>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Git Repository URL</label>
          <input
            value={gitRepo}
            onChange={(e) => setGitRepo(e.target.value)}
            placeholder="https://github.com/username/repo.git"
            className="w-full rounded-lg border border-input bg-background px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Branch</label>
          <input
            value={gitBranch}
            onChange={(e) => setGitBranch(e.target.value)}
            placeholder="main"
            className="w-full rounded-lg border border-input bg-background px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="space-y-3">
          <label className="text-sm font-medium">Environment Variables</label>
          <div className="flex gap-2">
            <input
              value={envKey}
              onChange={(e) => setEnvKey(e.target.value)}
              placeholder="KEY"
              className="flex-1 rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring font-mono"
            />
            <input
              value={envVal}
              onChange={(e) => setEnvVal(e.target.value)}
              placeholder="VALUE"
              className="flex-[2] rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring font-mono"
            />
            <button
              type="button"
              onClick={addEnvVar}
              className="rounded-lg bg-secondary text-secondary-foreground px-3 py-2 text-sm font-medium hover:opacity-80 transition-opacity"
            >
              Add
            </button>
          </div>
          {Object.keys(envVars).length > 0 && (
            <div className="space-y-1.5">
              {Object.entries(envVars).map(([k, v]) => (
                <div key={k} className="flex items-center justify-between rounded-lg bg-secondary px-3 py-2">
                  <span className="text-sm font-mono">
                    <span className="text-primary">{k}</span>=<span className="text-muted-foreground">{v}</span>
                  </span>
                  <button type="button" onClick={() => removeEnvVar(k)} className="text-muted-foreground hover:text-destructive text-sm">
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {error && (
          <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={loading || !name.trim() || !gitRepo.trim()}
          className="w-full rounded-lg bg-primary text-primary-foreground py-2.5 text-sm font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity inline-flex items-center justify-center gap-2"
        >
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
          {loading ? 'Creating...' : 'Create Project'}
        </button>
      </form>
    </div>
  );
}
