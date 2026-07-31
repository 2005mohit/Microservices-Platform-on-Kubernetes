const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Project {
  id: string;
  name: string;
  git_repo: string;
  git_branch: string;
  created_at: string;
  updated_at: string;
  status: string;
  deployment_count?: number;
}

export interface Deployment {
  id: string;
  project_id: string;
  status: string;
  branch: string;
  commit_sha: string;
  logs: string;
  domain: string;
  created_at: string;
  updated_at: string;
}

export async function fetchProjects(): Promise<Project[]> {
  const res = await fetch(`${API_BASE}/api/v1/projects`);
  if (!res.ok) throw new Error('Failed to fetch projects');
  return res.json();
}

export async function fetchProject(id: string): Promise<Project> {
  const res = await fetch(`${API_BASE}/api/v1/projects/${id}`);
  if (!res.ok) throw new Error('Project not found');
  return res.json();
}

export async function createProject(data: {
  name: string;
  git_repo: string;
  git_branch?: string;
  env_vars?: Record<string, string>;
}): Promise<Project> {
  const res = await fetch(`${API_BASE}/api/v1/projects`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || 'Failed to create project');
  }
  return res.json();
}

export async function deleteProject(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/v1/projects/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Failed to delete project');
}

export async function triggerDeploy(
  projectId: string,
  data?: { branch?: string; commit_sha?: string }
): Promise<{ deployment_id: string; status: string; domain: string }> {
  const res = await fetch(`${API_BASE}/api/v1/projects/${projectId}/deploy`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data || {}),
  });
  if (!res.ok) throw new Error('Failed to trigger deploy');
  return res.json();
}

export async function fetchDeployment(id: string): Promise<Deployment> {
  const res = await fetch(`${API_BASE}/api/v1/deployments/${id}`);
  if (!res.ok) throw new Error('Deployment not found');
  return res.json();
}

export async function fetchDeployments(projectId: string): Promise<Deployment[]> {
  const res = await fetch(`${API_BASE}/api/v1/projects/${projectId}/deployments`);
  if (!res.ok) throw new Error('Failed to fetch deployments');
  return res.json();
}
