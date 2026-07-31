'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { ProjectCard } from '@/components/dashboard/project-card';
import { fetchProjects } from '@/lib/api';
import { FolderGit2, Plus, Activity } from 'lucide-react';

export default function DashboardPage() {
  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: fetchProjects,
    refetchInterval: 10000,
  });

  const totalDeployments = projects?.reduce((acc, p) => acc + (p.deployment_count || 0), 0) || 0;
  const activeProjects = projects?.filter((p) => p.status === 'success' || p.status === 'deploying').length || 0;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground mt-1">Overview of your platform</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <FolderGit2 className="w-5 h-5 text-primary" />
            </div>
            <div>
              <p className="text-2xl font-bold">{projects?.length || 0}</p>
              <p className="text-sm text-muted-foreground">Total Projects</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-emerald-500/10 flex items-center justify-center">
              <Activity className="w-5 h-5 text-emerald-500" />
            </div>
            <div>
              <p className="text-2xl font-bold">{activeProjects}</p>
              <p className="text-sm text-muted-foreground">Active</p>
            </div>
          </div>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-violet-500/10 flex items-center justify-center">
              <Activity className="w-5 h-5 text-violet-500" />
            </div>
            <div>
              <p className="text-2xl font-bold">{totalDeployments}</p>
              <p className="text-sm text-muted-foreground">Total Deployments</p>
            </div>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Recent Projects</h2>
        <Link
          href="/dashboard/projects/new"
          className="inline-flex items-center gap-2 rounded-lg bg-primary text-primary-foreground px-4 py-2 text-sm font-medium hover:opacity-90 transition-opacity"
        >
          <Plus className="w-4 h-4" />
          New Project
        </Link>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border border-border bg-card p-6 animate-pulse">
              <div className="h-4 bg-secondary rounded w-3/4 mb-3" />
              <div className="h-3 bg-secondary rounded w-1/2 mb-2" />
              <div className="h-3 bg-secondary rounded w-2/3" />
            </div>
          ))}
        </div>
      ) : projects?.length === 0 ? (
        <div className="rounded-xl border border-border bg-card p-12 text-center">
          <FolderGit2 className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">No projects yet</h3>
          <p className="text-muted-foreground mb-6">Create your first project to get started</p>
          <Link
            href="/dashboard/projects/new"
            className="inline-flex items-center gap-2 rounded-lg bg-primary text-primary-foreground px-5 py-2.5 text-sm font-medium hover:opacity-90 transition-opacity"
          >
            <Plus className="w-4 h-4" />
            Create Project
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects?.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  );
}
