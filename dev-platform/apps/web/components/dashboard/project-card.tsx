'use client';

import Link from 'next/link';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { triggerDeploy, deleteProject, type Project } from '@/lib/api';
import { Rocket, Trash2, ExternalLink, Loader2, GitBranch } from 'lucide-react';
import { useState } from 'react';

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  success: { color: 'bg-emerald-500', label: 'Live' },
  building: { color: 'bg-blue-500', label: 'Building' },
  deploying: { color: 'bg-violet-500', label: 'Deploying' },
  queued: { color: 'bg-yellow-500', label: 'Queued' },
  failed: { color: 'bg-red-500', label: 'Failed' },
  inactive: { color: 'bg-gray-500', label: 'Inactive' },
};

export function ProjectCard({ project }: { project: Project }) {
  const queryClient = useQueryClient();
  const [deploying, setDeploying] = useState(false);
  const config = STATUS_CONFIG[project.status] || STATUS_CONFIG.inactive;

  const deployMutation = useMutation({
    mutationFn: () => triggerDeploy(project.id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      window.open(`/dashboard/deployments/${data.deployment_id}`, '_blank');
    },
    onSettled: () => setDeploying(false),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(project.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['projects'] }),
  });

  const handleDeploy = () => {
    setDeploying(true);
    deployMutation.mutate();
  };

  return (
    <div className="rounded-xl border border-border bg-card hover:border-primary/20 transition-colors group">
      <div className="p-5">
        <div className="flex items-start justify-between mb-3">
          <div>
            <Link href={`/dashboard/deployments?project=${project.id}`} className="font-semibold hover:text-primary transition-colors">
              {project.name}
            </Link>
            <div className="flex items-center gap-2 mt-1">
              <span className={`h-2 w-2 rounded-full ${config.color} ${project.status === 'building' || project.status === 'deploying' ? 'animate-pulse' : ''}`} />
              <span className="text-xs text-muted-foreground">{config.label}</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2 text-xs text-muted-foreground mb-4">
          <GitBranch className="w-3.5 h-3.5" />
          {project.git_branch || 'main'}
          <span className="text-border">|</span>
          <span className="truncate max-w-[200px]">{project.git_repo.replace('https://', '')}</span>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleDeploy}
            disabled={deploying}
            className="inline-flex items-center gap-1.5 rounded-lg bg-primary/10 text-primary px-3 py-1.5 text-xs font-medium hover:bg-primary/20 transition-colors disabled:opacity-50"
          >
            {deploying ? <Loader2 className="w-3 h-3 animate-spin" /> : <Rocket className="w-3 h-3" />}
            {deploying ? 'Deploying...' : 'Deploy'}
          </button>

          <button
            onClick={() => deleteMutation.mutate()}
            disabled={deleteMutation.isPending}
            className="inline-flex items-center gap-1.5 rounded-lg bg-destructive/10 text-destructive px-3 py-1.5 text-xs font-medium hover:bg-destructive/20 transition-colors disabled:opacity-50"
          >
            <Trash2 className="w-3 h-3" />
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}
