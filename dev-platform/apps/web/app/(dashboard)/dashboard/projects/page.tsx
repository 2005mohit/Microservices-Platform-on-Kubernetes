'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { ProjectCard } from '@/components/dashboard/project-card';
import { fetchProjects } from '@/lib/api';
import { FolderGit2, Plus } from 'lucide-react';
import { useState } from 'react';

export default function ProjectsPage() {
  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: fetchProjects,
    refetchInterval: 5000,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Projects</h1>
          <p className="text-muted-foreground mt-1">Manage your deployed applications</p>
        </div>
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
      ) : !projects?.length ? (
        <div className="rounded-xl border border-border bg-card p-12 text-center">
          <FolderGit2 className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">No projects</h3>
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
          {projects.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  );
}
